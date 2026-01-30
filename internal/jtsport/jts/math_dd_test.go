package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

const math_ddTest_valueDbl = 2.2

func TestSetValueDouble(t *testing.T) {
	junit.AssertTrue(t, math_ddTest_valueDbl == Math_NewDDFromFloat64(1).SetValueFloat64(math_ddTest_valueDbl).DoubleValue())
}

func TestSetValueDD(t *testing.T) {
	junit.AssertTrue(t, Math_NewDDFromFloat64(math_ddTest_valueDbl).Equals(Math_NewDDFromFloat64(1).SetValue(Math_NewDDFromFloat64(2.2))))
	junit.AssertTrue(t, Math_DD_Pi.Equals(Math_NewDDFromFloat64(1).SetValue(Math_DD_Pi)))
}

func TestCopy(t *testing.T) {
	junit.AssertTrue(t, Math_NewDDFromFloat64(math_ddTest_valueDbl).Equals(Math_DD_Copy(Math_NewDDFromFloat64(math_ddTest_valueDbl))))
	junit.AssertTrue(t, Math_DD_Pi.Equals(Math_DD_Copy(Math_DD_Pi)))
}
