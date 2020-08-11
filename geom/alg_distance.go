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
			return distanceBetweenXYAndPoint(xy, g2.AsPoint())
		case TypeLineString:
			return distanceBetweenXYAndLineString(xy, g2.AsLineString())
		case TypePolygon:
			//return distanceBetweenXYAndPolygon(xy, g2.AsPolygon())
			fallthrough
		case TypeMultiPoint:
			fallthrough
		case TypeMultiLineString:
			return distanceBetweenXYAndMultiPoint(xy, g2.AsMultiPoint())
		case TypeMultiPolygon:
			fallthrough
		case TypeGeometryCollection:
			// TODO
		}
	case TypeLineString:
		switch g2.Type() {
		case TypeLineString:
			fallthrough
		case TypePolygon:
			fallthrough
		case TypeMultiLineString:
			fallthrough
		case TypeMultiPolygon:
			fallthrough
		case TypeGeometryCollection:
			// TODO
		}
	case TypePolygon:
		switch g2.Type() {
		case TypePolygon:
			fallthrough
		case TypeMultiPoint:
			fallthrough
		case TypeMultiLineString:
			fallthrough
		case TypeMultiPolygon:
			fallthrough
		case TypeGeometryCollection:
			// TODO
		}
	case TypeMultiPoint:
		switch g2.Type() {
		case TypeMultiPoint:
			fallthrough
		case TypeMultiLineString:
			fallthrough
		case TypeMultiPolygon:
			fallthrough
		case TypeGeometryCollection:
			// TODO
		}
	case TypeMultiLineString:
		switch g2.Type() {
		case TypeMultiLineString:
			fallthrough
		case TypeMultiPolygon:
			fallthrough
		case TypeGeometryCollection:
			// TODO
		}
	case TypeMultiPolygon:
		switch g2.Type() {
		case TypeMultiPolygon:
			fallthrough
		case TypeGeometryCollection:
			// TODO
		}
	case TypeGeometryCollection:
		switch g2.Type() {
		case TypeGeometryCollection:
			// TODO
		}
	}

	panic(fmt.Sprintf("implementation error: unhandled geometry types %s and %s", g1.Type(), g2.Type()))
}

func distanceBetweenXYs(xy1, xy2 XY) float64 {
	sub := xy1.Sub(xy2)
	return math.Sqrt(sub.Dot(sub))
}

func distanceBetweenXYAndLine(xy XY, ln line) float64 {
	lnVec := ln.b.Sub(ln.a)
	lnVecUnit := lnVec.Unit()
	proj := xy.Sub(ln.a).Dot(lnVecUnit)
	var closest XY
	switch {
	case proj < 0:
		closest = ln.a
	case proj > 1:
		closest = ln.b
	default:
		scaled := lnVecUnit.Scale(proj)
		closest = scaled.Add(ln.a)
	}
	return distanceBetweenXYs(xy, closest)
}

func distanceBetweenXYAndPoint(xy XY, pt Point) (float64, bool) {
	other, ok := pt.XY()
	if !ok {
		return 0, false
	}
	return distanceBetweenXYs(xy, other), true
}

func distanceBetweenXYAndLineString(xy XY, ls LineString) (float64, bool) {
	if ls.IsEmpty() {
		return 0, false
	}
	minDist := math.Inf(+1)
	seq := ls.Coordinates()
	n := seq.Length()
	for i := 0; i < n; i++ {
		ln, ok := getLine(seq, i)
		if !ok {
			continue
		}
		dist := distanceBetweenXYAndLine(xy, ln)
		minDist = math.Min(minDist, dist)
	}
	return minDist, true
}

//func distanceBetweenXYAndPolygon(xy XY, poly Polygon) (float64, bool) {
//	// TODO: Do I need this check here? The distance to the boundary might do
//	// this for us.
//	if poly.IsEmpty() {
//		return 0, false
//	}
//
//	if hasIntersectionPointWithPolygon(NewPointFromXY(xy), poly) {
//		return 0, true
//	}
//
//	//poly.Boundary()
//	return 69, true
//}

func distanceBetweenXYAndMultiPoint(xy XY, mp MultiPoint) (float64, bool) {
	minDist := math.Inf(+1)
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		dist, ok := distanceBetweenXYAndPoint(xy, pt)
		if ok && dist < minDist {
			minDist = dist
		}
	}
	if math.IsInf(minDist, +1) {
		return 0, false
	}
	return minDist, true
}
