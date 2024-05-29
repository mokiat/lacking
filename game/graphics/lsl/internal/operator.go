package internal

var (
	operatorChars = []rune{
		'{', '}', '=', '(', ')', ',', '-', ';', '+', '*',
		'/', '%', '!', '<', '>', '&', '|', '^', '.',
	}
)

const (
	AssignmentOperatorEq  = "="
	AssignmentOperatorAdd = "+="
	AssignmentOperatorSub = "-="
	AssignmentOperatorMul = "*="
	AssignmentOperatorDiv = "/="
	AssignmentOperatorMod = "%="
	AssignmentOperatorShl = "<<="
	AssignmentOperatorShr = ">>="
	AssignmentOperatorAnd = "&="
	AssignmentOperatorOr  = "|="
	AssignmentOperatorXor = "^="
)

var assignmentOperators = []string{
	AssignmentOperatorEq,
	AssignmentOperatorAdd,
	AssignmentOperatorSub,
	AssignmentOperatorMul,
	AssignmentOperatorDiv,
	AssignmentOperatorMod,
	AssignmentOperatorShl,
	AssignmentOperatorShr,
	AssignmentOperatorAnd,
	AssignmentOperatorOr,
	AssignmentOperatorXor,
}

const (
	UnaryOperatorNot    = "!"
	UnaryOperatorNeg    = "-"
	UnaryOperatorPos    = "+"
	UnaryOperatorBitNot = "^"
)

var unaryOperators = []string{
	UnaryOperatorNot,
	UnaryOperatorNeg,
	UnaryOperatorPos,
	UnaryOperatorBitNot,
}

const (
	BinaryOperatorAdd       = "+"
	BinaryOperatorSub       = "-"
	BinaryOperatorMul       = "*"
	BinaryOperatorDiv       = "/"
	BinaryOperatorMod       = "%"
	BinaryOperatorShl       = "<<"
	BinaryOperatorShr       = ">>"
	BinaryOperatorEq        = "=="
	BinaryOperatorNotEq     = "!="
	BinaryOperatorLess      = "<"
	BinaryOperatorGreater   = ">"
	BinaryOperatorLessEq    = "<="
	BinaryOperatorGreaterEq = ">="
	BinaryOperatorBitAnd    = "&"
	BinaryOperatorBitOr     = "|"
	BinaryOperatorBitXor    = "^"
	BinaryOperatorAnd       = "&&"
	BinaryOperatorOr        = "||"
)

var binaryOperators = []string{
	BinaryOperatorAdd,
	BinaryOperatorSub,
	BinaryOperatorMul,
	BinaryOperatorDiv,
	BinaryOperatorMod,
	BinaryOperatorShl,
	BinaryOperatorShr,
	BinaryOperatorEq,
	BinaryOperatorNotEq,
	BinaryOperatorLess,
	BinaryOperatorGreater,
	BinaryOperatorLessEq,
	BinaryOperatorGreaterEq,
	BinaryOperatorBitAnd,
	BinaryOperatorBitOr,
	BinaryOperatorBitXor,
	BinaryOperatorAnd,
	BinaryOperatorOr,
}
