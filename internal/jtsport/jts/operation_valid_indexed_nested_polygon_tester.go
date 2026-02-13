package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationValid_IndexedNestedPolygonTester tests whether a MultiPolygon has any
// element polygon improperly nested inside another polygon, using a spatial
// index to speed up the comparisons.
//
// The logic assumes that the polygons do not overlap and have no collinear segments.
// So the polygon rings may touch at discrete points,
// but they are properly nested, and there are no duplicate rings.
type OperationValid_IndexedNestedPolygonTester struct {
	multiPoly *Geom_MultiPolygon
	index     Index_SpatialIndex
	locators  []*AlgorithmLocate_IndexedPointInAreaLocator
	nestedPt  *Geom_Coordinate
}

// OperationValid_NewIndexedNestedPolygonTester creates a new IndexedNestedPolygonTester
// for the given MultiPolygon.
func OperationValid_NewIndexedNestedPolygonTester(multiPoly *Geom_MultiPolygon) *OperationValid_IndexedNestedPolygonTester {
	tester := &OperationValid_IndexedNestedPolygonTester{
		multiPoly: multiPoly,
	}
	tester.loadIndex()
	return tester
}

func (t *OperationValid_IndexedNestedPolygonTester) loadIndex() {
	t.index = IndexStrtree_NewSTRtree()

	for i := 0; i < t.multiPoly.GetNumGeometries(); i++ {
		poly := java.Cast[*Geom_Polygon](t.multiPoly.GetGeometryN(i))
		env := poly.GetEnvelopeInternal()
		t.index.Insert(env, i)
	}
}

func (t *OperationValid_IndexedNestedPolygonTester) getLocator(polyIndex int) *AlgorithmLocate_IndexedPointInAreaLocator {
	if t.locators == nil {
		t.locators = make([]*AlgorithmLocate_IndexedPointInAreaLocator, t.multiPoly.GetNumGeometries())
	}
	locator := t.locators[polyIndex]
	if locator == nil {
		locator = AlgorithmLocate_NewIndexedPointInAreaLocator(t.multiPoly.GetGeometryN(polyIndex))
		t.locators[polyIndex] = locator
	}
	return locator
}

// GetNestedPoint gets a point on a nested polygon, if one exists.
//
// Returns a point on a nested polygon, or nil if none are nested.
func (t *OperationValid_IndexedNestedPolygonTester) GetNestedPoint() *Geom_Coordinate {
	return t.nestedPt
}

// IsNested tests if any polygon is improperly nested (contained) within another polygon.
// This is invalid.
// The nested point will be set to reflect this.
//
// Returns true if some polygon is nested.
func (t *OperationValid_IndexedNestedPolygonTester) IsNested() bool {
	for i := 0; i < t.multiPoly.GetNumGeometries(); i++ {
		poly := java.Cast[*Geom_Polygon](t.multiPoly.GetGeometryN(i))
		shell := poly.GetExteriorRing()

		results := t.index.Query(poly.GetEnvelopeInternal())
		for _, result := range results {
			polyIndex := result.(int)
			possibleOuterPoly := java.Cast[*Geom_Polygon](t.multiPoly.GetGeometryN(polyIndex))

			if poly == possibleOuterPoly {
				continue
			}
			// If polygon is not fully covered by candidate polygon it cannot be nested
			if !possibleOuterPoly.GetEnvelopeInternal().CoversEnvelope(poly.GetEnvelopeInternal()) {
				continue
			}

			t.nestedPt = t.findNestedPoint(shell, possibleOuterPoly, t.getLocator(polyIndex))
			if t.nestedPt != nil {
				return true
			}
		}
	}
	return false
}

// findNestedPoint finds an improperly nested point, if one exists.
//
// shell is the test polygon shell
// possibleOuterPoly is a polygon which may contain it
// locator is the locator for the outer polygon
//
// Returns a nested point, if one exists, or nil.
func (t *OperationValid_IndexedNestedPolygonTester) findNestedPoint(shell *Geom_LinearRing,
	possibleOuterPoly *Geom_Polygon, locator *AlgorithmLocate_IndexedPointInAreaLocator) *Geom_Coordinate {
	// Try checking two points, since checking point location is fast.
	shellPt0 := shell.GetCoordinateN(0)
	loc0 := locator.Locate(shellPt0)
	if loc0 == Geom_Location_Exterior {
		return nil
	}
	if loc0 == Geom_Location_Interior {
		return shellPt0
	}

	shellPt1 := shell.GetCoordinateN(1)
	loc1 := locator.Locate(shellPt1)
	if loc1 == Geom_Location_Exterior {
		return nil
	}
	if loc1 == Geom_Location_Interior {
		return shellPt1
	}

	// The shell points both lie on the boundary of
	// the polygon.
	// Nesting can be checked via the topology of the incident edges.
	return operationValid_IndexedNestedPolygonTester_findIncidentSegmentNestedPoint(shell, possibleOuterPoly)
}

// operationValid_IndexedNestedPolygonTester_findIncidentSegmentNestedPoint finds a point of a shell
// segment which lies inside a polygon, if any.
// The shell is assumed to touch the polygon only at shell vertices,
// and does not cross the polygon.
//
// shell is the shell to test
// poly is the polygon to test against
//
// Returns an interior segment point, or nil if the shell is nested correctly.
func operationValid_IndexedNestedPolygonTester_findIncidentSegmentNestedPoint(shell *Geom_LinearRing, poly *Geom_Polygon) *Geom_Coordinate {
	polyShell := poly.GetExteriorRing()
	if polyShell.IsEmpty() {
		return nil
	}

	if !OperationValid_PolygonTopologyAnalyzer_IsRingNested(shell, polyShell) {
		return nil
	}

	// Check if the shell is inside a hole (if there are any).
	// If so this is valid.
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		hole := poly.GetInteriorRingN(i)
		if hole.GetEnvelopeInternal().CoversEnvelope(shell.GetEnvelopeInternal()) &&
			OperationValid_PolygonTopologyAnalyzer_IsRingNested(shell, hole) {
			return nil
		}
	}

	// The shell is contained in the polygon, but is not contained in a hole.
	// This is invalid.
	return shell.GetCoordinateN(0)
}
