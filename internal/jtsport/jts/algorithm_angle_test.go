package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

const angleTolerance = 1e-5

func p(x, y float64) *jts.Geom_Coordinate {
	return jts.Geom_NewCoordinateWithXY(x, y)
}

func TestAngleAngle(t *testing.T) {
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_Angle(p(10, 0)), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi/4, jts.Algorithm_Angle_Angle(p(10, 10)), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi/2, jts.Algorithm_Angle_Angle(p(0, 10)), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.75*math.Pi, jts.Algorithm_Angle_Angle(p(-10, 10)), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_Angle(p(-10, 0)), angleTolerance)
	junit.AssertEqualsFloat64(t, -3.131592986903128, jts.Algorithm_Angle_Angle(p(-10, -0.1)), angleTolerance)
	junit.AssertEqualsFloat64(t, -0.75*math.Pi, jts.Algorithm_Angle_Angle(p(-10, -10)), angleTolerance)
}

func TestAngleIsAcute(t *testing.T) {
	junit.AssertEquals(t, true, jts.Algorithm_Angle_IsAcute(p(10, 0), p(0, 0), p(5, 10)))
	junit.AssertEquals(t, true, jts.Algorithm_Angle_IsAcute(p(10, 0), p(0, 0), p(5, -10)))
	// Angle of 0.
	junit.AssertEquals(t, true, jts.Algorithm_Angle_IsAcute(p(10, 0), p(0, 0), p(10, 0)))
	junit.AssertEquals(t, false, jts.Algorithm_Angle_IsAcute(p(10, 0), p(0, 0), p(-5, 10)))
	junit.AssertEquals(t, false, jts.Algorithm_Angle_IsAcute(p(10, 0), p(0, 0), p(-5, -10)))
}

func TestAngleNormalizePositive(t *testing.T) {
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_NormalizePositive(0.0), angleTolerance)
	junit.AssertEqualsFloat64(t, 1.5*math.Pi, jts.Algorithm_Angle_NormalizePositive(-0.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_NormalizePositive(-math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.5*math.Pi, jts.Algorithm_Angle_NormalizePositive(-1.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_NormalizePositive(-2*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 1.5*math.Pi, jts.Algorithm_Angle_NormalizePositive(-2.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_NormalizePositive(-3*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_NormalizePositive(-4*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.5*math.Pi, jts.Algorithm_Angle_NormalizePositive(0.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_NormalizePositive(math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 1.5*math.Pi, jts.Algorithm_Angle_NormalizePositive(1.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_NormalizePositive(2*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.5*math.Pi, jts.Algorithm_Angle_NormalizePositive(2.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_NormalizePositive(3*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_NormalizePositive(4*math.Pi), angleTolerance)
}

func TestAngleNormalize(t *testing.T) {
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_Normalize(0.0), angleTolerance)
	junit.AssertEqualsFloat64(t, -0.5*math.Pi, jts.Algorithm_Angle_Normalize(-0.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_Normalize(-math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.5*math.Pi, jts.Algorithm_Angle_Normalize(-1.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_Normalize(-2*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, -0.5*math.Pi, jts.Algorithm_Angle_Normalize(-2.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_Normalize(-3*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_Normalize(-4*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.5*math.Pi, jts.Algorithm_Angle_Normalize(0.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_Normalize(math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, -0.5*math.Pi, jts.Algorithm_Angle_Normalize(1.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_Normalize(2*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.5*math.Pi, jts.Algorithm_Angle_Normalize(2.5*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, math.Pi, jts.Algorithm_Angle_Normalize(3*math.Pi), angleTolerance)
	junit.AssertEqualsFloat64(t, 0.0, jts.Algorithm_Angle_Normalize(4*math.Pi), angleTolerance)
}

func TestAngleInteriorAngle(t *testing.T) {
	p1 := p(1, 2)
	p2 := p(3, 2)
	p3 := p(2, 1)

	// Tests all interior angles of a triangle "POLYGON ((1 2, 3 2, 2 1, 1 2))".
	junit.AssertEqualsFloat64(t, 45, math.Round(jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_InteriorAngle(p1, p2, p3))*100)/100, 0.01)
	junit.AssertEqualsFloat64(t, 90, math.Round(jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_InteriorAngle(p2, p3, p1))*100)/100, 0.01)
	junit.AssertEqualsFloat64(t, 45, math.Round(jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_InteriorAngle(p3, p1, p2))*100)/100, 0.01)
	// Tests interior angles greater than 180 degrees.
	junit.AssertEqualsFloat64(t, 315, math.Round(jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_InteriorAngle(p3, p2, p1))*100)/100, 0.01)
	junit.AssertEqualsFloat64(t, 270, math.Round(jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_InteriorAngle(p1, p3, p2))*100)/100, 0.01)
	junit.AssertEqualsFloat64(t, 315, math.Round(jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_InteriorAngle(p2, p1, p3))*100)/100, 0.01)
}

func TestAngleInteriorAngleTriangleSumProperty(t *testing.T) {
	// Tests that the sum of interior angles of any triangle equals PI (180 degrees).
	// This is a simplified version of testInteriorAngle_randomTriangles which
	// requires RandomPointsBuilder (not ported). Uses predetermined triangles instead.
	// NOTE: Triangles must be in CLOCKWISE order for Algorithm_Angle_InteriorAngle.
	triangles := [][3]*jts.Geom_Coordinate{
		// Same triangle as TestAngleInteriorAngle (clockwise).
		{p(1, 2), p(3, 2), p(2, 1)},
		// Right triangle (clockwise).
		{p(0, 0), p(0, 1), p(1, 0)},
		// Isosceles triangle (clockwise).
		{p(0, 0), p(1, 3), p(2, 0)},
		// Scalene triangle (clockwise).
		{p(0, 0), p(2, 4), p(5, 0)},
		// Thin triangle (clockwise).
		{p(0, 0), p(5, 0.1), p(10, 0)},
		// Large triangle (clockwise).
		{p(100, 200), p(150, 400), p(300, 50)},
		// Triangle with negative coordinates (clockwise).
		{p(-5, -3), p(-1, 4), p(2, -1)},
	}

	for i, tri := range triangles {
		c := tri[:]
		// Ensure triangles are clockwise (negative signed area).
		signedArea := 0.5 * (c[0].X*(c[1].Y-c[2].Y) + c[1].X*(c[2].Y-c[0].Y) + c[2].X*(c[0].Y-c[1].Y))
		if signedArea > 0 {
			// CCW, reverse to CW by swapping c[1] and c[2].
			c[1], c[2] = c[2], c[1]
		}
		sumOfInteriorAngles := jts.Algorithm_Angle_InteriorAngle(c[0], c[1], c[2]) +
			jts.Algorithm_Angle_InteriorAngle(c[1], c[2], c[0]) +
			jts.Algorithm_Angle_InteriorAngle(c[2], c[0], c[1])
		if math.Abs(sumOfInteriorAngles-math.Pi) > 0.01 {
			t.Errorf("triangle %d: sum of interior angles = %v, want %v (PI)", i, sumOfInteriorAngles, math.Pi)
		}
	}
}

func TestAngleBisector(t *testing.T) {
	junit.AssertEqualsFloat64(t, 45, jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_Bisector(p(0, 1), p(0, 0), p(1, 0))), 0.01)
	junit.AssertEqualsFloat64(t, 22.5, jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_Bisector(p(1, 1), p(0, 0), p(1, 0))), 0.01)
	junit.AssertEqualsFloat64(t, 67.5, jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_Bisector(p(-1, 1), p(0, 0), p(1, 0))), 0.01)
	junit.AssertEqualsFloat64(t, -45, jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_Bisector(p(0, -1), p(0, 0), p(1, 0))), 0.01)
	junit.AssertEqualsFloat64(t, 180, jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_Bisector(p(-1, -1), p(0, 0), p(-1, 1))), 0.01)
	junit.AssertEqualsFloat64(t, 45, jts.Algorithm_Angle_ToDegrees(jts.Algorithm_Angle_Bisector(p(13, 10), p(10, 10), p(10, 20))), 0.01)
}

func TestAngleSinCosSnap(t *testing.T) {
	// -720 to 720 degrees with 1 degree increments.
	for angdeg := -720; angdeg <= 720; angdeg++ {
		ang := jts.Algorithm_Angle_ToRadians(float64(angdeg))

		rSin := jts.Algorithm_Angle_SinSnap(ang)
		rCos := jts.Algorithm_Angle_CosSnap(ang)

		cSin := math.Sin(ang)
		cCos := math.Cos(ang)
		if angdeg%90 == 0 {
			// Not always the same for multiples of 90 degrees.
			junit.AssertTrue(t, math.Abs(rSin-cSin) < 1e-15)
			junit.AssertTrue(t, math.Abs(rCos-cCos) < 1e-15)
		} else {
			junit.AssertEquals(t, cSin, rSin)
			junit.AssertEquals(t, cCos, rCos)
		}
	}

	// Use radian increments that don't snap to exact degrees or zero.
	for angrad := -6.3; angrad < 6.3; angrad += 0.013 {
		rSin := jts.Algorithm_Angle_SinSnap(angrad)
		rCos := jts.Algorithm_Angle_CosSnap(angrad)

		junit.AssertEquals(t, math.Sin(angrad), rSin)
		junit.AssertEquals(t, math.Cos(angrad), rCos)
	}
}
