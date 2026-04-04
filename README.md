# finance-tracker

## motivation

The motivation behind this project was to simply determine my family's income vs. expenses and also understand our spending. I was frustrated
with YNAB because it couldn't answer these questions and was fundamentally incompatible with my finances. A list of issues I encountered includes:

* Does not support multiple currencies. I live in Canada and receive income typically in CAD. However, at one point I worked on contract for
an American company and was paid in USD. I work for another company now and receive stocks which I sell for USD and hold in a chequing account.
In order to track this in YNAB I manually added transactions and converted USD to CAD when inputting them.
* Does not automatically help categorize transactions. Every transaction needs to be manually categorized, and categories need to be setup.
A proper setup required me lots of time to understand the best way to do it and requires constant adjustment.
* Does not generate reports accurately. YNAB considers transfers to investment accounts as an expense, so both net worth and income vs. expense
reports are completely wrong and there is no way to correct them.
* Syncing of transactions continually breaks and needs re-authorization. Their integration with Plaid takes a really long time to connect
with each financial institution.

I knew that all my banking institutions allow for CSV downloads so if I could download the transactions and normalize them to a common format
then I could apply better accounting principles and generate better reports. This requires some manual processes but RBC for example makes it very
easy because they have a "download txns since last time" feature. Given my main chequing account is in RBC, I could just login and download a single
CSV file on occasion, and use it as input into my system. The only issue is that parsing the RBC CSV file requires some custom logic. I also need
to download transactions from Scotiabank. They make it relatively easy with a download buttom per account and they support transactions from up to 
2 years ago. Their CSV format is obviously different than RBC and requires different logic but it is fairly simple. My hypothesis is that there are
others like me who would prefer a method like this for managing their finances vs. systems like YNAB.

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
