package strconvs

type Op interface {
	isOp()
}

type Itoa struct{}

func (i Itoa) isOp() {}

type FormatInt struct {
	CastToInt64 bool
}

func (f FormatInt) isOp() {}

type FormatUint struct {
	CastToUint64 bool
}

func (f FormatUint) isOp() {}

type FormatFloat struct {
	// The format fmt is one of
	//   - 'b' (-ddddp±ddd, a binary exponent),
	//   - 'e' (-d.dddde±dd, a decimal exponent),
	//   - 'E' (-d.ddddE±dd, a decimal exponent),
	//   - 'f' (-ddd.dddd, no exponent),
	//   - 'g' ('e' for large exponents, 'f' otherwise),
	//   - 'G' ('E' for large exponents, 'f' otherwise),
	//   - 'x' (-0xd.ddddp±ddd, a hexadecimal fraction and binary exponent), or
	//   - 'X' (-0Xd.ddddP±ddd, a hexadecimal fraction and binary exponent).
	Fmt byte

	Prec int

	CastToFloat64 bool
}

func (f FormatFloat) isOp() {}
