package main

import (
	"io"
	"strconv"
)

type ErrorString string

func (e ErrorString) Error() string { return string(e) }

const (
	ErrMissingProtocol           = ErrorString("missing protocol")
	ErrUnknownProtocol           = ErrorString("unknown protocol")
	ErrMissingMethod             = ErrorString("missing method")
	ErrMissingRequestURI         = ErrorString("missing request uri")
	ErrMissingHeaderName         = ErrorString("missing header name")
	ErrMissingHeaderValue        = ErrorString("missing header value")
	ErrDuplicateHost             = ErrorString("duplicate Host header")
	ErrInconsistentContentLength = ErrorString("inconsistent Content-Length header")
	EOH                          = ErrorString("end of header")
	ErrHeaderTooLarge            = ErrorString("header too large")
	ErrMaxHeaderBytesTooSmall    = ErrorString("max header bytes too small")
)

type ErrExpected rune

func (e ErrExpected) Error() string { return "expected " + strconv.QuoteRuneToASCII(rune(e)) }

func unexpectedEOF(err error) error {
	if err == nil || err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

func coalesce(a, b error) error {
	if a != nil {
		return a
	}
	return b
}
