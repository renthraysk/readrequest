package main

import (
	"bufio"
	"net/http"
	"net/url"
	"reflect"
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
	assertEqual(t, "Close", r.Close, false)
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

func BenchmarkStdlibReadRequest(b *testing.B) {
	r0 := strings.NewReader(quickTest)
	r1 := bufio.NewReader(r0)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r0.Reset(quickTest)
		r1.Reset(r0)
		http.ReadRequest(r1)
	}
}
