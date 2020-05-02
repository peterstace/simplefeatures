package geom

import (
	"fmt"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func rank(g Geometry) int {
	switch g.gtype {
	case TypePoint:
		return 1
	case TypeLineString:
		return 2
	case TypePolygon:
		return 3
	case TypeMultiPoint:
		return 4
	case TypeMultiLineString:
		return 5
	case TypeMultiPolygon:
		return 6
	case TypeGeometryCollection:
		return 7
	default:
		panic(fmt.Sprintf("unknown geometry tag: %s", g.gtype))
	}
}

func mustEnv(env Envelope, ok bool) Envelope {
	if !ok {
		panic("not ok")
	}
	return env
}
