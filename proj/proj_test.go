package proj_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/proj"
)

func TestBasic(t *testing.T) {
	pj, err := proj.NewTransformation("EPSG:4326", "+proj=utm +zone=32 +datum=WGS84")
	if err != nil {
		t.Fatal(err)
	}
	defer pj.Release()

	coords := []float64{55, 12} // lon, lat
	t.Logf("coords: %v", coords)

	if err := pj.Forward(geom.DimXY, coords); err != nil {
		t.Fatal(err)
	}
	t.Logf("coords: %v", coords)

	if err := pj.Inverse(geom.DimXY, coords); err != nil {
		t.Fatal(err)
	}
	t.Logf("coords: %v", coords)
}
