package main

type parser struct {
	requestURI  int
	protocol    int
	headers     int
	headerCount int
	lineStart   int
	transform   func([]byte, int) int
}

// fn returns
// return values as follows
// - next
// the next parsing routine to call either a new state or a resumption of previous that run out of bytes
// - pos
// the current position
// - adv
// the current state has requested buf be expanded to atleast adv bytes
// - err
// fatal error occurred
type fn func(*parser, []byte, int) (next fn, pos int, adv int, err error)

func (p *parser) parseMethod(buf []byte, pos int) (fn, int, int, error) {
	for pos < len(buf) && isToken(buf[pos]) {
		pos++
	}
	if pos >= len(buf) {
		return (*parser).parseMethod, pos, pos, nil
	}
	if buf[pos] != ' ' {
		return nil, pos, 0, ErrExpectedSpace
	}
	pos++
	return (*parser).parseRequestURI, pos, pos, nil
}

func (p *parser) parseRequestURI(buf []byte, pos int) (fn, int, int, error) {
	if !isFieldVChar(buf[pos]) {
		return nil, pos, 0, ErrMissingRequestURI
	}
	p.requestURI = pos
	pos++
	return (*parser).parseRequestURI2, pos, pos, nil
}

func (p *parser) parseRequestURI2(buf []byte, pos int) (fn, int, int, error) {
	for pos < len(buf) && isFieldVChar(buf[pos]) {
		pos++
	}
	if pos >= len(buf) {
		return (*parser).parseRequestURI2, pos, pos, nil
	}
	if buf[pos] != ' ' {
		return nil, pos, 0, ErrExpectedSpace
	}
	pos++
	return (*parser).parseProtocol, pos, pos, nil
}

func (p *parser) parseProtocol(buf []byte, pos int) (fn, int, int, error) {
	const prefix = "HTTP/"
	const pattern = prefix + "0.0\r\n"

	if adv := pos + len(pattern); adv > len(buf) {
		return (*parser).parseProtocol, pos, adv, nil
	}
	if string(buf[pos:pos+len(prefix)]) != prefix {
		return nil, pos, 0, ErrUnknownProtocol
	}
	if !isDigit(buf[pos+len(prefix)]) ||
		buf[pos+len(prefix+"0")] != '.' ||
		!isDigit(buf[pos+len(prefix+"0.")]) {
		return nil, pos, 0, ErrUnknownProtocol
	}
	p.protocol = pos
	if buf[pos+len(prefix+"0.0")] != '\r' {
		return nil, pos, 0, ErrExpectedCarriageReturn
	}
	if buf[pos+len(prefix+"0.0\r")] != '\n' {
		return nil, pos, 0, ErrExpectedNewline
	}
	pos += len(pattern)
	p.headers = pos
	return (*parser).newline, pos, pos, nil
}

func (p *parser) newline(buf []byte, pos int) (fn, int, int, error) {
	if isToken(buf[pos]) {
		p.headerCount++
		p.lineStart = pos
		// First letter of header key should be upper case
		if isLower(buf[pos]) {
			buf[pos] ^= 0x20
		}
		pos++
		return (*parser).headerName, pos, pos, nil
	}
	if buf[pos] != '\r' {
		return nil, pos, 0, ErrExpectedCarriageReturn
	}
	pos++
	if pos >= len(buf) {
		// "unread" '\r' so can resume at this state
		return (*parser).newline, pos - len("\r"), pos, nil
	}
	if buf[pos] != '\n' {
		return nil, pos, 0, ErrExpectedNewline
	}
	return nil, pos + 1, 0, nil // Seen final \r\n\r\n
}

func none(buf []byte, pos int) int {
	for pos < len(buf) && isFieldVChar(buf[pos]) {
		pos++
	}
	return pos
}

func lower(buf []byte, pos int) int {
	for ; pos < len(buf) && isFieldVChar(buf[pos]); pos++ {
		if isUpper(buf[pos]) {
			buf[pos] += 'a' - 'A'
		}
	}
	return pos
}

func upper(buf []byte, pos int) int {
	for ; pos < len(buf) && isFieldVChar(buf[pos]); pos++ {
		if isLower(buf[pos]) {
			buf[pos] -= 'a' - 'A'
		}
	}
	return pos
}

func transform(b []byte) func([]byte, int) int {
	switch string(b) {
	case "Connection":
		return lower
	}
	return none
}

func (p *parser) headerName(buf []byte, pos int) (fn, int, int, error) {
	if !isToken(buf[pos]) {
		return nil, pos, 0, ErrMissingHeaderName
	}
	nextA := 'A'
	if pos <= 0 || buf[pos-1] == '-' {
		nextA = 'a'
	}
	for ; pos < len(buf) && isToken(buf[pos]); pos++ {
		if buf[pos]-byte(nextA) < 26 {
			buf[pos] ^= 0x20 // buf[pos] wrong case, toggle
		}
		nextA = 'A'
		if buf[pos] == '-' {
			nextA = 'a'
		}
	}
	if pos >= len(buf) {
		return (*parser).headerName, pos, pos, nil
	}
	if buf[pos] != ':' {
		return nil, pos, 0, ErrExpectedColon
	}
	p.transform = transform(buf[p.lineStart:pos])
	pos++
	return (*parser).ows, pos, pos, nil
}

func (p *parser) ows(buf []byte, pos int) (fn, int, int, error) {
	for pos < len(buf) && isHorizontalSpace(buf[pos]) {
		pos++
	}
	if pos >= len(buf) {
		return (*parser).ows, pos, pos, nil
	}
	return (*parser).value, pos, pos, nil
}

func (p *parser) value(buf []byte, pos int) (fn, int, int, error) {
	if !isFieldVChar(buf[pos]) {
		return nil, pos, 0, ErrMissingHeaderValue
	}
	pos = p.transform(buf, pos)
	if pos >= len(buf) {
		return (*parser).value, pos, pos, nil
	}
	return (*parser).ows1, pos, pos, nil
}

func (p *parser) ows1(buf []byte, pos int) (fn, int, int, error) {
	for pos < len(buf) && isHorizontalSpace(buf[pos]) {
		pos++
	}
	if adv := pos + 1; adv >= len(buf) {
		return (*parser).ows1, pos, adv, nil
	}
	if buf[pos] != '\r' {
		return nil, pos, 0, ErrExpectedCarriageReturn
	}
	pos++
	if buf[pos] == '\n' {
		pos++
		return (*parser).newline, pos, pos, nil
	}
	return nil, pos, 0, ErrExpectedNewline
}
