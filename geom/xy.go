package geom

import "math"

// XY represents a pair of X and Y coordinates. This can either represent a
// location on the XY plane, or a 2D vector in the real vector space.
type XY struct {
	X, Y float64
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
	return (w.X * o.Y) - (w.Y * o.X)
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
	return math.Sqrt(w.Dot(w))
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
