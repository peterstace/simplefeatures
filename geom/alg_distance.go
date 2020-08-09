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
		pt := g1.AsPoint()
		switch g2.Type() {
		case TypePoint:
			return distanceBetweenPointAndPoint(pt, g2.AsPoint())
		case TypeLineString:
			fallthrough
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

func distanceBetweenPointAndPoint(pt1, pt2 Point) (float64, bool) {
	xy1, ok := pt1.XY()
	if !ok {
		return 0, false
	}
	xy2, ok := pt2.XY()
	if !ok {
		return 0, false
	}
	return distanceBetweenXYs(xy1, xy2), true
}

func distanceBetweenXYs(xy1, xy2 XY) float64 {
	sub := xy1.Sub(xy2)
	return math.Sqrt(sub.Dot(sub))
}
