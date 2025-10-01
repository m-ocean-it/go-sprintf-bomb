package main

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

func main() {
	// fset := token.NewFileSet()
	expr, err := parser.ParseExpr(`fmt.Sprintf("Hello, %s!", "Max")`)
	if err != nil {
		panic(err)
	}
	_ = expr
}

func ProcessExpr(expr ast.Expr) (ast.Expr, error) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		exprNew := processCallExpr(e)
		return exprNew, nil
	default:
		return nil, errors.New("not implemented")
	}

	return expr, nil
}

func processCallExpr(callExpr *ast.CallExpr) ast.Expr {
	selExpr, _ := callExpr.Fun.(*ast.SelectorExpr)
	if selExpr == nil {
		return callExpr
	}

	xIdent, _ := selExpr.X.(*ast.Ident)
	if xIdent == nil {
		return callExpr
	}
	if xIdent.Name != "fmt" {
		return callExpr
	}

	if selExpr.Sel.Name != "Sprintf" {
		return callExpr
	}

	if len(callExpr.Args) < 2 {
		panic("todo")
	}

	return optimizeSprintf(callExpr)
}

func optimizeSprintf(callExpr *ast.CallExpr) ast.Expr {
	s := callExpr.Args[0].(*ast.BasicLit)

	if s.Kind != token.STRING {
		panic("todo")
		// TODO support any expression of type string
	}

	sVal, err := strconv.Unquote(s.Value)
	if err != nil {
		return callExpr
	}

	splitConcatedStr := SplitConcat(sVal)

	return splitConcatedStr.Fill(callExpr.Args[1:])
}

type SplitConcatedString struct {
	parts []part
}

type part struct {
	val    string
	isVerb bool
}

func (s *SplitConcatedString) Fill(args []ast.Expr) ast.Expr {
	res := &ast.BinaryExpr{
		Op: token.ADD,
	}

	var nextArg int
	for _, p := range s.parts {
		var e ast.Expr
		if p.isVerb {
			e = args[nextArg]
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

	return res
}

func SplitConcat(source string) *SplitConcatedString {
	// FIXME: won't work for many cases
	// TODO: account for escaping

	// sep := "%s" // TODO: support more verbs

	var parts []part

	sourceRunes := []rune(source)

	var percent bool
	var nextRuneIdx int

	for i, r := range sourceRunes {
		if r == 's' && percent {
			parts = append(parts, part{
				val: string(sourceRunes[nextRuneIdx : i-1]),
			})
			parts = append(parts, part{
				val:    "%s",
				isVerb: true,
			})
			nextRuneIdx = i + 1
		}
		if r == '%' {
			percent = true
		} else {
			percent = false
		}
	}
	if nextRuneIdx < len(sourceRunes) {
		parts = append(parts, part{
			val: string(sourceRunes[nextRuneIdx:]),
		})
	}

	return &SplitConcatedString{
		parts: parts,
	}
}
