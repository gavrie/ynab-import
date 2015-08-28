package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
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

func decodeRowsHtml() error {
	s, err := readUtf16(os.Stdin)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		return err
	}

	doc.Find("#tdCalGrid").Each(func(i int, s *goquery.Selection) {
		s.Find("tr").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				log.Println("Header:", i, s)
				s.Find("th").Each(func(i int, s *goquery.Selection) {
					log.Printf("Field: %v [%v]", s.Text(), i)
				})
				return
			}
			log.Println("Row:", i, s)
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				log.Printf("Field: %v [%v]", s.Text(), i)
			})
		})
	})

	return nil
}
