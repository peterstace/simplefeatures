package geom

import "math"

type side int

const (
	interior side = -1
	boundary side = 0
	exterior side = +1
)

// pointRingSide checks the side of a ring that a point is on. It assumes that
// the input ring is actually a ring (i.e. closed and simple).
func pointRingSide(pt XY, ring LineString) side {
	ptg := NewPointC(Coordinates{pt})
	// find max x coordinate
	// TODO: should be able to use envelope for this
	maxX := ring.lines[0].a.X
	for _, ln := range ring.lines {
		maxX = math.Max(maxX, ln.b.X)
		if !ToGeometry(mustIntersection(ln, ptg)).IsEmpty() {
			return boundary
		}
	}
	if pt.X > maxX {
		return exterior
	}

	ray := must(NewLineC(Coordinates{pt}, Coordinates{XY{maxX + 1, pt.Y}})).(Line)
	var count int
	for _, seg := range ring.lines {
		inter := ToGeometry(mustIntersection(seg, ray))
		if inter.IsEmpty() {
			continue
		}
		if inter.Dimension() == 1 {
			continue
		}
		ep1 := NewPointC(seg.a)
		ep2 := NewPointC(seg.b)
		if inter.EqualsExact(ep1.AsGeometry()) || inter.EqualsExact(ep2.AsGeometry()) {
			otherY := ep1.coords.Y
			if inter.EqualsExact(ep1.AsGeometry()) {
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
