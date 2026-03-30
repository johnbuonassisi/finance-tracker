package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/johnbuonassisi/finance-tracker/internal/auth"
)

const (
	spreadsheetID               = "1gVUDsKybwI0XOtB4WPDjJvTuYK4DN_-CNmqCDsv7o-E"
	sheetName                   = "Transactions"
	valueInputOptionUserEntered = "USER_ENTERED"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}

func run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {

	client, err := auth.NewDefaultClient()
	if err != nil {
		return fmt.Errorf("unable to create new Sheets client: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets Service: %w", err)
	}

	reader := csv.NewReader(stdin)
	reader.FieldsPerRecord = -1 // might not need this (CSV from RBC needed this as it had an extra comma on every line...)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	valueRange, err := convertRecords(records)
	if err != nil {
		return fmt.Errorf("failed to convert records, %w", err)
	}

	err = appendToSheet(srv, valueRange)
	if err != nil {
		return err
	}

	return nil
}

func appendToSheet(srv *sheets.Service, valueRange *sheets.ValueRange) error {
	appendRange := fmt.Sprintf("%s!A1", sheetName)
	_, err := srv.Spreadsheets.Values.Append(spreadsheetID,
		appendRange,
		valueRange).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return fmt.Errorf("failed to append values to sheet with range %s, %w", appendRange, err)
	}

	return nil
}

// convertRecords of a CSV file to a *sheets.ValuesRange containing equivalent cells of a sheet
func convertRecords(records [][]string) (*sheets.ValueRange, error) {
	// ignore the header if it is included
	if isHeader(records[0]) {
		records = records[1:]
	}

	// convert the record, [][]string, to Values, [][]any
	values := make([][]any, len(records))
	for idx, record := range records {
		value, err := convertRecord(record)
		if err != nil {
			return nil, fmt.Errorf("failed to convert record, %w", err)
		}
		values[idx] = value
	}

	return &sheets.ValueRange{
		Values: values,
	}, nil
}

// convertRecord, []string containing normalized csv fields to
// []any representing cells in a row in a google sheet, []any.
func convertRecord(record []string) ([]any, error) {
	if len(record) < 5 {
		return nil, errors.New("record must contain 5 columns")
	}

	// note: record[0] holds an id that is currently not used
	// in the future this could be used to de-depulicate transactions
	value := []any{
		record[1], // Account
		record[2], // Date
		record[3], // Description
		record[4], // Amount
	}
	return value, nil
}

// isHeader checks if a records contains the normalized CSV header
func isHeader(record []string) bool {
	if len(record) == 0 {
		return false
	}
	if record[0] != "ID" {
		return false
	}
	return true
}
