package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	inputFile  = flag.String("in", "", "input csv file")
	outputFile = flag.String("out", "", "output csv file")
	accounts   = map[string]string{
		"05920-5031885":    "RBC Chequing",
		"05920-5083274":    "RBC Savings",
		"4514011614694030": "RBC Visa",
	}
	outputHeader = []string{
		"Account",
		"Date",
		"Description",
		"Amount",
	}
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func logRecord(logger *slog.Logger, i int, record []string) {
	recordStr := strings.Join(record, ",")
	logger.Info(fmt.Sprintf("%d-%s", i, recordStr), "len", len(record))
}

func run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {

	logger := slog.New(slog.NewTextHandler(stderr, nil))

	reader := csv.NewReader(stdin)
	reader.FieldsPerRecord = -1 // turn of number of field checking
	writer := csv.NewWriter(stdout)
	defer writer.Flush()

	for i := 0; ; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if errors.Is(err, csv.ErrFieldCount) {
			logRecord(logger, i, record)
			return err
		}
		if err != nil {
			return err
		}
		logRecord(logger, i, record)

		if i == 0 {
			// skip the first line of headers
			err = writer.Write(outputHeader)
			if err != nil {
				return err
			}
			continue
		}

		in, err := NewInputTxn(record)
		if err != nil {
			return err
		}

		out, err := convert(in)
		if err != nil {
			return err
		}
		if out == nil {
			continue
		}
		logger.Info(fmt.Sprintf("converted record %d:", i), "converted", out)

		err = writer.Write(out.Record())
		if err != nil {
			return err
		}
	}
}

func convert(in InputTxn) (*OutputTxn, error) {
	out := OutputTxn{}

	// Account = "<Account Name>"
	accountName, ok := accounts[in.AccountNumber()]
	if !ok {
		return nil, nil
	}
	out.Account = fmt.Sprintf("%s", accountName)

	date, err := time.Parse("1/2/2006", in.Date())
	if err != nil {
		return nil, err
	}
	out.Date = date

	// default to description 2, fallback to description 1
	desc := in.Description2()
	if desc == "" {
		desc = in.Description1()
	}
	if desc == "" {
		desc = "NOT SPECIFIED"
	}
	out.Description = desc

	// Determine currency and transaction amount
	var (
		currency Currency
		amount   float64
	)
	if in.CAD() != "" {
		currency = CurrencyCAD
		amount, err = in.CADInt()
		if err != nil {
			return nil, err
		}
	} else if in.USD() != "" {
		currency = CurrencyUSD
		amount, err = in.USDInt()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no CAD or USD txn value included")
	}
	if err != nil {
		return nil, err
	}
	out.Currency = currency
	out.Amount = amount

	return &out, nil
}

func getCAD(in []string) (float64, error) {
	return strconv.ParseFloat(in[6], 32)
}

func getUSD(in []string) (float64, error) {
	return strconv.ParseFloat(in[7], 32)
}

func getTxn(in InputTxn) (float64, string, error) {
	cad, err := getCAD(in)
	if err != nil {
		usd, err := getUSD(in)
		if err != nil {
			return 0, "", fmt.Errorf("no cad/usd transaction value specified")
		}
		return usd, "usd", nil
	}
	return cad, "cad", nil
}

type Currency string

var (
	CurrencyCAD Currency = "cad"
	CurrencyUSD Currency = "usd"
)

type OutputTxn struct {
	Account     string
	Date        time.Time
	Description string
	Amount      float64
	Currency    Currency
}

func (o *OutputTxn) Record() []string {
	return []string{
		o.Account,
		o.Date.Format("1/2/2006"),
		strings.Trim(o.Description, " "),
		fmt.Sprintf("%.2f", o.Amount),
		string(o.Currency),
	}
}

type InputTxn []string

func NewInputTxn(record []string) (InputTxn, error) {
	return InputTxn(record), nil
}

func (i InputTxn) AccountType() string {
	return i[0]
}

func (i InputTxn) AccountNumber() string {
	return i[1]
}

func (i InputTxn) Date() string {
	return i[2]
}

func (i InputTxn) Description1() string {
	return i[3]
}

func (i InputTxn) Description2() string {
	return i[4]
}

func (i InputTxn) CAD() string {
	return i[6]
}

func (i InputTxn) CADInt() (float64, error) {
	return strconv.ParseFloat(i.CAD(), 32)
}

func (i InputTxn) USD() string {
	return i[7]
}

func (i InputTxn) USDInt() (float64, error) {
	return strconv.ParseFloat(i.USD(), 32)
}
