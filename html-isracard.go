package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func decodeRowsHtmlIsracard(r io.Reader) (rows []Row, err error) {
	log.Print("isracard")

	rInUTF8 := transform.NewReader(r, charmap.ISO8859_8I.NewDecoder())

	doc, err := goquery.NewDocumentFromReader(rInUTF8)
	if err != nil {
		return nil, err
	}

	done := false

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		if i < 3 || i == 4 {
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

		if len(row) < 2 {
			if !done {
				log.Println("Heuristic: wrong file format, giving up")
				rows = nil
				done = true
			}
			return
		}

		if row[0] == "" {
			row[0] = "1970-01-01"
		}

		if row[1] == "@סך חיוב בש\"ח:" {
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
