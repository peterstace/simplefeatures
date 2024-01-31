package geom

import "math"

func snapToGridXY(dp int) func(XY) XY {
	return func(pt XY) XY {
		return XY{
			snapToGridFloat64(pt.X, dp),
			snapToGridFloat64(pt.Y, dp),
		}
	}
}

// NOTE: This function would naively be implemented as:
//
//	return math.Round(f * math.Pow10(dp)) / math.Pow10(dp)
//
// However, this approach would causes two problems (which are fixed by having
// a slightly more complex implementation):
//
// 1. In floating point math, numbers of the form 10^dp can be represented
// exactly for values of dp such that 0 <= dp <= 15 (i.e. those less than
// 2^53). Numbers of the form 10^dp can never be represented exactly for
// negative values of dp (since the fractional part is recurring in base 2). To
// remedy this, the function is split into "positive", "negative", and "zero"
// dp cases.
//
// 2. For large values of dp, the input could overflow or underflow after being
// multiplied by the scale factor. This causes the wrong result when the scale
// factor is multiplied or divided out after rounding. This can be remedied by
// detecting this case and returning the input unaltered (for an overflow) or
// as zero (for an underflow).
func snapToGridFloat64(f float64, dp int) float64 {
	switch {
	case dp > 0:
		scale := math.Pow10(dp)
		scaled := f * scale
		if scaled > math.MaxFloat64 {
			return f
		}
		return math.Round(scaled) / scale
	case dp < 0:
		scale := math.Pow10(-dp)
		scaled := f / scale
		if scaled == 0 {
			return 0
		}
		return math.Round(scaled) * scale
	default:
		return math.Round(f)
	}
}
