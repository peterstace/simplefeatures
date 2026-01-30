package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geom_MultiLineString models a collection of LineStrings.
//
// Any collection of LineStrings is a valid MultiLineString.
type Geom_MultiLineString struct {
	*Geom_GeometryCollection
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mls *Geom_MultiLineString) GetChild() java.Polymorphic {
	return mls.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mls *Geom_MultiLineString) GetParent() java.Polymorphic {
	return mls.Geom_GeometryCollection
}

// Geom_NewMultiLineStringWithPrecisionModelAndSRID constructs a MultiLineString.
//
// Parameters:
//   - lineStrings: the LineStrings for this MultiLineString, or nil or an empty
//     slice to create the empty geometry. Elements may be empty LineStrings, but
//     not nils.
//
// Deprecated: Use GeometryFactory instead.
func Geom_NewMultiLineStringWithPrecisionModelAndSRID(lineStrings []*Geom_LineString, precisionModel *Geom_PrecisionModel, srid int) *Geom_MultiLineString {
	return Geom_NewMultiLineString(lineStrings, Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, srid))
}

// Geom_NewMultiLineString constructs a MultiLineString.
//
// Parameters:
//   - lineStrings: the LineStrings for this MultiLineString, or nil or an empty
//     slice to create the empty geometry. Elements may be empty LineStrings, but
//     not nils.
func Geom_NewMultiLineString(lineStrings []*Geom_LineString, factory *Geom_GeometryFactory) *Geom_MultiLineString {
	geometries := make([]*Geom_Geometry, len(lineStrings))
	for i, ls := range lineStrings {
		geometries[i] = ls.Geom_Geometry
	}
	gc := Geom_NewGeometryCollection(geometries, factory)
	mls := &Geom_MultiLineString{
		Geom_GeometryCollection: gc,
	}
	gc.child = mls
	return mls
}

func (mls *Geom_MultiLineString) GetDimension_BODY() int {
	return 1
}

func (mls *Geom_MultiLineString) HasDimension_BODY(dim int) bool {
	return dim == 1
}

func (mls *Geom_MultiLineString) GetBoundaryDimension_BODY() int {
	if mls.IsClosed() {
		return Geom_Dimension_False
	}
	return 0
}

func (mls *Geom_MultiLineString) GetGeometryType_BODY() string {
	return Geom_Geometry_TypeNameMultiLineString
}

// IsClosed reports whether all component LineStrings are closed.
// Returns false if the MultiLineString is empty.
func (mls *Geom_MultiLineString) IsClosed() bool {
	if mls.IsEmpty() {
		return false
	}
	for i := 0; i < len(mls.geometries); i++ {
		if !java.Cast[*Geom_LineString](mls.geometries[i]).IsClosed() {
			return false
		}
	}
	return true
}

// GetBoundary gets the boundary of this geometry.
// The boundary of a lineal geometry is always a zero-dimensional geometry
// (which may be empty).
func (mls *Geom_MultiLineString) GetBoundary_BODY() *Geom_Geometry {
	return Operation_NewBoundaryOp(mls.Geom_Geometry).GetBoundary()
}

// Reverse creates a MultiLineString in the reverse order to this object.
// Both the order of the component LineStrings and the order of their coordinate
// sequences are reversed.
func (mls *Geom_MultiLineString) Reverse() *Geom_MultiLineString {
	reversed := mls.Geom_Geometry.Reverse()
	return java.Cast[*Geom_MultiLineString](reversed)
}

func (mls *Geom_MultiLineString) Reverse_BODY() *Geom_Geometry {
	return mls.ReverseInternal().Geom_Geometry
}

func (mls *Geom_MultiLineString) ReverseInternal() *Geom_MultiLineString {
	lineStrings := make([]*Geom_LineString, len(mls.geometries))
	for i := range lineStrings {
		lineStrings[i] = java.Cast[*Geom_LineString](mls.geometries[i].Reverse())
	}
	return Geom_NewMultiLineString(lineStrings, mls.factory)
}

func (mls *Geom_MultiLineString) CopyInternal_BODY() *Geom_Geometry {
	lineStrings := make([]*Geom_LineString, len(mls.geometries))
	for i := range lineStrings {
		lineStrings[i] = java.Cast[*Geom_LineString](mls.geometries[i].Copy())
	}
	return Geom_NewMultiLineString(lineStrings, mls.factory).Geom_Geometry
}

func (mls *Geom_MultiLineString) EqualsExactWithTolerance_BODY(other *Geom_Geometry, tolerance float64) bool {
	if !mls.IsEquivalentClass(other) {
		return false
	}
	return mls.Geom_GeometryCollection.EqualsExactWithTolerance_BODY(other, tolerance)
}

func (mls *Geom_MultiLineString) GetTypeCode_BODY() int {
	return Geom_Geometry_TypeCodeMultiLineString
}

// isLineal implements the Lineal marker interface.
func (mls *Geom_MultiLineString) IsLineal() {}
