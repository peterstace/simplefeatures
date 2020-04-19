package geom

import (
	"math"

	"github.com/peterstace/simplefeatures/rtree"
)

type side int

const (
	interior side = -1
	boundary side = 0
	exterior side = +1
)

// relatePointToRing checks the side of a ring that a point is on. It assumes that
// the input ring is actually a ring (i.e. closed and simple) and is non-empty.
func relatePointToRing(pt XY, ring LineString) side {
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
		if incrementCountPointInRing(pt, ray, ln) {
			count++
		}
	}
	if count%2 == 1 {
		return interior
	}
	return exterior
}

func incrementCountPointInRing(pt XY, ray, iterLine line) bool {
	inter := ray.intersectLine(iterLine)
	if inter.empty {
		return false
	}
	if inter.ptA != inter.ptB {
		return false
	}
	if inter.ptA == iterLine.a || inter.ptA == iterLine.b {
		otherY := iterLine.a.Y
		if inter.ptA == iterLine.a {
			otherY = iterLine.b.Y
		}
		return otherY < pt.Y
	}
	return true
}

func relatePointToPolygon(pt XY, polyBoundary indexedLines) side {
	var onBoundary bool
	ptBox := rtree.Box{MinX: pt.X, MinY: pt.Y, MaxX: pt.X, MaxY: pt.Y}
	polyBoundary.tree.Search(ptBox, func(i int) error {
		ln := polyBoundary.lines[i]
		if ln.intersectsXY(pt) {
			onBoundary = true
			return rtree.Stop
		}
		return nil
	})
	if onBoundary {
		return boundary
	}

	extent, ok := polyBoundary.tree.Extent()
	if !ok {
		return exterior
	}
	ray := line{pt, XY{extent.MaxX + 1, pt.Y}}

	var count int
	polyBoundary.tree.Search(toBox(ray.envelope()), func(i int) error {
		ln := polyBoundary.lines[i]
		if incrementCountPointInRing(pt, ray, ln) {
			count++
		}
		return nil
	})
	if count%2 == 1 {
		return interior
	}
	return exterior
}
