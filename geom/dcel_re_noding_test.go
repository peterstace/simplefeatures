package geom

import (
	"strconv"
	"testing"
)

func TestDCELReNoding(t *testing.T) {
	for i, tt := range []struct {
		input string
		cut   string
		want  string
	}{
		{
			input: "POINT(1 2)",
			cut:   "POINT(2 1)",
			want:  "POINT(1 2)",
		},
		{
			input: "POINT(1 2)",
			cut:   "POINT(1 2)",
			want:  "POINT(1 2)",
		},
		{
			input: "MULTIPOINT(1 2,2 1)",
			cut:   "POINT(2 1)",
			want:  "MULTIPOINT(1 2,2 1)",
		},
		{
			input: "LINESTRING(0 0,1 1)",
			cut:   "POINT(0.5 0.5)",
			want:  "LINESTRING(0 0,0.5 0.5,1 1)",
		},
		{
			input: "LINESTRING(0 0,1 1)",
			cut:   "GEOMETRYCOLLECTION(POINT(0.5 0.5),LINESTRING(0.5 1,1 0.5))",
			want:  "LINESTRING(0 0,0.5 0.5,0.75 0.75,1 1)",
		},
		{
			input: "LINESTRING EMPTY",
			cut:   "LINESTRING(0 1,1 0)",
			want:  "LINESTRING EMPTY",
		},
		{
			input: "LINESTRING(0 0,1 1)",
			cut:   "LINESTRING EMPTY",
			want:  "LINESTRING(0 0,1 1)",
		},
		{
			input: "LINESTRING(0 0,1 1)",
			cut:   "LINESTRING(0 1,1 0)",
			want:  "LINESTRING(0 0,0.5 0.5,1 1)",
		},
		{
			input: "LINESTRING(1 1,0 0)",
			cut:   "LINESTRING(0 1,1 0)",
			want:  "LINESTRING(1 1,0.5 0.5,0 0)",
		},
		{
			input: "LINESTRING(0 0,1 2,2 0)",
			cut:   "LINESTRING(0 1,2 1)",
			want:  "LINESTRING(0 0,0.5 1,1 2,1.5 1,2 0)",
		},
		{
			input: "LINESTRING(0 1,2 1)",
			cut:   "LINESTRING(0 0,1 2,2 0)",
			want:  "LINESTRING(0 1,0.5 1,1.5 1,2 1)",
		},
		{
			input: "LINESTRING(2 1,0 1)",
			cut:   "LINESTRING(0 0,1 2,2 0)",
			want:  "LINESTRING(2 1,1.5 1,0.5 1,0 1)",
		},
		{
			input: "MULTILINESTRING((0 1,2 1),(0 2,2 2))",
			cut:   "LINESTRING(0 0,2 4)",
			want:  "MULTILINESTRING((0 1,0.5 1,2 1),(0 2,1 2,2 2))",
		},
		{
			input: "POLYGON((0 0,2 0,1 2,0 0))",
			cut:   "POLYGON((0 1,2 1,1 3,0 1))",
			want:  "POLYGON((0 0,2 0,1.5 1,1 2,0.5 1,0 0))",
		},
		{
			input: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
			cut:   "LINESTRING(-1 0.5,4 0.5)",
			want:  "MULTIPOLYGON(((0 0,0 0.5,0 1,1 1,1 0.5,1 0,0 0)),((2 0,2 0.5,2 1,3 1,3 0.5,3 0,2 0)))",
		},
		{
			input: "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 1,1 0,0 0)),LINESTRING(2 0,2 1))",
			cut:   "LINESTRING(-1 0.5,4 0.5)",
			want:  "GEOMETRYCOLLECTION(POLYGON((0 0,0 0.5,0 1,1 1,1 0.5,1 0,0 0)),LINESTRING(2 0,2 0.5,2 1))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("input: %v", tt.input)
			t.Logf("cut:   %v", tt.cut)
			t.Logf("want:  %v", tt.want)

			inputG, err := UnmarshalWKT(tt.input)
			if err != nil {
				t.Fatalf("could not unmarshal geometry: %v", err)
			}
			cutG, err := UnmarshalWKT(tt.cut)
			if err != nil {
				t.Fatalf("could not unmarshal geometry: %v", err)
			}
			wantG, err := UnmarshalWKT(tt.want)
			if err != nil {
				t.Fatalf("could not unmarshal geometry: %v", err)
			}

			cutSet := newCutSet(cutG, Geometry{})
			got := reNodeGeometry(inputG, cutSet)

			if !got.EqualsExact(wantG) {
				t.Logf("MISMATCH")
				t.Errorf("got: %v", got.AsText())
			}
		})
	}
}
