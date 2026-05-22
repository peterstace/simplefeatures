package geom

// ClipByRect2D clips a geometry to the 2D axis-aligned rectangle defined by
// the given [Envelope], returning the portion of the geometry that lies within
// the rectangle. It uses the Sutherland-Hodgman algorithm for polygon clipping
// and the Liang-Barsky algorithm for line clipping.
//
// The result is always 2D ([DimXY]): any Z or M values on the input are
// discarded. Linear interpolation of Z/M at clip-edge intersections is
// reasonably well-defined, but values that would have to be synthesised at
// rectangle corners (where the clipped boundary must traverse the rect) are
// not, so this function avoids the problem by reducing to two dimensions
// throughout.
func ClipByRect2D(g Geometry, rect Envelope) Geometry {
	return clipByRect2D(g.Force2D(), rect)
}

// clipByRect2D dispatches by geometry type. The input must already have been
// reduced to [DimXY] by the caller.
func clipByRect2D(g Geometry, rect Envelope) Geometry {
	switch g.Type() {
	case TypePoint:
		return clipPointByRect(g.MustAsPoint(), rect).AsGeometry()
	case TypeMultiPoint:
		return clipMultiPointByRect(g.MustAsMultiPoint(), rect).AsGeometry()
	case TypeLineString:
		return clipLineStringByRect(g.MustAsLineString(), rect)
	case TypeMultiLineString:
		return clipMultiLineStringByRect(g.MustAsMultiLineString(), rect).AsGeometry()
	case TypePolygon:
		return clipPolygonByRect(g.MustAsPolygon(), rect)
	case TypeMultiPolygon:
		return clipMultiPolygonByRect(g.MustAsMultiPolygon(), rect).AsGeometry()
	case TypeGeometryCollection:
		return clipGeometryCollectionByRect(g.MustAsGeometryCollection(), rect).AsGeometry()
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

func clipPointByRect(p Point, rect Envelope) Point {
	xy, ok := p.XY()
	if !ok {
		return p
	}
	if rect.Contains(xy) {
		return p
	}
	return Point{}
}

func clipMultiPointByRect(mp MultiPoint, rect Envelope) MultiPoint {
	n := mp.NumPoints()
	var pts []Point
	for i := 0; i < n; i++ {
		clipped := clipPointByRect(mp.PointN(i), rect)
		if !clipped.IsEmpty() {
			pts = append(pts, clipped)
		}
	}
	return NewMultiPoint(pts)
}

func clipMultiLineStringByRect(mls MultiLineString, rect Envelope) MultiLineString {
	n := mls.NumLineStrings()
	var lines []LineString
	for i := 0; i < n; i++ {
		clipped := clipLineStringByRect(mls.LineStringN(i), rect)
		switch clipped.Type() {
		case TypeLineString:
			ls := clipped.MustAsLineString()
			if !ls.IsEmpty() {
				lines = append(lines, ls)
			}
		case TypeMultiLineString:
			lines = append(lines, clipped.MustAsMultiLineString().Dump()...)
		default:
			panic("unexpected type from clipLineStringByRect: " + clipped.Type().String())
		}
	}
	return NewMultiLineString(lines)
}

func clipMultiPolygonByRect(mp MultiPolygon, rect Envelope) MultiPolygon {
	n := mp.NumPolygons()
	var polys []Polygon
	for i := 0; i < n; i++ {
		clipped := clipPolygonByRect(mp.PolygonN(i), rect)
		switch clipped.Type() {
		case TypePolygon:
			p := clipped.MustAsPolygon()
			if !p.IsEmpty() {
				polys = append(polys, p)
			}
		case TypeMultiPolygon:
			polys = append(polys, clipped.MustAsMultiPolygon().Dump()...)
		default:
			panic("unexpected type from clipPolygonByRect: " + clipped.Type().String())
		}
	}
	return NewMultiPolygon(polys)
}

func clipGeometryCollectionByRect(gc GeometryCollection, rect Envelope) GeometryCollection {
	n := gc.NumGeometries()
	var geoms []Geometry
	for i := 0; i < n; i++ {
		clipped := clipByRect2D(gc.GeometryN(i), rect)
		if !clipped.IsEmpty() {
			geoms = append(geoms, clipped)
		}
	}
	return NewGeometryCollection(geoms)
}

// xysToSeq builds a [DimXY] [Sequence] from a slice of [XY].
func xysToSeq(xys []XY) Sequence {
	floats := make([]float64, 0, 2*len(xys))
	for _, xy := range xys {
		floats = append(floats, xy.X, xy.Y)
	}
	return NewSequence(floats, DimXY)
}
