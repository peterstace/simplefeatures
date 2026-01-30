package jts

// OperationOverlayng_InputGeometry manages the input geometries for an overlay
// operation. The second geometry is allowed to be null, to support for instance
// precision reduction.
type OperationOverlayng_InputGeometry struct {
	geom        [2]*Geom_Geometry
	ptLocatorA  AlgorithmLocate_PointOnGeometryLocator
	ptLocatorB  AlgorithmLocate_PointOnGeometryLocator
	isCollapsed [2]bool
}

// OperationOverlayng_NewInputGeometry creates a new InputGeometry for the given
// geometries.
func OperationOverlayng_NewInputGeometry(geomA, geomB *Geom_Geometry) *OperationOverlayng_InputGeometry {
	return &OperationOverlayng_InputGeometry{
		geom: [2]*Geom_Geometry{geomA, geomB},
	}
}

// IsSingle returns true if only one input geometry was provided.
func (ig *OperationOverlayng_InputGeometry) IsSingle() bool {
	return ig.geom[1] == nil
}

// GetDimension returns the dimension of the geometry at the given index.
func (ig *OperationOverlayng_InputGeometry) GetDimension(index int) int {
	if ig.geom[index] == nil {
		return -1
	}
	return ig.geom[index].GetDimension()
}

// GetGeometry returns the geometry at the given index.
func (ig *OperationOverlayng_InputGeometry) GetGeometry(geomIndex int) *Geom_Geometry {
	return ig.geom[geomIndex]
}

// GetEnvelope returns the envelope of the geometry at the given index.
func (ig *OperationOverlayng_InputGeometry) GetEnvelope(geomIndex int) *Geom_Envelope {
	return ig.geom[geomIndex].GetEnvelopeInternal()
}

// IsEmpty returns whether the geometry at the given index is empty.
func (ig *OperationOverlayng_InputGeometry) IsEmpty(geomIndex int) bool {
	return ig.geom[geomIndex].IsEmpty()
}

// IsArea returns whether the geometry at the given index is an area.
func (ig *OperationOverlayng_InputGeometry) IsArea(geomIndex int) bool {
	return ig.geom[geomIndex] != nil && ig.geom[geomIndex].GetDimension() == 2
}

// GetAreaIndex gets the index of an input which is an area, if one exists.
// Otherwise returns -1. If both inputs are areas, returns the index of the
// first one (0).
func (ig *OperationOverlayng_InputGeometry) GetAreaIndex() int {
	if ig.GetDimension(0) == 2 {
		return 0
	}
	if ig.GetDimension(1) == 2 {
		return 1
	}
	return -1
}

// IsLine returns whether the geometry at the given index is a line.
func (ig *OperationOverlayng_InputGeometry) IsLine(geomIndex int) bool {
	return ig.GetDimension(geomIndex) == 1
}

// IsAllPoints returns whether both inputs are points.
func (ig *OperationOverlayng_InputGeometry) IsAllPoints() bool {
	return ig.GetDimension(0) == 0 && ig.geom[1] != nil && ig.GetDimension(1) == 0
}

// HasPoints returns whether either input is a point.
func (ig *OperationOverlayng_InputGeometry) HasPoints() bool {
	return ig.GetDimension(0) == 0 || ig.GetDimension(1) == 0
}

// HasEdges tests if an input geometry has edges. This indicates that topology
// needs to be computed for them.
func (ig *OperationOverlayng_InputGeometry) HasEdges(geomIndex int) bool {
	return ig.geom[geomIndex] != nil && ig.geom[geomIndex].GetDimension() > 0
}

// LocatePointInArea determines the location within an area geometry. This
// allows disconnected edges to be fully located.
func (ig *OperationOverlayng_InputGeometry) LocatePointInArea(geomIndex int, pt *Geom_Coordinate) int {
	// Assert: only called if dimension(geomIndex) = 2

	if ig.isCollapsed[geomIndex] {
		return Geom_Location_Exterior
	}

	// this check is required because IndexedPointInAreaLocator can't handle
	// empty polygons
	if ig.GetGeometry(geomIndex).IsEmpty() || ig.isCollapsed[geomIndex] {
		return Geom_Location_Exterior
	}

	ptLocator := ig.getLocator(geomIndex)
	return ptLocator.Locate(pt)
}

func (ig *OperationOverlayng_InputGeometry) getLocator(geomIndex int) AlgorithmLocate_PointOnGeometryLocator {
	if geomIndex == 0 {
		if ig.ptLocatorA == nil {
			ig.ptLocatorA = AlgorithmLocate_NewIndexedPointInAreaLocator(ig.GetGeometry(geomIndex))
		}
		return ig.ptLocatorA
	}
	if ig.ptLocatorB == nil {
		ig.ptLocatorB = AlgorithmLocate_NewIndexedPointInAreaLocator(ig.GetGeometry(geomIndex))
	}
	return ig.ptLocatorB
}

// SetCollapsed sets whether the geometry at the given index is collapsed.
func (ig *OperationOverlayng_InputGeometry) SetCollapsed(geomIndex int, isGeomCollapsed bool) {
	ig.isCollapsed[geomIndex] = isGeomCollapsed
}
