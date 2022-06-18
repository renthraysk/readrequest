package main

import "io"

type parser struct {
	method      int
	requestURI  int
	proto       int
	headerCount int
}

func (p *parser) parseMethod(buf []byte, pos int) (int, int, error) {
	for pos < len(buf) && isToken(buf[pos]) {
		pos++
	}
	p.method = pos
	if adv := pos + len(" / HTTP/0.0\r\n"); adv >= len(buf) {
		return pos, adv, nil
	}
	if buf[pos] != ' ' {
		return pos, 0, ErrExpectedSpace
	}
	pos++
	if !isFieldVChar(buf[pos]) {
		return pos, 0, ErrMissingRequestURI
	}
	for pos < len(buf) && isFieldVChar(buf[pos]) {
		pos++
	}
	p.requestURI = pos
	if adv := pos + len(" HTTP/0.0\r\n"); adv >= len(buf) {
		return 0, adv, nil
	}
	// Space between RequestURI and Protocol
	if buf[pos] != ' ' {
		return pos, 0, ErrExpectedSpace
	}
	pos++
	// Protocol
	if string(buf[pos:pos+len("HTTP/")]) != "HTTP/" {
		return pos, 0, ErrUnknownProtocol
	}
	pos += len("HTTP/")
	if !isDigit(buf[pos]) ||
		buf[pos+len("0")] != '.' ||
		!isDigit(buf[pos+len("0.")]) {
		return pos, 0, ErrUnknownProtocol
	}
	pos += len("0.0")
	p.proto = pos
	if buf[pos] != '\r' {
		return pos, 0, ErrExpectedCarriageReturn
	}
	pos++
	if buf[pos] != '\n' {
		return pos, 0, ErrExpectedNewline
	}
	pos++
	return pos, pos, nil
}

func (p *parser) newline(buf []byte, pos int) (int, int, error) {
	switch {
	case isToken(buf[pos]):
		return p.header(buf, pos)

	case buf[pos] == '\r':
		pos++
		if pos >= len(buf) {
			// "unread" '\r' so can resume at this state
			return pos - len("\r"), pos, nil
		}
		if buf[pos] != '\n' {
			return pos, 0, ErrExpectedNewline
		}
		return pos + 1, 0, io.EOF // Seen final \r\n\r\n

	default:
		return pos, 0, ErrExpectedCarriageReturn
	}
}

func key(buf []byte, pos int) int {
	for nextA := 'a'; pos < len(buf) && isToken(buf[pos]); pos++ {
		if buf[pos]-byte(nextA) < 26 {
			buf[pos] ^= 0x20 // buf[pos] wrong case, toggle
		}
		nextA = 'A'
		if buf[pos] == '-' {
			nextA = 'a'
		}
	}
	return pos
}

func skipOptionalSpace(buf []byte, pos int) int {
	for pos < len(buf) && isHorizontalSpace(buf[pos]) {
		pos++
	}
	return pos
}

func (p *parser) header(buf []byte, pos int) (int, int, error) {
	lineStart := pos
	pos = key(buf, pos)
	if pos >= len(buf) {
		return lineStart, pos, nil
	}
	// Colon
	if buf[pos] != ':' {
		return 0, 0, ErrExpectedColon
	}
	key := buf[lineStart:pos]
	pos = skipOptionalSpace(buf, pos+1)
	if pos >= len(buf) {
		return lineStart, pos, nil
	}
	// Header value
	if !isFieldVChar(buf[pos]) {
		return pos, 0, ErrMissingHeaderValue
	}
	switch string(key) {
	case "Connection", "Transfer-Encoding":
		// Lower case value
		for ; pos < len(buf) && isFieldVChar(buf[pos]); pos++ {
			if isUpper(buf[pos]) {
				buf[pos] += 'a' - 'A'
			}
		}
	default:
		for pos < len(buf) && isFieldVChar(buf[pos]) {
			pos++
		}
	}
	if pos >= len(buf) {
		return lineStart, pos, nil
	}
	pos = skipOptionalSpace(buf, pos)
	if pos >= len(buf) {
		return lineStart, pos, nil
	}
	if buf[pos] != '\r' {
		return pos, 0, ErrExpectedCarriageReturn
	}
	pos++
	if pos >= len(buf) {
		return lineStart, pos, nil
	}
	if buf[pos] != '\n' {
		return pos, 0, ErrExpectedNewline
	}
	p.headerCount++
	pos++
	return pos, pos, nil
}
