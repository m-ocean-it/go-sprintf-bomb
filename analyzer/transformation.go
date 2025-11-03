package analyzer

type StrConvFunc int

const (
	StrConvFunc_Itoa StrConvFunc = iota + 1
	StrConvFunc_FormatInt
	StrConvFunc_FormatUint
)

type transformation interface {
	isTransformation()
}

type NoOp struct{}

func (n NoOp) isTransformation() {}

type CallStringMethod struct{}

func (c CallStringMethod) isTransformation() {}

type ConvertToType struct {
	Type any // FIXME: specify correct type
}

func (c ConvertToType) isTransformation() {}

type StrConv struct {
	F StrConvFunc
}

func (s StrConv) isTransformation() {}
