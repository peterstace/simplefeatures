package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// Tests ported from PackedCoordinateSequenceFloatTest.java.

func TestPackedCoordinateSequenceFloat4dCoordinateSequence(t *testing.T) {
	cs := jts.GeomImpl_NewPackedCoordinateSequenceFactoryWithType(jts.GeomImpl_PackedCoordinateSequenceFactory_FLOAT).
		CreateFromFloatsWithMeasures([]float32{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0}, 4, 1)
	if cs.GetCoordinate(0).GetZ() != 2.0 {
		t.Errorf("expected Z=2.0, got %v", cs.GetCoordinate(0).GetZ())
	}
	if cs.GetCoordinate(0).GetM() != 3.0 {
		t.Errorf("expected M=3.0, got %v", cs.GetCoordinate(0).GetM())
	}
}
