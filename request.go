package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (p *parser) Set(r *http.Request, lines string) error {
	pos := 0
	if p.proto != 0 {
		var err error

		r.Method = lines[:p.method]
		pos = p.method + len(" ")
		r.RequestURI = lines[pos:p.requestURI]
		r.URL, err = url.Parse(r.RequestURI)
		if err != nil {
			return err
		}
		pos = p.requestURI + len(" ")
		r.Proto = lines[pos:p.proto]
		r.ProtoMajor = int(r.Proto[len("HTTP/")] - '0')
		r.ProtoMinor = int(r.Proto[len("HTTP/0.")] - '0')
		pos = p.proto + len("\r\n")
		p.proto = 0
	}

	index := p.headerCount
	p.headerCount = 0
	if r.Header == nil {
		r.Header = make(http.Header, index)
	}
	if index == 0 {
		return nil
	}
	values := make([]string, index)
	for pos < len(lines) {
		i := pos + strings.IndexByte(lines[pos:], ':')
		j := i + strings.IndexByte(lines[i:], '\r')
		key, value := lines[pos:i], trim(lines[i+1:j])
		pos = j + len("\r\n")
		if v, ok := r.Header[key]; ok {
			switch key {
			case "Host":
				return ErrDuplicateHost
			case "Content-Length":
				if len(v) > 0 && v[0] != value {
					return ErrInconsistentContentLength
				}
			default:
				r.Header[key] = append(v, value)
			}
		} else {
			index--
			values[index] = value
			r.Header[key] = values[index : index+1 : index+1]
		}
	}
	return nil
}

func ReadRequest(r *bufio.Reader) (*http.Request, error) {
	const peekInitial = 8 << 10
	const peekAdvance = 4 << 10

	buf, err := r.Peek(min(http.DefaultMaxHeaderBytes, peekInitial))
	if len(buf) <= 0 {
		return nil, coalesce(err, io.ErrUnexpectedEOF)
	}
	if !isToken(buf[0]) {
		return nil, ErrMissingMethod
	}

	p := new(parser)
	req := new(http.Request)
	size := 0
	pos, adv, err := p.parseFirstLine(buf, 0)
	for err == nil {
		if adv < len(buf) {
			pos, adv, err = p.newline(buf, pos)
			continue
		}
		if size > http.DefaultMaxHeaderBytes-pos {
			return nil, ErrHeaderTooLarge
		}
		size += pos
		if err = p.Set(req, string(buf[:pos])); err != nil {
			return nil, err
		}
		r.Discard(pos)
		adv -= pos
		buf, err = r.Peek(max(adv, min(http.DefaultMaxHeaderBytes-size, peekAdvance)))
		if adv >= len(buf) {
			return nil, unexpectedEOF(err)
		}
		pos, adv, err = p.newline(buf, 0)
	}
	if err != nil && err != EOH {
		return nil, err
	}
	if size > http.DefaultMaxHeaderBytes-pos {
		return nil, ErrHeaderTooLarge
	}
	size += pos
	if err = p.Set(req, string(buf[:pos-len("\r\n")])); err != nil {
		return nil, err
	}
	r.Discard(pos)

	if req.ContentLength, err = contentLength(req.Header); err != nil {
		return nil, fmt.Errorf("Content-Length: %w", err)
	}

	pragmaCacheControl(req.Header)

	req.Host = host(req.URL, req.Header)
	delete(req.Header, "Host")

	req.Close = close(req.ProtoMajor, req.ProtoMinor, req.Header)

	return req, nil
}

func host(u *url.URL, h http.Header) string {
	if u != nil && u.Host != "" {
		return u.Host
	}
	if v, ok := h["Host"]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

func close(protoMajor, protoMinor int, h http.Header) bool {
	switch protoMajor {
	case 1:
		v, ok := h["Connection"]
		if !ok {
			return true
		}
		var close, keepAlive bool
		for _, s := range v {
			switch s {
			case "close":
				close = true
			case "keep-alive":
				keepAlive = true
			}
		}
		return close && (protoMinor != 0 || !keepAlive)
	case 2:
	}
	return true
}

func pragmaCacheControl(h http.Header) {
	if v, ok := h["Pragma"]; ok && len(v) > 0 && v[0] == "no-cache" {
		if _, ok = h["Cache-Control"]; !ok {
			h["Cache-Control"] = v[:1:1]
		}
	}
}

func contentLength(h http.Header) (int64, error) {
	if v, ok := h["Content-Length"]; ok && len(v) > 0 {
		i, err := strconv.ParseInt(v[0], 10, 64)
		if err != nil {
			i = -1
		}
		return i, err
	}
	return -1, nil
}

func min[T int](a, b T) T {
	if b < a {
		a = b
	}
	return a
}

func max[T int](a, b T) T {
	if a < b {
		a = b
	}
	return a
}

func coalesce(a, b error) error {
	if a != nil {
		return a
	}
	return b
}

func unexpectedEOF(err error) error {
	if err == nil || err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}
