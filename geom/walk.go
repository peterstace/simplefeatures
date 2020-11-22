package geom

import "fmt"

// walkPoints calls fn for each control point in the geometry.
func walkPoints(g Geometry, fn func(XY)) {
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
		walkPoints(g.Boundary(), fn)
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
			walkPoints(mls.LineStringN(i).AsGeometry(), fn)
		}
	case TypeMultiPolygon:
		walkPoints(g.Boundary(), fn)
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			walkPoints(gc.GeometryN(i), fn)
		}
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}

// walkLines calls fn for each line segment in the geometry.
func walkLines(g Geometry, fn func(line)) {
	switch g.Type() {
	case TypePoint:
		// NO-OP
	case TypeLineString:
		seq := g.AsLineString().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			if ln, ok := getLine(seq, i); ok {
				fn(ln)
			}
		}
	case TypePolygon:
		walkLines(g.Boundary(), fn)
	case TypeMultiPoint:
		// NO-OP
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			walkLines(mls.LineStringN(i).AsGeometry(), fn)
		}
	case TypeMultiPolygon:
		walkLines(g.Boundary(), fn)
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			walkLines(gc.GeometryN(i), fn)
		}
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}
