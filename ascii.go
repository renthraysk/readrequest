package main

import "runtime"

type set256 [256 / 32]uint32

func (s *set256) Contains(c byte) bool { return (1<<(c%32))&s[c/32] != 0 }

const (
	upper  = ((1 << 26) - 1) << 'A'
	lower  = ((1 << 26) - 1) << 'a'
	digits = ((1 << 10) - 1) << '0'
	tokens = upper | lower | digits |
		1<<'!' | 1<<'#' | 1<<'$' | 1<<'%' | 1<<'&' | 1<<'\'' | 1<<'*' | 1<<'+' |
		1<<'-' | 1<<'.' | 1<<'^' | 1<<'_' | 1<<'`' | 1<<'|' | 1<<'~'
)

var tokenSet = set256{
	(tokens >> (0 * 32)) % (1 << 32),
	(tokens >> (1 * 32)) % (1 << 32),
	(tokens >> (2 * 32)) % (1 << 32),
	(tokens >> (3 * 32)) % (1 << 32),
	(tokens >> (4 * 32)) % (1 << 32),
	(tokens >> (5 * 32)) % (1 << 32),
	(tokens >> (6 * 32)) % (1 << 32),
	(tokens >> (7 * 32)) % (1 << 32),
}

func mask(c byte, lo, hi uint64) uint64 {
	if c >= 64 {
		lo = hi
		if hi != 0 && c >= 128 {
			lo = 0
		}
	}
	return lo
}

func isToken(c byte) bool {
	if false && tokens < (1<<128) {
		switch runtime.GOARCH {
		case "amd64", "arm64":
			// 64-bit and conditional movs available...
			return (1<<(c%64))&mask(c, tokens%(1<<64), tokens>>64) != 0
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
