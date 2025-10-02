package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
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
		return processCallExpr(e)
	default:
		return nil, errors.New("not implemented")
	}
}

func processCallExpr(callExpr *ast.CallExpr) (ast.Expr, error) {
	selExpr, _ := callExpr.Fun.(*ast.SelectorExpr)
	if selExpr == nil {
		return callExpr, nil
	}

	xIdent, _ := selExpr.X.(*ast.Ident)
	if xIdent == nil {
		return callExpr, nil
	}
	if xIdent.Name != "fmt" {
		return callExpr, nil
	}

	if selExpr.Sel.Name != "Sprintf" {
		return callExpr, nil
	}

	if len(callExpr.Args) < 2 {
		panic("todo")
	}

	result, err := optimizeSprintf(callExpr)
	if err != nil {
		return nil, fmt.Errorf("optimize Sprintf call: %w", err)
	}

	return result, nil
}

func optimizeSprintf(callExpr *ast.CallExpr) (ast.Expr, error) {
	s := callExpr.Args[0].(*ast.BasicLit)

	if s.Kind != token.STRING {
		panic("todo")
		// TODO support any expression of type string
	}

	sVal, err := strconv.Unquote(s.Value)
	if err != nil {
		return nil, fmt.Errorf("unquote: %w", err)
	}

	splitConcatedStr := SplitConcat(sVal)

	result, err := splitConcatedStr.Fill(callExpr.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("fill: %w", err)
	}

	return result, nil
}

type SplitConcatedString struct {
	parts []part
}

type part struct {
	val    string
	isVerb bool
}

func (s *SplitConcatedString) Fill(args []ast.Expr) (ast.Expr, error) {
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
				aLit, _ := a.(*ast.BasicLit)
				if aLit == nil {
					return nil, errors.New("aLit == nil") // TODO
				}
				if aLit.Kind != token.INT {
					return nil, errors.New("aLit.Kind != token.INT") // TODO
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
							Value: aLit.Value,
						},
					},
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

func SplitConcat(source string) *SplitConcatedString {
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
		} else if slices.Contains(supportedVerbs, string(v)) {
			parts = append(parts, part{val: string(v), isVerb: true})
			v = v[:0]
		}

		v = append(v, r)
	}
	if len(v) > 0 {
		parts = append(parts, part{val: string(v)})
	}

	return &SplitConcatedString{
		parts: parts,
	}
}
