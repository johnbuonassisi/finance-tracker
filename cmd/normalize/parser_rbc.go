package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	rbcHeaderRecord = []string{
		"Account Type",
		"Account Number",
		"Transaction Date",
		"Cheque Number",
		"Description 1",
		"Description 2",
		"CAD$",
		"USD$",
	}
	enabledAccounts = map[string]string{
		"05920-5031885":    "RBC Chequing",
		"05920-5083274":    "RBC Savings",
		"4514011614694030": "RBC Visa",
	}
)

// rbcParser parses RBC CSV records to the standard OutputTxn record
type rbcParser struct{}

func (r *rbcParser) MatchHeader(header []string) bool {

	// headers need to be the same size``
	if len(header) != len(rbcHeaderRecord) {
		return false
	}

	// all elements of headers need to match
	for idx, headerField := range header {
		if headerField != rbcHeaderRecord[idx] {
			return false
		}
	}

	return true
}

func (r *rbcParser) Parse(record []string) (*OutputTxn, error) {

	in, err := NewInputTxn(record)
	if err != nil {
		return nil, err
	}

	out, err := r.parse(in)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *rbcParser) parse(in RBCInputTxn) (*OutputTxn, error) {
	out := OutputTxn{}

	// Account = "<Account Name>"
	accountName, ok := enabledAccounts[in.AccountNumber()]
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
	desc := "NO DESCRIPTION"
	if strings.ToLower(in.AccountType()) == "visa" {
		// for visa accounts just grab the merchant in the first description, second
		// is always empty
		desc = in.Description1()
	} else {
		// for other accounts graph both descriptions as one will contain the txn type and
		// the other the merchant. Sometimes the type/merchant are swapped so just concat both.
		desc = fmt.Sprintf("%s %s", in.Description1(), in.Description2())
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

type RBCInputTxn []string

func NewInputTxn(record []string) (RBCInputTxn, error) {
	return RBCInputTxn(record), nil
}

func (i RBCInputTxn) AccountType() string {
	return i[0]
}

func (i RBCInputTxn) AccountNumber() string {
	return i[1]
}

func (i RBCInputTxn) Date() string {
	return i[2]
}

func (i RBCInputTxn) Description1() string {
	return i[4]
}

func (i RBCInputTxn) Description2() string {
	return i[5]
}

func (i RBCInputTxn) CAD() string {
	return i[6]
}

func (i RBCInputTxn) CADInt() (float64, error) {
	return strconv.ParseFloat(i.CAD(), 32)
}

func (i RBCInputTxn) USD() string {

	return i[7]
}

func (i RBCInputTxn) USDInt() (float64, error) {
	return strconv.ParseFloat(i.USD(), 32)
}

type Currency string

var (
	CurrencyCAD Currency = "cad"
	CurrencyUSD Currency = "usd"
)
