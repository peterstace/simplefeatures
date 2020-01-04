package geom

import "fmt"

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func rank(g Geometry) int {
	switch g.tag {
	case emptySetTag:
		return 1
	case pointTag:
		return 2
	case lineTag:
		return 3
	case lineStringTag:
		return 4
	case polygonTag:
		return 5
	case multiPointTag:
		return 6
	case multiLineStringTag:
		return 7
	case multiPolygonTag:
		return 8
	case geometryCollectionTag:
		return 9
	default:
		panic(fmt.Sprintf("unknown geometry tag: %s", g.tag))
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
