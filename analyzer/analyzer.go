package analyzer

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "foobar",
		URL:      "foobar",
		Doc:      "foobar",
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
		panic("todo")
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
		"foobar",
		[]analysis.SuggestedFix{
			{
				Message: "foobar",
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

var supportedVerbs = []string{"%s", "%d"} // TODO support more

func isVerb(rs string) bool {
	return slices.Contains(supportedVerbs, rs)
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}
