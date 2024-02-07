package geom

import (
	"errors"
	"fmt"
)

// imLocation is a location relative to a geometry. A location can be in the
// interior, boundary, or exterior of a geometry.
type imLocation uint32

const (
	imInterior imLocation = 0
	imBoundary imLocation = 1
	imExterior imLocation = 2
)

type matrix [9]byte

func newMatrix() matrix {
	return [9]byte{'F', 'F', 'F', 'F', 'F', 'F', 'F', 'F', 'F'}
}

func (m *matrix) set(locA, locB imLocation, entry byte) {
	idx := m.index(locA, locB)
	m[idx] = entry
}

func (m *matrix) get(locA, locB imLocation) byte {
	idx := m.index(locA, locB)
	return m[idx]
}

func (m *matrix) code() string {
	return string(m[:])
}

func (matrix) index(locA, locB imLocation) int {
	return int(3*locA + locB)
}

func (m *matrix) transpose() {
	cp := *m
	for _, locA := range []imLocation{imInterior, imBoundary, imExterior} {
		for _, locB := range []imLocation{imInterior, imBoundary, imExterior} {
			m.set(locB, locA, cp.get(locA, locB))
		}
	}
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
		return false, errors.New("invalid matrix: length is not 9")
	}
	if len(pat) != 9 {
		return false, errors.New("invalid matrix pattern: length is not 9")
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
