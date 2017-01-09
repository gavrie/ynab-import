package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func decodeRowsLeumicard(r io.Reader) (rows []Row, err error) {
	log.Print("leumicard")

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	doc.Find("#ctlRegularTransactions").Each(func(i int, s *goquery.Selection) {
		s.Find("tr.header").Each(func(i int, s *goquery.Selection) {
			var row Row
			s.Find("th").Each(func(i int, s *goquery.Selection) {
				value := strings.TrimSpace(s.Text())
				if value == "" {
					value = fmt.Sprintf("field_%v", i)
				}
				log.Printf("Field: '%v' [%v]", value, i)
				row = append(row, value)
			})
			log.Print(row)
			rows = append(rows, row)
		})

		s.Find("tr.alternatingItem").Each(func(i int, s *goquery.Selection) {
			done := false

			var row Row
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				if s.HasClass("footer_cell_total") {
					done = true
				}
				value := strings.TrimSpace(s.Text())
				row = append(row, value)
			})
			if !done && row != nil {
				log.Printf("%#v", row)
				rows = append(rows, row)
			}
		})
	})

	return rows, nil
}
