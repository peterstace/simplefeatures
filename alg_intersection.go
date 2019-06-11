package simplefeatures

import (
	"math"
)

func intersectLineWithLine(n1, n2 Line) Geometry {
	a := n1.a.XY
	b := n1.b.XY
	c := n2.a.XY
	d := n2.b.XY

	e := (c.Y-d.Y)*(a.X-c.X) + (d.X-c.X)*(a.Y-c.Y)
	f := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)
	g := (a.Y-b.Y)*(a.X-c.X) + (b.X-a.X)*(a.Y-c.Y)
	h := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)

	if f != 0 && h != 0 {
		p := e / f
		q := g / h
		if p < 0 || p > 1 || q < 0 || q > 1 {
			return NewGeometryCollection(nil)
		}
		pt, err := NewPoint(
			a.X+p*(b.X-a.X),
			a.Y+p*(b.Y-a.Y),
		)
		if err != nil {
			panic(err)
		}
		return pt
	}

	// They're parallel. But are they colinear?
	if cross(sub(b, a), sub(d, a)) == 0 {
		// The two segments are colinear, so the intersection is another line.
		u := XY{
			math.Max(math.Min(a.X, b.X), math.Min(c.X, d.X)),
			math.Max(math.Min(a.Y, b.Y), math.Min(c.Y, d.Y)),
		}
		v := XY{
			math.Min(math.Max(a.X, b.X), math.Max(c.X, d.X)),
			math.Min(math.Max(a.Y, b.Y), math.Max(c.Y, d.Y)),
		}
		if u.X > v.X || u.Y > v.Y {
			return NewGeometryCollection(nil)
		}
		if u == v {
			pt, err := NewPoint(u.X, u.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}
		ln, err := NewLine(Coordinates{u}, Coordinates{v})
		if err != nil {
			panic(err)
		}
		return ln
	}

	// Parrallel but not colinear. So cannot intersect anywhere.
	return NewGeometryCollection(nil)
}

func dot(a, b XY) float64 {
	return a.X*b.X + a.Y*b.Y
}

func sub(a, b XY) XY {
	return XY{a.X - b.X, a.Y - b.Y}
}

func cross(a, b XY) float64 {
	return a.X*b.Y - a.Y*b.X
}
