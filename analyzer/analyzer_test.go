package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		a := New()

		analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), a, "default")
	})
}
