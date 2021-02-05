package de9im

import "fmt"

// Matrix is a 3 by 3 intersection matrix that describes the intersection
// between two geometries. Specifically, it considers the Interior (I),
// Boundary (B), and Exterior (E) of each geometry separately, and shows how
// each part intersects with the 3 parts of the other geometry.
//
// Each entry in the matrix holds the dimension of the set formed when a
// specific combination of I, B, and E (one from each geometry) are intersected
// with each other. The entries are 2 for an areal intersection, 1 for a linear
// intersection, and 0 for a point intersection. The entry is F if there is no
// intersection at all (F stands for 'False').
//
// For example, the BI entry could contain a 1 if the set formed by
// intersecting the boundary of the first geometry and the interior of the
// second geometry has dimension 1.
//
// The zero value of Matrix is valid, and contains F entries everywhere
// (representing the empty intersection between two disjoint geometries).
type Matrix uint32

// Implementation detail for Matrix:
//
// Matrix is a bit field, where each matrix entry occupies 4 bit. The order of
// the encoding is: II, IB, IE, BI, BB, BE, EI, EB, EE. II is stored in the 2
// least significant bits. This uses 18 bits total, leaving the 14 most
// significant bits unused. Users SHOULD NOT manipulate the bits in a Matrix
// manually and treat the Matrix type opaquely. This is because the type's
// implementation details may change in the future.

// MatrixFromStringCode creates a matrix from its standard code representation.
// The standard code representation is a 9 digit string containing the
// characters '0', '1', '2', and 'F'. The order of the digits in the string is
// II, IB, IE, BI, BB, BE, EI, EB, EE.
func MatrixFromStringCode(code string) (Matrix, error) {
	if len(code) != 9 {
		return 0, fmt.Errorf("code length %d is invalid (must be 9)", len(code))
	}
	var m Matrix
	for i, c := range code {
		var dim Dimension
		switch c {
		case 'F':
			dim = Empty
		case '0':
			dim = Dim0
		case '1':
			dim = Dim1
		case '2':
			dim = Dim2
		default:
			return 0, fmt.Errorf("code is invalid, contains byte %d", c)
		}
		m |= Matrix(dim) << (i * 2)
	}
	return m, nil
}

// StringCode returns the standard code representation of the Matrix. It is a 9
// character string containing the characters '0', '1', '2', and 'F'. The order
// of the digits in the string is II, BI, EI, IB, BB, EB, IE, BE, EE.
func (m Matrix) StringCode() string {
	var buf [9]byte
	for i := 0; i < 9; i++ {
		shift := i * 2
		raw := byte((m & (3 << shift)) >> shift)
		buf[i] = [...]byte{'F', '0', '1', '2'}[raw]
	}
	return string(buf[:])
}

// Location is a location relative to a geometry. A location can be in the
// interior, boundary, or exterior of a geometry.
type Location uint32

const (
	Interior Location = 0
	Boundary Location = 1
	Exterior Location = 2
)

// String gives a textual representation of the location, returning "I"
// (Interior), "B" (Boundary), or "E" (Exterior).
func (o Location) String() string {
	switch o {
	case Interior:
		return "I"
	case Boundary:
		return "B"
	case Exterior:
		return "E"
	default:
		return fmt.Sprintf("unknown_location(%d)", o)
	}
}

// Dimension is the dimension of the set formed when two sets intersect.
type Dimension uint32

const (
	// Empty indicates that two intersecting sets are disjoint.
	Empty Dimension = 0

	// Dim0 indicates that the two sets intersect, but only at points (no
	// linear elements or areal elements).
	Dim0 Dimension = 1

	// Dim1 indicates that the set formed when two sets intersect contains
	// linear elements but no areal elements. Point elements may or may not be
	// present.
	Dim1 Dimension = 2

	// Dim2 indicatse that the set formed when two sets intersect contains
	// areal elements. Point and linear elements may or may not be present.
	Dim2 Dimension = 3
)

// Dimension gives a textual representation of the dimension, returning "F"
// (empty), "0" (dim 0), "1" (dim 1), or "2" (dim 2).
func (d Dimension) String() string {
	switch d {
	case Empty:
		return "F"
	case Dim0:
		return "0"
	case Dim1:
		return "1"
	case Dim2:
		return "2"
	default:
		return fmt.Sprintf("unknown_dimension(%d)", d)
	}
}

// MaxDimension finds the maximum dimension out of the two input
// dimensions.
func MaxDimension(dimA, dimB Dimension) Dimension {
	if dimA > dimB {
		return dimA
	}
	return dimB
}

// MinDimension finds the minimum dimension out of the two input
// dimensions.
func MinDimension(dimA, dimB Dimension) Dimension {
	if dimA < dimB {
		return dimA
	}
	return dimB
}

// With returns a new Matrix that has a single entry changed compared to the
// original. The original is not changed.
func (m Matrix) With(locA, locB Location, dim Dimension) Matrix {
	shift := (3*locA + locB) * 2
	var mask Matrix = 3 << shift
	return (m & ^mask) | Matrix(dim<<shift)
}

// Get returns an entry from the matrix.
func (m Matrix) Get(locA, locB Location) Dimension {
	shift := (3*locA + locB) * 2
	var mask Matrix = 3 << shift
	raw := (m & mask) >> shift
	return Dimension(raw)
}
