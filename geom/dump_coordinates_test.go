package geom_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestDumpCoordinatesPoint(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        Sequence
	}{
		{
			description: "empty",
			inputWKT:    "POINT EMPTY",
			want:        NewSequence(nil, DimXY),
		},
		{
			description: "empty z",
			inputWKT:    "POINT Z EMPTY",
			want:        NewSequence(nil, DimXYZ),
		},
		{
			description: "empty m",
			inputWKT:    "POINT M EMPTY",
			want:        NewSequence(nil, DimXYM),
		},
		{
			description: "empty zm",
			inputWKT:    "POINT ZM EMPTY",
			want:        NewSequence(nil, DimXYZM),
		},
		{
			description: "non-empty",
			inputWKT:    "POINT(1 2)",
			want:        NewSequence([]float64{1, 2}, DimXY),
		},
		{
			description: "non-empty z",
			inputWKT:    "POINT Z(1 2 3)",
			want:        NewSequence([]float64{1, 2, 3}, DimXYZ),
		},
		{
			description: "non-empty m",
			inputWKT:    "POINT M(1 2 3)",
			want:        NewSequence([]float64{1, 2, 3}, DimXYM),
		},
		{
			description: "non-empty zm",
			inputWKT:    "POINT ZM(1 2 3 4)",
			want:        NewSequence([]float64{1, 2, 3, 4}, DimXYZM),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).AsPoint().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

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

func TestDumpCoordinatesPolygon(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        Sequence
	}{
		{
			description: "empty",
			inputWKT:    "POLYGON EMPTY",
			want:        NewSequence(nil, DimXY),
		},
		{
			description: "contains single ring",
			inputWKT:    "POLYGON((0 0,0 1,1 0,0 0))",
			want:        NewSequence([]float64{0, 0, 0, 1, 1, 0, 0, 0}, DimXY),
		},
		{
			description: "multiple rings",
			inputWKT:    "POLYGON((0 0,0 10,10 0,0 0),(1 1,1 2,2 2,2 1,1 1))",
			want:        NewSequence([]float64{0, 0, 0, 10, 10, 0, 0, 0, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1}, DimXY),
		},
		{
			description: "Z coordinates",
			inputWKT:    "POLYGON Z((0 0 -1,0 10 -1,10 0 -1,0 0 -1),(1 1 -1,1 2 -1,2 2 -1,2 1 -1,1 1 -1))",
			want: NewSequence([]float64{
				0, 0, -1,
				0, 10, -1,
				10, 0, -1,
				0, 0, -1,
				1, 1, -1,
				1, 2, -1,
				2, 2, -1,
				2, 1, -1,
				1, 1, -1,
			}, DimXYZ),
		},
		{
			description: "M coordinates",
			inputWKT:    "POLYGON M((0 0 10,0 1 10,1 0 10,0 0 10))",
			want:        NewSequence([]float64{0, 0, 10, 0, 1, 10, 1, 0, 10, 0, 0, 10}, DimXYM),
		},
		{
			description: "ZM coordinates",
			inputWKT:    "POLYGON ZM((0 0 10 20,0 1 10 20,1 0 10 20,0 0 10 20))",
			want:        NewSequence([]float64{0, 0, 10, 20, 0, 1, 10, 20, 1, 0, 10, 20, 0, 0, 10, 20}, DimXYZM),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).AsPolygon().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

func TestDumpCoordinatesMultiPolygon(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        Sequence
	}{
		{
			description: "empty",
			inputWKT:    "MULTIPOLYGON EMPTY",
			want:        NewSequence(nil, DimXY),
		},
		{
			description: "multi polygon with empty polygon",
			inputWKT:    "MULTIPOLYGON(EMPTY)",
			want:        NewSequence(nil, DimXY),
		},
		{
			description: "contains single ring",
			inputWKT:    "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))",
			want:        NewSequence([]float64{0, 0, 0, 1, 1, 0, 0, 0}, DimXY),
		},
		{
			description: "multiple rings in a single polygon",
			inputWKT:    "MULTIPOLYGON(((0 0,0 10,10 0,0 0),(1 1,1 2,2 2,2 1,1 1)))",
			want:        NewSequence([]float64{0, 0, 0, 10, 10, 0, 0, 0, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1}, DimXY),
		},
		{
			description: "multiple polygons",
			inputWKT:    "MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((10 10,10 11,11 10,10 10)))",
			want:        NewSequence([]float64{0, 0, 0, 1, 1, 0, 0, 0, 10, 10, 10, 11, 11, 10, 10, 10}, DimXY),
		},
		{
			description: "Z coordinates",
			inputWKT:    "MULTIPOLYGON Z(((0 0 10,0 1 10,1 0 10,0 0 10)))",
			want:        NewSequence([]float64{0, 0, 10, 0, 1, 10, 1, 0, 10, 0, 0, 10}, DimXYZ),
		},
		{
			description: "M coordinates",
			inputWKT:    "MULTIPOLYGON M(((0 0 10,0 1 10,1 0 10,0 0 10)))",
			want:        NewSequence([]float64{0, 0, 10, 0, 1, 10, 1, 0, 10, 0, 0, 10}, DimXYM),
		},
		{
			description: "ZM coordinates",
			inputWKT:    "MULTIPOLYGON ZM(((0 0 20 10,0 1 20 10,1 0 20 10,0 0 20 10)))",
			want:        NewSequence([]float64{0, 0, 20, 10, 0, 1, 20, 10, 1, 0, 20, 10, 0, 0, 20, 10}, DimXYZM),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).AsMultiPolygon().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

func TestDumpCoordinatesGeometry(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        Sequence
	}{
		{
			description: "Point",
			inputWKT:    "POINT Z(0 1 2)",
			want:        NewSequence([]float64{0, 1, 2}, DimXYZ),
		},
		{
			description: "LineString",
			inputWKT:    "LINESTRING Z(0 1 2,3 4 5)",
			want:        NewSequence([]float64{0, 1, 2, 3, 4, 5}, DimXYZ),
		},
		{
			description: "Polygon",
			inputWKT:    "POLYGON Z((0 0 1,0 1 1,1 0 1,0 0 1))",
			want:        NewSequence([]float64{0, 0, 1, 0, 1, 1, 1, 0, 1, 0, 0, 1}, DimXYZ),
		},
		{
			description: "MultiPoint",
			inputWKT:    "MULTIPOINT Z(0 1 2,3 4 5)",
			want:        NewSequence([]float64{0, 1, 2, 3, 4, 5}, DimXYZ),
		},
		{
			description: "MultiLineString",
			inputWKT:    "MULTILINESTRING Z((0 1 2,3 4 5))",
			want:        NewSequence([]float64{0, 1, 2, 3, 4, 5}, DimXYZ),
		},
		{
			description: "MultiPolygon",
			inputWKT:    "MULTIPOLYGON Z(((0 0 1,0 1 1,1 0 1,0 0 1)))",
			want:        NewSequence([]float64{0, 0, 1, 0, 1, 1, 1, 0, 1, 0, 0, 1}, DimXYZ),
		},
		//{
		//	description: "GeometryCollection",
		//	inputWKT:    "GEOMETRYCOLLECTION Z(POINT Z(0 1 2))",
		//	want:        NewSequence([]float64{0, 1, 2}, DimXYZ),
		//},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}
