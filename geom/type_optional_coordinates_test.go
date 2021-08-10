package geom_test

import (
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestOptionalCoordinatesString(t *testing.T) {
	for i, tc := range []struct {
		input OptionalCoordinates
		want  string
	}{
		{
			OptionalCoordinates{},
			"OptionalCoordinates[XY] 0 0",
		},
		{
			OptionalCoordinates{Type: DimXY, XY: XY{X: 1, Y: 2}},
			"OptionalCoordinates[XY] 1 2",
		},
		{
			OptionalCoordinates{Type: DimXYZ, XY: XY{X: 1, Y: 2}, Z: 3},
			"OptionalCoordinates[XYZ] 1 2 3",
		},
		{
			OptionalCoordinates{Type: DimXYM, XY: XY{X: 1, Y: 2}, M: 3},
			"OptionalCoordinates[XYM] 1 2 3",
		},
		{
			OptionalCoordinates{Type: DimXYZM, XY: XY{X: 1, Y: 2}, Z: 3, M: 4},
			"OptionalCoordinates[XYZM] 1 2 3 4",
		},
		{
			OptionalCoordinates{Type: DimXY, Empty: true},
			"OptionalCoordinates[XY] EMPTY",
		},
		{
			OptionalCoordinates{Type: DimXYZ, Empty: true},
			"OptionalCoordinates[XYZ] EMPTY",
		},
		{
			OptionalCoordinates{Type: DimXYM, Empty: true},
			"OptionalCoordinates[XYM] EMPTY",
		},
		{
			OptionalCoordinates{Type: DimXYZM, Empty: true},
			"OptionalCoordinates[XYZM] EMPTY",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tc.input.String()
			expectStringEq(t, got, tc.want)
		})
	}
}
