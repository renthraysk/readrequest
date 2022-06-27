package main

import (
	"bufio"
	"strings"
	"testing"
)

func FuzzReadRequest(f *testing.F) {
	f.Add(quickTest, 8<<10)

	f.Fuzz(func(t *testing.T, in string, size int) {
		_, err := readRequest(bufio.NewReader(strings.NewReader(in)), size)

		headerSize := strings.Index(in, "\r\n\r\n")
		if err == nil {
			if headerSize < 0 {
				t.Error("no end of header marker, got no error")
			} else if headerSize+len("\r\n\r\n") > size {
				t.Errorf("header size exceeded %d, expected error", size)
			}
		}
	})
}
