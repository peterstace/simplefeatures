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
	switch g.tag {
	case pointTag:
		return 1
	case lineStringTag:
		return 2
	case polygonTag:
		return 3
	case multiPointTag:
		return 4
	case multiLineStringTag:
		return 5
	case multiPolygonTag:
		return 6
	case geometryCollectionTag:
		return 7
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
