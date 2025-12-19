package jts

import (
	stdmath "math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestWriteStandardNotation(t *testing.T) {
	// Standard cases.
	checkStandardNotationFloat64(t, 1.0, "1.0")
	checkStandardNotationFloat64(t, 0.0, "0.0")

	// Cases where hi is a power of 10 and lo is negative.
	checkStandardNotation(t, Math_DD_ValueOfFloat64(1e12).Subtract(Math_DD_ValueOfFloat64(1)), "999999999999.0")
	checkStandardNotation(t, Math_DD_ValueOfFloat64(1e14).Subtract(Math_DD_ValueOfFloat64(1)), "99999999999999.0")
	checkStandardNotation(t, Math_DD_ValueOfFloat64(1e16).Subtract(Math_DD_ValueOfFloat64(1)), "9999999999999999.0")

	num8Dec := Math_DD_ValueOfFloat64(-379363639).Divide(
		Math_DD_ValueOfFloat64(100000000))
	checkStandardNotation(t, num8Dec, "-3.79363639")

	checkStandardNotation(t, Math_NewDDFromHiLo(-3.79363639, 8.039137357367426e-17),
		"-3.7936363900000000000000000")

	checkStandardNotation(t, Math_DD_ValueOfFloat64(34).Divide(
		Math_DD_ValueOfFloat64(1000)), "0.034")
	checkStandardNotationFloat64(t, 1.05e3, "1050.0")
	checkStandardNotationFloat64(t, 0.34, "0.34000000000000002442490654175344")
	checkStandardNotation(t, Math_DD_ValueOfFloat64(34).Divide(
		Math_DD_ValueOfFloat64(100)), "0.34")
	checkStandardNotationFloat64(t, 14, "14.0")
}

func checkStandardNotationFloat64(t *testing.T, x float64, expectedStr string) {
	t.Helper()
	checkStandardNotation(t, Math_DD_ValueOfFloat64(x), expectedStr)
}

func checkStandardNotation(t *testing.T, x *Math_DD, expectedStr string) {
	t.Helper()
	xStr := x.ToStandardNotation()
	junit.AssertEquals(t, expectedStr, xStr)
}

func TestWriteSciNotation(t *testing.T) {
	checkSciNotationFloat64(t, 0.0, "0.0E0")
	checkSciNotationFloat64(t, 1.05e10, "1.05E10")
	checkSciNotationFloat64(t, 0.34, "3.4000000000000002442490654175344E-1")
	checkSciNotation(t, Math_DD_ValueOfFloat64(34).Divide(Math_DD_ValueOfFloat64(100)), "3.4E-1")
	checkSciNotationFloat64(t, 14, "1.4E1")
}

func checkSciNotationFloat64(t *testing.T, x float64, expectedStr string) {
	t.Helper()
	checkSciNotation(t, Math_DD_ValueOfFloat64(x), expectedStr)
}

func checkSciNotation(t *testing.T, x *Math_DD, expectedStr string) {
	t.Helper()
	xStr := x.ToSciNotation()
	junit.AssertEquals(t, xStr, expectedStr)
}

func TestParseInt(t *testing.T) {
	checkParse(t, "0", 0, 1e-32)
	checkParse(t, "00", 0, 1e-32)
	checkParse(t, "000", 0, 1e-32)

	checkParse(t, "1", 1, 1e-32)
	checkParse(t, "100", 100, 1e-32)
	checkParse(t, "00100", 100, 1e-32)

	checkParse(t, "-1", -1, 1e-32)
	checkParse(t, "-01", -1, 1e-32)
	checkParse(t, "-123", -123, 1e-32)
	checkParse(t, "-00123", -123, 1e-32)
}

func TestParseStandardNotation(t *testing.T) {
	checkParse(t, "1.0000000", 1, 1e-32)
	checkParse(t, "1.0", 1, 1e-32)
	checkParse(t, "1.", 1, 1e-32)
	checkParse(t, "01.", 1, 1e-32)

	checkParse(t, "-1.0", -1, 1e-32)
	checkParse(t, "-1.", -1, 1e-32)
	checkParse(t, "-01.0", -1, 1e-32)
	checkParse(t, "-123.0", -123, 1e-32)

	// The Java double-precision constant 1.4 gives rise to a value which
	// differs from the exact binary representation down around the 17th decimal
	// place. Thus it will not compare exactly to the Math_DD
	// representation of the same number. To avoid this, compute the expected
	// value using full Math_DD precision.
	checkParseDD(t, "1.4", Math_DD_ValueOfFloat64(14).Divide(Math_DD_ValueOfFloat64(10)), 1e-30)

	// 39.5D can be converted to an exact FP representation.
	checkParse(t, "39.5", 39.5, 1e-30)
	checkParse(t, "-39.5", -39.5, 1e-30)
}

func TestParseSciNotation(t *testing.T) {
	checkParse(t, "1.05e10", 1.05e10, 1e-32)
	checkParse(t, "01.05e10", 1.05e10, 1e-32)
	checkParse(t, "12.05e10", 1.205e11, 1e-32)

	checkParse(t, "-1.05e10", -1.05e10, 1e-32)

	checkParseDD(t, "1.05e-10", Math_DD_ValueOfFloat64(105.).Divide(
		Math_DD_ValueOfFloat64(100.)).Divide(Math_DD_ValueOfFloat64(1.0e10)), 1e-32)
	checkParseDD(t, "-1.05e-10", Math_DD_ValueOfFloat64(105.).Divide(
		Math_DD_ValueOfFloat64(100.)).Divide(Math_DD_ValueOfFloat64(1.0e10)).
		Negate(), 1e-32)
}

func checkParse(t *testing.T, str string, expectedVal float64, relErrBound float64) {
	t.Helper()
	checkParseDD(t, str, Math_NewDDFromFloat64(expectedVal), relErrBound)
}

func checkParseDD(t *testing.T, str string, expectedVal *Math_DD, relErrBound float64) {
	t.Helper()
	xdd, err := Math_DD_Parse(str)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", str, err)
	}
	errVal := xdd.Subtract(expectedVal).DoubleValue()
	xddd := xdd.DoubleValue()
	var relErr float64
	if xddd == 0 {
		relErr = errVal
	} else {
		relErr = stdmath.Abs(errVal / xddd)
	}
	junit.AssertTrue(t, relErr <= relErrBound)
}

func TestParseError(t *testing.T) {
	checkParseError(t, "-1.05E2w")
	checkParseError(t, "%-1.05E2w")
	checkParseError(t, "-1.0512345678t")
}

func checkParseError(t *testing.T, str string) {
	t.Helper()
	_, err := Math_DD_Parse(str)
	foundParseError := err != nil
	junit.AssertTrue(t, foundParseError)
}

func TestWriteRepeatedSqrt(t *testing.T) {
	writeRepeatedSqrt(t, Math_DD_ValueOfFloat64(1.0))
	writeRepeatedSqrt(t, Math_DD_ValueOfFloat64(.999999999999))
	writeRepeatedSqrt(t, Math_DD_Pi.Divide(Math_DD_ValueOfFloat64(10)))
}

func writeRepeatedSqrt(t *testing.T, xdd *Math_DD) {
	t.Helper()
	count := 0
	for xdd.DoubleValue() > 1e-300 {
		count++

		x := xdd.DoubleValue()
		xSqrt := xdd.Sqrt()
		s := xSqrt.String()

		xSqrt2, err := Math_DD_Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", s, err)
		}
		xx := xSqrt2.Multiply(xSqrt2)
		_ = stdmath.Abs(xx.DoubleValue() - x)

		xdd = xSqrt

		// Square roots converge on 1 - stop when very close.
		distFrom1DD := xSqrt.Subtract(Math_DD_ValueOfFloat64(1.0))
		distFrom1 := distFrom1DD.DoubleValue()
		if stdmath.Abs(distFrom1) < 1.0e-40 {
			break
		}
	}
	_ = count
}

func TestWriteRepeatedSqr(t *testing.T) {
	writeRepeatedSqr(t, Math_DD_ValueOfFloat64(.9))
	writeRepeatedSqr(t, Math_DD_Pi.Divide(Math_DD_ValueOfFloat64(10)))
}

func writeRepeatedSqr(t *testing.T, xdd *Math_DD) {
	t.Helper()
	if xdd.Ge(Math_DD_ValueOfFloat64(1)) {
		panic("Argument must be < 1")
	}

	count := 0
	for xdd.DoubleValue() > 1e-300 {
		count++
		if count == 100 {
			count = count
		}
		_ = xdd.DoubleValue()
		xSqr := xdd.Sqr()
		s := xSqr.String()

		_, err := Math_DD_Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", s, err)
		}

		xdd = xSqr
	}
}

func TestWriteSquaresStress(t *testing.T) {
	for i := 1; i < 10000; i++ {
		writeAndReadSqrt(t, float64(i))
	}
}

func writeAndReadSqrt(t *testing.T, x float64) {
	t.Helper()
	xdd := Math_DD_ValueOfFloat64(x)
	xSqrt := xdd.Sqrt()
	s := xSqrt.String()

	xSqrt2, err := Math_DD_Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", s, err)
	}
	xx := xSqrt2.Multiply(xSqrt2)
	xxStr := xx.String()

	xx2, err := Math_DD_Parse(xxStr)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", xxStr, err)
	}
	errVal := stdmath.Abs(xx2.DoubleValue() - x)
	junit.AssertTrue(t, errVal < 1e-10)
}
