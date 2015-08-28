package main

import (
	"encoding/xml"
	"io"
	"log"
)

/*
<Row>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">תאריך עסקה</Data>
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">תאריך חיוב</Data>
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">שם בית העסק</Data>
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">סוג עסקה</Data>
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">מטבע עסקה</Data>
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">סכום עסקה</Data>
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">סכום חיוב ₪</Data>
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">הערות</Data>
	</Cell>
</Row>

<Row>
	<Cell>
	<Data ss:Type="DateTime">2014-09-30T00:00:00</Data>
	</Cell>
	<Cell>
	<Data ss:Type="DateTime">2015-08-16T00:00:00</Data>
	</Cell>
	<Cell>
	<Data ss:Type="String">כללית סמייל בית שמש</Data>
	</Cell>
	<Cell>
	<Data ss:Type="String">תשלומים</Data>
	</Cell>
	<Cell>
	<Data ss:Type="String">₪</Data>
	</Cell>
	<Cell>
	<Data ss:Type="String">3229.00</Data>
	</Cell>
	<Cell>
	<Data ss:Type="Number">215</Data>
	</Cell>
	<Cell>
	<Data ss:Type="String">תשלום 11 מתוך 15</Data>
	</Cell>
</Row>
*/

type Document struct {
	XMLName xml.Name `xml:"Workbook"`
	Rows    []xmlRow `xml:"Worksheet>Table>Row"`
}

type xmlRow struct {
	Cells []xmlCell `xml:"Cell"`
}

type xmlCell struct {
	Data string
}

func decodeRowsXml(r io.Reader) (rows []Row, err error) {
	log.Printf("xml")

	d := xml.NewDecoder(r)
	doc := &Document{}

	for {
		err = d.Decode(doc)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
	}

	for _, xmlRow := range doc.Rows {
		var row Row
		for _, cell := range xmlRow.Cells {
			row = append(row, cell.Data)
		}
		rows = append(rows, row)
	}

	return rows, nil
}
