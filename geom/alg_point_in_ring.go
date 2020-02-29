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

	// find max x coordinate
	maxX := math.Inf(-1)
	for i := 0; i < n; i++ {
		maxX = math.Max(maxX, seq.GetXY(i).X)
	}
	if pt.X > maxX {
		return exterior
	}

	ptg := NewPointXY(pt)
	iter := newLineStringIterator(ring)
	for iter.next() {
		ln := iter.line()
		if hasIntersectionPointWithLine(ptg, ln) {
			return boundary
		}
	}

	ray, err := NewLineXY(pt, XY{maxX + 1, pt.Y})
	if err != nil {
		// Cannot occur because X coordinates are different.
		panic(err)
	}

	var count int
	iter = newLineStringIterator(ring)
	for iter.next() {
		seg := iter.line()
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
