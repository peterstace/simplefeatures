package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationBufferValidate_DistanceToPointFinder computes the Euclidean distance (L2 metric)
// from a Point to a Geometry.
// Also computes two points which are separated by the distance.
type OperationBufferValidate_DistanceToPointFinder struct{}

// OperationBufferValidate_NewDistanceToPointFinder creates a new DistanceToPointFinder.
func OperationBufferValidate_NewDistanceToPointFinder() *OperationBufferValidate_DistanceToPointFinder {
	return &OperationBufferValidate_DistanceToPointFinder{}
}

// OperationBufferValidate_DistanceToPointFinder_ComputeDistanceGeometry computes the distance
// from a geometry to a point.
func OperationBufferValidate_DistanceToPointFinder_ComputeDistanceGeometry(geom *Geom_Geometry, pt *Geom_Coordinate, ptDist *OperationBufferValidate_PointPairDistance) {
	if java.InstanceOf[*Geom_LineString](geom) {
		OperationBufferValidate_DistanceToPointFinder_ComputeDistanceLineString(java.Cast[*Geom_LineString](geom), pt, ptDist)
	} else if java.InstanceOf[*Geom_Polygon](geom) {
		OperationBufferValidate_DistanceToPointFinder_ComputeDistancePolygon(java.Cast[*Geom_Polygon](geom), pt, ptDist)
	} else if java.InstanceOf[*Geom_GeometryCollection](geom) {
		gc := java.Cast[*Geom_GeometryCollection](geom)
		for i := 0; i < gc.GetNumGeometries(); i++ {
			g := gc.GetGeometryN(i)
			OperationBufferValidate_DistanceToPointFinder_ComputeDistanceGeometry(g, pt, ptDist)
		}
	} else { // assume geom is Point
		ptDist.SetMinimum(geom.GetCoordinate(), pt)
	}
}

// OperationBufferValidate_DistanceToPointFinder_ComputeDistanceLineString computes the distance
// from a LineString to a point.
func OperationBufferValidate_DistanceToPointFinder_ComputeDistanceLineString(line *Geom_LineString, pt *Geom_Coordinate, ptDist *OperationBufferValidate_PointPairDistance) {
	coords := line.GetCoordinates()
	tempSegment := Geom_NewLineSegment()
	for i := 0; i < len(coords)-1; i++ {
		tempSegment.SetCoordinates(coords[i], coords[i+1])
		// this is somewhat inefficient - could do better
		closestPt := tempSegment.ClosestPoint(pt)
		ptDist.SetMinimum(closestPt, pt)
	}
}

// OperationBufferValidate_DistanceToPointFinder_ComputeDistanceLineSegment computes the distance
// from a LineSegment to a point.
func OperationBufferValidate_DistanceToPointFinder_ComputeDistanceLineSegment(segment *Geom_LineSegment, pt *Geom_Coordinate, ptDist *OperationBufferValidate_PointPairDistance) {
	closestPt := segment.ClosestPoint(pt)
	ptDist.SetMinimum(closestPt, pt)
}

// OperationBufferValidate_DistanceToPointFinder_ComputeDistancePolygon computes the distance
// from a Polygon to a point.
func OperationBufferValidate_DistanceToPointFinder_ComputeDistancePolygon(poly *Geom_Polygon, pt *Geom_Coordinate, ptDist *OperationBufferValidate_PointPairDistance) {
	OperationBufferValidate_DistanceToPointFinder_ComputeDistanceLineString(poly.GetExteriorRing().Geom_LineString, pt, ptDist)
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		OperationBufferValidate_DistanceToPointFinder_ComputeDistanceLineString(poly.GetInteriorRingN(i).Geom_LineString, pt, ptDist)
	}
}
