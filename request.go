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

func (p *parser) Set(r *http.Request, s string) error {
	pos := 0
	if p.proto != 0 {
		r.Method = s[:p.method]
		pos = p.method + len(" ")
		r.RequestURI = s[pos:p.requestURI]
		pos = p.requestURI + len(" ")
		r.Proto = s[pos:p.proto]
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
	for pos < len(s) {
		i := pos + strings.IndexByte(s[pos:], ':')
		j := i + strings.IndexByte(s[i:], '\r')
		key, value := s[pos:i], trim(s[i+1:j])
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

	p := new(parser)

	buf, err := r.Peek(peekInitial)
	if len(buf) <= 0 {
		return nil, coalesce(err, io.ErrUnexpectedEOF)
	}
	if !isToken(buf[0]) {
		return nil, ErrMissingMethod
	}

	req := &http.Request{}

	pos, adv, err := p.parseMethod(buf, 0)
	for err == nil {
		if adv < len(buf) {
			pos, adv, err = p.newline(buf, pos)
			continue
		}
		if err = p.Set(req, string(buf[:pos])); err != nil {
			return nil, err
		}
		r.Discard(pos)
		adv -= pos
		buf, err = r.Peek(max(adv, peekAdvance))
		if adv >= len(buf) {
			return nil, unexpectedEOF(err)
		}
		pos, adv, err = p.newline(buf, 0)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	if err = p.Set(req, string(buf[:pos-len("\r\n")])); err != nil {
		return nil, err
	}
	r.Discard(pos)

	if req.URL, err = url.Parse(req.RequestURI); err != nil {
		return nil, err
	}
	if req.ContentLength, err = ContentLength(req.Header); err != nil {
		return nil, fmt.Errorf("Content-Length: %w", err)
	}

	req.Host = Host(req.URL, req.Header)
	delete(req.Header, "Host")

	req.Close = Close(req.ProtoMajor, req.ProtoMinor, req.Header)

	return req, nil
}

func Host(u *url.URL, h http.Header) string {
	if u != nil && u.Host != "" {
		return u.Host
	}
	if v, ok := h["Host"]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

func Close(protoMajor, protoMinor int, h http.Header) bool {
	if protoMajor < 1 {
		return true
	}
	v, ok := h["Connection"]
	if !ok {
		return false
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
	if protoMajor == 1 && protoMinor == 0 {
		return close && !keepAlive
	}
	return close
}

func ContentLength(header http.Header) (int64, error) {
	if v, ok := header["Content-Length"]; ok && len(v) > 0 {
		i, err := strconv.ParseInt(v[0], 10, 64)
		if err != nil {
			i = -1
		}
		return i, err
	}
	return -1, nil
}

func max(a, b int) int {
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
