package geom

// Simplify returns a simplified version of the geometry using the
// Ramer-Douglas-Peucker algorithm.
func Simplify(g Geometry, threshold float64) (Geometry, error) {
	switch g.gtype {
	case TypeGeometryCollection:
		gc, err := simplifyGeometryCollection(g.AsGeometryCollection(), threshold)
		return gc.AsGeometry(), err
	case TypePoint:
		return g, nil
	case TypeLineString:
		ls, err := simplifyLineString(g.AsLineString(), threshold)
		return ls.AsGeometry(), err
	case TypePolygon:
		poly, err := simplifyPolygon(g.AsPolygon(), threshold)
		return poly.AsGeometry(), err
	case TypeMultiPoint:
		return g, nil
	case TypeMultiLineString:
		mls, err := simplifyMultiLineString(g.AsMultiLineString(), threshold)
		return mls.AsGeometry(), err
	case TypeMultiPolygon:
		mp, err := simplifyMultiPolygon(g.AsMultiPolygon(), threshold)
		return mp.AsGeometry(), err
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

func simplifyLineString(ls LineString, threshold float64) (LineString, error) {
	seq := ls.Coordinates()
	floats := ramerDouglasPeucker(nil, seq, threshold)
	seq = NewSequence(floats, seq.CoordinatesType())
	if seq.Length() > 0 && !hasAtLeast2DistinctPointsInSeq(seq) {
		return LineString{}, nil
	}
	return NewLineString(seq)
}

func simplifyMultiLineString(mls MultiLineString, threshold float64) (MultiLineString, error) {
	n := mls.NumLineStrings()
	lss := make([]LineString, 0, n)
	for i := 0; i < n; i++ {
		ls := mls.LineStringN(i)
		ls, err := simplifyLineString(ls, threshold)
		if err != nil {
			return MultiLineString{}, err
		}
		if !ls.IsEmpty() {
			lss = append(lss, ls)
		}
	}
	return NewMultiLineStringFromLineStrings(lss), nil
}

func simplifyPolygon(poly Polygon, threshold float64) (Polygon, error) {
	exterior, err := simplifyLineString(poly.ExteriorRing(), threshold)
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
		interior, err := simplifyLineString(poly.InteriorRingN(i), threshold)
		if err != nil {
			return Polygon{}, err
		}
		if interior.IsRing() {
			rings = append(rings, interior)
		}
	}
	return NewPolygonFromRings(rings)
}

func simplifyMultiPolygon(mp MultiPolygon, threshold float64) (MultiPolygon, error) {
	n := mp.NumPolygons()
	polys := make([]Polygon, 0, n)
	for i := 0; i < n; i++ {
		poly, err := simplifyPolygon(mp.PolygonN(i), threshold)
		if err != nil {
			return MultiPolygon{}, err
		}
		if !poly.IsEmpty() {
			polys = append(polys, poly)
		}
	}
	return NewMultiPolygonFromPolygons(polys)
}

func simplifyGeometryCollection(gc GeometryCollection, threshold float64) (GeometryCollection, error) {
	n := gc.NumGeometries()
	geoms := make([]Geometry, n)
	for i := 0; i < n; i++ {
		var err error
		geoms[i], err = Simplify(gc.GeometryN(i), threshold)
		if err != nil {
			return GeometryCollection{}, err
		}
	}
	return NewGeometryCollection(geoms), nil
}

func ramerDouglasPeucker(dst []float64, seq Sequence, threshold float64) []float64 {
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

	if maxDist <= threshold {
		dst = seq.appendPoint(dst, 0)
		dst = seq.appendPoint(dst, n-1)
		return dst
	}

	dst = ramerDouglasPeucker(dst, seq.Slice(0, maxDistIdx+1), threshold)
	stride := seq.CoordinatesType().Dimension()
	dst = dst[:len(dst)-stride]
	dst = ramerDouglasPeucker(dst, seq.Slice(maxDistIdx, n), threshold)
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
