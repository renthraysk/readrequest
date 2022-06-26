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
	if !p.parsedFirstLine {
		var err error

		method := strings.IndexByte(lines, ' ')
		if method < 0 {
			return ErrExpectedSpace
		}
		r.Method = lines[:method]
		method++
		requestURI := strings.IndexByte(lines[method:], ' ')
		if requestURI < 0 {
			return ErrExpectedSpace
		}
		requestURI += method
		r.RequestURI = lines[method:requestURI]
		requestURI++
		pos = strings.IndexByte(lines[requestURI:], '\r')
		if pos < 0 {
			return ErrExpectedCarriageReturn
		}
		pos += requestURI
		r.Proto = lines[requestURI:pos]
		r.URL, err = url.Parse(r.RequestURI)
		if err != nil {
			return err
		}
		r.ProtoMajor = int(r.Proto[len("HTTP/")] - '0')
		r.ProtoMinor = int(r.Proto[len("HTTP/0.")] - '0')
		p.parsedFirstLine = true
		pos += len("\r\n")
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
		i := strings.IndexByte(lines[pos:], ':')
		if i < 0 {
			return ErrExpectedColon
		}
		i += pos
		j := strings.IndexByte(lines[i:], '\r')
		if j < 0 {
			return ErrExpectedCarriageReturn
		}
		j += i
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
	return readRequest(r, http.DefaultMaxHeaderBytes)
}

func readRequest(r *bufio.Reader, maxHeaderBytes int) (*http.Request, error) {
	const peekInitial = 8 << 10
	const peekAdvance = 4 << 10

	if maxHeaderBytes < len("M / HTTP0.0/r/n/r/n") {
		return nil, ErrMaxHeaderBytesTooSmall
	}

	buf, err := r.Peek(min(maxHeaderBytes, peekInitial))
	if len(buf) <= 0 {
		return nil, coalesce(err, io.ErrUnexpectedEOF)
	}
	if !isToken(buf[0]) {
		return nil, ErrMissingMethod
	}

	p := &parser{
		remaining: maxHeaderBytes,
	}
	req := new(http.Request)

	pos, adv, err := p.parseFirstLine(buf)
	for err == nil && adv > len(buf) {
		buf, err = r.Peek(adv)
		if adv >= len(buf) {
			return nil, unexpectedEOF(err)
		}
		pos, adv, err = p.parseFirstLine(buf)
	}
	pos, adv, err = p.parseLines(buf, pos)
	for err == nil && adv > len(buf) {
		if pos > p.remaining {
			return nil, ErrHeaderTooLarge
		}
		p.remaining -= pos
		if err = p.Set(req, string(buf[:pos])); err != nil {
			return nil, err
		}
		r.Discard(pos)
		adv -= pos
		buf, err = r.Peek(max(adv, min(p.remaining, peekAdvance)))
		if adv >= len(buf) {
			return nil, unexpectedEOF(err)
		}
		pos, adv, err = p.parseLines(buf, 0)
	}
	if err != nil && err != EOH {
		return nil, err
	}
	if pos > p.remaining {
		return nil, ErrHeaderTooLarge
	}
	p.remaining -= pos
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
	if a <= b {
		return a
	}
	return b
}

func max[T int](a, b T) T {
	if a >= b {
		return a
	}
	return b
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
