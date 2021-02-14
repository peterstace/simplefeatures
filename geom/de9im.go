package geom

import (
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

// RelateMatches checks to see if an intersection matrix matches against an
// intersection matrix pattern. Each is a 9 character string that encodes a 3
// by 3 matrix.
//
// The intersection matrix has the same format as those computed by the Relate
// function. That is, it must be a 9 character string consisting of 'F', '0',
// '1', and '2' entries.
//
// The intersection matrix pattern is also 9 characters, and consists of 'F',
// '0', '1', '2', 'T', and '*' entries.
//
// An intersection matrix matches against an intersection matrix pattern if
// each entry in the intersection matrix matches against the corresponding
// entry in the intersection matrix pattern. An 'F' entry matches against an
// 'F' or '*' pattern. A '0' entry matches against '0', 'T', or '*'. A '1'
// entry matches against '1', 'T', or '*'. A '2' entry matches against '2',
// 'T', or '*'.
func RelateMatches(intersectionMatrix, intersectionMatrixPattern string) (bool, error) {
	mat := intersectionMatrix
	pat := intersectionMatrixPattern
	if len(mat) != 9 {
		return false, fmt.Errorf("invalid matrix: length is not 9")
	}
	if len(pat) != 9 {
		return false, fmt.Errorf("invalid matrix pattern: length is not 9")
	}

	for i, m := range mat {
		p := pat[i]
		switch p {
		case 'F', '0', '1', '2', 'T', '*':
		default:
			return false, fmt.Errorf("invalid character in intersection pattern: %c", p)
		}

		switch m {
		case 'F':
			if p != 'F' && p != '*' {
				return false, nil
			}
		case '0':
			if p != '0' && p != 'T' && p != '*' {
				return false, nil
			}
		case '1':
			if p != '1' && p != 'T' && p != '*' {
				return false, nil
			}
		case '2':
			if p != '2' && p != 'T' && p != '*' {
				return false, nil
			}
		default:
			return false, fmt.Errorf("invalid character in intersection matrix: %c", m)
		}
	}
	return true, nil
}
