package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestNaN(t *testing.T) {
	junit.AssertTrue(t, Math_DD_ValueOfFloat64(1).Divide(Math_DD_ValueOfFloat64(0)).IsNaN())
	junit.AssertTrue(t, Math_DD_ValueOfFloat64(1).Multiply(Math_DD_NaN).IsNaN())
}

func TestAddMult2(t *testing.T) {
	checkAddMult2(t, Math_NewDDFromFloat64(3))
	checkAddMult2(t, Math_DD_Pi)
}

func TestMultiplyDivide(t *testing.T) {
	checkMultiplyDivide(t, Math_DD_Pi, Math_DD_E, 1e-30)
	checkMultiplyDivide(t, Math_DD_TwoPi, Math_DD_E, 1e-30)
	checkMultiplyDivide(t, Math_DD_PiOver2, Math_DD_E, 1e-30)
	checkMultiplyDivide(t, Math_NewDDFromFloat64(39.4), Math_NewDDFromFloat64(10), 1e-30)
}

func TestDivideMultiply(t *testing.T) {
	checkDivideMultiply(t, Math_DD_Pi, Math_DD_E, 1e-30)
	checkDivideMultiply(t, Math_NewDDFromFloat64(39.4), Math_NewDDFromFloat64(10), 1e-30)
}

func TestSqrt(t *testing.T) {
	// The appropriate error bound is determined empirically.
	checkSqrt(t, Math_DD_Pi, 1e-30)
	checkSqrt(t, Math_DD_E, 1e-30)
	checkSqrt(t, Math_NewDDFromFloat64(999.0), 1e-28)
}

func checkSqrt(t *testing.T, x *Math_DD, errBound float64) {
	t.Helper()
	sqrt := x.Sqrt()
	x2 := sqrt.Multiply(sqrt)
	checkErrorBound(t, "Sqrt", x, x2, errBound)
}

func TestTrunc(t *testing.T) {
	checkTrunc(t, Math_DD_ValueOfFloat64(1e16).Subtract(Math_DD_ValueOfFloat64(1)),
		Math_DD_ValueOfFloat64(1e16).Subtract(Math_DD_ValueOfFloat64(1)))
	// The appropriate error bound is determined empirically.
	checkTrunc(t, Math_DD_Pi, Math_DD_ValueOfFloat64(3))
	checkTrunc(t, Math_DD_ValueOfFloat64(999.999), Math_DD_ValueOfFloat64(999))

	checkTrunc(t, Math_DD_E.Negate(), Math_DD_ValueOfFloat64(-2))
	checkTrunc(t, Math_DD_ValueOfFloat64(-999.999), Math_DD_ValueOfFloat64(-999))
}

func checkTrunc(t *testing.T, x, expected *Math_DD) {
	t.Helper()
	trunc := x.Trunc()
	isEqual := trunc.Equals(expected)
	junit.AssertTrue(t, isEqual)
}

func TestPow(t *testing.T) {
	checkPow(t, 0, 3, 16*Math_DD_Eps)
	checkPow(t, 14, 3, 16*Math_DD_Eps)
	checkPow(t, 3, -5, 16*Math_DD_Eps)
	checkPow(t, -3, 5, 16*Math_DD_Eps)
	checkPow(t, -3, -5, 16*Math_DD_Eps)
	checkPow(t, 0.12345, -5, 1e5*Math_DD_Eps)
}

func TestReciprocal(t *testing.T) {
	// Error bounds are chosen to be "close enough" (i.e. heuristically).

	// For some reason many reciprocals are exact.
	checkReciprocal(t, 3.0, 0)
	checkReciprocal(t, 99.0, 1e-29)
	checkReciprocal(t, 999.0, 0)
	checkReciprocal(t, 314159269.0, 0)
}

func TestDeterminant(t *testing.T) {
	checkDeterminant(t, 3, 8, 4, 6, -14, 0)
	checkDeterminantDD(t, 3, 8, 4, 6, -14, 0)
}

func TestDeterminantRobust(t *testing.T) {
	checkDeterminant(t, 1.0e9, 1.0e9-1, 1.0e9-1, 1.0e9-2, -1, 0)
	checkDeterminantDD(t, 1.0e9, 1.0e9-1, 1.0e9-1, 1.0e9-2, -1, 0)
}

func checkDeterminant(t *testing.T, x1, y1, x2, y2, expected, errBound float64) {
	t.Helper()
	det := Math_DD_DeterminantFloat64(x1, y1, x2, y2)
	checkErrorBound(t, "Determinant", det, Math_DD_ValueOfFloat64(expected), errBound)
}

func checkDeterminantDD(t *testing.T, x1, y1, x2, y2, expected, errBound float64) {
	t.Helper()
	det := Math_DD_DeterminantDD(
		Math_DD_ValueOfFloat64(x1), Math_DD_ValueOfFloat64(y1),
		Math_DD_ValueOfFloat64(x2), Math_DD_ValueOfFloat64(y2))
	checkErrorBound(t, "Determinant", det, Math_DD_ValueOfFloat64(expected), errBound)
}

func TestBinom(t *testing.T) {
	checkBinomialSquare(t, 100.0, 1.0)
	checkBinomialSquare(t, 1000.0, 1.0)
	checkBinomialSquare(t, 10000.0, 1.0)
	checkBinomialSquare(t, 100000.0, 1.0)
	checkBinomialSquare(t, 1000000.0, 1.0)
	checkBinomialSquare(t, 1e8, 1.0)
	checkBinomialSquare(t, 1e10, 1.0)
	checkBinomialSquare(t, 1e14, 1.0)
	// Following call will fail, because it requires 32 digits of precision.
	// checkBinomialSquare(t, 1e16, 1.0)

	checkBinomialSquare(t, 1e14, 291.0)
	checkBinomialSquare(t, 5e14, 291.0)
	checkBinomialSquare(t, 5e14, 345291.0)
}

func checkAddMult2(t *testing.T, dd *Math_DD) {
	t.Helper()
	sum := dd.Add(dd)
	prod := dd.Multiply(Math_NewDDFromFloat64(2.0))
	checkErrorBound(t, "AddMult2", sum, prod, 0.0)
}

func checkMultiplyDivide(t *testing.T, a, b *Math_DD, errBound float64) {
	t.Helper()
	a2 := a.Multiply(b).Divide(b)
	checkErrorBound(t, "MultiplyDivide", a, a2, errBound)
}

func checkDivideMultiply(t *testing.T, a, b *Math_DD, errBound float64) {
	t.Helper()
	a2 := a.Divide(b).Multiply(b)
	checkErrorBound(t, "DivideMultiply", a, a2, errBound)
}

func math_ddBasicTest_delta(x, y *Math_DD) *Math_DD {
	return x.Subtract(y).Abs()
}

func checkErrorBound(t *testing.T, tag string, x, y *Math_DD, errBound float64) {
	t.Helper()
	err := x.Subtract(y).Abs()
	isWithinEps := err.DoubleValue() <= errBound
	junit.AssertTrue(t, isWithinEps)
}

func checkBinomialSquare(t *testing.T, a, b float64) {
	t.Helper()
	// Binomial square.
	add := Math_NewDDFromFloat64(a)
	bdd := Math_NewDDFromFloat64(b)
	aPlusb := add.Add(bdd)
	abSq := aPlusb.Multiply(aPlusb)

	// Expansion.
	a2dd := add.Multiply(add)
	b2dd := bdd.Multiply(bdd)
	ab := add.Multiply(bdd)
	sum := b2dd.Add(ab).Add(ab)

	diff := abSq.Subtract(a2dd)

	delta := diff.Subtract(sum)

	math_ddBasicTest_printBinomialSquareDouble(a, b)

	isSame := diff.Equals(sum)
	junit.AssertTrue(t, isSame)
	isDeltaZero := delta.IsZero()
	junit.AssertTrue(t, isDeltaZero)
}

func math_ddBasicTest_printBinomialSquareDouble(a, b float64) {
	_ = 2*a*b + b*b
	_ = (a+b)*(a+b) - a*a
}

func TestBinomial2(t *testing.T) {
	checkBinomial2(t, 100.0, 1.0)
	checkBinomial2(t, 1000.0, 1.0)
	checkBinomial2(t, 10000.0, 1.0)
	checkBinomial2(t, 100000.0, 1.0)
	checkBinomial2(t, 1000000.0, 1.0)
	checkBinomial2(t, 1e8, 1.0)
	checkBinomial2(t, 1e10, 1.0)
	checkBinomial2(t, 1e14, 1.0)

	checkBinomial2(t, 1e14, 291.0)

	checkBinomial2(t, 5e14, 291.0)
	checkBinomial2(t, 5e14, 345291.0)
}

func checkBinomial2(t *testing.T, a, b float64) {
	t.Helper()
	// Binomial product.
	add := Math_NewDDFromFloat64(a)
	bdd := Math_NewDDFromFloat64(b)
	aPlusb := add.Add(bdd)
	aSubb := add.Subtract(bdd)
	abProd := aPlusb.Multiply(aSubb)

	// Expansion.
	a2dd := add.Multiply(add)
	b2dd := bdd.Multiply(bdd)

	// This should equal b^2.
	diff := abProd.Subtract(a2dd).Negate()

	delta := diff.Subtract(b2dd)

	isSame := diff.Equals(b2dd)
	junit.AssertTrue(t, isSame)
	isDeltaZero := delta.IsZero()
	junit.AssertTrue(t, isDeltaZero)
}

func checkReciprocal(t *testing.T, x float64, errBound float64) {
	t.Helper()
	xdd := Math_NewDDFromFloat64(x)
	rr := xdd.Reciprocal().Reciprocal()

	err := xdd.Subtract(rr).DoubleValue()

	junit.AssertTrue(t, err <= errBound)
}

func checkPow(t *testing.T, x float64, exp int, errBound float64) {
	t.Helper()
	xdd := Math_NewDDFromFloat64(x)
	pow := xdd.Pow(exp)
	pow2 := math_ddBasicTest_slowPow(xdd, exp)

	err := pow.Subtract(pow2).DoubleValue()

	junit.AssertTrue(t, err <= errBound)
}

func math_ddBasicTest_slowPow(x *Math_DD, exp int) *Math_DD {
	if exp == 0 {
		return Math_DD_ValueOfFloat64(1.0)
	}

	n := exp
	if n < 0 {
		n = -n
	}
	// MD - could use binary exponentiation for better precision & speed
	pow := Math_NewDDFromDD(x)
	for i := 1; i < n; i++ {
		pow = pow.Multiply(x)
	}
	if exp < 0 {
		return pow.Reciprocal()
	}
	return pow
}
