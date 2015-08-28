package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var fieldNames = map[string][]string{
	"amount": {"סכום חיוב ₪", "סכוםהחיוב", "סכום החיוב בש''ח"},
	"date":   {"תאריך עסקה", "תאריךהעסקה", "תאריך הקנייה"},
	"payee":  {"שם בית העסק", "שםבית העסק"},
	"memo":   {"הערות", "פירוט נוסף", "מידע נוסף"},
}

var dateFormats = []string{
	"2006-01-02T15:04:05",
	"02/01/06",
	"2006-01-02",
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
	stringAmount = strings.Replace(stringAmount, ",", "", -1)
	amount, err := strconv.ParseFloat(stringAmount, 64)
	if err != nil {
		log.Fatal("Non-numeric value encountered ", stringAmount)
	}

	var outflow, inflow string
	if amount > 0 {
		outflow = fmt.Sprintf("%v", amount)
	} else {
		inflow = fmt.Sprintf("%v", -amount)
	}

	stringDate := getCell(row, "date")
	date, err := parseDate(stringDate)
	if err != nil {
		log.Fatal("Invalid date encountered ", stringDate)
	}

	return &transaction{
		date:    date,
		payee:   getCell(row, "payee"),
		memo:    getCell(row, "memo"),
		outflow: outflow,
		inflow:  inflow,
	}
}

func fixDates(transactions []*transaction) {
	const maxPeriod = 45 * 24 * time.Hour

	firstDate := transactions[0].date
	lastDate := transactions[len(transactions)-1].date

	period := lastDate.Sub(firstDate)
	if period <= maxPeriod {
		return
	}

	log.Printf("Period is larger than a month (%v to %v)", firstDate, lastDate)

	midDate := transactions[len(transactions)/2].date
	adjustedDate := time.Date(midDate.Year(), midDate.Month(), 20, 0, 0, 0, 0, midDate.Location())

	// Adjust
	log.Printf("Middle date: %v", midDate)
	log.Printf("Adjusted date: %v", adjustedDate)

	for _, t := range transactions {
		if lastDate.Sub(t.date) > maxPeriod {
			log.Printf("Adjusting: %v -> %v", t.date, adjustedDate)
			t.date = adjustedDate
		}
	}
}

func exportRows(rows []Row, out io.Writer) error {
	if len(rows) < 2 {
		return errors.New("No transactions")
	}

	header, rows := rows[0], rows[1:]
	cellIndexByName = newCellIndexByName(header)

	// Create transactions

	var transactions []*transaction
	for _, row := range rows {
		t := newTransaction(row)
		transactions = append(transactions, t)
	}

	// Fix up transactions
	fixDates(transactions)

	// Export transactions

	w := csv.NewWriter(out)
	w.Write([]string{"Date", "Payee", "Category", "Memo", "Outflow", "Inflow"})

	for _, t := range transactions {
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
		decodeRowsHtmlCal,
		decodeRowsHtmlMizrahi,
	}

	for _, decode := range decoders {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		rows, err = decode(f)
		if err == nil && len(rows) > 0 {
			return rows, nil
		}
	}

	return nil, err
}

func basename(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

func decodeAll() error {
	inputDir := "data/input"
	outputDir := "data/output"

	fileInfos, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, fi := range fileInfos {
		filename := fi.Name()

		inFilename := filepath.Join(inputDir, filename)
		outFilename := filepath.Join(outputDir, fmt.Sprintf("%v.%v", basename(filename), "csv"))

		log.Printf("Processing %v: %v -> %v", filename, inFilename, outFilename)

		rows, err := decodeFile(inFilename)
		if err != nil {
			return err
		}

		outFile, err := os.Create(outFilename)
		if err != nil {
			return err
		}
		err = exportRows(rows, outFile)
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
