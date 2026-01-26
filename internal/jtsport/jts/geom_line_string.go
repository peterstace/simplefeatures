package jts

import (
	"fmt"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

type Geom_LineString struct {
	*Geom_Geometry
	points Geom_CoordinateSequence
	child  java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (ls *Geom_LineString) GetChild() java.Polymorphic {
	return ls.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (ls *Geom_LineString) GetParent() java.Polymorphic {
	return ls.Geom_Geometry
}

const Geom_LineString_MINIMUM_VALID_SIZE = 2

func Geom_NewLineStringWithPrecisionModel(points []*Geom_Coordinate, precisionModel *Geom_PrecisionModel, srid int) *Geom_LineString {
	factory := Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, srid)
	return Geom_NewLineString(factory.GetCoordinateSequenceFactory().CreateFromCoordinates(points), factory)
}

func Geom_NewLineString(points Geom_CoordinateSequence, factory *Geom_GeometryFactory) *Geom_LineString {
	geom := &Geom_Geometry{}
	ls := &Geom_LineString{
		Geom_Geometry: geom,
	}
	geom.child = ls
	ls.factory = factory
	ls.lineString_init(points)
	return ls
}

func (ls *Geom_LineString) lineString_init(points Geom_CoordinateSequence) {
	if points == nil {
		points = ls.GetFactory().GetCoordinateSequenceFactory().CreateFromCoordinates([]*Geom_Coordinate{})
	}
	if points.Size() > 0 && points.Size() < Geom_LineString_MINIMUM_VALID_SIZE {
		panic(fmt.Sprintf("Invalid number of points in LineString (found %d - must be 0 or >= %d)",
			points.Size(), Geom_LineString_MINIMUM_VALID_SIZE))
	}
	ls.points = points
}

func (ls *Geom_LineString) GetCoordinates_BODY() []*Geom_Coordinate {
	return ls.points.ToCoordinateArray()
}

func (ls *Geom_LineString) GetCoordinateSequence() Geom_CoordinateSequence {
	return ls.points
}

func (ls *Geom_LineString) GetCoordinateN(n int) *Geom_Coordinate {
	return ls.points.GetCoordinate(n)
}

func (ls *Geom_LineString) GetCoordinate_BODY() *Geom_Coordinate {
	if ls.IsEmpty_BODY() {
		return nil
	}
	return ls.points.GetCoordinate(0)
}

func (ls *Geom_LineString) GetDimension_BODY() int {
	return 1
}

func (ls *Geom_LineString) GetBoundaryDimension_BODY() int {
	if ls.IsClosed() {
		return Geom_Dimension_False
	}
	return 0
}

func (ls *Geom_LineString) IsEmpty_BODY() bool {
	return ls.points.Size() == 0
}

func (ls *Geom_LineString) GetNumPoints_BODY() int {
	return ls.points.Size()
}

func (ls *Geom_LineString) GetPointN(n int) *Geom_Point {
	return ls.GetFactory().CreatePointFromCoordinate(ls.points.GetCoordinate(n))
}

func (ls *Geom_LineString) GetStartPoint() *Geom_Point {
	if ls.IsEmpty_BODY() {
		return nil
	}
	return ls.GetPointN(0)
}

func (ls *Geom_LineString) GetEndPoint() *Geom_Point {
	if ls.IsEmpty_BODY() {
		return nil
	}
	return ls.GetPointN(ls.GetNumPoints() - 1)
}

func (ls *Geom_LineString) IsClosed() bool {
	if impl, ok := java.GetLeaf(ls).(interface{ IsClosed_BODY() bool }); ok {
		return impl.IsClosed_BODY()
	}
	return ls.IsClosed_BODY()
}

func (ls *Geom_LineString) IsClosed_BODY() bool {
	if ls.IsEmpty_BODY() {
		return false
	}
	return ls.GetCoordinateN(0).Equals2D(ls.GetCoordinateN(ls.GetNumPoints() - 1))
}

func (ls *Geom_LineString) IsRing() bool {
	return ls.IsClosed() && ls.IsSimple()
}

func (ls *Geom_LineString) GetGeometryType_BODY() string {
	return Geom_Geometry_TypeNameLineString
}

func (ls *Geom_LineString) GetLength_BODY() float64 {
	return Algorithm_Length_OfLine(ls.points)
}

func (ls *Geom_LineString) GetBoundary_BODY() *Geom_Geometry {
	return Operation_NewBoundaryOp(ls.Geom_Geometry).GetBoundary()
}

func (ls *Geom_LineString) Reverse() *Geom_LineString {
	reversed := ls.Geom_Geometry.Reverse()
	return java.Cast[*Geom_LineString](reversed)
}

func (ls *Geom_LineString) ReverseInternal_BODY() *Geom_Geometry {
	seq := ls.points.Copy()
	Geom_CoordinateSequences_Reverse(seq)
	lineString := ls.GetFactory().CreateLineStringFromCoordinateSequence(seq)
	return lineString.Geom_Geometry
}

func (ls *Geom_LineString) IsCoordinate(pt *Geom_Coordinate) bool {
	for i := 0; i < ls.points.Size(); i++ {
		if ls.points.GetCoordinate(i).Equals(pt) {
			return true
		}
	}
	return false
}

func (ls *Geom_LineString) ComputeEnvelopeInternal_BODY() *Geom_Envelope {
	if ls.IsEmpty_BODY() {
		return Geom_NewEnvelope()
	}
	return ls.points.ExpandEnvelope(Geom_NewEnvelope())
}

// IsEquivalentClass_BODY overrides the base implementation to treat LinearRing
// and LineString as equivalent types. This matches Java JTS behavior where
// LinearRing extends LineString.
func (ls *Geom_LineString) IsEquivalentClass_BODY(other *Geom_Geometry) bool {
	return java.InstanceOf[*Geom_LineString](other)
}

func (ls *Geom_LineString) EqualsExactWithTolerance_BODY(other *Geom_Geometry, tolerance float64) bool {
	if !ls.IsEquivalentClass(other) {
		return false
	}
	// Handle both LineString and LinearRing (which embeds LineString).
	var otherLineString *Geom_LineString
	switch o := java.GetLeaf(other).(type) {
	case *Geom_LineString:
		otherLineString = o
	case *Geom_LinearRing:
		otherLineString = o.Geom_LineString
	default:
		return false
	}
	if ls.points.Size() != otherLineString.points.Size() {
		return false
	}
	for i := 0; i < ls.points.Size(); i++ {
		if !ls.Geom_Geometry.EqualCoordinate(ls.points.GetCoordinate(i), otherLineString.points.GetCoordinate(i), tolerance) {
			return false
		}
	}
	return true
}

func (ls *Geom_LineString) ApplyCoordinateFilter_BODY(filter Geom_CoordinateFilter) {
	for i := 0; i < ls.points.Size(); i++ {
		filter.Filter(ls.points.GetCoordinate(i))
	}
}

func (ls *Geom_LineString) ApplyCoordinateSequenceFilter_BODY(filter Geom_CoordinateSequenceFilter) {
	if ls.points.Size() == 0 {
		return
	}
	for i := 0; i < ls.points.Size(); i++ {
		filter.Filter(ls.points, i)
		if filter.IsDone() {
			break
		}
	}
	if filter.IsGeometryChanged() {
		ls.GeometryChanged()
	}
}

func (ls *Geom_LineString) ApplyGeometryFilter_BODY(filter Geom_GeometryFilter) {
	filter.Filter(ls.Geom_Geometry)
}

func (ls *Geom_LineString) Apply_BODY(filter Geom_GeometryComponentFilter) {
	filter.Filter(ls.Geom_Geometry)
}

func (ls *Geom_LineString) CopyInternal_BODY() *Geom_Geometry {
	lineString := Geom_NewLineString(ls.points.Copy(), ls.factory)
	return lineString.Geom_Geometry
}

func (ls *Geom_LineString) Normalize_BODY() {
	for i := 0; i < ls.points.Size()/2; i++ {
		j := ls.points.Size() - 1 - i
		if !ls.points.GetCoordinate(i).Equals(ls.points.GetCoordinate(j)) {
			if ls.points.GetCoordinate(i).CompareTo(ls.points.GetCoordinate(j)) > 0 {
				copy := ls.points.Copy()
				Geom_CoordinateSequences_Reverse(copy)
				ls.points = copy
			}
			return
		}
	}
}

func (ls *Geom_LineString) CompareToSameClass_BODY(o any) int {
	line := java.Cast[*Geom_LineString](o.(*Geom_Geometry))
	i := 0
	j := 0
	for i < ls.points.Size() && j < line.points.Size() {
		comparison := ls.points.GetCoordinate(i).CompareTo(line.points.GetCoordinate(j))
		if comparison != 0 {
			return comparison
		}
		i++
		j++
	}
	if i < ls.points.Size() {
		return 1
	}
	if j < line.points.Size() {
		return -1
	}
	return 0
}

func (ls *Geom_LineString) CompareToSameClassWithComparator_BODY(o any, comp *Geom_CoordinateSequenceComparator) int {
	// o might be *Geom_LineString or *Geom_LinearRing.
	var otherPoints Geom_CoordinateSequence
	switch s := java.GetLeaf(o.(*Geom_Geometry)).(type) {
	case *Geom_LineString:
		otherPoints = s.points
	case *Geom_LinearRing:
		otherPoints = s.Geom_LineString.points
	default:
		panic("unexpected type in CompareToSameClassWithComparator_BODY")
	}
	return comp.Compare(ls.points, otherPoints)
}

func (ls *Geom_LineString) GetTypeCode_BODY() int {
	return Geom_Geometry_TypeCodeLineString
}

func (ls *Geom_LineString) IsLineal() {}
