package main

import "math"

func parseFirstLine(buf []byte) (pos int, adv int, err error) {
	for pos < len(buf) && isToken(buf[pos]) {
		pos++
	}
	if pos >= len(buf) {
		return pos, pos, nil
	}
	if pos >= math.MaxInt-len(" / HTTP/0.0\r\n") {
		return pos, 0, ErrHeaderTooLarge
	}
	if adv := pos + len(" / HTTP/0.0\r\n"); adv >= len(buf) {
		return pos, adv, nil
	}
	if buf[pos] != ' ' {
		return pos, 0, ErrExpected(' ')
	}
	pos++
	if !isFieldVChar(buf[pos]) {
		return pos, 0, ErrMissingRequestURI
	}
	pos++
	for pos < len(buf) && isFieldVChar(buf[pos]) {
		pos++
	}
	if pos >= math.MaxInt-len(" HTTP/0.0\r\n") {
		return pos, 0, ErrHeaderTooLarge
	}
	if adv := pos + len(" HTTP/0.0\r\n"); adv >= len(buf) {
		return 0, adv, nil
	}
	// Space between RequestURI and Protocol
	if buf[pos] != ' ' {
		return pos, 0, ErrExpected(' ')
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
		return pos - len("HTTP/"), 0, ErrUnknownProtocol
	}
	pos += len("0.0")
	if buf[pos] != '\r' {
		return pos, 0, ErrExpected('\r')
	}
	pos++
	if buf[pos] != '\n' {
		return pos, 0, ErrExpected('\n')
	}
	pos++
	return pos, pos, nil
}

func parseBlock(buf []byte, pos int) (_ int, adv int, headerCount int, err error) {
	for pos < len(buf) {
		lineStart := pos
		switch {
		case isToken(buf[pos]):
			pos = key(buf, pos)
			if pos >= len(buf) {
				return lineStart, pos, headerCount, nil
			}
			// Colon
			if buf[pos] != ':' {
				return pos, 0, 0, ErrExpected(':')
			}
			key := buf[lineStart:pos]
			pos = skipOptionalSpace(buf, pos+1)
			if pos >= len(buf) {
				return lineStart, pos, headerCount, nil
			}
			// Header value
			if !isFieldVChar(buf[pos]) {
				return pos, 0, 0, ErrMissingHeaderValue
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
			pos = skipOptionalSpace(buf, pos)
			if pos >= len(buf) {
				return lineStart, pos, headerCount, nil
			}
			if buf[pos] != '\r' {
				return pos, 0, 0, ErrExpected('\r')
			}
			pos++
			if pos >= len(buf) {
				return lineStart, pos, headerCount, nil
			}
			if buf[pos] != '\n' {
				return pos, 0, 0, ErrExpected('\n')
			}
			pos++
			headerCount++

		case buf[pos] == '\r':
			pos++
			if pos >= len(buf) {
				// "unread" '\r' so can resume at this state
				return pos - len("\r"), pos, headerCount, nil
			}
			if buf[pos] != '\n' {
				return pos, 0, 0, ErrExpected('\n')
			}
			return pos + 1, 0, headerCount, EOH // Seen final \r\n\r\n

		default:
			return pos, 0, 0, ErrExpected('\r')
		}
	}
	return pos, pos, headerCount, nil
}
