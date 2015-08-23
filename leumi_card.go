package main

import (
	"encoding/csv"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

/*
<Row>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">תאריך עסקה</Data> 0
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">תאריך חיוב</Data> 1
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">שם בית העסק</Data> 2
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">סוג עסקה</Data> 3
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">מטבע עסקה</Data> 4
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">סכום עסקה</Data> 5
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">סכום חיוב ₪</Data> 6
	</Cell>
	<Cell ss:StyleID="Header">
	<Data ss:Type="String">הערות</Data> 7
	</Cell>
</Row>
*/

type transaction struct {
	date     time.Time
	payee    string
	category string
	memo     string
	outflow  string
	inflow   string
}

func newTransaction(row *Row) *transaction {
	if len(row.Cells) != 8 {
		log.Fatal("Unexpected row length")
	}

	stringAmount := row.Cells[6].Data
	amount, err := strconv.ParseFloat(stringAmount, 64)
	if err != nil {
		log.Fatal("Non-numeric value encountered", stringAmount)
	}

	var outflow, inflow string
	if amount > 0 {
		outflow = stringAmount
	} else {
		inflow = stringAmount
	}

	stringDate := row.Cells[0].Data
	date, err := time.Parse("2006-01-02T15:04:05", stringDate)
	if err != nil {
		log.Fatal("Invalid date encountered", stringDate)
	}

	return &transaction{
		date:    date,
		payee:   row.Cells[2].Data,
		memo:    row.Cells[7].Data,
		outflow: outflow,
		inflow:  inflow,
	}
}

/*
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
	Rows    []Row    `xml:"Worksheet>Table>Row"`
}

type Row struct {
	Cells []Cell `xml:"Cell"`
}

type Cell struct {
	Data string
}

func main() {
	err := process()

	if err != nil {
		log.Fatal(err)
	}
}

func process() error {
	rows, err := decodeRows()
	if err != nil {
		return err
	}

	if len(rows) < 2 {
		return errors.New("No transactions")
	}

	rows = rows[1:] // Skip header

	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"Date", "Payee", "Category", "Memo", "Outflow", "Inflow"})

	for _, row := range rows {
		t := newTransaction(&row)

		w.Write([]string{
			t.date.Format("02/01/2006"),
			t.payee,
			t.category,
			t.memo,
			t.outflow,
			t.inflow,
		})
	}

	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func decodeRows() ([]Row, error) {
	d := xml.NewDecoder(os.Stdin)
	doc := &Document{}

	for {
		err := d.Decode(doc)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
	}

	return doc.Rows, nil
}
