package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

func main() {
	// fset := token.NewFileSet()
	expr, err := parser.ParseExpr(`fmt.Sprintf("Hello, %s!", "Max")`)
	if err != nil {
		panic(err)
	}
	_ = expr
}

func ProcessFile(file *ast.File) {
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		newNode, ok := ProcessNode(c.Node())
		if !ok {
			return true
		}

		c.Replace(newNode)

		return true
	})
}

func ProcessNode(node ast.Node) (ast.Node, bool) {
	if node == nil {
		return nil, false
	}

	expr, _ := node.(ast.Expr)
	if expr == nil {
		return nil, false
	}

	newExpr, err := ProcessExpr(expr)
	if err != nil {
		return nil, false
	}

	return newExpr, true
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
