package ir

// Operator is a verified binary operator. Back ends can dispatch on its code.
type Operator uint8

const (
	opInvalid Operator = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpAndBits
	OpOrBits
	OpXor
	OpShiftLeft
	OpShiftRight
	OpEqual
	OpNotEqual
	OpLess
	OpLessEqual
	OpGreater
	OpGreaterEqual
	OpAnd
	OpOr
)

var spelling = [...]string{
	OpAdd:          "+",
	OpSub:          "-",
	OpMul:          "*",
	OpDiv:          "/",
	OpMod:          "%",
	OpAndBits:      "&",
	OpOrBits:       "|",
	OpXor:          "^",
	OpShiftLeft:    "<<",
	OpShiftRight:   ">>",
	OpEqual:        "==",
	OpNotEqual:     "!=",
	OpLess:         "<",
	OpLessEqual:    "<=",
	OpGreater:      ">",
	OpGreaterEqual: ">=",
	OpAnd:          "&&",
	OpOr:           "||",
}

func (op Operator) String() string {
	if int(op) >= len(spelling) || op == opInvalid {
		return "<invalid operator>"
	}
	return spelling[op]
}

func operator(s string) Operator {
	for op, text := range spelling {
		if text == s && op != 0 {
			return Operator(op)
		}
	}
	panic("ir: unknown binary operator " + s)
}
