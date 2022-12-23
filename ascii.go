package main

import (
	"runtime"
	"strings"
)

func isIn(c byte, lo, hi uint64) bool {
	switch runtime.GOARCH {
	case "amd64", "arm64", "arm64be", "ppc64", "ppc64le":
		// 64 bit and conditional moves available...
		mask := lo
		if c >= 64 {
			mask = hi
			if c >= 128 {
				mask = 0
			}
		}
		return 1<<(c%64)&mask != 0
	}
	return (1<<c&lo)|(1<<(c-64)&hi) != 0
}

func isToken(c byte) bool {
	const (
		upper  = ((1 << 26) - 1) << 'A'
		lower  = ((1 << 26) - 1) << 'a'
		digits = ((1 << 10) - 1) << '0'
		tokens = upper | lower | digits |
			1<<'!' | 1<<'#' | 1<<'$' | 1<<'%' | 1<<'&' | 1<<'\'' | 1<<'*' | 1<<'+' |
			1<<'-' | 1<<'.' | 1<<'^' | 1<<'_' | 1<<'`' | 1<<'|' | 1<<'~'
	)

	return isIn(c, tokens%(1<<64), tokens>>64)
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
			buf[pos] = c ^ 0x20 // buf[pos] wrong case, toggle
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

func cut(s string, c byte) (string, string, bool) {
	i := strings.IndexByte(s, c)
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+1:], true
}
