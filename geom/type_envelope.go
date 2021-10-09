package geom

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/rtree"
)

// Envelope is a generalised axis-aligned rectangle (also known as an Axis
// Aligned Bounding Box or Minimum Bounding Rectangle). It usually represents a
// 2D area with non-zero width and height. But it can also represent degenerate
// cases where the width or height (or both) are zero, or the envelope is
// empty. Its bounds are validated so as to not be NaN or +/- Infinity.
//
// An envelope can be thought of as as being similar to a regular geometry, but
// can only represent an empty geometry, a single point, a horizontal or
// vertical line, or an axis aligned rectangle with some area.
//
// The Envelope zero value is the empty envelope. Envelopes are immutable after
// creation.
type Envelope struct {
	// nanXORMinX is the bit pattern of "min X" XORed with the bit pattern of
	// NaN. This is so that when Envelope has its zero value, the logical value
	// of "min X" is NaN. The logical value of "min X" being NaN is used to
	// signify that the Envelope is empty.
	nanXORMinX uint64

	minY float64
	maxX float64
	maxY float64
}

var nan = math.Float64bits(math.NaN())

// encodeFloat64WithNaN encodes a float64 by XORing it with NaN.
func encodeFloat64WithNaN(f float64) uint64 {
	return math.Float64bits(f) ^ nan
}

// minX decodes the logical value ("min X") of nanXORMinX.
func (e Envelope) minX() float64 {
	return math.Float64frombits(e.nanXORMinX ^ nan)
}

// NewEnvelope returns the smallest envelope that contains all provided XYs.
// It returns an error if any of the XYs contain NaN or +/- Infinity
// coordinates.
func NewEnvelope(xys []XY) (Envelope, error) {
	var env Envelope
	for _, xy := range xys {
		var err error
		env, err = env.ExtendToIncludeXY(xy)
		if err != nil {
			return Envelope{}, err
		}
	}
	return env, nil
}

func newUncheckedEnvelope(min, max XY) Envelope {
	return Envelope{
		nanXORMinX: encodeFloat64WithNaN(min.X),
		minY:       min.Y,
		maxX:       max.X,
		maxY:       max.Y,
	}
}

func (e Envelope) min() XY {
	return XY{e.minX(), e.minY}
}

func (e Envelope) max() XY {
	return XY{e.maxX, e.maxY}
}

// IsEmpty returns true if and only if this envelope is empty.
func (e Envelope) IsEmpty() bool {
	return math.IsNaN(e.minX())
}

// IsPoint returns true if and only if this envelope represents a single point.
func (e Envelope) IsPoint() bool {
	return !e.IsEmpty() && e.min() == e.max()
}

// IsLine returns true if and only if this envelope represents a single line
// (which must be either vertical or horizontal).
func (e Envelope) IsLine() bool {
	return !e.IsEmpty() && (e.minX() == e.maxX) != (e.minY == e.maxY)
}

// IsRectangle returns true if and only if this envelope represents a
// non-degenerate rectangle with some area.
func (e Envelope) IsRectangle() bool {
	return !e.IsEmpty() && e.minX() != e.maxX && e.minY != e.maxY
}

// AsGeometry returns the envelope as a Geometry. In the regular case where the
// envelope covers some area, then a Polygon geometry is returned. In
// degenerate cases where the envelope only covers a line or a point, a
// LineString or Point geometry is returned. In the case of an empty envelope,
// the zero value Geometry is returned (representing an empty
// GeometryCollection).
func (e Envelope) AsGeometry() Geometry {
	switch {
	case e.IsEmpty():
		return Geometry{}
	case e.IsPoint():
		return e.min().asUncheckedPoint().AsGeometry()
	case e.IsLine():
		ln := line{e.min(), e.max()}
		return ln.asLineString().AsGeometry()
	}

	minX := e.minX()
	floats := [...]float64{
		minX, e.minY,
		minX, e.maxY,
		e.maxX, e.maxY,
		e.maxX, e.minY,
		minX, e.minY,
	}
	seq := NewSequence(floats[:], DimXY)
	ls, err := NewLineString(seq)
	if err != nil {
		panic(fmt.Sprintf("constructing geometry from envelope: %v", err))
	}
	poly, err := NewPolygon([]LineString{ls})
	if err != nil {
		panic(fmt.Sprintf("constructing geometry from envelope: %v", err))
	}
	return poly.AsGeometry()
}

// Min returns the point in the envelope with the minimum X and Y values.
func (e Envelope) Min() Point {
	if e.IsEmpty() {
		return Point{}
	}
	return e.min().asUncheckedPoint()
}

// Max returns the point in the envelope with the maximum X and Y values.
func (e Envelope) Max() Point {
	if e.IsEmpty() {
		return Point{}
	}
	return e.max().asUncheckedPoint()
}

// ExtendToIncludeXY returns the smallest envelope that contains all of the
// points in this envelope along with the provided point. It gives an error if
// the XY contains NaN or +/- Infinite coordinates.
func (e Envelope) ExtendToIncludeXY(xy XY) (Envelope, error) {
	if err := xy.validate(); err != nil {
		return Envelope{}, err
	}
	return e.uncheckedExtend(xy), nil
}

// uncheckedExtend extends the envelope in the same manner as
// ExtendToIncludeXY but doesn't validate the XY. It should only be used
// when the XY doesn't come directly from user input.
func (e Envelope) uncheckedExtend(xy XY) Envelope {
	if e.IsEmpty() {
		return newUncheckedEnvelope(xy, xy)
	}
	return newUncheckedEnvelope(
		XY{fastMin(e.minX(), xy.X), fastMin(e.minY, xy.Y)},
		XY{fastMax(e.maxX, xy.X), fastMax(e.maxY, xy.Y)},
	)
}

// ExpandToIncludeEnvelope returns the smallest envelope that contains all of
// the points in this envelope and another envelope.
func (e Envelope) ExpandToIncludeEnvelope(o Envelope) Envelope {
	if e.IsEmpty() {
		return o
	}
	if o.IsEmpty() {
		return e
	}
	return newUncheckedEnvelope(
		XY{fastMin(e.minX(), o.minX()), fastMin(e.minY, o.minY)},
		XY{fastMax(e.maxX, o.maxX), fastMax(e.maxY, o.maxY)},
	)
}

// Contains returns true if and only if this envelope contains the given XY. It
// always returns false in the case where the XY contains NaN or +/- Infinity
// coordinates.
func (e Envelope) Contains(p XY) bool {
	return !e.IsEmpty() &&
		p.validate() == nil &&
		p.X >= e.minX() && p.X <= e.maxX &&
		p.Y >= e.minY && p.Y <= e.maxY
}

// Intersects returns true if and only if this envelope has any points in
// common with another envelope.
func (e Envelope) Intersects(o Envelope) bool {
	return !e.IsEmpty() && !o.IsEmpty() &&
		(e.minX() <= o.maxX) && (e.maxX >= o.minX()) &&
		(e.minY <= o.maxY) && (e.maxY >= o.minY)
}

// Center returns the center point of the envelope.
func (e Envelope) Center() Point {
	if e.IsEmpty() {
		return Point{}
	}
	return e.min().
		Add(e.max()).
		Scale(0.5).
		asUncheckedPoint()
}

// Covers returns true if and only if this envelope entirely covers another
// envelope (i.e. every point in the other envelope is contained within this
// envelope). An envelope can only cover another if it is non-empty.
// Furthermore, an envelope can only be covered if it is non-empty.
func (e Envelope) Covers(o Envelope) bool {
	return !e.IsEmpty() && !o.IsEmpty() &&
		e.minX() <= o.minX() && e.minY <= o.minY &&
		e.maxX >= o.maxX && e.maxY >= o.maxY
}

// Width returns the difference between the maximum and minimum X coordinates
// of the envelope.
func (e Envelope) Width() float64 {
	if e.IsEmpty() {
		return 0
	}
	return e.maxX - e.minX()
}

// Height returns the difference between the maximum and minimum X coordinates
// of the envelope.
func (e Envelope) Height() float64 {
	if e.IsEmpty() {
		return 0
	}
	return e.maxY - e.minY
}

// Area returns the area covered by the envelope.
func (e Envelope) Area() float64 {
	if e.IsEmpty() {
		return 0
	}
	return (e.maxX - e.minX()) * (e.maxY - e.minY)
}

// Distance calculates the shortest distance between this envelope and another
// envelope, which both must be non-empty for the distance to be well-defined
// (indicated by the bool return being true).  If the envelopes are both
// non-empty and intersect with each other, the distance between them is still
// well-defined, but zero.
func (e Envelope) Distance(o Envelope) (float64, bool) {
	if e.IsEmpty() || o.IsEmpty() {
		return 0, false
	}
	dx := fastMax(0, fastMax(o.minX()-e.maxX, e.minX()-o.maxX))
	dy := fastMax(0, fastMax(o.minY-e.maxY, e.minY-o.maxY))
	return math.Sqrt(dx*dx + dy*dy), true
}

func (e Envelope) box() (rtree.Box, bool) {
	return rtree.Box{
		MinX: e.minX(),
		MinY: e.minY,
		MaxX: e.maxX,
		MaxY: e.maxY,
	}, !e.IsEmpty()
}
