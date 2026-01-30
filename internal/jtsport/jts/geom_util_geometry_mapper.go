package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_GeometryMapper_MapOp is an interface for geometry functions that map
// a geometry input to a geometry output. The output may be nil if there is no
// valid output value for the given input value.
type GeomUtil_GeometryMapper_MapOp interface {
	Map(geom *Geom_Geometry) *Geom_Geometry
}

// GeomUtil_GeometryMapper_Map maps the members of a Geometry (which may be
// atomic or composite) into another Geometry of most specific type. nil results
// are skipped. In the case of hierarchical GeometryCollections, only the first
// level of members are mapped.
func GeomUtil_GeometryMapper_Map(geom *Geom_Geometry, op GeomUtil_GeometryMapper_MapOp) *Geom_Geometry {
	var mapped []*Geom_Geometry
	for i := 0; i < geom.GetNumGeometries(); i++ {
		g := op.Map(geom.GetGeometryN(i))
		if g != nil {
			mapped = append(mapped, g)
		}
	}
	return geom.GetFactory().BuildGeometry(mapped)
}

// GeomUtil_GeometryMapper_MapSlice maps a slice of geometries using the given
// operation.
func GeomUtil_GeometryMapper_MapSlice(geoms []*Geom_Geometry, op GeomUtil_GeometryMapper_MapOp) []*Geom_Geometry {
	var mapped []*Geom_Geometry
	for _, g := range geoms {
		gr := op.Map(g)
		if gr != nil {
			mapped = append(mapped, gr)
		}
	}
	return mapped
}

// GeomUtil_GeometryMapper_FlatMap maps the atomic elements of a Geometry (which
// may be atomic or composite) using a MapOp mapping operation into an atomic
// Geometry or a flat collection of the most specific type. nil and empty values
// returned from the mapping operation are discarded.
func GeomUtil_GeometryMapper_FlatMap(geom *Geom_Geometry, emptyDim int, op GeomUtil_GeometryMapper_MapOp) *Geom_Geometry {
	var mapped []*Geom_Geometry
	geomUtil_GeometryMapper_flatMap(geom, op, &mapped)

	if len(mapped) == 0 {
		return geom.GetFactory().CreateEmpty(emptyDim)
	}
	if len(mapped) == 1 {
		return mapped[0]
	}
	return geom.GetFactory().BuildGeometry(mapped)
}

func geomUtil_GeometryMapper_flatMap(geom *Geom_Geometry, op GeomUtil_GeometryMapper_MapOp, mapped *[]*Geom_Geometry) {
	for i := 0; i < geom.GetNumGeometries(); i++ {
		g := geom.GetGeometryN(i)
		if java.InstanceOf[*Geom_GeometryCollection](g) {
			geomUtil_GeometryMapper_flatMap(g, op, mapped)
		} else {
			res := op.Map(g)
			if res != nil && !res.IsEmpty() {
				geomUtil_GeometryMapper_addFlat(res, mapped)
			}
		}
	}
}

func geomUtil_GeometryMapper_addFlat(geom *Geom_Geometry, geomList *[]*Geom_Geometry) {
	if geom.IsEmpty() {
		return
	}
	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		for i := 0; i < geom.GetNumGeometries(); i++ {
			geomUtil_GeometryMapper_addFlat(geom.GetGeometryN(i), geomList)
		}
	} else {
		*geomList = append(*geomList, geom)
	}
}
