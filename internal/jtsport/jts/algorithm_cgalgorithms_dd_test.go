package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestCGAlgorithmsDDSignOfDet2x2(t *testing.T) {
	junit.AssertEquals(t, 0, Algorithm_CGAlgorithmsDD_SignOfDet2x2Float64(1, 1, 2, 2))
	junit.AssertEquals(t, 1, Algorithm_CGAlgorithmsDD_SignOfDet2x2Float64(1, 1, 2, 3))
	junit.AssertEquals(t, -1, Algorithm_CGAlgorithmsDD_SignOfDet2x2Float64(1, 1, 3, 2))
}
