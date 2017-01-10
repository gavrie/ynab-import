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
	"path"
)

var fieldNames = map[string][]string{
	"amount":  {"סכום חיוב ₪", "סכוםהחיוב", "סכום החיוב בש''ח", "סכום לחיוב", "סכום החיוב"},
	"inflow":  {"זכות"},
	"outflow": {"חובה"},
	"date":    {"תאריך עסקה", "תאריךהעסקה", "תאריך הקנייה", "תאריך רכישה", "תאריך", "תאריך העסקה"},
	"payee":   {"שם בית העסק", "שםבית העסק", "שם בית עסק", "סוג תנועה", "תיאור"},
	"memo":    {"הערות", "פירוט נוסף", "מידע נוסף", "פרוט נוסף", "אסמכתא", "פרטים"},
}

var dateFormats = []string{
	"2006-01-02T15:04:05",
	"02/01/06",
	"2006-01-02",
	"02/01/2006",
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

var ErrNoSuchCell = errors.New("No such cell")

func getCell(row Row, field string) (string, error) {
	names, ok := fieldNames[field]
	if !ok {
		err := fmt.Errorf("No cell names found for field '%v'", field)
		log.Panic(err)
		return "", err
	}

	for _, name := range names {
		i, ok := cellIndexByName[name]
		if ok {
			return row[i], nil
		}
	}
	log.Printf("%v", cellIndexByName)
	log.Printf("No cell found matching field '%v'", field)
	return "", ErrNoSuchCell
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
	//log.Printf("cellIndexByName: %#v", cellIndexByName)
	//log.Printf("row: %#v", row)
	//if len(row) != len(cellIndexByName) {
	//	log.Panic("Unexpected row length")
	//}

	var outflow, inflow string

	stringAmount, err := getCell(row, "amount")
	if err == ErrNoSuchCell {
		// We have separate inflow and outflow cells
		inflow, err = getCell(row, "inflow")
		if err != nil {
			log.Printf("%v",row)
			log.Panic(err)
		}
		outflow, err = getCell(row, "outflow")
		if err != nil {
			log.Panic(err)
		}
	} else if err != nil {
		log.Panic(err)
	} else {
		// We have a unified amount cell and need to split it into inflow and outflow
		stringAmount = strings.Replace(stringAmount, ",", "", -1)
		amount, err := strconv.ParseFloat(stringAmount, 64)
		if err != nil {
			log.Panic("Non-numeric value encountered ", stringAmount)
		}

		if amount > 0 {
			outflow = fmt.Sprintf("%v", amount)
		} else {
			inflow = fmt.Sprintf("%v", -amount)
		}
	}

	stringDate, err := getCell(row, "date")
	if err != nil {
		log.Panic(err)
	}
	date, err := parseDate(stringDate)
	if err != nil {
		log.Panicf("Invalid date encountered ", stringDate)
	}

	payee, err := getCell(row, "payee")
	if err != nil {
		log.Panic(err)
	}
	memo, err := getCell(row, "memo")
	if err != nil {
		memo = ""
		log.Print(err)
	}

	return &transaction{
		date:    date,
		payee:   payee,
		memo:    memo,
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
		log.Panic(err)
	}

	return nil
}

func decodeFile(filename string) (rows []Row, err error) {
	decoders := []func(io.Reader) ([]Row, error){
		// Credit cards
		decodeRowsXml,
		decodeRowsHtmlCal,
		decodeRowsHtmlMizrahiCC,
		decodeRowsHtmlIsracard,

		// Checking accounts
		decodeRowsHtmlMizrahiChecking,
		decodeRowsLeumiChecking,
		decodeRowsLeumicard,
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

func decodeAll(datapath string) error {
	inputDir := path.Join(datapath, "input")
	outputDir := path.Join(datapath, "output")

	fileInfos, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, fi := range fileInfos {
		filename := fi.Name()

		// Skip Mac folder attributes
		if filename == ".DS_Store" {
			continue
		}

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
	filePath := os.Args[1]

	err := decodeAll(filePath)
	if err != nil {
		log.Panic(err)
	}
}
