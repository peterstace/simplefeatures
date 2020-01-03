package geom

import "fmt"

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func rank(g Geometry) int {
	switch {
	case g.IsEmptySet():
		return 1
	case g.IsPoint():
		return 2
	case g.IsLine():
		return 3
	case g.IsLineString():
		return 4
	case g.IsPolygon():
		return 5
	case g.IsMultiPoint():
		return 6
	case g.IsMultiLineString():
		return 7
	case g.IsMultiPolygon():
		return 8
	case g.IsGeometryCollection():
		return 9
	default:
		panic(fmt.Sprintf("unknown geometry type: %s", g.tag))
	}
}

func must(x Geometry, err error) Geometry {
	if err != nil {
		panic(err)
	}
	return x
}

func mustEnv(env Envelope, ok bool) Envelope {
	if !ok {
		panic("not ok")
	}
	return env
}
