package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

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

	pos, adv, err := parseFirstLine(buf)
	for err == nil && adv >= len(buf) {
		if adv > maxHeaderBytes {
			return nil, ErrHeaderTooLarge
		}
		buf, err = r.Peek(adv)
		if adv >= len(buf) {
			return nil, unexpectedEOF(err)
		}
		pos, adv, err = parseFirstLine(buf)
	}
	if err != nil {
		return nil, err
	}
	if pos > maxHeaderBytes {
		return nil, ErrHeaderTooLarge
	}

	var req *http.Request
	var headerCount int

	remaining := maxHeaderBytes - pos
	pos, adv, headerCount, err = parseBlock(buf, pos)
	for err == nil && adv >= len(buf) {
		if pos > remaining {
			return nil, ErrHeaderTooLarge
		}
		remaining -= pos
		if req, err = buildRequest(req, string(buf[:pos]), headerCount); err != nil {
			return nil, err
		}
		r.Discard(pos)
		adv -= pos
		buf, err = r.Peek(max(adv, min(remaining, peekAdvance)))
		if adv >= len(buf) {
			return nil, unexpectedEOF(err)
		}
		pos, adv, headerCount, err = parseBlock(buf, 0)
	}
	if err != nil && err != EOH {
		return nil, err
	}
	if pos > remaining {
		return nil, ErrHeaderTooLarge
	}
	remaining -= pos

	if req, err = buildRequest(req, string(buf[:pos-len("\r\n")]), headerCount); err != nil {
		return nil, err
	}
	r.Discard(pos)

	if req.ContentLength, err = contentLength(req.Header); err != nil {
		return nil, fmt.Errorf("Content-Length: %w", err)
	}

	pragmaCacheControl(req.Header)

	req.Host = host(req.URL, req.Header)
	req.Close = close(req.ProtoMajor, req.ProtoMinor, req.Header)

	return req, nil
}

func host(u *url.URL, h http.Header) string {
	defer delete(h, "Host")
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
