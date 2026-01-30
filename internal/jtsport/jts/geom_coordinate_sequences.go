package jts

import (
	"math"
	"strconv"
	"strings"
)

// Geom_CoordinateSequences provides utility functions for manipulating
// Geom_CoordinateSequence objects.

// Geom_CoordinateSequences_Reverse reverses the coordinates in a sequence in-place.
func Geom_CoordinateSequences_Reverse(seq Geom_CoordinateSequence) {
	if seq.Size() <= 1 {
		return
	}
	last := seq.Size() - 1
	mid := last / 2
	for i := 0; i <= mid; i++ {
		Geom_CoordinateSequences_Swap(seq, i, last-i)
	}
}

// Geom_CoordinateSequences_Swap swaps two coordinates in a sequence.
func Geom_CoordinateSequences_Swap(seq Geom_CoordinateSequence, i, j int) {
	if i == j {
		return
	}
	for dim := 0; dim < seq.GetDimension(); dim++ {
		tmp := seq.GetOrdinate(i, dim)
		seq.SetOrdinate(i, dim, seq.GetOrdinate(j, dim))
		seq.SetOrdinate(j, dim, tmp)
	}
}

// Geom_CoordinateSequences_Copy copies a section of a Geom_CoordinateSequence to another Geom_CoordinateSequence.
// The sequences may have different dimensions; in this case only the common
// dimensions are copied.
func Geom_CoordinateSequences_Copy(src Geom_CoordinateSequence, srcPos int, dest Geom_CoordinateSequence, destPos int, length int) {
	for i := 0; i < length; i++ {
		Geom_CoordinateSequences_CopyCoord(src, srcPos+i, dest, destPos+i)
	}
}

// Geom_CoordinateSequences_CopyCoord copies a coordinate of a Geom_CoordinateSequence to another
// Geom_CoordinateSequence. The sequences may have different dimensions; in this case
// only the common dimensions are copied.
func Geom_CoordinateSequences_CopyCoord(src Geom_CoordinateSequence, srcPos int, dest Geom_CoordinateSequence, destPos int) {
	minDim := src.GetDimension()
	if dest.GetDimension() < minDim {
		minDim = dest.GetDimension()
	}
	for dim := 0; dim < minDim; dim++ {
		dest.SetOrdinate(destPos, dim, src.GetOrdinate(srcPos, dim))
	}
}

// Geom_CoordinateSequences_IsRing tests whether a Geom_CoordinateSequence forms a valid LinearRing, by
// checking the sequence length and closure (whether the first and last points
// are identical in 2D). Self-intersection is not checked.
func Geom_CoordinateSequences_IsRing(seq Geom_CoordinateSequence) bool {
	n := seq.Size()
	if n == 0 {
		return true
	}
	if n <= 3 {
		return false
	}
	return seq.GetOrdinate(0, Geom_CoordinateSequence_X) == seq.GetOrdinate(n-1, Geom_CoordinateSequence_X) &&
		seq.GetOrdinate(0, Geom_CoordinateSequence_Y) == seq.GetOrdinate(n-1, Geom_CoordinateSequence_Y)
}

// Geom_CoordinateSequences_EnsureValidRing ensures that a Geom_CoordinateSequence forms a valid ring,
// returning a new closed sequence of the correct length if required. If the
// input sequence is already a valid ring, it is returned without modification.
// If the input sequence is too short or is not closed, it is extended with one
// or more copies of the start point.
func Geom_CoordinateSequences_EnsureValidRing(fact Geom_CoordinateSequenceFactory, seq Geom_CoordinateSequence) Geom_CoordinateSequence {
	n := seq.Size()
	if n == 0 {
		return seq
	}
	if n <= 3 {
		return geom_CoordinateSequences_createClosedRing(fact, seq, 4)
	}
	isClosed := seq.GetOrdinate(0, Geom_CoordinateSequence_X) == seq.GetOrdinate(n-1, Geom_CoordinateSequence_X) &&
		seq.GetOrdinate(0, Geom_CoordinateSequence_Y) == seq.GetOrdinate(n-1, Geom_CoordinateSequence_Y)
	if isClosed {
		return seq
	}
	return geom_CoordinateSequences_createClosedRing(fact, seq, n+1)
}

func geom_CoordinateSequences_createClosedRing(fact Geom_CoordinateSequenceFactory, seq Geom_CoordinateSequence, size int) Geom_CoordinateSequence {
	newseq := fact.CreateWithSizeAndDimension(size, seq.GetDimension())
	n := seq.Size()
	Geom_CoordinateSequences_Copy(seq, 0, newseq, 0, n)
	for i := n; i < size; i++ {
		Geom_CoordinateSequences_Copy(seq, 0, newseq, i, 1)
	}
	return newseq
}

// Geom_CoordinateSequences_Extend extends a Geom_CoordinateSequence to the specified size. If the sequence is
// already at or greater than the specified size, it is returned unchanged. The
// new coordinates are filled with copies of the end point.
func Geom_CoordinateSequences_Extend(fact Geom_CoordinateSequenceFactory, seq Geom_CoordinateSequence, size int) Geom_CoordinateSequence {
	newseq := fact.CreateWithSizeAndDimension(size, seq.GetDimension())
	n := seq.Size()
	Geom_CoordinateSequences_Copy(seq, 0, newseq, 0, n)
	if n > 0 {
		for i := n; i < size; i++ {
			Geom_CoordinateSequences_Copy(seq, n-1, newseq, i, 1)
		}
	}
	return newseq
}

// Geom_CoordinateSequences_IsEqual tests whether two Geom_CoordinateSequences are equal. To be equal, the
// sequences must be the same length. They do not need to be of the same
// dimension, but the ordinate values for the smallest dimension of the two must
// be equal. Two NaN ordinate values are considered to be equal.
func Geom_CoordinateSequences_IsEqual(cs1, cs2 Geom_CoordinateSequence) bool {
	cs1Size := cs1.Size()
	cs2Size := cs2.Size()
	if cs1Size != cs2Size {
		return false
	}
	dim := cs1.GetDimension()
	if cs2.GetDimension() < dim {
		dim = cs2.GetDimension()
	}
	for i := 0; i < cs1Size; i++ {
		for d := 0; d < dim; d++ {
			v1 := cs1.GetOrdinate(i, d)
			v2 := cs2.GetOrdinate(i, d)
			if v1 == v2 {
				continue
			} else if math.IsNaN(v1) && math.IsNaN(v2) {
				continue
			} else {
				return false
			}
		}
	}
	return true
}

// Geom_CoordinateSequences_ToString creates a string representation of a Geom_CoordinateSequence. The format
// is: ( ord0,ord1.. ord0,ord1,... ... )
func Geom_CoordinateSequences_ToString(seq Geom_CoordinateSequence) string {
	size := seq.Size()
	if size == 0 {
		return "()"
	}
	dim := seq.GetDimension()
	var builder strings.Builder
	builder.WriteByte('(')
	for i := 0; i < size; i++ {
		if i > 0 {
			builder.WriteByte(' ')
		}
		for d := 0; d < dim; d++ {
			if d > 0 {
				builder.WriteByte(',')
			}
			builder.WriteString(geom_formatOrdinate(seq.GetOrdinate(i, d)))
		}
	}
	builder.WriteByte(')')
	return builder.String()
}

// geom_formatOrdinate formats an ordinate value as a string.
func geom_formatOrdinate(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// Geom_CoordinateSequences_MinCoordinate returns the minimum coordinate, using the usual lexicographic
// comparison.
func Geom_CoordinateSequences_MinCoordinate(seq Geom_CoordinateSequence) *Geom_Coordinate {
	var minCoord *Geom_Coordinate
	for i := 0; i < seq.Size(); i++ {
		testCoord := seq.GetCoordinate(i)
		if minCoord == nil || minCoord.CompareTo(testCoord) > 0 {
			minCoord = testCoord
		}
	}
	return minCoord
}

// Geom_CoordinateSequences_MinCoordinateIndex returns the index of the minimum coordinate of the whole
// coordinate sequence, using the usual lexicographic comparison.
func Geom_CoordinateSequences_MinCoordinateIndex(seq Geom_CoordinateSequence) int {
	return Geom_CoordinateSequences_MinCoordinateIndexInRange(seq, 0, seq.Size()-1)
}

// Geom_CoordinateSequences_MinCoordinateIndexInRange returns the index of the minimum coordinate of a
// part of the coordinate sequence (defined by from and to), using the usual
// lexicographic comparison.
func Geom_CoordinateSequences_MinCoordinateIndexInRange(seq Geom_CoordinateSequence, from, to int) int {
	minCoordIndex := -1
	var minCoord *Geom_Coordinate
	for i := from; i <= to; i++ {
		testCoord := seq.GetCoordinate(i)
		if minCoord == nil || minCoord.CompareTo(testCoord) > 0 {
			minCoord = testCoord
			minCoordIndex = i
		}
	}
	return minCoordIndex
}

// Geom_CoordinateSequences_ScrollToCoordinate shifts the positions of the coordinates until
// firstCoordinate is first.
func Geom_CoordinateSequences_ScrollToCoordinate(seq Geom_CoordinateSequence, firstCoordinate *Geom_Coordinate) {
	i := Geom_CoordinateSequences_IndexOf(firstCoordinate, seq)
	if i <= 0 {
		return
	}
	Geom_CoordinateSequences_ScrollToIndex(seq, i)
}

// Geom_CoordinateSequences_ScrollToIndex shifts the positions of the coordinates until the coordinate at
// indexOfFirstCoordinate is first.
func Geom_CoordinateSequences_ScrollToIndex(seq Geom_CoordinateSequence, indexOfFirstCoordinate int) {
	Geom_CoordinateSequences_ScrollToIndexWithRing(seq, indexOfFirstCoordinate, Geom_CoordinateSequences_IsRing(seq))
}

// Geom_CoordinateSequences_ScrollToIndexWithRing shifts the positions of the coordinates until the
// coordinate at indexOfFirstCoordinate is first.
func Geom_CoordinateSequences_ScrollToIndexWithRing(seq Geom_CoordinateSequence, indexOfFirstCoordinate int, ensureRing bool) {
	i := indexOfFirstCoordinate
	if i <= 0 {
		return
	}

	seqCopy := seq.Copy()

	last := seq.Size()
	if ensureRing {
		last = seq.Size() - 1
	}

	for j := 0; j < last; j++ {
		for k := 0; k < seq.GetDimension(); k++ {
			seq.SetOrdinate(j, k, seqCopy.GetOrdinate((indexOfFirstCoordinate+j)%last, k))
		}
	}

	if ensureRing {
		for k := 0; k < seq.GetDimension(); k++ {
			seq.SetOrdinate(last, k, seq.GetOrdinate(0, k))
		}
	}
}

// Geom_CoordinateSequences_IndexOf returns the index of coordinate in a Geom_CoordinateSequence. The first
// position is 0; the second, 1; etc. Returns -1 if not found.
func Geom_CoordinateSequences_IndexOf(coordinate *Geom_Coordinate, seq Geom_CoordinateSequence) int {
	for i := 0; i < seq.Size(); i++ {
		if coordinate.X == seq.GetOrdinate(i, Geom_CoordinateSequence_X) &&
			coordinate.Y == seq.GetOrdinate(i, Geom_CoordinateSequence_Y) {
			return i
		}
	}
	return -1
}
