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
	full bool
	min  XY
	max  XY
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

// IsEmpty returns true if and only if this envelope is empty.
func (e Envelope) IsEmpty() bool {
	return !e.full
}

// IsPoint returns true if and only if this envelope represents a single point.
func (e Envelope) IsPoint() bool {
	return e.full && e.min == e.max
}

// IsLine returns true if and only if this envelope represents a single line
// (which must be either vertical or horizontal).
func (e Envelope) IsLine() bool {
	return e.full && (e.min.X == e.max.X) != (e.min.Y == e.max.Y)
}

// IsRectangle returns true if and only if this envelope represents a
// non-degenerate rectangle with some area.
func (e Envelope) IsRectangle() bool {
	return e.full && e.min.X != e.max.X && e.min.Y != e.max.Y
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
		return e.min.asUncheckedPoint().AsGeometry()
	case e.IsLine():
		ln := line{e.min, e.max}
		return ln.asLineString().AsGeometry()
	}

	floats := [...]float64{
		e.min.X, e.min.Y,
		e.min.X, e.max.Y,
		e.max.X, e.max.Y,
		e.max.X, e.min.Y,
		e.min.X, e.min.Y,
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
	if !e.full {
		return Point{}
	}
	return e.min.asUncheckedPoint()
}

// Max returns the point in the envelope with the maximum X and Y values.
func (e Envelope) Max() Point {
	if !e.full {
		return Point{}
	}
	return e.max.asUncheckedPoint()
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
		return Envelope{full: true, min: xy, max: xy}
	}
	return Envelope{
		full: true,
		min:  XY{fastMin(e.min.X, xy.X), fastMin(e.min.Y, xy.Y)},
		max:  XY{fastMax(e.max.X, xy.X), fastMax(e.max.Y, xy.Y)},
	}
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
	return Envelope{
		full: true,
		min:  XY{fastMin(e.min.X, o.min.X), fastMin(e.min.Y, o.min.Y)},
		max:  XY{fastMax(e.max.X, o.max.X), fastMax(e.max.Y, o.max.Y)},
	}
}

// Contains returns true if and only if this envelope contains the given XY. It
// always returns false in the case where the XY contains NaN or +/- Infinity
// coordinates.
func (e Envelope) Contains(p XY) bool {
	return e.full &&
		p.validate() == nil &&
		p.X >= e.min.X && p.X <= e.max.X &&
		p.Y >= e.min.Y && p.Y <= e.max.Y
}

// Intersects returns true if and only if this envelope has any points in
// common with another envelope.
func (e Envelope) Intersects(o Envelope) bool {
	return e.full && o.full &&
		(e.min.X <= o.max.X) && (e.max.X >= o.min.X) &&
		(e.min.Y <= o.max.Y) && (e.max.Y >= o.min.Y)
}

// Center returns the center point of the envelope.
func (e Envelope) Center() Point {
	if !e.full {
		return Point{}
	}
	return e.min.Add(e.max).Scale(0.5).asUncheckedPoint()
}

// Covers returns true if and only if this envelope entirely covers another
// envelope (i.e. every point in the other envelope is contained within this
// envelope). An envelope can only cover another if it is non-empty.
// Furthermore, an envelope can only be covered if it is non-empty.
func (e Envelope) Covers(o Envelope) bool {
	return e.full && o.full &&
		e.min.X <= o.min.X && e.min.Y <= o.min.Y &&
		e.max.X >= o.max.X && e.max.Y >= o.max.Y
}

// Width returns the difference between the maximum and minimum X coordinates
// of the envelope.
func (e Envelope) Width() float64 {
	return e.max.X - e.min.X
}

// Height returns the difference between the maximum and minimum X coordinates
// of the envelope.
func (e Envelope) Height() float64 {
	return e.max.Y - e.min.Y
}

// Area returns the area covered by the envelope.
func (e Envelope) Area() float64 {
	return (e.max.X - e.min.X) * (e.max.Y - e.min.Y)
}

// Distance calculates the shortest distance between this envelope and another
// envelope, which both must be non-empty for the distance to be well-defined
// (indicated by the bool return being true).  If the envelopes are both
// non-empty and intersect with each other, the distance between them is still
// well-defined, but zero.
func (e Envelope) Distance(o Envelope) (float64, bool) {
	if !e.full || !o.full {
		return 0, false
	}
	dx := fastMax(0, fastMax(o.min.X-e.max.X, e.min.X-o.max.X))
	dy := fastMax(0, fastMax(o.min.Y-e.max.Y, e.min.Y-o.max.Y))
	return math.Sqrt(dx*dx + dy*dy), true
}

func (e Envelope) box() (rtree.Box, bool) {
	return rtree.Box{
		MinX: e.min.X,
		MinY: e.min.Y,
		MaxX: e.max.X,
		MaxY: e.max.Y,
	}, e.full
}
