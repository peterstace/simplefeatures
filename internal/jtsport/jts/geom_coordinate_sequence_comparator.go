package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geom_CoordinateSequenceComparator_Compare compares two float64 values, allowing for NaN values. NaN is treated
// as being less than any valid number.
//
// Returns -1, 0, or 1 depending on whether a is less than, equal to or greater
// than b.
func Geom_CoordinateSequenceComparator_Compare(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	if math.IsNaN(a) {
		if math.IsNaN(b) {
			return 0
		}
		return -1
	}
	if math.IsNaN(b) {
		return 1
	}
	return 0
}

// Geom_CoordinateSequenceComparator compares two Geom_CoordinateSequences. For sequences
// of the same dimension, the ordering is lexicographic. Otherwise, lower
// dimensions are sorted before higher. The dimensions compared can be limited;
// if this is done ordinate dimensions above the limit will not be compared.
//
// If different behaviour is required for comparing size, dimension, or
// coordinate values, any or all methods can be overridden.
type Geom_CoordinateSequenceComparator struct {
	child java.Polymorphic
	// dimensionLimit is the number of dimensions to test.
	dimensionLimit int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (csc *Geom_CoordinateSequenceComparator) GetChild() java.Polymorphic {
	return csc.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (csc *Geom_CoordinateSequenceComparator) GetParent() java.Polymorphic {
	return nil
}

// Geom_NewCoordinateSequenceComparator creates a comparator which will test all
// dimensions.
func Geom_NewCoordinateSequenceComparator() *Geom_CoordinateSequenceComparator {
	return &Geom_CoordinateSequenceComparator{
		dimensionLimit: math.MaxInt,
	}
}

// Geom_NewCoordinateSequenceComparatorWithDimensionLimit creates a comparator which
// will test only the specified number of dimensions.
func Geom_NewCoordinateSequenceComparatorWithDimensionLimit(dimensionLimit int) *Geom_CoordinateSequenceComparator {
	return &Geom_CoordinateSequenceComparator{
		dimensionLimit: dimensionLimit,
	}
}

// Compare compares two Geom_CoordinateSequences for relative order.
//
// Returns -1, 0, or 1 depending on whether s1 is less than, equal to, or
// greater than s2.
func (c *Geom_CoordinateSequenceComparator) Compare(s1, s2 Geom_CoordinateSequence) int {
	if impl, ok := java.GetLeaf(c).(interface {
		Compare_BODY(Geom_CoordinateSequence, Geom_CoordinateSequence) int
	}); ok {
		return impl.Compare_BODY(s1, s2)
	}
	return c.Compare_BODY(s1, s2)
}

func (c *Geom_CoordinateSequenceComparator) Compare_BODY(s1, s2 Geom_CoordinateSequence) int {
	size1 := s1.Size()
	size2 := s2.Size()

	dim1 := s1.GetDimension()
	dim2 := s2.GetDimension()

	minDim := dim1
	if dim2 < minDim {
		minDim = dim2
	}
	dimLimited := false
	if c.dimensionLimit <= minDim {
		minDim = c.dimensionLimit
		dimLimited = true
	}

	// Lower dimension is less than higher.
	if !dimLimited {
		if dim1 < dim2 {
			return -1
		}
		if dim1 > dim2 {
			return 1
		}
	}

	// Lexicographic ordering of point sequences.
	i := 0
	for i < size1 && i < size2 {
		ptComp := c.CompareCoordinate(s1, s2, i, minDim)
		if ptComp != 0 {
			return ptComp
		}
		i++
	}
	if i < size1 {
		return 1
	}
	if i < size2 {
		return -1
	}
	return 0
}

// CompareCoordinate compares the same coordinate of two Geom_CoordinateSequences
// along the given number of dimensions.
//
// Returns -1, 0, or 1 depending on whether s1[i] is less than, equal to, or
// greater than s2[i].
func (c *Geom_CoordinateSequenceComparator) CompareCoordinate(s1, s2 Geom_CoordinateSequence, i, dimension int) int {
	if impl, ok := java.GetLeaf(c).(interface {
		CompareCoordinate_BODY(Geom_CoordinateSequence, Geom_CoordinateSequence, int, int) int
	}); ok {
		return impl.CompareCoordinate_BODY(s1, s2, i, dimension)
	}
	return c.CompareCoordinate_BODY(s1, s2, i, dimension)
}

func (c *Geom_CoordinateSequenceComparator) CompareCoordinate_BODY(s1, s2 Geom_CoordinateSequence, i, dimension int) int {
	for d := 0; d < dimension; d++ {
		ord1 := s1.GetOrdinate(i, d)
		ord2 := s2.GetOrdinate(i, d)
		comp := Geom_CoordinateSequenceComparator_Compare(ord1, ord2)
		if comp != 0 {
			return comp
		}
	}
	return 0
}
