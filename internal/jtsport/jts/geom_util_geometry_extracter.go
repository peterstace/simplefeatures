package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_GeometryExtracter extracts the components of a given type from a
// Geometry.
type GeomUtil_GeometryExtracter struct {
	geometryType string
	comps        []*Geom_Geometry
}

var _ Geom_GeometryFilter = (*GeomUtil_GeometryExtracter)(nil)

func (ge *GeomUtil_GeometryExtracter) IsGeom_GeometryFilter() {}

// GeomUtil_GeometryExtracter_ExtractToSlice extracts the components of
// geometryType from a Geometry and adds them to the provided slice.
func GeomUtil_GeometryExtracter_ExtractToSlice(geom *Geom_Geometry, geometryType string, list []*Geom_Geometry) []*Geom_Geometry {
	if geom.GetGeometryType() == geometryType {
		return append(list, geom)
	}
	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		extracter := GeomUtil_NewGeometryExtracter(geometryType, list)
		geom.ApplyGeometryFilter(extracter)
		return extracter.comps
	}
	// Skip non-matching elemental geometries.
	return list
}

// GeomUtil_GeometryExtracter_Extract extracts the components of geometryType
// from a Geometry and returns them in a slice.
func GeomUtil_GeometryExtracter_Extract(geom *Geom_Geometry, geometryType string) []*Geom_Geometry {
	return GeomUtil_GeometryExtracter_ExtractToSlice(geom, geometryType, nil)
}

// GeomUtil_NewGeometryExtracter constructs a GeometryExtracter with a geometry
// type to extract and a slice in which to store the extracted geometries.
func GeomUtil_NewGeometryExtracter(geometryType string, comps []*Geom_Geometry) *GeomUtil_GeometryExtracter {
	return &GeomUtil_GeometryExtracter{
		geometryType: geometryType,
		comps:        comps,
	}
}

// GeomUtil_GeometryExtracter_IsOfType checks if the geometry is of the
// specified type. LinearRings are considered LineStrings.
func GeomUtil_GeometryExtracter_IsOfType(geom *Geom_Geometry, geometryType string) bool {
	if geom.GetGeometryType() == geometryType {
		return true
	}
	if geometryType == Geom_Geometry_TypeNameLineString &&
		geom.GetGeometryType() == Geom_Geometry_TypeNameLinearRing {
		return true
	}
	return false
}

// Filter implements the GeometryFilter interface.
func (ge *GeomUtil_GeometryExtracter) Filter(geom *Geom_Geometry) {
	if ge.geometryType == "" || GeomUtil_GeometryExtracter_IsOfType(geom, ge.geometryType) {
		ge.comps = append(ge.comps, geom)
	}
}
