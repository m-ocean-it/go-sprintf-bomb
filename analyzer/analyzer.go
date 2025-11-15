package analyzer

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "SprintfBomb",
		URL:      "https://github.com/m-ocean-it/go-sprintf-bomb",
		Doc:      "https://github.com/m-ocean-it/go-sprintf-bomb",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		diagnostic := processNode(pass.Fset, pass.TypesInfo, node)
		if diagnostic == nil {
			return
		}

		pass.Report(*diagnostic)
	})

	return nil, nil
}

func newAnalysisDiagnostic(
	analysisRange analysis.Range,
	message string,
	suggestedFixes []analysis.SuggestedFix,
) *analysis.Diagnostic {

	return &analysis.Diagnostic{
		Pos:            analysisRange.Pos(),
		End:            analysisRange.End(),
		SuggestedFixes: suggestedFixes,
		Message:        message,
	}
}

func processNode(fset *token.FileSet, typesInfo *types.Info, node ast.Node) *analysis.Diagnostic {
	if node == nil {
		return nil
	}

	expr, _ := node.(ast.Expr)
	if expr == nil {
		return nil
	}

	return processExpr(fset, typesInfo, expr)
}

func processExpr(fset *token.FileSet, typesInfo *types.Info, expr ast.Expr) *analysis.Diagnostic {
	switch e := expr.(type) {
	case *ast.CallExpr:
		return processCallExpr(fset, typesInfo, e)
	default:
		return nil
	}
}

func processCallExpr(fset *token.FileSet, typesInfo *types.Info, callExpr *ast.CallExpr) *analysis.Diagnostic {
	selExpr, _ := callExpr.Fun.(*ast.SelectorExpr)
	if selExpr == nil {
		return nil
	}

	xIdent, _ := selExpr.X.(*ast.Ident)
	if xIdent == nil {
		return nil
	}
	if xIdent.Name != "fmt" {
		return nil
	}

	if selExpr.Sel.Name != "Sprintf" {
		return nil
	}

	if len(callExpr.Args) < 2 {
		// TODO: handle case
		return nil
	}

	return optimizeSprintf(fset, typesInfo, callExpr)
}

func optimizeSprintf(fset *token.FileSet, typesInfo *types.Info, callExpr *ast.CallExpr) *analysis.Diagnostic {
	newExpr, ok := ProcessSprintfCall(typesInfo, callExpr)
	if !ok {
		return nil
	}

	return newAnalysisDiagnostic(
		callExpr,
		"Sprintf could be optimized away",
		[]analysis.SuggestedFix{
			{
				Message: "Sprintf could be optimized away",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     callExpr.Pos(),
						End:     callExpr.End(),
						NewText: []byte(formatNode(fset, newExpr)),
					},
				},
			},
		},
	)
}

var supportedVerbs = []string{"%s", "%d", "%f"} // TODO support more

func isVerb(rs string) bool {
	return slices.Contains(supportedVerbs, rs)
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		return formatBinaryExpr(fset, n)
	default:
		return formatAnyNode(fset, n)
	}
}

func formatAnyNode(fset *token.FileSet, node ast.Node) string {
	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, node); err != nil {
		return ""
	}

	return buf.String()
}

func formatBinaryExpr(fset *token.FileSet, node ast.Node) string {
	binExpr := node.(*ast.BinaryExpr)

	stringBuilder := &strings.Builder{}

	stack := []*ast.BinaryExpr{binExpr}
	var popped bool

	for len(stack) > 0 {
		lastBinExpr := stack[len(stack)-1]

		if lastBinExpr.Op != token.ADD {
			panic("wrong op")
		}

		if popped {
			stringBuilder.WriteString(" + ")

			if err := format.Node(stringBuilder, fset, lastBinExpr.Y); err != nil {
				return ""
			}
			stack = stack[:len(stack)-1] // pop
			popped = true

			continue
		}

		switch x := lastBinExpr.X.(type) {
		case *ast.BinaryExpr:
			stack = append(stack, x)
			popped = false
			continue
		default:
			if err := format.Node(stringBuilder, fset, x); err != nil {
				return ""
			}
			popped = true
		}
	}

	return stringBuilder.String()
}
