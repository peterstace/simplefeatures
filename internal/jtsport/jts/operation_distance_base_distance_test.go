package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

// operationDistance_baseDistanceTest provides abstract test methods for distance tests.
type operationDistance_baseDistanceTest struct {
	t              *testing.T
	distanceFn     func(g1, g2 *Geom_Geometry) float64
	isWithinDistFn func(g1, g2 *Geom_Geometry, distance float64) bool
	nearestPtsFn   func(g1, g2 *Geom_Geometry) []*Geom_Coordinate
}

func (bt *operationDistance_baseDistanceTest) testDisjointCollinearSegments() {
	g1 := operationDistance_baseDistanceTest_read("LINESTRING (0.0 0.0, 9.9 1.4)")
	g2 := operationDistance_baseDistanceTest_read("LINESTRING (11.88 1.68, 21.78 3.08)")

	dist := bt.distanceFn(g1, g2)
	junit.AssertEqualsFloat64(bt.t, 1.9996999774966246, dist, 0.0001)

	junit.AssertTrue(bt.t, !bt.isWithinDistFn(g1, g2, 1))
	junit.AssertTrue(bt.t, bt.isWithinDistFn(g1, g2, 3))
}

func (bt *operationDistance_baseDistanceTest) testPolygonsDisjoint() {
	g1 := operationDistance_baseDistanceTest_read("POLYGON ((40 320, 200 380, 320 80, 40 40, 40 320),  (180 280, 80 280, 100 100, 220 140, 180 280))")
	g2 := operationDistance_baseDistanceTest_read("POLYGON ((160 240, 120 240, 120 160, 160 140, 160 240))")
	junit.AssertEqualsFloat64(bt.t, 18.97366596, bt.distanceFn(g1, g2), 1e-5)

	junit.AssertTrue(bt.t, !bt.isWithinDistFn(g1, g2, 0))
	junit.AssertTrue(bt.t, !bt.isWithinDistFn(g1, g2, 10))
	junit.AssertTrue(bt.t, bt.isWithinDistFn(g1, g2, 20))
}

func (bt *operationDistance_baseDistanceTest) testPolygonsOverlapping() {
	g1 := operationDistance_baseDistanceTest_read("POLYGON ((40 320, 200 380, 320 80, 40 40, 40 320),  (180 280, 80 280, 100 100, 220 140, 180 280))")
	g3 := operationDistance_baseDistanceTest_read("POLYGON ((160 240, 120 240, 120 160, 180 100, 160 240))")

	junit.AssertEqualsFloat64(bt.t, 0.0, bt.distanceFn(g1, g3), 1e-9)
	junit.AssertTrue(bt.t, bt.isWithinDistFn(g1, g3, 0.0))
}

func (bt *operationDistance_baseDistanceTest) testLinesIdentical() {
	l1 := operationDistance_baseDistanceTest_read("LINESTRING(10 10, 20 20, 30 40)")
	junit.AssertEqualsFloat64(bt.t, 0.0, bt.distanceFn(l1, l1), 1e-5)

	junit.AssertTrue(bt.t, bt.isWithinDistFn(l1, l1, 0))
}

func (bt *operationDistance_baseDistanceTest) testEmpty() {
	g1 := operationDistance_baseDistanceTest_read("POINT (0 0)")
	g2 := operationDistance_baseDistanceTest_read("POLYGON EMPTY")
	junit.AssertEqualsFloat64(bt.t, 0.0, g1.Distance(g2), 0.0)
}

func (bt *operationDistance_baseDistanceTest) testClosestPoints1() {
	bt.checkDistanceNearestPoints("POLYGON ((200 180, 60 140, 60 260, 200 180))", "POINT (140 280)", 57.05597791103589, Geom_NewCoordinateWithXY(111.6923076923077, 230.46153846153845), Geom_NewCoordinateWithXY(140, 280))
}

func (bt *operationDistance_baseDistanceTest) testClosestPoints2() {
	bt.checkDistanceNearestPoints("POLYGON ((200 180, 60 140, 60 260, 200 180))", "MULTIPOINT ((140 280), (140 320))", 57.05597791103589, Geom_NewCoordinateWithXY(111.6923076923077, 230.46153846153845), Geom_NewCoordinateWithXY(140, 280))
}

func (bt *operationDistance_baseDistanceTest) testClosestPoints3() {
	bt.checkDistanceNearestPoints("LINESTRING (100 100, 200 100, 200 200, 100 200, 100 100)", "POINT (10 10)", 127.27922061357856, Geom_NewCoordinateWithXY(100, 100), Geom_NewCoordinateWithXY(10, 10))
}

func (bt *operationDistance_baseDistanceTest) testClosestPoints4() {
	bt.checkDistanceNearestPoints("LINESTRING (100 100, 200 200)", "LINESTRING (100 200, 200 100)", 0.0, Geom_NewCoordinateWithXY(150, 150), Geom_NewCoordinateWithXY(150, 150))
}

func (bt *operationDistance_baseDistanceTest) testClosestPoints5() {
	bt.checkDistanceNearestPoints("LINESTRING (100 100, 200 200)", "LINESTRING (150 121, 200 0)", 20.506096654409877, Geom_NewCoordinateWithXY(135.5, 135.5), Geom_NewCoordinateWithXY(150, 121))
}

func (bt *operationDistance_baseDistanceTest) testClosestPoints6() {
	bt.checkDistanceNearestPoints("POLYGON ((76 185, 125 283, 331 276, 324 122, 177 70, 184 155, 69 123, 76 185), (267 237, 148 248, 135 185, 223 189, 251 151, 286 183, 267 237))", "LINESTRING (153 204, 185 224, 209 207, 238 222, 254 186)", 13.788860460124573, Geom_NewCoordinateWithXY(139.4956500724988, 206.78661188980183), Geom_NewCoordinateWithXY(153, 204))
}

func (bt *operationDistance_baseDistanceTest) testClosestPoints7() {
	bt.checkDistanceNearestPoints("POLYGON ((76 185, 125 283, 331 276, 324 122, 177 70, 184 155, 69 123, 76 185), (267 237, 148 248, 135 185, 223 189, 251 151, 286 183, 267 237))", "LINESTRING (120 215, 185 224, 209 207, 238 222, 254 186)", 0.0, Geom_NewCoordinateWithXY(120, 215), Geom_NewCoordinateWithXY(120, 215))
}

const operationDistance_baseDistanceTest_TOLERANCE = 1e-10

func (bt *operationDistance_baseDistanceTest) checkDistanceNearestPoints(wkt0, wkt1 string, distance float64, p0, p1 *Geom_Coordinate) {
	g0 := operationDistance_baseDistanceTest_read(wkt0)
	g1 := operationDistance_baseDistanceTest_read(wkt1)

	nearestPoints := bt.nearestPtsFn(g0, g1)

	junit.AssertEqualsFloat64(bt.t, distance, nearestPoints[0].Distance(nearestPoints[1]), operationDistance_baseDistanceTest_TOLERANCE)
	junit.AssertEqualsFloat64(bt.t, p0.X, nearestPoints[0].X, operationDistance_baseDistanceTest_TOLERANCE)
	junit.AssertEqualsFloat64(bt.t, p0.Y, nearestPoints[0].Y, operationDistance_baseDistanceTest_TOLERANCE)
	junit.AssertEqualsFloat64(bt.t, p1.X, nearestPoints[1].X, operationDistance_baseDistanceTest_TOLERANCE)
	junit.AssertEqualsFloat64(bt.t, p1.Y, nearestPoints[1].Y, operationDistance_baseDistanceTest_TOLERANCE)
}

func operationDistance_baseDistanceTest_read(wkt string) *Geom_Geometry {
	reader := Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		panic(err)
	}
	return geom
}
