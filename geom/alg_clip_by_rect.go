package geom

// ClipByRect clips a geometry to an axis-aligned rectangle defined by the
// given [Envelope], returning the portion of the geometry that lies within the
// rectangle. It uses the Sutherland-Hodgman algorithm for polygon clipping
// and related approaches for other geometry types.
func ClipByRect(g Geometry, rect Envelope) Geometry {
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
	return NewEmptyPoint(p.CoordinatesType())
}

func clipMultiPointByRect(mp MultiPoint, rect Envelope) MultiPoint {
	n := mp.NumPoints()
	var pts []Point
	for i := 0; i < n; i++ {
		p := mp.PointN(i)
		clipped := clipPointByRect(p, rect)
		if !clipped.IsEmpty() {
			pts = append(pts, clipped)
		}
	}
	if len(pts) == 0 {
		return NewMultiPoint(nil).ForceCoordinatesType(mp.CoordinatesType())
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
	if len(lines) == 0 {
		return NewMultiLineString(nil).ForceCoordinatesType(mls.CoordinatesType())
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
	if len(polys) == 0 {
		return NewMultiPolygon(nil).ForceCoordinatesType(mp.CoordinatesType())
	}
	return NewMultiPolygon(polys)
}

func clipGeometryCollectionByRect(gc GeometryCollection, rect Envelope) GeometryCollection {
	panic("TODO")
}
