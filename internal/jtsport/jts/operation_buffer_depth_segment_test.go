package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestDepthSegmentCompareTipToTail(t *testing.T) {
	ds0 := depthSeg(0.7, 0.2, 1.4, 0.9)
	ds1 := depthSeg(0.7, 0.2, 0.3, 1.1)
	checkDepthSegmentCompare(t, ds0, ds1, 1)
}

func TestDepthSegmentCompare2(t *testing.T) {
	ds0 := depthSeg(0.5, 1.0, 0.1, 1.9)
	ds1 := depthSeg(1.0, 0.9, 1.9, 1.4)
	checkDepthSegmentCompare(t, ds0, ds1, -1)
}

func TestDepthSegmentCompareVertical(t *testing.T) {
	ds0 := depthSeg(1, 1, 1, 2)
	ds1 := depthSeg(1, 0, 1, 1)
	checkDepthSegmentCompare(t, ds0, ds1, 1)
}

func TestDepthSegmentCompareOrientBug(t *testing.T) {
	ds0 := depthSeg(146.268, -8.42361, 146.263, -8.3875)
	ds1 := depthSeg(146.269, -8.42889, 146.268, -8.42361)
	checkDepthSegmentCompare(t, ds0, ds1, -1)
}

func TestDepthSegmentCompareEqual(t *testing.T) {
	ds0 := depthSeg(1, 1, 2, 2)
	checkDepthSegmentCompare(t, ds0, ds0, 0)
}

func checkDepthSegmentCompare(t *testing.T, ds0, ds1 *operationBuffer_DepthSegment, expectedComp int) {
	t.Helper()
	junit.AssertTrue(t, ds0.isUpward())
	junit.AssertTrue(t, ds1.isUpward())

	// check compareTo contract - should never have ds1 < ds2 && ds2 < ds1
	comp0 := ds0.compareTo(ds1)
	comp1 := ds1.compareTo(ds0)
	junit.AssertEquals(t, expectedComp, comp0)
	junit.AssertTrue(t, comp0 == -comp1)
}

func depthSeg(x0, y0, x1, y1 float64) *operationBuffer_DepthSegment {
	return operationBuffer_newDepthSegment(Geom_NewLineSegmentFromXY(x0, y0, x1, y1), 0)
}
