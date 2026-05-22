package geom

func clipLineStringByRect(ls LineString, rect Envelope) Geometry {
	emptyLine := LineString{}.AsGeometry()

	seq := ls.Coordinates()
	n := seq.Length()
	if n == 0 {
		return emptyLine
	}

	min, max, ok := rect.MinMaxXYs()
	if !ok {
		return emptyLine
	}

	var chains [][]XY
	var cur []XY

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

		ca := lerpXY(a, b, tMin)
		cb := lerpXY(a, b, tMax)

		if len(cur) > 0 && cur[len(cur)-1] == ca {
			cur = append(cur, cb)
		} else {
			if len(cur) > 0 {
				chains = append(chains, cur)
			}
			cur = []XY{ca, cb}
		}
	}
	if len(cur) > 0 {
		chains = append(chains, cur)
	}

	if len(chains) == 0 {
		return emptyLine
	}

	lines := make([]LineString, len(chains))
	for i, c := range chains {
		lines[i] = NewLineString(xysToSeq(c))
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
