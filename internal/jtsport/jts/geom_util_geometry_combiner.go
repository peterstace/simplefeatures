package jts

// GeomUtil_GeometryCombiner combines Geometries to produce a GeometryCollection
// of the most appropriate type. Input geometries which are already collections
// will have their elements extracted first. No validation of the result
// geometry is performed. (The only case where invalidity is possible is where
// Polygonal geometries are combined and result in a self-intersection).
type GeomUtil_GeometryCombiner struct {
	geomFactory *Geom_GeometryFactory
	skipEmpty   bool
	inputGeoms  []*Geom_Geometry
}

// GeomUtil_GeometryCombiner_CombineSlice combines a slice of geometries.
func GeomUtil_GeometryCombiner_CombineSlice(geoms []*Geom_Geometry) *Geom_Geometry {
	combiner := GeomUtil_NewGeometryCombiner(geoms)
	return combiner.Combine()
}

// GeomUtil_GeometryCombiner_Combine2 combines two geometries.
func GeomUtil_GeometryCombiner_Combine2(g0, g1 *Geom_Geometry) *Geom_Geometry {
	combiner := GeomUtil_NewGeometryCombiner([]*Geom_Geometry{g0, g1})
	return combiner.Combine()
}

// GeomUtil_GeometryCombiner_Combine3 combines three geometries.
func GeomUtil_GeometryCombiner_Combine3(g0, g1, g2 *Geom_Geometry) *Geom_Geometry {
	combiner := GeomUtil_NewGeometryCombiner([]*Geom_Geometry{g0, g1, g2})
	return combiner.Combine()
}

// GeomUtil_GeometryCombiner_ExtractFactory extracts the GeometryFactory used by
// the geometries in a slice.
func GeomUtil_GeometryCombiner_ExtractFactory(geoms []*Geom_Geometry) *Geom_GeometryFactory {
	if len(geoms) == 0 {
		return nil
	}
	return geoms[0].GetFactory()
}

// GeomUtil_NewGeometryCombiner creates a new combiner for a slice of
// geometries.
func GeomUtil_NewGeometryCombiner(geoms []*Geom_Geometry) *GeomUtil_GeometryCombiner {
	return &GeomUtil_GeometryCombiner{
		geomFactory: GeomUtil_GeometryCombiner_ExtractFactory(geoms),
		skipEmpty:   false,
		inputGeoms:  geoms,
	}
}

// Combine computes the combination of the input geometries to produce the most
// appropriate Geometry or GeometryCollection.
func (gc *GeomUtil_GeometryCombiner) Combine() *Geom_Geometry {
	var elems []*Geom_Geometry
	for _, g := range gc.inputGeoms {
		elems = gc.extractElements(g, elems)
	}

	if len(elems) == 0 {
		if gc.geomFactory != nil {
			// Return an empty GeometryCollection.
			return gc.geomFactory.CreateGeometryCollection().Geom_Geometry
		}
		return nil
	}
	// Return the "simplest possible" geometry.
	return gc.geomFactory.BuildGeometry(elems)
}

// extractElements extracts elements from a geometry and appends them to the
// elems slice.
func (gc *GeomUtil_GeometryCombiner) extractElements(geom *Geom_Geometry, elems []*Geom_Geometry) []*Geom_Geometry {
	if geom == nil {
		return elems
	}

	for i := 0; i < geom.GetNumGeometries(); i++ {
		elemGeom := geom.GetGeometryN(i)
		if gc.skipEmpty && elemGeom.IsEmpty() {
			continue
		}
		elems = append(elems, elemGeom)
	}
	return elems
}
