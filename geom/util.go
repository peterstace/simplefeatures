package geom

import "fmt"

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func rank(g GeometryX) int {
	switch g.(type) {
	case EmptySet:
		return 1
	case Point:
		return 2
	case Line:
		return 3
	case LineString:
		return 4
	case Polygon:
		return 5
	case MultiPoint:
		return 6
	case MultiLineString:
		return 7
	case MultiPolygon:
		return 8
	case GeometryCollection:
		return 9
	default:
		panic(fmt.Sprintf("unknown geometry type: %T", g))
	}
}

func must(x GeometryX, err error) GeometryX {
	if err != nil {
		panic(err)
	}
	return x
}
