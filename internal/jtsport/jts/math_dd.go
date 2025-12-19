package jts

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Math_DD implements extended-precision floating-point numbers
// which maintain 106 bits (approximately 30 decimal digits) of precision.
//
// A Math_DD uses a representation containing two double-precision values.
// A number x is represented as a pair of doubles, x.hi and x.lo,
// such that the number represented by x is x.hi + x.lo, where
//
//	|x.lo| <= 0.5*ulp(x.hi)
//
// and ulp(y) means "unit in the last place of y".
// The basic arithmetic operations are implemented using
// convenient properties of IEEE-754 floating-point arithmetic.
//
// The range of values which can be represented is the same as in IEEE-754.
// The precision of the representable numbers
// is twice as great as IEEE-754 double precision.
//
// The correctness of the arithmetic algorithms relies on operations
// being performed with standard IEEE-754 double precision and rounding.
//
// The API provides both a set of value-oriented operations
// and a set of mutating operations.
// Value-oriented operations treat Math_DD values as
// immutable; operations on them return new objects carrying the result
// of the operation. This provides a simple and safe semantics for
// writing Math_DD expressions. However, there is a performance
// penalty for the object allocations required.
// The mutable interface updates object values in-place.
// It provides optimum memory performance, but requires
// care to ensure that aliasing errors are not created
// and constant values are not changed.
//
// This implementation uses algorithms originally designed variously by
// Knuth, Kahan, Dekker, and Linnainmaa.
// Douglas Priest developed the first C implementation of these techniques.
// Other more recent C++ implementations are due to Keith M. Briggs and David Bailey et al.
type Math_DD struct {
	hi float64
	lo float64
}

// Math_DD_Pi is the value nearest to the constant Pi.
var Math_DD_Pi = &Math_DD{hi: 3.141592653589793116e+00, lo: 1.224646799147353207e-16}

// Math_DD_TwoPi is the value nearest to the constant 2 * Pi.
var Math_DD_TwoPi = &Math_DD{hi: 6.283185307179586232e+00, lo: 2.449293598294706414e-16}

// Math_DD_PiOver2 is the value nearest to the constant Pi / 2.
var Math_DD_PiOver2 = &Math_DD{hi: 1.570796326794896558e+00, lo: 6.123233995736766036e-17}

// Math_DD_E is the value nearest to the constant e (the natural logarithm base).
var Math_DD_E = &Math_DD{hi: 2.718281828459045091e+00, lo: 1.445646891729250158e-16}

// Math_DD_NaN is a value representing the result of an operation which does not return a valid number.
var Math_DD_NaN = &Math_DD{hi: math.NaN(), lo: math.NaN()}

// Math_DD_Eps is the smallest representable relative difference between two Math_DD values.
const Math_DD_Eps = 1.23259516440783e-32 // 2^-106

func math_createNaN() *Math_DD {
	return Math_NewDDFromHiLo(math.NaN(), math.NaN())
}

// Math_DD_ValueOfString converts the string argument to a Math_DD number.
func Math_DD_ValueOfString(str string) (*Math_DD, error) {
	return Math_DD_Parse(str)
}

// Math_DD_ValueOfFloat64 converts the float64 argument to a Math_DD number.
func Math_DD_ValueOfFloat64(x float64) *Math_DD {
	return Math_NewDDFromFloat64(x)
}

const math_dd_split = 134217729.0 // 2^27+1, for IEEE double

// Math_NewDD creates a new Math_DD with value 0.0.
func Math_NewDD() *Math_DD {
	d := &Math_DD{}
	d.initFloat64(0.0)
	return d
}

// Math_NewDDFromFloat64 creates a new Math_DD with the given float64 value.
func Math_NewDDFromFloat64(x float64) *Math_DD {
	d := &Math_DD{}
	d.initFloat64(x)
	return d
}

// Math_NewDDFromHiLo creates a new Math_DD with value (hi, lo).
func Math_NewDDFromHiLo(hi, lo float64) *Math_DD {
	d := &Math_DD{}
	d.initHiLo(hi, lo)
	return d
}

// Math_NewDDFromDD creates a new Math_DD with value equal to the argument.
func Math_NewDDFromDD(dd *Math_DD) *Math_DD {
	d := &Math_DD{}
	d.initDD(dd)
	return d
}

// Math_NewDDFromString creates a new Math_DD with value equal to the parsed string.
func Math_NewDDFromString(str string) (*Math_DD, error) {
	return Math_DD_Parse(str)
}

// Math_DD_Copy creates a new Math_DD with the value of the argument.
func Math_DD_Copy(dd *Math_DD) *Math_DD {
	return Math_NewDDFromDD(dd)
}

// Clone creates and returns a copy of this value.
func (d *Math_DD) Clone() *Math_DD {
	return Math_NewDDFromDD(d)
}

func (d *Math_DD) initFloat64(x float64) {
	d.hi = x
	d.lo = 0.0
}

func (d *Math_DD) initHiLo(hi, lo float64) {
	d.hi = hi
	d.lo = lo
}

func (d *Math_DD) initDD(dd *Math_DD) {
	d.hi = dd.hi
	d.lo = dd.lo
}

// SetValue sets the value for the Math_DD object from another Math_DD.
// This method supports the mutating operations concept.
func (d *Math_DD) SetValue(value *Math_DD) *Math_DD {
	d.initDD(value)
	return d
}

// SetValueFloat64 sets the value for the Math_DD object from a float64.
// This method supports the mutating operations concept.
func (d *Math_DD) SetValueFloat64(value float64) *Math_DD {
	d.initFloat64(value)
	return d
}

// Add returns a new Math_DD whose value is (this + y).
func (d *Math_DD) Add(y *Math_DD) *Math_DD {
	return Math_DD_Copy(d).SelfAdd(y)
}

// AddFloat64 returns a new Math_DD whose value is (this + y).
func (d *Math_DD) AddFloat64(y float64) *Math_DD {
	return Math_DD_Copy(d).SelfAddFloat64(y)
}

// SelfAdd adds the argument to the value of this Math_DD.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
func (d *Math_DD) SelfAdd(y *Math_DD) *Math_DD {
	return d.selfAddHiLo(y.hi, y.lo)
}

// SelfAddFloat64 adds the float64 argument to the value of this Math_DD.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
//
// Note: The float64() conversions below are essential to prevent FMA
// (fused multiply-add) optimizations that would break the double-double
// algorithm by not rounding intermediate results.
func (d *Math_DD) SelfAddFloat64(y float64) *Math_DD {
	var H, h, S, s, e, f float64
	S = float64(d.hi + y)
	e = float64(S - d.hi)
	s = float64(S - e)
	s = float64(float64(y-e) + float64(d.hi-s))
	f = float64(s + d.lo)
	H = float64(S + f)
	h = float64(f + float64(S-H))
	d.hi = float64(H + h)
	d.lo = float64(h + float64(H-d.hi))
	return d
}

func (d *Math_DD) selfAddHiLo(yhi, ylo float64) *Math_DD {
	var H, h, T, t, S, s, e, f float64
	S = float64(d.hi + yhi)
	T = float64(d.lo + ylo)
	e = float64(S - d.hi)
	f = float64(T - d.lo)
	s = float64(S - e)
	t = float64(T - f)
	s = float64(float64(yhi-e) + float64(d.hi-s))
	t = float64(float64(ylo-f) + float64(d.lo-t))
	e = float64(s + T)
	H = float64(S + e)
	h = float64(e + float64(S-H))
	e = float64(t + h)

	zhi := float64(H + e)
	zlo := float64(e + float64(H-zhi))
	d.hi = zhi
	d.lo = zlo
	return d
}

// Subtract computes a new Math_DD whose value is (this - y).
func (d *Math_DD) Subtract(y *Math_DD) *Math_DD {
	return d.Add(y.Negate())
}

// SubtractFloat64 computes a new Math_DD whose value is (this - y).
func (d *Math_DD) SubtractFloat64(y float64) *Math_DD {
	return d.AddFloat64(-y)
}

// SelfSubtract subtracts the argument from the value of this Math_DD.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
func (d *Math_DD) SelfSubtract(y *Math_DD) *Math_DD {
	if d.IsNaN() {
		return d
	}
	return d.selfAddHiLo(-y.hi, -y.lo)
}

// SelfSubtractFloat64 subtracts the float64 argument from the value of this Math_DD.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
func (d *Math_DD) SelfSubtractFloat64(y float64) *Math_DD {
	if d.IsNaN() {
		return d
	}
	return d.selfAddHiLo(-y, 0.0)
}

// Negate returns a new Math_DD whose value is -this.
func (d *Math_DD) Negate() *Math_DD {
	if d.IsNaN() {
		return d
	}
	return Math_NewDDFromHiLo(-d.hi, -d.lo)
}

// Multiply returns a new Math_DD whose value is (this * y).
func (d *Math_DD) Multiply(y *Math_DD) *Math_DD {
	if y.IsNaN() {
		return math_createNaN()
	}
	return Math_DD_Copy(d).SelfMultiply(y)
}

// MultiplyFloat64 returns a new Math_DD whose value is (this * y).
func (d *Math_DD) MultiplyFloat64(y float64) *Math_DD {
	if math.IsNaN(y) {
		return math_createNaN()
	}
	return Math_DD_Copy(d).selfMultiplyHiLo(y, 0.0)
}

// SelfMultiply multiplies this object by the argument, returning this.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
func (d *Math_DD) SelfMultiply(y *Math_DD) *Math_DD {
	return d.selfMultiplyHiLo(y.hi, y.lo)
}

// SelfMultiplyFloat64 multiplies this object by the float64 argument, returning this.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
func (d *Math_DD) SelfMultiplyFloat64(y float64) *Math_DD {
	return d.selfMultiplyHiLo(y, 0.0)
}

func (d *Math_DD) selfMultiplyHiLo(yhi, ylo float64) *Math_DD {
	var hx, tx, hy, ty, C, c float64
	C = float64(math_dd_split * d.hi)
	hx = float64(C - d.hi)
	c = float64(math_dd_split * yhi)
	hx = float64(C - hx)
	tx = float64(d.hi - hx)
	hy = float64(c - yhi)
	C = float64(d.hi * yhi)
	hy = float64(c - hy)
	ty = float64(yhi - hy)
	c = float64(float64(float64(float64(float64(hx*hy)-C)+float64(hx*ty))+float64(tx*hy))+float64(tx*ty)) + float64(float64(d.hi*ylo)+float64(d.lo*yhi))
	zhi := float64(C + c)
	hx = float64(C - zhi)
	zlo := float64(c + hx)
	d.hi = zhi
	d.lo = zlo
	return d
}

// Divide computes a new Math_DD whose value is (this / y).
func (d *Math_DD) Divide(y *Math_DD) *Math_DD {
	var hc, tc, hy, ty, C, c, U, u float64
	C = float64(d.hi / y.hi)
	c = float64(math_dd_split * C)
	hc = float64(c - C)
	u = float64(math_dd_split * y.hi)
	hc = float64(c - hc)
	tc = float64(C - hc)
	hy = float64(u - y.hi)
	U = float64(C * y.hi)
	hy = float64(u - hy)
	ty = float64(y.hi - hy)
	u = float64(float64(float64(float64(hc*hy)-U)+float64(hc*ty))+float64(tc*hy)) + float64(tc*ty)
	c = float64(float64(float64(float64(d.hi-U)-u)+d.lo)-float64(C*y.lo)) / y.hi
	u = float64(C + c)

	zhi := u
	zlo := float64(float64(C-u) + c)
	return Math_NewDDFromHiLo(zhi, zlo)
}

// DivideFloat64 computes a new Math_DD whose value is (this / y).
func (d *Math_DD) DivideFloat64(y float64) *Math_DD {
	if math.IsNaN(y) {
		return math_createNaN()
	}
	return Math_DD_Copy(d).selfDivideHiLo(y, 0.0)
}

// SelfDivide divides this object by the argument, returning this.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
func (d *Math_DD) SelfDivide(y *Math_DD) *Math_DD {
	return d.selfDivideHiLo(y.hi, y.lo)
}

// SelfDivideFloat64 divides this object by the float64 argument, returning this.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
func (d *Math_DD) SelfDivideFloat64(y float64) *Math_DD {
	return d.selfDivideHiLo(y, 0.0)
}

func (d *Math_DD) selfDivideHiLo(yhi, ylo float64) *Math_DD {
	var hc, tc, hy, ty, C, c, U, u float64
	C = float64(d.hi / yhi)
	c = float64(math_dd_split * C)
	hc = float64(c - C)
	u = float64(math_dd_split * yhi)
	hc = float64(c - hc)
	tc = float64(C - hc)
	hy = float64(u - yhi)
	U = float64(C * yhi)
	hy = float64(u - hy)
	ty = float64(yhi - hy)
	u = float64(float64(float64(float64(hc*hy)-U)+float64(hc*ty))+float64(tc*hy)) + float64(tc*ty)
	c = float64(float64(float64(float64(d.hi-U)-u)+d.lo)-float64(C*ylo)) / yhi
	u = float64(C + c)

	d.hi = u
	d.lo = float64(float64(C-u) + c)
	return d
}

// Reciprocal returns a Math_DD whose value is 1 / this.
func (d *Math_DD) Reciprocal() *Math_DD {
	var hc, tc, hy, ty, C, c, U, u float64
	C = float64(1.0 / d.hi)
	c = float64(math_dd_split * C)
	hc = float64(c - C)
	u = float64(math_dd_split * d.hi)
	hc = float64(c - hc)
	tc = float64(C - hc)
	hy = float64(u - d.hi)
	U = float64(C * d.hi)
	hy = float64(u - hy)
	ty = float64(d.hi - hy)
	u = float64(float64(float64(float64(hc*hy)-U)+float64(hc*ty))+float64(tc*hy)) + float64(tc*ty)
	c = float64(float64(float64(1.0-U)-u)-float64(C*d.lo)) / d.hi

	zhi := float64(C + c)
	zlo := float64(float64(C-zhi) + c)
	return Math_NewDDFromHiLo(zhi, zlo)
}

// Floor returns the largest value that is not greater than the argument
// and is equal to a mathematical integer.
// If this value is NaN, returns NaN.
func (d *Math_DD) Floor() *Math_DD {
	if d.IsNaN() {
		return Math_DD_NaN
	}
	fhi := math.Floor(d.hi)
	flo := 0.0
	// Hi is already integral. Floor the low word.
	if fhi == d.hi {
		flo = math.Floor(d.lo)
	}
	return Math_NewDDFromHiLo(fhi, flo)
}

// Ceil returns the smallest value that is not less than the argument
// and is equal to a mathematical integer.
// If this value is NaN, returns NaN.
func (d *Math_DD) Ceil() *Math_DD {
	if d.IsNaN() {
		return Math_DD_NaN
	}
	fhi := math.Ceil(d.hi)
	flo := 0.0
	// Hi is already integral. Ceil the low word.
	if fhi == d.hi {
		flo = math.Ceil(d.lo)
	}
	return Math_NewDDFromHiLo(fhi, flo)
}

// Signum returns an integer indicating the sign of this value.
// Returns 1 if > 0, -1 if < 0, 0 if = 0 or NaN.
func (d *Math_DD) Signum() int {
	if d.hi > 0 {
		return 1
	}
	if d.hi < 0 {
		return -1
	}
	if d.lo > 0 {
		return 1
	}
	if d.lo < 0 {
		return -1
	}
	return 0
}

// Rint rounds this value to the nearest integer.
// The value is rounded to an integer by adding 1/2 and taking the floor of the result.
// If this value is NaN, returns NaN.
func (d *Math_DD) Rint() *Math_DD {
	if d.IsNaN() {
		return d
	}
	plus5 := d.AddFloat64(0.5)
	return plus5.Floor()
}

// Trunc returns the integer which is largest in absolute value and not further
// from zero than this value.
// If this value is NaN, returns NaN.
func (d *Math_DD) Trunc() *Math_DD {
	if d.IsNaN() {
		return Math_DD_NaN
	}
	if d.IsPositive() {
		return d.Floor()
	}
	return d.Ceil()
}

// Abs returns the absolute value of this value.
// If this value is NaN, it is returned.
func (d *Math_DD) Abs() *Math_DD {
	if d.IsNaN() {
		return Math_DD_NaN
	}
	if d.IsNegative() {
		return d.Negate()
	}
	return Math_NewDDFromDD(d)
}

// Sqr computes the square of this value.
func (d *Math_DD) Sqr() *Math_DD {
	return d.Multiply(d)
}

// SelfSqr squares this object.
// To prevent altering constants, this method must only be used on values
// known to be newly created.
func (d *Math_DD) SelfSqr() *Math_DD {
	return d.SelfMultiply(d)
}

// Math_DD_SqrFloat64 computes the square of a float64 value.
func Math_DD_SqrFloat64(x float64) *Math_DD {
	return Math_DD_ValueOfFloat64(x).SelfMultiplyFloat64(x)
}

// Sqrt computes the positive square root of this value.
// If the number is NaN or negative, NaN is returned.
func (d *Math_DD) Sqrt() *Math_DD {
	// Strategy: Use Karp's trick: if x is an approximation
	// to sqrt(a), then
	//
	//    sqrt(a) = a*x + [a - (a*x)^2] * x / 2   (approx)
	//
	// The approximation is accurate to twice the accuracy of x.
	// Also, the multiplication (a*x) and [-]*x can be done with
	// only half the precision.

	if d.IsZero() {
		return Math_DD_ValueOfFloat64(0.0)
	}

	if d.IsNegative() {
		return Math_DD_NaN
	}

	x := 1.0 / math.Sqrt(d.hi)
	ax := d.hi * x

	axdd := Math_DD_ValueOfFloat64(ax)
	diffSq := d.Subtract(axdd.Sqr())
	d2 := diffSq.hi * (x * 0.5)

	return axdd.AddFloat64(d2)
}

// Math_DD_SqrtFloat64 computes the positive square root of a float64 value.
func Math_DD_SqrtFloat64(x float64) *Math_DD {
	return Math_DD_ValueOfFloat64(x).Sqrt()
}

// Pow computes the value of this number raised to an integral power.
// Follows semantics of Java Math.pow as closely as possible.
func (d *Math_DD) Pow(exp int) *Math_DD {
	if exp == 0 {
		return Math_DD_ValueOfFloat64(1.0)
	}

	r := Math_NewDDFromDD(d)
	s := Math_DD_ValueOfFloat64(1.0)
	n := java.AbsInt(exp)

	if n > 1 {
		// Use binary exponentiation.
		for n > 0 {
			if n%2 == 1 {
				s.SelfMultiply(r)
			}
			n /= 2
			if n > 0 {
				r = r.Sqr()
			}
		}
	} else {
		s = r
	}

	// Compute the reciprocal if exp is negative.
	if exp < 0 {
		return s.Reciprocal()
	}
	return s
}

// Math_DD_DeterminantFloat64 computes the determinant of the 2x2 matrix with the given entries.
func Math_DD_DeterminantFloat64(x1, y1, x2, y2 float64) *Math_DD {
	return Math_DD_DeterminantDD(
		Math_DD_ValueOfFloat64(x1), Math_DD_ValueOfFloat64(y1),
		Math_DD_ValueOfFloat64(x2), Math_DD_ValueOfFloat64(y2),
	)
}

// Math_DD_DeterminantDD computes the determinant of the 2x2 matrix with the given Math_DD entries.
func Math_DD_DeterminantDD(x1, y1, x2, y2 *Math_DD) *Math_DD {
	return x1.Multiply(y2).SelfSubtract(y1.Multiply(x2))
}

// Min computes the minimum of this and another Math_DD number.
func (d *Math_DD) Min(x *Math_DD) *Math_DD {
	if d.Le(x) {
		return d
	}
	return x
}

// Max computes the maximum of this and another Math_DD number.
func (d *Math_DD) Max(x *Math_DD) *Math_DD {
	if d.Ge(x) {
		return d
	}
	return x
}

// DoubleValue converts this value to the nearest double-precision number.
func (d *Math_DD) DoubleValue() float64 {
	return d.hi + d.lo
}

// IntValue converts this value to the nearest integer.
func (d *Math_DD) IntValue() int {
	return int(d.hi)
}

// IsZero tests whether this value is equal to 0.
func (d *Math_DD) IsZero() bool {
	return d.hi == 0.0 && d.lo == 0.0
}

// IsNegative tests whether this value is less than 0.
func (d *Math_DD) IsNegative() bool {
	return d.hi < 0.0 || (d.hi == 0.0 && d.lo < 0.0)
}

// IsPositive tests whether this value is greater than 0.
func (d *Math_DD) IsPositive() bool {
	return d.hi > 0.0 || (d.hi == 0.0 && d.lo > 0.0)
}

// IsNaN tests whether this value is NaN.
func (d *Math_DD) IsNaN() bool {
	return math.IsNaN(d.hi)
}

// Equals tests whether this value is equal to another Math_DD value.
func (d *Math_DD) Equals(y *Math_DD) bool {
	return d.hi == y.hi && d.lo == y.lo
}

// Gt tests whether this value is greater than another Math_DD value.
func (d *Math_DD) Gt(y *Math_DD) bool {
	return (d.hi > y.hi) || (d.hi == y.hi && d.lo > y.lo)
}

// Ge tests whether this value is greater than or equal to another Math_DD value.
func (d *Math_DD) Ge(y *Math_DD) bool {
	return (d.hi > y.hi) || (d.hi == y.hi && d.lo >= y.lo)
}

// Lt tests whether this value is less than another Math_DD value.
func (d *Math_DD) Lt(y *Math_DD) bool {
	return (d.hi < y.hi) || (d.hi == y.hi && d.lo < y.lo)
}

// Le tests whether this value is less than or equal to another Math_DD value.
func (d *Math_DD) Le(y *Math_DD) bool {
	return (d.hi < y.hi) || (d.hi == y.hi && d.lo <= y.lo)
}

// CompareTo compares two Math_DD objects numerically.
// Returns -1, 0, or 1 depending on whether this value is less than, equal to,
// or greater than the value of other.
func (d *Math_DD) CompareTo(other *Math_DD) int {
	if d.hi < other.hi {
		return -1
	}
	if d.hi > other.hi {
		return 1
	}
	if d.lo < other.lo {
		return -1
	}
	if d.lo > other.lo {
		return 1
	}
	return 0
}

/*------------------------------------------------------------
 *   Output
 *------------------------------------------------------------
 */

const math_dd_maxPrintDigits = 32

var math_dd_ten = func() *Math_DD {
	return Math_NewDDFromFloat64(10.0)
}()

var math_dd_one = func() *Math_DD {
	return Math_NewDDFromFloat64(1.0)
}()

const math_dd_sciNotExponentChar = "E"
const math_dd_sciNotZero = "0.0E0"

// Dump dumps the components of this number to a string.
func (d *Math_DD) Dump() string {
	return fmt.Sprintf("Math_DD<%v, %v>", d.hi, d.lo)
}

// String returns a string representation of this number, in either standard or scientific notation.
// If the magnitude of the number is in the range [10^-3, 10^8]
// standard notation will be used. Otherwise, scientific notation will be used.
func (d *Math_DD) String() string {
	mag := math_magnitude(d.hi)
	if mag >= -3 && mag <= 20 {
		return d.ToStandardNotation()
	}
	return d.ToSciNotation()
}

// ToStandardNotation returns the string representation of this value in standard notation.
func (d *Math_DD) ToStandardNotation() string {
	specialStr := d.getSpecialNumberString()
	if specialStr != "" {
		return specialStr
	}

	var mag int
	sigDigits := d.extractSignificantDigits(true, &mag)
	decimalPointPos := mag + 1

	num := sigDigits
	// Add a leading 0 if the decimal point is the first char.
	if len(sigDigits) > 0 && sigDigits[0] == '.' {
		num = "0" + sigDigits
	} else if decimalPointPos < 0 {
		num = "0." + math_stringOfChar('0', -decimalPointPos) + sigDigits
	} else if !strings.Contains(sigDigits, ".") {
		// No point inserted - sig digits must be smaller than magnitude of number.
		// Add zeroes to end to make number the correct size.
		numZeroes := decimalPointPos - len(sigDigits)
		zeroes := math_stringOfChar('0', numZeroes)
		num = sigDigits + zeroes + ".0"
	}

	if d.IsNegative() {
		return "-" + num
	}
	return num
}

// ToSciNotation returns the string representation of this value in scientific notation.
func (d *Math_DD) ToSciNotation() string {
	// Special case zero.
	if d.IsZero() {
		return math_dd_sciNotZero
	}

	specialStr := d.getSpecialNumberString()
	if specialStr != "" {
		return specialStr
	}

	var mag int
	digits := d.extractSignificantDigits(false, &mag)
	expStr := math_dd_sciNotExponentChar + strconv.Itoa(mag)

	// Should never have leading zeroes.
	if len(digits) > 0 && digits[0] == '0' {
		panic(fmt.Sprintf("Found leading zero: %s", digits))
	}

	// Add decimal point.
	trailingDigits := ""
	if len(digits) > 1 {
		trailingDigits = digits[1:]
	}
	digitsWithDecimal := string(digits[0]) + "." + trailingDigits

	if d.IsNegative() {
		return "-" + digitsWithDecimal + expStr
	}
	return digitsWithDecimal + expStr
}

func (d *Math_DD) extractSignificantDigits(insertDecimalPoint bool, magnitude *int) string {
	y := d.Abs()
	// Compute *correct* magnitude of y.
	mag := math_magnitude(y.hi)
	scale := math_dd_ten.Pow(mag)
	y = y.Divide(scale)

	// Fix magnitude if off by one.
	if y.Gt(math_dd_ten) {
		y = y.Divide(math_dd_ten)
		mag++
	} else if y.Lt(math_dd_one) {
		y = y.Multiply(math_dd_ten)
		mag--
	}

	decimalPointPos := mag + 1
	var buf strings.Builder
	numDigits := math_dd_maxPrintDigits - 1
	for i := 0; i <= numDigits; i++ {
		if insertDecimalPoint && i == decimalPointPos {
			buf.WriteByte('.')
		}
		digit := int(y.hi)

		// If a negative remainder is encountered, simply terminate the extraction.
		// This is robust, but maybe slightly inaccurate.
		if digit < 0 {
			break
		}
		rebiasBy10 := false
		var digitChar byte
		if digit > 9 {
			// Set flag to re-bias after next 10-shift.
			// Output digit will end up being '9'.
			rebiasBy10 = true
			digitChar = '9'
		} else {
			digitChar = byte('0' + digit)
		}
		buf.WriteByte(digitChar)
		y = y.Subtract(Math_DD_ValueOfFloat64(float64(digit))).Multiply(math_dd_ten)
		if rebiasBy10 {
			y.SelfAdd(math_dd_ten)
		}

		continueExtractingDigits := true
		// Check if remaining digits will be 0, and if so don't output them.
		// Do this by comparing the magnitude of the remainder with the expected precision.
		remMag := math_magnitude(y.hi)
		if remMag < 0 && java.AbsInt(remMag) >= (numDigits-i) {
			continueExtractingDigits = false
		}
		if !continueExtractingDigits {
			break
		}
	}
	*magnitude = mag
	return buf.String()
}

func math_stringOfChar(ch byte, length int) string {
	var buf strings.Builder
	for i := 0; i < length; i++ {
		buf.WriteByte(ch)
	}
	return buf.String()
}

func (d *Math_DD) getSpecialNumberString() string {
	if d.IsZero() {
		return "0.0"
	}
	if d.IsNaN() {
		return "NaN "
	}
	return ""
}

func math_magnitude(x float64) int {
	xAbs := math.Abs(x)
	xLog10 := math.Log(xAbs) / math.Log(10)
	xMag := int(math.Floor(xLog10))
	// Since log computation is inexact, there may be an off-by-one error
	// in the computed magnitude.
	// Following tests that magnitude is correct, and adjusts it if not.
	xApprox := math.Pow(10, float64(xMag))
	if xApprox*10 <= xAbs {
		xMag++
	}
	return xMag
}

/*------------------------------------------------------------
 *   Input
 *------------------------------------------------------------
 */

// Math_DD_Parse converts a string representation of a real number into a Math_DD value.
// The format accepted is similar to the standard Go real number syntax.
func Math_DD_Parse(str string) (*Math_DD, error) {
	i := 0
	strlen := len(str)

	// Skip leading whitespace.
	for i < strlen && unicode.IsSpace(rune(str[i])) {
		i++
	}

	// Check for sign.
	isNegative := false
	if i < strlen {
		signCh := str[i]
		if signCh == '-' || signCh == '+' {
			i++
			if signCh == '-' {
				isNegative = true
			}
		}
	}

	// Scan all digits and accumulate into an integral value.
	// Keep track of the location of the decimal point (if any) to allow scaling later.
	val := Math_NewDD()

	numDigits := 0
	numBeforeDec := 0
	exp := 0
	hasDecimalChar := false

	for i < strlen {
		ch := str[i]
		i++
		if ch >= '0' && ch <= '9' {
			digit := float64(ch - '0')
			val.SelfMultiply(math_dd_ten)
			val.SelfAddFloat64(digit)
			numDigits++
			continue
		}
		if ch == '.' {
			numBeforeDec = numDigits
			hasDecimalChar = true
			continue
		}
		if ch == 'e' || ch == 'E' {
			expStr := str[i:]
			var err error
			exp, err = strconv.Atoi(expStr)
			if err != nil {
				return nil, fmt.Errorf("invalid exponent %s in string %s", expStr, str)
			}
			break
		}
		return nil, fmt.Errorf("unexpected character '%c' at position %d in string %s", ch, i, str)
	}
	val2 := val

	// Correct number of digits before decimal sign if we don't have a decimal sign in the string.
	if !hasDecimalChar {
		numBeforeDec = numDigits
	}

	// Scale the number correctly.
	numDecPlaces := numDigits - numBeforeDec - exp
	if numDecPlaces == 0 {
		val2 = val
	} else if numDecPlaces > 0 {
		scale := math_dd_ten.Pow(numDecPlaces)
		val2 = val.Divide(scale)
	} else {
		scale := math_dd_ten.Pow(-numDecPlaces)
		val2 = val.Multiply(scale)
	}

	// Apply leading sign, if any.
	if isNegative {
		return val2.Negate(), nil
	}
	return val2, nil
}
