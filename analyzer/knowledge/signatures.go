package knowledge

// The contents of this file were borrowed from https://github.com/dominikh/go-tools/blob/master/knowledge/signatures.go.
// The following license, therefore, applies.

/*
Copyright (c) 2016 Dominik Honnef

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

import (
	"go/token"
	"go/types"
)

var Signatures = map[string]*types.Signature{
	"(fmt.Stringer).String": types.NewSignatureType(nil, nil, nil,
		types.NewTuple(),
		types.NewTuple(
			types.NewParam(token.NoPos, nil, "", types.Typ[types.String]),
		),
		false,
	),
}

var Interfaces = map[string]*types.Interface{
	"fmt.Stringer": types.NewInterfaceType(
		[]*types.Func{
			types.NewFunc(token.NoPos, nil, "String", Signatures["(fmt.Stringer).String"]),
		},
		nil,
	).Complete(),

	"error": types.Universe.Lookup("error").Type().Underlying().(*types.Interface),
}
