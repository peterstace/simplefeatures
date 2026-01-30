package jts

import (
	"math"
	"strconv"
	"strings"
)

const io_ordinateFormat_decimalPattern = "0"

// Io_OrdinateFormat_RepPosInf is the output representation of positive infinity.
const Io_OrdinateFormat_RepPosInf = "Inf"

// Io_OrdinateFormat_RepNegInf is the output representation of negative infinity.
const Io_OrdinateFormat_RepNegInf = "-Inf"

// Io_OrdinateFormat_RepNaN is the output representation of NaN.
const Io_OrdinateFormat_RepNaN = "NaN"

// Io_OrdinateFormat_MaxFractionDigits is the maximum number of fraction digits
// to support output of reasonable ordinate values.
// The default is chosen to allow representing the smallest possible IEEE-754
// double-precision value, although this is not expected to occur (and is not
// supported by other areas of the JTS code).
const Io_OrdinateFormat_MaxFractionDigits = 325

// Io_OrdinateFormat_Default is the default formatter using the maximum number
// of digits in the fraction portion of a number.
var Io_OrdinateFormat_Default = Io_NewOrdinateFormat()

// Io_OrdinateFormat_Create creates a new formatter with the given maximum
// number of digits in the fraction portion of a number.
func Io_OrdinateFormat_Create(maximumFractionDigits int) *Io_OrdinateFormat {
	return Io_NewOrdinateFormatWithMaxFractionDigits(maximumFractionDigits)
}

// Io_OrdinateFormat formats numeric values for ordinates in a consistent,
// accurate way. The format has the following characteristics:
// - It is consistent in all locales (the decimal separator is always a period).
// - Scientific notation is never output, even for very large numbers.
// - The maximum number of decimal places reflects the available precision.
// - NaN values are represented as "NaN".
// - Inf values are represented as "Inf" or "-Inf".
type Io_OrdinateFormat struct {
	maximumFractionDigits int
}

// Io_NewOrdinateFormat creates an OrdinateFormat using the default maximum
// number of fraction digits.
func Io_NewOrdinateFormat() *Io_OrdinateFormat {
	return &Io_OrdinateFormat{maximumFractionDigits: Io_OrdinateFormat_MaxFractionDigits}
}

// Io_NewOrdinateFormatWithMaxFractionDigits creates an OrdinateFormat using
// the given maximum number of fraction digits.
func Io_NewOrdinateFormatWithMaxFractionDigits(maximumFractionDigits int) *Io_OrdinateFormat {
	return &Io_OrdinateFormat{maximumFractionDigits: maximumFractionDigits}
}

// io_ordinateFormat_createFormat is not needed in Go (uses strconv.FormatFloat
// directly). This is a placeholder to maintain 1-1 correspondence with Java's
// private static createFormat method.

// Format returns a string representation of the given ordinate numeric value.
func (f *Io_OrdinateFormat) Format(ord float64) string {
	// FUTURE: If it seems better to use scientific notation for very large/small
	// numbers then this can be done here.

	if math.IsNaN(ord) {
		return Io_OrdinateFormat_RepNaN
	}
	if math.IsInf(ord, 1) {
		return Io_OrdinateFormat_RepPosInf
	}
	if math.IsInf(ord, -1) {
		return Io_OrdinateFormat_RepNegInf
	}

	// Format the number without scientific notation.
	// Use -1 precision first to get the full representation.
	s := strconv.FormatFloat(ord, 'f', -1, 64)

	// Check if we need to limit fraction digits (with rounding).
	if dotIdx := strings.Index(s, "."); dotIdx >= 0 {
		fractionLen := len(s) - dotIdx - 1
		if fractionLen > f.maximumFractionDigits {
			// Re-format with the desired precision to get proper rounding.
			s = strconv.FormatFloat(ord, 'f', f.maximumFractionDigits, 64)
		}
	}

	return s
}
