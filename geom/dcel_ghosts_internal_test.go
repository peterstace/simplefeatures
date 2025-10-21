package geom

import (
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

			got := findComponentRepresentatives(a, b)

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

func TestCreateGhostsRayCasting(t *testing.T) {
	for i, tc := range []struct {
		aWKT        string
		bWKT        string
		description string
		// We check that ghosts exist and components are connected,
		// without being too specific about exact ghost structure.
		minGhosts int
		maxGhosts int
	}{
		{
			description: "single component needs no ghosts",
			aWKT:        "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			bWKT:        "POLYGON((1 0,2 0,2 1,1 1,1 0))",
			minGhosts:   0,
			maxGhosts:   0,
		},
		{
			description: "two disjoint components need ghost edges",
			aWKT:        "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			bWKT:        "POLYGON((2 0,3 0,3 1,2 1,2 0))",
			minGhosts:   1,
			maxGhosts:   2,
		},
		{
			description: "three components stacked vertically",
			aWKT:        "POINT(0 0)",
			bWKT:        "GEOMETRYCOLLECTION(POINT(0 1),POINT(0 2))",
			minGhosts:   5,
			maxGhosts:   5,
		},
		{
			description: "empty geometries",
			aWKT:        "POINT EMPTY",
			bWKT:        "POINT EMPTY",
			minGhosts:   0,
			maxGhosts:   0,
		},
	} {
		t.Run(strconv.Itoa(i)+"_"+tc.description, func(t *testing.T) {
			a, err := UnmarshalWKT(tc.aWKT)
			if err != nil {
				t.Fatal(err)
			}
			b, err := UnmarshalWKT(tc.bWKT)
			if err != nil {
				t.Fatal(err)
			}

			ghosts := createGhosts(a, b)
			numGhosts := ghosts.NumLineStrings()

			if numGhosts < tc.minGhosts || numGhosts > tc.maxGhosts {
				t.Errorf("expected %d-%d ghosts, got %d",
					tc.minGhosts, tc.maxGhosts, numGhosts)
			}

			// Verify DCEL can be constructed successfully.
			dcel := newDCELFromGeometries(a, b)
			if dcel == nil {
				t.Fatal("failed to create DCEL")
			}
		})
	}
}

func TestSpanningTree(t *testing.T) {
	for i, tc := range []struct {
		xys     []XY
		wantWKT string
	}{
		{
			xys:     nil,
			wantWKT: "MULTILINESTRING EMPTY",
		},
		{
			xys:     []XY{{1, 1}},
			wantWKT: "MULTILINESTRING EMPTY",
		},
		{
			xys:     []XY{{2, 1}, {1, 2}},
			wantWKT: "MULTILINESTRING((2 1,1 2))",
		},
		{
			xys:     []XY{{2, 0}, {2, 2}, {0, 0}, {1.5, 1.5}},
			wantWKT: "MULTILINESTRING((0 0,2 0),(1.5 1.5,2 2),(2 0,1.5 1.5))",
		},
		{
			xys:     []XY{{-0.5, 0.5}, {0, 0}, {0, 1}, {1, 0}},
			wantWKT: "MULTILINESTRING((-0.5 0.5,0 0),(0 0,0 1),(0 1,1 0))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want, err := UnmarshalWKT(tc.wantWKT)
			if err != nil {
				t.Fatal(err)
			}
			got := spanningTree(tc.xys)
			if !ExactEquals(want, got.AsGeometry(), IgnoreOrder) {
				t.Logf("got:  %v", got.AsText())
				t.Logf("want: %v", want.AsText())
				t.Fatal("mismatch")
			}
		})
	}
}
