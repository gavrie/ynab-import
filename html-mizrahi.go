package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func decodeRowsHtmlMizrahi(r io.Reader) (rows []Row, err error) {
	log.Print("mizrahi")

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

		row[1] = strings.TrimPrefix(row[1], "\u200f")

		if !done {
			log.Printf("%#v", row)
			rows = append(rows, row)
		}
	})

	return rows, nil
}
