package geom

import (
	"fmt"
	"math"
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
	switch g.tag {
	case pointTag:
		return 1
	case lineTag:
		return 2
	case lineStringTag:
		return 3
	case polygonTag:
		return 4
	case multiPointTag:
		return 5
	case multiLineStringTag:
		return 6
	case multiPolygonTag:
		return 7
	case geometryCollectionTag:
		return 8
	default:
		panic(fmt.Sprintf("unknown geometry tag: %s", g.tag))
	}
}

func mustEnv(env Envelope, ok bool) Envelope {
	if !ok {
		panic("not ok")
	}
	return env
}

func minX(ln Line) float64 {
	return math.Min(ln.StartPoint().X, ln.EndPoint().X)
}

func maxX(ln Line) float64 {
	return math.Max(ln.StartPoint().X, ln.EndPoint().X)
}
