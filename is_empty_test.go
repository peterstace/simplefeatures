package simplefeatures_test

import (
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestIsEmpty(t *testing.T) {
	for _, tt := range []struct {
		wkt  string
		want bool
	}{
		{"POINT EMPTY", true},
		{"POINT(1 1)", false},
		{"LINESTRING EMPTY", true},
		{"LINESTRING(0 0,1 1)", false},
		{"LINESTRING(0 0,1 1,2 2)", false},
		{"LINESTRING(0 0,1 1,1 0,0 0)", false},
		// TODO: linear ring
		{"POLYGON EMPTY", true},
		{"POLYGON((0 0,1 1,1 0,0 0))", false},
		{"MULTIPOINT EMPTY", true},
		{"MULTIPOINT((0 0))", false},
		{"MULTIPOINT((0 0),(1 1))", false},
		{"MULTILINESTRING EMPTY", true},
		{"MULTILINESTRING((0 0,1 1,2 2))", false},
		{"MULTIPOLYGON EMPTY", true},
		{"MULTIPOLYGON(((0 0,1 0,1 1,0 0)))", false},
		{"GEOMETRYCOLLECTION EMPTY", true},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", true},
		{"GEOMETRYCOLLECTION(POLYGON EMPTY)", true},
		{"GEOMETRYCOLLECTION(POINT(1 1))", false},
	} {
		t.Run(tt.wkt, func(t *testing.T) {
			geom, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err != nil {
				t.Fatal(err)
			}
			got := geom.IsEmpty()
			if got != tt.want {
				t.Errorf("want=%v got=%v", tt.want, got)
			}
		})
	}
}

func TestIsEmptyLineString(t *testing.T) {
	// Tested on its own, since it cannot be constructed from WKT.
	ring, err := NewLinearRing([]Coordinates{
		{XY: XY{0, 0}}, {XY: XY{1, 0}}, {XY: XY{1, 1}}, {XY: XY{0, 0}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ring.IsEmpty() {
		t.Errorf("expected to not be empty")
	}
}
