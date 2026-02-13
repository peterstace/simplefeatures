package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geom_Point represents a single point.
//
// A Point is topologically valid if and only if:
//   - the coordinate which defines it (if any) is a valid coordinate
//     (i.e. does not have an NaN X or Y ordinate)
type Geom_Point struct {
	*Geom_Geometry
	coordinates Geom_CoordinateSequence
	child       java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (p *Geom_Point) GetChild() java.Polymorphic {
	return p.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *Geom_Point) GetParent() java.Polymorphic {
	return p.Geom_Geometry
}

// Geom_NewPointWithPrecisionModel constructs a Point with the given coordinate.
//
// Deprecated: Use GeometryFactory instead.
func Geom_NewPointWithPrecisionModel(coordinate *Geom_Coordinate, precisionModel *Geom_PrecisionModel, srid int) *Geom_Point {
	factory := Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, srid)
	var coords Geom_CoordinateSequence
	if coordinate != nil {
		coords = factory.GetCoordinateSequenceFactory().CreateFromCoordinates([]*Geom_Coordinate{coordinate})
	} else {
		coords = factory.GetCoordinateSequenceFactory().CreateFromCoordinates([]*Geom_Coordinate{})
	}
	return geom_newPointInternal(coords, factory)
}

// Geom_NewPoint constructs a Point from a CoordinateSequence.
// The coordinates parameter contains the single coordinate on which to base this Point,
// or nil to create the empty geometry.
func Geom_NewPoint(coordinates Geom_CoordinateSequence, factory *Geom_GeometryFactory) *Geom_Point {
	return geom_newPointInternal(coordinates, factory)
}

func geom_newPointInternal(coordinates Geom_CoordinateSequence, factory *Geom_GeometryFactory) *Geom_Point {
	geom := &Geom_Geometry{factory: factory}
	p := &Geom_Point{
		Geom_Geometry: geom,
	}
	geom.child = p
	p.init(coordinates)
	return p
}

func (p *Geom_Point) init(coordinates Geom_CoordinateSequence) {
	if coordinates == nil {
		coordinates = p.GetFactory().GetCoordinateSequenceFactory().CreateFromCoordinates([]*Geom_Coordinate{})
	}
	Util_Assert_IsTrue(coordinates.Size() <= 1)
	p.coordinates = coordinates
}

func (p *Geom_Point) GetCoordinates_BODY() []*Geom_Coordinate {
	if p.IsEmpty() {
		return []*Geom_Coordinate{}
	}
	return []*Geom_Coordinate{p.GetCoordinate()}
}

func (p *Geom_Point) GetNumPoints_BODY() int {
	if p.IsEmpty() {
		return 0
	}
	return 1
}

func (p *Geom_Point) IsEmpty_BODY() bool {
	return p.coordinates.Size() == 0
}

func (p *Geom_Point) IsSimple_BODY() bool {
	return true
}

func (p *Geom_Point) GetDimension_BODY() int {
	return 0
}

func (p *Geom_Point) GetBoundaryDimension_BODY() int {
	return Geom_Dimension_False
}

// GetX returns the X ordinate value.
// Panics if called on an empty Point.
func (p *Geom_Point) GetX() float64 {
	if p.GetCoordinate() == nil {
		panic("getX called on empty Point")
	}
	return p.GetCoordinate().X
}

// GetY returns the Y ordinate value.
// Panics if called on an empty Point.
func (p *Geom_Point) GetY() float64 {
	if p.GetCoordinate() == nil {
		panic("getY called on empty Point")
	}
	return p.GetCoordinate().Y
}

// GetCoordinate_BODY returns the Coordinate or nil if this Point is empty.
func (p *Geom_Point) GetCoordinate_BODY() *Geom_Coordinate {
	if p.coordinates.Size() != 0 {
		return p.coordinates.GetCoordinate(0)
	}
	return nil
}

func (p *Geom_Point) GetGeometryType_BODY() string {
	return Geom_Geometry_TypeNamePoint
}

// GetBoundary_BODY gets the boundary of this geometry.
// Zero-dimensional geometries have no boundary by definition,
// so an empty GeometryCollection is returned.
func (p *Geom_Point) GetBoundary_BODY() *Geom_Geometry {
	gc := p.GetFactory().CreateGeometryCollection()
	return gc.Geom_Geometry
}

func (p *Geom_Point) ComputeEnvelopeInternal_BODY() *Geom_Envelope {
	if p.IsEmpty() {
		return Geom_NewEnvelope()
	}
	env := Geom_NewEnvelope()
	env.ExpandToIncludeXY(p.coordinates.GetX(0), p.coordinates.GetY(0))
	return env
}

func (p *Geom_Point) EqualsExactWithTolerance_BODY(other *Geom_Geometry, tolerance float64) bool {
	if !p.IsEquivalentClass(other) {
		return false
	}
	if p.IsEmpty() && other.IsEmpty() {
		return true
	}
	if p.IsEmpty() != other.IsEmpty() {
		return false
	}
	otherPoint := java.Cast[*Geom_Point](other)
	return p.EqualCoordinate(otherPoint.GetCoordinate(), p.GetCoordinate(), tolerance)
}

func (p *Geom_Point) ApplyCoordinateFilter_BODY(filter Geom_CoordinateFilter) {
	if p.IsEmpty() {
		return
	}
	filter.Filter(p.GetCoordinate())
}

func (p *Geom_Point) ApplyCoordinateSequenceFilter_BODY(filter Geom_CoordinateSequenceFilter) {
	if p.IsEmpty() {
		return
	}
	filter.Filter(p.coordinates, 0)
	if filter.IsGeometryChanged() {
		p.GeometryChanged()
	}
}

func (p *Geom_Point) ApplyGeometryFilter_BODY(filter Geom_GeometryFilter) {
	filter.Filter(p.Geom_Geometry)
}

func (p *Geom_Point) Apply_BODY(filter Geom_GeometryComponentFilter) {
	filter.Filter(p.Geom_Geometry)
}

func (p *Geom_Point) CopyInternal_BODY() *Geom_Geometry {
	point := Geom_NewPoint(p.coordinates.Copy(), p.factory)
	return point.Geom_Geometry
}

func (p *Geom_Point) Reverse_BODY() *Geom_Geometry {
	return p.ReverseInternal_BODY()
}

func (p *Geom_Point) ReverseInternal_BODY() *Geom_Geometry {
	point := p.GetFactory().CreatePointFromCoordinateSequence(p.coordinates.Copy())
	return point.Geom_Geometry
}

func (p *Geom_Point) Normalize_BODY() {
	// A Point is always in normalized form.
}

func (p *Geom_Point) CompareToSameClass_BODY(other any) int {
	otherPoint := java.Cast[*Geom_Point](other.(*Geom_Geometry))
	return p.GetCoordinate().CompareTo(otherPoint.GetCoordinate())
}

func (p *Geom_Point) CompareToSameClassWithComparator_BODY(other any, comp *Geom_CoordinateSequenceComparator) int {
	otherPoint := java.Cast[*Geom_Point](other.(*Geom_Geometry))
	return comp.Compare(p.coordinates, otherPoint.coordinates)
}

func (p *Geom_Point) GetTypeCode_BODY() int {
	return Geom_Geometry_TypeCodePoint
}

// GetCoordinateSequence returns the CoordinateSequence containing the coordinates.
func (p *Geom_Point) GetCoordinateSequence() Geom_CoordinateSequence {
	return p.coordinates
}

// isPuntal implements the Puntal marker interface.
func (p *Geom_Point) IsPuntal() {}
