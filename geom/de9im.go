package geom

import "fmt"

// IntersectionMatrix is a 3 by 3 matrix that describes the intersection
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
// The zero value of IntersectionMatrix is valid, and contains F entries everywhere
// (representing the empty intersection between two disjoint geometries).
//
// TODO: make this a struct with a non-exported field so that it's not
// modifiable by users.
type IntersectionMatrix uint32

// Implementation detail for Matrix:
//
// Matrix is a bit field, where each matrix entry occupies 4 bit. The order of
// the encoding is: II, IB, IE, BI, BB, BE, EI, EB, EE. II is stored in the 2
// least significant bits. This uses 18 bits total, leaving the 14 most
// significant bits unused. Users SHOULD NOT manipulate the bits in a Matrix
// manually and treat the Matrix type opaquely. This is because the type's
// implementation details may change in the future.

// IntersectionMatrixFromStringCode creates a matrix from its standard code
// representation.  The standard code representation is a 9 digit string
// containing the characters '0', '1', '2', and 'F'. The order of the digits in
// the string is II, IB, IE, BI, BB, BE, EI, EB, EE.
func IntersectionMatrixFromStringCode(code string) (IntersectionMatrix, error) {
	if len(code) != 9 {
		return 0, fmt.Errorf("code length %d is invalid (must be 9)", len(code))
	}
	var m IntersectionMatrix
	for i, c := range code {
		var dim imEntry
		switch c {
		case 'F':
			dim = imEntryF
		case '0':
			dim = imEntry0
		case '1':
			dim = imEntry1
		case '2':
			dim = imEntry2
		default:
			return 0, fmt.Errorf("code is invalid, contains byte %d", c)
		}
		m |= IntersectionMatrix(dim) << (i * 2)
	}
	return m, nil
}

// StringCode returns the standard code representation of the
// IntersectionMatrix. It is a 9 character string containing the characters
// '0', '1', '2', and 'F'. The order of the digits in the string is II, BI, EI,
// IB, BB, EB, IE, BE, EE.
func (m IntersectionMatrix) StringCode() string {
	var buf [9]byte
	for i := 0; i < 9; i++ {
		shift := i * 2
		raw := byte((m & (3 << shift)) >> shift)
		buf[i] = [...]byte{'F', '0', '1', '2'}[raw]
	}
	return string(buf[:])
}

// imLocation is a location relative to a geometry. A location can be in the
// interior, boundary, or exterior of a geometry.
type imLocation uint32

const (
	imInterior imLocation = 0
	imBoundary imLocation = 1
	imExterior imLocation = 2
)

// String gives a textual representation of the location, returning "I"
// (Interior), "B" (Boundary), or "E" (Exterior).
func (o imLocation) String() string {
	switch o {
	case imInterior:
		return "I"
	case imBoundary:
		return "B"
	case imExterior:
		return "E"
	default:
		return fmt.Sprintf("unknown_location(%d)", o)
	}
}

// imEntry is the value of an entry in an intersection matrix. It represents
// the dimension of the set formed when two sets intersect.
type imEntry uint32

const (
	// imEntryF indicates that two intersecting sets are disjoint.
	imEntryF imEntry = 0

	// imEntry0 indicates that the two sets intersect, but only at points (no
	// linear elements or areal elements).
	imEntry0 imEntry = 1

	// imEntry1 indicates that the set formed when two sets intersect contains
	// linear elements but no areal elements. Point elements may or may not be
	// present.
	imEntry1 imEntry = 2

	// imEntry2 indicatse that the set formed when two sets intersect contains
	// areal elements. Point and linear elements may or may not be present.
	imEntry2 imEntry = 3
)

// maxDimension finds the maximum dimension out of the two input
// dimensions.
func maxDimension(dimA, dimB imEntry) imEntry {
	if dimA > dimB {
		return dimA
	}
	return dimB
}

// with returns a new Matrix that has a single entry changed compared to the
// original. The original is not changed.
func (m IntersectionMatrix) with(locA, locB imLocation, dim imEntry) IntersectionMatrix {
	shift := (3*locA + locB) * 2
	var mask IntersectionMatrix = 3 << shift
	return (m & ^mask) | IntersectionMatrix(dim<<shift)
}

// get returns an entry from the matrix.
func (m IntersectionMatrix) get(locA, locB imLocation) imEntry {
	shift := (3*locA + locB) * 2
	var mask IntersectionMatrix = 3 << shift
	raw := (m & mask) >> shift
	return imEntry(raw)
}

// upgradeEntry returns a new intersection matrix that has a single entry set
// to a new dimension. If the new dimension would be lower than the existing
// one, then the original intersection matrix is returned.
func (m IntersectionMatrix) upgradeEntry(locA, locB imLocation, dim imEntry) IntersectionMatrix {
	oldEntry := m.get(locA, locB)
	newEntry := maxDimension(oldEntry, dim)
	return m.with(locA, locB, newEntry)
}
