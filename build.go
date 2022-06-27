package main

import (
	"net/http"
	"net/url"
)

// buildRequest creates or alters a http.Request{} based on blocks of lines
// from the header. If r is nil, assumes lines includes the first line of the http request,
// and will create the http.Request otherwise it assumes lines is entirely headers.
func buildRequest(r *http.Request, lines string, headerCount int) (*http.Request, error) {
	if r == nil {
		var method, requestURI, proto string
		var ok bool
		var err error

		if method, lines, ok = cut(lines, ' '); !ok {
			return nil, ErrExpected(' ')
		}
		if requestURI, lines, ok = cut(lines, ' '); !ok {
			return nil, ErrExpected(' ')
		}
		if proto, lines, ok = cut(lines, '\r'); !ok {
			return nil, ErrExpected('\r')
		}
		r = &http.Request{
			Method:     method,
			RequestURI: requestURI,
			Proto:      proto,
			ProtoMajor: int(proto[len("HTTP/")] - '0'),
			ProtoMinor: int(proto[len("HTTP/0.")] - '0'),
		}
		if r.URL, err = url.Parse(requestURI); err != nil {
			return nil, err
		}
		lines = lines[len("\n"):]
	}

	if headerCount <= 0 {
		return r, nil
	}
	if r.Header == nil {
		r.Header = make(http.Header, headerCount)
	}
	values := make([]string, headerCount)
	for index := headerCount; len(lines) > 0; lines = lines[len("\n"):] {
		var key, value string
		var ok bool

		if key, lines, ok = cut(lines, ':'); !ok {
			return nil, ErrExpected(':')
		}
		if value, lines, ok = cut(lines, '\r'); !ok {
			return nil, ErrExpected('\r')
		}
		value = trim(value)
		if v, ok := r.Header[key]; ok {
			switch key {
			case "Host":
				return nil, ErrDuplicateHost
			case "Content-Length":
				if len(v) > 0 && v[0] != value {
					return nil, ErrInconsistentContentLength
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
	return r, nil
}
