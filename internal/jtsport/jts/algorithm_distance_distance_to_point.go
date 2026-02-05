package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// AlgorithmDistance_DistanceToPoint computes the Euclidean distance (L2 metric)
// from a Coordinate to a Geometry.
// Also computes two points on the geometry which are separated by the distance found.
type AlgorithmDistance_DistanceToPoint struct{}

// AlgorithmDistance_NewDistanceToPoint creates a new DistanceToPoint.
func AlgorithmDistance_NewDistanceToPoint() *AlgorithmDistance_DistanceToPoint {
	return &AlgorithmDistance_DistanceToPoint{}
}

// AlgorithmDistance_DistanceToPoint_ComputeDistanceGeometry computes the distance
// from a Coordinate to a Geometry.
func AlgorithmDistance_DistanceToPoint_ComputeDistanceGeometry(geom *Geom_Geometry, pt *Geom_Coordinate, ptDist *AlgorithmDistance_PointPairDistance) {
	if java.InstanceOf[*Geom_LineString](geom) {
		AlgorithmDistance_DistanceToPoint_ComputeDistanceLineString(java.Cast[*Geom_LineString](geom), pt, ptDist)
	} else if java.InstanceOf[*Geom_Polygon](geom) {
		AlgorithmDistance_DistanceToPoint_ComputeDistancePolygon(java.Cast[*Geom_Polygon](geom), pt, ptDist)
	} else if java.InstanceOf[*Geom_GeometryCollection](geom) {
		gc := java.Cast[*Geom_GeometryCollection](geom)
		for i := 0; i < gc.GetNumGeometries(); i++ {
			g := gc.GetGeometryN(i)
			AlgorithmDistance_DistanceToPoint_ComputeDistanceGeometry(g, pt, ptDist)
		}
	} else { // assume geom is Point
		ptDist.SetMinimum(geom.GetCoordinate(), pt)
	}
}

// AlgorithmDistance_DistanceToPoint_ComputeDistanceLineString computes the distance
// from a Coordinate to a LineString.
func AlgorithmDistance_DistanceToPoint_ComputeDistanceLineString(line *Geom_LineString, pt *Geom_Coordinate, ptDist *AlgorithmDistance_PointPairDistance) {
	tempSegment := Geom_NewLineSegment()
	coords := line.GetCoordinates()
	for i := 0; i < len(coords)-1; i++ {
		tempSegment.SetCoordinates(coords[i], coords[i+1])
		// this is somewhat inefficient - could do better
		closestPt := tempSegment.ClosestPoint(pt)
		ptDist.SetMinimum(closestPt, pt)
	}
}

// AlgorithmDistance_DistanceToPoint_ComputeDistanceLineSegment computes the distance
// from a Coordinate to a LineSegment.
func AlgorithmDistance_DistanceToPoint_ComputeDistanceLineSegment(segment *Geom_LineSegment, pt *Geom_Coordinate, ptDist *AlgorithmDistance_PointPairDistance) {
	closestPt := segment.ClosestPoint(pt)
	ptDist.SetMinimum(closestPt, pt)
}

// AlgorithmDistance_DistanceToPoint_ComputeDistancePolygon computes the distance
// from a Coordinate to a Polygon.
func AlgorithmDistance_DistanceToPoint_ComputeDistancePolygon(poly *Geom_Polygon, pt *Geom_Coordinate, ptDist *AlgorithmDistance_PointPairDistance) {
	AlgorithmDistance_DistanceToPoint_ComputeDistanceLineString(poly.GetExteriorRing().Geom_LineString, pt, ptDist)
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		AlgorithmDistance_DistanceToPoint_ComputeDistanceLineString(poly.GetInteriorRingN(i).Geom_LineString, pt, ptDist)
	}
}
