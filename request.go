package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func ReadRequest(r *bufio.Reader) (*http.Request, error) {
	builder, err := New(r)
	if err != nil {
		return nil, err
	}
	return builder.Build()
}

// Builder a thing that creates http.Request{} from a bufio.Reader
type Builder struct {
	s           string
	requestURI  int
	protocol    int
	headers     int
	headerCount int
}

func New(r *bufio.Reader) (*Builder, error) {
	const peekInitial = 4 << 10
	const peekAdvance = 1 << 10

	var p parser

	buf, err := r.Peek(peekInitial)
	if len(buf) <= 0 {
		return nil, coalesce(err, io.ErrUnexpectedEOF)
	}
	if !isToken(buf[0]) {
		return nil, ErrMissingMethod
	}

	next, pos, adv, err := p.parseMethod(buf, 1)
	for next != nil {
		if adv < len(buf) {
			next, pos, adv, err = next(&p, buf, pos)
			continue
		}
		// buf expansion required
		buf, err = r.Peek(max(adv, pos+peekAdvance))
		if adv >= len(buf) {
			return nil, unexpectedEOF(err)
		}
		prev := pos
		next, pos, adv, err = next(&p, buf, pos)
		if prev >= pos {
			return nil, errors.New("parser stuck?")
		}
	}
	if err != nil {
		return nil, err
	}
	defer r.Discard(pos)
	return &Builder{s: string(buf[:pos-len("\r\n")]),
		requestURI:  p.requestURI,
		protocol:    p.protocol,
		headers:     p.headers,
		headerCount: p.headerCount,
	}, nil
}

func (b *Builder) Method() string     { return b.s[:b.requestURI-len(" ")] }
func (b *Builder) RequestURI() string { return b.s[b.requestURI : b.protocol-len(" ")] }
func (b *Builder) Proto() string      { return b.s[b.protocol : b.headers-len("\r\n")] }
func (b *Builder) ProtoMajor() int    { return int(b.s[b.protocol+len("HTTP/")] - '0') }
func (b *Builder) ProtoMinor() int    { return int(b.s[b.protocol+len("HTTP/0.")] - '0') }

func (b *Builder) URL() (*url.URL, error) { return url.Parse(b.RequestURI()) }

func Host(u *url.URL, h http.Header) string {
	if u != nil && u.Host != "" {
		return u.Host
	}
	if v, ok := h["Host"]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

func (b *Builder) Header() (http.Header, error) {
	ssIndex := b.headerCount
	header := make(http.Header, ssIndex)
	if ssIndex == 0 {
		return header, nil
	}
	ssValues := make([]string, ssIndex)
	for i := b.headers; i < len(b.s); {
		j := i + strings.IndexByte(b.s[i:], ':')
		k := j + strings.IndexByte(b.s[j:], '\r')
		key, value := b.s[i:j], trimHorizontalSpace(b.s[j+1:k])
		i = k + len("\r\n")

		if v, ok := header[key]; ok {
			switch key {
			case "Host":
				return nil, errors.New("duplicate Host headers")
			case "Content-Length":
				if len(v) > 0 && v[0] != value {
					return nil, errors.New("duplicate Content-Length headers")
				}
			default:
				header[key] = append(v, value)
			}
		} else {
			ssIndex--
			ssValues[ssIndex] = value
			header[key] = ssValues[ssIndex : ssIndex+1 : ssIndex+1]
		}
	}
	return header, nil
}

func (b *Builder) Close(header http.Header) bool {
	if b.ProtoMajor() < 1 {
		return true
	}
	if v, ok := header["Connection"]; ok {
		var close, keepAlive bool
		for _, s := range v {
			switch s {
			case "close":
				close = true
			case "keep-alive":
				keepAlive = true
			}
		}
		if b.ProtoMajor() == 1 && b.ProtoMinor() == 0 {
			return close && !keepAlive
		}
		return close
	}
	return false
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

// Builder builds a http.Request
func (b *Builder) Build() (*http.Request, error) {
	URL, err := b.URL()
	if err != nil {
		return nil, fmt.Errorf("failed to parse RequestURI: %w", err)
	}
	header, err := b.Header()
	if err != nil {
		return nil, err
	}
	contentLength, err := ContentLength(header)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Content-Length: %w", err)
	}
	host := Host(URL, header)
	delete(header, "Host")

	return &http.Request{
		Method:        b.Method(),
		RequestURI:    b.RequestURI(),
		Proto:         b.Proto(),
		ProtoMajor:    b.ProtoMajor(),
		ProtoMinor:    b.ProtoMinor(),
		Header:        header,
		URL:           URL,
		Host:          host,
		Close:         b.Close(header),
		ContentLength: contentLength,
	}, nil
}

func max(a, b int) int {
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
