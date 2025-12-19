package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestAreaOfRing(t *testing.T) {
	ring := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(100, 200),
		Geom_NewCoordinateWithXY(200, 200),
		Geom_NewCoordinateWithXY(200, 100),
		Geom_NewCoordinateWithXY(100, 100),
		Geom_NewCoordinateWithXY(100, 200),
	}

	junit.AssertEquals(t, 10000.0, Algorithm_Area_OfRing(ring))

	seq := GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures(ring, 2, 0)
	junit.AssertEquals(t, 10000.0, Algorithm_Area_OfRingSeq(seq))
}

func TestAreaOfRingSignedCW(t *testing.T) {
	ring := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(100, 200),
		Geom_NewCoordinateWithXY(200, 200),
		Geom_NewCoordinateWithXY(200, 100),
		Geom_NewCoordinateWithXY(100, 100),
		Geom_NewCoordinateWithXY(100, 200),
	}

	junit.AssertEquals(t, 10000.0, Algorithm_Area_OfRingSigned(ring))

	seq := GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures(ring, 2, 0)
	junit.AssertEquals(t, 10000.0, Algorithm_Area_OfRingSignedSeq(seq))
}

func TestAreaOfRingSignedCCW(t *testing.T) {
	ring := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(100, 200),
		Geom_NewCoordinateWithXY(100, 100),
		Geom_NewCoordinateWithXY(200, 100),
		Geom_NewCoordinateWithXY(200, 200),
		Geom_NewCoordinateWithXY(100, 200),
	}

	junit.AssertEquals(t, -10000.0, Algorithm_Area_OfRingSigned(ring))

	seq := GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures(ring, 2, 0)
	junit.AssertEquals(t, -10000.0, Algorithm_Area_OfRingSignedSeq(seq))
}
