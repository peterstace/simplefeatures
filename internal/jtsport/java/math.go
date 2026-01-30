package java

import "math"

// Round implements Java's Math.round() semantics: rounds to the nearest
// integer with ties going towards positive infinity. This differs from Go's
// math.Round() which rounds ties away from zero.
//
// Examples:
//
//	Round(1.5)    // 2 (same as Go)
//	Round(-1.5)   // -1 (Go returns -2)
//	Round(-1232.5) // -1232 (Go returns -1233)
func Round(val float64) float64 {
	return math.Floor(val + 0.5)
}

// AbsInt implements Java's Math.abs(int) for integers.
// Go's math.Abs() only works on float64, so this provides the integer version.
func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// CanonicalNaN is Java's canonical NaN value (Double.NaN).
// Go's math.NaN() may produce a different NaN bit pattern (0x7FF8000000000001)
// than Java's canonical NaN (0x7FF8000000000000). This difference matters for
// binary formats like WKB where byte-for-byte compatibility with Java is needed.
var CanonicalNaN = math.Float64frombits(0x7FF8000000000000)
