package geom

func canonicalPointsAndLines(points []Point, lines []Line) (Geometry, error) {
	// Deduplicate.
	points = dedupPoints(points)
	lines = dedupLines(lines)

	// Remove any points that are covered by lines.
	var newPoints []Point
	for _, pt := range points {
		hasInter := false
		for _, ln := range lines {
			if pt.Intersects(ln) {
				hasInter = true
				break
			}
		}
		if !hasInter {
			newPoints = append(newPoints, pt)
		}
	}
	points = newPoints

	switch {
	case len(points) == 0 && len(lines) == 0:
		return NewGeometryCollection(nil), nil
	case len(points) == 0:
		// Lines only.
		if len(lines) == 1 {
			return lines[0], nil
		}
		var lineStrings []LineString
		for _, ln := range lines {
			lnStr, err := NewLineStringC(ln.Coordinates())
			if err != nil {
				return nil, err
			}
			lineStrings = append(lineStrings, lnStr)
		}
		return NewMultiLineString(lineStrings), nil
	case len(lines) == 0:
		// Points only.
		if len(points) == 1 {
			return points[0], nil
		}
		return NewMultiPoint(points), nil
	default:
		all := make([]Geometry, len(points)+len(lines))
		for i, pt := range points {
			all[i] = pt
		}
		for i, ln := range lines {
			all[len(points)+i] = ln
		}
		return NewGeometryCollection(all), nil
	}
}

func dedupPoints(pts []Point) []Point {
	var dedup []Point
	seen := make(map[XY]bool)
	for _, pt := range pts {
		xy := pt.XY()
		if !seen[xy] {
			dedup = append(dedup, pt)
			seen[xy] = true
		}
	}
	return dedup
}

func dedupLines(lines []Line) []Line {
	type xyxy struct {
		a, b XY
	}
	var dedup []Line
	seen := make(map[xyxy]bool)
	for _, ln := range lines {
		k := xyxy{ln.a.XY, ln.b.XY}
		if !k.a.Less(k.b) {
			k.a, k.b = k.b, k.a
		}
		if !seen[k] {
			dedup = append(dedup, ln)
			seen[k] = true
		}
	}
	return dedup
}
