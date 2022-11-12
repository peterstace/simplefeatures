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

func mergeBools(dst *[2]bool, src [2]bool) {
	dst[0] = dst[0] || src[0]
	dst[1] = dst[1] || src[1]
}

type location struct {
	interior bool
	boundary bool
}

func mergeLocations(dst *[2]location, src [2]location) {
	dst[0].interior = dst[0].interior || src[0].interior
	dst[1].interior = dst[1].interior || src[1].interior
	dst[0].boundary = dst[0].boundary || src[0].boundary
	dst[1].boundary = dst[1].boundary || src[1].boundary
}

func newLocationsOnBoundary(operand operand) [2]location {
	var locs [2]location
	locs[operand].boundary = true
	return locs
}
