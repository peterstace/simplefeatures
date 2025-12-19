package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationValid_IsSimpleOp tests whether a Geometry is simple as defined by
// the OGC SFS specification.
//
// Simplicity is defined for each Geometry type as follows:
//   - Point geometries are simple.
//   - MultiPoint geometries are simple if every point is unique.
//   - LineString geometries are simple if they do not self-intersect at interior
//     points (i.e. points other than the endpoints). Closed linestrings which
//     intersect only at their endpoints are simple (i.e. valid LinearRings).
//   - MultiLineString geometries are simple if their elements are simple and
//     they intersect only at points which are boundary points of both elements.
//     (The notion of boundary points can be user-specified - see below).
//   - Polygonal geometries have no definition of simplicity. The isSimple code
//     checks if all polygon rings are simple. (Note: this means that isSimple
//     cannot be used to test for all self-intersections in Polygons. In order to
//     check if a Polygonal geometry has self-intersections, use Geometry.IsValid()).
//   - GeometryCollection geometries are simple if all their elements are simple.
//   - Empty geometries are simple.
//
// For Lineal geometries the evaluation of simplicity can be customized by
// supplying a BoundaryNodeRule to define how boundary points are determined.
// The default is the SFS-standard MOD2_BOUNDARY_RULE.
//
// Note that under the Mod-2 rule, closed LineStrings (rings) have no boundary.
// This means that an intersection at the endpoints of two closed LineStrings
// makes the geometry non-simple. If it is required to test whether a set of
// LineStrings touch only at their endpoints, use ENDPOINT_BOUNDARY_RULE.
// For example, this can be used to validate that a collection of lines form a
// topologically valid linear network.
//
// By default this class finds a single non-simple location. To find all
// non-simple locations, set SetFindAllLocations(true) before calling IsSimple(),
// and retrieve the locations via GetNonSimpleLocations().
type OperationValid_IsSimpleOp struct {
	inputGeom                   *Geom_Geometry
	isClosedEndpointsInInterior bool
	isFindAllLocations          bool
	isSimple                    bool
	nonSimplePts                []*Geom_Coordinate
}

// OperationValid_IsSimpleOp_IsSimple tests whether a geometry is simple.
func OperationValid_IsSimpleOp_IsSimple(geom *Geom_Geometry) bool {
	op := OperationValid_NewIsSimpleOp(geom)
	return op.IsSimple()
}

// OperationValid_IsSimpleOp_GetNonSimpleLocation gets a non-simple location in a
// geometry, if any.
func OperationValid_IsSimpleOp_GetNonSimpleLocation(geom *Geom_Geometry) *Geom_Coordinate {
	op := OperationValid_NewIsSimpleOp(geom)
	return op.GetNonSimpleLocation()
}

// OperationValid_NewIsSimpleOp creates a simplicity checker using the default
// SFS Mod-2 Boundary Node Rule.
func OperationValid_NewIsSimpleOp(geom *Geom_Geometry) *OperationValid_IsSimpleOp {
	return OperationValid_NewIsSimpleOpWithBoundaryNodeRule(geom, Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE)
}

// OperationValid_NewIsSimpleOpWithBoundaryNodeRule creates a simplicity checker
// using a given BoundaryNodeRule.
func OperationValid_NewIsSimpleOpWithBoundaryNodeRule(geom *Geom_Geometry, boundaryNodeRule Algorithm_BoundaryNodeRule) *OperationValid_IsSimpleOp {
	return &OperationValid_IsSimpleOp{
		inputGeom:                   geom,
		isClosedEndpointsInInterior: !boundaryNodeRule.IsInBoundary(2),
	}
}

// SetFindAllLocations sets whether all non-simple intersection points will be
// found.
func (op *OperationValid_IsSimpleOp) SetFindAllLocations(isFindAll bool) {
	op.isFindAllLocations = isFindAll
}

// IsSimple tests whether the geometry is simple.
func (op *OperationValid_IsSimpleOp) IsSimple() bool {
	op.compute()
	return op.isSimple
}

// GetNonSimpleLocation gets the coordinate for a location where the geometry
// fails to be simple (i.e. where it has a non-boundary self-intersection).
func (op *OperationValid_IsSimpleOp) GetNonSimpleLocation() *Geom_Coordinate {
	op.compute()
	if len(op.nonSimplePts) == 0 {
		return nil
	}
	return op.nonSimplePts[0]
}

// GetNonSimpleLocations gets all non-simple intersection locations.
func (op *OperationValid_IsSimpleOp) GetNonSimpleLocations() []*Geom_Coordinate {
	op.compute()
	return op.nonSimplePts
}

func (op *OperationValid_IsSimpleOp) compute() {
	if op.nonSimplePts != nil {
		return
	}
	op.nonSimplePts = make([]*Geom_Coordinate, 0)
	op.isSimple = op.computeSimple(op.inputGeom)
}

func (op *OperationValid_IsSimpleOp) computeSimple(geom *Geom_Geometry) bool {
	if geom.IsEmpty() {
		return true
	}
	if java.InstanceOf[*Geom_Point](geom) {
		return true
	}
	if java.InstanceOf[*Geom_LineString](geom) {
		return op.isSimpleLinearGeometry(geom)
	}
	if java.InstanceOf[*Geom_MultiLineString](geom) {
		return op.isSimpleLinearGeometry(geom)
	}
	if java.InstanceOf[*Geom_MultiPoint](geom) {
		return op.isSimpleMultiPoint(java.Cast[*Geom_MultiPoint](geom))
	}
	if java.InstanceOf[Geom_Polygonal](geom) {
		return op.isSimplePolygonal(geom)
	}
	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		return op.isSimpleGeometryCollection(geom)
	}
	// All other geometry types are simple by definition.
	return true
}

// coordKey2D is used as a map key for coordinate comparison using only X and Y.
// This matches Java's Coordinate.hashCode() which only uses X and Y.
type coordKey2D struct {
	x, y float64
}

func (op *OperationValid_IsSimpleOp) isSimpleMultiPoint(mp *Geom_MultiPoint) bool {
	if mp.IsEmpty() {
		return true
	}
	isSimple := true
	points := make(map[coordKey2D]bool)
	for i := 0; i < mp.GetNumGeometries(); i++ {
		pt := java.Cast[*Geom_Point](mp.GetGeometryN(i))
		p := pt.GetCoordinate()
		if p == nil {
			continue
		}
		key := coordKey2D{p.X, p.Y}
		if points[key] {
			op.nonSimplePts = append(op.nonSimplePts, p)
			isSimple = false
			if !op.isFindAllLocations {
				break
			}
		} else {
			points[key] = true
		}
	}
	return isSimple
}

// isSimplePolygonal computes simplicity for polygonal geometries.
// Polygonal geometries are simple if and only if all of their component rings
// are simple.
func (op *OperationValid_IsSimpleOp) isSimplePolygonal(geom *Geom_Geometry) bool {
	isSimple := true
	rings := GeomUtil_LinearComponentExtracter_GetLines(geom)
	for _, ring := range rings {
		if !op.isSimpleLinearGeometry(ring.Geom_Geometry) {
			isSimple = false
			if !op.isFindAllLocations {
				break
			}
		}
	}
	return isSimple
}

// isSimpleGeometryCollection tests simplicity of a GeometryCollection.
// Semantics: simple iff all components are simple.
func (op *OperationValid_IsSimpleOp) isSimpleGeometryCollection(geom *Geom_Geometry) bool {
	isSimple := true
	for i := 0; i < geom.GetNumGeometries(); i++ {
		comp := geom.GetGeometryN(i)
		if !op.computeSimple(comp) {
			isSimple = false
			if !op.isFindAllLocations {
				break
			}
		}
	}
	return isSimple
}

func (op *OperationValid_IsSimpleOp) isSimpleLinearGeometry(geom *Geom_Geometry) bool {
	if geom.IsEmpty() {
		return true
	}
	segStrings := operationValid_extractSegmentStrings(geom)
	segInt := operationValid_NewNonSimpleIntersectionFinder(op.isClosedEndpointsInInterior, op.isFindAllLocations, &op.nonSimplePts)
	noder := Noding_NewMCIndexNoder()
	noder.SetSegmentIntersector(segInt)
	noder.ComputeNodes(segStrings)
	if segInt.hasIntersection() {
		return false
	}
	return true
}

func operationValid_extractSegmentStrings(geom *Geom_Geometry) []Noding_SegmentString {
	segStrings := make([]Noding_SegmentString, 0)
	for i := 0; i < geom.GetNumGeometries(); i++ {
		line := java.Cast[*Geom_LineString](geom.GetGeometryN(i))
		trimPts := operationValid_trimRepeatedPoints(line.GetCoordinates())
		if trimPts != nil {
			ss := Noding_NewBasicSegmentString(trimPts, nil)
			segStrings = append(segStrings, ss)
		}
	}
	return segStrings
}

func operationValid_trimRepeatedPoints(pts []*Geom_Coordinate) []*Geom_Coordinate {
	if len(pts) <= 2 {
		return pts
	}

	length := len(pts)
	hasRepeatedStart := pts[0].Equals2D(pts[1])
	hasRepeatedEnd := pts[length-1].Equals2D(pts[length-2])
	if !hasRepeatedStart && !hasRepeatedEnd {
		return pts
	}

	// Trim ends.
	startIndex := 0
	startPt := pts[0]
	for startIndex < length-1 && startPt.Equals2D(pts[startIndex+1]) {
		startIndex++
	}
	endIndex := length - 1
	endPt := pts[endIndex]
	for endIndex > 0 && endPt.Equals2D(pts[endIndex-1]) {
		endIndex--
	}
	// Are all points identical?
	if endIndex-startIndex < 1 {
		return nil
	}
	trimPts := Geom_CoordinateArrays_Extract(pts, startIndex, endIndex)
	return trimPts
}

// operationValid_NonSimpleIntersectionFinder is the intersection finder for
// IsSimpleOp.
type operationValid_NonSimpleIntersectionFinder struct {
	isClosedEndpointsInInterior bool
	isFindAll                   bool
	li                          *Algorithm_LineIntersector
	intersectionPts             *[]*Geom_Coordinate
}

var _ Noding_SegmentIntersector = (*operationValid_NonSimpleIntersectionFinder)(nil)

// IsNoding_SegmentIntersector is a marker method for interface identification.
func (f *operationValid_NonSimpleIntersectionFinder) IsNoding_SegmentIntersector() {}

func operationValid_NewNonSimpleIntersectionFinder(isClosedEndpointsInInterior, isFindAll bool, intersectionPts *[]*Geom_Coordinate) *operationValid_NonSimpleIntersectionFinder {
	return &operationValid_NonSimpleIntersectionFinder{
		isClosedEndpointsInInterior: isClosedEndpointsInInterior,
		isFindAll:                   isFindAll,
		li:                          Algorithm_NewRobustLineIntersector().Algorithm_LineIntersector,
		intersectionPts:             intersectionPts,
	}
}

// hasIntersection tests whether an intersection was found.
func (f *operationValid_NonSimpleIntersectionFinder) hasIntersection() bool {
	return len(*f.intersectionPts) > 0
}

// ProcessIntersections processes intersections between two segment strings.
func (f *operationValid_NonSimpleIntersectionFinder) ProcessIntersections(ss0 Noding_SegmentString, segIndex0 int, ss1 Noding_SegmentString, segIndex1 int) {
	// Don't test a segment with itself.
	isSameSegString := ss0 == ss1
	isSameSegment := isSameSegString && segIndex0 == segIndex1
	if isSameSegment {
		return
	}

	hasInt := f.findIntersection(ss0, segIndex0, ss1, segIndex1)

	if hasInt {
		// Found an intersection!
		*f.intersectionPts = append(*f.intersectionPts, f.li.GetIntersection(0))
	}
}

func (f *operationValid_NonSimpleIntersectionFinder) findIntersection(ss0 Noding_SegmentString, segIndex0 int, ss1 Noding_SegmentString, segIndex1 int) bool {
	p00 := ss0.GetCoordinate(segIndex0)
	p01 := ss0.GetCoordinate(segIndex0 + 1)
	p10 := ss1.GetCoordinate(segIndex1)
	p11 := ss1.GetCoordinate(segIndex1 + 1)

	f.li.ComputeIntersection(p00, p01, p10, p11)
	if !f.li.HasIntersection() {
		return false
	}

	// Check for an intersection in the interior of a segment.
	hasInteriorInt := f.li.IsInteriorIntersection()
	if hasInteriorInt {
		return true
	}

	// Check for equal segments (which will produce two intersection points).
	// These also intersect in interior points, so are non-simple.
	// (This is not triggered by zero-length segments, since they are filtered
	// out by the MC index).
	hasEqualSegments := f.li.GetIntersectionNum() >= 2
	if hasEqualSegments {
		return true
	}

	// Following tests assume non-adjacent segments.
	isSameSegString := ss0 == ss1
	isAdjacentSegment := isSameSegString && int(math.Abs(float64(segIndex1-segIndex0))) <= 1
	if isAdjacentSegment {
		return false
	}

	// At this point there is a single intersection point which is a vertex in
	// each segString. Classify them as endpoints or interior.
	isIntersectionEndpt0 := operationValid_isIntersectionEndpoint(ss0, segIndex0, f.li, 0)
	isIntersectionEndpt1 := operationValid_isIntersectionEndpoint(ss1, segIndex1, f.li, 1)

	hasInteriorVertexInt := !(isIntersectionEndpt0 && isIntersectionEndpt1)
	if hasInteriorVertexInt {
		return true
	}

	// Both intersection vertices must be endpoints.
	// Final check is if one or both of them is interior due to being endpoint
	// of a closed ring. This only applies to different lines (which avoids
	// reporting ring endpoints).
	if f.isClosedEndpointsInInterior && !isSameSegString {
		hasInteriorEndpointInt := ss0.IsClosed() || ss1.IsClosed()
		if hasInteriorEndpointInt {
			return true
		}
	}
	return false
}

// operationValid_isIntersectionEndpoint tests whether an intersection vertex is
// an endpoint of a segment string.
func operationValid_isIntersectionEndpoint(ss Noding_SegmentString, ssIndex int, li *Algorithm_LineIntersector, liSegmentIndex int) bool {
	vertexIndex := operationValid_intersectionVertexIndex(li, liSegmentIndex)
	// If the vertex is the first one of the segment, check if it is the start
	// endpoint. Otherwise check if it is the end endpoint.
	if vertexIndex == 0 {
		return ssIndex == 0
	}
	return ssIndex+2 == ss.Size()
}

// operationValid_intersectionVertexIndex finds the vertex index in a segment of
// an intersection which is known to be a vertex.
func operationValid_intersectionVertexIndex(li *Algorithm_LineIntersector, segmentIndex int) int {
	intPt := li.GetIntersection(0)
	endPt0 := li.GetEndpoint(segmentIndex, 0)
	if intPt.Equals2D(endPt0) {
		return 0
	}
	return 1
}

// IsDone tests whether processing should stop.
func (f *operationValid_NonSimpleIntersectionFinder) IsDone() bool {
	if f.isFindAll {
		return false
	}
	return len(*f.intersectionPts) > 0
}
