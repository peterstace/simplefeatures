package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelateng_RelatePointLocator locates a point on a geometry, including
// mixed-type collections. The dimension of the containing geometry element is
// also determined. GeometryCollections are handled with union semantics; i.e.
// the location of a point is that location of that point on the union of the
// elements of the collection.
//
// Union semantics for GeometryCollections has the following behaviours:
//  1. For a mixed-dimension (heterogeneous) collection a point may lie on two
//     geometry elements with different dimensions. In this case the location on
//     the largest-dimension element is reported.
//  2. For a collection with overlapping or adjacent polygons, points on polygon
//     element boundaries may lie in the effective interior of the collection
//     geometry.
//
// Prepared mode is supported via cached spatial indexes.
//
// Supports specifying the BoundaryNodeRule to use for line endpoints.
type OperationRelateng_RelatePointLocator struct {
	geom           *Geom_Geometry
	isPrepared     bool
	boundaryRule   Algorithm_BoundaryNodeRule
	adjEdgeLocator *OperationRelateng_AdjacentEdgeLocator
	points         map[coord2DKey]bool
	lines          []*Geom_LineString
	polygons       []*Geom_Geometry
	polyLocator    []AlgorithmLocate_PointOnGeometryLocator
	lineBoundary   *OperationRelateng_LinearBoundary
	isEmpty        bool
}

// OperationRelateng_NewRelatePointLocator creates a new RelatePointLocator with
// the default boundary node rule.
func OperationRelateng_NewRelatePointLocator(geom *Geom_Geometry) *OperationRelateng_RelatePointLocator {
	return OperationRelateng_NewRelatePointLocatorWithOptions(geom, false, Algorithm_BoundaryNodeRule_OGC_SFS_BOUNDARY_RULE)
}

// OperationRelateng_NewRelatePointLocatorWithOptions creates a new
// RelatePointLocator with specified options.
func OperationRelateng_NewRelatePointLocatorWithOptions(geom *Geom_Geometry, isPrepared bool, bnRule Algorithm_BoundaryNodeRule) *OperationRelateng_RelatePointLocator {
	rpl := &OperationRelateng_RelatePointLocator{
		geom:         geom,
		isPrepared:   isPrepared,
		boundaryRule: bnRule,
	}
	rpl.init(geom)
	return rpl
}

func (rpl *OperationRelateng_RelatePointLocator) init(geom *Geom_Geometry) {
	// Cache empty status, since may be checked many times.
	rpl.isEmpty = geom.IsEmpty()
	rpl.extractElements(geom)

	if rpl.lines != nil {
		rpl.lineBoundary = OperationRelateng_NewLinearBoundary(rpl.lines, rpl.boundaryRule)
	}

	if rpl.polygons != nil {
		if rpl.isPrepared {
			rpl.polyLocator = make([]AlgorithmLocate_PointOnGeometryLocator, len(rpl.polygons))
		} else {
			rpl.polyLocator = make([]AlgorithmLocate_PointOnGeometryLocator, len(rpl.polygons))
		}
	}
}

// HasBoundary reports whether the geometry has a boundary.
func (rpl *OperationRelateng_RelatePointLocator) HasBoundary() bool {
	return rpl.lineBoundary.HasBoundary()
}

func (rpl *OperationRelateng_RelatePointLocator) extractElements(geom *Geom_Geometry) {
	if geom.IsEmpty() {
		return
	}

	if java.InstanceOf[*Geom_Point](geom) {
		rpl.addPoint(java.Cast[*Geom_Point](geom))
	} else if java.InstanceOf[*Geom_LineString](geom) {
		rpl.addLine(java.Cast[*Geom_LineString](geom))
	} else if java.InstanceOf[*Geom_Polygon](geom) ||
		java.InstanceOf[*Geom_MultiPolygon](geom) {
		rpl.addPolygonal(geom)
	} else if java.InstanceOf[*Geom_GeometryCollection](geom) {
		for i := 0; i < geom.GetNumGeometries(); i++ {
			g := geom.GetGeometryN(i)
			rpl.extractElements(g)
		}
	}
}

func (rpl *OperationRelateng_RelatePointLocator) addPoint(pt *Geom_Point) {
	if rpl.points == nil {
		rpl.points = make(map[coord2DKey]bool)
	}
	c := pt.GetCoordinate()
	key := coord2DKey{x: c.X, y: c.Y}
	rpl.points[key] = true
}

func (rpl *OperationRelateng_RelatePointLocator) addLine(line *Geom_LineString) {
	if rpl.lines == nil {
		rpl.lines = make([]*Geom_LineString, 0)
	}
	rpl.lines = append(rpl.lines, line)
}

func (rpl *OperationRelateng_RelatePointLocator) addPolygonal(polygonal *Geom_Geometry) {
	if rpl.polygons == nil {
		rpl.polygons = make([]*Geom_Geometry, 0)
	}
	rpl.polygons = append(rpl.polygons, polygonal)
}

// Locate returns the location of the point.
func (rpl *OperationRelateng_RelatePointLocator) Locate(p *Geom_Coordinate) int {
	return OperationRelateng_DimensionLocation_Location(rpl.LocateWithDim(p))
}

// LocateLineEndWithDim locates a line endpoint, as a DimensionLocation. In a
// mixed-dim GC, the line end point may also lie in an area. In this case the
// area location is reported. Otherwise, the dimLoc is either LINE_BOUNDARY or
// LINE_INTERIOR, depending on the endpoint valence and the BoundaryNodeRule in
// place.
func (rpl *OperationRelateng_RelatePointLocator) LocateLineEndWithDim(p *Geom_Coordinate) int {
	// If a GC with areas, check for point on area.
	if rpl.polygons != nil {
		locPoly := rpl.locateOnPolygons(p, false, nil)
		if locPoly != Geom_Location_Exterior {
			return OperationRelateng_DimensionLocation_LocationArea(locPoly)
		}
	}
	// Not in area, so return line end location.
	if rpl.lineBoundary.IsBoundary(p) {
		return OperationRelateng_DimensionLocation_LINE_BOUNDARY
	}
	return OperationRelateng_DimensionLocation_LINE_INTERIOR
}

// LocateNode locates a point which is known to be a node of the geometry (i.e.
// a vertex or on an edge).
func (rpl *OperationRelateng_RelatePointLocator) LocateNode(p *Geom_Coordinate, parentPolygonal *Geom_Geometry) int {
	return OperationRelateng_DimensionLocation_Location(rpl.LocateNodeWithDim(p, parentPolygonal))
}

// LocateNodeWithDim locates a point which is known to be a node of the
// geometry, as a DimensionLocation.
func (rpl *OperationRelateng_RelatePointLocator) LocateNodeWithDim(p *Geom_Coordinate, parentPolygonal *Geom_Geometry) int {
	return rpl.locateWithDim(p, true, parentPolygonal)
}

// LocateWithDim computes the topological location of a single point in a
// Geometry, as well as the dimension of the geometry element the point is
// located in (if not in the Exterior). It handles both single-element and
// multi-element Geometries. The algorithm for multi-part Geometries takes into
// account the SFS Boundary Determination Rule.
func (rpl *OperationRelateng_RelatePointLocator) LocateWithDim(p *Geom_Coordinate) int {
	return rpl.locateWithDim(p, false, nil)
}

func (rpl *OperationRelateng_RelatePointLocator) locateWithDim(p *Geom_Coordinate, isNode bool, parentPolygonal *Geom_Geometry) int {
	if rpl.isEmpty {
		return OperationRelateng_DimensionLocation_EXTERIOR
	}

	// In a polygonal geometry a node must be on the boundary. (This is not the
	// case for a mixed collection, since the node may be in the interior of a
	// polygon.)
	if isNode && (java.InstanceOf[*Geom_Polygon](rpl.geom) ||
		java.InstanceOf[*Geom_MultiPolygon](rpl.geom)) {
		return OperationRelateng_DimensionLocation_AREA_BOUNDARY
	}

	return rpl.computeDimLocation(p, isNode, parentPolygonal)
}

func (rpl *OperationRelateng_RelatePointLocator) computeDimLocation(p *Geom_Coordinate, isNode bool, parentPolygonal *Geom_Geometry) int {
	// Check dimensions in order of precedence.
	if rpl.polygons != nil {
		locPoly := rpl.locateOnPolygons(p, isNode, parentPolygonal)
		if locPoly != Geom_Location_Exterior {
			return OperationRelateng_DimensionLocation_LocationArea(locPoly)
		}
	}
	if rpl.lines != nil {
		locLine := rpl.locateOnLines(p, isNode)
		if locLine != Geom_Location_Exterior {
			return OperationRelateng_DimensionLocation_LocationLine(locLine)
		}
	}
	if rpl.points != nil {
		locPt := rpl.locateOnPoints(p)
		if locPt != Geom_Location_Exterior {
			return OperationRelateng_DimensionLocation_LocationPoint(locPt)
		}
	}
	return OperationRelateng_DimensionLocation_EXTERIOR
}

func (rpl *OperationRelateng_RelatePointLocator) locateOnPoints(p *Geom_Coordinate) int {
	key := coord2DKey{x: p.X, y: p.Y}
	if rpl.points[key] {
		return Geom_Location_Interior
	}
	return Geom_Location_Exterior
}

func (rpl *OperationRelateng_RelatePointLocator) locateOnLines(p *Geom_Coordinate, isNode bool) int {
	if rpl.lineBoundary != nil && rpl.lineBoundary.IsBoundary(p) {
		return Geom_Location_Boundary
	}
	// Must be on line, in interior.
	if isNode {
		return Geom_Location_Interior
	}

	// TODO: index the lines.
	for _, line := range rpl.lines {
		// Have to check every line, since any/all may contain point.
		loc := rpl.locateOnLine(p, isNode, line)
		if loc != Geom_Location_Exterior {
			return loc
		}
		// TODO: minor optimization - some BoundaryNodeRules can short-circuit.
	}
	return Geom_Location_Exterior
}

func (rpl *OperationRelateng_RelatePointLocator) locateOnLine(p *Geom_Coordinate, isNode bool, l *Geom_LineString) int {
	// Bounding-box check.
	if !l.GetEnvelopeInternal().IntersectsCoordinate(p) {
		return Geom_Location_Exterior
	}

	seq := l.GetCoordinateSequence()
	if Algorithm_PointLocation_IsOnLineSeq(p, seq) {
		return Geom_Location_Interior
	}
	return Geom_Location_Exterior
}

func (rpl *OperationRelateng_RelatePointLocator) locateOnPolygons(p *Geom_Coordinate, isNode bool, parentPolygonal *Geom_Geometry) int {
	numBdy := 0
	// TODO: use a spatial index on the polygons.
	for i := range rpl.polygons {
		loc := rpl.locateOnPolygonal(p, isNode, parentPolygonal, i)
		if loc == Geom_Location_Interior {
			return Geom_Location_Interior
		}
		if loc == Geom_Location_Boundary {
			numBdy++
		}
	}
	if numBdy == 1 {
		return Geom_Location_Boundary
	}
	// Check for point lying on adjacent boundaries.
	if numBdy > 1 {
		if rpl.adjEdgeLocator == nil {
			rpl.adjEdgeLocator = OperationRelateng_NewAdjacentEdgeLocator(rpl.geom)
		}
		return rpl.adjEdgeLocator.Locate(p)
	}
	return Geom_Location_Exterior
}

func (rpl *OperationRelateng_RelatePointLocator) locateOnPolygonal(p *Geom_Coordinate, isNode bool, parentPolygonal *Geom_Geometry, index int) int {
	polygonal := rpl.polygons[index]
	if isNode && parentPolygonal == polygonal {
		return Geom_Location_Boundary
	}
	locator := rpl.getLocator(index)
	return locator.Locate(p)
}

func (rpl *OperationRelateng_RelatePointLocator) getLocator(index int) AlgorithmLocate_PointOnGeometryLocator {
	locator := rpl.polyLocator[index]
	if locator == nil {
		polygonal := rpl.polygons[index]
		if rpl.isPrepared {
			locator = AlgorithmLocate_NewIndexedPointInAreaLocator(polygonal)
		} else {
			locator = AlgorithmLocate_NewSimplePointInAreaLocator(polygonal)
		}
		rpl.polyLocator[index] = locator
	}
	return locator
}
