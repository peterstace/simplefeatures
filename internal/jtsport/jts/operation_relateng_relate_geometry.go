package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

const (
	OperationRelateng_RelateGeometry_GEOM_A = true
	OperationRelateng_RelateGeometry_GEOM_B = false
)

// OperationRelateng_RelateGeometry_Name returns "A" or "B" based on the
// isA parameter.
func OperationRelateng_RelateGeometry_Name(isA bool) string {
	if isA {
		return "A"
	}
	return "B"
}

// OperationRelateng_RelateGeometry wraps a geometry with analysis capabilities
// for RelateNG.
type OperationRelateng_RelateGeometry struct {
	geom          *Geom_Geometry
	isPrepared    bool
	geomEnv       *Geom_Envelope
	geomDim       int
	uniquePoints  map[coord2DKey]bool
	boundaryRule  Algorithm_BoundaryNodeRule
	locator       *OperationRelateng_RelatePointLocator
	elementId     int
	hasPoints     bool
	hasLines      bool
	hasAreas      bool
	isLineZeroLen bool
	isGeomEmpty   bool
}

// OperationRelateng_NewRelateGeometry creates a new RelateGeometry with the
// default boundary node rule.
func OperationRelateng_NewRelateGeometry(input *Geom_Geometry) *OperationRelateng_RelateGeometry {
	return OperationRelateng_NewRelateGeometryWithRule(input, Algorithm_BoundaryNodeRule_OGC_SFS_BOUNDARY_RULE)
}

// OperationRelateng_NewRelateGeometryWithRule creates a new RelateGeometry with
// the specified boundary node rule.
func OperationRelateng_NewRelateGeometryWithRule(input *Geom_Geometry, bnRule Algorithm_BoundaryNodeRule) *OperationRelateng_RelateGeometry {
	return OperationRelateng_NewRelateGeometryWithOptions(input, false, bnRule)
}

// OperationRelateng_NewRelateGeometryWithOptions creates a new RelateGeometry
// with the specified options.
func OperationRelateng_NewRelateGeometryWithOptions(input *Geom_Geometry, isPrepared bool, bnRule Algorithm_BoundaryNodeRule) *OperationRelateng_RelateGeometry {
	rg := &OperationRelateng_RelateGeometry{
		geom:         input,
		geomEnv:      input.GetEnvelopeInternal(),
		isPrepared:   isPrepared,
		boundaryRule: bnRule,
		geomDim:      Geom_Dimension_False,
	}
	// Cache geometry metadata.
	rg.isGeomEmpty = input.IsEmpty()
	rg.geomDim = input.GetDimension()
	rg.analyzeDimensions()
	rg.isLineZeroLen = rg.isZeroLengthLine(input)
	return rg
}

func (rg *OperationRelateng_RelateGeometry) isZeroLengthLine(geom *Geom_Geometry) bool {
	// Avoid expensive zero-length calculation if not linear.
	if rg.GetDimension() != Geom_Dimension_L {
		return false
	}
	return operationRelateng_RelateGeometry_isZeroLength(geom)
}

func (rg *OperationRelateng_RelateGeometry) analyzeDimensions() {
	if rg.isGeomEmpty {
		return
	}
	if java.InstanceOf[*Geom_Point](rg.geom) ||
		java.InstanceOf[*Geom_MultiPoint](rg.geom) {
		rg.hasPoints = true
		rg.geomDim = Geom_Dimension_P
		return
	}
	if java.InstanceOf[*Geom_LineString](rg.geom) ||
		java.InstanceOf[*Geom_MultiLineString](rg.geom) {
		rg.hasLines = true
		rg.geomDim = Geom_Dimension_L
		return
	}
	if java.InstanceOf[*Geom_Polygon](rg.geom) ||
		java.InstanceOf[*Geom_MultiPolygon](rg.geom) {
		rg.hasAreas = true
		rg.geomDim = Geom_Dimension_A
		return
	}
	// Analyze a (possibly mixed type) collection.
	geomi := Geom_NewGeometryCollectionIterator(rg.geom)
	for geomi.HasNext() {
		elem := geomi.Next()
		if elem.IsEmpty() {
			continue
		}
		if java.InstanceOf[*Geom_Point](elem) {
			rg.hasPoints = true
			if rg.geomDim < Geom_Dimension_P {
				rg.geomDim = Geom_Dimension_P
			}
		}
		if java.InstanceOf[*Geom_LineString](elem) {
			rg.hasLines = true
			if rg.geomDim < Geom_Dimension_L {
				rg.geomDim = Geom_Dimension_L
			}
		}
		if java.InstanceOf[*Geom_Polygon](elem) {
			rg.hasAreas = true
			if rg.geomDim < Geom_Dimension_A {
				rg.geomDim = Geom_Dimension_A
			}
		}
	}
}

// operationRelateng_RelateGeometry_isZeroLength tests if all geometry linear
// elements are zero-length.
func operationRelateng_RelateGeometry_isZeroLength(geom *Geom_Geometry) bool {
	geomi := Geom_NewGeometryCollectionIterator(geom)
	for geomi.HasNext() {
		elem := geomi.Next()
		if java.InstanceOf[*Geom_LineString](elem) {
			ls := java.Cast[*Geom_LineString](elem)
			if !operationRelateng_RelateGeometry_isZeroLengthLine(ls) {
				return false
			}
		}
	}
	return true
}

func operationRelateng_RelateGeometry_isZeroLengthLine(line *Geom_LineString) bool {
	if line.GetNumPoints() >= 2 {
		p0 := line.GetCoordinateN(0)
		for i := 0; i < line.GetNumPoints(); i++ {
			pi := line.GetCoordinateN(i)
			// Most non-zero-len lines will trigger this right away.
			if !p0.Equals2D(pi) {
				return false
			}
		}
	}
	return true
}

// GetGeometry returns the wrapped geometry.
func (rg *OperationRelateng_RelateGeometry) GetGeometry() *Geom_Geometry {
	return rg.geom
}

// IsPrepared returns true if this geometry is in prepared mode.
func (rg *OperationRelateng_RelateGeometry) IsPrepared() bool {
	return rg.isPrepared
}

// GetEnvelope returns the envelope of the geometry.
func (rg *OperationRelateng_RelateGeometry) GetEnvelope() *Geom_Envelope {
	return rg.geomEnv
}

// GetDimension returns the dimension of the geometry.
func (rg *OperationRelateng_RelateGeometry) GetDimension() int {
	return rg.geomDim
}

// HasDimension tests if the geometry has the specified dimension.
func (rg *OperationRelateng_RelateGeometry) HasDimension(dim int) bool {
	switch dim {
	case Geom_Dimension_P:
		return rg.hasPoints
	case Geom_Dimension_L:
		return rg.hasLines
	case Geom_Dimension_A:
		return rg.hasAreas
	}
	return false
}

// GetDimensionReal gets the actual non-empty dimension of the geometry.
// Zero-length LineStrings are treated as Points.
func (rg *OperationRelateng_RelateGeometry) GetDimensionReal() int {
	if rg.isGeomEmpty {
		return Geom_Dimension_False
	}
	if rg.GetDimension() == 1 && rg.isLineZeroLen {
		return Geom_Dimension_P
	}
	if rg.hasAreas {
		return Geom_Dimension_A
	}
	if rg.hasLines {
		return Geom_Dimension_L
	}
	return Geom_Dimension_P
}

// HasEdges returns true if the geometry has linear or areal components.
func (rg *OperationRelateng_RelateGeometry) HasEdges() bool {
	return rg.hasLines || rg.hasAreas
}

func (rg *OperationRelateng_RelateGeometry) getLocator() *OperationRelateng_RelatePointLocator {
	if rg.locator == nil {
		rg.locator = OperationRelateng_NewRelatePointLocatorWithOptions(rg.geom, rg.isPrepared, rg.boundaryRule)
	}
	return rg.locator
}

// IsNodeInArea tests if a node point is in the area of this geometry.
func (rg *OperationRelateng_RelateGeometry) IsNodeInArea(nodePt *Geom_Coordinate, parentPolygonal *Geom_Geometry) bool {
	loc := rg.getLocator().LocateNodeWithDim(nodePt, parentPolygonal)
	return loc == OperationRelateng_DimensionLocation_AREA_INTERIOR
}

// LocateLineEndWithDim locates a line endpoint with dimension information.
func (rg *OperationRelateng_RelateGeometry) LocateLineEndWithDim(p *Geom_Coordinate) int {
	return rg.getLocator().LocateLineEndWithDim(p)
}

// LocateAreaVertex locates a vertex of a polygon.
func (rg *OperationRelateng_RelateGeometry) LocateAreaVertex(pt *Geom_Coordinate) int {
	return rg.LocateNode(pt, nil)
}

// LocateNode locates a point which is known to be a node of the geometry.
func (rg *OperationRelateng_RelateGeometry) LocateNode(pt *Geom_Coordinate, parentPolygonal *Geom_Geometry) int {
	return rg.getLocator().LocateNode(pt, parentPolygonal)
}

// LocateWithDim computes the topological location and dimension of a point.
func (rg *OperationRelateng_RelateGeometry) LocateWithDim(pt *Geom_Coordinate) int {
	return rg.getLocator().LocateWithDim(pt)
}

// IsSelfNodingRequired indicates whether the geometry requires self-noding for
// correct evaluation of specific spatial predicates.
func (rg *OperationRelateng_RelateGeometry) IsSelfNodingRequired() bool {
	if java.InstanceOf[*Geom_Point](rg.geom) ||
		java.InstanceOf[*Geom_MultiPoint](rg.geom) ||
		java.InstanceOf[*Geom_Polygon](rg.geom) ||
		java.InstanceOf[*Geom_MultiPolygon](rg.geom) {
		return false
	}
	// A GC with a single polygon does not need noding.
	if rg.hasAreas && rg.geom.GetNumGeometries() == 1 {
		return false
	}
	return true
}

// IsPolygonal tests whether the geometry has polygonal topology.
func (rg *OperationRelateng_RelateGeometry) IsPolygonal() bool {
	return java.InstanceOf[*Geom_Polygon](rg.geom) ||
		java.InstanceOf[*Geom_MultiPolygon](rg.geom)
}

// IsEmpty returns true if the geometry is empty.
func (rg *OperationRelateng_RelateGeometry) IsEmpty() bool {
	return rg.isGeomEmpty
}

// HasBoundary reports whether the geometry has a boundary.
func (rg *OperationRelateng_RelateGeometry) HasBoundary() bool {
	return rg.getLocator().HasBoundary()
}

// GetUniquePoints returns the set of unique points in the geometry.
// The map uses 2D coordinates (X, Y only) as keys since that's how Java's
// Coordinate.equals() and hashCode() work.
func (rg *OperationRelateng_RelateGeometry) GetUniquePoints() map[coord2DKey]bool {
	// Will be re-used in prepared mode.
	if rg.uniquePoints == nil {
		rg.uniquePoints = rg.createUniquePoints()
	}
	return rg.uniquePoints
}

func (rg *OperationRelateng_RelateGeometry) createUniquePoints() map[coord2DKey]bool {
	// Only called on P geometries.
	pts := GeomUtil_ComponentCoordinateExtracter_GetCoordinates(rg.geom)
	set := make(map[coord2DKey]bool)
	for _, pt := range pts {
		key := coord2DKey{x: pt.X, y: pt.Y}
		set[key] = true
	}
	return set
}

// GetEffectivePoints returns the points not covered by another element.
func (rg *OperationRelateng_RelateGeometry) GetEffectivePoints() []*Geom_Point {
	ptListAll := GeomUtil_PointExtracter_GetPoints(rg.geom)

	if rg.GetDimensionReal() <= Geom_Dimension_P {
		return ptListAll
	}

	// Only return Points not covered by another element.
	var ptList []*Geom_Point
	for _, p := range ptListAll {
		if p.IsEmpty() {
			continue
		}
		locDim := rg.LocateWithDim(p.GetCoordinate())
		if OperationRelateng_DimensionLocation_Dimension(locDim) == Geom_Dimension_P {
			ptList = append(ptList, p)
		}
	}
	return ptList
}

// ExtractSegmentStrings extracts RelateSegmentStrings from the geometry which
// intersect a given envelope. If the envelope is nil all edges are extracted.
func (rg *OperationRelateng_RelateGeometry) ExtractSegmentStrings(isA bool, env *Geom_Envelope) []*OperationRelateng_RelateSegmentString {
	var segStrings []*OperationRelateng_RelateSegmentString
	rg.extractSegmentStrings(isA, env, rg.geom, &segStrings)
	return segStrings
}

func (rg *OperationRelateng_RelateGeometry) extractSegmentStrings(isA bool, env *Geom_Envelope, geom *Geom_Geometry, segStrings *[]*OperationRelateng_RelateSegmentString) {
	// Record if parent is MultiPolygon.
	// Note: parentPolygonal may be nil if geom is not a MultiPolygon, which is handled later.
	var parentPolygonal *Geom_MultiPolygon
	if java.InstanceOf[*Geom_MultiPolygon](geom) {
		parentPolygonal = java.Cast[*Geom_MultiPolygon](geom)
	}

	for i := 0; i < geom.GetNumGeometries(); i++ {
		g := geom.GetGeometryN(i)
		if java.InstanceOf[*Geom_GeometryCollection](g) {
			rg.extractSegmentStrings(isA, env, g, segStrings)
		} else {
			rg.extractSegmentStringsFromAtomic(isA, g, parentPolygonal, env, segStrings)
		}
	}
}

func (rg *OperationRelateng_RelateGeometry) extractSegmentStringsFromAtomic(isA bool, geom *Geom_Geometry, parentPolygonal *Geom_MultiPolygon, env *Geom_Envelope, segStrings *[]*OperationRelateng_RelateSegmentString) {
	if geom.IsEmpty() {
		return
	}
	doExtract := env == nil || env.IntersectsEnvelope(geom.GetEnvelopeInternal())
	if !doExtract {
		return
	}

	rg.elementId++
	if java.InstanceOf[*Geom_LineString](geom) {
		ss := OperationRelateng_RelateSegmentString_CreateLine(geom.GetCoordinates(), isA, rg.elementId, rg)
		*segStrings = append(*segStrings, ss)
	} else if java.InstanceOf[*Geom_Polygon](geom) {
		poly := java.Cast[*Geom_Polygon](geom)
		var parentPoly *Geom_Geometry
		if parentPolygonal != nil {
			parentPoly = parentPolygonal.Geom_Geometry
		} else {
			parentPoly = poly.Geom_Geometry
		}
		rg.extractRingToSegmentString(isA, poly.GetExteriorRing(), 0, env, parentPoly, segStrings)
		for i := 0; i < poly.GetNumInteriorRing(); i++ {
			rg.extractRingToSegmentString(isA, poly.GetInteriorRingN(i), i+1, env, parentPoly, segStrings)
		}
	}
}

func (rg *OperationRelateng_RelateGeometry) extractRingToSegmentString(isA bool, ring *Geom_LinearRing, ringId int, env *Geom_Envelope, parentPoly *Geom_Geometry, segStrings *[]*OperationRelateng_RelateSegmentString) {
	if ring.IsEmpty() {
		return
	}
	if env != nil && !env.IntersectsEnvelope(ring.GetEnvelopeInternal()) {
		return
	}

	// Orient the points if required.
	requireCW := ringId == 0
	pts := OperationRelateng_RelateGeometry_Orient(ring.GetCoordinates(), requireCW)
	ss := OperationRelateng_RelateSegmentString_CreateRing(pts, isA, rg.elementId, ringId, parentPoly, rg)
	*segStrings = append(*segStrings, ss)
}

// OperationRelateng_RelateGeometry_Orient orients a coordinate array to the
// specified orientation (CW or CCW).
func OperationRelateng_RelateGeometry_Orient(pts []*Geom_Coordinate, orientCW bool) []*Geom_Coordinate {
	isFlipped := orientCW == Algorithm_Orientation_IsCCW(pts)
	if isFlipped {
		// Clone and reverse.
		result := make([]*Geom_Coordinate, len(pts))
		copy(result, pts)
		Geom_CoordinateArrays_Reverse(result)
		return result
	}
	return pts
}

// String returns a string representation of the geometry.
func (rg *OperationRelateng_RelateGeometry) String() string {
	return rg.geom.String()
}
