package geom_test

import (
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestCoordinatesString(t *testing.T) {
	for i, tc := range []struct {
		coords Coordinates
		want   string
	}{
		{
			Coordinates{},
			"Coordinates[XY] 0 0",
		},
		{
			Coordinates{XY: XY{X: 1, Y: 2}},
			"Coordinates[XY] 1 2",
		},
		{
			Coordinates{XY: XY{X: 1, Y: 2}, Z: 3, Type: DimXYZ},
			"Coordinates[XYZ] 1 2 3",
		},
		{
			Coordinates{XY: XY{X: 1, Y: 2}, M: 3, Type: DimXYM},
			"Coordinates[XYM] 1 2 3",
		},
		{
			Coordinates{XY: XY{X: 1, Y: 2}, Z: 3, M: 4, Type: DimXYZM},
			"Coordinates[XYZM] 1 2 3 4",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tc.coords.String()
			expectStringEq(t, got, tc.want)
		})
	}
}
