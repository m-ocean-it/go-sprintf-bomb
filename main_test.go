package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitConcatedString(t *testing.T) {
	t.Parallel()

	scs := SplitConcatedString{
		parts: []string{`"Hello, "`, `"!"`},
	}

	expected := `"Hello, " + "Max" + "!"`
	got := scs.Fill([]string{`"Max"`})

	require.Equal(t, expected, got)
}

func TestSplitConcat(t *testing.T) {
	t.Parallel()

	got := SplitConcat("Hello, %s!")

	require.Equal(t, []string{`"Hello, "`, `"!"`}, got.parts)
}
