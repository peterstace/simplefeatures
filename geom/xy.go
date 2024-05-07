package geom

import (
	"math"

	"github.com/peterstace/simplefeatures/rtree"
)

// XY represents a pair of X and Y coordinates. This can either represent a
// location on the XY plane, or a 2D vector in the real vector space.
type XY struct {
	X, Y float64
}

// validate checks if the XY value contains NaN, -inf, or +inf.
func (w XY) validate() error {
	if math.IsNaN(w.X) || math.IsNaN(w.Y) {
		return violateNaN.errAtXY(w)
	}
	if math.IsInf(w.X, 0) || math.IsInf(w.Y, 0) {
		return violateInf.errAtXY(w)
	}
	return nil
}

// AsPoint is a convenience function to convert this XY value into a Point
// geometry.
func (w XY) AsPoint() Point {
	coords := Coordinates{XY: w, Type: DimXY}
	return NewPoint(coords)
}

// uncheckedEnvelope is a convenience function to convert this XY value into
// a (degenerate) envelope that represents a single XY location (i.e. a zero
// area envelope). It may be used internally when the caller is sure that the
// XY value doesn't come directly from outline the library without first being
// validated.
func (w XY) uncheckedEnvelope() Envelope {
	return newUncheckedEnvelope(w, w)
}

// Sub returns the result of subtracting the other XY from this XY (in the same
// manner as vector subtraction).
func (w XY) Sub(o XY) XY {
	return XY{
		w.X - o.X,
		w.Y - o.Y,
	}
}

// Add returns the result of adding this XY to another XY (in the same manner
// as vector addition).
func (w XY) Add(o XY) XY {
	return XY{
		w.X + o.X,
		w.Y + o.Y,
	}
}

// Scale returns the XY where the X and Y have been scaled by s.
func (w XY) Scale(s float64) XY {
	return XY{
		w.X * s,
		w.Y * s,
	}
}

// Cross returns the 2D cross product of this and another XY. This is defined
// as the 'z' coordinate of the regular 3D cross product.
func (w XY) Cross(o XY) float64 {
	// Avoid fused multiply-add by explicitly converting intermediate products
	// to float64. This ensures that the cross product is *exactly* zero for
	// all linearly dependent inputs.
	return float64(w.X*o.Y) - float64(w.Y*o.X)
}

// Midpoint returns the midpoint of this and another XY.
func (w XY) Midpoint(o XY) XY {
	return w.Add(o).Scale(0.5)
}

// Dot returns the dot product of this and another XY.
func (w XY) Dot(o XY) float64 {
	return w.X*o.X + w.Y*o.Y
}

// Unit treats the XY as a vector, and scales it to have unit length.
func (w XY) Unit() XY {
	return w.Scale(1 / w.Length())
}

// Length treats XY as a vector, and returns its length.
func (w XY) Length() float64 {
	return math.Sqrt(w.lengthSq())
}

// lengthSq treats XY as a vector, and returns its squared length.
func (w XY) lengthSq() float64 {
	return w.Dot(w)
}

// Less gives an ordering on XYs. If two XYs have different X values, then the
// one with the lower X value is ordered before the one with the higher X
// value. If the X values are then same, then the Y values are used (the lower
// Y value comes first).
func (w XY) Less(o XY) bool {
	if w.X != o.X {
		return w.X < o.X
	}
	return w.Y < o.Y
}

func (w XY) distanceTo(o XY) float64 {
	return math.Sqrt(w.distanceSquaredTo(o))
}

func (w XY) distanceSquaredTo(o XY) float64 {
	delta := o.Sub(w)
	return delta.Dot(delta)
}

func (w XY) box() rtree.Box {
	return rtree.Box{
		MinX: w.X,
		MinY: w.Y,
		MaxX: w.X,
		MaxY: w.Y,
	}
}

// rotateCCW90 treats the XY as a vector, rotating it 90 degrees in a counter
// clockwise direction (assuming a right handed/positive orientation).
func (w XY) rotateCCW90() XY {
	return XY{
		X: -w.Y,
		Y: w.X,
	}
}

// rotate180 treats the XY as a vector, rotating it 180 degrees.
func (w XY) rotate180() XY {
	return XY{
		X: -w.X,
		Y: -w.Y,
	}
}

// identity returns the XY unchanged.
func (w XY) identity() XY {
	return w
}

// proj returns the projection of w onto o.
func (w XY) proj(o XY) XY {
	return o.Scale(w.Dot(o) / o.Dot(o))
}
