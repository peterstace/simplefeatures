package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationDistance_ConnectedElementPointFilter extracts a single point from
// each connected element in a Geometry (e.g. a polygon, linestring or point)
// and returns them in a list.
type OperationDistance_ConnectedElementPointFilter struct {
	pts []*Geom_Coordinate
}

var _ Geom_GeometryFilter = (*OperationDistance_ConnectedElementPointFilter)(nil)

func (f *OperationDistance_ConnectedElementPointFilter) IsGeom_GeometryFilter() {}

// OperationDistance_ConnectedElementPointFilter_GetCoordinates returns a list
// containing a Coordinate from each Polygon, LineString, and Point found
// inside the specified geometry. Thus, if the specified geometry is not a
// GeometryCollection, an empty list will be returned.
func OperationDistance_ConnectedElementPointFilter_GetCoordinates(geom *Geom_Geometry) []*Geom_Coordinate {
	var pts []*Geom_Coordinate
	filter := &OperationDistance_ConnectedElementPointFilter{pts: pts}
	geom.ApplyGeometryFilter(filter)
	return filter.pts
}

// operationDistance_NewConnectedElementPointFilter constructs a
// ConnectedElementPointFilter.
func operationDistance_NewConnectedElementPointFilter(pts []*Geom_Coordinate) *OperationDistance_ConnectedElementPointFilter {
	return &OperationDistance_ConnectedElementPointFilter{pts: pts}
}

// Filter implements the GeometryFilter interface.
func (f *OperationDistance_ConnectedElementPointFilter) Filter(geom *Geom_Geometry) {
	if java.InstanceOf[*Geom_Point](geom) ||
		java.InstanceOf[*Geom_LineString](geom) ||
		java.InstanceOf[*Geom_Polygon](geom) {
		f.pts = append(f.pts, geom.GetCoordinate())
	}
}
