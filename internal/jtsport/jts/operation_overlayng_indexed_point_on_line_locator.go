package jts

// Compile-time interface check.
var _ AlgorithmLocate_PointOnGeometryLocator = (*OperationOverlayng_IndexedPointOnLineLocator)(nil)

// OperationOverlayng_IndexedPointOnLineLocator locates points on a linear
// geometry, using a spatial index to provide good performance.
type OperationOverlayng_IndexedPointOnLineLocator struct {
	inputGeom *Geom_Geometry
}

// IsAlgorithmLocate_PointOnGeometryLocator is a marker method for the interface.
func (ipoll *OperationOverlayng_IndexedPointOnLineLocator) IsAlgorithmLocate_PointOnGeometryLocator() {
}

// OperationOverlayng_NewIndexedPointOnLineLocator creates a new
// IndexedPointOnLineLocator.
func OperationOverlayng_NewIndexedPointOnLineLocator(geomLinear *Geom_Geometry) *OperationOverlayng_IndexedPointOnLineLocator {
	return &OperationOverlayng_IndexedPointOnLineLocator{
		inputGeom: geomLinear,
	}
}

// Locate implements the point location algorithm.
func (ipoll *OperationOverlayng_IndexedPointOnLineLocator) Locate(p *Geom_Coordinate) int {
	// TODO: optimize this with a segment index.
	locator := Algorithm_NewPointLocator()
	return locator.Locate(p, ipoll.inputGeom)
}
