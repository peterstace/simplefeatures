package geom_test

import (
	"strconv"
	"testing"
)

func TestDumpGeometry(t *testing.T) {
	for i, tc := range []struct {
		inputWKT      string
		wantOutputWKT []string
	}{
		{
			"POINT(1 2)",
			[]string{"POINT(1 2)"},
		},
		{
			"POINT EMPTY",
			[]string{"POINT EMPTY"},
		},
		{
			"LINESTRING(0 0,1 1)",
			[]string{"LINESTRING(0 0,1 1)"},
		},
		{
			"LINESTRING EMPTY",
			[]string{"LINESTRING EMPTY"},
		},
		{
			"POLYGON((0 0,0 1,1 0,0 0))",
			[]string{"POLYGON((0 0,0 1,1 0,0 0))"},
		},
		{
			"POLYGON EMPTY",
			[]string{"POLYGON EMPTY"},
		},
		{
			"MULTIPOINT EMPTY",
			[]string{},
		},
		{
			"MULTIPOINT(1 2,EMPTY)",
			[]string{"POINT(1 2)", "POINT EMPTY"},
		},
		{
			"MULTILINESTRING EMPTY",
			[]string{},
		},
		{
			"MULTILINESTRING(EMPTY,(0 0,1 1))",
			[]string{"LINESTRING EMPTY", "LINESTRING(0 0,1 1)"},
		},
		{
			"MULTIPOLYGON EMPTY",
			[]string{},
		},
		{
			"MULTIPOLYGON(EMPTY,((0 0,0 1,1 0,0 0)))",
			[]string{"POLYGON EMPTY", "POLYGON((0 0,0 1,1 0,0 0))"},
		},
		{
			"GEOMETRYCOLLECTION EMPTY",
			[]string{},
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)",
			[]string{},
		},
		{
			"GEOMETRYCOLLECTION(POINT EMPTY)",
			[]string{"POINT EMPTY"},
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT EMPTY))",
			[]string{"POINT EMPTY"},
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT EMPTY))",
			[]string{"POINT EMPTY"},
		},
		{
			"GEOMETRYCOLLECTION(POINT(1 2),GEOMETRYCOLLECTION(POINT(3 4)))",
			[]string{"POINT(1 2)", "POINT(3 4)"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := geomsFromWKTs(t, tc.wantOutputWKT)
			got := geomFromWKT(t, tc.inputWKT).Dump()
			expectGeomsEq(t, got, want)
		})
	}
}

func TestDumpMultiPoint(t *testing.T) {
	for i, tc := range []struct {
		inputWKT      string
		wantOutputWKT []string
	}{
		{
			"MULTIPOINT EMPTY",
			[]string{},
		},
		{
			"MULTIPOINT(EMPTY)",
			[]string{"POINT EMPTY"},
		},
		{
			"MULTIPOINT(1 2)",
			[]string{"POINT(1 2)"},
		},
		{
			"MULTIPOINT(1 2,EMPTY)",
			[]string{"POINT(1 2)", "POINT EMPTY"},
		},
		{
			"MULTIPOINT(1 2,4 5)",
			[]string{"POINT(1 2)", "POINT(4 5)"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := geomsFromWKTs(t, tc.wantOutputWKT)
			got := upcastPoints(geomFromWKT(t, tc.inputWKT).AsMultiPoint().Dump())
			expectGeomsEq(t, got, want)
		})
	}
}

func TestDumpMultiLineString(t *testing.T) {
	for i, tc := range []struct {
		inputWKT      string
		wantOutputWKT []string
	}{
		{
			"MULTILINESTRING EMPTY",
			[]string{},
		},
		{
			"MULTILINESTRING(EMPTY)",
			[]string{"LINESTRING EMPTY"},
		},
		{
			"MULTILINESTRING((1 2,2 3))",
			[]string{"LINESTRING(1 2,2 3)"},
		},
		{
			"MULTILINESTRING((1 2,2 3),EMPTY)",
			[]string{"LINESTRING(1 2,2 3)", "LINESTRING EMPTY"},
		},
		{
			"MULTILINESTRING((1 2,2 3),(4 5,1 2))",
			[]string{"LINESTRING(1 2,2 3)", "LINESTRING(4 5,1 2)"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := geomsFromWKTs(t, tc.wantOutputWKT)
			got := upcastLineStrings(geomFromWKT(t, tc.inputWKT).AsMultiLineString().Dump())
			expectGeomsEq(t, got, want)
		})
	}
}

func TestDumpMultiPolygon(t *testing.T) {
	for i, tc := range []struct {
		inputWKT      string
		wantOutputWKT []string
	}{
		{
			"MULTIPOLYGON EMPTY",
			[]string{},
		},
		{
			"MULTIPOLYGON(EMPTY)",
			[]string{"POLYGON EMPTY"},
		},
		{
			"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))",
			[]string{"POLYGON((0 0,0 1,1 0,0 0))"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := geomsFromWKTs(t, tc.wantOutputWKT)
			got := upcastPolygons(geomFromWKT(t, tc.inputWKT).AsMultiPolygon().Dump())
			expectGeomsEq(t, got, want)
		})
	}
}

func TestDumpGeometryCollection(t *testing.T) {
	for i, tc := range []struct {
		inputWKT      string
		wantOutputWKT []string
	}{
		{
			"GEOMETRYCOLLECTION EMPTY",
			[]string{},
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)",
			[]string{},
		},
		{
			"GEOMETRYCOLLECTION(POINT EMPTY)",
			[]string{"POINT EMPTY"},
		},
		{
			"GEOMETRYCOLLECTION(MULTIPOINT(0 0,1 1))",
			[]string{"POINT(0 0)", "POINT(1 1)"},
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(4 4)))",
			[]string{"POINT(4 4)"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := geomsFromWKTs(t, tc.wantOutputWKT)
			got := geomFromWKT(t, tc.inputWKT).AsGeometryCollection().Dump()
			expectGeomsEq(t, got, want)
		})
	}
}
