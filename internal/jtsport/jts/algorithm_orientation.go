package jts

// Functions to compute the orientation of basic geometric structures including
// point triplets (triangles) and rings. Orientation is a fundamental property
// of planar geometries (and more generally geometry on two-dimensional
// manifolds).
//
// Determining triangle orientation is notoriously subject to numerical
// precision errors in the case of collinear or nearly collinear points. JTS
// uses extended-precision arithmetic to increase the robustness of the
// computation.

// Orientation constants.
const (
	// Algorithm_Orientation_Clockwise indicates an orientation of clockwise, or a right turn.
	Algorithm_Orientation_Clockwise = -1
	// Algorithm_Orientation_Right indicates an orientation of clockwise, or a right turn.
	Algorithm_Orientation_Right = Algorithm_Orientation_Clockwise
	// Algorithm_Orientation_Counterclockwise indicates an orientation of counterclockwise, or a left turn.
	Algorithm_Orientation_Counterclockwise = 1
	// Algorithm_Orientation_Left indicates an orientation of counterclockwise, or a left turn.
	Algorithm_Orientation_Left = Algorithm_Orientation_Counterclockwise
	// Algorithm_Orientation_Collinear indicates an orientation of collinear, or no turn (straight).
	Algorithm_Orientation_Collinear = 0
	// Algorithm_Orientation_Straight indicates an orientation of collinear, or no turn (straight).
	Algorithm_Orientation_Straight = Algorithm_Orientation_Collinear
)

// Algorithm_Orientation_Index returns the orientation index of the direction of the point
// q relative to a directed infinite line specified by p1-p2. The index
// indicates whether the point lies to the LEFT or RIGHT of the line, or lies on
// it COLLINEAR. The index also indicates the orientation of the triangle formed
// by the three points (COUNTERCLOCKWISE, CLOCKWISE, or STRAIGHT).
//
// Returns:
//
//	-1 (CLOCKWISE or RIGHT) if q is clockwise (right) from p1-p2
//	 1 (COUNTERCLOCKWISE or LEFT) if q is counter-clockwise (left) from p1-p2
//	 0 (COLLINEAR or STRAIGHT) if q is collinear with p1-p2
func Algorithm_Orientation_Index(p1, p2, q *Geom_Coordinate) int {
	return Algorithm_CGAlgorithmsDD_OrientationIndex(p1, p2, q)
}

// Algorithm_Orientation_IsCCW tests if a ring defined by an array of Coordinates is
// oriented counter-clockwise.
//   - The list of points is assumed to have the first and last points equal.
//   - This handles coordinate lists which contain repeated points.
//   - This handles rings which contain collapsed segments (in particular, along
//     the top of the ring).
//
// This algorithm is guaranteed to work with valid rings. It also works with
// "mildly invalid" rings which contain collapsed (coincident) flat segments
// along the top of the ring. If the ring is "more" invalid (e.g. self-crosses
// or touches), the computed result may not be correct.
//
// Returns true if the ring is oriented counter-clockwise.
func Algorithm_Orientation_IsCCW(ring []*Geom_Coordinate) bool {
	casSeq := GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures(ring, 2, 0)
	return Algorithm_Orientation_IsCCWSeq(casSeq)
}

// Algorithm_Orientation_IsCCWSeq tests if a ring defined by a CoordinateSequence is
// oriented counter-clockwise.
//   - The list of points is assumed to have the first and last points equal.
//   - This handles coordinate lists which contain repeated points.
//   - This handles rings which contain collapsed segments (in particular, along
//     the top of the ring).
//
// This algorithm is guaranteed to work with valid rings. It also works with
// "mildly invalid" rings which contain collapsed (coincident) flat segments
// along the top of the ring. If the ring is "more" invalid (e.g. self-crosses
// or touches), the computed result may not be correct.
//
// Returns true if the ring is oriented counter-clockwise.
func Algorithm_Orientation_IsCCWSeq(ring Geom_CoordinateSequence) bool {
	nPts := ring.Size() - 1
	if nPts < 3 {
		return false
	}

	upHiPt := ring.GetCoordinate(0)
	prevY := upHiPt.GetY()
	var upLowPt *Geom_Coordinate
	iUpHi := 0
	for i := 1; i <= nPts; i++ {
		py := ring.GetOrdinate(i, Geom_Coordinate_Y)
		if py > prevY && py >= upHiPt.GetY() {
			upHiPt = ring.GetCoordinate(i)
			iUpHi = i
			upLowPt = ring.GetCoordinate(i - 1)
		}
		prevY = py
	}

	if iUpHi == 0 {
		return false
	}

	iDownLow := iUpHi
	for {
		iDownLow = (iDownLow + 1) % nPts
		if iDownLow == iUpHi || ring.GetOrdinate(iDownLow, Geom_Coordinate_Y) != upHiPt.GetY() {
			break
		}
	}

	downLowPt := ring.GetCoordinate(iDownLow)
	iDownHi := iDownLow - 1
	if iDownLow == 0 {
		iDownHi = nPts - 1
	}
	downHiPt := ring.GetCoordinate(iDownHi)

	if upHiPt.Equals2D(downHiPt) {
		if upLowPt.Equals2D(upHiPt) || downLowPt.Equals2D(upHiPt) || upLowPt.Equals2D(downLowPt) {
			return false
		}

		index := Algorithm_Orientation_Index(upLowPt, upHiPt, downLowPt)
		return index == Algorithm_Orientation_Counterclockwise
	}

	delX := downHiPt.GetX() - upHiPt.GetX()
	return delX < 0
}

// Algorithm_Orientation_IsCCWArea tests if a ring defined by an array of Coordinates is
// oriented counter-clockwise, using the signed area of the ring.
//   - The list of points is assumed to have the first and last points equal.
//   - This handles coordinate lists which contain repeated points.
//   - This handles rings which contain collapsed segments (in particular, along
//     the top of the ring).
//   - This handles rings which are invalid due to self-intersection.
//
// This algorithm is guaranteed to work with valid rings. For invalid rings
// (containing self-intersections), the algorithm determines the orientation of
// the largest enclosed area (including overlaps). This provides a more useful
// result in some situations, such as buffering.
//
// However, this approach may be less accurate in the case of rings with almost
// zero area. (Note that the orientation of rings with zero area is essentially
// undefined, and hence non-deterministic.)
//
// Returns true if the ring is oriented counter-clockwise.
func Algorithm_Orientation_IsCCWArea(ring []*Geom_Coordinate) bool {
	return Algorithm_Area_OfRingSigned(ring) < 0
}
