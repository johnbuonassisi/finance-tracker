package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/johnbuonassisi/finance-tracker/internal/auth"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {

	_, err := auth.NewDefaultClient()
	if err != nil {
		return fmt.Errorf("unable to create new Sheets client: %w", err)
	}

	return nil
}
