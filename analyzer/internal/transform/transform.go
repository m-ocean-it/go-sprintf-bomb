package transform

import "github.com/m-ocean-it/go-sprintf-bomb/analyzer/internal/strconvs"

type Transformation interface {
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
	Op strconvs.Op
}

func (s StrConv) isTransformation() {}

// TODO: add CallErrorMethod
