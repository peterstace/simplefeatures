package geom

import (
	"strconv"
	"testing"
)

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
			minGhosts:   2,
			maxGhosts:   3,
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
