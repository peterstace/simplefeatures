package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestDumpCoordinatesPoint(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        geom.Sequence
	}{
		{
			description: "empty",
			inputWKT:    "POINT EMPTY",
			want:        geom.NewSequence(nil, geom.DimXY),
		},
		{
			description: "empty z",
			inputWKT:    "POINT Z EMPTY",
			want:        geom.NewSequence(nil, geom.DimXYZ),
		},
		{
			description: "empty m",
			inputWKT:    "POINT M EMPTY",
			want:        geom.NewSequence(nil, geom.DimXYM),
		},
		{
			description: "empty zm",
			inputWKT:    "POINT ZM EMPTY",
			want:        geom.NewSequence(nil, geom.DimXYZM),
		},
		{
			description: "non-empty",
			inputWKT:    "POINT(1 2)",
			want:        geom.NewSequence([]float64{1, 2}, geom.DimXY),
		},
		{
			description: "non-empty z",
			inputWKT:    "POINT Z(1 2 3)",
			want:        geom.NewSequence([]float64{1, 2, 3}, geom.DimXYZ),
		},
		{
			description: "non-empty m",
			inputWKT:    "POINT M(1 2 3)",
			want:        geom.NewSequence([]float64{1, 2, 3}, geom.DimXYM),
		},
		{
			description: "non-empty zm",
			inputWKT:    "POINT ZM(1 2 3 4)",
			want:        geom.NewSequence([]float64{1, 2, 3, 4}, geom.DimXYZM),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).MustAsPoint().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

func TestDumpCoordinatesMultiLineString(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        geom.Sequence
	}{
		{
			description: "empty",
			inputWKT:    "MULTILINESTRING EMPTY",
			want:        geom.NewSequence(nil, geom.DimXY),
		},
		{
			description: "contains empty LineString",
			inputWKT:    "MULTILINESTRING(EMPTY)",
			want:        geom.NewSequence(nil, geom.DimXY),
		},
		{
			description: "single non-empty LineString",
			inputWKT:    "MULTILINESTRING((1 2,3 4))",
			want:        geom.NewSequence([]float64{1, 2, 3, 4}, geom.DimXY),
		},
		{
			description: "multiple non-empty LineStrings",
			inputWKT:    "MULTILINESTRING((1 2,3 4),(5 6,7 8))",
			want:        geom.NewSequence([]float64{1, 2, 3, 4, 5, 6, 7, 8}, geom.DimXY),
		},
		{
			description: "mix of empty and non-empty LineStrings",
			inputWKT:    "MULTILINESTRING(EMPTY,(1 2,3 4))",
			want:        geom.NewSequence([]float64{1, 2, 3, 4}, geom.DimXY),
		},
		{
			description: "Z coordinates",
			inputWKT:    "MULTILINESTRING Z((1 2 3,3 4 5))",
			want:        geom.NewSequence([]float64{1, 2, 3, 3, 4, 5}, geom.DimXYZ),
		},
		{
			description: "M coordinates",
			inputWKT:    "MULTILINESTRING M((1 2 3,3 4 5))",
			want:        geom.NewSequence([]float64{1, 2, 3, 3, 4, 5}, geom.DimXYM),
		},
		{
			description: "ZM coordinates",
			inputWKT:    "MULTILINESTRING ZM((1 2 3 4,3 4 5 6))",
			want:        geom.NewSequence([]float64{1, 2, 3, 4, 3, 4, 5, 6}, geom.DimXYZM),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).MustAsMultiLineString().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

func TestDumpCoordinatesPolygon(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        geom.Sequence
	}{
		{
			description: "empty",
			inputWKT:    "POLYGON EMPTY",
			want:        geom.NewSequence(nil, geom.DimXY),
		},
		{
			description: "contains single ring",
			inputWKT:    "POLYGON((0 0,0 1,1 0,0 0))",
			want:        geom.NewSequence([]float64{0, 0, 0, 1, 1, 0, 0, 0}, geom.DimXY),
		},
		{
			description: "multiple rings",
			inputWKT:    "POLYGON((0 0,0 10,10 0,0 0),(1 1,1 2,2 2,2 1,1 1))",
			want:        geom.NewSequence([]float64{0, 0, 0, 10, 10, 0, 0, 0, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1}, geom.DimXY),
		},
		{
			description: "Z coordinates",
			inputWKT:    "POLYGON Z((0 0 -1,0 10 -1,10 0 -1,0 0 -1),(1 1 -1,1 2 -1,2 2 -1,2 1 -1,1 1 -1))",
			want: geom.NewSequence([]float64{
				0, 0, -1,
				0, 10, -1,
				10, 0, -1,
				0, 0, -1,
				1, 1, -1,
				1, 2, -1,
				2, 2, -1,
				2, 1, -1,
				1, 1, -1,
			}, geom.DimXYZ),
		},
		{
			description: "M coordinates",
			inputWKT:    "POLYGON M((0 0 10,0 1 10,1 0 10,0 0 10))",
			want:        geom.NewSequence([]float64{0, 0, 10, 0, 1, 10, 1, 0, 10, 0, 0, 10}, geom.DimXYM),
		},
		{
			description: "ZM coordinates",
			inputWKT:    "POLYGON ZM((0 0 10 20,0 1 10 20,1 0 10 20,0 0 10 20))",
			want:        geom.NewSequence([]float64{0, 0, 10, 20, 0, 1, 10, 20, 1, 0, 10, 20, 0, 0, 10, 20}, geom.DimXYZM),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).MustAsPolygon().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

func TestDumpCoordinatesMultiPolygon(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        geom.Sequence
	}{
		{
			description: "empty",
			inputWKT:    "MULTIPOLYGON EMPTY",
			want:        geom.NewSequence(nil, geom.DimXY),
		},
		{
			description: "multi polygon with empty polygon",
			inputWKT:    "MULTIPOLYGON(EMPTY)",
			want:        geom.NewSequence(nil, geom.DimXY),
		},
		{
			description: "contains single ring",
			inputWKT:    "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))",
			want:        geom.NewSequence([]float64{0, 0, 0, 1, 1, 0, 0, 0}, geom.DimXY),
		},
		{
			description: "multiple rings in a single polygon",
			inputWKT:    "MULTIPOLYGON(((0 0,0 10,10 0,0 0),(1 1,1 2,2 2,2 1,1 1)))",
			want:        geom.NewSequence([]float64{0, 0, 0, 10, 10, 0, 0, 0, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1}, geom.DimXY),
		},
		{
			description: "multiple polygons",
			inputWKT:    "MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((10 10,10 11,11 10,10 10)))",
			want:        geom.NewSequence([]float64{0, 0, 0, 1, 1, 0, 0, 0, 10, 10, 10, 11, 11, 10, 10, 10}, geom.DimXY),
		},
		{
			description: "Z coordinates",
			inputWKT:    "MULTIPOLYGON Z(((0 0 10,0 1 10,1 0 10,0 0 10)))",
			want:        geom.NewSequence([]float64{0, 0, 10, 0, 1, 10, 1, 0, 10, 0, 0, 10}, geom.DimXYZ),
		},
		{
			description: "M coordinates",
			inputWKT:    "MULTIPOLYGON M(((0 0 10,0 1 10,1 0 10,0 0 10)))",
			want:        geom.NewSequence([]float64{0, 0, 10, 0, 1, 10, 1, 0, 10, 0, 0, 10}, geom.DimXYM),
		},
		{
			description: "ZM coordinates",
			inputWKT:    "MULTIPOLYGON ZM(((0 0 20 10,0 1 20 10,1 0 20 10,0 0 20 10)))",
			want:        geom.NewSequence([]float64{0, 0, 20, 10, 0, 1, 20, 10, 1, 0, 20, 10, 0, 0, 20, 10}, geom.DimXYZM),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).MustAsMultiPolygon().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

func TestDumpCoordinatesGeometryCollection(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        geom.Sequence
	}{
		{
			description: "empty",
			inputWKT:    "GEOMETRYCOLLECTION EMPTY",
			want:        geom.NewSequence(nil, geom.DimXY),
		},
		{
			description: "empty z",
			inputWKT:    "GEOMETRYCOLLECTION Z EMPTY",
			want:        geom.NewSequence(nil, geom.DimXYZ),
		},
		{
			description: "single point",
			inputWKT:    "GEOMETRYCOLLECTION(POINT(1 2))",
			want:        geom.NewSequence([]float64{1, 2}, geom.DimXY),
		},
		{
			description: "single point z",
			inputWKT:    "GEOMETRYCOLLECTION Z(POINT Z(1 2 0))",
			want:        geom.NewSequence([]float64{1, 2, 0}, geom.DimXYZ),
		},
		{
			description: "nested",
			inputWKT:    "GEOMETRYCOLLECTION Z(GEOMETRYCOLLECTION Z(POINT Z(1 2 0)))",
			want:        geom.NewSequence([]float64{1, 2, 0}, geom.DimXYZ),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).MustAsGeometryCollection().DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}

func TestDumpCoordinatesGeometry(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputWKT    string
		want        geom.Sequence
	}{
		{
			description: "Point",
			inputWKT:    "POINT Z(0 1 2)",
			want:        geom.NewSequence([]float64{0, 1, 2}, geom.DimXYZ),
		},
		{
			description: "LineString",
			inputWKT:    "LINESTRING Z(0 1 2,3 4 5)",
			want:        geom.NewSequence([]float64{0, 1, 2, 3, 4, 5}, geom.DimXYZ),
		},
		{
			description: "Polygon",
			inputWKT:    "POLYGON Z((0 0 1,0 1 1,1 0 1,0 0 1))",
			want:        geom.NewSequence([]float64{0, 0, 1, 0, 1, 1, 1, 0, 1, 0, 0, 1}, geom.DimXYZ),
		},
		{
			description: "MultiPoint",
			inputWKT:    "MULTIPOINT Z(0 1 2,3 4 5)",
			want:        geom.NewSequence([]float64{0, 1, 2, 3, 4, 5}, geom.DimXYZ),
		},
		{
			description: "MultiLineString",
			inputWKT:    "MULTILINESTRING Z((0 1 2,3 4 5))",
			want:        geom.NewSequence([]float64{0, 1, 2, 3, 4, 5}, geom.DimXYZ),
		},
		{
			description: "MultiPolygon",
			inputWKT:    "MULTIPOLYGON Z(((0 0 1,0 1 1,1 0 1,0 0 1)))",
			want:        geom.NewSequence([]float64{0, 0, 1, 0, 1, 1, 1, 0, 1, 0, 0, 1}, geom.DimXYZ),
		},
		{
			description: "GeometryCollection",
			inputWKT:    "GEOMETRYCOLLECTION Z(POINT Z(0 1 2))",
			want:        geom.NewSequence([]float64{0, 1, 2}, geom.DimXYZ),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).DumpCoordinates()
			expectSequenceEq(t, got, tc.want)
		})
	}
}
