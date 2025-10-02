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

func TestProcessWholeFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		source   string
		expected string
	}{
		// =======================================================
		{
			name: "int variable",
			source: `
package main
import "fmt"
func main() {
	n := int(3)

	_ = strconv.Sprintf("High %d!", n)
}
`,
			expected: `package main

import "fmt"

func main() {
	n := int(3)

	_ = "High " + strconv.Itoa(n) + "!"
}
`,
		},
		// =======================================================
		{
			name: "int64 variable",
			source: `
package main
import "fmt"
func main() {
	n := int64(3)

	_ = fmt.Sprintf("High %d!", n)
}
`,
			expected: `package main

import "fmt"

func main() {
	n := int64(3)

	_ = "High " + strconv.FormatInt(n, 10) + "!"
}
`,
		},
		// =======================================================
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tc.source, 0)
			if err != nil {
				panic(err)
			}

			ProcessFile(file)

			got := formatNode(fset, file)

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestProcessCallExpr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:   "Hello, %s",
			source: fmt.Sprintf(common, `fmt.Sprintf("Hello, %s!", "Max")`),
			expected: `"Hello, " +

	"Max" + "!"`,
		},
		{
			name:     "%s, hello!",
			source:   fmt.Sprintf(common, `fmt.Sprintf("%s, hello!", "Max")`),
			expected: `"Max" + ", hello!"`,
		},
		{
			name:     "number literal",
			source:   fmt.Sprintf(common, `fmt.Sprintf("High %d!", 5)`),
			expected: `"High " + strconv.Itoa(5) + "!"`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tc.source, 0)
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

func TestSplitConcat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		source   string
		expected []part
	}{
		{
			name:   "Hello, %s",
			source: "Hello, %s!",
			expected: []part{
				{val: "Hello, "},
				{val: "%s", isVerb: true},
				{val: "!"},
			},
		},
		{
			name:   "%s, hello!",
			source: "%s, hello!",
			expected: []part{
				{val: "%s", isVerb: true},
				{val: ", hello!"},
			},
		},
		{
			name:   "High %d!",
			source: "High %d!",
			expected: []part{
				{val: "High "},
				{val: "%d", isVerb: true},
				{val: "!"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SplitConcat(tc.source)

			require.Equal(t, tc.expected, got.parts)
		})
	}

}

func formatNode(fset *token.FileSet, node ast.Node) string {

	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}
