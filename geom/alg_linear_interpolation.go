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

	p0 := l.seq.Get(idx + 0)
	p1 := l.seq.Get(idx + 1)

	partial := frac * l.total
	if idx-1 >= 0 {
		partial -= l.cumulative[idx-1]
	}
	partial /= p0.XY.distanceTo(p1.XY)

	return interpolateCoords(p0, p1, partial).AsPoint()
}

func lerp(a, b, ratio float64) float64 {
	return a + ratio*(b-a)
}

func interpolateCoords(c0, c1 Coordinates, frac float64) Coordinates {
	return Coordinates{
		XY: XY{
			X: lerp(c0.X, c1.X, frac),
			Y: lerp(c0.Y, c1.Y, frac),
		},
		Z:    lerp(c0.Z, c1.Z, frac),
		M:    lerp(c0.M, c1.M, frac),
		Type: c0.Type & c1.Type,
	}
}
