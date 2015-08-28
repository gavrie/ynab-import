package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func decodeRowsHtmlCal(r io.Reader) (rows []Row, err error) {
	log.Print("cal")

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
