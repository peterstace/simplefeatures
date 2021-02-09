package geom

import (
	"fmt"
)

// Relate calculates the DE-9IM matrix between two geometries, describing how
// the two geometries relate to each other.
func Relate(a, b Geometry) (IntersectionMatrix, error) {
	if a.IsEmpty() || b.IsEmpty() {
		var m IntersectionMatrix
		m = m.with(imExterior, imExterior, imEntry2)
		if a.IsEmpty() && b.IsEmpty() {
			return m, nil
		}

		var flip bool
		if b.IsEmpty() {
			b, a = a, b
			flip = true
		}
		switch b.Dimension() {
		case 0:
			m = m.with(imExterior, imInterior, imEntry0)
			m = m.with(imExterior, imBoundary, imEntryF)
		case 1:
			m = m.with(imExterior, imInterior, imEntry1)
			if !b.Boundary().IsEmpty() {
				m = m.with(imExterior, imBoundary, imEntry0)
			}
		case 2:
			m = m.with(imExterior, imInterior, imEntry2)
			m = m.with(imExterior, imBoundary, imEntry1)
		}
		if flip {
			m = m.transpose()
		}
		return m, nil
	}

	overlay, err := createOverlay(a, b)
	if err != nil {
		return IntersectionMatrix{}, fmt.Errorf("internal error creating overlay: %v", err)
	}
	return overlay.extractIntersectionMatrix(), nil
}
