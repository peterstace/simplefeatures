package jts

import (
	"fmt"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geom_LinearRing models an OGC SFS LinearRing.
// A LinearRing is a LineString which is both closed and simple.
// In other words, the first and last coordinate in the ring must be equal,
// and the ring must not self-intersect.
// Either orientation of the ring is allowed.
//
// A ring must have either 0 or 3 or more points.
// The first and last points must be equal (in 2D).
// If these conditions are not met, the constructors panic with an
// IllegalArgumentException. A ring with 3 points is invalid, because it is
// collapsed and thus has a self-intersection. It is allowed to be constructed
// so that it can be represented, and repaired if needed.
type Geom_LinearRing struct {
	*Geom_LineString
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (lr *Geom_LinearRing) GetChild() java.Polymorphic {
	return lr.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (lr *Geom_LinearRing) GetParent() java.Polymorphic {
	return lr.Geom_LineString
}

// Geom_LinearRing_MinimumValidSize is the minimum number of vertices allowed in a
// valid non-empty ring. Empty rings with 0 vertices are also valid.
const Geom_LinearRing_MinimumValidSize = 3

// Geom_NewLinearRing constructs a LinearRing with the vertices specified by the
// given CoordinateSequence.
func Geom_NewLinearRing(points Geom_CoordinateSequence, factory *Geom_GeometryFactory) *Geom_LinearRing {
	ls := Geom_NewLineString(points, factory)
	lr := &Geom_LinearRing{
		Geom_LineString: ls,
	}
	ls.child = lr
	lr.validateConstruction()
	return lr
}

// Geom_NewLinearRingWithCoordinates constructs a LinearRing with the given points.
// Deprecated: Use GeometryFactory instead.
func Geom_NewLinearRingWithCoordinates(points []*Geom_Coordinate, precisionModel *Geom_PrecisionModel, srid int) *Geom_LinearRing {
	factory := Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, srid)
	return geom_newLinearRingWithCoordinatesAndFactory(points, factory)
}

// geom_newLinearRingWithCoordinatesAndFactory is used to avoid deprecation.
func geom_newLinearRingWithCoordinatesAndFactory(points []*Geom_Coordinate, factory *Geom_GeometryFactory) *Geom_LinearRing {
	seq := factory.GetCoordinateSequenceFactory().CreateFromCoordinates(points)
	return Geom_NewLinearRing(seq, factory)
}

func (lr *Geom_LinearRing) validateConstruction() {
	if !lr.IsEmpty() && !lr.IsClosed() {
		panic("Points of LinearRing do not form a closed linestring")
	}
	size := lr.GetCoordinateSequence().Size()
	if size >= 1 && size < Geom_LinearRing_MinimumValidSize {
		panic(fmt.Sprintf("Invalid number of points in LinearRing (found %d - must be 0 or >= %d)",
			size, Geom_LinearRing_MinimumValidSize))
	}
}

// GetBoundaryDimension_BODY returns Dimension_False, since by definition
// LinearRings do not have a boundary.
func (lr *Geom_LinearRing) GetBoundaryDimension_BODY() int {
	return Geom_Dimension_False
}

// IsClosed_BODY tests whether this ring is closed.
// Empty rings are closed by definition.
func (lr *Geom_LinearRing) IsClosed_BODY() bool {
	if lr.IsEmpty() {
		return true
	}
	return lr.Geom_LineString.IsClosed_BODY()
}

// GetGeometryType_BODY returns the geometry type.
func (lr *Geom_LinearRing) GetGeometryType_BODY() string {
	return Geom_Geometry_TypeNameLinearRing
}

// GetTypeCode_BODY returns the type code.
func (lr *Geom_LinearRing) GetTypeCode_BODY() int {
	return Geom_Geometry_TypeCodeLinearRing
}

// CopyInternal_BODY creates a deep copy of this LinearRing.
func (lr *Geom_LinearRing) CopyInternal_BODY() *Geom_Geometry {
	copied := Geom_NewLinearRing(lr.points.Copy(), lr.factory)
	return copied.Geom_Geometry
}

// Reverse returns a reversed copy of this LinearRing.
func (lr *Geom_LinearRing) Reverse() *Geom_LinearRing {
	return java.Cast[*Geom_LinearRing](lr.Geom_LineString.Geom_Geometry.Reverse())
}

// ReverseInternal_BODY creates a reversed copy of this LinearRing.
func (lr *Geom_LinearRing) ReverseInternal_BODY() *Geom_Geometry {
	seq := lr.points.Copy()
	Geom_CoordinateSequences_Reverse(seq)
	reversed := lr.GetFactory().CreateLinearRingFromCoordinateSequence(seq)
	return reversed.Geom_Geometry
}
