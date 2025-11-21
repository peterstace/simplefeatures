package geom

import "fmt"

// walk calls fn for each control point in the geometry.
func walk(g Geometry, fn func(XY)) {
	switch g.Type() {
	case TypePoint:
		if xy, ok := g.MustAsPoint().XY(); ok {
			fn(xy)
		}
	case TypeLineString:
		seq := g.MustAsLineString().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			fn(seq.GetXY(i))
		}
	case TypePolygon:
		walk(g.Boundary(), fn)
	case TypeMultiPoint:
		mp := g.MustAsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			if xy, ok := mp.PointN(i).XY(); ok {
				fn(xy)
			}
		}
	case TypeMultiLineString:
		mls := g.MustAsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			walk(mls.LineStringN(i).AsGeometry(), fn)
		}
	case TypeMultiPolygon:
		walk(g.Boundary(), fn)
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			walk(gc.GeometryN(i), fn)
		}
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}
