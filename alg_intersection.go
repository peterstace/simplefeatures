package simplefeatures

import (
	"math"
)

func intersection(g1, g2 Geometry) Geometry {
	ln1, ok1 := g1.(Line)
	ln2, ok2 := g2.(Line)
	if ok1 && ok2 {
		return intersectLineWithLine(ln1, ln2)
	}
	panic("not implemented")
}

func intersectLineWithLine(n1, n2 Line) Geometry {
	a := n1.a.XY
	b := n1.b.XY
	c := n2.a.XY
	d := n2.b.XY

	if parallel := cross(sub(b, a), sub(d, c)) == 0; !parallel {
		e := (c.Y-d.Y)*(a.X-c.X) + (d.X-c.X)*(a.Y-c.Y)
		f := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)
		g := (a.Y-b.Y)*(a.X-c.X) + (b.X-a.X)*(a.Y-c.Y)
		h := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)
		// Division by zero is not possible, since the lines are not parallel.
		p := e / f
		q := g / h
		if p < 0 || p > 1 || q < 0 || q > 1 {
			// Intersection between lines occurs beyond line endpoints.
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

	if colinear := cross(sub(b, a), sub(d, a)) == 0; colinear {
		u := XY{
			math.Max(math.Min(a.X, b.X), math.Min(c.X, d.X)),
			math.Max(math.Min(a.Y, b.Y), math.Min(c.Y, d.Y)),
		}
		v := XY{
			math.Min(math.Max(a.X, b.X), math.Max(c.X, d.X)),
			math.Min(math.Max(a.Y, b.Y), math.Max(c.Y, d.Y)),
		}
		if u.X > v.X || u.Y > v.Y {
			// Line segments don't overlap at all.
			return NewGeometryCollection(nil)
		}
		if u == v {
			// Line segments overlap at a point.
			pt, err := NewPoint(u.X, u.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}
		// Line segments overlap over a line segment.
		ln, err := NewLine(Coordinates{u}, Coordinates{v})
		if err != nil {
			panic(err)
		}
		return ln
	}

	// Parrallel but not colinear, so cannot intersect anywhere.
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
