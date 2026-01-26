package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geom_MultiPolygon models a collection of Polygons.
//
// As per the OGC SFS specification, the Polygons in a MultiPolygon may not
// overlap, and may only touch at single points. This allows the topological
// point-set semantics to be well-defined.
type Geom_MultiPolygon struct {
	*Geom_GeometryCollection
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mp *Geom_MultiPolygon) GetChild() java.Polymorphic {
	return mp.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mp *Geom_MultiPolygon) GetParent() java.Polymorphic {
	return mp.Geom_GeometryCollection
}

// Geom_NewMultiPolygonWithPrecisionModelAndSRID constructs a MultiPolygon.
//
// Parameters:
//   - polygons: the Polygons for this MultiPolygon, or nil or an empty slice to
//     create the empty geometry. Elements may be empty Polygons, but not nils.
//     The polygons must conform to the assertions specified in the OpenGIS
//     Simple Features Specification for SQL.
//
// Deprecated: Use GeometryFactory instead.
func Geom_NewMultiPolygonWithPrecisionModelAndSRID(polygons []*Geom_Polygon, precisionModel *Geom_PrecisionModel, srid int) *Geom_MultiPolygon {
	return Geom_NewMultiPolygon(polygons, Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, srid))
}

// Geom_NewMultiPolygon constructs a MultiPolygon.
//
// Parameters:
//   - polygons: the Polygons for this MultiPolygon, or nil or an empty slice to
//     create the empty geometry. Elements may be empty Polygons, but not nils.
//     The polygons must conform to the assertions specified in the OpenGIS
//     Simple Features Specification for SQL.
func Geom_NewMultiPolygon(polygons []*Geom_Polygon, factory *Geom_GeometryFactory) *Geom_MultiPolygon {
	geometries := make([]*Geom_Geometry, len(polygons))
	for i, p := range polygons {
		geometries[i] = p.Geom_Geometry
	}
	gc := Geom_NewGeometryCollection(geometries, factory)
	mp := &Geom_MultiPolygon{
		Geom_GeometryCollection: gc,
	}
	gc.child = mp
	return mp
}

func (mp *Geom_MultiPolygon) GetDimension_BODY() int {
	return 2
}

func (mp *Geom_MultiPolygon) HasDimension_BODY(dim int) bool {
	return dim == 2
}

func (mp *Geom_MultiPolygon) GetBoundaryDimension_BODY() int {
	return 1
}

func (mp *Geom_MultiPolygon) GetGeometryType_BODY() string {
	return Geom_Geometry_TypeNameMultiPolygon
}

// GetBoundary computes the boundary of this geometry.
func (mp *Geom_MultiPolygon) GetBoundary_BODY() *Geom_Geometry {
	if mp.IsEmpty() {
		return mp.GetFactory().CreateMultiLineString().Geom_Geometry
	}
	allRings := []*Geom_LineString{}
	for i := 0; i < len(mp.geometries); i++ {
		polygon := java.Cast[*Geom_Polygon](mp.geometries[i])
		rings := polygon.GetBoundary()
		for j := 0; j < rings.GetNumGeometries(); j++ {
			allRings = append(allRings, java.Cast[*Geom_LineString](rings.GetGeometryN(j)))
		}
	}
	return mp.GetFactory().CreateMultiLineStringFromLineStrings(allRings).Geom_Geometry
}

// Reverse creates a MultiPolygon with every component reversed. The order of
// the components in the collection are not reversed.
func (mp *Geom_MultiPolygon) Reverse() *Geom_MultiPolygon {
	reversed := mp.Geom_Geometry.Reverse()
	return java.Cast[*Geom_MultiPolygon](reversed)
}

func (mp *Geom_MultiPolygon) Reverse_BODY() *Geom_Geometry {
	return mp.ReverseInternal().Geom_Geometry
}

func (mp *Geom_MultiPolygon) ReverseInternal() *Geom_MultiPolygon {
	polygons := make([]*Geom_Polygon, len(mp.geometries))
	for i := range polygons {
		polygons[i] = java.Cast[*Geom_Polygon](mp.geometries[i].Reverse())
	}
	return Geom_NewMultiPolygon(polygons, mp.factory)
}

func (mp *Geom_MultiPolygon) CopyInternal_BODY() *Geom_Geometry {
	polygons := make([]*Geom_Polygon, len(mp.geometries))
	for i := range polygons {
		polygons[i] = java.Cast[*Geom_Polygon](mp.geometries[i].Copy())
	}
	return Geom_NewMultiPolygon(polygons, mp.factory).Geom_Geometry
}

func (mp *Geom_MultiPolygon) EqualsExactWithTolerance_BODY(other *Geom_Geometry, tolerance float64) bool {
	if !mp.IsEquivalentClass(other) {
		return false
	}
	return mp.Geom_GeometryCollection.EqualsExactWithTolerance_BODY(other, tolerance)
}

func (mp *Geom_MultiPolygon) GetTypeCode_BODY() int {
	return Geom_Geometry_TypeCodeMultiPolygon
}

// isPolygonal implements the Polygonal marker interface.
func (mp *Geom_MultiPolygon) IsPolygonal() {}
