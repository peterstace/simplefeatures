package geom

import (
	"math"
	"strconv"
	"strings"

	"github.com/peterstace/simplefeatures/rtree"
)

// Envelope is a generalised axis-aligned rectangle (also known as an Axis
// Aligned Bounding Box or Minimum Bounding Rectangle). It usually represents a
// 2D area with non-zero width and height. But it can also represent degenerate
// cases where the width or height (or both) are zero, or the envelope is
// empty. Its bounds are validated so as to not be NaN or +/- Infinity.
//
// An envelope can be thought of as being similar to a regular geometry, but
// can only represent an empty geometry, a single point, a horizontal or
// vertical line, or an axis aligned rectangle with some area.
//
// The Envelope zero value is the empty envelope. Envelopes are immutable after
// creation.
type Envelope struct {
	min, max XY
	nonEmpty bool
}

// NewEnvelope returns the smallest envelope that contains all provided XYs.
// It returns an error if any of the XYs contain NaN or +/- Infinity
// coordinates.
func NewEnvelope(xys ...XY) Envelope {
	var env Envelope
	for _, xy := range xys {
		env = env.ExpandToIncludeXY(xy)
	}
	return env
}

func newUncheckedEnvelope(minXY, maxXY XY) Envelope {
	return Envelope{minXY, maxXY, true}
}

// Validate checks if the Envelope is valid. The only validation rule is that
// the coordinates the Envelope was constructed from must not be NaN or +/-
// infinity. An empty Envelope is always valid.
func (e Envelope) Validate() error {
	if e.IsEmpty() {
		return nil
	}
	if err := e.min.validate(); err != nil {
		return wrap(err, "min coords invalid")
	}
	if err := e.max.validate(); err != nil {
		return wrap(err, "max coords invalid")
	}
	return nil
}

// IsEmpty returns true if and only if this envelope is empty.
func (e Envelope) IsEmpty() bool {
	return !e.nonEmpty
}

// IsPoint returns true if and only if this envelope represents a single point.
func (e Envelope) IsPoint() bool {
	return !e.IsEmpty() && e.min == e.max
}

// IsLine returns true if and only if this envelope represents a single line
// (which must be either vertical or horizontal).
func (e Envelope) IsLine() bool {
	return !e.IsEmpty() && (e.min.X == e.max.X) != (e.min.Y == e.max.Y)
}

// IsRectangle returns true if and only if this envelope represents a
// non-degenerate rectangle with some area.
func (e Envelope) IsRectangle() bool {
	return !e.IsEmpty() && e.min.X != e.max.X && e.min.Y != e.max.Y
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
		return e.min.AsPoint().AsGeometry()
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
	ring := NewLineString(seq)
	poly := NewPolygon([]LineString{ring})
	return poly.AsGeometry()
}

// Min returns the point in the envelope with the minimum X and Y values.
func (e Envelope) Min() Point {
	if e.IsEmpty() {
		return Point{}
	}
	return e.min.AsPoint()
}

// Max returns the point in the envelope with the maximum X and Y values.
func (e Envelope) Max() Point {
	if e.IsEmpty() {
		return Point{}
	}
	return e.max.AsPoint()
}

// MinMaxXYs returns the two XY values in the envelope that contain the minimum
// (first return value) and maximum (second return value) X and Y values in the
// envelope. The third return value is true if and only if the Envelope is
// non-empty and thus the first two return values are populated.
func (e Envelope) MinMaxXYs() (XY, XY, bool) {
	if e.IsEmpty() {
		return XY{}, XY{}, false
	}
	return e.min, e.max, true
}

// ExpandToIncludeXY returns the smallest envelope that contains all of the
// points in this envelope along with the provided point. It will produce an
// invalid envelope if any of the coordinates in the existing envelope or new
// XY contain NaN or +/- infinity.
func (e Envelope) ExpandToIncludeXY(xy XY) Envelope {
	if e.IsEmpty() {
		return newUncheckedEnvelope(xy, xy)
	}
	return newUncheckedEnvelope(
		XY{fastMin(e.min.X, xy.X), fastMin(e.min.Y, xy.Y)},
		XY{fastMax(e.max.X, xy.X), fastMax(e.max.Y, xy.Y)},
	)
}

// ExpandToIncludeEnvelope returns the smallest envelope that contains all of
// the points in this envelope and another envelope. It will produce an invalid
// envelope if min or max coordinates of either envelope contain NaN or +/-
// infinity.
func (e Envelope) ExpandToIncludeEnvelope(o Envelope) Envelope {
	if e.IsEmpty() {
		return o
	}
	if o.IsEmpty() {
		return e
	}
	return newUncheckedEnvelope(
		XY{fastMin(e.min.X, o.min.X), fastMin(e.min.Y, o.min.Y)},
		XY{fastMax(e.max.X, o.max.X), fastMax(e.max.Y, o.max.Y)},
	)
}

// Contains returns true if and only if this envelope contains the given XY. It
// always returns false in the case where the XY contains NaN or +/- Infinity
// coordinates.
func (e Envelope) Contains(p XY) bool {
	return !e.IsEmpty() &&
		p.validate() == nil &&
		p.X >= e.min.X && p.X <= e.max.X &&
		p.Y >= e.min.Y && p.Y <= e.max.Y
}

// Intersects returns true if and only if this envelope has any points in
// common with another envelope.
func (e Envelope) Intersects(o Envelope) bool {
	return !e.IsEmpty() && !o.IsEmpty() &&
		(e.min.X <= o.max.X) && (e.max.X >= o.min.X) &&
		(e.min.Y <= o.max.Y) && (e.max.Y >= o.min.Y)
}

// Center returns the center point of the envelope.
func (e Envelope) Center() Point {
	if e.IsEmpty() {
		return Point{}
	}
	return e.min.
		Add(e.max).
		Scale(0.5).
		AsPoint()
}

// Covers returns true if and only if this envelope entirely covers another
// envelope (i.e. every point in the other envelope is contained within this
// envelope). An envelope can only cover another if it is non-empty.
// Furthermore, an envelope can only be covered if it is non-empty.
func (e Envelope) Covers(o Envelope) bool {
	return !e.IsEmpty() && !o.IsEmpty() &&
		e.min.X <= o.min.X && e.min.Y <= o.min.Y &&
		e.max.X >= o.max.X && e.max.Y >= o.max.Y
}

// Width returns the difference between the maximum and minimum X coordinates
// of the envelope.
func (e Envelope) Width() float64 {
	if e.IsEmpty() {
		return 0
	}
	return e.max.X - e.min.X
}

// Height returns the difference between the maximum and minimum X coordinates
// of the envelope.
func (e Envelope) Height() float64 {
	if e.IsEmpty() {
		return 0
	}
	return e.max.Y - e.min.Y
}

// Area returns the area covered by the envelope.
func (e Envelope) Area() float64 {
	if e.IsEmpty() {
		return 0
	}
	return (e.max.X - e.min.X) * (e.max.Y - e.min.Y)
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
	dx := fastMax(0, fastMax(o.min.X-e.max.X, e.min.X-o.max.X))
	dy := fastMax(0, fastMax(o.min.Y-e.max.Y, e.min.Y-o.max.Y))
	return math.Sqrt(dx*dx + dy*dy), true
}

// TransformXY transforms this Envelope into another Envelope according to fn.
func (e Envelope) TransformXY(fn func(XY) XY) Envelope {
	u, v, ok := e.MinMaxXYs()
	if !ok {
		return Envelope{}
	}
	u = fn(u)
	v = fn(v)
	return newUncheckedEnvelope(
		XY{fastMin(u.X, v.X), fastMin(u.Y, v.Y)},
		XY{fastMax(u.X, v.X), fastMax(u.Y, v.Y)},
	)
}

// AsBox converts this Envelope to an rtree.Box.
func (e Envelope) AsBox() (rtree.Box, bool) {
	return rtree.Box{
		MinX: e.min.X,
		MinY: e.min.Y,
		MaxX: e.max.X,
		MaxY: e.max.Y,
	}, !e.IsEmpty()
}

// BoundingDiagonal returns the LineString that goes from the point returned by
// Min() to the point returned by Max(). If the envelope is degenerate and
// represents a single point, then a Point is returned instead of a LineString.
// If the Envelope is empty, then the empty Geometry (representing an empty
// GeometryCollection) is returned.
func (e Envelope) BoundingDiagonal() Geometry {
	if e.IsEmpty() {
		return Geometry{}
	}
	if e.IsPoint() {
		return e.min.AsPoint().AsGeometry()
	}

	coords := []float64{e.min.X, e.min.Y, e.max.X, e.max.Y}
	seq := NewSequence(coords, DimXY)
	return NewLineString(seq).AsGeometry()
}

// String implements the fmt.Stringer interface by printing the envelope in a
// pseudo-WKT style.
func (e Envelope) String() string {
	var sb strings.Builder
	sb.WriteString("ENVELOPE")
	if e.IsEmpty() {
		sb.WriteString(" EMPTY")
		return sb.String()
	}
	sb.WriteRune('(')
	add := func(f float64, r rune) {
		sb.WriteString(strconv.FormatFloat(f, 'f', -1, 64))
		sb.WriteRune(r)
	}
	add(e.min.X, ' ')
	add(e.min.Y, ',')
	add(e.max.X, ' ')
	add(e.max.Y, ')')
	return sb.String()
}
