package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_PolygonExtracter extracts all the Polygon elements from a Geometry.
type GeomUtil_PolygonExtracter struct {
	comps []*Geom_Polygon
}

var _ Geom_GeometryFilter = (*GeomUtil_PolygonExtracter)(nil)

func (pe *GeomUtil_PolygonExtracter) IsGeom_GeometryFilter() {}

// GeomUtil_PolygonExtracter_GetPolygonsToSlice extracts the Polygon elements
// from a single Geometry and adds them to the provided slice.
func GeomUtil_PolygonExtracter_GetPolygonsToSlice(geom *Geom_Geometry, list []*Geom_Polygon) []*Geom_Polygon {
	if java.InstanceOf[*Geom_Polygon](geom) {
		return append(list, java.Cast[*Geom_Polygon](geom))
	}
	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		extracter := GeomUtil_NewPolygonExtracter(list)
		geom.ApplyGeometryFilter(extracter)
		return extracter.comps
	}
	// Skip non-Polygonal elemental geometries.
	return list
}

// GeomUtil_PolygonExtracter_GetPolygons extracts the Polygon elements from a
// single Geometry and returns them in a slice.
func GeomUtil_PolygonExtracter_GetPolygons(geom *Geom_Geometry) []*Geom_Polygon {
	return GeomUtil_PolygonExtracter_GetPolygonsToSlice(geom, nil)
}

// GeomUtil_NewPolygonExtracter constructs a PolygonExtracter with a slice in
// which to store Polygons found.
func GeomUtil_NewPolygonExtracter(comps []*Geom_Polygon) *GeomUtil_PolygonExtracter {
	return &GeomUtil_PolygonExtracter{comps: comps}
}

// Filter implements the GeometryFilter interface.
func (pe *GeomUtil_PolygonExtracter) Filter(geom *Geom_Geometry) {
	if java.InstanceOf[*Geom_Polygon](geom) {
		pe.comps = append(pe.comps, java.Cast[*Geom_Polygon](geom))
	}
}
