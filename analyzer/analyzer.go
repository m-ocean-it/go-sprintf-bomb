package analyzer

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"slices"
	"strconv"
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

type filePath = string
type packagesOutput = map[filePath]*packagesFileResult

type packagesFileResult struct {
	fmtCount     int
	addedStrConv bool
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	packagesResult := packagesOutput{}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		diagnostic := processNode(pass.Fset, pass.TypesInfo, node, packagesResult)
		if diagnostic == nil {
			return
		}

		pass.Report(*diagnostic)
	})

	importSpecFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	insp.Preorder(importSpecFilter, func(n ast.Node) {
		genDecl, _ := n.(*ast.GenDecl)
		if genDecl == nil {
			return // just in case...
		}

		if genDecl.Tok != token.IMPORT {
			return
		}

		fPath := pass.Fset.Position(genDecl.TokPos).Filename
		filePkgResult := packagesResult[fPath]
		if filePkgResult == nil {
			return
		}

		importDiagnostic := processImportBlock(pass.Fset, genDecl, filePkgResult)
		if importDiagnostic == nil {
			return
		}

		pass.Report(*importDiagnostic)
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

func processNode(
	fset *token.FileSet,
	typesInfo *types.Info,
	node ast.Node,
	pkgOut packagesOutput,
) *analysis.Diagnostic {
	if node == nil {
		return nil
	}

	expr, _ := node.(ast.Expr)
	if expr == nil {
		return nil
	}

	fPath := fset.Position(expr.Pos()).Filename
	filePkgOut := pkgOut[fPath]
	if filePkgOut == nil {
		filePkgOut = &packagesFileResult{}
		pkgOut[fPath] = filePkgOut
	}

	return processExpr(fset, typesInfo, expr, filePkgOut)
}

func processExpr(
	fset *token.FileSet,
	typesInfo *types.Info,
	expr ast.Expr,
	filePkgOut *packagesFileResult,
) *analysis.Diagnostic {
	switch e := expr.(type) {
	case *ast.CallExpr:
		return processCallExpr(fset, typesInfo, e, filePkgOut)
	default:
		return nil
	}
}

func processCallExpr(
	fset *token.FileSet,
	typesInfo *types.Info,
	callExpr *ast.CallExpr,
	filePkgOut *packagesFileResult,
) *analysis.Diagnostic {
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

	filePkgOut.fmtCount++

	if selExpr.Sel.Name != "Sprintf" {
		return nil
	}

	if len(callExpr.Args) < 2 {
		// TODO: handle case
		return nil
	}

	return optimizeSprintf(fset, typesInfo, callExpr, filePkgOut)
}

func optimizeSprintf(
	fset *token.FileSet,
	typesInfo *types.Info,
	callExpr *ast.CallExpr,
	filePkgOut *packagesFileResult,
) *analysis.Diagnostic {
	newExpr, ok := ProcessSprintfCall(typesInfo, callExpr, filePkgOut)
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

func processImportBlock(
	fset *token.FileSet,
	genDecl *ast.GenDecl,
	filePkgResult *packagesFileResult,
) *analysis.Diagnostic {
	if filePkgResult.fmtCount > 0 && !filePkgResult.addedStrConv {
		return nil
	}

	var alreadyHasStrConv bool
	fmtImportIndex := -1

	for i, spec := range genDecl.Specs {
		importSpec, _ := spec.(*ast.ImportSpec)
		if importSpec == nil {
			continue // just in case...
		}

		pathLit := importSpec.Path
		if pathLit.Kind != token.STRING {
			return nil
		}

		importPath, err := strconv.Unquote(pathLit.Value)
		if err != nil {
			return nil
		}

		switch importPath {
		case "fmt":
			fmtImportIndex = i
		case "strconv":
			alreadyHasStrConv = true
		}
	}

	if filePkgResult.fmtCount == 0 && fmtImportIndex > -1 {
		genDecl.Specs = slices.Delete(genDecl.Specs, fmtImportIndex, fmtImportIndex+1)
	}
	if filePkgResult.addedStrConv && !alreadyHasStrConv {
		genDecl.Specs = append(genDecl.Specs, &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("strconv")},
		})
	}

	return newAnalysisDiagnostic(
		genDecl,
		"Fix imports",
		[]analysis.SuggestedFix{
			{
				Message: "Fix imports",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     genDecl.Pos(),
						End:     genDecl.End(),
						NewText: []byte(formatNode(fset, genDecl)),
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
