package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"unicode/utf8"

	"github.com/m-ocean-it/go-sprintf-bomb/analyzer/internal/strconvs"
	"github.com/m-ocean-it/go-sprintf-bomb/analyzer/internal/transform"
	"github.com/m-ocean-it/go-sprintf-bomb/analyzer/knowledge"
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
	transformation transform.Transformation
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

			t := resolveTransformation(typesInfo, verbArg, string(verbBuf))
			if t == nil {
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

func resolveTransformation(typesInfo *types.Info, arg ast.Expr, verb string) transform.Transformation {
	dataType, ok := typesInfo.Types[arg]
	if !ok {
		return nil
	}
	if dataType.Type == nil {
		return nil
	}
	dataTypeName := dataType.Type.String()

	switch verb {
	case "%s":
		return resolveTransformationForSVerb(dataType.Type)
	case "%d":
		return resolveTransformationForDVerb(dataTypeName)
	case "%f":
		return resolveTransformationForFVerb(dataTypeName, verb)
	default:
		// TODO: support more verbs
		return nil
	}
}

func resolveTransformationForSVerb(t types.Type) transform.Transformation {
	// TODO: check, also, types.Implements(T, knowledge.Interfaces["error"] and return CallErrorMethod if does

	if types.Implements(t, knowledge.Interfaces["fmt.Stringer"]) {
		return transform.CallStringMethod{}
	}

	switch t.String() {
	case "string": // TODO: check, also, if underlying type is string
		return transform.NoOp{}
	default:
		return nil
	}
}

func resolveTransformationForDVerb(typeName string) transform.Transformation {
	switch typeName {
	case "int":
		return transform.StrConv{Op: strconvs.Itoa{}}
	case "int64":
		return transform.StrConv{Op: strconvs.FormatInt{}}
	case "uint64":
		return transform.StrConv{Op: strconvs.FormatUint{}}
	// TODO: support more types
	default:
		return nil
	}
}

func resolveTransformationForFVerb(typeName string, verb string) transform.Transformation {
	var castToFloat64 bool

	switch typeName {
	case "float64":
		castToFloat64 = false
	case "float32":
		castToFloat64 = true
	default:
		return nil
	}

	fmt, prec, ok := getFmtAndPrecFromVerb(verb)
	if !ok {
		return nil
	}

	return transform.StrConv{Op: strconvs.FormatFloat{
		Fmt:           fmt,
		Prec:          prec,
		CastToFloat64: castToFloat64,
	}}
}

func getFmtAndPrecFromVerb(verb string) (byte, int, bool) {
	switch verb {
	case "%f":
		// The special precision -1 uses the smallest number of digits necessary
		// such that ParseFloat will return f exactly.
		return 'f', -1, true
	// TODO: parse more floating-point verbs
	default:
		return 0, 0, false
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

func transformValue(value ast.Expr, t transform.Transformation) ast.Expr {
	switch tt := t.(type) {
	case transform.NoOp:
		return value
	case transform.CallStringMethod:
		return transformValueToCallStringMethod(value)
	case transform.ConvertToType:
		panic("unimplemented") // TODO
	case transform.StrConv:
		return transformValueWithStrConv(value, tt)
	default:
		panic("unknown transformation")
	}
}

func transformValueWithStrConv(value ast.Expr, tStrConv transform.StrConv) ast.Expr {
	// TODO: point to actual strconv object? or at least dedupe strconv-ident pointers?

	switch op := tStrConv.Op.(type) {
	case strconvs.Itoa:
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "strconv"},
				Sel: &ast.Ident{Name: "Itoa"},
			},
			Args: []ast.Expr{value},
		}
	case strconvs.FormatInt:
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "strconv"},
				Sel: &ast.Ident{Name: "FormatInt"},
			},
			Args: []ast.Expr{value, &ast.BasicLit{Value: "10", Kind: token.INT}},
		}
	case strconvs.FormatUint:
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "strconv"},
				Sel: &ast.Ident{Name: "FormatUint"},
			},
			Args: []ast.Expr{value, &ast.BasicLit{Value: "10", Kind: token.INT}},
		}
	case strconvs.FormatFloat:
		val := value
		if op.CastToFloat64 {
			val = &ast.CallExpr{Fun: &ast.Ident{Name: "float64"}, Args: []ast.Expr{val}}
		}

		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "strconv"},
				Sel: &ast.Ident{Name: "FormatFloat"},
			},
			Args: []ast.Expr{
				val,
				&ast.BasicLit{Value: strconv.QuoteRune(rune(op.Fmt)), Kind: token.CHAR},
				&ast.BasicLit{Value: strconv.Itoa(op.Prec), Kind: token.INT},
				&ast.BasicLit{Value: "64", Kind: token.INT},
			},
		}
	default:
		panic("unknown strconv operation")
	}
}

func transformValueToCallStringMethod(value ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X: value,
			// X:   &ast.Ident{Name: "strconv"}, // FIXME: remove
			Sel: &ast.Ident{Name: "String"},
		},
	}
}
