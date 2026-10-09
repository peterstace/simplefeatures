package java

import (
	crand "crypto/rand"
	"encoding/binary"
	"math"
)

// This file implements java.lang.Math and java.lang.StrictMath. Both follow
// JDK 21's StrictMath (see Runtime Semantics in internal/jts/DESIGN.md): the
// transcendental functions use the fdlibm port in fdlibm.go, and the rest port
// the JDK's Java code.
//
// Go's math package is only used for operations whose result IEEE 754 fully
// specifies (sqrt, floor, ceil, rint, abs).
//
// Every floating-point product is wrapped in an explicit float64 or float32
// conversion so that Go cannot fuse it with an addition (see fdlibm.go).

// Math_E is Math.E.
const Math_E float64 = 2.718281828459045

// Math_PI is Math.PI.
const Math_PI float64 = 3.141592653589793

// StrictMath_E is StrictMath.E.
const StrictMath_E float64 = Math_E

// StrictMath_PI is StrictMath.PI.
const StrictMath_PI float64 = Math_PI

// Math_sin implements Math.sin.
func Math_sin(a float64) float64 { return fdlibmSin(a) }

// Math_cos implements Math.cos.
func Math_cos(a float64) float64 { return fdlibmCos(a) }

// Math_tan implements Math.tan.
func Math_tan(a float64) float64 { return fdlibmTan(a) }

// Math_asin implements Math.asin.
func Math_asin(a float64) float64 { return fdlibmAsin(a) }

// Math_acos implements Math.acos.
func Math_acos(a float64) float64 { return fdlibmAcos(a) }

// Math_atan implements Math.atan.
func Math_atan(a float64) float64 { return fdlibmAtan(a) }

// Math_toRadians implements Math.toRadians, which multiplies by a rounded
// constant rather than computing angdeg / 180 * PI.
func Math_toRadians(angdeg float64) float64 {
	const degreesToRadians = 0.017453292519943295
	return float64(angdeg * degreesToRadians)
}

// Math_toDegrees implements Math.toDegrees, which multiplies by a rounded
// constant rather than computing angrad * 180 / PI.
func Math_toDegrees(angrad float64) float64 {
	const radiansToDegrees = 57.29577951308232
	return float64(angrad * radiansToDegrees)
}

// Math_exp implements Math.exp.
func Math_exp(a float64) float64 { return fdlibmExp(a) }

// Math_log implements Math.log.
func Math_log(a float64) float64 { return fdlibmLog(a) }

// Math_log10 implements Math.log10.
func Math_log10(a float64) float64 { return fdlibmLog10(a) }

// Math_sqrt implements Math.sqrt (correctly rounded).
func Math_sqrt(a float64) float64 { return math.Sqrt(a) }

// Math_cbrt implements Math.cbrt.
func Math_cbrt(a float64) float64 { return fdlibmCbrt(a) }

// Math_IEEEremainder implements Math.IEEEremainder.
func Math_IEEEremainder(f1, f2 float64) float64 { return fdlibmIEEEremainder(f1, f2) }

// Math_ceil implements Math.ceil.
func Math_ceil(a float64) float64 { return math.Ceil(a) }

// Math_floor implements Math.floor.
func Math_floor(a float64) float64 { return math.Floor(a) }

// Math_rint implements Math.rint: round half to even, keeping the sign of
// zero.
func Math_rint(a float64) float64 { return math.RoundToEven(a) }

// Math_atan2 implements Math.atan2.
func Math_atan2(y, x float64) float64 { return fdlibmAtan2(y, x) }

// Math_pow implements Math.pow.
func Math_pow(a, b float64) float64 { return fdlibmPow(a, b) }

// Math_round_float implements Math.round(float): floor(a + 1/2) computed
// without rounding error, with NaN mapping to 0 and saturation at the int
// range.
func Math_round_float(a float32) int32 {
	const (
		significandWidth = 24
		expBias          = 127
		expBitMask       = 0x7f800000
		signifBitMask    = 0x007fffff
	)
	intBits := int32(math.Float32bits(a))
	biasedExp := (intBits & expBitMask) >> (significandWidth - 1)
	shift := (significandWidth - 2 + expBias) - biasedExp
	if (shift & -32) == 0 { // shift >= 0 && shift < 32
		// a is a finite number such that pow(2,-32) <= ulp(a) < 1.
		r := (intBits & signifBitMask) | (signifBitMask + 1)
		if intBits < 0 {
			r = -r
		}
		return ((r >> shift) + 1) >> 1
	}
	// a is a finite number with abs(a) < 1/2, a finite number with
	// ulp(a) >= 1 (and hence a mathematical integer), an infinity, or NaN.
	return F2I(float64(a))
}

// Math_round_double implements Math.round(double): floor(a + 1/2) computed
// without rounding error, with NaN mapping to 0 and saturation at the long
// range.
func Math_round_double(a float64) int64 {
	const (
		significandWidth = 53
		expBias          = 1023
		expBitMask       = 0x7ff0000000000000
		signifBitMask    = 0x000fffffffffffff
	)
	longBits := int64(math.Float64bits(a))
	biasedExp := (longBits & expBitMask) >> (significandWidth - 1)
	shift := (significandWidth - 2 + expBias) - biasedExp
	if (shift & -64) == 0 { // shift >= 0 && shift < 64
		// a is a finite number such that pow(2,-64) <= ulp(a) < 1.
		r := (longBits & signifBitMask) | (signifBitMask + 1)
		if longBits < 0 {
			r = -r
		}
		return ((r >> shift) + 1) >> 1
	}
	// a is a finite number with abs(a) < 1/2, a finite number with
	// ulp(a) >= 1 (and hence a mathematical integer), an infinity, or NaN.
	return F2L(a)
}

// Math_random implements Math.random. As in Java, the sequence is not
// reproducible: each value is a uniformly distributed multiple of 2^-53 in
// [0, 1).
func Math_random() float64 {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		panic(err)
	}
	return float64(binary.LittleEndian.Uint64(b[:])>>11) * 0x1.0p-53
}

// Math_addExact_int_int implements Math.addExact(int, int).
func Math_addExact_int_int(x, y int32) int32 {
	r := x + y
	if ((x ^ r) & (y ^ r)) < 0 {
		panic(ArithmeticException_new_String(Lit("integer overflow")))
	}
	return r
}

// Math_addExact_long_long implements Math.addExact(long, long).
func Math_addExact_long_long(x, y int64) int64 {
	r := x + y
	if ((x ^ r) & (y ^ r)) < 0 {
		panic(ArithmeticException_new_String(Lit("long overflow")))
	}
	return r
}

// Math_subtractExact_int_int implements Math.subtractExact(int, int).
func Math_subtractExact_int_int(x, y int32) int32 {
	r := x - y
	if ((x ^ y) & (x ^ r)) < 0 {
		panic(ArithmeticException_new_String(Lit("integer overflow")))
	}
	return r
}

// Math_subtractExact_long_long implements Math.subtractExact(long, long).
func Math_subtractExact_long_long(x, y int64) int64 {
	r := x - y
	if ((x ^ y) & (x ^ r)) < 0 {
		panic(ArithmeticException_new_String(Lit("long overflow")))
	}
	return r
}

// Math_multiplyExact_int_int implements Math.multiplyExact(int, int).
func Math_multiplyExact_int_int(x, y int32) int32 {
	r := int64(x) * int64(y)
	if int64(int32(r)) != r {
		panic(ArithmeticException_new_String(Lit("integer overflow")))
	}
	return int32(r)
}

// Math_multiplyExact_long_long implements Math.multiplyExact(long, long).
func Math_multiplyExact_long_long(x, y int64) int64 {
	r := x * y
	ax := Math_abs_long(x)
	ay := Math_abs_long(y)
	if uint64(ax|ay)>>31 != 0 {
		// Some bits greater than 2^31 that might cause overflow. Check the
		// result using the divide operator and check for the special case of
		// Long.MIN_VALUE * -1.
		if (y != 0 && r/y != x) || (x == math.MinInt64 && y == -1) {
			panic(ArithmeticException_new_String(Lit("long overflow")))
		}
	}
	return r
}

// Math_incrementExact_int implements Math.incrementExact(int).
func Math_incrementExact_int(a int32) int32 {
	if a == math.MaxInt32 {
		panic(ArithmeticException_new_String(Lit("integer overflow")))
	}
	return a + 1
}

// Math_incrementExact_long implements Math.incrementExact(long).
func Math_incrementExact_long(a int64) int64 {
	if a == math.MaxInt64 {
		panic(ArithmeticException_new_String(Lit("long overflow")))
	}
	return a + 1
}

// Math_decrementExact_int implements Math.decrementExact(int).
func Math_decrementExact_int(a int32) int32 {
	if a == math.MinInt32 {
		panic(ArithmeticException_new_String(Lit("integer overflow")))
	}
	return a - 1
}

// Math_decrementExact_long implements Math.decrementExact(long).
func Math_decrementExact_long(a int64) int64 {
	if a == math.MinInt64 {
		panic(ArithmeticException_new_String(Lit("long overflow")))
	}
	return a - 1
}

// Math_negateExact_int implements Math.negateExact(int).
func Math_negateExact_int(a int32) int32 {
	if a == math.MinInt32 {
		panic(ArithmeticException_new_String(Lit("integer overflow")))
	}
	return -a
}

// Math_negateExact_long implements Math.negateExact(long).
func Math_negateExact_long(a int64) int64 {
	if a == math.MinInt64 {
		panic(ArithmeticException_new_String(Lit("long overflow")))
	}
	return -a
}

// Math_toIntExact implements Math.toIntExact.
func Math_toIntExact(value int64) int32 {
	if int64(int32(value)) != value {
		panic(ArithmeticException_new_String(Lit("integer overflow")))
	}
	return int32(value)
}

// Math_floorDiv_int_int implements Math.floorDiv(int, int). Division by zero
// panics with Go's runtime error, which the runtime converts to Java's
// ArithmeticException.
func Math_floorDiv_int_int(x, y int32) int32 {
	q := x / y
	// If the signs are different and modulo not zero, round down.
	if (x^y) < 0 && q*y != x {
		return q - 1
	}
	return q
}

// Math_floorDiv_long_long implements Math.floorDiv(long, long).
func Math_floorDiv_long_long(x, y int64) int64 {
	q := x / y
	// If the signs are different and modulo not zero, round down.
	if (x^y) < 0 && q*y != x {
		return q - 1
	}
	return q
}

// Math_floorMod_int_int implements Math.floorMod(int, int).
func Math_floorMod_int_int(x, y int32) int32 {
	r := x % y
	// If the signs are different and modulo not zero, adjust result.
	if (x^y) < 0 && r != 0 {
		return r + y
	}
	return r
}

// Math_floorMod_long_long implements Math.floorMod(long, long).
func Math_floorMod_long_long(x, y int64) int64 {
	r := x % y
	// If the signs are different and modulo not zero, adjust result.
	if (x^y) < 0 && r != 0 {
		return r + y
	}
	return r
}

// Math_abs_int implements Math.abs(int); Integer.MIN_VALUE maps to itself.
func Math_abs_int(a int32) int32 {
	if a < 0 {
		return -a
	}
	return a
}

// Math_abs_long implements Math.abs(long); Long.MIN_VALUE maps to itself.
func Math_abs_long(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

// Math_abs_float implements Math.abs(float) by clearing the sign bit.
func Math_abs_float(a float32) float32 {
	return math.Float32frombits(math.Float32bits(a) &^ (1 << 31))
}

// Math_abs_double implements Math.abs(double) by clearing the sign bit.
func Math_abs_double(a float64) float64 {
	return math.Abs(a)
}

// Math_max_int_int implements Math.max(int, int).
func Math_max_int_int(a, b int32) int32 {
	if a >= b {
		return a
	}
	return b
}

// Math_max_long_long implements Math.max(long, long).
func Math_max_long_long(a, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}

// Math_max_float_float implements Math.max(float, float): NaN if either
// argument is NaN, and 0.0f is greater than -0.0f.
func Math_max_float_float(a, b float32) float32 {
	if a != a {
		return a
	}
	if a == 0 && b == 0 && math.Float32bits(a) == 1<<31 {
		return b
	}
	if a >= b {
		return a
	}
	return b
}

// Math_max_double_double implements Math.max(double, double): NaN if either
// argument is NaN, and 0.0 is greater than -0.0.
func Math_max_double_double(a, b float64) float64 {
	if a != a {
		return a
	}
	if a == 0 && b == 0 && math.Float64bits(a) == 1<<63 {
		return b
	}
	if a >= b {
		return a
	}
	return b
}

// Math_min_int_int implements Math.min(int, int).
func Math_min_int_int(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
}

// Math_min_long_long implements Math.min(long, long).
func Math_min_long_long(a, b int64) int64 {
	if a <= b {
		return a
	}
	return b
}

// Math_min_float_float implements Math.min(float, float): NaN if either
// argument is NaN, and -0.0f is less than 0.0f.
func Math_min_float_float(a, b float32) float32 {
	if a != a {
		return a
	}
	if a == 0 && b == 0 && math.Float32bits(b) == 1<<31 {
		return b
	}
	if a <= b {
		return a
	}
	return b
}

// Math_min_double_double implements Math.min(double, double): NaN if either
// argument is NaN, and -0.0 is less than 0.0.
func Math_min_double_double(a, b float64) float64 {
	if a != a {
		return a
	}
	if a == 0 && b == 0 && math.Float64bits(b) == 1<<63 {
		return b
	}
	if a <= b {
		return a
	}
	return b
}

// Math_ulp_double implements Math.ulp(double).
func Math_ulp_double(d float64) float64 {
	exp := Math_getExponent_double(d)
	switch exp {
	case 1023 + 1: // NaN or infinity
		return math.Abs(d)
	case -1022 - 1: // zero or subnormal
		return 0x1p-1074
	}
	exp -= 53 - 1
	if exp >= -1022 {
		return mathPowerOfTwoD(exp)
	}
	// Return a subnormal result.
	return math.Float64frombits(1 << (exp - (-1022 - (53 - 1))))
}

// Math_ulp_float implements Math.ulp(float).
func Math_ulp_float(f float32) float32 {
	exp := Math_getExponent_float(f)
	switch exp {
	case 127 + 1: // NaN or infinity
		return Math_abs_float(f)
	case -126 - 1: // zero or subnormal
		return 0x1p-149
	}
	exp -= 24 - 1
	if exp >= -126 {
		return mathPowerOfTwoF(exp)
	}
	// Return a subnormal result.
	return math.Float32frombits(1 << (exp - (-126 - (24 - 1))))
}

// Math_signum_double implements Math.signum(double): zeros and NaN are
// returned unchanged.
func Math_signum_double(d float64) float64 {
	if d == 0 || d != d {
		return d
	}
	return Math_copySign_double_double(1.0, d)
}

// Math_signum_float implements Math.signum(float): zeros and NaN are returned
// unchanged.
func Math_signum_float(f float32) float32 {
	if f == 0 || f != f {
		return f
	}
	return Math_copySign_float_float(1.0, f)
}

// Math_sinh implements Math.sinh.
func Math_sinh(x float64) float64 { return fdlibmSinh(x) }

// Math_cosh implements Math.cosh.
func Math_cosh(x float64) float64 { return fdlibmCosh(x) }

// Math_tanh implements Math.tanh.
func Math_tanh(x float64) float64 { return fdlibmTanh(x) }

// Math_hypot implements Math.hypot.
func Math_hypot(x, y float64) float64 { return fdlibmHypot(x, y) }

// Math_expm1 implements Math.expm1.
func Math_expm1(x float64) float64 { return fdlibmExpm1(x) }

// Math_log1p implements Math.log1p.
func Math_log1p(x float64) float64 { return fdlibmLog1p(x) }

// Math_copySign_double_double implements Math.copySign(double, double), which
// treats the sign bit of a NaN sign argument as given.
func Math_copySign_double_double(magnitude, sign float64) float64 {
	const signBit = 1 << 63
	return math.Float64frombits(math.Float64bits(sign)&signBit | math.Float64bits(magnitude)&^signBit)
}

// Math_copySign_float_float implements Math.copySign(float, float), which
// treats the sign bit of a NaN sign argument as given.
func Math_copySign_float_float(magnitude, sign float32) float32 {
	const signBit = 1 << 31
	return math.Float32frombits(math.Float32bits(sign)&signBit | math.Float32bits(magnitude)&^signBit)
}

// Math_getExponent_float implements Math.getExponent(float).
func Math_getExponent_float(f float32) int32 {
	return int32((math.Float32bits(f)&0x7f800000)>>23) - 127
}

// Math_getExponent_double implements Math.getExponent(double).
func Math_getExponent_double(d float64) int32 {
	return int32((math.Float64bits(d)&0x7ff0000000000000)>>52) - 1023
}

// Math_nextAfter_double_double implements Math.nextAfter(double, double).
func Math_nextAfter_double_double(start, direction float64) float64 {
	switch {
	case start > direction: // descending
		if start != 0 {
			transducer := int64(math.Float64bits(start))
			if transducer > 0 {
				return math.Float64frombits(uint64(transducer - 1))
			}
			return math.Float64frombits(uint64(transducer + 1))
		}
		return -0x1p-1074
	case start < direction: // ascending
		// Add +0.0 to get rid of a -0.0 (+0.0 + -0.0 => +0.0).
		transducer := int64(math.Float64bits(start + 0.0))
		if transducer >= 0 {
			return math.Float64frombits(uint64(transducer + 1))
		}
		return math.Float64frombits(uint64(transducer - 1))
	case start == direction:
		return direction
	}
	return start + direction // start or direction is NaN
}

// Math_nextAfter_float_double implements Math.nextAfter(float, double).
func Math_nextAfter_float_double(start float32, direction float64) float32 {
	switch {
	case float64(start) > direction: // descending
		if start != 0 {
			transducer := int32(math.Float32bits(start))
			if transducer > 0 {
				return math.Float32frombits(uint32(transducer - 1))
			}
			return math.Float32frombits(uint32(transducer + 1))
		}
		return -0x1p-149
	case float64(start) < direction: // ascending
		// Add +0.0 to get rid of a -0.0 (+0.0 + -0.0 => +0.0).
		transducer := int32(math.Float32bits(start + 0.0))
		if transducer >= 0 {
			return math.Float32frombits(uint32(transducer + 1))
		}
		return math.Float32frombits(uint32(transducer - 1))
	case float64(start) == direction:
		return float32(direction)
	}
	return start + float32(direction) // start or direction is NaN
}

// Math_nextUp_double implements Math.nextUp(double).
func Math_nextUp_double(d float64) float64 {
	if d < math.Inf(1) {
		// Add +0.0 to get rid of a -0.0 (+0.0 + -0.0 => +0.0).
		transducer := int64(math.Float64bits(d + 0.0))
		if transducer >= 0 {
			return math.Float64frombits(uint64(transducer + 1))
		}
		return math.Float64frombits(uint64(transducer - 1))
	}
	return d // NaN or +Infinity
}

// Math_nextUp_float implements Math.nextUp(float).
func Math_nextUp_float(f float32) float32 {
	if f < float32(math.Inf(1)) {
		// Add +0.0 to get rid of a -0.0 (+0.0 + -0.0 => +0.0).
		transducer := int32(math.Float32bits(f + 0.0))
		if transducer >= 0 {
			return math.Float32frombits(uint32(transducer + 1))
		}
		return math.Float32frombits(uint32(transducer - 1))
	}
	return f // NaN or +Infinity
}

// Math_nextDown_double implements Math.nextDown(double).
func Math_nextDown_double(d float64) float64 {
	if d != d || math.IsInf(d, -1) {
		return d
	}
	if d == 0 {
		return -0x1p-1074
	}
	if d > 0 {
		return math.Float64frombits(math.Float64bits(d) - 1)
	}
	return math.Float64frombits(math.Float64bits(d) + 1)
}

// Math_nextDown_float implements Math.nextDown(float).
func Math_nextDown_float(f float32) float32 {
	if f != f || math.IsInf(float64(f), -1) {
		return f
	}
	if f == 0 {
		return -0x1p-149
	}
	if f > 0 {
		return math.Float32frombits(math.Float32bits(f) - 1)
	}
	return math.Float32frombits(math.Float32bits(f) + 1)
}

// Math_scalb_double_int implements Math.scalb(double, int).
func Math_scalb_double_int(d float64, scaleFactor int32) float64 {
	// Magnitude of a power of two so large that scaling a finite nonzero
	// value by it would be guaranteed to over or underflow; due to rounding,
	// scaling down takes an additional power of two which is reflected here.
	const maxScale = 1023 + 1022 + 53 + 1
	var scaleIncrement int32
	var expDelta float64

	// Make sure scaling factor is in a reasonable range.
	if scaleFactor < 0 {
		scaleFactor = Math_max_int_int(scaleFactor, -maxScale)
		scaleIncrement = -512
		expDelta = 0x1p-512
	} else {
		scaleFactor = Math_min_int_int(scaleFactor, maxScale)
		scaleIncrement = 512
		expDelta = 0x1p512
	}

	// Calculate (scaleFactor % +/-512), 512 = 2^9, using technique from
	// "Hacker's Delight" section 10-2.
	t := int32(uint32(scaleFactor>>(9-1)) >> (32 - 9))
	expAdjust := ((scaleFactor + t) & (512 - 1)) - t

	d = float64(d * mathPowerOfTwoD(expAdjust))
	scaleFactor -= expAdjust

	for scaleFactor != 0 {
		d = float64(d * expDelta)
		scaleFactor -= scaleIncrement
	}
	return d
}

// Math_scalb_float_int implements Math.scalb(float, int).
func Math_scalb_float_int(f float32, scaleFactor int32) float32 {
	const maxScale = 127 + 126 + 24 + 1
	scaleFactor = Math_max_int_int(Math_min_int_int(scaleFactor, maxScale), -maxScale)
	return float32(float64(float64(f) * mathPowerOfTwoD(scaleFactor)))
}

// mathPowerOfTwoD returns 2^n for a normal exponent n (Math.powerOfTwoD).
func mathPowerOfTwoD(n int32) float64 {
	return math.Float64frombits(uint64(int64(n)+1023) << 52 & 0x7ff0000000000000)
}

// mathPowerOfTwoF returns 2^n for a normal exponent n (Math.powerOfTwoF).
func mathPowerOfTwoF(n int32) float32 {
	return math.Float32frombits(uint32(n+127) << 23 & 0x7f800000)
}

// StrictMath_sin implements StrictMath.sin.
func StrictMath_sin(a float64) float64 { return fdlibmSin(a) }

// StrictMath_cos implements StrictMath.cos.
func StrictMath_cos(a float64) float64 { return fdlibmCos(a) }

// StrictMath_tan implements StrictMath.tan.
func StrictMath_tan(a float64) float64 { return fdlibmTan(a) }

// StrictMath_asin implements StrictMath.asin.
func StrictMath_asin(a float64) float64 { return fdlibmAsin(a) }

// StrictMath_acos implements StrictMath.acos.
func StrictMath_acos(a float64) float64 { return fdlibmAcos(a) }

// StrictMath_atan implements StrictMath.atan.
func StrictMath_atan(a float64) float64 { return fdlibmAtan(a) }

// StrictMath_toRadians implements StrictMath.toRadians.
func StrictMath_toRadians(angdeg float64) float64 { return Math_toRadians(angdeg) }

// StrictMath_toDegrees implements StrictMath.toDegrees.
func StrictMath_toDegrees(angrad float64) float64 { return Math_toDegrees(angrad) }

// StrictMath_exp implements StrictMath.exp.
func StrictMath_exp(a float64) float64 { return fdlibmExp(a) }

// StrictMath_log implements StrictMath.log.
func StrictMath_log(a float64) float64 { return fdlibmLog(a) }

// StrictMath_log10 implements StrictMath.log10.
func StrictMath_log10(a float64) float64 { return fdlibmLog10(a) }

// StrictMath_sqrt implements StrictMath.sqrt (correctly rounded).
func StrictMath_sqrt(a float64) float64 { return math.Sqrt(a) }

// StrictMath_cbrt implements StrictMath.cbrt.
func StrictMath_cbrt(a float64) float64 { return fdlibmCbrt(a) }

// StrictMath_IEEEremainder implements StrictMath.IEEEremainder.
func StrictMath_IEEEremainder(f1, f2 float64) float64 { return fdlibmIEEEremainder(f1, f2) }

// StrictMath_ceil implements StrictMath.ceil.
func StrictMath_ceil(a float64) float64 { return math.Ceil(a) }

// StrictMath_floor implements StrictMath.floor.
func StrictMath_floor(a float64) float64 { return math.Floor(a) }

// StrictMath_rint implements StrictMath.rint.
func StrictMath_rint(a float64) float64 { return math.RoundToEven(a) }

// StrictMath_atan2 implements StrictMath.atan2.
func StrictMath_atan2(y, x float64) float64 { return fdlibmAtan2(y, x) }

// StrictMath_pow implements StrictMath.pow.
func StrictMath_pow(a, b float64) float64 { return fdlibmPow(a, b) }

// StrictMath_round_float implements StrictMath.round(float).
func StrictMath_round_float(a float32) int32 { return Math_round_float(a) }

// StrictMath_round_double implements StrictMath.round(double).
func StrictMath_round_double(a float64) int64 { return Math_round_double(a) }

// StrictMath_random implements StrictMath.random. As in Java, the sequence is
// not reproducible.
func StrictMath_random() float64 { return Math_random() }

// StrictMath_addExact_int_int implements StrictMath.addExact(int, int).
func StrictMath_addExact_int_int(x, y int32) int32 { return Math_addExact_int_int(x, y) }

// StrictMath_addExact_long_long implements StrictMath.addExact(long, long).
func StrictMath_addExact_long_long(x, y int64) int64 { return Math_addExact_long_long(x, y) }

// StrictMath_subtractExact_int_int implements StrictMath.subtractExact(int,
// int).
func StrictMath_subtractExact_int_int(x, y int32) int32 { return Math_subtractExact_int_int(x, y) }

// StrictMath_subtractExact_long_long implements
// StrictMath.subtractExact(long, long).
func StrictMath_subtractExact_long_long(x, y int64) int64 {
	return Math_subtractExact_long_long(x, y)
}

// StrictMath_multiplyExact_int_int implements StrictMath.multiplyExact(int,
// int).
func StrictMath_multiplyExact_int_int(x, y int32) int32 { return Math_multiplyExact_int_int(x, y) }

// StrictMath_multiplyExact_long_long implements
// StrictMath.multiplyExact(long, long).
func StrictMath_multiplyExact_long_long(x, y int64) int64 {
	return Math_multiplyExact_long_long(x, y)
}

// StrictMath_toIntExact implements StrictMath.toIntExact.
func StrictMath_toIntExact(value int64) int32 { return Math_toIntExact(value) }

// StrictMath_floorDiv_int_int implements StrictMath.floorDiv(int, int).
func StrictMath_floorDiv_int_int(x, y int32) int32 { return Math_floorDiv_int_int(x, y) }

// StrictMath_floorDiv_long_long implements StrictMath.floorDiv(long, long).
func StrictMath_floorDiv_long_long(x, y int64) int64 { return Math_floorDiv_long_long(x, y) }

// StrictMath_floorMod_int_int implements StrictMath.floorMod(int, int).
func StrictMath_floorMod_int_int(x, y int32) int32 { return Math_floorMod_int_int(x, y) }

// StrictMath_floorMod_long_long implements StrictMath.floorMod(long, long).
func StrictMath_floorMod_long_long(x, y int64) int64 { return Math_floorMod_long_long(x, y) }

// StrictMath_abs_int implements StrictMath.abs(int).
func StrictMath_abs_int(a int32) int32 { return Math_abs_int(a) }

// StrictMath_abs_long implements StrictMath.abs(long).
func StrictMath_abs_long(a int64) int64 { return Math_abs_long(a) }

// StrictMath_abs_float implements StrictMath.abs(float).
func StrictMath_abs_float(a float32) float32 { return Math_abs_float(a) }

// StrictMath_abs_double implements StrictMath.abs(double).
func StrictMath_abs_double(a float64) float64 { return Math_abs_double(a) }

// StrictMath_max_int_int implements StrictMath.max(int, int).
func StrictMath_max_int_int(a, b int32) int32 { return Math_max_int_int(a, b) }

// StrictMath_max_long_long implements StrictMath.max(long, long).
func StrictMath_max_long_long(a, b int64) int64 { return Math_max_long_long(a, b) }

// StrictMath_max_float_float implements StrictMath.max(float, float).
func StrictMath_max_float_float(a, b float32) float32 { return Math_max_float_float(a, b) }

// StrictMath_max_double_double implements StrictMath.max(double, double).
func StrictMath_max_double_double(a, b float64) float64 { return Math_max_double_double(a, b) }

// StrictMath_min_int_int implements StrictMath.min(int, int).
func StrictMath_min_int_int(a, b int32) int32 { return Math_min_int_int(a, b) }

// StrictMath_min_long_long implements StrictMath.min(long, long).
func StrictMath_min_long_long(a, b int64) int64 { return Math_min_long_long(a, b) }

// StrictMath_min_float_float implements StrictMath.min(float, float).
func StrictMath_min_float_float(a, b float32) float32 { return Math_min_float_float(a, b) }

// StrictMath_min_double_double implements StrictMath.min(double, double).
func StrictMath_min_double_double(a, b float64) float64 { return Math_min_double_double(a, b) }

// StrictMath_ulp_double implements StrictMath.ulp(double).
func StrictMath_ulp_double(d float64) float64 { return Math_ulp_double(d) }

// StrictMath_ulp_float implements StrictMath.ulp(float).
func StrictMath_ulp_float(f float32) float32 { return Math_ulp_float(f) }

// StrictMath_signum_double implements StrictMath.signum(double).
func StrictMath_signum_double(d float64) float64 { return Math_signum_double(d) }

// StrictMath_signum_float implements StrictMath.signum(float).
func StrictMath_signum_float(f float32) float32 { return Math_signum_float(f) }

// StrictMath_sinh implements StrictMath.sinh.
func StrictMath_sinh(x float64) float64 { return fdlibmSinh(x) }

// StrictMath_cosh implements StrictMath.cosh.
func StrictMath_cosh(x float64) float64 { return fdlibmCosh(x) }

// StrictMath_tanh implements StrictMath.tanh.
func StrictMath_tanh(x float64) float64 { return fdlibmTanh(x) }

// StrictMath_hypot implements StrictMath.hypot.
func StrictMath_hypot(x, y float64) float64 { return fdlibmHypot(x, y) }

// StrictMath_expm1 implements StrictMath.expm1.
func StrictMath_expm1(x float64) float64 { return fdlibmExpm1(x) }

// StrictMath_log1p implements StrictMath.log1p.
func StrictMath_log1p(x float64) float64 { return fdlibmLog1p(x) }

// StrictMath_copySign_double_double implements StrictMath.copySign(double,
// double). Unlike Math.copySign, Java specifies that a NaN sign argument is
// treated as positive.
func StrictMath_copySign_double_double(magnitude, sign float64) float64 {
	if sign != sign {
		sign = 1.0
	}
	return Math_copySign_double_double(magnitude, sign)
}

// StrictMath_copySign_float_float implements StrictMath.copySign(float,
// float). Unlike Math.copySign, Java specifies that a NaN sign argument is
// treated as positive.
func StrictMath_copySign_float_float(magnitude, sign float32) float32 {
	if sign != sign {
		sign = 1.0
	}
	return Math_copySign_float_float(magnitude, sign)
}

// StrictMath_getExponent_float implements StrictMath.getExponent(float).
func StrictMath_getExponent_float(f float32) int32 { return Math_getExponent_float(f) }

// StrictMath_getExponent_double implements StrictMath.getExponent(double).
func StrictMath_getExponent_double(d float64) int32 { return Math_getExponent_double(d) }

// StrictMath_nextAfter_double_double implements StrictMath.nextAfter(double,
// double).
func StrictMath_nextAfter_double_double(start, direction float64) float64 {
	return Math_nextAfter_double_double(start, direction)
}

// StrictMath_nextAfter_float_double implements StrictMath.nextAfter(float,
// double).
func StrictMath_nextAfter_float_double(start float32, direction float64) float32 {
	return Math_nextAfter_float_double(start, direction)
}

// StrictMath_nextUp_double implements StrictMath.nextUp(double).
func StrictMath_nextUp_double(d float64) float64 { return Math_nextUp_double(d) }

// StrictMath_nextUp_float implements StrictMath.nextUp(float).
func StrictMath_nextUp_float(f float32) float32 { return Math_nextUp_float(f) }

// StrictMath_nextDown_double implements StrictMath.nextDown(double).
func StrictMath_nextDown_double(d float64) float64 { return Math_nextDown_double(d) }

// StrictMath_nextDown_float implements StrictMath.nextDown(float).
func StrictMath_nextDown_float(f float32) float32 { return Math_nextDown_float(f) }

// StrictMath_scalb_double_int implements StrictMath.scalb(double, int).
func StrictMath_scalb_double_int(d float64, scaleFactor int32) float64 {
	return Math_scalb_double_int(d, scaleFactor)
}

// StrictMath_scalb_float_int implements StrictMath.scalb(float, int).
func StrictMath_scalb_float_int(f float32, scaleFactor int32) float32 {
	return Math_scalb_float_int(f, scaleFactor)
}
