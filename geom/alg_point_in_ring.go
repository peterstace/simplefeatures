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
	maxX := ring.LineN(0).StartPoint().XY().X
	for i := 0; i < ring.NumLines(); i++ {
		ln := ring.LineN(i)
		maxX = math.Max(maxX, ln.EndPoint().XY().X)
		if hasIntersectionPointWithLine(ptg, ln) {
			return boundary
		}
	}
	if pt.X > maxX {
		return exterior
	}

	ray, err := NewLineC(Coordinates{pt}, Coordinates{XY{maxX + 1, pt.Y}})
	if err != nil {
		// Cannot occur because X coordinates are different.
		panic(err)
	}

	var count int
	for i := 0; i < ring.NumLines(); i++ {
		seg := ring.LineN(i)
		inter := intersectLineWithLineNoAlloc(seg, ray)
		if inter.empty {
			continue
		}
		if inter.ptA != inter.ptB {
			continue
		}
		if inter.ptA == seg.a.XY || inter.ptA == seg.b.XY {
			otherY := seg.a.Y
			if inter.ptA == seg.a.XY {
				otherY = seg.b.Y
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
