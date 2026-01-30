package jts

import (
	stdmath "math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestEByTaylorSeries(t *testing.T) {
	testE := computeEByTaylorSeries()
	err := stdmath.Abs(testE.Subtract(Math_DD_E).DoubleValue())
	junit.AssertTrue(t, err < 64*Math_DD_Eps)
}

func computeEByTaylorSeries() *Math_DD {
	s := Math_DD_ValueOfFloat64(2.0)
	ddT := Math_DD_ValueOfFloat64(1.0)
	n := 1.0
	i := 0

	for ddT.DoubleValue() > Math_DD_Eps {
		i++
		n += 1.0
		ddT = ddT.Divide(Math_DD_ValueOfFloat64(n))
		s = s.Add(ddT)
	}
	_ = i
	return s
}

func TestPiByMachin(t *testing.T) {
	testE := computePiByMachin()
	err := stdmath.Abs(testE.Subtract(Math_DD_Pi).DoubleValue())
	junit.AssertTrue(t, err < 8*Math_DD_Eps)
}

func computePiByMachin() *Math_DD {
	t1 := Math_DD_ValueOfFloat64(1.0).Divide(Math_DD_ValueOfFloat64(5.0))
	t2 := Math_DD_ValueOfFloat64(1.0).Divide(Math_DD_ValueOfFloat64(239.0))

	pi4 := Math_DD_ValueOfFloat64(4.0).
		Multiply(arctan(t1)).
		Subtract(arctan(t2))
	pi := Math_DD_ValueOfFloat64(4.0).Multiply(pi4)
	return pi
}

func arctan(x *Math_DD) *Math_DD {
	ddT := x
	t2 := ddT.Sqr()
	at := Math_NewDDFromFloat64(0.0)
	two := Math_NewDDFromFloat64(2.0)
	k := 0
	d := Math_NewDDFromFloat64(1.0)
	sign := 1
	for ddT.DoubleValue() > Math_DD_Eps {
		k++
		if sign < 0 {
			at = at.Subtract(ddT.Divide(d))
		} else {
			at = at.Add(ddT.Divide(d))
		}

		d = d.Add(two)
		ddT = ddT.Multiply(t2)
		sign = -sign
	}
	_ = k
	return at
}
