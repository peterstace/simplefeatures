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

func float64AsInt64(f float64) int64 {
	if f < 0 {
		return -int64(math.Float64bits(-f))
	} else {
		return int64(math.Float64bits(f))
	}
}

// diffULP calculates f1-f1, but in terms of ULP (units of least precision)
// instead of the regular float64 result. This is the number of discrete
// floating point values between the two inputs.
func diffULP(f1, f2 float64) int64 {
	return float64AsInt64(f1) - float64AsInt64(f2)
}
