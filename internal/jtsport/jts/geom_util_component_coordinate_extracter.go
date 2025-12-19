package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_ComponentCoordinateExtracter extracts a representative Coordinate
// from each connected component of a Geometry.
type GeomUtil_ComponentCoordinateExtracter struct {
	coords []*Geom_Coordinate
}

var _ Geom_GeometryComponentFilter = (*GeomUtil_ComponentCoordinateExtracter)(nil)

func (cce *GeomUtil_ComponentCoordinateExtracter) IsGeom_GeometryComponentFilter() {}

// GeomUtil_ComponentCoordinateExtracter_GetCoordinates extracts a
// representative Coordinate from each connected component in a geometry.
//
// If more than one geometry is to be processed, it is more efficient to create
// a single ComponentCoordinateExtracter instance and pass it to each geometry.
func GeomUtil_ComponentCoordinateExtracter_GetCoordinates(geom *Geom_Geometry) []*Geom_Coordinate {
	var coords []*Geom_Coordinate
	extracter := GeomUtil_NewComponentCoordinateExtracter(coords)
	geom.Apply(extracter)
	return extracter.coords
}

// GeomUtil_NewComponentCoordinateExtracter constructs a
// ComponentCoordinateExtracter with a slice in which to store Coordinates found.
func GeomUtil_NewComponentCoordinateExtracter(coords []*Geom_Coordinate) *GeomUtil_ComponentCoordinateExtracter {
	return &GeomUtil_ComponentCoordinateExtracter{coords: coords}
}

// Filter implements the GeometryComponentFilter interface.
func (cce *GeomUtil_ComponentCoordinateExtracter) Filter(geom *Geom_Geometry) {
	if geom.IsEmpty() {
		return
	}
	// Add coordinates from connected components.
	// Point.GetCoordinate() is not polymorphic (no body override), so we must
	// cast to the concrete type to call the method directly.
	if java.InstanceOf[*Geom_LineString](geom) {
		cce.coords = append(cce.coords, java.Cast[*Geom_LineString](geom).GetCoordinate())
	} else if java.InstanceOf[*Geom_Point](geom) {
		cce.coords = append(cce.coords, java.Cast[*Geom_Point](geom).GetCoordinate())
	}
}
