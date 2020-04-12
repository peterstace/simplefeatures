package geom

import "math"

type side int

const (
	interior side = -1
	boundary side = 0
	exterior side = +1
)

// pointRingSide checks the side of a ring that a point is on. It assumes that
// the input ring is actually a ring (i.e. closed and simple) and is non-empty.
func pointRingSide(pt XY, ring LineString) side {
	if !ring.IsClosed() {
		// We don't explicitly check for simplicity, since that's an expensive
		// operation. If a ring is closed, then that implies that it's also
		// non-empty.
		panic("pointRingSide called with non-closed ring")
	}

	seq := ring.Coordinates()
	n := seq.Length()

	maxX := math.Inf(-1)
	for i := 0; i < n; i++ {
		maxX = math.Max(maxX, seq.GetXY(i).X)
		ln, ok := getLine(seq, i)
		if ok && ln.intersectsXY(pt) {
			return boundary
		}
	}
	if pt.X > maxX {
		return exterior
	}

	ray := line{pt, XY{maxX + 1, pt.Y}}
	var count int
	for i := 0; i < n; i++ {
		ln, ok := getLine(seq, i)
		if !ok {
			continue
		}
		inter := ln.intersectLine(ray)
		if inter.empty {
			continue
		}
		if inter.ptA != inter.ptB {
			continue
		}
		if inter.ptA == ln.a || inter.ptA == ln.b {
			otherY := ln.a.Y
			if inter.ptA == ln.a {
				otherY = ln.b.Y
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
