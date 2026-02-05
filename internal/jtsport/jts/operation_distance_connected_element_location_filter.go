package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationDistance_ConnectedElementLocationFilter extracts a single point
// from each connected element in a Geometry (e.g. a polygon, linestring or
// point) and returns them in a list. The elements of the list are
// GeometryLocations. Empty geometries do not provide a location item.
type OperationDistance_ConnectedElementLocationFilter struct {
	locations []*OperationDistance_GeometryLocation
}

var _ Geom_GeometryFilter = (*OperationDistance_ConnectedElementLocationFilter)(nil)

func (f *OperationDistance_ConnectedElementLocationFilter) IsGeom_GeometryFilter() {}

// OperationDistance_ConnectedElementLocationFilter_GetLocations returns a list
// containing a point from each Polygon, LineString, and Point found inside the
// specified geometry. Thus, if the specified geometry is not a
// GeometryCollection, an empty list will be returned. The elements of the list
// are GeometryLocations.
func OperationDistance_ConnectedElementLocationFilter_GetLocations(geom *Geom_Geometry) []*OperationDistance_GeometryLocation {
	var locations []*OperationDistance_GeometryLocation
	filter := &OperationDistance_ConnectedElementLocationFilter{locations: locations}
	geom.ApplyGeometryFilter(filter)
	return filter.locations
}

// operationDistance_NewConnectedElementLocationFilter constructs a
// ConnectedElementLocationFilter.
func operationDistance_NewConnectedElementLocationFilter(locations []*OperationDistance_GeometryLocation) *OperationDistance_ConnectedElementLocationFilter {
	return &OperationDistance_ConnectedElementLocationFilter{locations: locations}
}

// Filter implements the GeometryFilter interface.
func (f *OperationDistance_ConnectedElementLocationFilter) Filter(geom *Geom_Geometry) {
	// Empty geometries do not provide a location.
	if geom.IsEmpty() {
		return
	}
	if java.InstanceOf[*Geom_Point](geom) ||
		java.InstanceOf[*Geom_LineString](geom) ||
		java.InstanceOf[*Geom_Polygon](geom) {
		f.locations = append(f.locations, OperationDistance_NewGeometryLocation(geom, 0, geom.GetCoordinate()))
	}
}
