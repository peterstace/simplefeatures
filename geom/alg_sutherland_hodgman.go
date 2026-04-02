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
	panic("TODO")
}

func clipLineStringByRect(ls LineString, rect Envelope) Geometry {
	panic("TODO")
}

func clipMultiLineStringByRect(mls MultiLineString, rect Envelope) MultiLineString {
	panic("TODO")
}

func clipPolygonByRect(p Polygon, rect Envelope) Geometry {
	panic("TODO")
}

func clipMultiPolygonByRect(mp MultiPolygon, rect Envelope) MultiPolygon {
	panic("TODO")
}

func clipGeometryCollectionByRect(gc GeometryCollection, rect Envelope) GeometryCollection {
	panic("TODO")
}
