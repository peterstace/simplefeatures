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

// lerp calculates the linear interpolation (or extrapolation) between a and b
// at t. Mathematically, the result is:
//
//	a + t×(b-a)
//
// or equivalently:
//
//	(1-t)×a + t×b
//
// In IEEE floating point math, the implementation can't be that simple due to
// rounding and over/underflow of intermediate results. Instead, we used the
// hybrid approach described by Davis Herring in [1]. It's is much more complex
// than the naive approach, but fixes a lot of edge cases that would otherwise
// occur.
//
// [1]: https://www.open-std.org/jtc1/sc22/wg21/docs/papers/2018/p0811r2.html
func lerp(a, b, t float64) float64 {
	if a <= 0 && b >= 0 || a >= 0 && b <= 0 {
		return t*b + (1-t)*a
	}
	if t == 1 {
		return b
	}
	x := a + t*(b-a)
	if (t > 1) == (b > a) {
		return math.Max(b, x)
	}
	return math.Min(b, x)
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
