package jts

import "testing"

func TestDistanceOp_DisjointCollinearSegments(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testDisjointCollinearSegments()
}

func TestDistanceOp_PolygonsDisjoint(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testPolygonsDisjoint()
}

func TestDistanceOp_PolygonsOverlapping(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testPolygonsOverlapping()
}

func TestDistanceOp_LinesIdentical(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testLinesIdentical()
}

func TestDistanceOp_Empty(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testEmpty()
}

func TestDistanceOp_ClosestPoints1(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testClosestPoints1()
}

func TestDistanceOp_ClosestPoints2(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testClosestPoints2()
}

func TestDistanceOp_ClosestPoints3(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testClosestPoints3()
}

func TestDistanceOp_ClosestPoints4(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testClosestPoints4()
}

func TestDistanceOp_ClosestPoints5(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testClosestPoints5()
}

func TestDistanceOp_ClosestPoints6(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testClosestPoints6()
}

func TestDistanceOp_ClosestPoints7(t *testing.T) {
	bt := newOperationDistance_distanceTest(t)
	bt.testClosestPoints7()
}

func newOperationDistance_distanceTest(t *testing.T) *operationDistance_baseDistanceTest {
	return &operationDistance_baseDistanceTest{
		t: t,
		distanceFn: func(g1, g2 *Geom_Geometry) float64 {
			return g1.Distance(g2)
		},
		isWithinDistFn: func(g1, g2 *Geom_Geometry, distance float64) bool {
			return g1.IsWithinDistance(g2, distance)
		},
		nearestPtsFn: func(g1, g2 *Geom_Geometry) []*Geom_Coordinate {
			return OperationDistance_DistanceOp_NearestPoints(g1, g2)
		},
	}
}
