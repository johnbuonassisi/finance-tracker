package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

var (
	parsers = []Parser{
		&rbcParser{},
		&scotiaParser{},
	}
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {

	logger := slog.New(slog.NewTextHandler(stderr, nil))

	reader := csv.NewReader(stdin)
	reader.FieldsPerRecord = -1 // turn of number of field checking

	writer := csv.NewWriter(stdout)
	defer writer.Flush()

	err := parse(ctx, logger, reader, writer, parsers)
	if err != nil {
		return err
	}

	return nil
}

func parse(ctx context.Context, logger *slog.Logger, reader *csv.Reader, writer *csv.Writer, parsers []Parser) error {

	parser, err := detectParser(reader, parsers)
	if err != nil {
		return err
	}

	err = writer.Write(outputHeader)
	if err != nil {
		return err
	}

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

		out, err := parser.Parse(record)
		if err != nil {
			return err
		}
		if out == nil {
			logger.Debug(fmt.Sprintf("skipped record %d", i))
			continue
		}
		logger.Debug(fmt.Sprintf("converted record %d:", i), "converted", out)

		err = writer.Write(out.Record())
		if err != nil {
			return err
		}
	}
}

func detectParser(reader *csv.Reader, parsers []Parser) (Parser, error) {
	header, err := reader.Read()
	if err == io.EOF {
		return nil, errors.New("no header found")
	}
	if err != nil {
		return nil, err
	}

	for _, p := range parsers {
		isMatch := p.MatchHeader(header)
		if isMatch {
			return p, nil
		}
	}
	return nil, errors.New("no parser found")
}

func logRecord(logger *slog.Logger, i int, record []string) {
	recordStr := strings.Join(record, ",")
	logger.Info(fmt.Sprintf("%d-%s", i, recordStr), "len", len(record))
}
