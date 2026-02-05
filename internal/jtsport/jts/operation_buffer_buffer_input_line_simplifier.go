package jts

import "math"

// OperationBuffer_BufferInputLineSimplifier_Simplify simplifies the input coordinate list.
// If the distance tolerance is positive,
// concavities on the LEFT side of the line are simplified.
// If the supplied distance tolerance is negative,
// concavities on the RIGHT side of the line are simplified.
func OperationBuffer_BufferInputLineSimplifier_Simplify(inputLine []*Geom_Coordinate, distanceTol float64) []*Geom_Coordinate {
	simp := operationBuffer_newBufferInputLineSimplifier(inputLine)
	return simp.simplify(distanceTol)
}

const operationBuffer_bufferInputLineSimplifier_delete = 1

// operationBuffer_BufferInputLineSimplifier simplifies a buffer input line to
// remove concavities with shallow depth.
//
// The major benefit of doing this is to reduce the number of points and the complexity of
// shape which will be buffered.
// This improve performance and robustness.
// It also reduces the risk of gores created by
// the quantized fillet arcs (although this issue
// should be eliminated by the
// offset curve generation logic).
//
// A key aspect of the simplification is that it
// affects inside (concave or inward) corners only.
// Convex (outward) corners are preserved, since they
// are required to ensure that the generated buffer curve
// lies at the correct distance from the input geometry.
//
// Another important heuristic used is that the end segments
// of linear inputs are never simplified. This ensures that
// the client buffer code is able to generate end caps faithfully.
// Ring inputs can have end segments removed by simplification.
//
// No attempt is made to avoid self-intersections in the output.
// This is acceptable for use for generating a buffer offset curve,
// since the buffer algorithm is insensitive to invalid polygonal
// geometry. However, this means that this algorithm
// cannot be used as a general-purpose polygon simplification technique.
type operationBuffer_BufferInputLineSimplifier struct {
	inputLine        []*Geom_Coordinate
	distanceTol      float64
	isRing           bool
	isDeleted        []bool
	angleOrientation int
}

// operationBuffer_newBufferInputLineSimplifier creates a new BufferInputLineSimplifier.
func operationBuffer_newBufferInputLineSimplifier(inputLine []*Geom_Coordinate) *operationBuffer_BufferInputLineSimplifier {
	return &operationBuffer_BufferInputLineSimplifier{
		inputLine:        inputLine,
		isRing:           Geom_CoordinateArrays_IsRing(inputLine),
		angleOrientation: Algorithm_Orientation_Counterclockwise,
	}
}

// simplify simplifies the input coordinate list.
// If the distance tolerance is positive,
// concavities on the LEFT side of the line are simplified.
// If the supplied distance tolerance is negative,
// concavities on the RIGHT side of the line are simplified.
func (bils *operationBuffer_BufferInputLineSimplifier) simplify(distanceTol float64) []*Geom_Coordinate {
	bils.distanceTol = math.Abs(distanceTol)
	bils.angleOrientation = Algorithm_Orientation_Counterclockwise
	if distanceTol < 0 {
		bils.angleOrientation = Algorithm_Orientation_Clockwise
	}

	// rely on fact that boolean array is filled with false values
	bils.isDeleted = make([]bool, len(bils.inputLine))

	isChanged := false
	for {
		isChanged = bils.deleteShallowConcavities()
		if !isChanged {
			break
		}
	}

	return bils.collapseLine()
}

// deleteShallowConcavities uses a sliding window containing 3 vertices to detect shallow angles
// in which the middle vertex can be deleted, since it does not
// affect the shape of the resulting buffer in a significant way.
//
// Returns true if any vertices were deleted.
func (bils *operationBuffer_BufferInputLineSimplifier) deleteShallowConcavities() bool {
	// Do not simplify end line segments of lines.
	// This ensures that end caps are generated consistently.
	index := 0
	if !bils.isRing {
		index = 1
	}

	midIndex := bils.nextIndex(index)
	lastIndex := bils.nextIndex(midIndex)

	isChanged := false
	for lastIndex < len(bils.inputLine) {
		// test triple for shallow concavity
		isMiddleVertexDeleted := false
		if bils.isDeletable(index, midIndex, lastIndex, bils.distanceTol) {
			bils.isDeleted[midIndex] = true
			isMiddleVertexDeleted = true
			isChanged = true
		}
		// move simplification window forward
		if isMiddleVertexDeleted {
			index = lastIndex
		} else {
			index = midIndex
		}

		midIndex = bils.nextIndex(index)
		lastIndex = bils.nextIndex(midIndex)
	}
	return isChanged
}

// nextIndex finds the next non-deleted index, or the end of the point array if none.
func (bils *operationBuffer_BufferInputLineSimplifier) nextIndex(index int) int {
	next := index + 1
	for next < len(bils.inputLine) && bils.isDeleted[next] {
		next++
	}
	return next
}

func (bils *operationBuffer_BufferInputLineSimplifier) collapseLine() []*Geom_Coordinate {
	coordList := Geom_NewCoordinateList()
	for i := 0; i < len(bils.inputLine); i++ {
		if !bils.isDeleted[i] {
			coordList.AddCoordinate(bils.inputLine[i], true)
		}
	}
	return coordList.ToCoordinateArray()
}

func (bils *operationBuffer_BufferInputLineSimplifier) isDeletable(i0, i1, i2 int, distanceTol float64) bool {
	p0 := bils.inputLine[i0]
	p1 := bils.inputLine[i1]
	p2 := bils.inputLine[i2]

	if !bils.isConcave(p0, p1, p2) {
		return false
	}
	if !operationBuffer_bufferInputLineSimplifier_isShallow(p0, p1, p2, distanceTol) {
		return false
	}

	return bils.isShallowSampled(p0, p2, i0, i2, distanceTol)
}

const operationBuffer_bufferInputLineSimplifier_num_pts_to_check = 10

// isShallowSampled checks for shallowness over a sample of points in the given section.
// This helps prevents the simplification from incrementally
// "skipping" over points which are in fact non-shallow.
func (bils *operationBuffer_BufferInputLineSimplifier) isShallowSampled(p0, p2 *Geom_Coordinate, i0, i2 int, distanceTol float64) bool {
	// check every n'th point to see if it is within tolerance
	inc := (i2 - i0) / operationBuffer_bufferInputLineSimplifier_num_pts_to_check
	if inc <= 0 {
		inc = 1
	}

	for i := i0; i < i2; i += inc {
		if !operationBuffer_bufferInputLineSimplifier_isShallow(p0, bils.inputLine[i], p2, distanceTol) {
			return false
		}
	}
	return true
}

func operationBuffer_bufferInputLineSimplifier_isShallow(p0, p1, p2 *Geom_Coordinate, distanceTol float64) bool {
	dist := Algorithm_Distance_PointToSegment(p1, p0, p2)
	return dist < distanceTol
}

func (bils *operationBuffer_BufferInputLineSimplifier) isConcave(p0, p1, p2 *Geom_Coordinate) bool {
	orientation := Algorithm_Orientation_Index(p0, p1, p2)
	isConcave := orientation == bils.angleOrientation
	return isConcave
}
