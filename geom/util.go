package geom

import (
	"fmt"
	"sort"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// sortAndUniquifyXYs sorts the xys, and then makes them unique. The input
// slice is modified, however the result is in the returned slice since it may
// have its size changed due to uniquification.
func sortAndUniquifyXYs(xys []XY) []XY {
	if len(xys) == 0 {
		return xys
	}
	sort.Slice(xys, func(i, j int) bool {
		ptI := xys[i]
		ptJ := xys[j]
		if ptI.X != ptJ.X {
			return ptI.X < ptJ.X
		}
		return ptI.Y < ptJ.Y
	})
	n := 1
	for i := 1; i < len(xys); i++ {
		if xys[i] != xys[i-1] {
			xys[n] = xys[i]
			n++
		}
	}
	return xys[:n]
}

//func reverseXYs(fwd []XY) []XY {
//	rev := make([]XY, len(fwd))
//	for i := range rev {
//		rev[i] = fwd[len(fwd)-1-i]
//	}
//	return rev
//}

func reverseSequence(seq Sequence) Sequence {
	n := seq.Length()
	coords := make([]float64, 0, 2*n)
	for i := n - 1; i >= 0; i-- {
		xy := seq.GetXY(i)
		coords = append(coords, xy.X, xy.Y)
	}
	return NewSequence(coords, DimXY)
}

// fastMin is a faster but not functionally identical version of math.Min.
func fastMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// fastMax is a faster but not functionally identical version of math.Max.
func fastMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// sortFloat64Pair returns a and b in sorted order.
func sortFloat64Pair(a, b float64) (float64, float64) {
	if a > b {
		return b, a
	}
	return a, b
}
