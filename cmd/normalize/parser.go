package main

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)

var (
	outputHeader = []string{
		"ID",
		"Account",
		"Date",
		"Description",
		"Amount",
	}
)

// Parser for normalizing CSV records to OutputTxn. Also
// includes a utility for matching the header.
type Parser interface {
	MatchHeader(record []string) bool
	Parse([]string) (*OutputTxn, error)
}

// OutputTxn is the standardized format of a Transaction
type OutputTxn struct {
	Account     string
	Date        time.Time
	Description string
	Amount      float64
	Currency    Currency
}

// Record generates a CSV record for the Transaction
func (o *OutputTxn) Record() []string {
	hashID := fmt.Sprintf("%s-%s-%s-%.2f", o.Account, o.Date, o.Description, o.Amount)
	hashBytes := sha1.Sum([]byte(hashID))
	return []string{
		fmt.Sprintf("%x", hashBytes),
		o.Account,
		o.Date.Format("1/2/2006"),
		strings.Trim(o.Description, " "),
		fmt.Sprintf("%.2f", o.Amount),
		string(o.Currency),
	}
}
