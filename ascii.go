package main

import "runtime"

type set256 [256 / 32]uint32

func (s *set256) Contains(c byte) bool { return (1<<(c%32))&tokenSet[c/32] != 0 }

const tokenMask = ((1<<10)-1)<<'0' |
	((1<<26)-1)<<'a' |
	((1<<26)-1)<<'A' |
	1<<'!' |
	1<<'#' |
	1<<'$' |
	1<<'%' |
	1<<'&' |
	1<<'\'' |
	1<<'*' |
	1<<'+' |
	1<<'-' |
	1<<'.' |
	1<<'^' |
	1<<'_' |
	1<<'`' |
	1<<'|' |
	1<<'~'

var tokenSet = set256{
	(tokenMask >> (0 * 32)) % (1 << 32),
	(tokenMask >> (1 * 32)) % (1 << 32),
	(tokenMask >> (2 * 32)) % (1 << 32),
	(tokenMask >> (3 * 32)) % (1 << 32),
	(tokenMask >> (4 * 32)) % (1 << 32),
	(tokenMask >> (5 * 32)) % (1 << 32),
	(tokenMask >> (6 * 32)) % (1 << 32),
	(tokenMask >> (7 * 32)) % (1 << 32),
}

func isToken(c byte) bool {
	if tokenMask < (1 << 128) {
		// @TODO Worth it?

		// If arch & compiler supports conditional moves...
		switch runtime.GOARCH {
		case "amd64", "arm64":
			m := uint64(tokenMask % (1 << 64))
			if c >= 64 {
				m = tokenMask >> 64
				if tokenMask >= (1<<64) && c >= 128 {
					m = 0
				}
			}
			return (1<<(c%64))&m != 0
		}
	}
	return tokenSet.Contains(c)
}

func isLower(c byte) bool           { return c-'a' <= 'z'-'a' }
func isUpper(c byte) bool           { return c-'A' <= 'Z'-'A' }
func isDigit(c byte) bool           { return c-'0' <= 9 }
func isFieldVChar(c byte) bool      { return c > ' ' && c != 0x7F }
func isHorizontalSpace(c byte) bool { return c == ' ' || c == '\t' }

func key(buf []byte, pos int) int {
	for nextA := 'a'; pos < len(buf) && isToken(buf[pos]); pos++ {
		c := buf[pos]
		if c-byte(nextA) < 26 {
			buf[pos] ^= 0x20 // buf[pos] wrong case, toggle
		}
		nextA = 'A'
		if c == '-' {
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

func trim(s string) string {
	i := 0
	for i < len(s) && isHorizontalSpace(s[i]) {
		i++
	}
	n := len(s)
	for n > i && isHorizontalSpace(s[n-1]) {
		n--
	}
	return s[i:n]
}
