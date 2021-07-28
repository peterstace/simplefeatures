package geom_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestDumpCoordinatesMultiPoint(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        Sequence
	}{
		{
			description: "empty",
			inputWKT:    "MULTIPOINT EMPTY",
			want:        NewSequence(nil, DimXY),
		},
		{
			description: "contains empty point",
			inputWKT:    "MULTIPOINT(EMPTY)",
			want:        NewSequence(nil, DimXY),
		},
		{
			description: "single non-empty point",
			inputWKT:    "MULTIPOINT(1 2)",
			want:        NewSequence([]float64{1, 2}, DimXY),
		},
		{
			description: "multiple non-empty points",
			inputWKT:    "MULTIPOINT(1 2,3 4,5 6)",
			want:        NewSequence([]float64{1, 2, 3, 4, 5, 6}, DimXY),
		},
		{
			description: "mix of empty and non-empty points",
			inputWKT:    "MULTIPOINT(EMPTY,3 4)",
			want:        NewSequence([]float64{3, 4}, DimXY),
		},
		{
			description: "Z coordinates",
			inputWKT:    "MULTIPOINT Z(3 4 5)",
			want:        NewSequence([]float64{3, 4, 5}, DimXYZ),
		},
		{
			description: "M coordinates",
			inputWKT:    "MULTIPOINT M(3 4 6)",
			want:        NewSequence([]float64{3, 4, 6}, DimXYM),
		},
		{
			description: "ZM coordinates",
			inputWKT:    "MULTIPOINT ZM(3 4 5 6)",
			want:        NewSequence([]float64{3, 4, 5, 6}, DimXYZM),
		},
		{
			description: "reproduce bug",
			inputWKT:    "MULTIPOINT Z(3 4 5,6 7 8)",
			want:        NewSequence([]float64{3, 4, 5, 6, 7, 8}, DimXYZ),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).AsMultiPoint().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

func TestDumpCoordinatesMultiLineString(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        Sequence
	}{
		{
			description: "empty",
			inputWKT:    "MULTILINESTRING EMPTY",
			want:        NewSequence(nil, DimXY),
		},
		{
			description: "contains empty LineString",
			inputWKT:    "MULTILINESTRING(EMPTY)",
			want:        NewSequence(nil, DimXY),
		},
		{
			description: "single non-empty LineString",
			inputWKT:    "MULTILINESTRING((1 2,3 4))",
			want:        NewSequence([]float64{1, 2, 3, 4}, DimXY),
		},
		{
			description: "multiple non-empty LineStrings",
			inputWKT:    "MULTILINESTRING((1 2,3 4),(5 6,7 8))",
			want:        NewSequence([]float64{1, 2, 3, 4, 5, 6, 7, 8}, DimXY),
		},
		{
			description: "mix of empty and non-empty LineStrings",
			inputWKT:    "MULTILINESTRING(EMPTY,(1 2,3 4))",
			want:        NewSequence([]float64{1, 2, 3, 4}, DimXY),
		},
		{
			description: "Z coordinates",
			inputWKT:    "MULTILINESTRING Z((1 2 3,3 4 5))",
			want:        NewSequence([]float64{1, 2, 3, 3, 4, 5}, DimXYZ),
		},
		{
			description: "M coordinates",
			inputWKT:    "MULTILINESTRING M((1 2 3,3 4 5))",
			want:        NewSequence([]float64{1, 2, 3, 3, 4, 5}, DimXYM),
		},
		{
			description: "ZM coordinates",
			inputWKT:    "MULTILINESTRING ZM((1 2 3 4,3 4 5 6))",
			want:        NewSequence([]float64{1, 2, 3, 4, 3, 4, 5, 6}, DimXYZM),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).AsMultiLineString().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}
