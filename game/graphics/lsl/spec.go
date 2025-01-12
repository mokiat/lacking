package lsl

var (
	operatorChars = []rune{
		'{', '}', '=', '(', ')', ',', '-', ';', '+', '*',
		'/', '%', '!', '<', '>', '&', '|', '^', '.', ':',
	}
)

const (
	// AssignmentOperatorEq is the assignment operator "=". It assigns the
	// value to the variable.
	AssignmentOperatorEq = "="

	// AssignmentOperatorAuto is the assignment operator ":=", where the type
	// of the variable is inferred from the value.
	AssignmentOperatorAuto = ":="

	// AssignmentOperatorAdd is the assignment operator "+=". It adds the value
	// to the variable.
	AssignmentOperatorAdd = "+="

	// AssignmentOperatorSub is the assignment operator "-=". It subtracts the
	// value from the variable.
	AssignmentOperatorSub = "-="

	// AssignmentOperatorMul is the assignment operator "*=". It multiplies the
	// variable by the value.
	AssignmentOperatorMul = "*="

	// AssignmentOperatorDiv is the assignment operator "/=". It divides the
	// variable by the value.
	AssignmentOperatorDiv = "/="

	// AssignmentOperatorMod is the assignment operator "%=". It takes the
	// modulo of the variable and the value.
	AssignmentOperatorMod = "%="

	// AssignmentOperatorShl is the assignment operator "<<=". It shifts the
	// variable to the left by the value.
	AssignmentOperatorShl = "<<="

	// AssignmentOperatorShr is the assignment operator ">>=". It shifts the
	// variable to the right by the value.
	AssignmentOperatorShr = ">>="

	// AssignmentOperatorAnd is the assignment operator "&=". It performs a
	// bitwise AND operation on the variable and the value.
	AssignmentOperatorAnd = "&="

	// AssignmentOperatorOr is the assignment operator "|=". It performs a
	// bitwise OR operation on the variable and the value.
	AssignmentOperatorOr = "|="

	// AssignmentOperatorXor is the assignment operator "^=". It performs a
	// bitwise XOR operation on the variable and the value.
	AssignmentOperatorXor = "^="
)

var assignmentOperators = []string{
	AssignmentOperatorEq,
	AssignmentOperatorAuto,
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
	// UnaryOperatorNot is the unary operator "!". It inverts the value.
	UnaryOperatorNot = "!"

	// UnaryOperatorNeg is the unary operator "-". It negates the value.
	UnaryOperatorNeg = "-"

	// UnaryOperatorPos is the unary operator "+". It is a no-op.
	UnaryOperatorPos = "+"

	// UnaryOperatorBitNot is the unary operator "^". It performs a bitwise
	// NOT operation on the value.
	UnaryOperatorBitNot = "^"
)

var unaryOperators = []string{
	UnaryOperatorNot,
	UnaryOperatorNeg,
	UnaryOperatorPos,
	UnaryOperatorBitNot,
}

const (
	// BinaryOperatorAdd is the binary operator "+". It adds two values.
	BinaryOperatorAdd = "+"

	// BinaryOperatorSub is the binary operator "-". It subtracts two values.
	BinaryOperatorSub = "-"

	// BinaryOperatorMul is the binary operator "*". It multiplies two values.
	BinaryOperatorMul = "*"

	// BinaryOperatorDiv is the binary operator "/". It divides two values.
	BinaryOperatorDiv = "/"

	// BinaryOperatorMod is the binary operator "%". It takes the modulo of two
	// values.
	BinaryOperatorMod = "%"

	// BinaryOperatorShl is the binary operator "<<". It shifts the first value
	// to the left by the second value.
	BinaryOperatorShl = "<<"

	// BinaryOperatorShr is the binary operator ">>". It shifts the first value
	// to the right by the second value.
	BinaryOperatorShr = ">>"

	// BinaryOperatorEq is the binary operator "==". It checks if two values are
	// equal.
	BinaryOperatorEq = "=="

	// BinaryOperatorNotEq is the binary operator "!=". It checks if two values
	// are not equal.
	BinaryOperatorNotEq = "!="

	// BinaryOperatorLess is the binary operator "<". It checks if the first
	// value is less than the second value.
	BinaryOperatorLess = "<"

	// BinaryOperatorGreater is the binary operator ">". It checks if the first
	// value is greater than the second value.
	BinaryOperatorGreater = ">"

	// BinaryOperatorLessEq is the binary operator "<=". It checks if the first
	// value is less than or equal to the second value.
	BinaryOperatorLessEq = "<="

	// BinaryOperatorGreaterEq is the binary operator ">=". It checks if the
	// first value is greater than or equal to the second value.
	BinaryOperatorGreaterEq = ">="

	// BinaryOperatorBitAnd is the binary operator "&". It performs a bitwise
	// AND operation on two values.
	BinaryOperatorBitAnd = "&"

	// BinaryOperatorBitOr is the binary operator "|". It performs a bitwise OR
	// operation on two values.
	BinaryOperatorBitOr = "|"

	// BinaryOperatorBitXor is the binary operator "^". It performs a bitwise
	// XOR operation on two values.
	BinaryOperatorBitXor = "^"

	// BinaryOperatorAnd is the binary operator "&&". It performs a logical AND
	// operation on two values.
	BinaryOperatorAnd = "&&"

	// BinaryOperatorOr is the binary operator "||". It performs a logical OR
	// operation on two values.
	BinaryOperatorOr = "||"
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
