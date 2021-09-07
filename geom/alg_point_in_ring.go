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
	seq := ring.Coordinates()
	n := seq.Length()

	var count int
	for i := 0; i < n; i++ {
		ln, ok := getLine(seq, i)
		if !ok {
			continue
		}
		crossing, onLine := hasCrossing(pt, ln)
		if onLine {
			return boundary
		}
		if crossing {
			count++
		}
	}
	if count%2 == 0 {
		return exterior
	}
	return interior
}

func hasCrossing(pt XY, ln line) (crossing, onLine bool) {
	lower, upper := ln.a, ln.b
	if lower.Y > upper.Y {
		lower, upper = upper, lower
	}
	o := orientation(lower, upper, pt)

	crossing = pt.Y >= lower.Y && pt.Y < upper.Y && o == rightTurn
	onLine = ln.uncheckedEnvelope().Contains(pt) && o == collinear
	return
}

func relatePointToPolygon(pt XY, polyBoundary indexedLines) side {
	box := rtree.Box{
		MinX: math.Inf(-1),
		MinY: pt.Y,
		MaxX: pt.X,
		MaxY: pt.Y,
	}
	var onBound bool
	var count int
	polyBoundary.tree.RangeSearch(box, func(i int) error {
		ln := polyBoundary.lines[i]
		crossing, onLine := hasCrossing(pt, ln)
		if onLine {
			onBound = true
			return rtree.Stop
		}
		if crossing {
			count++
		}
		return nil
	})
	if onBound {
		return boundary
	}
	if count%2 == 1 {
		return interior
	}
	return exterior
}
