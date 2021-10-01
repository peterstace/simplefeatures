package geom

// Simplify returns a simplified version of the geometry using the
// Ramer-Douglas-Peucker algorithm. Sometimes a simplified geometry can become
// invalid, in which case an error is returned rather than attempting to fix
// the geometry. Validation of the result can be skipped by making use of the
// geometry constructor options.
func Simplify(g Geometry, threshold float64, opts ...ConstructorOption) (Geometry, error) {
	s := simplifier{threshold, opts}
	switch g.gtype {
	case TypeGeometryCollection:
		gc, err := s.simplifyGeometryCollection(g.AsGeometryCollection())
		return gc.AsGeometry(), wrapSimplified(err)
	case TypePoint:
		return g, nil
	case TypeLineString:
		ls, err := s.simplifyLineString(g.AsLineString())
		return ls.AsGeometry(), wrapSimplified(err)
	case TypePolygon:
		poly, err := s.simplifyPolygon(g.AsPolygon())
		return poly.AsGeometry(), wrapSimplified(err)
	case TypeMultiPoint:
		return g, nil
	case TypeMultiLineString:
		mls, err := s.simplifyMultiLineString(g.AsMultiLineString())
		return mls.AsGeometry(), wrapSimplified(err)
	case TypeMultiPolygon:
		mp, err := s.simplifyMultiPolygon(g.AsMultiPolygon())
		return mp.AsGeometry(), wrapSimplified(err)
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

type simplifier struct {
	threshold float64
	opts      []ConstructorOption
}

func (s simplifier) simplifyLineString(ls LineString) (LineString, error) {
	seq := ls.Coordinates()
	floats := s.ramerDouglasPeucker(nil, seq)
	seq = NewSequence(floats, seq.CoordinatesType())
	if seq.Length() > 0 && !hasAtLeast2DistinctPointsInSeq(seq) {
		return LineString{}, nil
	}
	return NewLineString(seq, s.opts...)
}

func (s simplifier) simplifyMultiLineString(mls MultiLineString) (MultiLineString, error) {
	n := mls.NumLineStrings()
	lss := make([]LineString, 0, n)
	for i := 0; i < n; i++ {
		ls := mls.LineStringN(i)
		ls, err := s.simplifyLineString(ls)
		if err != nil {
			return MultiLineString{}, err
		}
		if !ls.IsEmpty() {
			lss = append(lss, ls)
		}
	}
	return NewMultiLineString(lss, s.opts...), nil
}

func (s simplifier) simplifyPolygon(poly Polygon) (Polygon, error) {
	exterior, err := s.simplifyLineString(poly.ExteriorRing())
	if err != nil {
		return Polygon{}, err
	}

	// If we don't have at least 4 coordinates, then we can't form a ring, and
	// the polygon has collapsed either to a point or a single linear element.
	// Both cases are represented by an empty polygon.
	if exterior.Coordinates().Length() < 4 {
		return Polygon{}, nil
	}

	n := poly.NumInteriorRings()
	rings := make([]LineString, 0, n+1)
	rings = append(rings, exterior)
	for i := 0; i < n; i++ {
		interior, err := s.simplifyLineString(poly.InteriorRingN(i))
		if err != nil {
			return Polygon{}, err
		}
		if interior.IsRing() {
			rings = append(rings, interior)
		}
	}
	return NewPolygon(rings, s.opts...)
}

func (s simplifier) simplifyMultiPolygon(mp MultiPolygon) (MultiPolygon, error) {
	n := mp.NumPolygons()
	polys := make([]Polygon, 0, n)
	for i := 0; i < n; i++ {
		poly, err := s.simplifyPolygon(mp.PolygonN(i))
		if err != nil {
			return MultiPolygon{}, err
		}
		if !poly.IsEmpty() {
			polys = append(polys, poly)
		}
	}
	return NewMultiPolygon(polys, s.opts...)
}

func (s simplifier) simplifyGeometryCollection(gc GeometryCollection) (GeometryCollection, error) {
	n := gc.NumGeometries()
	geoms := make([]Geometry, n)
	for i := 0; i < n; i++ {
		var err error
		geoms[i], err = Simplify(gc.GeometryN(i), s.threshold)
		if err != nil {
			return GeometryCollection{}, err
		}
	}
	return NewGeometryCollection(geoms, s.opts...), nil
}

func (s simplifier) ramerDouglasPeucker(dst []float64, seq Sequence) []float64 {
	if seq.Length() <= 2 {
		return seq.appendAllPoints(dst)
	}

	start := 0
	end := seq.Length() - 1

	for start < end {
		dst = seq.appendPoint(dst, start)
		newEnd := end
		for {
			var maxDist float64
			var maxDistIdx int
			for i := start + 1; i < newEnd; i++ {
				if d := perpendicularDistance(
					seq.GetXY(i),
					seq.GetXY(start),
					seq.GetXY(newEnd),
				); d > maxDist {
					maxDistIdx = i
					maxDist = d
				}
			}
			if maxDist <= s.threshold {
				break
			}
			newEnd = maxDistIdx
		}
		start = newEnd
	}
	dst = seq.appendPoint(dst, end)
	return dst
}

// perpendicularDistance is the distance from 'p' to the infinite line going
// through 'a' and 'b'. If 'a' and 'b' are the same, then the distance between
// 'a'/'b' and 'p' is returned.
func perpendicularDistance(p, a, b XY) float64 {
	if a == b {
		return p.Sub(a).Length()
	}
	aSubP := a.Sub(p)
	bSubA := b.Sub(a)
	unit := bSubA.Scale(1 / bSubA.Length())
	perpendicular := aSubP.Sub(unit.Scale(aSubP.Dot(unit)))
	return perpendicular.Length()
}
