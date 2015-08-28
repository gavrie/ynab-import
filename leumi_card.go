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
	Rows    []Row    `xml:"Worksheet>Table>Row"`
}

type Row struct {
	Cells []Cell `xml:"Cell"`
}

type Cell struct {
	Data string
}

func getCellNames(row Row) map[string]int {
	cellsByName := map[string]int{}

	for i, cell := range row.Cells {
		cellsByName[cell.Data] = i
	}
	return cellsByName
}

////////////////////////////////////////////////////////////////////////////////

type transaction struct {
	date     time.Time
	payee    string
	category string
	memo     string
	outflow  string
	inflow   string
}

func newTransaction(cellsByName map[string]int, row *Row) *transaction {
	if len(row.Cells) != len(cellsByName) {
		log.Fatal("Unexpected row length")
	}

	getCell := func(name string) string {
		i, ok := cellsByName[name]
		if !ok {
			log.Fatal("No cell found with name '%v'", name)
		}
		return row.Cells[i].Data
	}

	stringAmount := getCell("סכום חיוב ₪")
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

	stringDate := getCell("תאריך עסקה")
	date, err := time.Parse("2006-01-02T15:04:05", stringDate)
	if err != nil {
		log.Fatal("Invalid date encountered", stringDate)
	}

	return &transaction{
		date:    date,
		payee:   getCell("שם בית העסק"),
		memo:    getCell("הערות"),
		outflow: outflow,
		inflow:  inflow,
	}
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

func process() error {
	rows, err := decodeRows()
	if err != nil {
		return err
	}

	if len(rows) < 2 {
		return errors.New("No transactions")
	}

	header, rows := rows[0], rows[1:]

	cellsByName := getCellNames(header)

	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"Date", "Payee", "Category", "Memo", "Outflow", "Inflow"})

	for _, row := range rows {
		t := newTransaction(cellsByName, &row)

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

func main() {
	err := process()

	if err != nil {
		log.Fatal(err)
	}
}
