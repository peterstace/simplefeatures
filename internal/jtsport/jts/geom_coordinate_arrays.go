package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geom_CoordinateArrays provides useful utility functions for handling Geom_Coordinate
// arrays.

// Geom_CoordinateArrays_Dimension determines the dimension based on the subclass of Geom_Coordinate.
func Geom_CoordinateArrays_Dimension(pts []*Geom_Coordinate) int {
	if pts == nil || len(pts) == 0 {
		return 3 // unknown, assume default
	}
	dimension := 0
	for _, coordinate := range pts {
		d := Geom_Coordinates_Dimension(coordinate)
		if d > dimension {
			dimension = d
		}
	}
	return dimension
}

// Geom_CoordinateArrays_Measures determines the number of measures based on the subclass of
// Geom_Coordinate.
func Geom_CoordinateArrays_Measures(pts []*Geom_Coordinate) int {
	if pts == nil || len(pts) == 0 {
		return 0 // unknown, assume default
	}
	measures := 0
	for _, coordinate := range pts {
		m := Geom_Coordinates_Measures(coordinate)
		if m > measures {
			measures = m
		}
	}
	return measures
}

// Geom_CoordinateArrays_EnforceConsistency ensures array contents are of consistent dimension and
// measures. Array is modified in place if required, coordinates are replaced in
// the array as required to ensure all coordinates have the same dimension and
// measures. The final dimension and measures used are the maximum found when
// checking the array.
func Geom_CoordinateArrays_EnforceConsistency(array []*Geom_Coordinate) {
	if array == nil {
		return
	}
	// Step one: check.
	maxDimension := -1
	maxMeasures := -1
	isConsistent := true
	for i := 0; i < len(array); i++ {
		coordinate := array[i]
		if coordinate != nil {
			d := Geom_Coordinates_Dimension(coordinate)
			m := Geom_Coordinates_Measures(coordinate)
			if maxDimension == -1 {
				maxDimension = d
				maxMeasures = m
				continue
			}
			if d != maxDimension || m != maxMeasures {
				isConsistent = false
				if d > maxDimension {
					maxDimension = d
				}
				if m > maxMeasures {
					maxMeasures = m
				}
			}
		}
	}
	if !isConsistent {
		// Step two: fix.
		sample := Geom_Coordinates_CreateWithMeasures(maxDimension, maxMeasures)
		sampleType := geom_CoordinateArrays_coordinateType(sample)

		for i := 0; i < len(array); i++ {
			coordinate := array[i]
			if coordinate != nil && geom_CoordinateArrays_coordinateType(coordinate) != sampleType {
				duplicate := Geom_Coordinates_CreateWithMeasures(maxDimension, maxMeasures)
				duplicate.SetCoordinate(coordinate)
				array[i] = duplicate
			}
		}
	}
}

// Geom_CoordinateArrays_EnforceConsistencyWithDimension ensures array contents are of the specified
// dimension and measures. Array is returned unmodified if consistent, or a copy
// of the array is made with each inconsistent coordinate duplicated into an
// instance of the correct dimension and measures.
func Geom_CoordinateArrays_EnforceConsistencyWithDimension(array []*Geom_Coordinate, dimension, measures int) []*Geom_Coordinate {
	sample := Geom_Coordinates_CreateWithMeasures(dimension, measures)
	sampleType := geom_CoordinateArrays_coordinateType(sample)
	isConsistent := true
	for i := 0; i < len(array); i++ {
		coordinate := array[i]
		if coordinate != nil && geom_CoordinateArrays_coordinateType(coordinate) != sampleType {
			isConsistent = false
			break
		}
	}
	if isConsistent {
		return array
	}
	copyArr := make([]*Geom_Coordinate, len(array))
	for i := 0; i < len(copyArr); i++ {
		coordinate := array[i]
		if coordinate != nil && geom_CoordinateArrays_coordinateType(coordinate) != sampleType {
			duplicate := Geom_Coordinates_CreateWithMeasures(dimension, measures)
			duplicate.SetCoordinate(coordinate)
			copyArr[i] = duplicate
		} else {
			copyArr[i] = coordinate
		}
	}
	return copyArr
}

// geom_CoordinateArrays_coordinateType returns a string representing the type of the coordinate.
func geom_CoordinateArrays_coordinateType(c *Geom_Coordinate) string {
	self := java.GetLeaf(c)
	if _, ok := self.(*Geom_CoordinateXY); ok {
		return "XY"
	} else if _, ok := self.(*Geom_CoordinateXYM); ok {
		return "XYM"
	} else if _, ok := self.(*Geom_CoordinateXYZM); ok {
		return "XYZM"
	}
	return "XYZ"
}

// Geom_CoordinateArrays_IsRing tests whether an array of Geom_Coordinates forms a ring, by checking length
// and closure. Self-intersection is not checked.
func Geom_CoordinateArrays_IsRing(pts []*Geom_Coordinate) bool {
	if len(pts) < 4 {
		return false
	}
	if !pts[0].Equals2D(pts[len(pts)-1]) {
		return false
	}
	return true
}

// Geom_CoordinateArrays_PtNotInList finds a point in a list of points which is not contained in
// another list of points. Returns a Geom_Coordinate from testPts which is not in
// pts, or nil.
func Geom_CoordinateArrays_PtNotInList(testPts, pts []*Geom_Coordinate) *Geom_Coordinate {
	for i := 0; i < len(testPts); i++ {
		testPt := testPts[i]
		if Geom_CoordinateArrays_IndexOf(testPt, pts) < 0 {
			return testPt
		}
	}
	return nil
}

// Geom_CoordinateArrays_Compare compares two Geom_Coordinate arrays in the forward direction of their
// coordinates, using lexicographic ordering.
func Geom_CoordinateArrays_Compare(pts1, pts2 []*Geom_Coordinate) int {
	i := 0
	for i < len(pts1) && i < len(pts2) {
		compare := pts1[i].CompareTo(pts2[i])
		if compare != 0 {
			return compare
		}
		i++
	}
	// Handle situation when arrays are of different length.
	if i < len(pts2) {
		return -1
	}
	if i < len(pts1) {
		return 1
	}
	return 0
}

// Geom_ForwardComparator is a Comparator for Geom_Coordinate arrays in the forward
// direction of their coordinates, using lexicographic ordering.
type Geom_ForwardComparator struct{}

// Compare compares two Geom_Coordinate arrays.
func (fc *Geom_ForwardComparator) Compare(pts1, pts2 []*Geom_Coordinate) int {
	return Geom_CoordinateArrays_Compare(pts1, pts2)
}

// Geom_CoordinateArrays_IncreasingDirection determines which orientation of the Geom_Coordinate array is
// (overall) increasing. In other words, determines which end of the array is
// "smaller" (using the standard ordering on Geom_Coordinate). Returns an integer
// indicating the increasing direction. If the sequence is a palindrome, it is
// defined to be oriented in a positive direction.
//
// Returns 1 if the array is smaller at the start or is a palindrome, -1 if
// smaller at the end.
func Geom_CoordinateArrays_IncreasingDirection(pts []*Geom_Coordinate) int {
	for i := 0; i < len(pts)/2; i++ {
		j := len(pts) - 1 - i
		// Skip equal points on both ends.
		comp := pts[i].CompareTo(pts[j])
		if comp != 0 {
			return comp
		}
	}
	// Array must be a palindrome - defined to be in positive direction.
	return 1
}

// geom_CoordinateArrays_isEqualReversed determines whether two Geom_Coordinate arrays of equal length are
// equal in opposite directions.
func geom_CoordinateArrays_isEqualReversed(pts1, pts2 []*Geom_Coordinate) bool {
	for i := 0; i < len(pts1); i++ {
		p1 := pts1[i]
		p2 := pts2[len(pts1)-i-1]
		if p1.CompareTo(p2) != 0 {
			return false
		}
	}
	return true
}

// Geom_BidirectionalComparator is a Comparator for Geom_Coordinate arrays modulo their
// directionality. E.g. if two coordinate arrays are identical but reversed they
// will compare as equal under this ordering. If the arrays are not equal, the
// ordering returned is the ordering in the forward direction.
type Geom_BidirectionalComparator struct{}

// Compare compares two Geom_Coordinate arrays.
func (bc *Geom_BidirectionalComparator) Compare(pts1, pts2 []*Geom_Coordinate) int {
	if len(pts1) < len(pts2) {
		return -1
	}
	if len(pts1) > len(pts2) {
		return 1
	}
	if len(pts1) == 0 {
		return 0
	}
	forwardComp := Geom_CoordinateArrays_Compare(pts1, pts2)
	isEqualRev := geom_CoordinateArrays_isEqualReversed(pts1, pts2)
	if isEqualRev {
		return 0
	}
	return forwardComp
}

// Geom_CoordinateArrays_CopyDeep creates a deep copy of the argument Geom_Coordinate array.
func Geom_CoordinateArrays_CopyDeep(coordinates []*Geom_Coordinate) []*Geom_Coordinate {
	copyArr := make([]*Geom_Coordinate, len(coordinates))
	for i := 0; i < len(coordinates); i++ {
		copyArr[i] = coordinates[i].Copy()
	}
	return copyArr
}

// Geom_CoordinateArrays_CopyDeepRange creates a deep copy of a given section of a source Geom_Coordinate
// array into a destination Geom_Coordinate array. The destination array must be an
// appropriate size to receive the copied coordinates.
func Geom_CoordinateArrays_CopyDeepRange(src []*Geom_Coordinate, srcStart int, dest []*Geom_Coordinate, destStart, length int) {
	for i := 0; i < length; i++ {
		dest[destStart+i] = src[srcStart+i].Copy()
	}
}

// Geom_CoordinateArrays_ToCoordinateArray converts the given slice of Geom_Coordinates into a Geom_Coordinate
// array. This is essentially a no-op in Go since slices are already the native
// collection type.
func Geom_CoordinateArrays_ToCoordinateArray(coordList []*Geom_Coordinate) []*Geom_Coordinate {
	result := make([]*Geom_Coordinate, len(coordList))
	copy(result, coordList)
	return result
}

// Geom_CoordinateArrays_HasRepeatedPoints tests whether Geom_Coordinate.Equals returns true for any two
// consecutive Geom_Coordinates in the given array.
func Geom_CoordinateArrays_HasRepeatedPoints(coord []*Geom_Coordinate) bool {
	for i := 1; i < len(coord); i++ {
		if coord[i-1].Equals(coord[i]) {
			return true
		}
	}
	return false
}

// Geom_CoordinateArrays_AtLeastNCoordinatesOrNothing returns either the given coordinate array if its
// length is greater than or equal to the given amount, or an empty coordinate
// array.
func Geom_CoordinateArrays_AtLeastNCoordinatesOrNothing(n int, c []*Geom_Coordinate) []*Geom_Coordinate {
	if len(c) >= n {
		return c
	}
	return []*Geom_Coordinate{}
}

// Geom_CoordinateArrays_RemoveRepeatedPoints constructs a new array containing no repeated points if
// the coordinate array argument has repeated points. Otherwise, returns the
// argument.
func Geom_CoordinateArrays_RemoveRepeatedPoints(coord []*Geom_Coordinate) []*Geom_Coordinate {
	if !Geom_CoordinateArrays_HasRepeatedPoints(coord) {
		return coord
	}
	coordList := Geom_NewCoordinateListFromCoordinatesAllowRepeated(coord, false)
	return coordList.ToCoordinateArray()
}

// Geom_CoordinateArrays_HasRepeatedOrInvalidPoints tests whether an array has any repeated or invalid
// coordinates.
func Geom_CoordinateArrays_HasRepeatedOrInvalidPoints(coord []*Geom_Coordinate) bool {
	for i := 1; i < len(coord); i++ {
		if !coord[i].IsValid() {
			return true
		}
		if coord[i-1].Equals(coord[i]) {
			return true
		}
	}
	return false
}

// Geom_CoordinateArrays_RemoveRepeatedOrInvalidPoints constructs a new array containing no repeated
// or invalid points if the coordinate array argument has repeated or invalid
// points. Otherwise, returns the argument.
func Geom_CoordinateArrays_RemoveRepeatedOrInvalidPoints(coord []*Geom_Coordinate) []*Geom_Coordinate {
	if !Geom_CoordinateArrays_HasRepeatedOrInvalidPoints(coord) {
		return coord
	}
	coordList := Geom_NewCoordinateList()
	for i := 0; i < len(coord); i++ {
		if !coord[i].IsValid() {
			continue
		}
		coordList.AddCoordinate(coord[i], false)
	}
	return coordList.ToCoordinateArray()
}

// Geom_CoordinateArrays_RemoveNull collapses a coordinate array to remove all nil elements.
func Geom_CoordinateArrays_RemoveNull(coord []*Geom_Coordinate) []*Geom_Coordinate {
	nonNull := 0
	for i := 0; i < len(coord); i++ {
		if coord[i] != nil {
			nonNull++
		}
	}
	newCoord := make([]*Geom_Coordinate, nonNull)
	// Empty case.
	if nonNull == 0 {
		return newCoord
	}
	j := 0
	for i := 0; i < len(coord); i++ {
		if coord[i] != nil {
			newCoord[j] = coord[i]
			j++
		}
	}
	return newCoord
}

// Geom_CoordinateArrays_Reverse reverses the coordinates in an array in-place.
func Geom_CoordinateArrays_Reverse(coord []*Geom_Coordinate) {
	if len(coord) <= 1 {
		return
	}
	last := len(coord) - 1
	mid := last / 2
	for i := 0; i <= mid; i++ {
		tmp := coord[i]
		coord[i] = coord[last-i]
		coord[last-i] = tmp
	}
}

// Geom_CoordinateArrays_Equals returns true if the two arrays are identical, both nil, or pointwise
// equal (as compared using Geom_Coordinate.Equals).
func Geom_CoordinateArrays_Equals(coord1, coord2 []*Geom_Coordinate) bool {
	if len(coord1) == 0 && len(coord2) == 0 {
		return true
	}
	if coord1 == nil && coord2 == nil {
		return true
	}
	if coord1 == nil || coord2 == nil {
		return false
	}
	if len(coord1) != len(coord2) {
		return false
	}
	for i := 0; i < len(coord1); i++ {
		if !coord1[i].Equals(coord2[i]) {
			return false
		}
	}
	return true
}

// Geom_CoordinateComparator is an interface for comparing Geom_Coordinates.
type Geom_CoordinateComparator interface {
	Compare(c1, c2 *Geom_Coordinate) int
}

// Geom_CoordinateArrays_EqualsWithComparator returns true if the two arrays are identical, both nil,
// or pointwise equal, using a user-defined Comparator for Geom_Coordinates.
func Geom_CoordinateArrays_EqualsWithComparator(coord1, coord2 []*Geom_Coordinate, coordinateComparator Geom_CoordinateComparator) bool {
	if len(coord1) == 0 && len(coord2) == 0 {
		return true
	}
	if coord1 == nil && coord2 == nil {
		return true
	}
	if coord1 == nil || coord2 == nil {
		return false
	}
	if len(coord1) != len(coord2) {
		return false
	}
	for i := 0; i < len(coord1); i++ {
		if coordinateComparator.Compare(coord1[i], coord2[i]) != 0 {
			return false
		}
	}
	return true
}

// Geom_CoordinateArrays_MinCoordinate returns the minimum coordinate, using the usual lexicographic
// comparison.
func Geom_CoordinateArrays_MinCoordinate(coordinates []*Geom_Coordinate) *Geom_Coordinate {
	var minCoord *Geom_Coordinate
	for i := 0; i < len(coordinates); i++ {
		if minCoord == nil || minCoord.CompareTo(coordinates[i]) > 0 {
			minCoord = coordinates[i]
		}
	}
	return minCoord
}

// Geom_CoordinateArrays_Scroll shifts the positions of the coordinates until firstCoordinate is
// first.
func Geom_CoordinateArrays_Scroll(coordinates []*Geom_Coordinate, firstCoordinate *Geom_Coordinate) {
	i := Geom_CoordinateArrays_IndexOf(firstCoordinate, coordinates)
	Geom_CoordinateArrays_ScrollToIndex(coordinates, i)
}

// Geom_CoordinateArrays_ScrollToIndex shifts the positions of the coordinates until the coordinate at
// indexOfFirstCoordinate is first.
func Geom_CoordinateArrays_ScrollToIndex(coordinates []*Geom_Coordinate, indexOfFirstCoordinate int) {
	Geom_CoordinateArrays_ScrollToIndexWithRing(coordinates, indexOfFirstCoordinate, Geom_CoordinateArrays_IsRing(coordinates))
}

// Geom_CoordinateArrays_ScrollToIndexWithRing shifts the positions of the coordinates until the
// coordinate at indexOfFirstCoordinate is first. If ensureRing is true, first
// and last coordinate of the returned array are equal.
func Geom_CoordinateArrays_ScrollToIndexWithRing(coordinates []*Geom_Coordinate, indexOfFirstCoordinate int, ensureRing bool) {
	i := indexOfFirstCoordinate
	if i <= 0 {
		return
	}
	newCoordinates := make([]*Geom_Coordinate, len(coordinates))
	if !ensureRing {
		copy(newCoordinates[0:], coordinates[i:])
		copy(newCoordinates[len(coordinates)-i:], coordinates[0:i])
	} else {
		last := len(coordinates) - 1
		// Fill in values.
		j := 0
		for ; j < last; j++ {
			newCoordinates[j] = coordinates[(i+j)%last]
		}
		// Fix the ring (first == last).
		newCoordinates[j] = newCoordinates[0].Copy()
	}
	copy(coordinates, newCoordinates)
}

// Geom_CoordinateArrays_IndexOf returns the index of coordinate in coordinates. The first position is
// 0; the second, 1; etc. Returns -1 if not found.
func Geom_CoordinateArrays_IndexOf(coordinate *Geom_Coordinate, coordinates []*Geom_Coordinate) int {
	for i := 0; i < len(coordinates); i++ {
		if coordinate.Equals(coordinates[i]) {
			return i
		}
	}
	return -1
}

// Geom_CoordinateArrays_Extract extracts a subsequence of the input Geom_Coordinate array from indices
// start to end (inclusive). The input indices are clamped to the array size; If
// the end index is less than the start index, the extracted array will be
// empty.
func Geom_CoordinateArrays_Extract(pts []*Geom_Coordinate, start, end int) []*Geom_Coordinate {
	start = Math_MathUtil_ClampInt(start, 0, len(pts))
	end = Math_MathUtil_ClampInt(end, -1, len(pts))

	npts := end - start + 1
	if end < 0 {
		npts = 0
	}
	if start >= len(pts) {
		npts = 0
	}
	if end < start {
		npts = 0
	}

	extractPts := make([]*Geom_Coordinate, npts)
	if npts == 0 {
		return extractPts
	}

	iPts := 0
	for i := start; i <= end; i++ {
		extractPts[iPts] = pts[i]
		iPts++
	}
	return extractPts
}

// Geom_CoordinateArrays_Envelope computes the envelope of the coordinates.
func Geom_CoordinateArrays_Envelope(coordinates []*Geom_Coordinate) *Geom_Envelope {
	env := Geom_NewEnvelope()
	for i := 0; i < len(coordinates); i++ {
		env.ExpandToIncludeCoordinate(coordinates[i])
	}
	return env
}

// Geom_CoordinateArrays_Intersection extracts the coordinates which intersect a Geom_Envelope.
func Geom_CoordinateArrays_Intersection(coordinates []*Geom_Coordinate, env *Geom_Envelope) []*Geom_Coordinate {
	coordList := Geom_NewCoordinateList()
	for i := 0; i < len(coordinates); i++ {
		if env.IntersectsCoordinate(coordinates[i]) {
			coordList.AddCoordinate(coordinates[i], true)
		}
	}
	return coordList.ToCoordinateArray()
}
