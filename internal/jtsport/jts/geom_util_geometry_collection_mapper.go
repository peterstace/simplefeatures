package jts

// GeomUtil_GeometryCollectionMapper maps the members of a GeometryCollection
// into another GeometryCollection via a defined mapping function.
type GeomUtil_GeometryCollectionMapper struct {
	mapOp GeomUtil_GeometryMapper_MapOp
}

// GeomUtil_GeometryCollectionMapper_Map is a static method that maps a
// GeometryCollection using the given function. This variant accepts a function
// directly for convenience.
func GeomUtil_GeometryCollectionMapper_Map(gc *Geom_GeometryCollection, op func(*Geom_Geometry) *Geom_Geometry) *Geom_Geometry {
	var mapped []*Geom_Geometry
	for i := 0; i < gc.GetNumGeometries(); i++ {
		g := op(gc.GetGeometryN(i))
		if g != nil && !g.IsEmpty() {
			mapped = append(mapped, g)
		}
	}
	return gc.GetFactory().BuildGeometry(mapped)
}

// GeomUtil_GeometryCollectionMapper_MapWithOp is a static method that maps a
// GeometryCollection using the given MapOp interface.
func GeomUtil_GeometryCollectionMapper_MapWithOp(gc *Geom_GeometryCollection, op GeomUtil_GeometryMapper_MapOp) *Geom_GeometryCollection {
	mapper := GeomUtil_NewGeometryCollectionMapper(op)
	return mapper.Map(gc)
}

// GeomUtil_NewGeometryCollectionMapper creates a new GeometryCollectionMapper
// with the given MapOp.
func GeomUtil_NewGeometryCollectionMapper(mapOp GeomUtil_GeometryMapper_MapOp) *GeomUtil_GeometryCollectionMapper {
	return &GeomUtil_GeometryCollectionMapper{
		mapOp: mapOp,
	}
}

// Map maps the GeometryCollection using the configured MapOp.
func (gcm *GeomUtil_GeometryCollectionMapper) Map(gc *Geom_GeometryCollection) *Geom_GeometryCollection {
	var mapped []*Geom_Geometry
	for i := 0; i < gc.GetNumGeometries(); i++ {
		g := gcm.mapOp.Map(gc.GetGeometryN(i))
		if !g.IsEmpty() {
			mapped = append(mapped, g)
		}
	}
	return gc.GetFactory().CreateGeometryCollectionFromGeometries(
		Geom_GeometryFactory_ToGeometryArray(mapped))
}
