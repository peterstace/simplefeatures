package geom

import "math"

type side int

const (
	interior side = -1
	boundary side = 0
	exterior side = +1
)

func pointRingSide(pt XY, ring LinearRing) side {
	ptg := NewPointC(Coordinates{pt})
	// find max x coordinate
	// TODO: should be able to use envelope for this
	maxX := ring.ls.lines[0].a.X
	for _, ln := range ring.ls.lines {
		maxX = math.Max(maxX, ln.b.X)
		if !ln.Intersection(ptg).IsEmpty() {
			return boundary
		}
	}
	if pt.X > maxX {
		return exterior
	}

	ray := must(NewLineC(Coordinates{pt}, Coordinates{XY{maxX + 1, pt.Y}})).(Line)
	var count int
	for _, seg := range ring.ls.lines {
		inter := seg.Intersection(ray)
		if inter.IsEmpty() {
			continue
		}
		if inter.Dimension() == 1 {
			continue
		}
		ep1 := NewPointC(seg.a)
		ep2 := NewPointC(seg.b)
		if inter.Equals(ep1) || inter.Equals(ep2) {
			otherY := ep1.coords.Y
			if inter.Equals(ep1) {
				otherY = ep2.coords.Y
			}
			if otherY < pt.Y {
				count++
			}
		} else {
			count++
		}
	}
	if count%2 == 1 {
		return interior
	}
	return exterior
}
