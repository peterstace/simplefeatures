package geom

import "fmt"

// walkXY calls fn for each control point (XY) in the geometry.
func walkXY(g Geometry, fn func(XY)) {
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
		walkXY(g.Boundary(), fn)
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
			walkXY(mls.LineStringN(i).AsGeometry(), fn)
		}
	case TypeMultiPolygon:
		walkXY(g.Boundary(), fn)
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			walkXY(gc.GeometryN(i), fn)
		}
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}

// walkLines calls fn for each line in the geometry.
func walkLines(g Geometry, fn func(line)) {
	switch g.Type() {
	case TypePoint:
		// Points have no edges.
	case TypeLineString:
		seq := g.MustAsLineString().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			ln, ok := getLine(seq, i)
			if ok {
				fn(ln)
			}
		}
	case TypePolygon:
		walkLines(g.Boundary(), fn)
	case TypeMultiPoint:
		// MultiPoints have no edges.
	case TypeMultiLineString:
		mls := g.MustAsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			walkLines(mls.LineStringN(i).AsGeometry(), fn)
		}
	case TypeMultiPolygon:
		walkLines(g.Boundary(), fn)
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			walkLines(gc.GeometryN(i), fn)
		}
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}
