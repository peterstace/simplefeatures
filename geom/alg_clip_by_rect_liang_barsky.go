package geom

func clipLineStringByRect(ls LineString, rect Envelope) Geometry {
	seq := ls.Coordinates()
	n := seq.Length()
	if n == 0 {
		return ls.AsGeometry()
	}

	min, max, ok := rect.MinMaxXYs()
	if !ok {
		return NewLineString(NewSequence(nil, seq.CoordinatesType())).AsGeometry()
	}

	ctype := seq.CoordinatesType()
	dim := ctype.Dimension()

	var chains [][]float64
	var cur []float64

	for i := 0; i < n-1; i++ {
		a := seq.GetXY(i)
		b := seq.GetXY(i + 1)

		tMin, tMax, ok := clipSegmentParams(a, b, min, max)
		if !ok {
			if len(cur) > 0 {
				chains = append(chains, cur)
				cur = nil
			}
			continue
		}

		ca := interpolateSeqCoord(seq, i, i+1, tMin)
		cb := interpolateSeqCoord(seq, i, i+1, tMax)

		if len(cur) > 0 && cur[len(cur)-dim] == ca.X && cur[len(cur)-dim+1] == ca.Y {
			cur = cb.appendFloat64s(cur)
		} else {
			if len(cur) > 0 {
				chains = append(chains, cur)
			}
			cur = ca.appendFloat64s(nil)
			cur = cb.appendFloat64s(cur)
		}
	}
	if len(cur) > 0 {
		chains = append(chains, cur)
	}

	if len(chains) == 0 {
		return NewLineString(NewSequence(nil, ctype)).AsGeometry()
	}

	lines := make([]LineString, len(chains))
	for i, c := range chains {
		lines[i] = NewLineString(NewSequence(c, ctype))
	}

	if len(lines) == 1 {
		return lines[0].AsGeometry()
	}
	return NewMultiLineString(lines).AsGeometry()
}

// clipSegmentParams uses the Liang-Barsky algorithm to compute the parametric
// range [tMin, tMax] of segment a->b that lies inside the axis-aligned
// rectangle from min to max. It returns false if no segment of positive length
// survives.
func clipSegmentParams(a, b, min, max XY) (float64, float64, bool) {
	tMin := 0.0
	tMax := 1.0
	dx := b.X - a.X
	dy := b.Y - a.Y
	for _, pq := range [4][2]float64{
		{-dx, a.X - min.X}, // left
		{dx, max.X - a.X},  // right
		{-dy, a.Y - min.Y}, // bottom
		{dy, max.Y - a.Y},  // top
	} {
		p, q := pq[0], pq[1]
		if p == 0 {
			if q < 0 {
				return 0, 0, false
			}
			continue
		}
		t := q / p
		if p < 0 {
			if t > tMin {
				tMin = t
			}
		} else {
			if t < tMax {
				tMax = t
			}
		}
		if tMin >= tMax {
			return 0, 0, false
		}
	}
	return tMin, tMax, true
}

// interpolateSeqCoord linearly interpolates between coordinates at index i and
// j in seq at parameter t (0=coord i, 1=coord j). Delegates to
// [interpolateCoords] which uses a numerically robust lerp.
func interpolateSeqCoord(seq Sequence, i, j int, t float64) Coordinates {
	if t == 0 {
		return seq.Get(i)
	}
	if t == 1 {
		return seq.Get(j)
	}
	return interpolateCoords(seq.Get(i), seq.Get(j), t)
}
