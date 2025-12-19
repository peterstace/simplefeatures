package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_LineStringExtracter extracts all the LineString elements from a
// Geometry.
type GeomUtil_LineStringExtracter struct {
	comps []*Geom_LineString
}

var _ Geom_GeometryFilter = (*GeomUtil_LineStringExtracter)(nil)

func (lse *GeomUtil_LineStringExtracter) IsGeom_GeometryFilter() {}

// GeomUtil_LineStringExtracter_GetLinesToSlice extracts the LineString elements
// from a single Geometry and adds them to the provided slice.
func GeomUtil_LineStringExtracter_GetLinesToSlice(geom *Geom_Geometry, lines []*Geom_LineString) []*Geom_LineString {
	if java.InstanceOf[*Geom_LineString](geom) {
		return append(lines, java.Cast[*Geom_LineString](geom))
	}
	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		extracter := GeomUtil_NewLineStringExtracter(lines)
		geom.ApplyGeometryFilter(extracter)
		return extracter.comps
	}
	// Skip non-LineString elemental geometries.
	return lines
}

// GeomUtil_LineStringExtracter_GetLines extracts the LineString elements from a
// single Geometry and returns them in a slice.
func GeomUtil_LineStringExtracter_GetLines(geom *Geom_Geometry) []*Geom_LineString {
	return GeomUtil_LineStringExtracter_GetLinesToSlice(geom, nil)
}

// GeomUtil_LineStringExtracter_GetGeometry extracts the LineString elements
// from a single Geometry and returns them as either a LineString or
// MultiLineString.
func GeomUtil_LineStringExtracter_GetGeometry(geom *Geom_Geometry) *Geom_Geometry {
	lines := GeomUtil_LineStringExtracter_GetLines(geom)
	geoms := make([]*Geom_Geometry, len(lines))
	for i, line := range lines {
		geoms[i] = line.Geom_Geometry
	}
	return geom.GetFactory().BuildGeometry(geoms)
}

// GeomUtil_NewLineStringExtracter constructs a LineStringExtracter with a slice
// in which to store LineStrings found.
func GeomUtil_NewLineStringExtracter(comps []*Geom_LineString) *GeomUtil_LineStringExtracter {
	return &GeomUtil_LineStringExtracter{comps: comps}
}

// Filter implements the GeometryFilter interface.
func (lse *GeomUtil_LineStringExtracter) Filter(geom *Geom_Geometry) {
	if java.InstanceOf[*Geom_LineString](geom) {
		lse.comps = append(lse.comps, java.Cast[*Geom_LineString](geom))
	}
}
