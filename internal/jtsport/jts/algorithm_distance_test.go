package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestDistancePointToLinePerpendicular(t *testing.T) {
	junit.AssertEqualsFloat64(t, 0.5, jts.Algorithm_Distance_PointToLinePerpendicular(
		jts.Geom_NewCoordinateWithXY(0.5, 0.5), jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_NewCoordinateWithXY(1, 0)), 0.000001)
	junit.AssertEqualsFloat64(t, 0.5, jts.Algorithm_Distance_PointToLinePerpendicular(
		jts.Geom_NewCoordinateWithXY(3.5, 0.5), jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_NewCoordinateWithXY(1, 0)), 0.000001)
	junit.AssertEqualsFloat64(t, 0.707106, jts.Algorithm_Distance_PointToLinePerpendicular(
		jts.Geom_NewCoordinateWithXY(1, 0), jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_NewCoordinateWithXY(1, 1)), 0.000001)
}

func TestDistancePointToSegment(t *testing.T) {
	junit.AssertEqualsFloat64(t, 0.5, jts.Algorithm_Distance_PointToSegment(
		jts.Geom_NewCoordinateWithXY(0.5, 0.5), jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_NewCoordinateWithXY(1, 0)), 0.000001)
	junit.AssertEqualsFloat64(t, 1.0, jts.Algorithm_Distance_PointToSegment(
		jts.Geom_NewCoordinateWithXY(2, 0), jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_NewCoordinateWithXY(1, 0)), 0.000001)
}

func TestDistanceSegmentToSegmentDisjointCollinear(t *testing.T) {
	junit.AssertEqualsFloat64(t, 1.999699, jts.Algorithm_Distance_SegmentToSegment(
		jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_NewCoordinateWithXY(9.9, 1.4),
		jts.Geom_NewCoordinateWithXY(11.88, 1.68), jts.Geom_NewCoordinateWithXY(21.78, 3.08)), 0.000001)
}
