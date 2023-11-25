package geom

import "math"

//nolint:unused
func lineLineIntersection(ln1, ln2 line) (XY, bool) {
	// See https://en.wikipedia.org/wiki/Line-line_intersection
	x1 := ln1.a.X
	x2 := ln1.b.X
	x3 := ln2.a.X
	x4 := ln2.b.X
	y1 := ln1.a.Y
	y2 := ln1.b.Y
	y3 := ln2.a.Y
	y4 := ln2.b.Y

	x12 := x1 - x2
	x34 := x3 - x4
	y12 := y1 - y2
	y34 := y3 - y4

	var (
		denom  = det(x12, y12, x34, y34)
		detLn1 = det(x1, y1, x2, y2)
		detLn2 = det(x3, y3, x4, y4)
		xr     = det(detLn1, x12, detLn2, x34) / denom
		yr     = det(detLn1, y12, detLn2, y34) / denom
	)

	// TODO: perhaps this is not symmetric?
	if math.Abs(x12) > math.Abs(y12) {
		if (xr < x1 || xr > x2) && (xr < x2 || xr > x1) {
			return XY{xr, yr}, false
		}
	} else {
		if (yr < y1 || yr > y2) && (yr < y2 || yr > y1) {
			return XY{xr, yr}, false
		}
	}

	if math.Abs(x34) > math.Abs(y34) {
		if (xr < x3 || xr > x4) && (xr < x4 || xr > x3) {
			return XY{xr, yr}, false
		}
	} else {
		if (yr < y3 || yr > y4) && (yr < y4 || yr > y3) {
			return XY{xr, yr}, false
		}
	}

	return XY{xr, yr}, true

	//fmt.Println("DEBUG geom/line_line_intersection.go:30 ((xr >= x1 && xr <= x2) || (xr >= x2 && xr <= x1))", ((xr >= x1 && xr <= x2) || (xr >= x2 && xr <= x1))) // XXX
	//fmt.Println("DEBUG geom/line_line_intersection.go:32 ((xr >= x3 && xr <= x4) || (xr >= x4 && xr <= x3))", ((xr >= x3 && xr <= x4) || (xr >= x4 && xr <= x3))) // XXX
	//fmt.Println("DEBUG geom/line_line_intersection.go:34 ((yr >= y1 && yr <= y2) || (yr >= y2 && yr <= y1))", ((yr >= y1 && yr <= y2) || (yr >= y2 && yr <= y1))) // XXX
	//fmt.Println("DEBUG geom/line_line_intersection.go:36 ((yr >= y3 && yr <= y4) || (yr >= y4 && yr <= y3))", ((yr >= y3 && yr <= y4) || (yr >= y4 && yr <= y3))) // XXX

	//return XY{xr, yr}, true &&
	//	(((xr >= x1 && xr <= x2) || (xr >= x2 && xr <= x1)) &&
	//		((xr >= x3 && xr <= x4) || (xr >= x4 && xr <= x3))) ||
	//	(((yr >= y1 && yr <= y2) || (yr >= y2 && yr <= y1)) &&
	//		((yr >= y3 && yr <= y4) || (yr >= y4 && yr <= y3)))

	// ln1Xd := x12 * x12
	// fmt.Println("DEBUG geom/line_line_intersection.go:39 ln1Xd", ln1Xd) // XXX
	// ln1Yd := y12 * y12
	// fmt.Println("DEBUG geom/line_line_intersection.go:41 ln1Yd", ln1Yd) // XXX
	// ln2Xd := x34 * x34
	// fmt.Println("DEBUG geom/line_line_intersection.go:43 ln2Xd", ln2Xd) // XXX
	// ln2Yd := y34 * y34
	// fmt.Println("DEBUG geom/line_line_intersection.go:45 ln2Yd", ln2Yd) // XXX
	//
	// dx1 := (xr - x1) * (xr - x1)
	// fmt.Println("DEBUG geom/line_line_intersection.go:48 dx1", dx1) // XXX
	// dx2 := (xr - x2) * (xr - x2)
	// fmt.Println("DEBUG geom/line_line_intersection.go:50 dx2", dx2) // XXX
	// dx3 := (xr - x3) * (xr - x3)
	// fmt.Println("DEBUG geom/line_line_intersection.go:52 dx3", dx3) // XXX
	// dx4 := (xr - x4) * (xr - x4)
	// fmt.Println("DEBUG geom/line_line_intersection.go:54 dx4", dx4) // XXX
	//
	// dy1 := (yr - y1) * (yr - y1)
	// fmt.Println("DEBUG geom/line_line_intersection.go:57 dy1", dy1) // XXX
	// dy2 := (yr - y2) * (yr - y2)
	// fmt.Println("DEBUG geom/line_line_intersection.go:59 dy2", dy2) // XXX
	// dy3 := (yr - y3) * (yr - y3)
	// fmt.Println("DEBUG geom/line_line_intersection.go:61 dy3", dy3) // XXX
	// dy4 := (yr - y4) * (yr - y4)
	// fmt.Println("DEBUG geom/line_line_intersection.go:63 dy4", dy4) // XXX
	//
	// fmt.Println("DEBUG geom/line_line_intersection.go:69 dx1 <= ln1Xd", dx1 <= ln1Xd) // XXX
	// fmt.Println("DEBUG geom/line_line_intersection.go:69 dx2 <= ln1Xd", dx2 <= ln1Xd) // XXX
	// fmt.Println("DEBUG geom/line_line_intersection.go:69 dy1 <= ln1Yd", dy1 <= ln1Yd) // XXX
	// fmt.Println("DEBUG geom/line_line_intersection.go:69 dy2 <= ln1Yd", dy2 <= ln1Yd) // XXX
	//
	// fmt.Println("DEBUG geom/line_line_intersection.go:74 dx3 <= ln2Xd", dx3 <= ln2Xd) // XXX
	// fmt.Println("DEBUG geom/line_line_intersection.go:74 dx4 <= ln2Xd", dx4 <= ln2Xd) // XXX
	// fmt.Println("DEBUG geom/line_line_intersection.go:74 dy3 <= ln2Yd", dy3 <= ln2Yd) // XXX
	// fmt.Println("DEBUG geom/line_line_intersection.go:74 dy4 <= ln2Yd", dy4 <= ln2Yd) // XXX
	//
	// return XY{xr, yr}, true &&
	//
	//	((dx1 <= ln1Xd && dx2 <= ln1Xd) || (dy1 <= ln1Yd && dy2 <= ln1Yd)) &&
	//	((dx3 <= ln2Xd && dx4 <= ln2Xd) || (dy3 <= ln2Yd && dy4 <= ln2Yd))
}

// det calculates the determinant of the 2x2 matrix:
//
//	a b
//	c d
//
//nolint:unused
func det(a, b, c, d float64) float64 {
	return float64(a*d) - float64(b*c)
}
