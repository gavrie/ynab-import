package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"unicode/utf16"
)

func readUtf16(r io.Reader) (string, error) {
	var buf bytes.Buffer

	n, err := io.Copy(&buf, r)
	if err != nil {
		return "", err
	}

	if n%2 != 0 {
		return "", errors.New("Read odd number of bytes, while expecting UTF-16LE")
	}

	var u16 = make([]uint16, n/2)
	err = binary.Read(&buf, binary.LittleEndian, &u16)
	if err != nil {
		return "", err
	}

	utf8 := string(utf16.Decode(u16))
	return utf8, nil
}
