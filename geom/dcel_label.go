package geom

// operand represents either the first (A) or second (B) geometry in a binary
// operation (such as Union or Covers).
type operand int

const (
	operandA operand = 0
	operandB operand = 1
)

func forEachOperand(fn func(operand)) {
	fn(operandA)
	fn(operandB)
}

type location struct {
	interior bool
	boundary bool
}

func selectUnion(inSet [2]bool) bool {
	return inSet[0] || inSet[1]
}

func selectIntersection(inSet [2]bool) bool {
	return inSet[0] && inSet[1]
}

func selectDifference(inSet [2]bool) bool {
	return inSet[0] && !inSet[1]
}

func selectSymmetricDifference(inSet [2]bool) bool {
	return inSet[0] != inSet[1]
}
