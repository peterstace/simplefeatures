package geom

import (
	"errors"
	"fmt"
)

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
// The zero value of IntersectionMatrix is well defined and contains F entries
// everywhere (although it doesn't represent a sensible intersection between 2
// geometries).
type IntersectionMatrix struct {
	// Implementation details: The matrix is stored in a bit field, where each
	// entry occupies 4 bits.  The order of the encoding is: II, IB, IE, BI,
	// BB, BE, EI, EB, EE.  II is stored in the 2 least significant bits. This
	// uses 18 bits total, leaving the 14 most significant bits unused.
	bits uint32
}

// IntersectionMatrixFromStringCode creates a matrix from its standard code
// representation.  The standard code representation is a 9 digit string
// containing the characters '0', '1', '2', and 'F'. The order of the digits in
// the string is II, IB, IE, BI, BB, BE, EI, EB, EE.
func IntersectionMatrixFromStringCode(code string) (IntersectionMatrix, error) {
	if len(code) != 9 {
		return IntersectionMatrix{}, fmt.Errorf("code length %d is invalid (must be 9)", len(code))
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
			return IntersectionMatrix{}, fmt.Errorf("code is invalid, contains byte %d", c)
		}
		m.bits |= uint32(dim) << (i * 2)
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
		raw := byte((m.bits & (3 << shift)) >> shift)
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

// with returns a new IntersectionMatrix that has a single entry changed
// compared to the original. The original is not changed.
func (m IntersectionMatrix) with(locA, locB imLocation, dim imEntry) IntersectionMatrix {
	shift := (3*locA + locB) * 2
	var mask uint32 = 3 << shift
	return IntersectionMatrix{(m.bits & ^mask) | (uint32(dim) << shift)}
}

// get returns an entry from the matrix.
func (m IntersectionMatrix) get(locA, locB imLocation) imEntry {
	shift := (3*locA + locB) * 2
	var mask uint32 = 3 << shift
	raw := (m.bits & mask) >> shift
	return imEntry(raw)
}

// transpose returns the original intersection matrix, but flipped along its
// diagonal.
func (m IntersectionMatrix) transpose() IntersectionMatrix {
	var t IntersectionMatrix
	for _, locA := range []imLocation{imInterior, imBoundary, imExterior} {
		for _, locB := range []imLocation{imInterior, imBoundary, imExterior} {
			t = t.with(locB, locA, m.get(locA, locB))
		}
	}
	return t
}

// IntersectionMask is a 3 by 3 matrix mask describing a spatial predicate
// between two geometries. It is a mask for an IntersectionMatrix. Each entry
// is one of 'F', '0', '1', '2', 'T', or '*'.
//
// The spatial predicate between two geometries represented by the mask is true
// if and only if each entry of the mask matches the corresponding entry in the
// intersection matrix. The matching rules are:
//
// 1. 'F' matches against 'F' or '*'.
//
// 2. '0' matches against '0', 'T', or '*'.
//
// 3. '1' matches against '1', 'T', or '*'.
//
// 4. '2' matches against '2', 'T', or '*'.
//
// The zero value of IntersectionMask is well defined and contains all 'F'
// entries.
type IntersectionMask struct {
	// Implementation details: the mask matrix is stored as a bit field, with
	// each entry storing 6 possible values ('F', '0', '1', '2', 'T', or '*').
	// We need 3 bits per entry, so 27 bits total. This leaves 5 bits unused.
	bits uint32
}

// Matches checks if all entries in the intersection matrix match against the
// corresponding entries in the intersection mask. If Matches returns true,
// then the spatial predicate described by the mask is true for the pair of
// geometries used to create the intersection matrix.
func (m IntersectionMask) Matches(im IntersectionMatrix) bool {
	return false // TODO
}

// IntersectionMaskFromStringCode creates a new intersection mask from its
// code. The code should be 9 characters long and consist of only 'F', '0',
// '1', '2', 'T', and '*'.
func IntersectionMaskFromStringCode(code string) (IntersectionMask, error) {
	// TODO
	return IntersectionMask{}, errors.New("not implemented")
}

// StringCode returns the 9 character string code representing the intersection
// matrix.
func (m IntersectionMask) StringCode() string {
	return "NOT IMPLEMENTED" // TODO
}
