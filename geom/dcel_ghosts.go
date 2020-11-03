package geom

import "fmt"

func connectGeometry(g Geometry) MultiLineString {
	var ghostLSs []LineString
	var seenFirst bool
	var first XY
	addComponent := func(pt Point) {
		xy, ok := pt.XY()
		if !ok {
			return
		}
		if seenFirst {
			if first != xy {
				seq := NewSequence([]float64{first.X, first.Y, xy.X, xy.Y}, DimXY)
				ghostLS, err := NewLineString(seq)
				if err != nil {
					// Can't happen, since first and pt are not the same.
					panic(fmt.Sprintf("could not construct LineString: %v", err))
				}
				ghostLSs = append(ghostLSs, ghostLS)
			}
		} else {
			seenFirst = true
			first = xy
		}
	}

	switch g.Type() {
	case TypePoint:
		// A single Point is already trivially connected.
	case TypeMultiPoint:
		mp := g.AsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			addComponent(mp.PointN(i))
		}
	case TypeLineString:
		// LineStrings are already connected.
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			ls := mls.LineStringN(i)
			addComponent(ls.StartPoint())
		}
	case TypePolygon:
		poly := g.AsPolygon()
		addComponent(poly.ExteriorRing().StartPoint())
		n := poly.NumInteriorRings()
		for i := 0; i < n; i++ {
			addComponent(poly.InteriorRingN(i).StartPoint())
		}
	case TypeMultiPolygon:
		mp := g.AsMultiPolygon()
		n := mp.NumPolygons()
		for i := 0; i < n; i++ {
			poly := mp.PolygonN(i)
			addComponent(poly.ExteriorRing().StartPoint())
			m := poly.NumInteriorRings()
			for j := 0; j < m; j++ {
				addComponent(poly.InteriorRingN(j).StartPoint())
			}
		}
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			addComponent(pointOnGeometry(gc.GeometryN(i)))
		}
	default:
		panic(fmt.Sprintf("unknown geometry type: %v", g.Type()))
	}

	return NewMultiLineStringFromLineStrings(ghostLSs)
}

func connectGeometries(g1, g2 Geometry) LineString {
	pt1 := pointOnGeometry(g1)
	pt2 := pointOnGeometry(g2)

	xy1, ok1 := pt1.XY()
	xy2, ok2 := pt2.XY()
	if !ok1 || !ok2 || xy1 == xy2 {
		return LineString{}
	}

	coords := []float64{xy1.X, xy1.Y, xy2.X, xy2.Y}
	ls, err := NewLineString(NewSequence(coords, DimXY))
	if err != nil {
		// Can't happen, since we have already checked that xy1 != xy2.
		panic(fmt.Sprintf("could not create lines string: %v", err))
	}
	return ls
}

func pointOnGeometry(g Geometry) Point {
	switch g.Type() {
	case TypePoint:
		return g.AsPoint()
	case TypeMultiPoint:
		mp := g.AsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			pt := mp.PointN(i)
			if !pt.IsEmpty() {
				return pt
			}
		}
		return Point{}
	case TypeLineString:
		return g.AsLineString().StartPoint()
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			pt := mls.LineStringN(i).StartPoint()
			if !pt.IsEmpty() {
				return pt
			}
		}
		return Point{}
	case TypePolygon:
		return pointOnGeometry(g.Boundary())
	case TypeMultiPolygon:
		return pointOnGeometry(g.Boundary())
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			pt := pointOnGeometry(gc.GeometryN(i))
			if !pt.IsEmpty() {
				return pt
			}
		}
		return Point{}
	default:
		panic(fmt.Sprintf("unknown geometry type: %v", g.Type()))
	}
}

func mergeMultiLineStrings(mlss []MultiLineString) MultiLineString {
	var lss []LineString
	for _, mls := range mlss {
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			lss = append(lss, mls.LineStringN(i))
		}
	}
	return NewMultiLineStringFromLineStrings(lss)
}
