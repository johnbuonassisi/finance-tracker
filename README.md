# finance-tracker

## normalize

A simple command line utility that can read an RBC generated CSV file and output a normalized CSV file that
can be uploaded to the Finance Tracker google sheet.

The command line utility reads from stdin and writes the normalized csv to stdout, while it writes logs to stderr.

An RBC generated CSV file contains the following columns:
`"Account Type","Account Number","Transaction Date","Cheque Number","Description 1","Description 2","CAD$","USD$"`

Whereas the normalized CSV file contains the following columns:
`Account,Date,Description,Amount`


### usage

```bash
go run cmd/normalize/main.go < ./csvs/csv18320.csv
```

## upload

A simple command line utility that can read a normalized CSV file and upload its contents to the Finance Tracker Google sheet.

The command line utility reads from stdin and writes the normalized csv to stdout, while it writes logs to stderr.


### usage

```bash
go run cmd/upload/main.go < ./csvs/csv18320_normalized.csv
```

## categorize

A simple command line uitlity that can interact with the Finance Tracker Google sheet to automatically categorize transactions.

```bash
go run cmd/categorize/main.go
```
