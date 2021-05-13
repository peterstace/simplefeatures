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
		return gc.AsGeometry(), err
	case TypePoint:
		return g, nil
	case TypeLineString:
		ls, err := s.simplifyLineString(g.AsLineString())
		return ls.AsGeometry(), err
	case TypePolygon:
		poly, err := s.simplifyPolygon(g.AsPolygon())
		return poly.AsGeometry(), err
	case TypeMultiPoint:
		return g, nil
	case TypeMultiLineString:
		mls, err := s.simplifyMultiLineString(g.AsMultiLineString())
		return mls.AsGeometry(), err
	case TypeMultiPolygon:
		mp, err := s.simplifyMultiPolygon(g.AsMultiPolygon())
		return mp.AsGeometry(), err
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
	return NewMultiLineStringFromLineStrings(lss, s.opts...), nil
}

func (s simplifier) simplifyPolygon(poly Polygon) (Polygon, error) {
	exterior, err := s.simplifyLineString(poly.ExteriorRing())
	if err != nil {
		return Polygon{}, err
	}
	if !exterior.IsRing() {
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
	return NewPolygonFromRings(rings, s.opts...)
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
	return NewMultiPolygonFromPolygons(polys, s.opts...)
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
	n := seq.Length()
	if n <= 2 {
		return seq.appendAllPoints(dst)
	}

	first, last := seq.GetXY(0), seq.GetXY(n-1)

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
		if dist := calcDist(seq.GetXY(i)); dist > maxDist {
			maxDist = dist
			maxDistIdx = i
		}
	}

	if maxDist <= s.threshold {
		dst = seq.appendPoint(dst, 0)
		dst = seq.appendPoint(dst, n-1)
		return dst
	}

	dst = s.ramerDouglasPeucker(dst, seq.Slice(0, maxDistIdx+1))
	stride := seq.CoordinatesType().Dimension()
	dst = dst[:len(dst)-stride]
	dst = s.ramerDouglasPeucker(dst, seq.Slice(maxDistIdx, n))
	return dst
}

// perpendicularDistance is the distance from p to the infinite line going
// through ln.
func perpendicularDistance(p XY, ln line) float64 {
	aSubP := ln.a.Sub(p)
	unit := ln.b.Sub(ln.a).Scale(1 / ln.length())
	perpendicular := aSubP.Sub(unit.Scale(aSubP.Dot(unit)))
	return perpendicular.Length()
}
