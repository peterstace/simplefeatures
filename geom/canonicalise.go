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
			if pt.Intersects(ln.AsGeometry()) {
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
		return GeometryCollection{}.AsGeometry(), nil
	case len(points) == 0:
		// Lines only.
		if len(lines) == 1 {
			return lines[0].AsGeometry(), nil
		}
		var lineStrings []LineString
		for _, ln := range lines {
			lnStr, err := NewLineStringFromSequence(ln.Coordinates())
			if err != nil {
				return Geometry{}, err
			}
			lineStrings = append(lineStrings, lnStr)
		}
		ctype := DimXY
		if len(lineStrings) > 0 {
			ctype = lineStrings[0].CoordinatesType()
		}
		mls, err := NewMultiLineString(lineStrings, ctype)
		return mls.AsGeometry(), err
	case len(lines) == 0:
		// Points only.
		if len(points) == 1 {
			return points[0].AsGeometry(), nil
		}
		ctype := DimXY
		if len(points) > 0 {
			ctype = points[0].CoordinatesType()
		}
		mp, err := NewMultiPoint(points, ctype)
		return mp.AsGeometry(), err
	default:
		all := make([]Geometry, len(points)+len(lines))
		for i, pt := range points {
			all[i] = pt.AsGeometry()
		}
		for i, ln := range lines {
			all[len(points)+i] = ln.AsGeometry()
		}
		ctype := DimXY
		if len(all) > 0 {
			ctype = all[0].CoordinatesType()
		}
		gc, err := NewGeometryCollection(all, ctype)
		return gc.AsGeometry(), err
	}
}

func dedupPoints(pts []Point) []Point {
	var dedup []Point
	seen := make(map[XY]bool)
	var haveEmpty bool
	for _, pt := range pts {
		xy, ok := pt.XY()
		if !ok {
			haveEmpty = true
		} else if !seen[xy] {
			dedup = append(dedup, pt)
			seen[xy] = true
		}
	}
	if haveEmpty {
		dedup = append(dedup, NewEmptyPoint(DimXY))
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
