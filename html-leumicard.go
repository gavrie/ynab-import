package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getHeaders(id string, doc *goquery.Document) (headers Row, err error) {
	doc.Find(id).Each(func(i int, s *goquery.Selection) {
		s.Find("tr.header").Each(func(i int, s *goquery.Selection) {
			s.Find("th").Each(func(i int, s *goquery.Selection) {
				value := strings.TrimSpace(s.Text())
				if value == "" {
					value = fmt.Sprintf("field_%v", i)
				}
				//log.Printf("Field: '%v' [%v]", value, i)
				headers = append(headers, value)
			})
		})
	})
	//log.Printf("Header: %v", headers)
	return headers, err

}

func getRows(id string, doc *goquery.Document) (rows []Row, err error) {
	doc.Find(id).Each(func(i int, s *goquery.Selection) {
		s.Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
			if !s.HasClass("footer") {
				var row Row
				s.Find("td").Each(func(i int, s *goquery.Selection) {
					value := strings.TrimSpace(s.Text())
					row = append(row, value)
				})
				if row != nil {
					//log.Printf("%#v", row)
					rows = append(rows, row)
				}
			}
		})
	})
	return rows, err
}

func decodeRowsLeumicard(r io.Reader) (rows []Row, err error) {
	log.Print("leumicard")

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	// ctlRegularTransactionsDebit
	header, err := getHeaders("#ctlRegularTransactions", doc)
	rows1, err := getRows("#ctlRegularTransactions", doc)
	//rows2, err := getRows("#ctlRegularTransactionsDebit", doc)
	//rowsAll := append(rows1, rows2...)

	rows = append([]Row{header}, rows1...)
	return rows, nil
}
