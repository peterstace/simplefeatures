package geom

import (
	"fmt"
	"math"
	"sort"
)

// TODO: Remove this when we require Go 1.21 (the max builtin can be used instead).
func maxInt(a, b int) int {
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
		return xys[i].Less(xys[j])
	})
	return uniquifyGroupedXYs(xys)
}

// uniquifyGroupedXYs uniquifies the xys, assuming that equal values are always
// grouped adjacent to each other. The input slice is modified, however the
// result is in the returned slice since it may have its size changed due to
// uniquification.
func uniquifyGroupedXYs(xys []XY) []XY {
	if len(xys) == 0 {
		return xys
	}
	n := 1
	for i := 1; i < len(xys); i++ {
		if xys[i] != xys[i-1] {
			xys[n] = xys[i]
			n++
		}
	}
	return xys[:n]
}

func sequenceToXYs(seq Sequence) []XY {
	n := seq.Length()
	xys := make([]XY, seq.Length())
	for i := 0; i < n; i++ {
		xys[i] = seq.GetXY(i)
	}
	return xys
}

// fastMin is a faster but not functionally identical version of math.Min.
func fastMin(a, b float64) float64 {
	if math.IsNaN(a) || a < b {
		return a
	}
	return b
}

// fastMax is a faster but not functionally identical version of math.Max.
func fastMax(a, b float64) float64 {
	if math.IsNaN(a) || a > b {
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

func arbitraryControlPoint(g Geometry) Point {
	switch typ := g.Type(); typ {
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if pt := arbitraryControlPoint(gc.GeometryN(i)); !pt.IsEmpty() {
				return pt
			}
		}
		return Point{}
	case TypePoint:
		return g.MustAsPoint()
	case TypeLineString:
		return g.MustAsLineString().StartPoint()
	case TypePolygon:
		return g.MustAsPolygon().ExteriorRing().StartPoint()
	case TypeMultiPoint:
		for _, pt := range g.MustAsMultiPoint().Dump() {
			if !pt.IsEmpty() {
				return pt
			}
		}
		return Point{}
	case TypeMultiLineString:
		for _, ls := range g.MustAsMultiLineString().Dump() {
			if pt := ls.StartPoint(); !pt.IsEmpty() {
				return pt
			}
		}
		return Point{}
	case TypeMultiPolygon:
		for _, p := range g.MustAsMultiPolygon().Dump() {
			pt := p.ExteriorRing().StartPoint()
			if !pt.IsEmpty() {
				return pt
			}
		}
		return Point{}
	default:
		panic(fmt.Sprintf("invalid geometry type: %d", int(typ)))
	}
}
