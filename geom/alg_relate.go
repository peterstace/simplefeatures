package geom

import (
	"fmt"
)

// Relate calculates the DE-9IM matrix between two geometries, describing how
// the two geometries relate to each other.
func Relate(a, b Geometry) (IntersectionMatrix, error) {
	if a.IsEmpty() || b.IsEmpty() {
		// TODO: Eliminate duplication here. There might be a more elegant way
		// to do this.
		var m IntersectionMatrix
		m = m.with(imExterior, imExterior, imEntry2)
		if !b.IsEmpty() {
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
		}
		if !a.IsEmpty() {
			switch a.Dimension() {
			case 0:
				m = m.with(imInterior, imExterior, imEntry0)
				m = m.with(imBoundary, imExterior, imEntryF)
			case 1:
				m = m.with(imInterior, imExterior, imEntry1)
				if !a.Boundary().IsEmpty() {
					m = m.with(imBoundary, imExterior, imEntry0)
				}
			case 2:
				m = m.with(imInterior, imExterior, imEntry2)
				m = m.with(imBoundary, imExterior, imEntry1)
			}
		}
		return m, nil
	}

	overlay, err := createOverlay(a, b)
	if err != nil {
		return IntersectionMatrix{}, fmt.Errorf("internal error creating overlay: %v", err)
	}
	return overlay.extractIntersectionMatrix(), nil
}
