package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf16"

	"github.com/PuerkitoBio/goquery"
)

/*

 */

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

func decodeRowsHtml(r io.Reader) (rows []Row, err error) {
	s, err := readUtf16(r)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		return nil, err
	}

	doc.Find("#tdCalGrid").Each(func(i int, s *goquery.Selection) {
		s.Find("thead").Each(func(i int, s *goquery.Selection) {
			s.Find("tr").Each(func(i int, s *goquery.Selection) {
				var row Row
				s.Find("th").Each(func(i int, s *goquery.Selection) {
					value := s.Text()
					if value == "" {
						value = fmt.Sprintf("field_%v", i)
					}
					// log.Printf("Field: %v [%v]", value, i)
					row = append(row, value)
				})
				// log.Print(row)
				rows = append(rows, row)
			})
		})

		s.Find("tbody").Each(func(i int, s *goquery.Selection) {
			done := false

			s.Find("tr").Each(func(i int, s *goquery.Selection) {
				var row Row
				s.Find("td").Each(func(i int, s *goquery.Selection) {
					if s.HasClass("footer_cell_total") {
						done = true
					}
					value := strings.TrimSpace(s.Text())
					row = append(row, value)
				})
				if !done {
					// log.Print(row)
					rows = append(rows, row)
				}
			})
		})
	})

	return rows, nil
}
