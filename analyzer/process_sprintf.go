package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"unicode/utf8"
)

func ProcessSprintfCall(typesInfo *types.Info, call *ast.CallExpr) (ast.Expr, bool) {
	analyzed, ok := analyzeSprintfCall(typesInfo, call)
	if !ok {
		return nil, false
	}

	result, ok := constructResult(analyzed)
	if !ok {
		return nil, false
	}

	return result, true
}

type analyzedSprintfCall struct {
	originalText string
	args         []sprintfArg
}

type sprintfArg struct {
	position       [2]int
	value          ast.Expr
	transformation transformation
}

func analyzeSprintfCall(typesInfo *types.Info, call *ast.CallExpr) (analyzedSprintfCall, bool) {
	// TODO: account for numbered placeholders (%[1]s, etc.)
	// TODO: account for escaping

	var zero analyzedSprintfCall

	if len(call.Args) < 1 {
		return zero, false
	}
	s, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return zero, false
	}

	verbArgs := call.Args[1:]
	if len(verbArgs) == 0 {
		// TODO: just use the string without fmt.Sprintf

		return zero, false
	}

	if s.Kind != token.STRING {
		// TODO support any expression of type string

		return zero, false
	}

	sprintfString, err := strconv.Unquote(s.Value)
	if err != nil {
		return zero, false
	}

	var entries []sprintfArg
	var verbBuf []rune

	for i, r := range sprintfString {
		if r == '%' {
			verbBuf = verbBuf[:0]
			verbBuf = append(verbBuf, '%')
		} else if len(verbBuf) > 0 {
			verbBuf = append(verbBuf, r)
		} else {
			continue
		}

		if isVerb(string(verbBuf)) {
			if len(entries) >= len(verbArgs) {
				return zero, false
			}

			verbArg := verbArgs[len(entries)]

			t, ok := resolveTransformation(typesInfo, verbArg, string(verbBuf))
			if !ok {
				return zero, false
			}

			entries = append(entries, sprintfArg{
				position:       [2]int{i - len(verbBuf) + 1, i + 1},
				transformation: t,
				value:          verbArg,
			})

			verbBuf = verbBuf[:0]
		}
	}

	return analyzedSprintfCall{
		originalText: sprintfString,
		args:         entries,
	}, true
}

func resolveTransformation(typesInfo *types.Info, arg ast.Expr, verb string) (transformation, bool) {
	dataType, ok := typesInfo.Types[arg]
	if !ok {
		return nil, false
	}
	if dataType.Type == nil {
		return nil, false
	}
	dataTypeName := dataType.Type.String()

	switch verb {
	case "%s":
		return resolveTransformationForSVerb(dataTypeName)
	case "%d":
		return resolveTransformationForDVerb(dataTypeName)
	default:
		// TODO: support more verbs
		return nil, false
	}
}

func resolveTransformationForSVerb(typeName string) (transformation, bool) {
	// TODO: check if is Stringer, then use String()

	switch typeName {
	case "string":
		return NoOp{}, true
	default:
		return nil, false
	}
}

func resolveTransformationForDVerb(typeName string) (transformation, bool) {
	switch typeName {
	case "int":
		return StrConv{F: StrConvFunc_Itoa}, true
	case "int64":
		return StrConv{F: StrConvFunc_FormatInt}, true
	case "uint64":
		return StrConv{F: StrConvFunc_FormatUint}, true
	// TODO: support more types
	default:
		return nil, false
	}
}

func constructResult(analyzed analyzedSprintfCall) (ast.Expr, bool) {
	if len(analyzed.args) == 0 {
		return nil, false
	}

	res := &ast.BinaryExpr{
		Op: token.ADD,
	}

	var cursor int

	for _, arg := range analyzed.args {
		head := analyzed.originalText[cursor:arg.position[0]]
		if head != "" {
			res = addExprToSum(res, &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(head),
			})
		}

		newValueExpr := transformValue(arg.value, arg.transformation)

		res = addExprToSum(res, newValueExpr)

		cursor = arg.position[1]
	}

	if cursor < utf8.RuneCountInString(analyzed.originalText) {
		res = addExprToSum(res, &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(analyzed.originalText[cursor:]),
		})
	}

	return res, true
}

func addExprToSum(base *ast.BinaryExpr, e ast.Expr) *ast.BinaryExpr {
	if base.X == nil {
		base.X = e
	} else if base.Y == nil {
		base.Y = e
	} else {
		base = &ast.BinaryExpr{X: base, Y: e, Op: token.ADD}
	}

	return base
}

func transformValue(value ast.Expr, t transformation) ast.Expr {
	switch tt := t.(type) {
	case NoOp:
		return value
	case CallStringMethod:
		panic("unimplemented") // TODO
	case ConvertToType:
		panic("unimplemented") // TODO
	case StrConv:
		return transformValueWithStrConv(value, tt)
	default:
		panic("unknown transformation")
	}
}

func transformValueWithStrConv(value ast.Expr, tStrConv StrConv) ast.Expr {
	// TODO: point to actual strconv object?

	switch tStrConv.F {
	case StrConvFunc_Itoa:
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "strconv"},
				Sel: &ast.Ident{Name: "Itoa"},
			},
			Args: []ast.Expr{value},
		}
	case StrConvFunc_FormatInt:
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "strconv"},
				Sel: &ast.Ident{Name: "FormatInt"},
			},
			Args: []ast.Expr{value, &ast.BasicLit{Value: "10", Kind: token.INT}},
		}
	case StrConvFunc_FormatUint:
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "strconv"},
				Sel: &ast.Ident{Name: "FormatUint"},
			},
			Args: []ast.Expr{value, &ast.BasicLit{Value: "10", Kind: token.INT}},
		}
	default:
		panic("unknown strconv operation")
	}
}
