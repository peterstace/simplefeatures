package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationDistance_DistanceOp_Distance computes the distance between the
// nearest points of two geometries.
func OperationDistance_DistanceOp_Distance(g0, g1 *Geom_Geometry) float64 {
	distOp := OperationDistance_NewDistanceOp(g0, g1)
	return distOp.Distance()
}

// OperationDistance_DistanceOp_IsWithinDistance tests whether two geometries
// lie within a given distance of each other.
func OperationDistance_DistanceOp_IsWithinDistance(g0, g1 *Geom_Geometry, distance float64) bool {
	// Check envelope distance for a short-circuit negative result.
	envDist := g0.GetEnvelopeInternal().Distance(g1.GetEnvelopeInternal())
	if envDist > distance {
		return false
	}

	// MD - could improve this further with a positive short-circuit based on envelope MinMaxDist

	distOp := OperationDistance_NewDistanceOpWithTerminate(g0, g1, distance)
	return distOp.Distance() <= distance
}

// OperationDistance_DistanceOp_NearestPoints computes the nearest points of
// two geometries. The points are presented in the same order as the input
// Geometries.
func OperationDistance_DistanceOp_NearestPoints(g0, g1 *Geom_Geometry) []*Geom_Coordinate {
	distOp := OperationDistance_NewDistanceOp(g0, g1)
	return distOp.NearestPoints()
}

// OperationDistance_DistanceOp_ClosestPoints computes the closest points of
// two geometries. The points are presented in the same order as the input
// Geometries.
// Deprecated: renamed to NearestPoints.
func OperationDistance_DistanceOp_ClosestPoints(g0, g1 *Geom_Geometry) []*Geom_Coordinate {
	distOp := OperationDistance_NewDistanceOp(g0, g1)
	return distOp.NearestPoints()
}

// OperationDistance_DistanceOp finds two points on two Geometries which lie
// within a given distance, or else are the nearest points on the geometries
// (in which case this also provides the distance between the geometries).
//
// The distance computation also finds a pair of points in the input geometries
// which have the minimum distance between them. If a point lies in the interior
// of a line segment, the coordinate computed is a close approximation to the
// exact point.
//
// Empty geometry collection components are ignored.
//
// The algorithms used are straightforward O(n^2) comparisons. This worst-case
// performance could be improved on by using Voronoi techniques or spatial
// indexes.
type OperationDistance_DistanceOp struct {
	// Input.
	geom              [2]*Geom_Geometry
	terminateDistance float64
	// Working.
	ptLocator           *Algorithm_PointLocator
	minDistanceLocation [2]*OperationDistance_GeometryLocation
	minDistance         float64
}

// OperationDistance_NewDistanceOp constructs a DistanceOp that computes the
// distance and nearest points between the two specified geometries.
func OperationDistance_NewDistanceOp(g0, g1 *Geom_Geometry) *OperationDistance_DistanceOp {
	return OperationDistance_NewDistanceOpWithTerminate(g0, g1, 0.0)
}

// OperationDistance_NewDistanceOpWithTerminate constructs a DistanceOp that
// computes the distance and nearest points between the two specified
// geometries.
func OperationDistance_NewDistanceOpWithTerminate(g0, g1 *Geom_Geometry, terminateDistance float64) *OperationDistance_DistanceOp {
	return &OperationDistance_DistanceOp{
		geom:              [2]*Geom_Geometry{g0, g1},
		terminateDistance: terminateDistance,
		ptLocator:         Algorithm_NewPointLocator(),
		minDistance:       math.MaxFloat64,
	}
}

// Distance reports the distance between the nearest points on the input
// geometries.
//
// Returns the distance between the geometries or 0 if either input geometry is
// empty. Panics if either input geometry is null.
func (op *OperationDistance_DistanceOp) Distance() float64 {
	if op.geom[0] == nil || op.geom[1] == nil {
		panic("null geometries are not supported")
	}
	if op.geom[0].IsEmpty() || op.geom[1].IsEmpty() {
		return 0.0
	}

	// Optimization for Point/Point case.
	if java.InstanceOf[*Geom_Point](op.geom[0]) && java.InstanceOf[*Geom_Point](op.geom[1]) {
		return op.geom[0].GetCoordinate().Distance(op.geom[1].GetCoordinate())
	}

	op.computeMinDistance()
	return op.minDistance
}

// NearestPoints reports the coordinates of the nearest points in the input
// geometries. The points are presented in the same order as the input
// Geometries.
func (op *OperationDistance_DistanceOp) NearestPoints() []*Geom_Coordinate {
	op.computeMinDistance()
	nearestPts := []*Geom_Coordinate{
		op.minDistanceLocation[0].GetCoordinate(),
		op.minDistanceLocation[1].GetCoordinate(),
	}
	return nearestPts
}

// ClosestPoints returns a pair of Coordinates of the nearest points.
// Deprecated: renamed to NearestPoints.
func (op *OperationDistance_DistanceOp) ClosestPoints() []*Geom_Coordinate {
	return op.NearestPoints()
}

// NearestLocations reports the locations of the nearest points in the input
// geometries. The locations are presented in the same order as the input
// Geometries.
func (op *OperationDistance_DistanceOp) NearestLocations() [2]*OperationDistance_GeometryLocation {
	op.computeMinDistance()
	return op.minDistanceLocation
}

// ClosestLocations returns a pair of GeometryLocations for the nearest points.
// Deprecated: renamed to NearestLocations.
func (op *OperationDistance_DistanceOp) ClosestLocations() [2]*OperationDistance_GeometryLocation {
	return op.NearestLocations()
}

func (op *OperationDistance_DistanceOp) updateMinDistance(locGeom [2]*OperationDistance_GeometryLocation, flip bool) {
	// If not set then don't update.
	if locGeom[0] == nil {
		return
	}

	if flip {
		op.minDistanceLocation[0] = locGeom[1]
		op.minDistanceLocation[1] = locGeom[0]
	} else {
		op.minDistanceLocation[0] = locGeom[0]
		op.minDistanceLocation[1] = locGeom[1]
	}
}

func (op *OperationDistance_DistanceOp) computeMinDistance() {
	// Only compute once!
	if op.minDistanceLocation[0] != nil {
		return
	}

	op.computeContainmentDistance()
	if op.minDistance <= op.terminateDistance {
		return
	}
	op.computeFacetDistance()
}

func (op *OperationDistance_DistanceOp) computeContainmentDistance() {
	var locPtPoly [2]*OperationDistance_GeometryLocation
	// Test if either geometry has a vertex inside the other.
	op.computeContainmentDistanceForIndex(0, &locPtPoly)
	if op.minDistance <= op.terminateDistance {
		return
	}
	op.computeContainmentDistanceForIndex(1, &locPtPoly)
}

func (op *OperationDistance_DistanceOp) computeContainmentDistanceForIndex(polyGeomIndex int, locPtPoly *[2]*OperationDistance_GeometryLocation) {
	polyGeom := op.geom[polyGeomIndex]
	// If no polygon then nothing to do.
	if polyGeom.GetDimension() < 2 {
		return
	}

	locationsIndex := 1 - polyGeomIndex
	polys := GeomUtil_PolygonExtracter_GetPolygons(polyGeom)
	if len(polys) > 0 {
		insideLocs := OperationDistance_ConnectedElementLocationFilter_GetLocations(op.geom[locationsIndex])
		op.computeContainmentDistanceLocsPolys(insideLocs, polys, locPtPoly)
		if op.minDistance <= op.terminateDistance {
			// This assignment is determined by the order of the args in the computeInside call above.
			op.minDistanceLocation[locationsIndex] = locPtPoly[0]
			op.minDistanceLocation[polyGeomIndex] = locPtPoly[1]
			return
		}
	}
}

func (op *OperationDistance_DistanceOp) computeContainmentDistanceLocsPolys(locs []*OperationDistance_GeometryLocation, polys []*Geom_Polygon, locPtPoly *[2]*OperationDistance_GeometryLocation) {
	for i := 0; i < len(locs); i++ {
		loc := locs[i]
		for j := 0; j < len(polys); j++ {
			op.computeContainmentDistanceLocPoly(loc, polys[j], locPtPoly)
			if op.minDistance <= op.terminateDistance {
				return
			}
		}
	}
}

func (op *OperationDistance_DistanceOp) computeContainmentDistanceLocPoly(ptLoc *OperationDistance_GeometryLocation, poly *Geom_Polygon, locPtPoly *[2]*OperationDistance_GeometryLocation) {
	pt := ptLoc.GetCoordinate()
	// If pt is not in exterior, distance to geom is 0.
	if Geom_Location_Exterior != op.ptLocator.Locate(pt, poly.Geom_Geometry) {
		op.minDistance = 0.0
		locPtPoly[0] = ptLoc
		locPtPoly[1] = OperationDistance_NewGeometryLocationInsideArea(poly.Geom_Geometry, pt)
		return
	}
}

// computeFacetDistance computes distance between facets (lines and points) of
// input geometries.
func (op *OperationDistance_DistanceOp) computeFacetDistance() {
	var locGeom [2]*OperationDistance_GeometryLocation

	// Geometries are not wholly inside, so compute distance from lines and
	// points of one to lines and points of the other.
	lines0 := GeomUtil_LinearComponentExtracter_GetLines(op.geom[0])
	lines1 := GeomUtil_LinearComponentExtracter_GetLines(op.geom[1])

	pts0 := GeomUtil_PointExtracter_GetPoints(op.geom[0])
	pts1 := GeomUtil_PointExtracter_GetPoints(op.geom[1])

	// Exit whenever minDistance goes LE than terminateDistance.
	op.computeMinDistanceLines(lines0, lines1, &locGeom)
	op.updateMinDistance(locGeom, false)
	if op.minDistance <= op.terminateDistance {
		return
	}

	locGeom[0] = nil
	locGeom[1] = nil
	op.computeMinDistanceLinesPoints(lines0, pts1, &locGeom)
	op.updateMinDistance(locGeom, false)
	if op.minDistance <= op.terminateDistance {
		return
	}

	locGeom[0] = nil
	locGeom[1] = nil
	op.computeMinDistanceLinesPoints(lines1, pts0, &locGeom)
	op.updateMinDistance(locGeom, true)
	if op.minDistance <= op.terminateDistance {
		return
	}

	locGeom[0] = nil
	locGeom[1] = nil
	op.computeMinDistancePoints(pts0, pts1, &locGeom)
	op.updateMinDistance(locGeom, false)
}

func (op *OperationDistance_DistanceOp) computeMinDistanceLines(lines0, lines1 []*Geom_LineString, locGeom *[2]*OperationDistance_GeometryLocation) {
	for i := 0; i < len(lines0); i++ {
		line0 := lines0[i]
		for j := 0; j < len(lines1); j++ {
			line1 := lines1[j]
			op.computeMinDistanceLineToLine(line0, line1, locGeom)
			if op.minDistance <= op.terminateDistance {
				return
			}
		}
	}
}

func (op *OperationDistance_DistanceOp) computeMinDistancePoints(points0, points1 []*Geom_Point, locGeom *[2]*OperationDistance_GeometryLocation) {
	for i := 0; i < len(points0); i++ {
		pt0 := points0[i]
		if pt0.IsEmpty() {
			continue
		}
		for j := 0; j < len(points1); j++ {
			pt1 := points1[j]
			if pt1.IsEmpty() {
				continue
			}
			dist := pt0.GetCoordinate().Distance(pt1.GetCoordinate())
			if dist < op.minDistance {
				op.minDistance = dist
				locGeom[0] = OperationDistance_NewGeometryLocation(pt0.Geom_Geometry, 0, pt0.GetCoordinate())
				locGeom[1] = OperationDistance_NewGeometryLocation(pt1.Geom_Geometry, 0, pt1.GetCoordinate())
			}
			if op.minDistance <= op.terminateDistance {
				return
			}
		}
	}
}

func (op *OperationDistance_DistanceOp) computeMinDistanceLinesPoints(lines []*Geom_LineString, points []*Geom_Point, locGeom *[2]*OperationDistance_GeometryLocation) {
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		for j := 0; j < len(points); j++ {
			pt := points[j]
			if pt.IsEmpty() {
				continue
			}
			op.computeMinDistanceLineToPoint(line, pt, locGeom)
			if op.minDistance <= op.terminateDistance {
				return
			}
		}
	}
}

func (op *OperationDistance_DistanceOp) computeMinDistanceLineToLine(line0, line1 *Geom_LineString, locGeom *[2]*OperationDistance_GeometryLocation) {
	if line0.GetEnvelopeInternal().Distance(line1.GetEnvelopeInternal()) > op.minDistance {
		return
	}
	coord0 := line0.GetCoordinates()
	coord1 := line1.GetCoordinates()
	// Brute force approach!
	for i := 0; i < len(coord0)-1; i++ {
		// Short-circuit if line segment is far from line.
		segEnv0 := Geom_NewEnvelopeFromCoordinates(coord0[i], coord0[i+1])
		if segEnv0.Distance(line1.GetEnvelopeInternal()) > op.minDistance {
			continue
		}

		for j := 0; j < len(coord1)-1; j++ {
			// Short-circuit if line segments are far apart.
			segEnv1 := Geom_NewEnvelopeFromCoordinates(coord1[j], coord1[j+1])
			if segEnv0.Distance(segEnv1) > op.minDistance {
				continue
			}

			dist := Algorithm_Distance_SegmentToSegment(
				coord0[i], coord0[i+1],
				coord1[j], coord1[j+1])
			if dist < op.minDistance {
				op.minDistance = dist
				seg0 := Geom_NewLineSegmentFromCoordinates(coord0[i], coord0[i+1])
				seg1 := Geom_NewLineSegmentFromCoordinates(coord1[j], coord1[j+1])
				closestPt := seg0.ClosestPoints(seg1)
				locGeom[0] = OperationDistance_NewGeometryLocation(line0.Geom_Geometry, i, closestPt[0])
				locGeom[1] = OperationDistance_NewGeometryLocation(line1.Geom_Geometry, j, closestPt[1])
			}
			if op.minDistance <= op.terminateDistance {
				return
			}
		}
	}
}

func (op *OperationDistance_DistanceOp) computeMinDistanceLineToPoint(line *Geom_LineString, pt *Geom_Point, locGeom *[2]*OperationDistance_GeometryLocation) {
	if line.GetEnvelopeInternal().Distance(pt.GetEnvelopeInternal()) > op.minDistance {
		return
	}
	coord0 := line.GetCoordinates()
	coord := pt.GetCoordinate()
	// Brute force approach!
	for i := 0; i < len(coord0)-1; i++ {
		dist := Algorithm_Distance_PointToSegment(coord, coord0[i], coord0[i+1])
		if dist < op.minDistance {
			op.minDistance = dist
			seg := Geom_NewLineSegmentFromCoordinates(coord0[i], coord0[i+1])
			segClosestPoint := seg.ClosestPoint(coord)
			locGeom[0] = OperationDistance_NewGeometryLocation(line.Geom_Geometry, i, segClosestPoint)
			locGeom[1] = OperationDistance_NewGeometryLocation(pt.Geom_Geometry, 0, coord)
		}
		if op.minDistance <= op.terminateDistance {
			return
		}
	}
}
