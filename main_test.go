package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

const common = `
package main
import "fmt"
func main() {
	_ = %s
}
`

func TestProcessCallExpr(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:   "foo",
			source: `fmt.Sprintf("Hello, %s!", "Max")`,
			expected: `"Hello, " +

	"Max" + "!"`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			source := fmt.Sprintf(common, tc.source)
			f, err := parser.ParseFile(fset, "", source, 0)
			if err != nil {
				panic(err)
			}
			expr := f.Decls[1].(*ast.FuncDecl).
				Body.
				List[0].(*ast.AssignStmt).
				Rhs[0]

			fixedExpr, err := ProcessExpr(expr)

			require.NoError(t, err)
			got := formatNode(fset, fixedExpr)
			require.Equal(t, tc.expected, got)
		})
	}
}

// func TestSplitConcatedString(t *testing.T) {
// 	t.Parallel()

// 	scs := SplitConcatedString{
// 		parts: []part{val: `"Hello, "`, `"!"`},
// 	}

// 	expected := `"Hello, " + "Max" + "!"`
// 	got := scs.Fill([]string{`"Max"`})

// 	require.Equal(t, expected, got)
// }

func TestSplitConcat(t *testing.T) {
	t.Parallel()

	got := SplitConcat("Hello, %s!")

	require.Equal(
		t,
		[]part{
			{val: "Hello, "},
			{val: "%s", isVerb: true},
			{val: "!"},
		},
		got.parts,
	)
}

// func TestConcatedStringFill(t *testing.T) {
// 	t.Parallel()

// 	s := SplitConcatedString{}

// 	// gots.Fill()
// }

func formatNode(fset *token.FileSet, node ast.Node) string {

	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}
