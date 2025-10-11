package analyzer

import (
	"bytes"
	"errors"
	"go/ast"
	"go/format"
	"go/token"
	"slices"
	"strconv"

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
		diagnostic := processNode(pass.Fset, node)
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

func processNode(fset *token.FileSet, node ast.Node) *analysis.Diagnostic {
	if node == nil {
		return nil
	}

	expr, _ := node.(ast.Expr)
	if expr == nil {
		return nil
	}

	return processExpr(fset, expr)
}

func processExpr(fset *token.FileSet, expr ast.Expr) *analysis.Diagnostic {
	switch e := expr.(type) {
	case *ast.CallExpr:
		return processCallExpr(fset, e)
	default:
		return nil
	}
}

func processCallExpr(fset *token.FileSet, callExpr *ast.CallExpr) *analysis.Diagnostic {
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

	return optimizeSprintf(fset, callExpr)
}

func optimizeSprintf(fset *token.FileSet, callExpr *ast.CallExpr) *analysis.Diagnostic {
	s := callExpr.Args[0].(*ast.BasicLit)

	if s.Kind != token.STRING {
		panic("todo")
		// TODO support any expression of type string
	}

	sVal, err := strconv.Unquote(s.Value)
	if err != nil {
		return nil
	}

	splitConcatedStr := splitConcat(sVal)

	result, err := splitConcatedStr.fill(callExpr.Args[1:])
	if err != nil {
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
						NewText: []byte(formatNode(fset, result)),
					},
				},
			},
		},
	)
}

type splitConcatedString struct {
	parts []part
}

type part struct {
	val    string
	isVerb bool
}

func (s *splitConcatedString) fill(args []ast.Expr) (ast.Expr, error) {
	res := &ast.BinaryExpr{
		Op: token.ADD,
	}

	var nextArg int
	for _, p := range s.parts {
		var e ast.Expr
		if p.isVerb {
			switch p.val {
			case "%s":
				e = args[nextArg]
			case "%d":
				a := args[nextArg]
				switch aTyped := a.(type) {
				case *ast.BasicLit:
					if aTyped.Kind != token.INT {
						return nil, errors.New("aTyped.Kind != token.INT")
					}

					e = &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "strconv",
							},
							Sel: &ast.Ident{
								Name: "Itoa",
							},
						},
						Args: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.INT,
								Value: aTyped.Value,
							},
						},
					}
				case *ast.Ident:
					e = &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "strconv",
							},
							Sel: &ast.Ident{
								Name: "FormatInt", // TODO we need type info here (what if it's uint?)
							},
						},
						Args: []ast.Expr{
							aTyped,
							&ast.BasicLit{
								Kind:  token.INT,
								Value: "10",
							},
						},
					}
				}

			default:
				panic("todo")
			}

			nextArg++
		} else {
			e = &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(p.val)}
		}

		if res.X == nil {
			res.X = e
		} else if res.Y == nil {
			res.Y = e
		} else {
			res = &ast.BinaryExpr{X: res, Y: e, Op: token.ADD}
		}
	}

	return res, nil
}

var supportedVerbs = []string{"%s", "%d"} // TODO support more

func splitConcat(source string) *splitConcatedString {
	// TODO: account for numbered placeholders (%[1]s, etc.)
	// TODO: account for escaping

	var parts []part

	var v []rune

	for _, r := range source {
		if r == '%' {
			if len(v) > 0 {
				parts = append(parts, part{val: string(v)})
			}
			v = v[:0]
		} else if isVerb(v) {
			parts = append(parts, part{val: string(v), isVerb: true})
			v = v[:0]
		}

		v = append(v, r)
	}
	if len(v) > 0 {
		parts = append(parts, part{
			val:    string(v),
			isVerb: isVerb(v),
		})
	}

	return &splitConcatedString{
		parts: parts,
	}
}

func isVerb(rs []rune) bool {
	return slices.Contains(supportedVerbs, string(rs))
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}
