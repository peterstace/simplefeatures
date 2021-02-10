package geom

import (
	"errors"
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

func relateWithMask(a, b Geometry, mask string) (bool, error) {
	return false, errors.New("not implemented")
}

func Equals(a, b Geometry) (bool, error) {
	if a.IsEmpty() && b.IsEmpty() {
		// Part of the mask is 'dim(I(a) ∩ I(b)) = T'.  If both inputs are
		// empty, then their interiors will be empty, and thus 'dim(I(a) ∩ I(b)
		// = F'. However, we want to return 'true' for this case. So we just
		// return true manually rather than using DE-9IM.
		return true, nil
	}
	return relateWithMask(a, b, "T*F**FFF*")
}

func Disjoint(a, b Geometry) (bool, error) {
	// TODO
	return false, errors.New("not implemented")
}

func Touches(a, b Geometry) (bool, error) {
	// TODO
	return false, errors.New("not implemented")
}

func Contains(a, b Geometry) (bool, error) {
	// TODO
	return false, errors.New("not implemented")
}

func Covers(a, b Geometry) (bool, error) {
	// TODO
	return false, errors.New("not implemented")
}

func Within(a, b Geometry) (bool, error) {
	// TODO
	return false, errors.New("not implemented")
}

func CoveredBy(a, b Geometry) (bool, error) {
	// TODO
	return false, errors.New("not implemented")
}

func Crosses(a, b Geometry) (bool, error) {
	// TODO
	return false, errors.New("not implemented")
}

func Overlaps(a, b Geometry) (bool, error) {
	// TODO
	return false, errors.New("not implemented")
}
