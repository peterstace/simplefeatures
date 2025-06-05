package geom

import (
	"testing"
)

func TestGeometryCollectionFlipCoordinates(t *testing.T) {
	pt := NewPointXY(1, 2)
	ls := NewLineStringXY(3, 4, 5, 6)
	gc := NewGeometryCollection([]Geometry{pt.AsGeometry(), ls.AsGeometry()})
	flipped := gc.FlipCoordinates()
	geoms := flipped.Dump()
	g0 := geoms[0].MustAsPoint()
	g1 := geoms[1].MustAsLineString()
	xy0, _ := g0.XY()
	seq := g1.Coordinates()
	xy1 := seq.GetXY(0)
	xy2 := seq.GetXY(1)
	if xy0 != (XY{2, 1}) || xy1 != (XY{4, 3}) || xy2 != (XY{6, 5}) {
		t.Errorf("expected [(2,1),(4,3),(6,5)], got [%v,%v,%v]", xy0, xy1, xy2)
	}
}
