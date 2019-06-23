package simplefeatures

import (
	"strings"
)

type XY struct {
	X, Y Scalar
}

func (xy XY) Equals(other XY) bool {
	return xyeq(xy, other)
}

type xyString string

func xykey(xy XY) xyString {
	var buf strings.Builder
	z := xy.X.AsRat(nil)
	buf.WriteString(z.String())
	buf.WriteByte(',')
	xy.Y.AsRat(z)
	buf.WriteString(z.String())
	return xyString(buf.String())
}

func xysub(a, b XY) XY {
	return XY{
		ssub(a.X, b.X),
		//a.X - b.X,
		ssub(a.Y, b.Y),
		//a.Y - b.Y,
	}
}

func xycross(a, b XY) Scalar {
	return ssub(smul(a.X, b.Y), smul(a.Y, b.X))
	//return a.X*b.Y - a.Y*b.X
}

func xyeq(a, b XY) bool {
	return seq(a.X, b.X) && seq(a.Y, b.Y)
}

type Coordinates struct {
	XY
	// TODO: Put optional Z and M here.
}

type OptionalCoordinates struct {
	Empty bool
	Value Coordinates
}
