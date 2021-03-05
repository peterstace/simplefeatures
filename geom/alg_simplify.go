package geom

import "errors"

// Simplify returns a simplified version of the geometry using the
// Ramer-Douglas-Peucker algorithm.
func Simplify(g Geometry, threshold float64) (Geometry, error) {
	if !g.IsLineString() {
		return Geometry{}, errors.New("not implemented")
	}

	seq := g.AsLineString().Coordinates()
	n := seq.Length()
	xys := make([]XY, n)
	for i := 0; i < n; i++ {
		xys[i] = seq.GetXY(i)
	}

	xys = ramerDouglasPeucker(xys, threshold)
	var floats []float64
	for _, xy := range xys {
		floats = append(floats, xy.X, xy.Y)
	}
	seq = NewSequence(floats, DimXY)
	if seq.Length() > 0 && !hasAtLeast2DistinctPoints(seq) {
		return LineString{}.AsGeometry(), nil
	}
	ls, err := NewLineString(seq)
	return ls.AsGeometry(), err
}

// TODO: handle Z and M.

func ramerDouglasPeucker(seq []XY, threshold float64) []XY {
	if len(seq) <= 2 {
		return seq
	}

	n := len(seq)
	first, last := seq[0], seq[n-1]

	var calcDist func(pt XY) float64
	if first != last {
		calcDist = func(pt XY) float64 {
			return perpendicularDistance(pt, line{first, last})
		}
	} else {
		calcDist = func(pt XY) float64 {
			return pt.Sub(first).Length()
		}
	}

	var maxDist float64
	var maxDistIdx int
	for i := 0; i < n; i++ {
		if dist := calcDist(seq[i]); dist > maxDist {
			maxDist = dist
			maxDistIdx = i
		}
	}

	if maxDist <= threshold {
		return []XY{first, last}
	}

	h1 := ramerDouglasPeucker(seq[:maxDistIdx+1], threshold)
	h2 := ramerDouglasPeucker(seq[maxDistIdx:], threshold)
	return append(h1, h2[1:]...)
}

// perpendicularDistance is the distance from p to the infinite line going
// through ln.
func perpendicularDistance(p XY, ln line) float64 {
	aSubP := ln.a.Sub(p)
	unit := ln.b.Sub(ln.a).Scale(1 / ln.length())
	perpendicular := aSubP.Sub(unit.Scale(aSubP.Dot(unit)))
	return perpendicular.Length()
}
