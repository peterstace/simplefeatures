package geom_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestGeoJSONUnmarshalValid(t *testing.T) {
	// Test data from the following query:
	/*
		SELECT
			ST_AsText(ST_GeomFromText(wkt)),
			ST_AsGeoJSON(ST_GeomFromText(wkt)) AS geojson
		FROM (
			VALUES
			('POINT EMPTY'),
			('POINT(1 2)'),
			('POINTZ(1 2 3)'),
			('LINESTRING EMPTY'),
			('LINESTRING(1 2,3 4)'),
			('LINESTRINGZ(1 2 3,4 5 6)'),
			('LINESTRING(1 2,3 4,5 6)'),
			('LINESTRINGZ(1 2 3,3 4 5,5 6 7)'),
			('POLYGON EMPTY'),
			('POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))'),
			('POLYGONZ((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))'),
			('MULTIPOINT EMPTY'),
			('MULTIPOINT(1 2)'),
			('MULTIPOINTZ(1 2 3)'),
			('MULTIPOINT(1 2,3 4)'),
			('MULTIPOINTZ(1 2 3,3 4 5)'),
			('MULTILINESTRING EMPTY'),
			('MULTILINESTRINGZ EMPTY'),
			('MULTILINESTRING((0 1,2 3,4 5))'),
			('MULTILINESTRINGZ((0 1 8,2 3 8,4 5 8))'),
			('MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))'),
			('MULTILINESTRINGZ((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))'),
			('MULTIPOLYGON EMPTY'),
			('MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))'),
			('MULTIPOLYGONZ(((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))'),
			('MULTIPOLYGON(EMPTY,((1 0,2 0,1 1,1 0)))'),
			('MULTIPOLYGONZ(EMPTY,((1 0 9,2 0 9,1 1 9,1 0 9)))'),
			('GEOMETRYCOLLECTION EMPTY'),
			('GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))'),
			('GEOMETRYCOLLECTIONZ(POINTZ(1 2 3),POINTZ(3 4 5))')
		) AS q (wkt);
	*/
	for i, tt := range []struct {
		geojson string
		wkt     string
	}{
		{
			geojson: `{"type":"Point","coordinates":[]}`,
			wkt:     "POINT EMPTY",
		},
		{
			geojson: `{"type":"Point","coordinates":[1,2]}`,
			wkt:     "POINT(1 2)",
		},
		{
			geojson: `{"type":"Point","coordinates":[1,2,3]}`,
			wkt:     "POINT Z (1 2 3)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[]}`,
			wkt:     "LINESTRING EMPTY",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2],[3,4]]}`,
			wkt:     "LINESTRING(1 2,3 4)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3],[4,5,6]]}`,
			wkt:     "LINESTRING Z (1 2 3,4 5 6)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2],[3,4],[5,6]]}`,
			wkt:     "LINESTRING(1 2,3 4,5 6)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3],[3,4,5],[5,6,7]]}`,
			wkt:     "LINESTRING Z (1 2 3,3 4 5,5 6 7)",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[]}`,
			wkt:     "POLYGON EMPTY",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[[[0,0],[4,0],[0,4],[0,0]],[[1,1],[2,1],[1,2],[1,1]]]}`,
			wkt:     "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[[[0,0,9],[4,0,9],[0,4,9],[0,0,9]],[[1,1,9],[2,1,9],[1,2,9],[1,1,9]]]}`,
			wkt:     "POLYGON Z ((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[]}`,
			wkt:     "MULTIPOINT EMPTY",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
			wkt:     "MULTIPOINT(1 2)",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3]]}`,
			wkt:     "MULTIPOINT Z (1 2 3)",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`,
			wkt:     "MULTIPOINT(1 2,3 4)",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3],[3,4,5]]}`,
			wkt:     "MULTIPOINT Z (1 2 3,3 4 5)",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[]}`,
			wkt:     "MULTILINESTRING EMPTY",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3],[4,5]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3,4 5))",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,8],[2,3,8],[4,5,8]]]}`,
			wkt:     "MULTILINESTRING Z ((0 1 8,2 3 8,4 5 8))",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[[4,5],[6,7],[8,9]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,9],[2,3,9]],[[4,5,9],[6,7,9],[8,9,9]]]}`,
			wkt:     "MULTILINESTRING Z ((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[]}`,
			wkt:     "MULTIPOLYGON EMPTY",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
			wkt:     "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0,9],[1,0,9],[0,1,9],[0,0,9]]],[[[1,0,9],[2,0,9],[1,1,9],[1,0,9]]]]}`,
			wkt:     "MULTIPOLYGON Z (((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
			wkt:     "MULTIPOLYGON(EMPTY,((1 0,2 0,1 1,1 0)))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[],[[[1,0,9],[2,0,9],[1,1,9],[1,0,9]]]]}`,
			wkt:     "MULTIPOLYGON Z (EMPTY,((1 0 9,2 0 9,1 1 9,1 0 9)))",
		},
		{
			geojson: `{"type":"GeometryCollection","geometries":[]}`,
			wkt:     "GEOMETRYCOLLECTION EMPTY",
		},
		{
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]}`,
			wkt:     "GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))",
		},
		{
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2,3]},{"type":"Point","coordinates":[3,4,5]}]}`,
			wkt:     "GEOMETRYCOLLECTION Z (POINT Z (1 2 3),POINT Z (3 4 5))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := geom.UnmarshalGeoJSON([]byte(tt.geojson))
			expectNoErr(t, err)
			want := geomFromWKT(t, tt.wkt)
			expectGeomEq(t, got, want)
		})
	}
}

func TestGeoJSONUnmarshalValidXYZM(t *testing.T) {
	// The GeoJSON spec states:
	//
	// > The interpretation and meaning of additional elements is beyond the
	// > scope of this specification, and additional elements MAY be ignored by
	// > parsers.
	//
	// However, some users find that being able to unmarshal XYZM coordinates
	// is useful, and isn't strictly prohibited by the spec.

	for i, tt := range []struct {
		geojson string
		wkt     string
	}{
		{
			geojson: `{"type":"Point","coordinates":[1,2,3,4]}`,
			wkt:     "POINT ZM (1 2 3 4)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3,1],[3,4,5,2],[5,6,7,3]]}`,
			wkt:     "LINESTRING ZM (1 2 3 1,3 4 5 2,5 6 7 3)",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[[[0,0,9,1],[4,0,9,2],[0,4,9,3],[0,0,9,4]],[[1,1,9,5],[2,1,9,6],[1,2,9,7],[1,1,9,8]]]}`,
			wkt:     "POLYGON ZM ((0 0 9 1,4 0 9 2,0 4 9 3,0 0 9 4),(1 1 9 5,2 1 9 6,1 2 9 7,1 1 9 8))",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3,1],[3,4,5,2]]}`,
			wkt:     "MULTIPOINT ZM (1 2 3 1,3 4 5 2)",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,9,1],[2,3,9,2]],[[4,5,9,3],[6,7,9,4],[8,9,9,5]]]}`,
			wkt:     "MULTILINESTRING ZM ((0 1 9 1,2 3 9 2),(4 5 9 3,6 7 9 4,8 9 9 5))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0,9,1],[1,0,9,2],[0,1,9,3],[0,0,9,4]]],[[[1,0,9,5],[2,0,9,6],[1,1,9,7],[1,0,9,8]]]]}`,
			wkt:     "MULTIPOLYGON ZM (((0 0 9 1,1 0 9 2,0 1 9 3,0 0 9 4)),((1 0 9 5,2 0 9 6,1 1 9 7,1 0 9 8)))",
		},
		{
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2,3,1]},{"type":"Point","coordinates":[3,4,5,2]}]}`,
			wkt:     "GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 1),POINT ZM (3 4 5 2))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := geom.UnmarshalGeoJSON([]byte(tt.geojson))
			expectNoErr(t, err)
			want := geomFromWKT(t, tt.wkt)
			expectGeomEq(t, got, want)
		})
	}
}

func TestGeoJSONUnmarshalValidAllowAdditionalCoordDimensions(t *testing.T) {
	for i, tt := range []struct {
		geojson string
		wkt     string
	}{
		{
			geojson: `{"type":"Point","coordinates":[1,2,3,4,5]}`,
			wkt:     "POINT ZM (1 2 3 4)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3,4,5],[2,3,4,5,6]]}`,
			wkt:     "LINESTRING ZM (1 2 3 4,2 3 4 5)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3,4,5],[2,3,4,5,6],[3,4,5,6,7,8]]}`,
			wkt:     "LINESTRING ZM (1 2 3 4,2 3 4 5,3 4 5 6)",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[[[0,0,0,0,6],[0,1,0,0,7],[1,0,0,0,8],[0,0,0,0,9]]]}`,
			wkt:     "POLYGON ZM ((0 0 0 0,0 1 0 0,1 0 0 0,0 0 0 0))",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3,4,5]]}`,
			wkt:     "MULTIPOINT ZM (1 2 3 4)",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[1,2,3,4,5],[2,3,4,5,6]]]}`,
			wkt:     "MULTILINESTRING ZM ((1 2 3 4,2 3 4 5))",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[1,2,3,4,5],[2,3,4,5,6],[3,4,5,6,7,8]]]}`,
			wkt:     "MULTILINESTRING ZM ((1 2 3 4,2 3 4 5,3 4 5 6))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0,0,0,1],[0,1,0,0,2],[1,0,0,0,3],[0,0,0,0,4]]]]}`,
			wkt:     "MULTIPOLYGON ZM (((0 0 0 0,0 1 0 0,1 0 0 0,0 0 0 0)))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := geom.UnmarshalGeoJSON([]byte(tt.geojson))
			expectNoErr(t, err)
			want := geomFromWKT(t, tt.wkt)
			expectGeomEq(t, got, want)
		})
	}
}

func TestGeoJSONSyntaxError(t *testing.T) {
	for _, tc := range []struct {
		description string
		geojson     string
		errorText   string
	}{
		{
			"invalid json",
			`{zort}`,
			"invalid character 'z' looking for beginning of object key string",
		},
		{
			"unknown geometry type",
			`{"type":"foo","coordinates":[[0,0],[1,1]]}`,
			"unknown geometry type: 'foo'",
		},

		{
			"bad coordinate length - point 1",
			`{"type":"Point","coordinates":[0]}`,
			"invalid geojson coordinate length: 1",
		},
		{
			"bad coordinate length - linestring 0",
			`{"type":"LineString","coordinates":[[],[0]]}`,
			"invalid geojson coordinate length: 0",
		},
		{
			"bad coordinate length - linestring 1",
			`{"type":"LineString","coordinates":[[0],[0]]}`,
			"invalid geojson coordinate length: 1",
		},
		{
			"bad coordinate length - polygon 0",
			`{"type":"Polygon","coordinates":[[[],[0]]]}`,
			"invalid geojson coordinate length: 0",
		},
		{
			"bad coordinate length - polygon 1",
			`{"type":"Polygon","coordinates":[[[0],[0]]]}`,
			"invalid geojson coordinate length: 1",
		},
		{
			"bad coordinate length - multipoint 0",
			`{"type":"MultiPoint","coordinates":[[],[0]]}`,
			"invalid geojson coordinate length: 0",
		},
		{
			"bad coordinate length - multipoint 1",
			`{"type":"MultiPoint","coordinates":[[0],[0]]}`,
			"invalid geojson coordinate length: 1",
		},
		{
			"bad coordinate length - multilinestring 0",
			`{"type":"MultiLineString","coordinates":[[[],[0]]]}`,
			"invalid geojson coordinate length: 0",
		},
		{
			"bad coordinate length - multilinestring 1",
			`{"type":"MultiLineString","coordinates":[[[0],[0]]]}`,
			"invalid geojson coordinate length: 1",
		},
		{
			"bad coordinate length - multipolygon 0",
			`{"type":"MultiPolygon","coordinates":[[[[],[0]]]]}`,
			"invalid geojson coordinate length: 0",
		},
		{
			"bad coordinate length - multipolygon 1",
			`{"type":"MultiPolygon","coordinates":[[[[0],[0]]]]}`,
			"invalid geojson coordinate length: 1",
		},

		{
			"bad coordinates shape - point",
			`{"type":"Point","coordinates":[[0,0]]}`,
			"json: cannot unmarshal array into Go value of type float64",
		},
		{
			"bad coordinates shape - linestring",
			`{"type":"LineString","coordinates":[[[0,0]]]}`,
			"json: cannot unmarshal array into Go value of type float64",
		},
		{
			"bad coordinates shape - polygon",
			`{"type":"Polygon","coordinates":[[0,0]]}`,
			"json: cannot unmarshal number into Go value of type []float64",
		},
		{
			"bad coordinates shape - multipolygon",
			`{"type":"MultiPolygon","coordinates":[[0,0]]}`,
			"json: cannot unmarshal number into Go value of type [][]float64",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, err := geom.UnmarshalGeoJSON([]byte(tc.geojson))
			if err == nil {
				t.Fatal("expected an error but got nil")
			}
			want := "invalid GeoJSON syntax: " + tc.errorText
			if err.Error() != want {
				t.Logf("got:  %q", err.Error())
				t.Logf("want: %q", want)
				t.Errorf("mismatch")
			}
		})
	}
}

func TestGeoJSONUnmarshalInvalid(t *testing.T) {
	for i, geojson := range []string{
		// GeoJSON cannot represent empty points in MultiPoints. When parsing,
		// we should complain that the dimensionality cannot be zero.
		`{"type":"MultiPoint","coordinates":[[0,1],[]]}`,
		`{"type":"MultiPoint","coordinates":[[],[0,1]]}`,
		`{"type":"MultiPoint","coordinates":[[]]}`,

		// Coordinates (other than the first level) must have either 2 or 3 (or more) parts.
		`{"type":"LineString","coordinates":[[0,1],[]]}`,
		`{"type":"LineString","coordinates":[[0,1],[3]]}`,
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var g geom.Geometry
			err := json.NewDecoder(strings.NewReader(geojson)).Decode(&g)
			if err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}

func TestGeoJSONUnmarshalNoValidate(t *testing.T) {
	for i, geojson := range []string{
		`{"type":"LineString","coordinates":[[0,0],[0,0]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,1],[1,0],[0,1],[0,0]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[0,0]]]}`,
		`{"type":"MultiLineString","coordinates":[[[0,0],[0,0]]]}`,
		`{"type":"MultiPolygon","coordinates":[[[[0,0],[1,1],[1,0],[0,1],[0,0]]]]}`,
		`{"type":"MultiPolygon","coordinates":[[[[0,0],[0,0]]]]}`,
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if _, err := geom.UnmarshalGeoJSON([]byte(geojson)); err == nil {
				t.Fatal("invalid test case -- geometry should be invalid")
			}
			if _, err := geom.UnmarshalGeoJSON([]byte(geojson), geom.NoValidate{}); err != nil {
				t.Errorf("got error but didn't expect one because validations are disabled")
			}
		})
	}
}

func TestGeoJSONUnmarshalIntoConcreteGeometryValid(t *testing.T) {
	for _, tc := range []struct {
		target interface {
			json.Unmarshaler
			AsGeometry() geom.Geometry
			Type() geom.GeometryType
		}
		geojson string
		wantWKT string
	}{
		{
			target:  new(geom.GeometryCollection),
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]}]}`,
			wantWKT: "GEOMETRYCOLLECTION(POINT(1 2))",
		},
		{
			target:  new(geom.Point),
			geojson: `{"type":"Point","coordinates":[1,2]}`,
			wantWKT: "POINT(1 2)",
		},
		{
			target:  new(geom.LineString),
			geojson: `{"type":"LineString","coordinates":[[1,2],[3,4]]}`,
			wantWKT: "LINESTRING(1 2,3 4)",
		},
		{
			target:  new(geom.Polygon),
			geojson: `{"type":"Polygon","coordinates":[[[0,0],[0,1],[1,0],[0,0]]]}`,
			wantWKT: "POLYGON((0 0,0 1,1 0,0 0))",
		},
		{
			target:  new(geom.MultiPoint),
			geojson: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
			wantWKT: "MULTIPOINT((1 2))",
		},
		{
			target:  new(geom.MultiLineString),
			geojson: `{"type":"MultiLineString","coordinates":[[[1,2],[3,4]]]}`,
			wantWKT: "MULTILINESTRING((1 2,3 4))",
		},
		{
			target:  new(geom.MultiPolygon),
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0],[0,1],[1,0],[0,0]]]]}`,
			wantWKT: "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))",
		},
	} {
		t.Run(tc.target.Type().String(), func(t *testing.T) {
			err := json.Unmarshal([]byte(tc.geojson), tc.target)
			expectNoErr(t, err)
			expectGeomEq(t, tc.target.AsGeometry(), geomFromWKT(t, tc.wantWKT))
		})
	}
}

func TestGeoJSONUnmarshalIntoConcreteGeometryWrongType(t *testing.T) {
	for _, tc := range []struct {
		dest interface {
			json.Unmarshaler
			Type() geom.GeometryType
		}
	}{
		{new(geom.GeometryCollection)},
		{new(geom.Point)},
		{new(geom.LineString)},
		{new(geom.Polygon)},
		{new(geom.MultiPoint)},
		{new(geom.MultiLineString)},
		{new(geom.MultiPolygon)},
	} {
		t.Run("dest_"+tc.dest.Type().String(), func(t *testing.T) {
			for _, geojson := range []string{
				`{"type":"Point","coordinates":[1,2]}`,
				`{"type":"MultiPoint","coordinates":[[1,2]]}`,
				`{"type":"LineString","coordinates":[[1,2],[3,4]]}`,
				`{"type":"MultiLineString","coordinates":[[[1,2],[3,4]]]}`,
				`{"type":"Polygon","coordinates":[[[0,0],[0,1],[1,0],[0,0]]]}`,
				`{"type":"MultiPolygon","coordinates":[[[[0,0],[0,1],[1,0],[0,0]]]]}`,
				`{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]}]}`,
			} {
				srcTyp := geomFromGeoJSON(t, geojson).Type()
				t.Run("source_"+srcTyp.String(), func(t *testing.T) {
					destType := tc.dest.Type()
					if srcTyp == destType {
						// This test suite is for negative test cases, however
						// this test case would always succeed.
						return
					}
					err := json.Unmarshal([]byte(geojson), tc.dest)
					want := geom.UnmarshalGeoJSONSourceDestinationMismatchError{
						SourceType:      srcTyp,
						DestinationType: destType,
					}
					expectErrIs(t, err, want)
				})
			}
		})
	}
}

func TestGeoJSONUnmarshalIntoConcreteGeometryDoesNotAlterParent(t *testing.T) {
	t.Run("MultiPoint", func(t *testing.T) {
		const parentWKT = "MULTIPOINT((1 2))"
		parent := geomFromWKT(t, parentWKT).MustAsMultiPoint()
		child := parent.PointN(0)
		err := json.Unmarshal([]byte(`{"type":"Point","coordinates":[9,9]}`), &child)
		expectNoErr(t, err)
		expectGeomEq(t, parent.AsGeometry(), geomFromWKT(t, parentWKT))
	})
	t.Run("MultiLineString", func(t *testing.T) {
		const parentWKT = "MULTILINESTRING((1 2,3 4))"
		parent := geomFromWKT(t, parentWKT).MustAsMultiLineString()
		child := parent.LineStringN(0)
		err := json.Unmarshal([]byte(`{"type":"LineString","coordinates":[[9,9],[8,8]]}`), &child)
		expectNoErr(t, err)
		expectGeomEq(t, parent.AsGeometry(), geomFromWKT(t, parentWKT))
	})
	t.Run("MultiPolygon", func(t *testing.T) {
		const parentWKT = "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))"
		parent := geomFromWKT(t, parentWKT).MustAsMultiPolygon()
		child := parent.PolygonN(0)
		err := json.Unmarshal([]byte(`{"type":"Polygon","coordinates":[[[4,4],[4,5],[5,4],[4,4]]]}`), &child)
		expectNoErr(t, err)
		expectGeomEq(t, parent.AsGeometry(), geomFromWKT(t, parentWKT))
	})
	t.Run("GeometryCollection", func(t *testing.T) {
		const parentWKT = "GEOMETRYCOLLECTION(POINT(1 2))"
		parent := geomFromWKT(t, parentWKT).MustAsGeometryCollection()
		child := parent.GeometryN(0)
		err := json.Unmarshal([]byte(`{"type":"Point","coordinates":[9,9]}`), &child)
		expectNoErr(t, err)
		expectGeomEq(t, parent.AsGeometry(), geomFromWKT(t, parentWKT))
	})
}
