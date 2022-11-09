package geom

// operand represents either the first (A) or second (B) geometry in a binary
// operation (such as Union or Covers).
type operand int

const (
	operandA operand = 0
	operandB operand = 1
)

func forEachOperand(fn func(operand operand)) {
	fn(operandA)
	fn(operandB)
}

type location struct {
	interior bool
	boundary bool
}
