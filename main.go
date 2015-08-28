package main

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

var fieldNames = map[string][]string{
	"amount": {"סכום חיוב ₪", "סכוםהחיוב"},
	"date":   {"תאריך עסקה", "תאריךהעסקה"},
	"payee":  {"שם בית העסק", "שםבית העסק"},
	"memo":   {"הערות", "פירוט נוסף"},
}

var dateFormats = []string{
	"2006-01-02T15:04:05",
	"02/01/06",
}

func parseDate(s string) (date time.Time, err error) {
	for _, f := range dateFormats {
		date, err = time.Parse(f, s)
		if err == nil {
			return date, nil
			break
		}
	}
	return date, err
}

type Row []string

////////////////////////////////////////////////////////////////////////////////

var cellIndexByName map[string]int

func newCellIndexByName(row Row) map[string]int {
	cellIndexByName := map[string]int{}

	for i, cell := range row {
		cellIndexByName[cell] = i
	}
	return cellIndexByName
}

func getCell(row Row, field string) string {
	names, ok := fieldNames[field]
	if !ok {
		log.Fatal("No cell names found for field '%v'", field)
	}

	for _, name := range names {
		i, ok := cellIndexByName[name]
		if ok {
			return row[i]
		}
	}
	log.Fatal("No cell found matching field '%v'", field)
	return "<invalid>"
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

func newTransaction(row Row) *transaction {
	// log.Printf("cellIndexByName: %#v", cellIndexByName)

	if len(row) != len(cellIndexByName) {
		log.Fatal("Unexpected row length")
	}

	stringAmount := getCell(row, "amount")
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

	stringDate := getCell(row, "date")
	date, err := parseDate(stringDate)
	if err != nil {
		log.Fatal("Invalid date encountered", stringDate)
	}

	return &transaction{
		date:    date,
		payee:   getCell(row, "payee"),
		memo:    getCell(row, "memo"),
		outflow: outflow,
		inflow:  inflow,
	}
}

func processRows(rows []Row) error {
	if len(rows) < 2 {
		return errors.New("No transactions")
	}

	header, rows := rows[0], rows[1:]

	cellIndexByName = newCellIndexByName(header)

	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"Date", "Payee", "Category", "Memo", "Outflow", "Inflow"})

	for _, row := range rows {
		t := newTransaction(row)

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

func decodeFile(filename string) (rows []Row, err error) {
	decoders := []func(io.Reader) ([]Row, error){
		decodeRowsXml,
		decodeRowsHtml,
	}

	for _, decode := range decoders {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		rows, err = decode(f)
		if err == nil {
			return rows, nil
		}
	}

	return nil, err

}

func decodeAll() error {
	files := []string{
		"6025.xml",
		"4105.html",
	}

	for _, filename := range files {
		rows, err := decodeFile(filename)
		if err != nil {
			return err
		}

		err = processRows(rows)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	err := decodeAll()
	if err != nil {
		log.Fatal(err)
	}
}
