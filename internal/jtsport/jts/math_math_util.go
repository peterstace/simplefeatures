package jts

import "math"

// Math_MathUtil_ClampFloat64 clamps a float64 value to a given range.
func Math_MathUtil_ClampFloat64(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// Math_MathUtil_ClampInt clamps an int value to a given range.
func Math_MathUtil_ClampInt(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// Math_MathUtil_ClampMax clamps an integer to a given maximum limit.
func Math_MathUtil_ClampMax(x, max int) int {
	if x > max {
		return max
	}
	return x
}

// Math_MathUtil_Ceil computes the ceiling function of the dividend of two integers.
func Math_MathUtil_Ceil(num, denom int) int {
	div := num / denom
	if div*denom >= num {
		return div
	}
	return div + 1
}

var math_MathUtil_log10 = math.Log(10)

// Math_MathUtil_Log10 computes the base-10 logarithm of a float64 value.
//   - If the argument is NaN or less than zero, then the result is NaN.
//   - If the argument is positive infinity, then the result is positive infinity.
//   - If the argument is positive zero or negative zero, then the result is negative infinity.
func Math_MathUtil_Log10(x float64) float64 {
	ln := math.Log(x)
	if math.IsInf(ln, 0) {
		return ln
	}
	if math.IsNaN(ln) {
		return ln
	}
	return ln / math_MathUtil_log10
}

// Math_MathUtil_Wrap computes an index which wraps around a given maximum value.
// For values >= 0, this equals val % max.
// For values < 0, this equals max - (-val) % max.
func Math_MathUtil_Wrap(index, max int) int {
	if index < 0 {
		return max - ((-index) % max)
	}
	return index % max
}

// Math_MathUtil_Average computes the average of two numbers.
func Math_MathUtil_Average(x1, x2 float64) float64 {
	return (x1 + x2) / 2.0
}

// Math_MathUtil_Max3 returns the maximum of three values.
func Math_MathUtil_Max3(v1, v2, v3 float64) float64 {
	max := v1
	if v2 > max {
		max = v2
	}
	if v3 > max {
		max = v3
	}
	return max
}

// Math_MathUtil_Max4 returns the maximum of four values.
func Math_MathUtil_Max4(v1, v2, v3, v4 float64) float64 {
	max := v1
	if v2 > max {
		max = v2
	}
	if v3 > max {
		max = v3
	}
	if v4 > max {
		max = v4
	}
	return max
}

// Math_MathUtil_Min4 returns the minimum of four values.
func Math_MathUtil_Min4(v1, v2, v3, v4 float64) float64 {
	min := v1
	if v2 < min {
		min = v2
	}
	if v3 < min {
		min = v3
	}
	if v4 < min {
		min = v4
	}
	return min
}

// Math_MathUtil_PhiInv is the inverse of the Golden Ratio phi.
var Math_MathUtil_PhiInv = (math.Sqrt(5) - 1.0) / 2.0

// Math_MathUtil_Quasirandom generates a quasi-random sequence of numbers in the range [0,1].
// They are produced by an additive recurrence with 1/phi as the constant.
// This produces a low-discrepancy sequence which is more evenly
// distributed than random numbers.
//
// The sequence is initialized by calling it with any positive fractional number;
// 0 works well for most uses.
func Math_MathUtil_Quasirandom(curr float64) float64 {
	return Math_MathUtil_QuasirandomWithAlpha(curr, Math_MathUtil_PhiInv)
}

// Math_MathUtil_QuasirandomWithAlpha generates a quasi-random sequence of numbers in the range [0,1].
// They are produced by an additive recurrence with constant alpha.
// When alpha is irrational this produces a low discrepancy sequence
// which is more evenly distributed than random numbers.
//
// The sequence is initialized by calling it with any positive fractional number.
// 0 works well for most uses.
func Math_MathUtil_QuasirandomWithAlpha(curr, alpha float64) float64 {
	next := curr + alpha
	if next < 1 {
		return next
	}
	return next - math.Floor(next)
}

// Math_MathUtil_Shuffle generates a randomly-shuffled list of the integers from [0..n-1].
// One use is to randomize points inserted into a KDtree.
func Math_MathUtil_Shuffle(n int) []int {
	rnd := &math_lcgRandom{state: 13}
	ints := make([]int, n)
	for i := 0; i < n; i++ {
		ints[i] = i
	}
	for i := n - 1; i >= 1; i-- {
		j := rnd.nextInt(i + 1)
		last := ints[i]
		ints[i] = ints[j]
		ints[j] = last
	}
	return ints
}

// math_lcgRandom is a simple linear congruential generator to match Java's Random behavior.
type math_lcgRandom struct {
	state int64
}

func (r *math_lcgRandom) nextInt(bound int) int {
	// Java's Random uses a 48-bit LCG.
	r.state = (r.state*0x5DEECE66D + 0xB) & ((1 << 48) - 1)
	return int((r.state >> 17) % int64(bound))
}
