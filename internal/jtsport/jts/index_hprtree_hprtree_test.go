package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// Tests ported from HPRtreeTest.java.

func TestHPRtreeEmptyTreeUsingListQuery(t *testing.T) {
	tree := jts.IndexHprtree_NewHPRtree()
	list := tree.Query(jts.Geom_NewEnvelopeFromXY(0, 1, 0, 1))
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestHPRtreeEmptyTreeUsingItemVisitorQuery(t *testing.T) {
	tree := jts.IndexHprtree_NewHPRtreeWithCapacity(0)
	visited := false
	visitor := jts.Index_NewItemVisitorFunc(func(item any) {
		visited = true
	})
	tree.QueryWithVisitor(jts.Geom_NewEnvelopeFromXY(0, 1, 0, 1), visitor)
	if visited {
		t.Error("visitor should not have been called for empty tree")
	}
}

func TestHPRtreeDisallowedInserts(t *testing.T) {
	tree := jts.IndexHprtree_NewHPRtreeWithCapacity(3)
	tree.Insert(jts.Geom_NewEnvelopeFromXY(0, 0, 0, 0), "item1")
	tree.Insert(jts.Geom_NewEnvelopeFromXY(0, 0, 0, 0), "item2")
	tree.Query(jts.Geom_NewEnvelope())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for insert after query, but got none")
		}
	}()
	tree.Insert(jts.Geom_NewEnvelopeFromXY(0, 0, 0, 0), "item3")
}

func TestHPRtreeQuery(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	tree := jts.IndexHprtree_NewHPRtreeWithCapacity(3)

	// Create line strings and insert them.
	ls1 := factory.CreateLineStringFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
	})
	tree.Insert(ls1.GetEnvelopeInternal(), "obj1")

	ls2 := factory.CreateLineStringFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(30, 30),
	})
	tree.Insert(ls2.GetEnvelopeInternal(), "obj2")

	ls3 := factory.CreateLineStringFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(30, 30),
	})
	tree.Insert(ls3.GetEnvelopeInternal(), "obj3")

	// Trigger build.
	tree.Query(jts.Geom_NewEnvelopeFromXY(5, 6, 5, 6))

	checkHPRtreeQuerySize(t, tree, 5, 6, 5, 6, 1)
	checkHPRtreeQuerySize(t, tree, 20, 30, 0, 10, 0)
	checkHPRtreeQuerySize(t, tree, 25, 26, 25, 26, 2)
	checkHPRtreeQuerySize(t, tree, 0, 100, 0, 100, 3)
}

func TestHPRtreeQuery3(t *testing.T) {
	tree := jts.IndexHprtree_NewHPRtree()
	for i := 0; i < 3; i++ {
		tree.Insert(jts.Geom_NewEnvelopeFromXY(float64(i), float64(i+1), float64(i), float64(i+1)), i)
	}
	tree.Query(jts.Geom_NewEnvelopeFromXY(0, 1, 0, 1))

	checkHPRtreeQuerySize(t, tree, 1, 2, 1, 2, 3)
	checkHPRtreeQuerySize(t, tree, 9, 10, 9, 10, 0)
}

func TestHPRtreeQuery10(t *testing.T) {
	tree := jts.IndexHprtree_NewHPRtree()
	for i := 0; i < 10; i++ {
		tree.Insert(jts.Geom_NewEnvelopeFromXY(float64(i), float64(i+1), float64(i), float64(i+1)), i)
	}
	tree.Query(jts.Geom_NewEnvelopeFromXY(0, 1, 0, 1))

	checkHPRtreeQuerySize(t, tree, 5, 6, 5, 6, 3)
	checkHPRtreeQuerySize(t, tree, 9, 10, 9, 10, 2)
	checkHPRtreeQuerySize(t, tree, 25, 26, 25, 26, 0)
	checkHPRtreeQuerySize(t, tree, 0, 10, 0, 10, 10)
}

func TestHPRtreeQuery100(t *testing.T) {
	checkHPRtreeQueryGrid(t, 100, jts.IndexHprtree_NewHPRtree())
}

func TestHPRtreeQuery100Cap8(t *testing.T) {
	checkHPRtreeQueryGrid(t, 100, jts.IndexHprtree_NewHPRtreeWithCapacity(8))
}

func TestHPRtreeQuery100Cap2(t *testing.T) {
	checkHPRtreeQueryGrid(t, 100, jts.IndexHprtree_NewHPRtreeWithCapacity(2))
}

func checkHPRtreeQueryGrid(t *testing.T, size int, tree *jts.IndexHprtree_HPRtree) {
	t.Helper()
	for i := 0; i < size; i++ {
		tree.Insert(jts.Geom_NewEnvelopeFromXY(float64(i), float64(i+1), float64(i), float64(i+1)), i)
	}
	tree.Query(jts.Geom_NewEnvelopeFromXY(0, 1, 0, 1))

	checkHPRtreeQuerySize(t, tree, 5, 6, 5, 6, 3)
	checkHPRtreeQuerySize(t, tree, 9, 10, 9, 10, 3)
	checkHPRtreeQuerySize(t, tree, 25, 26, 25, 26, 3)
	checkHPRtreeQuerySize(t, tree, 0, 10, 0, 10, 11)
}

func checkHPRtreeQuerySize(t *testing.T, tree *jts.IndexHprtree_HPRtree, x1, x2, y1, y2 float64, expected int) {
	t.Helper()
	result := tree.Query(jts.Geom_NewEnvelopeFromXY(x1, x2, y1, y2))
	if len(result) != expected {
		t.Errorf("Query([%v,%v],[%v,%v]): expected %d items, got %d", x1, x2, y1, y2, expected, len(result))
	}
}

func TestHPRtreeSpatialIndex(t *testing.T) {
	// Ported from SpatialIndexTester.java.
	const (
		cellExtent       = 20.31
		cellsPerGridSide = 10
		featureExtent    = 10.1
		offset           = 5.03
		queryExtent1     = 1.009
		queryExtent2     = 11.7
	)

	// Build source data: two grids of envelopes.
	var sourceData []*jts.Geom_Envelope
	addSourceData := func(off float64) {
		for i := 0; i < cellsPerGridSide; i++ {
			minx := float64(i)*cellExtent + off
			maxx := minx + featureExtent
			for j := 0; j < cellsPerGridSide; j++ {
				miny := float64(j)*cellExtent + off
				maxy := miny + featureExtent
				sourceData = append(sourceData, jts.Geom_NewEnvelopeFromXY(minx, maxx, miny, maxy))
			}
		}
	}
	addSourceData(0)
	addSourceData(offset)

	// Insert all envelopes into the tree.
	tree := jts.IndexHprtree_NewHPRtree()
	for _, env := range sourceData {
		tree.Insert(env, env)
	}

	// Helper to find envelopes that intersect a query envelope.
	intersectingEnvelopes := func(query *jts.Geom_Envelope) []*jts.Geom_Envelope {
		var result []*jts.Geom_Envelope
		for _, env := range sourceData {
			if env.IntersectsEnvelope(query) {
				result = append(result, env)
			}
		}
		return result
	}

	// Run queries at two different envelope extents.
	doTest := func(queryExtent float64) {
		gridExtent := cellExtent * cellsPerGridSide
		for x := 0.0; x < gridExtent; x += queryExtent {
			for y := 0.0; y < gridExtent; y += queryExtent {
				query := jts.Geom_NewEnvelopeFromXY(x, x+queryExtent, y, y+queryExtent)
				expected := intersectingEnvelopes(query)
				actual := tree.Query(query)
				// Index returns candidates, so it may return more than expected.
				if len(expected) > len(actual) {
					t.Errorf("Query at (%v,%v) extent %v: expected at least %d matches, got %d",
						x, y, queryExtent, len(expected), len(actual))
				}
				// Verify all expected matches are present.
				for _, exp := range expected {
					found := false
					for _, act := range actual {
						if act.(*jts.Geom_Envelope).Equals(exp) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Query at (%v,%v) extent %v: missing expected envelope %v",
							x, y, queryExtent, exp)
					}
				}
			}
		}
	}

	doTest(queryExtent1)
	doTest(queryExtent2)
}
