package main

type ErrorString string

func (e ErrorString) Error() string { return string(e) }

const (
	ErrExpectedColon             = ErrorString("expected colon")
	ErrExpectedCarriageReturn    = ErrorString("expected carriage return")
	ErrExpectedNewline           = ErrorString("expected newline")
	ErrExpectedSpace             = ErrorString("expected space")
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
