package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestCoordinateListForward(t *testing.T) {
	checkCoordListValue(t, coordListFromOrds(0, 0, 1, 1, 2, 2).ToCoordinateArrayWithDirection(true),
		0, 0, 1, 1, 2, 2)
}

func TestCoordinateListReverse(t *testing.T) {
	checkCoordListValue(t, coordListFromOrds(0, 0, 1, 1, 2, 2).ToCoordinateArrayWithDirection(false),
		2, 2, 1, 1, 0, 0)
}

func TestCoordinateListReverseEmpty(t *testing.T) {
	checkCoordListValue(t, coordListFromOrds().ToCoordinateArrayWithDirection(false))
}

func checkCoordListValue(t *testing.T, coordArray []*jts.Geom_Coordinate, ords ...float64) {
	t.Helper()
	if len(coordArray)*2 != len(ords) {
		t.Fatalf("length mismatch: coordArray len %d, ords len %d", len(coordArray), len(ords))
	}
	for i := 0; i < len(coordArray); i++ {
		pt := coordArray[i]
		if pt.GetX() != ords[2*i] {
			t.Errorf("coordinate[%d].X: expected %v, got %v", i, ords[2*i], pt.GetX())
		}
		if pt.GetY() != ords[2*i+1] {
			t.Errorf("coordinate[%d].Y: expected %v, got %v", i, ords[2*i+1], pt.GetY())
		}
	}
}

func coordListFromOrds(ords ...float64) *jts.Geom_CoordinateList {
	cl := jts.Geom_NewCoordinateList()
	for i := 0; i < len(ords); i += 2 {
		cl.AddCoordinate(jts.Geom_NewCoordinateWithXY(ords[i], ords[i+1]), false)
	}
	return cl
}
