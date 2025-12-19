package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_PointExtracter extracts all the 0-dimensional (Point) components
// from a Geometry.
type GeomUtil_PointExtracter struct {
	pts []*Geom_Point
}

var _ Geom_GeometryFilter = (*GeomUtil_PointExtracter)(nil)

func (pe *GeomUtil_PointExtracter) IsGeom_GeometryFilter() {}

// GeomUtil_PointExtracter_GetPointsToSlice extracts the Point elements from a
// single Geometry and adds them to the provided slice.
func GeomUtil_PointExtracter_GetPointsToSlice(geom *Geom_Geometry, list []*Geom_Point) []*Geom_Point {
	if java.InstanceOf[*Geom_Point](geom) {
		return append(list, java.Cast[*Geom_Point](geom))
	}
	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		extracter := GeomUtil_NewPointExtracter(list)
		geom.ApplyGeometryFilter(extracter)
		return extracter.pts
	}
	// Skip non-Point elemental geometries.
	return list
}

// GeomUtil_PointExtracter_GetPoints extracts the Point elements from a single
// Geometry and returns them in a slice.
func GeomUtil_PointExtracter_GetPoints(geom *Geom_Geometry) []*Geom_Point {
	if java.InstanceOf[*Geom_Point](geom) {
		return []*Geom_Point{java.Cast[*Geom_Point](geom)}
	}
	return GeomUtil_PointExtracter_GetPointsToSlice(geom, nil)
}

// GeomUtil_NewPointExtracter constructs a PointExtracter with a slice in which
// to store Points found.
func GeomUtil_NewPointExtracter(pts []*Geom_Point) *GeomUtil_PointExtracter {
	return &GeomUtil_PointExtracter{pts: pts}
}

// Filter implements the GeometryFilter interface.
func (pe *GeomUtil_PointExtracter) Filter(geom *Geom_Geometry) {
	if java.InstanceOf[*Geom_Point](geom) {
		pe.pts = append(pe.pts, java.Cast[*Geom_Point](geom))
	}
}
