package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func decodeRowsHtmlMizrahiCC(r io.Reader) (rows []Row, err error) {
	log.Print("mizrahi-cc")

	s, err := readUtf16(r)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		return nil, err
	}

	done := false

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		if i < 7 {
			return
		}

		isHeader := (len(rows) == 0)

		var row Row
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			value := strings.TrimSpace(s.Text())
			value = strings.TrimPrefix(value, "\u200f")

			if isHeader {
				if value == "" {
					value = fmt.Sprintf("field_%v", i)
				}
			}
			// log.Printf("Field: %v [%v]", value, i)

			row = append(row, value)
		})

		if row[0] == "" {
			row[0] = "1970-01-01"
		}

		if row[0] == "כרטיס:" {
			done = true
		}

		if row[3] == "לידיעה בלבד" {
			return
		}

		if !done {
			log.Printf("%#v", row)
			rows = append(rows, row)
		}
	})

	if len(rows) > 0 && rows[0][0] == "תנועות אחרונות" {
		return nil, nil // Not a CC file but probably a checking file
	}

	return rows, nil
}

func decodeRowsHtmlMizrahiChecking(r io.Reader) (rows []Row, err error) {
	log.Print("mizrahi-checking")

	s, err := readUtf16(r)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		return nil, err
	}

	done := false

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		if i < 9 {
			return
		}

		isHeader := (len(rows) == 0)

		var row Row
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			value := strings.TrimSpace(s.Text())
			value = strings.TrimPrefix(value, "\u200f")

			if isHeader {
				if value == "" {
					value = fmt.Sprintf("field_%v", i)
				}
			}
			// log.Printf("Field: %v [%v]", value, i)
			row = append(row, value)
		})

		if row[0] == "" {
			done = true
		}

		if !done {
			log.Printf("%#v", row)
			rows = append(rows, row)
		}
	})

	return rows, nil
}
