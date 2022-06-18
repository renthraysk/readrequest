package main

func isToken(c byte) bool {
	const mask = 0 |
		((1<<10)-1)<<'0' |
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

	m := uint64(mask & ((1 << 64) - 1))
	if c >= 64 {
		m = mask >> 64
		if mask >= 1<<64 && c >= 128 {
			m = mask >> 128
			if mask >= 1<<128 && c >= 192 {
				m = mask >> 192
			}
		}
	}
	return (1<<(c%64))&m != 0
}

func isLower(c byte) bool           { return c-'a' <= 'z'-'a' }
func isUpper(c byte) bool           { return c-'A' <= 'Z'-'A' }
func isDigit(c byte) bool           { return c-'0' <= 9 }
func isFieldVChar(c byte) bool      { return c > ' ' && c != 0x7F }
func isHorizontalSpace(c byte) bool { return c == ' ' || c == '\t' }

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
