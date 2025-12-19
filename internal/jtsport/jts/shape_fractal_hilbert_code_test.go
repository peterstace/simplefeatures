package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// Tests ported from HilbertCodeTest.java.

func TestHilbertCodeSize(t *testing.T) {
	checkHilbertCodeSize(t, 0, 1)
	checkHilbertCodeSize(t, 1, 4)
	checkHilbertCodeSize(t, 2, 16)
	checkHilbertCodeSize(t, 3, 64)
	checkHilbertCodeSize(t, 4, 256)
	checkHilbertCodeSize(t, 5, 1024)
	checkHilbertCodeSize(t, 6, 4096)
}

func checkHilbertCodeSize(t *testing.T, level, expected int) {
	t.Helper()
	actual := jts.ShapeFractal_HilbertCode_Size(level)
	if actual != expected {
		t.Errorf("Size(%d): expected %d, got %d", level, expected, actual)
	}
}

func TestHilbertCodeLevel(t *testing.T) {
	checkHilbertCodeLevel(t, 1, 0)

	checkHilbertCodeLevel(t, 2, 1)
	checkHilbertCodeLevel(t, 3, 1)
	checkHilbertCodeLevel(t, 4, 1)

	checkHilbertCodeLevel(t, 5, 2)
	checkHilbertCodeLevel(t, 13, 2)
	checkHilbertCodeLevel(t, 15, 2)
	checkHilbertCodeLevel(t, 16, 2)

	checkHilbertCodeLevel(t, 17, 3)
	checkHilbertCodeLevel(t, 63, 3)
	checkHilbertCodeLevel(t, 64, 3)

	checkHilbertCodeLevel(t, 65, 4)
	checkHilbertCodeLevel(t, 255, 4)
	checkHilbertCodeLevel(t, 256, 4)
}

func checkHilbertCodeLevel(t *testing.T, numPoints, expected int) {
	t.Helper()
	actual := jts.ShapeFractal_HilbertCode_Level(numPoints)
	if actual != expected {
		t.Errorf("Level(%d): expected %d, got %d", numPoints, expected, actual)
	}
}

func TestHilbertCodeDecode(t *testing.T) {
	checkHilbertCodeDecode(t, 1, 0, 0, 0)

	checkHilbertCodeDecode(t, 1, 0, 0, 0)
	checkHilbertCodeDecode(t, 1, 1, 0, 1)

	checkHilbertCodeDecode(t, 3, 0, 0, 0)
	checkHilbertCodeDecode(t, 3, 1, 0, 1)

	checkHilbertCodeDecode(t, 4, 0, 0, 0)
	checkHilbertCodeDecode(t, 4, 1, 1, 0)
	checkHilbertCodeDecode(t, 4, 24, 6, 2)
	checkHilbertCodeDecode(t, 4, 255, 15, 0)

	checkHilbertCodeDecode(t, 5, 124, 8, 6)
}

func checkHilbertCodeDecode(t *testing.T, order, index, x, y int) {
	t.Helper()
	p := jts.ShapeFractal_HilbertCode_Decode(order, index)
	actualX := int(p.GetX())
	actualY := int(p.GetY())
	if actualX != x {
		t.Errorf("Decode(%d, %d).X: expected %d, got %d", order, index, x, actualX)
	}
	if actualY != y {
		t.Errorf("Decode(%d, %d).Y: expected %d, got %d", order, index, y, actualY)
	}
}

func TestHilbertCodeDecodeEncode(t *testing.T) {
	checkHilbertCodeDecodeEncodeForLevel(t, 4)
	checkHilbertCodeDecodeEncodeForLevel(t, 5)
}

func checkHilbertCodeDecodeEncodeForLevel(t *testing.T, level int) {
	t.Helper()
	n := jts.ShapeFractal_HilbertCode_Size(level)
	for i := 0; i < n; i++ {
		checkHilbertCodeDecodeEncode(t, level, i)
	}
}

func checkHilbertCodeDecodeEncode(t *testing.T, level, index int) {
	t.Helper()
	p := jts.ShapeFractal_HilbertCode_Decode(level, index)
	encode := jts.ShapeFractal_HilbertCode_Encode(level, int(p.GetX()), int(p.GetY()))
	if encode != index {
		t.Errorf("DecodeEncode(%d, %d): expected %d, got %d", level, index, index, encode)
	}
}
