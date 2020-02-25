package geom

import "math"

type side int

const (
	interior side = -1
	boundary side = 0
	exterior side = +1
)

// TODO: check to see if all usages of pointRingSide are non-empty.

// pointRingSide checks the side of a ring that a point is on. It assumes that
// the input ring is actually a ring (i.e. closed and simple) and is non-empty.
func pointRingSide(pt XY, ring LineString) side {
	if !ring.IsClosed() {
		// We don't explicitly check for simplicity, since that's an expensive
		// operation. If a ring is closed, then that implies that it's also
		// non-empty.
		panic("pointRingSide called with non-closed ring")
	}

	ptg := NewPointC(Coordinates{pt})
	// find max x coordinate
	// TODO: should be able to use envelope for this
	maxX := ring.LineN(0).StartPoint().X
	for i := 0; i < ring.NumLines(); i++ {
		ln := ring.LineN(i)
		maxX = math.Max(maxX, ln.EndPoint().X)
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
