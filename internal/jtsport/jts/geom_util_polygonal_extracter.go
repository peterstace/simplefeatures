package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_PolygonalExtracter extracts the Polygon and MultiPolygon elements
// from a Geometry.
type GeomUtil_PolygonalExtracter struct{}

// GeomUtil_PolygonalExtracter_GetPolygonalsToSlice extracts the Polygon and
// MultiPolygon elements from a Geometry and adds them to the provided slice.
func GeomUtil_PolygonalExtracter_GetPolygonalsToSlice(geom *Geom_Geometry, list []*Geom_Geometry) []*Geom_Geometry {
	if java.InstanceOf[*Geom_Polygon](geom) || java.InstanceOf[*Geom_MultiPolygon](geom) {
		switch s := java.GetLeaf(geom).(type) {
		case *Geom_Polygon:
			return append(list, s.Geom_Geometry)
		case *Geom_MultiPolygon:
			return append(list, s.Geom_Geometry)
		}
	}
	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		for i := 0; i < geom.GetNumGeometries(); i++ {
			list = GeomUtil_PolygonalExtracter_GetPolygonalsToSlice(geom.GetGeometryN(i), list)
		}
	}
	// Skip non-Polygonal elemental geometries.
	return list
}

// GeomUtil_PolygonalExtracter_GetPolygonals extracts the Polygon and
// MultiPolygon elements from a Geometry and returns them in a slice.
func GeomUtil_PolygonalExtracter_GetPolygonals(geom *Geom_Geometry) []*Geom_Geometry {
	return GeomUtil_PolygonalExtracter_GetPolygonalsToSlice(geom, nil)
}
