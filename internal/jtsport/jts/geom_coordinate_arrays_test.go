package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

var (
	coordArraysTestCoords1 = []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(1, 1),
		jts.Geom_NewCoordinateWithXY(2, 2),
		jts.Geom_NewCoordinateWithXY(3, 3),
	}
	coordArraysTestCoordsEmpty = []*jts.Geom_Coordinate{}
)

func TestCoordinateArraysPtNotInList1(t *testing.T) {
	result := jts.Geom_CoordinateArrays_PtNotInList(
		[]*jts.Geom_Coordinate{
			jts.Geom_NewCoordinateWithXY(1, 1),
			jts.Geom_NewCoordinateWithXY(2, 2),
			jts.Geom_NewCoordinateWithXY(3, 3),
		},
		[]*jts.Geom_Coordinate{
			jts.Geom_NewCoordinateWithXY(1, 1),
			jts.Geom_NewCoordinateWithXY(1, 2),
			jts.Geom_NewCoordinateWithXY(1, 3),
		},
	)
	junit.AssertTrue(t, result.Equals2D(jts.Geom_NewCoordinateWithXY(2, 2)))
}

func TestCoordinateArraysPtNotInList2(t *testing.T) {
	result := jts.Geom_CoordinateArrays_PtNotInList(
		[]*jts.Geom_Coordinate{
			jts.Geom_NewCoordinateWithXY(1, 1),
			jts.Geom_NewCoordinateWithXY(2, 2),
			jts.Geom_NewCoordinateWithXY(3, 3),
		},
		[]*jts.Geom_Coordinate{
			jts.Geom_NewCoordinateWithXY(1, 1),
			jts.Geom_NewCoordinateWithXY(2, 2),
			jts.Geom_NewCoordinateWithXY(3, 3),
		},
	)
	junit.AssertTrue(t, result == nil)
}

func TestCoordinateArraysEnvelope1(t *testing.T) {
	junit.AssertTrue(t, jts.Geom_CoordinateArrays_Envelope(coordArraysTestCoords1).Equals(jts.Geom_NewEnvelopeFromXY(1, 3, 1, 3)))
}

func TestCoordinateArraysEnvelopeEmpty(t *testing.T) {
	junit.AssertTrue(t, jts.Geom_CoordinateArrays_Envelope(coordArraysTestCoordsEmpty).Equals(jts.Geom_NewEnvelope()))
}

func TestCoordinateArraysIntersectionEnvelope1(t *testing.T) {
	junit.AssertTrue(t, jts.Geom_CoordinateArrays_Equals(
		jts.Geom_CoordinateArrays_Intersection(coordArraysTestCoords1, jts.Geom_NewEnvelopeFromXY(1, 2, 1, 2)),
		[]*jts.Geom_Coordinate{jts.Geom_NewCoordinateWithXY(1, 1), jts.Geom_NewCoordinateWithXY(2, 2)}))
}

func TestCoordinateArraysIntersectionEnvelopeDisjoint(t *testing.T) {
	junit.AssertTrue(t, jts.Geom_CoordinateArrays_Equals(
		jts.Geom_CoordinateArrays_Intersection(coordArraysTestCoords1, jts.Geom_NewEnvelopeFromXY(10, 20, 10, 20)), coordArraysTestCoordsEmpty))
}

func TestCoordinateArraysIntersectionEmptyEnvelope(t *testing.T) {
	junit.AssertTrue(t, jts.Geom_CoordinateArrays_Equals(
		jts.Geom_CoordinateArrays_Intersection(coordArraysTestCoordsEmpty, jts.Geom_NewEnvelopeFromXY(1, 2, 1, 2)), coordArraysTestCoordsEmpty))
}

func TestCoordinateArraysIntersectionCoordsEmptyEnvelope(t *testing.T) {
	junit.AssertTrue(t, jts.Geom_CoordinateArrays_Equals(
		jts.Geom_CoordinateArrays_Intersection(coordArraysTestCoords1, jts.Geom_NewEnvelope()), coordArraysTestCoordsEmpty))
}

func TestCoordinateArraysReverseEmpty(t *testing.T) {
	pts := []*jts.Geom_Coordinate{}
	checkReversed(t, pts)
}

func TestCoordinateArraysReverseSingleElement(t *testing.T) {
	pts := []*jts.Geom_Coordinate{jts.Geom_NewCoordinateWithXY(1, 1)}
	checkReversed(t, pts)
}

func TestCoordinateArraysReverse2(t *testing.T) {
	pts := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(1, 1),
		jts.Geom_NewCoordinateWithXY(2, 2),
	}
	checkReversed(t, pts)
}

func TestCoordinateArraysReverse3(t *testing.T) {
	pts := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(1, 1),
		jts.Geom_NewCoordinateWithXY(2, 2),
		jts.Geom_NewCoordinateWithXY(3, 3),
	}
	checkReversed(t, pts)
}

func checkReversed(t *testing.T, pts []*jts.Geom_Coordinate) {
	ptsRev := jts.Geom_CoordinateArrays_CopyDeep(pts)
	jts.Geom_CoordinateArrays_Reverse(ptsRev)
	junit.AssertEquals(t, len(pts), len(ptsRev))
	length := len(pts)
	for i := range pts {
		checkCoordArraysEqualXY(t, pts[i], ptsRev[length-1-i])
	}
}

func checkCoordArraysEqualXY(t *testing.T, c1, c2 *jts.Geom_Coordinate) {
	junit.AssertEquals(t, c1.GetX(), c2.GetX())
	junit.AssertEquals(t, c1.GetY(), c2.GetY())
}

func TestCoordinateArraysScrollRing(t *testing.T) {
	sequence := createCircle(jts.Geom_NewCoordinateWithXY(10, 10), 9.0)
	scrolled := createCircle(jts.Geom_NewCoordinateWithXY(10, 10), 9.0)

	jts.Geom_CoordinateArrays_ScrollToIndex(scrolled, 12)

	io := 12
	for is := 0; is < len(scrolled)-1; is++ {
		checkCoordinateAt(t, sequence, io, scrolled, is)
		io++
		io %= len(scrolled) - 1
	}
	checkCoordinateAt(t, scrolled, 0, scrolled, len(scrolled)-1)
}

func TestCoordinateArraysScroll(t *testing.T) {
	sequence := createCircularString(jts.Geom_NewCoordinateWithXY(20, 20), 7.0, 0.1, 22)
	scrolled := createCircularString(jts.Geom_NewCoordinateWithXY(20, 20), 7.0, 0.1, 22)

	jts.Geom_CoordinateArrays_ScrollToIndex(scrolled, 12)

	io := 12
	for is := 0; is < len(scrolled)-1; is++ {
		checkCoordinateAt(t, sequence, io, scrolled, is)
		io++
		io %= len(scrolled)
	}
}

func TestCoordinateArraysEnforceConsistency(t *testing.T) {
	array := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXYZ(1.0, 1.0, 0.0),
		jts.Geom_NewCoordinateXYM3DWithXYM(2.0, 2.0, 1.0).Geom_Coordinate,
	}
	array2 := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateXY2DWithXY(1.0, 1.0).Geom_Coordinate,
		jts.Geom_NewCoordinateXY2DWithXY(2.0, 2.0).Geom_Coordinate,
	}

	// Process into array with dimension 3 and measures 1.
	jts.Geom_CoordinateArrays_EnforceConsistency(array)
	junit.AssertEquals(t, 3, jts.Geom_CoordinateArrays_Dimension(array))
	junit.AssertEquals(t, 1, jts.Geom_CoordinateArrays_Measures(array))

	jts.Geom_CoordinateArrays_EnforceConsistency(array2)

	fixed := jts.Geom_CoordinateArrays_EnforceConsistencyWithDimension(array2, 2, 0)
	// No processing required, should be same slice.
	if &fixed[0] != &array2[0] {
		t.Error("assertSame: expected same array when no processing required")
	}

	fixed = jts.Geom_CoordinateArrays_EnforceConsistencyWithDimension(array, 3, 0)
	// Copied into new array.
	junit.AssertTrue(t, &fixed[0] != &array[0])
	// Processing needed to CoordinateXYZM.
	junit.AssertTrue(t, array[0] != fixed[0])
	junit.AssertTrue(t, array[1] != fixed[1])
}

func checkCoordinateAt(t *testing.T, seq1 []*jts.Geom_Coordinate, pos1 int, seq2 []*jts.Geom_Coordinate, pos2 int) {
	c1, c2 := seq1[pos1], seq2[pos2]
	junit.AssertEquals(t, c1.GetX(), c2.GetX())
	junit.AssertEquals(t, c1.GetY(), c2.GetY())
}

func createCircle(center *jts.Geom_Coordinate, radius float64) []*jts.Geom_Coordinate {
	res := createCircularString(center, radius, 0.0, 49)
	res[48] = res[0].Copy()
	return res
}

func createCircularString(center *jts.Geom_Coordinate, radius, startAngle float64, numPoints int) []*jts.Geom_Coordinate {
	const numSegmentsCircle = 48
	const angleCircle = 2 * math.Pi
	const angleStep = angleCircle / numSegmentsCircle

	sequence := make([]*jts.Geom_Coordinate, numPoints)
	pm := jts.Geom_NewPrecisionModelWithScale(1000)
	angle := startAngle
	for i := 0; i < numPoints; i++ {
		dx := math.Cos(angle) * radius
		dy := math.Sin(angle) * radius
		sequence[i] = jts.Geom_NewCoordinateXY2DWithXY(
			pm.MakePrecise(center.X+dx),
			pm.MakePrecise(center.Y+dy),
		).Geom_Coordinate
		angle += angleStep
		angle = math.Mod(angle, angleCircle)
	}
	return sequence
}
