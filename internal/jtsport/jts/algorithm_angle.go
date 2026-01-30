package jts

import "math"

// Utility functions for working with angles.
// Unless otherwise noted, methods in this package express angles in radians.

// Angle constants.
const (
	// Algorithm_Angle_PiTimes2 is the value of 2*Pi.
	Algorithm_Angle_PiTimes2 = 2.0 * math.Pi
	// Algorithm_Angle_PiOver2 is the value of Pi/2.
	Algorithm_Angle_PiOver2 = math.Pi / 2.0
	// Algorithm_Angle_PiOver4 is the value of Pi/4.
	Algorithm_Angle_PiOver4 = math.Pi / 4.0
)

// Orientation constants (duplicated from Orientation for convenience).
const (
	// Algorithm_Angle_Counterclockwise represents counterclockwise orientation.
	Algorithm_Angle_Counterclockwise = Algorithm_Orientation_Counterclockwise
	// Algorithm_Angle_Clockwise represents clockwise orientation.
	Algorithm_Angle_Clockwise = Algorithm_Orientation_Clockwise
	// Algorithm_Angle_None represents no orientation (collinear).
	Algorithm_Angle_None = Algorithm_Orientation_Collinear
)

// Algorithm_Angle_ToDegrees converts from radians to degrees.
func Algorithm_Angle_ToDegrees(radians float64) float64 {
	return (radians * 180) / math.Pi
}

// Algorithm_Angle_ToRadians converts from degrees to radians.
func Algorithm_Angle_ToRadians(angleDegrees float64) float64 {
	return (angleDegrees * math.Pi) / 180.0
}

// Algorithm_Angle_AngleBetweenPoints returns the angle of the vector from p0 to p1,
// relative to the positive X-axis. The angle is normalized to be in the range
// [ -Pi, Pi ].
func Algorithm_Angle_AngleBetweenPoints(p0, p1 *Geom_Coordinate) float64 {
	dx := p1.GetX() - p0.GetX()
	dy := p1.GetY() - p0.GetY()
	return math.Atan2(dy, dx)
}

// Algorithm_Angle_Angle returns the angle of the vector from (0,0) to p,
// relative to the positive X-axis. The angle is normalized to be in the range
// ( -Pi, Pi ].
func Algorithm_Angle_Angle(p *Geom_Coordinate) float64 {
	return math.Atan2(p.GetY(), p.GetX())
}

// Algorithm_Angle_IsAcute tests whether the angle between p0-p1-p2 is acute.
// An angle is acute if it is less than 90 degrees.
//
// Note: this implementation is not precise (deterministic) for angles very
// close to 90 degrees.
func Algorithm_Angle_IsAcute(p0, p1, p2 *Geom_Coordinate) bool {
	// Relies on fact that A dot B is positive if A ang B is acute.
	dx0 := p0.GetX() - p1.GetX()
	dy0 := p0.GetY() - p1.GetY()
	dx1 := p2.GetX() - p1.GetX()
	dy1 := p2.GetY() - p1.GetY()
	dotprod := dx0*dx1 + dy0*dy1
	return dotprod > 0
}

// Algorithm_Angle_IsObtuse tests whether the angle between p0-p1-p2 is obtuse.
// An angle is obtuse if it is greater than 90 degrees.
//
// Note: this implementation is not precise (deterministic) for angles very
// close to 90 degrees.
func Algorithm_Angle_IsObtuse(p0, p1, p2 *Geom_Coordinate) bool {
	// Relies on fact that A dot B is negative if A ang B is obtuse.
	dx0 := p0.GetX() - p1.GetX()
	dy0 := p0.GetY() - p1.GetY()
	dx1 := p2.GetX() - p1.GetX()
	dy1 := p2.GetY() - p1.GetY()
	dotprod := dx0*dx1 + dy0*dy1
	return dotprod < 0
}

// Algorithm_Angle_AngleBetween returns the unoriented smallest angle between
// two vectors. The computed angle will be in the range [0, Pi).
func Algorithm_Angle_AngleBetween(tip1, tail, tip2 *Geom_Coordinate) float64 {
	a1 := Algorithm_Angle_AngleBetweenPoints(tail, tip1)
	a2 := Algorithm_Angle_AngleBetweenPoints(tail, tip2)
	return Algorithm_Angle_Diff(a1, a2)
}

// Algorithm_Angle_AngleBetweenOriented returns the oriented smallest angle
// between two vectors. The computed angle will be in the range (-Pi, Pi].
// A positive result corresponds to a counterclockwise (CCW) rotation from v1
// to v2; a negative result corresponds to a clockwise (CW) rotation; a zero
// result corresponds to no rotation.
func Algorithm_Angle_AngleBetweenOriented(tip1, tail, tip2 *Geom_Coordinate) float64 {
	a1 := Algorithm_Angle_AngleBetweenPoints(tail, tip1)
	a2 := Algorithm_Angle_AngleBetweenPoints(tail, tip2)
	angDel := a2 - a1

	// Normalize, maintaining orientation.
	if angDel <= -math.Pi {
		return angDel + Algorithm_Angle_PiTimes2
	}
	if angDel > math.Pi {
		return angDel - Algorithm_Angle_PiTimes2
	}
	return angDel
}

// Algorithm_Angle_Bisector computes the angle of the unoriented bisector of
// the smallest angle between two vectors. The computed angle will be in the
// range (-Pi, Pi].
func Algorithm_Angle_Bisector(tip1, tail, tip2 *Geom_Coordinate) float64 {
	angDel := Algorithm_Angle_AngleBetweenOriented(tip1, tail, tip2)
	angBi := Algorithm_Angle_AngleBetweenPoints(tail, tip1) + angDel/2
	return Algorithm_Angle_Normalize(angBi)
}

// Algorithm_Angle_InteriorAngle computes the interior angle between two
// segments of a ring. The ring is assumed to be oriented in a clockwise
// direction. The computed angle will be in the range [0, 2Pi].
func Algorithm_Angle_InteriorAngle(p0, p1, p2 *Geom_Coordinate) float64 {
	anglePrev := Algorithm_Angle_AngleBetweenPoints(p1, p0)
	angleNext := Algorithm_Angle_AngleBetweenPoints(p1, p2)
	return Algorithm_Angle_NormalizePositive(angleNext - anglePrev)
}

// Algorithm_Angle_GetTurn returns whether an angle must turn clockwise or
// counterclockwise to overlap another angle.
func Algorithm_Angle_GetTurn(ang1, ang2 float64) int {
	crossproduct := math.Sin(ang2 - ang1)

	if crossproduct > 0 {
		return Algorithm_Angle_Counterclockwise
	}
	if crossproduct < 0 {
		return Algorithm_Angle_Clockwise
	}
	return Algorithm_Angle_None
}

// Algorithm_Angle_Normalize computes the normalized value of an angle, which
// is the equivalent angle in the range ( -Pi, Pi ].
func Algorithm_Angle_Normalize(angle float64) float64 {
	for angle > math.Pi {
		angle -= Algorithm_Angle_PiTimes2
	}
	for angle <= -math.Pi {
		angle += Algorithm_Angle_PiTimes2
	}
	return angle
}

// Algorithm_Angle_NormalizePositive computes the normalized positive value of
// an angle, which is the equivalent angle in the range [ 0, 2*Pi ).
func Algorithm_Angle_NormalizePositive(angle float64) float64 {
	if angle < 0.0 {
		for angle < 0.0 {
			angle += Algorithm_Angle_PiTimes2
		}
		// In case round-off error bumps the value over.
		if angle >= Algorithm_Angle_PiTimes2 {
			angle = 0.0
		}
	} else {
		for angle >= Algorithm_Angle_PiTimes2 {
			angle -= Algorithm_Angle_PiTimes2
		}
		// In case round-off error bumps the value under.
		if angle < 0.0 {
			angle = 0.0
		}
	}
	return angle
}

// Algorithm_Angle_Diff computes the unoriented smallest difference between two
// angles. The angles are assumed to be normalized to the range [-Pi, Pi]. The
// result will be in the range [0, Pi].
func Algorithm_Angle_Diff(ang1, ang2 float64) float64 {
	var delAngle float64

	if ang1 < ang2 {
		delAngle = ang2 - ang1
	} else {
		delAngle = ang1 - ang2
	}

	if delAngle > math.Pi {
		delAngle = Algorithm_Angle_PiTimes2 - delAngle
	}

	return delAngle
}

// Algorithm_Angle_SinSnap computes sin of an angle, snapping near-zero values
// to zero.
func Algorithm_Angle_SinSnap(ang float64) float64 {
	res := math.Sin(ang)
	if math.Abs(res) < 5e-16 {
		return 0.0
	}
	return res
}

// Algorithm_Angle_CosSnap computes cos of an angle, snapping near-zero values
// to zero.
func Algorithm_Angle_CosSnap(ang float64) float64 {
	res := math.Cos(ang)
	if math.Abs(res) < 5e-16 {
		return 0.0
	}
	return res
}

// Algorithm_Angle_Project projects a point by a given angle and distance.
func Algorithm_Angle_Project(p *Geom_Coordinate, angle, dist float64) *Geom_Coordinate {
	x := p.GetX() + dist*Algorithm_Angle_CosSnap(angle)
	y := p.GetY() + dist*Algorithm_Angle_SinSnap(angle)
	return Geom_NewCoordinateWithXY(x, y)
}
