package geom

import (
	"fmt"
	"strconv"
	"testing"
)

func TestFindComponentRepresentatives(t *testing.T) {
	for i, tc := range []struct {
		aWKT string
		bWKT string
		want []XY
	}{
		{
			aWKT: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			bWKT: "POLYGON((1 0,2 0,2 1,1 1,1 0))",
			want: []XY{{2, 1}},
		},
		{
			aWKT: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			bWKT: "POLYGON((2 0,3 0,3 1,2 1,2 0))",
			want: []XY{{1, 1}, {3, 1}},
		},
		{
			aWKT: "LINESTRING(0 0,1 0,2 0)",
			bWKT: "LINESTRING(3 0,4 0)",
			want: []XY{{2, 0}, {4, 0}},
		},
		{
			aWKT: "LINESTRING(0 0,0 1,0 2)",
			bWKT: "LINESTRING(1 0,1 1,1 2)",
			want: []XY{{0, 2}, {1, 2}},
		},
		{
			aWKT: "LINESTRING(5 5,5 3,5 1)",
			bWKT: "POINT EMPTY",
			want: []XY{{5, 5}},
		},
		{
			aWKT: "POINT EMPTY",
			bWKT: "POINT EMPTY",
			want: nil,
		},
		{
			aWKT: "GEOMETRYCOLLECTION(POLYGON((0 0,1 0,1 1,0 1,0 0)),POLYGON((1 0,2 0,2 1,1 1,1 0)))",
			bWKT: "POINT EMPTY",
			want: []XY{{2, 1}},
		},
		{
			aWKT: "MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((3 0,4 0,4 1,3 1,3 0)))",
			bWKT: "POINT EMPTY",
			want: []XY{{1, 1}, {4, 1}},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a, err := UnmarshalWKT(tc.aWKT)
			if err != nil {
				t.Fatal(err)
			}
			b, err := UnmarshalWKT(tc.bWKT)
			if err != nil {
				t.Fatal(err)
			}

			xys := collectControlPoints(a, b)
			lines := appendLines(nil, NewGeometryCollection([]Geometry{a, b}).AsGeometry())
			got := findConnectedComponentRepresentatives(xys, lines)

			if len(got) != len(tc.want) {
				t.Fatalf("length mismatch: got %d, want %d", len(got), len(tc.want))
			}

			gotSet := make(map[XY]bool)
			for _, xy := range got {
				gotSet[xy] = true
			}
			wantSet := make(map[XY]bool)
			for _, xy := range tc.want {
				wantSet[xy] = true
			}

			for xy := range wantSet {
				if !gotSet[xy] {
					t.Errorf("missing expected point: %v", xy)
				}
			}
			for xy := range gotSet {
				if !wantSet[xy] {
					t.Errorf("unexpected point: %v", xy)
				}
			}
		})
	}
}

func TestPrepareGeometriesForDCEL(t *testing.T) {
	for i, tc := range []struct {
		name   string
		inputA string
		inputB string
		wantA  string
		wantB  string
		wantG  string
	}{
		// Test cases for linking disjoint components together:
		{
			name:   "empty inputs",
			inputA: "GEOMETRYCOLLECTION EMPTY",
			inputB: "GEOMETRYCOLLECTION EMPTY",
			wantA:  "GEOMETRYCOLLECTION EMPTY",
			wantB:  "GEOMETRYCOLLECTION EMPTY",
			wantG:  "MULTILINESTRING EMPTY",
		},
		{
			name:   "simple polygon as one input",
			inputA: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			inputB: "GEOMETRYCOLLECTION EMPTY",
			wantA:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wantB:  "GEOMETRYCOLLECTION EMPTY",
			wantG:  "MULTILINESTRING EMPTY",
		},
		{
			name:   "simple polygon as the other input",
			inputA: "GEOMETRYCOLLECTION EMPTY",
			inputB: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wantA:  "GEOMETRYCOLLECTION EMPTY",
			wantB:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wantG:  "MULTILINESTRING EMPTY",
		},
		{
			name:   "polygon with a hole",
			inputA: "POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))",
			inputB: "GEOMETRYCOLLECTION EMPTY",
			wantA:  "POLYGON((0 0,0 3,3 3,3 2,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))",
			wantB:  "GEOMETRYCOLLECTION EMPTY",
			wantG:  "MULTILINESTRING((2 2,3 2))",
		},
		{
			name:   "polygon with two vertically stacked holes",
			inputA: "POLYGON((0 0,0 5,3 5,3 0,0 0),(1 1,2 1,2 2,1 2,1 1),(1 3,2 3,2 4,1 4,1 3))",
			inputB: "GEOMETRYCOLLECTION EMPTY",
			wantA:  "POLYGON((0 0,0 5,3 5,3 4,3 2,3 0,0 0),(1 1,2 1,2 2,1 2,1 1),(1 3,2 3,2 4,1 4,1 3))",
			wantB:  "GEOMETRYCOLLECTION EMPTY",
			wantG:  "MULTILINESTRING((2 2,3 2),(2 4,3 4))",
		},
		{
			name:   "polygon with two horizontally stacked holes",
			inputA: "POLYGON((0 0,0 3,5 3,5 2,5 0,0 0),(1 1,1 2,2 2,2 1,1 1),(3 1,3 2,4 2,4 1,3 1))",
			inputB: "GEOMETRYCOLLECTION EMPTY",
			wantA:  "POLYGON((0 0,0 3,5 3,5 2,5 0,0 0),(1 1,1 2,2 2,2 1,1 1),(3 1,3 2,4 2,4 1,3 1))",
			wantB:  "GEOMETRYCOLLECTION EMPTY",
			wantG:  "MULTILINESTRING((2 2,3 2),(4 2,5 2))",
		},
		{
			name:   "two horizontally stacked polygons",
			inputA: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			inputB: "POLYGON((2 0,2 1,3 1,3 0,2 0))",
			wantA:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wantB:  "POLYGON((2 0,2 1,3 1,3 0,2 0))",
			wantG:  "MULTILINESTRING((1 1,2 1))",
		},
		{
			name:   "two vertically stacked polygons",
			inputA: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			inputB: "POLYGON((0 2,0 3,1 3,1 2,0 2))",
			wantA:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wantB:  "POLYGON((0 2,0 3,1 3,1 2,0 2))",
			wantG:  "MULTILINESTRING((1 1,2 1),(1 3,2 3),(2 1,2 3))",
		},

		// Test cases for the specifics of _how_ disjoint components are linked:
		{
			name:   "link to the closest point (bottom)",
			inputA: "POLYGON((0 0,0 2,2 2,0 0))",
			inputB: "LINESTRING(3 1,3 4)",
			wantA:  "POLYGON((0 0,0 2,2 2,0 0))",
			wantB:  "LINESTRING(3 1,3 2,3 4)",
			wantG:  "MULTILINESTRING((2 2,3 2))",
		},
		{
			name:   "link to the closest point (top)",
			inputA: "POLYGON((0 0,0 2,2 2,0 0))",
			inputB: "LINESTRING(3 0,3 3)",
			wantA:  "POLYGON((0 0,0 2,2 2,0 0))",
			wantB:  "LINESTRING(3 0,3 2,3 3)",
			wantG:  "MULTILINESTRING((2 2,3 2))",
		},
		{
			name:   "link to the highest point if both are equal distance",
			inputA: "POLYGON((0 0,0 2,2 2,0 0))",
			inputB: "LINESTRING(3 0,3 4)",
			wantA:  "POLYGON((0 0,0 2,2 2,0 0))",
			wantB:  "LINESTRING(3 0,3 2,3 4)",
			wantG:  "MULTILINESTRING((2 2,3 2))",
		},
		{
			name:   "only link to a point if it is unobstructed (bottom)",
			inputA: "LINESTRING(0 0,2 2)",
			inputB: "LINESTRING(8 0,4 4,4 3,3 4)",
			wantA:  "LINESTRING(0 0,2 2)",
			wantB:  "LINESTRING(8 0,6 2,4 4,4 3,3 4)",
			wantG:  "MULTILINESTRING((2 2,6 2))",
		},
		{
			name:   "only link to a point if it is unobstructed (top)",
			inputA: "LINESTRING(0 4,2 2)",
			inputB: "LINESTRING(8 4,4 0,4 1,3 0)",
			wantA:  "LINESTRING(0 4,2 2)",
			wantB:  "LINESTRING(8 4,6 2,4 0,4 1,3 0)",
			wantG:  "MULTILINESTRING((2 2,6 2))",
		},
		{
			name:   "only link to the right",
			inputA: "LINESTRING(0 0,2 2)",
			inputB: "LINESTRING(1 4,5 0)",
			wantA:  "LINESTRING(0 0,2 2)",
			wantB:  "LINESTRING(1 4,3 2,5 0)",
			wantG:  "MULTILINESTRING((2 2,3 2))",
		},
		{
			name:   "split edge if both endpoints are obstructed",
			inputA: "POINT(0 2)",
			inputB: "MULTILINESTRING((0.5 0.5,1.5 1.5),(0.5 3.5,1.5 2.5),(2 0,2 4))",
			wantA:  "POINT(0 2)",
			wantB:  "MULTILINESTRING((0.5 0.5,1.5 1.5),(0.5 3.5,1.5 2.5),(2 0,2 1.5,2 2,2 2.5,2 4))",
			wantG:  "MULTILINESTRING((0 2,2 2),(1.5 2.5,2 2.5),(1.5 1.5,2 1.5))",
		},

		// Test cases for creating new nodes:
		{
			name:   "two linked polygons",
			inputA: "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			inputB: "POLYGON((1 1,1 3,3 3,3 1,1 1))",
			wantA:  "POLYGON((0 0,0 2,1 2,2 2,2 1,2 0,0 0))",
			wantB:  "POLYGON((1 1,1 2,1 3,3 3,3 1,2 1,1 1))",
			wantG:  "MULTILINESTRING EMPTY",
		},
		{
			name:   "polygon with shared edge",
			inputA: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			inputB: "POLYGON((1 0,1 1,2 1,2 0,1 0))",
			wantA:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wantB:  "POLYGON((1 0,1 1,2 1,2 0,1 0))",
			wantG:  "MULTILINESTRING EMPTY",
		},
		{
			name:   "polygon with partially shared edge",
			inputA: "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			inputB: "POLYGON((2 1,2 3,4 3,4 1,2 1))",
			wantA:  "POLYGON((0 0,0 2,2 2,2 1,2 0,0 0))",
			wantB:  "POLYGON((2 1,2 2,2 3,4 3,4 1,2 1))",
			wantG:  "MULTILINESTRING EMPTY",
		},
		{
			name:   "polygon vertex touches the edge of another polygon",
			inputA: "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			inputB: "POLYGON((2 1,4 0,4 2,2 1))",
			wantA:  "POLYGON((0 0,0 2,2 2,2 1,2 0,0 0))",
			wantB:  "POLYGON((2 1,4 0,4 2,2 1))",
			wantG:  "MULTILINESTRING EMPTY",
		},
	} {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			inputA := wktToGeom(t, tc.inputA)
			inputB := wktToGeom(t, tc.inputB)
			gotA, gotB, gotG := prepareGeometriesForDCEL(inputA, inputB)
			testExactEqualsWKT(t, gotA, tc.wantA, IgnoreOrder)
			testExactEqualsWKT(t, gotB, tc.wantB, IgnoreOrder)
			testExactEqualsWKT(t, gotG.AsGeometry(), tc.wantG, IgnoreOrder)
		})
	}
}
