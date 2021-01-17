package geom

import (
	"fmt"

	"github.com/peterstace/simplefeatures/de9im"
)

// Relate calculates the DE-9IM matrix between two geometries, describing how
// the two geometries relate to each other.
func Relate(a, b Geometry) (de9im.Matrix, error) {
	if a.IsEmpty() || b.IsEmpty() {
		// TODO: Eliminate duplication here. There might be a more elegant way
		// to do this.
		var m de9im.Matrix
		m = m.With(de9im.Exterior, de9im.Exterior, de9im.Dim2)
		if !b.IsEmpty() {
			switch b.Dimension() {
			case 0:
				m = m.With(de9im.Exterior, de9im.Interior, de9im.Dim0)
				m = m.With(de9im.Exterior, de9im.Boundary, de9im.Empty)
			case 1:
				m = m.With(de9im.Exterior, de9im.Interior, de9im.Dim1)
				m = m.With(de9im.Exterior, de9im.Boundary, de9im.Dim0)
			case 2:
				m = m.With(de9im.Exterior, de9im.Interior, de9im.Dim2)
				m = m.With(de9im.Exterior, de9im.Boundary, de9im.Dim1)
			}
		}
		if !a.IsEmpty() {
			switch a.Dimension() {
			case 0:
				m = m.With(de9im.Interior, de9im.Exterior, de9im.Dim0)
				m = m.With(de9im.Boundary, de9im.Exterior, de9im.Empty)
			case 1:
				m = m.With(de9im.Interior, de9im.Exterior, de9im.Dim1)
				m = m.With(de9im.Boundary, de9im.Exterior, de9im.Dim0)
			case 2:
				m = m.With(de9im.Interior, de9im.Exterior, de9im.Dim2)
				m = m.With(de9im.Boundary, de9im.Exterior, de9im.Dim1)
			}
		}
		return m, nil
	}

	overlay, err := createOverlay(a, b)
	if err != nil {
		return 0, fmt.Errorf("internal error creating overlay: %v", err)
	}
	return overlay.extractIntersectionMatrix(), nil
}
