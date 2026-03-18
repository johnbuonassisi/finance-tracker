package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	file, err := os.Open("testdata/csv18320.csv")
	require.NoError(t, err)

	buffer := new(bytes.Buffer)
	err = run(t.Context(), file, buffer, t.Output())
	require.NoError(t, err)

	t.Logf("buffer %s", buffer.String())
}
