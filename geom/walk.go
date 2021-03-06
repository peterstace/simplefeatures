package geom

import "fmt"

// walk calls fn for each control point in the geometry.
func walk(g Geometry, fn func(XY)) {
	switch g.Type() {
	case TypePoint:
		if xy, ok := g.AsPoint().XY(); ok {
			fn(xy)
		}
	case TypeLineString:
		seq := g.AsLineString().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			fn(seq.GetXY(i))
		}
	case TypePolygon:
		walk(g.Boundary(), fn)
	case TypeMultiPoint:
		seq, empty := g.AsMultiPoint().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			if !empty.Get(i) {
				fn(seq.GetXY(i))
			}
		}
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			walk(mls.LineStringN(i).AsGeometry(), fn)
		}
	case TypeMultiPolygon:
		walk(g.Boundary(), fn)
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			walk(gc.GeometryN(i), fn)
		}
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}
