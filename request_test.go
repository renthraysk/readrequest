package main

import (
	"bufio"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var quickTest = "" +
	"GET http://www.techcrunch.com/ HTTP/1.1\r\n" +
	"Host: www.techcrunch.com\r\n" +
	"User-Agent: Fake\r\n" +
	"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\n" +
	"Accept-Language: en-us,en;q=0.5\r\n" +
	"Accept-Encoding: gzip,deflate\r\n" +
	"Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7\r\n" +
	"Keep-Alive: 300\r\n" +
	"Content-Length: 7\r\n" +
	"Proxy-Connection: keep-alive\r\n\r\n" +
	"abcdef\n???"

func TestQuickReadRequest(t *testing.T) {
	rdr := bufio.NewReader(strings.NewReader(quickTest))
	r, err := ReadRequest(rdr)
	if err != nil {
		t.Fatalf("error occured: %v", err)
	}
	assertEqual(t, "Method", r.Method, "GET")
	assertEqual(t, "RequestURI", r.RequestURI, "http://www.techcrunch.com/")
	assertEqual(t, "Proto", r.Proto, "HTTP/1.1")
	assertEqual(t, "ProtoMajor", r.ProtoMajor, 1)
	assertEqual(t, "ProtoMinor", r.ProtoMinor, 1)
	assertEqual(t, "Host", r.Host, "www.techcrunch.com")
	assertEqual(t, "Close", r.Close, true)
	assertEqual(t, "ContentLength", r.ContentLength, 7)
	assertAnyEqual(t, "Header", r.Header, http.Header{
		"Accept":           {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
		"Accept-Language":  {"en-us,en;q=0.5"},
		"Accept-Encoding":  {"gzip,deflate"},
		"Accept-Charset":   {"ISO-8859-1,utf-8;q=0.7,*;q=0.7"},
		"Keep-Alive":       {"300"},
		"Proxy-Connection": {"keep-alive"},
		"Content-Length":   {"7"},
		"User-Agent":       {"Fake"},
	})
	assertAnyEqual(t, "URL", r.URL, &url.URL{
		Scheme: "http",
		Host:   "www.techcrunch.com",
		Path:   "/",
	})
}

func TestDuplicateHosts(t *testing.T) {
	cases := []struct {
		in  string
		err error
	}{
		{"GET / HTTP/1.1\r\n" + "Host: example.org\r\n" + "Host: example.org\r\n\r\n", ErrDuplicateHost},
		{"GET / HTTP/1.1\r\n" + "host: example.org\r\n" + "HOST: example.org\r\n\r\n", ErrDuplicateHost},
	}
	for _, s := range cases {
		rdr := bufio.NewReader(strings.NewReader(s.in))
		_, err := ReadRequest(rdr)
		if err != s.err {
			t.Fatalf("expected error %q, got %q", s.err, err)
		}
	}
}

func TestContentLength(t *testing.T) {
	cases := []struct {
		in  string
		err error
	}{
		{"GET / HTTP/1.1\r\n" + "Content-Length: 7\r\n" + "Content-Length: 7\r\n\r\n", nil},
		{"GET / HTTP/1.1\r\n" + "Content-Length: 7\r\n" + "Content-Length: 8\r\n\r\n", ErrInconsistentContentLength},
		{"GET / HTTP/1.1\r\n" + "Content-Length: 7\r\n" + "Content-Length: 7\r\n" + "Content-Length: 8\r\n\r\n", ErrInconsistentContentLength},
	}
	for _, s := range cases {
		rdr := bufio.NewReader(strings.NewReader(s.in))
		r, err := ReadRequest(rdr)
		if err != s.err {
			t.Fatalf("expected error %q, got %q", s.err, err)
		}
		if s.err == nil {
			assertEqual(t, "Content-Length", r.ContentLength, 7)
		}
	}
}

func TestConnection(t *testing.T) {
	cases := []struct {
		in    string
		close bool
	}{
		{"GET / HTTP/0.0\r\n\r\n", true},
		{"GET / HTTP/0.0\r\nConnection: Close\r\n\r\n", true},
		{"GET / HTTP/0.0\r\nConnection: Keep-Alive\r\n\r\n", true},
		{"GET / HTTP/1.0\r\n\r\n", true},
		{"GET / HTTP/1.0\r\nConnection: Close\r\n\r\n", true},
		{"GET / HTTP/1.0\r\nConnection: Keep-Alive\r\n\r\n", false},
		{"GET / HTTP/1.0\r\nConnection: Close\r\nConnection: Keep-Alive\r\n\r\n", false},
		{"GET / HTTP/1.1\r\n\r\n", true},
		{"GET / HTTP/1.1\r\nConnection: Close\r\n\r\n", true},
		{"GET / HTTP/1.1\r\nConnection: Keep-Alive\r\n\r\n", false},
		{"GET / HTTP/1.1\r\nConnection: Close\r\nConnection: Keep-Alive\r\n\r\n", true},
	}

	for i, s := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			rdr := bufio.NewReader(strings.NewReader(s.in))
			r, err := ReadRequest(rdr)
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}
			assertEqual(t, "Close", r.Close, s.close)
		})
	}
}

func TestPragma(t *testing.T) {
	cases := []struct {
		in           string
		cacheControl []string
	}{
		{"GET / HTTP/0.0\r\n\r\n", nil},
		{"GET / HTTP/0.0\r\nPragma: no-cache\r\n\r\n", []string{"no-cache"}},
		{"GET / HTTP/0.0\r\nPragma: no-cache\r\nCache-Control: public\r\n\r\n", []string{"public"}},
	}
	for i, s := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			rdr := bufio.NewReader(strings.NewReader(s.in))
			r, err := ReadRequest(rdr)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			assertAnyEqual(t, "Cache-Control", r.Header["Cache-Control"], s.cacheControl)
		})
	}
}

func BenchmarkReadRequest(b *testing.B) {
	sr := strings.NewReader(quickTest)
	br := bufio.NewReader(sr)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sr.Reset(quickTest)
		br.Reset(sr)
		ReadRequest(br)
	}
}

func XBenchmarkReadRequest(b *testing.B) {
	r0 := strings.NewReader(quickTest)
	r1 := bufio.NewReader(r0)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r0.Reset(quickTest)
		r1.Reset(r0)
		http.ReadRequest(r1)
	}
}

func assertEqual[T comparable](tb testing.TB, name string, got, expected T) {
	tb.Helper()
	if got != expected {
		tb.Errorf("%s expected %v, got %v", name, expected, got)
	}
}

func assertAnyEqual[T any](tb testing.TB, name string, got, expected T) {
	tb.Helper()
	if !reflect.DeepEqual(got, expected) {
		tb.Errorf("%s\n%s", name, cmp.Diff(expected, got))
	}
}
