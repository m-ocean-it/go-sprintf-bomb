package transform

import "github.com/m-ocean-it/go-sprintf-bomb/analyzer/internal/strconvs"

type Transformation interface {
	isTransformation()
}

type NoOp struct{}

func (n NoOp) isTransformation() {}

type CallStringMethod struct{}

func (c CallStringMethod) isTransformation() {}

type CallErrorMethod struct{}

func (c CallErrorMethod) isTransformation() {}

type Wrap struct {
	Wrapper string
}

func (c Wrap) isTransformation() {}

type StrConv struct {
	Op strconvs.Op
}

func (s StrConv) isTransformation() {}
