package geom

func Dump(g Geometry) []Geometry {
	var geoms []Geometry
	switch g.Type() {
	case TypePoint, TypeLineString, TypePolygon:
		geoms = append(geoms, g)
	case TypeMultiPoint:
		mp := g.AsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			geoms = append(geoms, mp.PointN(i).AsGeometry())
		}
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			geoms = append(geoms, mls.LineStringN(i).AsGeometry())
		}
	case TypeMultiPolygon:
		mp := g.AsMultiPolygon()
		n := mp.NumPolygons()
		for i := 0; i < n; i++ {
			geoms = append(geoms, mp.PolygonN(i).AsGeometry())
		}
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			geoms = append(geoms, Dump(gc.GeometryN(i))...)
		}
	default:
		panic("unknown type: " + g.Type().String())
	}
	return geoms
}
