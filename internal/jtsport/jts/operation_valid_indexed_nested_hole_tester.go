package jts

// OperationValid_IndexedNestedHoleTester tests whether any holes of a Polygon are
// nested inside another hole, using a spatial index to speed up the comparisons.
//
// The logic assumes that the holes do not overlap and have no collinear segments
// (so they are properly nested, and there are no duplicate holes).
//
// The situation where every vertex of a hole touches another hole
// is invalid because either the hole is nested,
// or else it disconnects the polygon interior.
// This class detects the nested situation.
// The disconnected interior situation must be checked elsewhere.
type OperationValid_IndexedNestedHoleTester struct {
	polygon  *Geom_Polygon
	index    Index_SpatialIndex
	nestedPt *Geom_Coordinate
}

// OperationValid_NewIndexedNestedHoleTester creates a new IndexedNestedHoleTester for
// the given polygon.
func OperationValid_NewIndexedNestedHoleTester(poly *Geom_Polygon) *OperationValid_IndexedNestedHoleTester {
	tester := &OperationValid_IndexedNestedHoleTester{
		polygon: poly,
	}
	tester.loadIndex()
	return tester
}

func (t *OperationValid_IndexedNestedHoleTester) loadIndex() {
	t.index = IndexStrtree_NewSTRtree()

	for i := 0; i < t.polygon.GetNumInteriorRing(); i++ {
		hole := t.polygon.GetInteriorRingN(i)
		env := hole.GetEnvelopeInternal()
		t.index.Insert(env, hole)
	}
}

// GetNestedPoint gets a point on a nested hole, if one exists.
//
// Returns a point on a nested hole, or nil if none are nested.
func (t *OperationValid_IndexedNestedHoleTester) GetNestedPoint() *Geom_Coordinate {
	return t.nestedPt
}

// IsNested tests if any hole is nested (contained) within another hole.
// This is invalid.
// The nested point will be set to reflect this.
//
// Returns true if some hole is nested.
func (t *OperationValid_IndexedNestedHoleTester) IsNested() bool {
	for i := 0; i < t.polygon.GetNumInteriorRing(); i++ {
		hole := t.polygon.GetInteriorRingN(i)

		results := t.index.Query(hole.GetEnvelopeInternal())
		for _, result := range results {
			testHole := result.(*Geom_LinearRing)
			if hole == testHole {
				continue
			}

			// Hole is not fully covered by test hole, so cannot be nested
			if !testHole.GetEnvelopeInternal().CoversEnvelope(hole.GetEnvelopeInternal()) {
				continue
			}

			if OperationValid_PolygonTopologyAnalyzer_IsRingNested(hole, testHole) {
				//TODO: find a hole point known to be inside
				t.nestedPt = hole.GetCoordinateN(0)
				return true
			}
		}
	}
	return false
}
