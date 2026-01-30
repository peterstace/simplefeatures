package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geom_MultiPoint models a collection of Points.
//
// Any collection of Points is a valid MultiPoint.
type Geom_MultiPoint struct {
	*Geom_GeometryCollection
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mp *Geom_MultiPoint) GetChild() java.Polymorphic {
	return mp.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mp *Geom_MultiPoint) GetParent() java.Polymorphic {
	return mp.Geom_GeometryCollection
}

// Geom_NewMultiPointWithPrecisionModelAndSRID constructs a MultiPoint.
//
// Parameters:
//   - points: the Points for this MultiPoint, or nil or an empty slice to
//     create the empty geometry. Elements may be empty Points, but not nils.
//
// Deprecated: Use GeometryFactory instead.
func Geom_NewMultiPointWithPrecisionModelAndSRID(points []*Geom_Point, precisionModel *Geom_PrecisionModel, srid int) *Geom_MultiPoint {
	return Geom_NewMultiPoint(points, Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, srid))
}

// Geom_NewMultiPoint constructs a MultiPoint.
//
// Parameters:
//   - points: the Points for this MultiPoint, or nil or an empty slice to
//     create the empty geometry. Elements may be empty Points, but not nils.
func Geom_NewMultiPoint(points []*Geom_Point, factory *Geom_GeometryFactory) *Geom_MultiPoint {
	geometries := make([]*Geom_Geometry, len(points))
	for i, p := range points {
		geometries[i] = p.Geom_Geometry
	}
	gc := Geom_NewGeometryCollection(geometries, factory)
	mp := &Geom_MultiPoint{
		Geom_GeometryCollection: gc,
	}
	gc.child = mp
	return mp
}

func (mp *Geom_MultiPoint) GetDimension_BODY() int {
	return 0
}

func (mp *Geom_MultiPoint) HasDimension_BODY(dim int) bool {
	return dim == 0
}

func (mp *Geom_MultiPoint) GetBoundaryDimension_BODY() int {
	return Geom_Dimension_False
}

func (mp *Geom_MultiPoint) GetGeometryType_BODY() string {
	return Geom_Geometry_TypeNameMultiPoint
}

// GetBoundary gets the boundary of this geometry.
// Zero-dimensional geometries have no boundary by definition,
// so an empty GeometryCollection is returned.
func (mp *Geom_MultiPoint) GetBoundary_BODY() *Geom_Geometry {
	return mp.GetFactory().CreateGeometryCollection().Geom_Geometry
}

func (mp *Geom_MultiPoint) Reverse() *Geom_MultiPoint {
	reversed := mp.Geom_Geometry.Reverse()
	return java.Cast[*Geom_MultiPoint](reversed)
}

func (mp *Geom_MultiPoint) Reverse_BODY() *Geom_Geometry {
	return mp.ReverseInternal().Geom_Geometry
}

func (mp *Geom_MultiPoint) ReverseInternal() *Geom_MultiPoint {
	points := make([]*Geom_Point, len(mp.geometries))
	for i := range points {
		points[i] = java.Cast[*Geom_Point](mp.geometries[i].Copy())
	}
	return Geom_NewMultiPoint(points, mp.factory)
}

func (mp *Geom_MultiPoint) EqualsExactWithTolerance_BODY(other *Geom_Geometry, tolerance float64) bool {
	if !mp.IsEquivalentClass(other) {
		return false
	}
	return mp.Geom_GeometryCollection.EqualsExactWithTolerance_BODY(other, tolerance)
}

// getCoordinate returns the Coordinate at the given position.
func (mp *Geom_MultiPoint) getCoordinate(n int) *Geom_Coordinate {
	return java.Cast[*Geom_Point](mp.geometries[n]).GetCoordinate()
}

func (mp *Geom_MultiPoint) CopyInternal_BODY() *Geom_Geometry {
	points := make([]*Geom_Point, len(mp.geometries))
	for i := range points {
		points[i] = java.Cast[*Geom_Point](mp.geometries[i].Copy())
	}
	return Geom_NewMultiPoint(points, mp.factory).Geom_Geometry
}

func (mp *Geom_MultiPoint) GetTypeCode_BODY() int {
	return Geom_Geometry_TypeCodeMultiPoint
}

// isPuntal implements the Puntal marker interface.
func (mp *Geom_MultiPoint) IsPuntal() {}
