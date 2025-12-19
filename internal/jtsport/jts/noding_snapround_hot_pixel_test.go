package jts

import "testing"

func TestHotPixelBelow(t *testing.T) {
	checkHotPixelIntersects(t, false, 1, 1, 100,
		1, 0.98, 3, 0.5)
}

func TestHotPixelAbove(t *testing.T) {
	checkHotPixelIntersects(t, false, 1, 1, 100,
		1, 1.011, 3, 1.5)
}

func TestHotPixelRightSideVerticalTouchAbove(t *testing.T) {
	checkHotPixelIntersects(t, false, 1.2, 1.2, 10,
		1.25, 1.25, 1.25, 2)
}

func TestHotPixelRightSideVerticalTouchBelow(t *testing.T) {
	checkHotPixelIntersects(t, false, 1.2, 1.2, 10,
		1.25, 0, 1.25, 1.15)
}

func TestHotPixelRightSideVerticalOverlap(t *testing.T) {
	checkHotPixelIntersects(t, false, 1.2, 1.2, 10,
		1.25, 0, 1.25, 1.5)
}

func TestHotPixelTopSideHorizontalTouchRight(t *testing.T) {
	checkHotPixelIntersects(t, false, 1.2, 1.2, 10,
		1.25, 1.25, 2, 1.25)
}

func TestHotPixelTopSideHorizontalTouchLeft(t *testing.T) {
	checkHotPixelIntersects(t, false, 1.2, 1.2, 10,
		0, 1.25, 1.15, 1.25)
}

func TestHotPixelTopSideHorizontalOverlap(t *testing.T) {
	checkHotPixelIntersects(t, false, 1.2, 1.2, 10,
		0, 1.25, 1.9, 1.25)
}

func TestHotPixelLeftSideVerticalTouchAbove(t *testing.T) {
	checkHotPixelIntersects(t, false, 1.2, 1.2, 10,
		1.15, 1.25, 1.15, 2)
}

func TestHotPixelLeftSideVerticalOverlap(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		1.15, 0, 1.15, 1.8)
}

func TestHotPixelLeftSideVerticalTouchBelow(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		1.15, 0, 1.15, 1.15)
}

func TestHotPixelLeftSideCrossRight(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		0, 1.19, 2, 1.21)
}

func TestHotPixelLeftSideCrossTop(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		0.8, 0.8, 1.3, 1.39)
}

func TestHotPixelLeftSideCrossBottom(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		1, 1.5, 1.3, 0.9)
}

func TestHotPixelBottomSideHorizontalTouchRight(t *testing.T) {
	checkHotPixelIntersects(t, false, 1.2, 1.2, 10,
		1.25, 1.15, 2, 1.15)
}

func TestHotPixelBottomSideHorizontalTouchLeft(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		0, 1.15, 1.15, 1.15)
}

func TestHotPixelBottomSideHorizontalOverlapLeft(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		0, 1.15, 1.2, 1.15)
}

func TestHotPixelBottomSideHorizontalOverlap(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		0, 1.15, 1.9, 1.15)
}

func TestHotPixelBottomSideHorizontalOverlapRight(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		1.2, 1.15, 1.4, 1.15)
}

func TestHotPixelBottomSideCrossRight(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		1.1, 1, 1.4, 1.4)
}

func TestHotPixelBottomSideCrossTop(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		1.1, 0.9, 1.3, 1.6)
}

func TestHotPixelDiagonalDown(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		0.9, 1.5, 1.4, 1)
}

func TestHotPixelDiagonalUp(t *testing.T) {
	checkHotPixelIntersects(t, true, 1.2, 1.2, 10,
		0.9, 0.9, 1.5, 1.5)
}

func TestHotPixelCornerULEndInside(t *testing.T) {
	checkHotPixelIntersects(t, true, 1, 1, 10,
		0.7, 1.3, 0.98, 1.02)
}

func TestHotPixelCornerLLEndInside(t *testing.T) {
	checkHotPixelIntersects(t, true, 1, 1, 10,
		0.8, 0.8, 0.98, 0.98)
}

func TestHotPixelCornerURStartInside(t *testing.T) {
	checkHotPixelIntersects(t, true, 1, 1, 10,
		1.02, 1.02, 1.3, 1.3)
}

func TestHotPixelCornerLRStartInside(t *testing.T) {
	checkHotPixelIntersects(t, true, 1, 1, 10,
		1.02, 0.98, 1.3, 0.7)
}

func TestHotPixelCornerLLTangent(t *testing.T) {
	checkHotPixelIntersects(t, true, 1, 1, 10,
		0.9, 1, 1, 0.9)
}

func TestHotPixelCornerLLTangentNoTouch(t *testing.T) {
	checkHotPixelIntersects(t, false, 1, 1, 10,
		0.9, 0.9, 1, 0.9)
}

func TestHotPixelCornerULTangent(t *testing.T) {
	// Does not intersect due to open top.
	checkHotPixelIntersects(t, false, 1, 1, 10,
		0.9, 1, 1, 1.1)
}

func TestHotPixelCornerURTangent(t *testing.T) {
	// Does not intersect due to open top.
	checkHotPixelIntersects(t, false, 1, 1, 10,
		1, 1.1, 1.1, 1)
}

func TestHotPixelCornerLRTangent(t *testing.T) {
	// Does not intersect due to open right side.
	checkHotPixelIntersects(t, false, 1, 1, 10,
		1, 0.9, 1.1, 1)
}

func TestHotPixelCornerULTouchEnd(t *testing.T) {
	// Does not intersect due to bounding box check for open top.
	checkHotPixelIntersects(t, false, 1, 1, 10,
		0.9, 1.1, 0.95, 1.05)
}

func checkHotPixelIntersects(t *testing.T, expected bool,
	x, y, scale float64,
	x1, y1, x2, y2 float64) {
	t.Helper()
	hp := NodingSnapround_NewHotPixel(Geom_NewCoordinateWithXY(x, y), scale)
	p1 := Geom_NewCoordinateWithXY(x1, y1)
	p2 := Geom_NewCoordinateWithXY(x2, y2)
	actual := hp.IntersectsSegment(p1, p2)
	if actual != expected {
		t.Errorf("expected %v but got %v for HotPixel(%v,%v,%v).IntersectsSegment((%v,%v),(%v,%v))",
			expected, actual, x, y, scale, x1, y1, x2, y2)
	}
}
