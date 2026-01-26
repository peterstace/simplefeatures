package jts

import (
	"sort"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Operation_BoundaryOp computes the boundary of a Geometry.
// Allows specifying the BoundaryNodeRule to be used.
// This operation will always return a Geometry of the appropriate
// dimension for the boundary (even if the input geometry is empty).
// The boundary of zero-dimensional geometries (Points) is
// always the empty GeometryCollection.
type Operation_BoundaryOp struct {
	geom        *Geom_Geometry
	geomFact    *Geom_GeometryFactory
	bnRule      Algorithm_BoundaryNodeRule
	endpointMap map[operation_CoordKey]*operation_EndpointEntry
}

// operation_EndpointEntry stores a coordinate and its count.
type operation_EndpointEntry struct {
	coord *Geom_Coordinate
	count int
}

// operation_CoordKey is a map key for coordinates using only X and Y.
// This matches Java's Coordinate.compareTo which only compares X and Y.
type operation_CoordKey struct {
	x, y float64
}

func operation_makeCoordKey(c *Geom_Coordinate) operation_CoordKey {
	return operation_CoordKey{x: c.X, y: c.Y}
}

// Operation_BoundaryOp_GetBoundary computes a geometry representing the
// boundary of a geometry.
func Operation_BoundaryOp_GetBoundary(g *Geom_Geometry) *Geom_Geometry {
	bop := Operation_NewBoundaryOp(g)
	return bop.GetBoundary()
}

// Operation_BoundaryOp_GetBoundaryWithRule computes a geometry representing the
// boundary of a geometry, using an explicit BoundaryNodeRule.
func Operation_BoundaryOp_GetBoundaryWithRule(g *Geom_Geometry, bnRule Algorithm_BoundaryNodeRule) *Geom_Geometry {
	bop := Operation_NewBoundaryOpWithRule(g, bnRule)
	return bop.GetBoundary()
}

// Operation_BoundaryOp_HasBoundary tests if a geometry has a boundary (it is
// non-empty). The semantics are:
//   - Empty geometries do not have boundaries.
//   - Points do not have boundaries.
//   - For linear geometries the existence of the boundary is determined by the
//     BoundaryNodeRule.
//   - Non-empty polygons always have a boundary.
func Operation_BoundaryOp_HasBoundary(geom *Geom_Geometry, boundaryNodeRule Algorithm_BoundaryNodeRule) bool {
	// Note that this does not handle geometry collections with a non-empty linear element.
	if geom.IsEmpty() {
		return false
	}
	switch geom.GetDimension() {
	case Geom_Dimension_P:
		return false
	case Geom_Dimension_L:
		// Linear geometries might have an empty boundary due to boundary node rule.
		boundary := Operation_BoundaryOp_GetBoundaryWithRule(geom, boundaryNodeRule)
		return !boundary.IsEmpty()
	case Geom_Dimension_A:
		return true
	}
	return true
}

// Operation_NewBoundaryOp creates a new instance for the given geometry.
func Operation_NewBoundaryOp(geom *Geom_Geometry) *Operation_BoundaryOp {
	return Operation_NewBoundaryOpWithRule(geom, Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE)
}

// Operation_NewBoundaryOpWithRule creates a new instance for the given geometry
// with an explicit BoundaryNodeRule.
func Operation_NewBoundaryOpWithRule(geom *Geom_Geometry, bnRule Algorithm_BoundaryNodeRule) *Operation_BoundaryOp {
	return &Operation_BoundaryOp{
		geom:     geom,
		geomFact: geom.GetFactory(),
		bnRule:   bnRule,
	}
}

// GetBoundary gets the computed boundary.
func (bop *Operation_BoundaryOp) GetBoundary() *Geom_Geometry {
	if java.InstanceOf[*Geom_LineString](bop.geom) {
		return bop.boundaryLineString(java.Cast[*Geom_LineString](bop.geom))
	}
	if java.InstanceOf[*Geom_MultiLineString](bop.geom) {
		return bop.boundaryMultiLineString(java.Cast[*Geom_MultiLineString](bop.geom))
	}
	return bop.geom.GetBoundary()
}

func (bop *Operation_BoundaryOp) getEmptyMultiPoint() *Geom_MultiPoint {
	return bop.geomFact.CreateMultiPoint()
}

func (bop *Operation_BoundaryOp) boundaryMultiLineString(mLine *Geom_MultiLineString) *Geom_Geometry {
	if bop.geom.IsEmpty() {
		return bop.getEmptyMultiPoint().Geom_Geometry
	}

	bdyPts := bop.computeBoundaryCoordinates(mLine)

	// Return Point or MultiPoint.
	if len(bdyPts) == 1 {
		return bop.geomFact.CreatePointFromCoordinate(bdyPts[0]).Geom_Geometry
	}
	// This handles 0 points case as well.
	return bop.geomFact.CreateMultiPointFromCoords(bdyPts).Geom_Geometry
}

func (bop *Operation_BoundaryOp) computeBoundaryCoordinates(mLine *Geom_MultiLineString) []*Geom_Coordinate {
	var bdyPts []*Geom_Coordinate
	bop.endpointMap = make(map[operation_CoordKey]*operation_EndpointEntry)

	for i := 0; i < mLine.GetNumGeometries(); i++ {
		line := java.GetLeaf(mLine.GetGeometryN(i)).(*Geom_LineString)
		if line.GetNumPoints() == 0 {
			continue
		}
		bop.addEndpoint(line.GetCoordinateN(0))
		bop.addEndpoint(line.GetCoordinateN(line.GetNumPoints() - 1))
	}

	// Collect coordinates from endpoints that are in the boundary.
	// Use sorted iteration for deterministic output.
	var entries []*operation_EndpointEntry
	for _, entry := range bop.endpointMap {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].coord.CompareTo(entries[j].coord) < 0
	})

	for _, entry := range entries {
		if bop.bnRule.IsInBoundary(entry.count) {
			bdyPts = append(bdyPts, entry.coord)
		}
	}

	return bdyPts
}

func (bop *Operation_BoundaryOp) addEndpoint(pt *Geom_Coordinate) {
	key := operation_makeCoordKey(pt)
	entry, exists := bop.endpointMap[key]
	if !exists {
		entry = &operation_EndpointEntry{coord: pt}
		bop.endpointMap[key] = entry
	}
	entry.count++
}

func (bop *Operation_BoundaryOp) boundaryLineString(line *Geom_LineString) *Geom_Geometry {
	if bop.geom.IsEmpty() {
		return bop.getEmptyMultiPoint().Geom_Geometry
	}

	if line.IsClosed() {
		// Check whether endpoints of valence 2 are on the boundary or not.
		closedEndpointOnBoundary := bop.bnRule.IsInBoundary(2)
		if closedEndpointOnBoundary {
			return line.GetStartPoint().Geom_Geometry
		}
		return bop.geomFact.CreateMultiPoint().Geom_Geometry
	}
	return bop.geomFact.CreateMultiPointFromPoints([]*Geom_Point{
		line.GetStartPoint(),
		line.GetEndPoint(),
	}).Geom_Geometry
}
