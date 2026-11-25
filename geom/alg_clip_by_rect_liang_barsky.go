package geom

func clipLineStringByRect(ls LineString, rect Envelope) Geometry {
	emptyLine := LineString{}.AsGeometry()

	seq := ls.Coordinates()
	n := seq.Length()

	lo, hi, ok := rect.MinMaxXYs()
	if !ok {
		return emptyLine
	}

	var chains [][]XY
	var cur []XY

	for i := 0; i < n-1; i++ {
		a := seq.GetXY(i)
		b := seq.GetXY(i + 1)

		ca, cb, ok := clipSegment(a, b, lo, hi)
		if !ok {
			if len(cur) > 0 {
				chains = append(chains, cur)
				cur = nil
			}
			continue
		}

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

// clipSegment uses the Liang-Barsky algorithm to compute the portion of segment
// a->b that lies inside the axis-aligned rectangle from lo to hi. It returns the
// clipped endpoints ca and cb, and false if no segment of positive length
// survives.
//
// An endpoint that a rect edge introduced (i.e. where the segment was cut by
// that edge) is snapped exactly onto the edge's coordinate, rather than being
// left a few ULPs off by interpolation. This matches the exact-boundary
// guarantee that the polygon clipper provides, and keeps results stable under
// this package's zero-tolerance equality.
func clipSegment(a, b, lo, hi XY) (ca, cb XY, ok bool) {
	tMin := 0.0
	tMax := 1.0
	dx := b.X - a.X
	dy := b.Y - a.Y

	// Each rect edge contributes a Liang-Barsky constraint.
	edges := [4]clipEdge{
		{-dx, a.X - lo.X, true, lo.X},  // left
		{dx, hi.X - a.X, true, hi.X},   // right
		{-dy, a.Y - lo.Y, false, lo.Y}, // bottom
		{dy, hi.Y - a.Y, false, hi.Y},  // top
	}

	// minEdge/maxEdge record which edge last bound tMin/tMax, or -1 if the
	// endpoint is still the original a/b (inside the rect) and must not be
	// snapped.
	minEdge, maxEdge := -1, -1
	for i, e := range edges {
		if e.p == 0 {
			if e.q < 0 {
				return XY{}, XY{}, false
			}
			continue
		}
		t := e.q / e.p
		if e.p < 0 {
			if t > tMin {
				tMin = t
				minEdge = i
			}
		} else {
			if t < tMax {
				tMax = t
				maxEdge = i
			}
		}
		if tMin >= tMax {
			return XY{}, XY{}, false
		}
	}

	ca = snapToEdge(lerpXY(a, b, tMin), edges, minEdge)
	cb = snapToEdge(lerpXY(a, b, tMax), edges, maxEdge)
	return ca, cb, true
}

// clipEdge describes one Liang-Barsky constraint, corresponding to a single rect
// edge. p and q are the Liang-Barsky numerator/denominator for the edge. axisX
// reports whether the edge fixes the X coordinate (left/right) or the Y
// coordinate (bottom/top); val is the boundary coordinate to snap that axis to.
type clipEdge struct {
	p, q  float64
	axisX bool
	val   float64
}

// snapToEdge fixes the axis controlled by edge i exactly onto its boundary
// coordinate, leaving the interpolated point xy unchanged when i is -1 (the
// endpoint lies inside the rect and was not introduced by an edge).
func snapToEdge(xy XY, edges [4]clipEdge, i int) XY {
	if i < 0 {
		return xy
	}
	if edges[i].axisX {
		xy.X = edges[i].val
	} else {
		xy.Y = edges[i].val
	}
	return xy
}
