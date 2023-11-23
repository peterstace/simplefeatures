package geom

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

	var (
		denom = det(x1-x2, y1-y2, x3-x4, y3-y4)
		detA  = det(x1, y1, x2, y2)
		detB  = det(x3, y3, x4, y4)
		xr    = det(detA, x1-x2, detB, x3-x4) / denom
		xy    = det(detA, y1-y2, detB, y3-y4) / denom
	)

	inX12 := (xr >= x1 && xr <= x2) || (xr >= x2 && xr <= x1)
	inX34 := (xr >= x3 && xr <= x4) || (xr >= x4 && xr <= x3)
	inY12 := (xy >= y1 && xy <= y2) || (xy >= y2 && xy <= y1)
	inY34 := (xy >= y3 && xy <= y4) || (xy >= y4 && xy <= y3)
	return XY{xr, xy}, inX12 && inX34 && inY12 && inY34
}

// det calculates the determinant of the 2x2 matrix:
//
//	a b
//	c d
//
//nolint:unused
func det(a, b, c, d float64) float64 {
	return a*d - b*c
}
