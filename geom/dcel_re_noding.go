package geom

import "fmt"

// reNodeGeometry returns a geometry that is spatially equivalent to g, but
// with additional nodes (i.e. control points). The cut set is used to
// determine the location of additional nodes (they will occur wherever the cut
// set intersects with g). Because intersection between line segments is
// non-commutative (due to numerical precision issues), flip is used to reverse
// the order between the operator and operand.
func reNodeGeometry(g Geometry, cut cutSet, flip bool) Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return reNodeGeometryCollection(g.AsGeometryCollection(), cut, flip).AsGeometry()
	case TypeLineString:
		return reNodeLineString(g.AsLineString(), cut, flip).AsGeometry()
	case TypePolygon:
		return reNodePolygon(g.AsPolygon(), cut, flip).AsGeometry()
	case TypeMultiLineString:
		return reNodeMultiLineString(g.AsMultiLineString(), cut, flip).AsGeometry()
	case TypeMultiPolygon:
		return reNodeMultiPolygonString(g.AsMultiPolygon(), cut, flip).AsGeometry()
	case TypePoint, TypeMultiPoint:
		// It doesn't make sense to re-node point geometries, since they have
		// no edges.
		return g
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}

type cutSet struct {
	lnIndex indexedLines
	ptIndex indexedPoints
}

func newCutSet(g Geometry) cutSet {
	return cutSet{
		lnIndex: newIndexedLines(appendLines(nil, g)),
		ptIndex: newIndexedPoints(appendPoints(nil, g)),
	}
}

func appendLines(lines []line, g Geometry) []line {
	switch g.Type() {
	case TypeLineString:
		seq := g.AsLineString().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			ln, ok := getLine(seq, i)
			if ok {
				lines = append(lines, ln)
			}
		}
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		for i := 0; i < mls.NumLineStrings(); i++ {
			ls := mls.LineStringN(i)
			lines = appendLines(lines, ls.AsGeometry())
		}
	case TypePolygon:
		lines = appendLines(lines, g.AsPolygon().Boundary().AsGeometry())
	case TypeMultiPolygon:
		lines = appendLines(lines, g.AsMultiPolygon().Boundary().AsGeometry())
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			lines = appendLines(lines, gc.GeometryN(i))
		}
	}
	return lines
}

func appendPoints(points []XY, g Geometry) []XY {
	switch g.Type() {
	case TypePoint:
		coords, ok := g.AsPoint().Coordinates()
		if ok {
			points = append(points, coords.XY)
		}
	case TypeMultiPoint:
		mp := g.AsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			points = appendPoints(points, mp.PointN(i).AsGeometry())
		}
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			points = appendPoints(points, gc.GeometryN(i))
		}
	}
	return points
}

func reNodeLineString(ls LineString, cut cutSet, flip bool) LineString {
	var newCoords []float64
	seq := ls.Coordinates()
	n := seq.Length()
	for lnIdx := 0; lnIdx < n; lnIdx++ {
		ln, ok := getLine(seq, lnIdx)
		if !ok {
			continue
		}

		// Collect cut locations.
		xys := []XY{ln.a, ln.b}
		cut.lnIndex.tree.RangeSearch(ln.envelope().box(), func(i int) error {
			other := cut.lnIndex.lines[i]
			var inter lineWithLineIntersection
			if flip {
				inter = other.intersectLine(ln)
			} else {
				inter = ln.intersectLine(other)
			}
			if inter.empty {
				return nil
			}
			xys = append(xys, inter.ptA, inter.ptB)
			return nil
		})
		cut.ptIndex.tree.RangeSearch(ln.envelope().box(), func(i int) error {
			other := cut.ptIndex.points[i]
			if ln.intersectsXY(other) {
				xys = append(xys, other)
			}
			return nil
		})

		xys = sortAndUniquifyXYs(xys) // TODO: make common function

		// Reverse order to match direction of edge.
		if xys[0] != ln.a {
			for i := 0; i < len(xys)/2; i++ {
				j := len(xys) - i - 1
				xys[i], xys[j] = xys[j], xys[i]
			}
		}

		// Add coords related to this line segment. The end of the previous
		// line is the same as the first point of this line, so we skip it to
		// avoid doubling up.
		if len(newCoords) == 0 {
			newCoords = append(newCoords, xys[0].X, xys[0].Y)
		}
		for _, xy := range xys[1:] {
			newCoords = append(newCoords, xy.X, xy.Y)
		}
	}

	newLS, err := NewLineString(NewSequence(newCoords, DimXY), DisableAllValidations)
	if err != nil {
		panic(fmt.Sprintf("could not re-node LineString: %v", err))
	}
	return newLS
}

func reNodeMultiLineString(mls MultiLineString, cut cutSet, flip bool) MultiLineString {
	n := mls.NumLineStrings()
	lss := make([]LineString, n)
	for i := 0; i < n; i++ {
		lss[i] = reNodeLineString(mls.LineStringN(i), cut, flip)
	}
	return NewMultiLineStringFromLineStrings(lss, DisableAllValidations)
}

func reNodePolygon(poly Polygon, cut cutSet, flip bool) Polygon {
	reNodedBoundary := reNodeMultiLineString(poly.Boundary(), cut, flip)
	n := reNodedBoundary.NumLineStrings()
	rings := make([]LineString, n)
	for i := 0; i < n; i++ {
		rings[i] = reNodedBoundary.LineStringN(i)
	}
	reNodedPoly, err := NewPolygonFromRings(rings, DisableAllValidations)
	if err != nil {
		panic(err)
	}
	return reNodedPoly
}

func reNodeMultiPolygonString(mp MultiPolygon, cut cutSet, flip bool) MultiPolygon {
	n := mp.NumPolygons()
	polys := make([]Polygon, n)
	for i := 0; i < n; i++ {
		polys[i] = reNodePolygon(mp.PolygonN(i), cut, flip)
	}
	reNodedMP, err := NewMultiPolygonFromPolygons(polys, DisableAllValidations)
	if err != nil {
		panic(err)
	}
	return reNodedMP
}

func reNodeGeometryCollection(gc GeometryCollection, cut cutSet, flip bool) GeometryCollection {
	n := gc.NumGeometries()
	geoms := make([]Geometry, n)
	for i := 0; i < n; i++ {
		geoms[i] = reNodeGeometry(gc.GeometryN(i), cut, flip)
	}
	return NewGeometryCollection(geoms, DisableAllValidations)
}
