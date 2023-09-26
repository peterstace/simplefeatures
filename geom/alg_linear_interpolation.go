package geom

import (
	"math"
	"sort"
)

type linearInterpolator struct {
	seq        Sequence
	cumulative []float64
	total      float64
}

func newLinearInterpolator(seq Sequence) linearInterpolator {
	n := seq.Length()
	if n == 0 {
		panic("empty seq in newLinearInterpolator")
	}
	var total float64
	cumulative := make([]float64, n-1)
	for i := 0; i < n-1; i++ {
		total += seq.GetXY(i).distanceTo(seq.GetXY(i + 1))
		cumulative[i] = total
	}
	return linearInterpolator{seq, cumulative, total}
}

func (l linearInterpolator) interpolate(frac float64) Point {
	frac = math.Max(0, math.Min(1, frac))
	idx := sort.SearchFloat64s(l.cumulative, frac*l.total)
	if idx == l.seq.Length() {
		return l.seq.Get(idx - 1).AsPoint()
	}

	p0 := l.seq.Get(idx)
	p1 := l.seq.Get(idx + 1)

	partial := frac * l.total
	if idx-1 >= 0 {
		partial -= l.cumulative[idx-1]
	}
	partial /= p0.XY.distanceTo(p1.XY)

	return Coordinates{
		XY: XY{
			X: lerp(p0.X, p1.X, partial),
			Y: lerp(p0.Y, p1.Y, partial),
		},
		Z:    lerp(p0.Z, p1.Z, partial),
		M:    lerp(p0.M, p1.M, partial),
		Type: l.seq.CoordinatesType(),
	}.AsPoint()
}

func lerp(a, b, ratio float64) float64 {
	return a + ratio*(b-a)
}
