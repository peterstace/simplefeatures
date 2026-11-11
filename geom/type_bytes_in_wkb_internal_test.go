package geom

import "testing"

// wkbTestGeometries returns a geometry of each type, including empty and null
// edge cases.
func wkbTestGeometries() []Geometry {
	seqXY := NewSequence([]float64{0, 0, 1, 0, 1, 1, 0, 0}, DimXY)
	seqXYZM := NewSequence([]float64{0, 0, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4}, DimXYZM)
	ring := NewLineString(seqXY)
	poly := NewPolygon([]LineString{ring})
	polyZ := NewPolygon([]LineString{NewLineString(seqXYZM)})
	pt := NewPoint(Coordinates{XY: XY{1, 2}, Type: DimXYZ, Z: 3})
	emptyPt := NewEmptyPoint(DimXYM)
	mp := NewMultiPoint([]Point{pt, emptyPt})
	mls := NewMultiLineString([]LineString{ring, NewLineString(seqXYZM)})
	mpoly := NewMultiPolygon([]Polygon{poly, polyZ})
	gc := NewGeometryCollection([]Geometry{
		pt.AsGeometry(),
		ring.AsGeometry(),
		poly.AsGeometry(),
		mp.AsGeometry(),
		mls.AsGeometry(),
		mpoly.AsGeometry(),
		NewGeometryCollection(nil).AsGeometry(),
	})
	return []Geometry{
		pt.AsGeometry(),
		emptyPt.AsGeometry(),
		ring.AsGeometry(),
		poly.AsGeometry(),
		mp.AsGeometry(),
		mls.AsGeometry(),
		mpoly.AsGeometry(),
		gc.AsGeometry(),
		{},
	}
}

// TestBytesInWKBMatchesAsBinary checks that AsBinary allocates its output
// buffer once, sized exactly to the WKB (so cap equals len and no growth
// happens during appending).
func TestBytesInWKBMatchesAsBinary(t *testing.T) {
	for i, g := range wkbTestGeometries() {
		b := g.AsBinary()
		if cap(b) != len(b) {
			t.Errorf("AsBinary buffer not sized exactly: geometry %d len=%d cap=%d", i, len(b), cap(b))
		}
		allocs := testing.AllocsPerRun(10, func() { _ = g.AsBinary() })
		if allocs != 1 {
			t.Errorf("AsBinary allocation count: geometry %d allocs=%v, want 1", i, allocs)
		}
	}
}
