package main

func isToken(c byte) bool           { return isTokenTable[c] != 0 }
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

const isTokenTable = "" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x01\x00\x01\x01\x01\x01\x01\x00\x00\x01\x01\x00\x01\x01\x00" +
	"\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x00\x00\x00\x00\x00\x00" +
	"\x00\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01" +
	"\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x00\x00\x00\x01\x01" +
	"\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01" +
	"\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x00\x01\x00\x01\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" +
	"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
