package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize_RBC(t *testing.T) {
	inputFile, err := os.Open("testdata/rbc_txns_input.csv")
	require.NoError(t, err)

	outputFile, err := os.Open("testdata/rbc_txns_output.csv")
	require.NoError(t, err)

	buffer := new(bytes.Buffer)
	err = run(t.Context(), inputFile, buffer, t.Output())
	require.NoError(t, err)

	expectedOutput, err := io.ReadAll(outputFile)
	require.NoError(t, err)
	require.Equal(t, string(expectedOutput), buffer.String())
}

func TestNormalize_Scotia(t *testing.T) {
	inputFile, err := os.Open("testdata/scotia_txns_input.csv")
	require.NoError(t, err)

	outputFile, err := os.Open("testdata/scotia_txns_output.csv")
	require.NoError(t, err)

	buffer := new(bytes.Buffer)
	err = run(t.Context(), inputFile, buffer, t.Output())
	require.NoError(t, err)

	expectedOutput, err := io.ReadAll(outputFile)
	require.NoError(t, err)
	require.Equal(t, string(expectedOutput), buffer.String())
}
