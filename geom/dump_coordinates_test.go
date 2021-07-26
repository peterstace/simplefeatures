package geom_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestDumpCoordinatesMultiPoint(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        []Coordinates
	}{
		{
			description: "empty",
			inputWKT:    "MULTIPOINT EMPTY",
			want:        nil,
		},
		{
			description: "contains empty point",
			inputWKT:    "MULTIPOINT(EMPTY)",
			want:        nil,
		},
		{
			description: "single non-empty point",
			inputWKT:    "MULTIPOINT(1 2)",
			want: []Coordinates{
				NewXYCoordinates(1, 2),
			},
		},
		{
			description: "multiple non-empty points",
			inputWKT:    "MULTIPOINT(1 2,3 4)",
			want: []Coordinates{
				NewXYCoordinates(1, 2),
				NewXYCoordinates(3, 4),
			},
		},
		{
			description: "mix of empty and non-empty points",
			inputWKT:    "MULTIPOINT(EMPTY,3 4)",
			want: []Coordinates{
				NewXYCoordinates(3, 4),
			},
		},
		{
			description: "Z coordinates",
			inputWKT:    "MULTIPOINT Z(3 4 5)",
			want: []Coordinates{
				NewXYZCoordinates(3, 4, 5),
			},
		},
		{
			description: "M coordinates",
			inputWKT:    "MULTIPOINT M(3 4 6)",
			want: []Coordinates{
				NewXYMCoordinates(3, 4, 6),
			},
		},
		{
			description: "ZM coordinates",
			inputWKT:    "MULTIPOINT ZM(3 4 5 6)",
			want: []Coordinates{
				NewXYZMCoordinates(3, 4, 5, 6),
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			mp := geomFromWKT(t, tc.inputWKT).AsMultiPoint()
			got := mp.DumpCoordinates()
			expectCoordinateSliceEq(t, got, tc.want)
		})
	}
}
