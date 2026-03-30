package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

var (
	scotiaHeader = []string{
		"Filter",
		"Date",
		"Description",
		"Sub-description",
		"Type of Transaction",
		"Amount",
		"Balance",
	}
	scotiaRecordLength = 7
	scotiaAccountName  = "Scotia Chequing"
)

type scotiaParser struct{}

func (s *scotiaParser) MatchHeader(header []string) bool {
	if len(header) != len(scotiaHeader) {
		return false
	}

	for idx, headerField := range header {
		if headerField != scotiaHeader[idx] {
			return false
		}
	}

	return true
}

func (s *scotiaParser) Parse(record []string) (*OutputTxn, error) {
	scotiaIn, err := newScotiaInputTxn(record)
	if err != nil {
		return nil, err
	}
	return scotiaIn.OutputTxn(), nil
}

type scotiaInputTxn struct {
	Filter            string
	Date              time.Time
	Description       string
	SubDescription    string
	TypeOfTransaction string
	Amount            float64
	Balance           float64
}

func (s *scotiaInputTxn) OutputTxn() *OutputTxn {
	return &OutputTxn{
		Account:     scotiaAccountName,
		Date:        s.Date,
		Description: s.Description,
		Amount:      s.Amount,
		Currency:    CurrencyCAD,
	}
}

func newScotiaInputTxn(record []string) (*scotiaInputTxn, error) {
	if len(record) != scotiaRecordLength {
		return nil,
			fmt.Errorf("scotia record length size unexpected, expected %d, found %d",
				scotiaRecordLength, len(record))
	}
	in := &scotiaInputTxn{
		Filter:            record[0],
		Description:       record[2],
		SubDescription:    record[3],
		TypeOfTransaction: record[4],
	}
	date, err := time.Parse("2006-01-02", record[1])
	if err != nil {
		return nil, errors.New("failed to parse date from scotia csv")
	}
	in.Date = date

	amount, err := strconv.ParseFloat(record[5], 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount from scotia csv, %w", err)
	}
	in.Amount = amount

	balance, err := strconv.ParseFloat(record[6], 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance from scotia csv, %w", err)
	}
	in.Balance = balance

	return in, nil
}
