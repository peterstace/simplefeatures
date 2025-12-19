package jts

// AlgorithmLocate_PointOnGeometryLocator is an interface for classes which
// determine the Location of points in a Geometry.
type AlgorithmLocate_PointOnGeometryLocator interface {
	// Locate determines the Location of a point in the Geometry.
	Locate(pt *Geom_Coordinate) int

	// IsAlgorithmLocate_PointOnGeometryLocator is a marker method for the interface.
	IsAlgorithmLocate_PointOnGeometryLocator()
}
