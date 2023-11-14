package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestCoordinatesString(t *testing.T) {
	for i, tc := range []struct {
		coords geom.Coordinates
		want   string
	}{
		{
			geom.Coordinates{},
			"Coordinates[XY] 0 0",
		},
		{
			geom.Coordinates{XY: geom.XY{X: 1, Y: 2}},
			"Coordinates[XY] 1 2",
		},
		{
			geom.Coordinates{XY: geom.XY{X: 1, Y: 2}, Z: 3, Type: geom.DimXYZ},
			"Coordinates[XYZ] 1 2 3",
		},
		{
			geom.Coordinates{XY: geom.XY{X: 1, Y: 2}, M: 3, Type: geom.DimXYM},
			"Coordinates[XYM] 1 2 3",
		},
		{
			geom.Coordinates{XY: geom.XY{X: 1, Y: 2}, Z: 3, M: 4, Type: geom.DimXYZM},
			"Coordinates[XYZM] 1 2 3 4",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tc.coords.String()
			expectStringEq(t, got, tc.want)
		})
	}
}
