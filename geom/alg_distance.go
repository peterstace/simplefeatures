package geom

import (
	"fmt"
	"math"
)

func dispatchDistance(g1, g2 Geometry) (float64, bool) {
	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}

	switch g1.Type() {
	case TypePoint:
		xy, ok := g1.AsPoint().XY()
		if !ok {
			return 0, false
		}
		switch g2.Type() {
		case TypePoint:
			return distBetweenXYAndPoint(xy, g2.AsPoint())
		case TypeLineString:
			return distBetweenXYAndLineString(xy, g2.AsLineString())
		case TypePolygon:
			return distBetweenXYAndPolygon(xy, g2.AsPolygon())
		case TypeMultiPoint:
			return distBetweenXYAndMultiPoint(xy, g2.AsMultiPoint())
		case TypeMultiLineString:
			return distBetweenXYAndMultiLineString(xy, g2.AsMultiLineString())
		case TypeMultiPolygon:
			return distBetweenXYAndMultiPolygon(xy, g2.AsMultiPolygon())
		case TypeGeometryCollection:
			return distBetweenXYAndGeometryCollection(xy, g2.AsGeometryCollection())
		}
	case TypeLineString:
		ls := g1.AsLineString()
		switch g2.Type() {
		case TypeLineString:
			return distBetweenLineStringAndLineString(ls, g2.AsLineString())
		case TypePolygon:
			break
		case TypeMultiLineString:
			break
		case TypeMultiPolygon:
			break
		case TypeGeometryCollection:
			break
		}
	case TypePolygon:
		switch g2.Type() {
		case TypePolygon:
			break
		case TypeMultiPoint:
			break
		case TypeMultiLineString:
			break
		case TypeMultiPolygon:
			break
		case TypeGeometryCollection:
			break
		}
	case TypeMultiPoint:
		switch g2.Type() {
		case TypeMultiPoint:
			break
		case TypeMultiLineString:
			break
		case TypeMultiPolygon:
			break
		case TypeGeometryCollection:
			break
		}
	case TypeMultiLineString:
		switch g2.Type() {
		case TypeMultiLineString:
			break
		case TypeMultiPolygon:
			break
		case TypeGeometryCollection:
			break
		}
	case TypeMultiPolygon:
		switch g2.Type() {
		case TypeMultiPolygon:
			break
		case TypeGeometryCollection:
			break
		}
	case TypeGeometryCollection:
		switch g2.Type() {
		case TypeGeometryCollection:
			break
		}
	}

	panic(fmt.Sprintf("implementation error: unhandled geometry types %s and %s", g1.Type(), g2.Type()))
}

func distBetweenXYs(xy1, xy2 XY) float64 {
	sub := xy1.Sub(xy2)
	return math.Sqrt(sub.Dot(sub))
}

func distBetweenXYAndLine(xy XY, ln line) float64 {
	ab := ln.b.Sub(ln.a)
	abLen := ab.Length()
	proj := xy.Sub(ln.a).Dot(ab) / abLen
	var closest XY
	switch {
	case proj < 0:
		closest = ln.a
	case proj > abLen:
		closest = ln.b
	default:
		scaled := ab.Scale(proj / abLen)
		closest = scaled.Add(ln.a)
	}
	return distBetweenXYs(xy, closest)
}

func distBetweenLineAndLine(ln1, ln2 line) float64 {
	minDist := math.Inf(+1)
	for _, dist := range [4]float64{
		distBetweenXYAndLine(ln1.a, ln2),
		distBetweenXYAndLine(ln1.b, ln2),
		distBetweenXYAndLine(ln2.a, ln1),
		distBetweenXYAndLine(ln2.b, ln1),
	} {
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

type distAggregator struct {
	dist float64
}

func newDistAggregator() distAggregator {
	return distAggregator{math.Inf(+1)}
}

func (a *distAggregator) agg(dist float64, ok bool) {
	if ok && dist < a.dist {
		a.dist = dist
	}
}

func (a *distAggregator) result() (float64, bool) {
	if math.IsInf(a.dist, +1) {
		return 0, false
	}
	return a.dist, true
}

func distBetweenXYAndPoint(xy XY, pt Point) (float64, bool) {
	other, ok := pt.XY()
	if !ok {
		return 0, false
	}
	return distBetweenXYs(xy, other), true
}

func distBetweenXYAndLineString(xy XY, ls LineString) (float64, bool) {
	dist := newDistAggregator()
	seq := ls.Coordinates()
	n := seq.Length()
	for i := 0; i < n; i++ {
		ln, ok := getLine(seq, i)
		if ok {
			dist.agg(distBetweenXYAndLine(xy, ln), true)
		}
	}
	return dist.result()
}

func distBetweenXYAndPolygon(xy XY, poly Polygon) (float64, bool) {
	if hasIntersectionPointWithPolygon(NewPointFromXY(xy), poly) {
		return 0, true
	}
	return distBetweenXYAndMultiLineString(xy, poly.Boundary())
}

func distBetweenXYAndMultiPoint(xy XY, mp MultiPoint) (float64, bool) {
	dist := newDistAggregator()
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		dist.agg(distBetweenXYAndPoint(xy, pt))
	}
	return dist.result()
}

func distBetweenXYAndMultiLineString(xy XY, mls MultiLineString) (float64, bool) {
	dist := newDistAggregator()
	n := mls.NumLineStrings()
	for i := 0; i < n; i++ {
		ls := mls.LineStringN(i)
		dist.agg(distBetweenXYAndLineString(xy, ls))
	}
	return dist.result()
}

func distBetweenXYAndMultiPolygon(xy XY, mp MultiPolygon) (float64, bool) {
	dist := newDistAggregator()
	n := mp.NumPolygons()
	for i := 0; i < n; i++ {
		p := mp.PolygonN(i)
		dist.agg(distBetweenXYAndPolygon(xy, p))
	}
	return dist.result()
}

func distBetweenXYAndGeometryCollection(xy XY, gc GeometryCollection) (float64, bool) {
	pt := NewPointFromXY(xy)
	dist := newDistAggregator()
	gc.walk(func(g Geometry) {
		dist.agg(pt.Distance(g))
	})
	return dist.result()
}

func distBetweenLineStringAndLineString(ls1, ls2 LineString) (float64, bool) {
	seq1 := ls1.Coordinates()
	seq2 := ls2.Coordinates()
	n1 := seq1.Length()
	n2 := seq2.Length()

	dist := newDistAggregator()
	for i := 0; i < n1; i++ {
		ln1, ok := getLine(seq1, i)
		if !ok {
			continue
		}
		for j := 0; j < n2; j++ {
			ln2, ok := getLine(seq2, j)
			if !ok {
				continue
			}
			dist.agg(distBetweenLineAndLine(ln1, ln2), true)
		}
	}
	return dist.result()
}

func extractXYsAndLines(g Geometry) ([]XY, []line) {
	switch g.Type() {
	case TypePoint:
		return g.AsPoint().asXYs(), nil
	case TypeLineString:
		return nil, g.AsLineString().asLines()
	case TypePolygon:
		return nil, g.AsPolygon().Boundary().asLines()
	case TypeMultiPoint:
		return g.AsMultiPoint().asXYs(), nil
	case TypeMultiLineString:
		return nil, g.AsMultiLineString().asLines()
	case TypeMultiPolygon:
		return nil, g.AsMultiPolygon().Boundary().asLines()
	case TypeGeometryCollection:
		var allXYs []XY
		var allLines []line
		g.AsGeometryCollection().walk(func(child Geometry) {
			xys, lns := extractXYsAndLines(child)
			allXYs = append(allXYs, xys...)
			allLines = append(allLines, lns...)
		})
		return allXYs, allLines
	default:
		panic(fmt.Sprintf("implementation error: unhandled geometry types %s", g.Type()))
	}
}
