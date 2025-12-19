package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

// Tests ported from OrdinateFormatTest.java.

func TestOrdinateFormatLargeNumber(t *testing.T) {
	// Ensure scientific notation is not used.
	checkOrdinateFormat(t, 1234567890.0, "1234567890")
}

func TestOrdinateFormatVeryLargeNumber(t *testing.T) {
	// Ensure scientific notation is not used.
	// Note output is rounded since it exceeds double precision accuracy.
	checkOrdinateFormat(t, 12345678901234567890.0, "12345678901234567000")
}

func TestOrdinateFormatDecimalPoint(t *testing.T) {
	checkOrdinateFormat(t, 1.123, "1.123")
}

func TestOrdinateFormatNegative(t *testing.T) {
	checkOrdinateFormat(t, -1.123, "-1.123")
}

func TestOrdinateFormatFractionDigits(t *testing.T) {
	checkOrdinateFormat(t, 1.123456789012345, "1.123456789012345")
	checkOrdinateFormat(t, 0.0123456789012345, "0.0123456789012345")
}

func TestOrdinateFormatLimitedFractionDigits(t *testing.T) {
	checkOrdinateFormatWithMaxDigits(t, 1.123456789012345, 2, "1.12")
	checkOrdinateFormatWithMaxDigits(t, 1.123456789012345, 3, "1.123")
	checkOrdinateFormatWithMaxDigits(t, 1.123456789012345, 4, "1.1235")
	checkOrdinateFormatWithMaxDigits(t, 1.123456789012345, 5, "1.12346")
	checkOrdinateFormatWithMaxDigits(t, 1.123456789012345, 6, "1.123457")
}

func TestOrdinateFormatMaximumFractionDigits(t *testing.T) {
	checkOrdinateFormat(t, 0.0000000000123456789012345, "0.0000000000123456789012345")
}

func TestOrdinateFormatPi(t *testing.T) {
	checkOrdinateFormat(t, math.Pi, "3.141592653589793")
}

func TestOrdinateFormatNaN(t *testing.T) {
	checkOrdinateFormat(t, math.NaN(), "NaN")
}

func TestOrdinateFormatInf(t *testing.T) {
	checkOrdinateFormat(t, math.Inf(1), "Inf")
	checkOrdinateFormat(t, math.Inf(-1), "-Inf")
}

func checkOrdinateFormat(t *testing.T, d float64, expected string) {
	actual := jts.Io_OrdinateFormat_Default.Format(d)
	junit.AssertEquals(t, expected, actual)
}

func checkOrdinateFormatWithMaxDigits(t *testing.T, d float64, maxFractionDigits int, expected string) {
	format := jts.Io_OrdinateFormat_Create(maxFractionDigits)
	actual := format.Format(d)
	junit.AssertEquals(t, expected, actual)
}

func checkOrdinateFormatAllLocales(t *testing.T, d float64, maxFractionDigits int, expected string) {
	format := jts.Io_OrdinateFormat_Create(maxFractionDigits)
	actual := format.Format(d)
	junit.AssertEquals(t, expected, actual)
}

func checkOrdinateFormatLocales(t *testing.T, locale any, d float64, maxFractionDigits int, expected string) {
	format := jts.Io_OrdinateFormat_Create(maxFractionDigits)
	actual := format.Format(d)
	junit.AssertEquals(t, expected, actual)
}
