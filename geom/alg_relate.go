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

func relateMatchesAnyPattern(a, b Geometry, patterns ...string) (bool, error) {
	//intersectionMatrix, err := Relate(a, b)
	//if err != nil {
	//return false, err
	//}
	//for _, pat := range patterns {
	// TODO: check against pattern
	//}
	//return true, nil
	return false, errors.New("not implemented")
}

// Equals returns true if and only if the input geometries are spatially equal,
// i.e. they represent exactly the same set of points.
func Equals(a, b Geometry) (bool, error) {
	if a.IsEmpty() && b.IsEmpty() {
		// Part of the mask is 'dim(I(a) ∩ I(b)) = T'.  If both inputs are
		// empty, then their interiors will be empty, and thus 'dim(I(a) ∩ I(b)
		// = F'. However, we want to return 'true' for this case. So we just
		// return true manually rather than using DE-9IM.
		return true, nil
	}
	return relateMatchesAnyPattern(a, b, "T*F**FFF*")
}

// Disjoint returns true if and only if the input geometries have no points in
// common.
func Disjoint(a, b Geometry) (bool, error) {
	return relateMatchesAnyPattern(a, b, "FF*FF****")
}

// Touches returns true if and only if the geometries have at least 1 point in
// common, but their interiors don't intersect.
func Touches(a, b Geometry) (bool, error) {
	return relateMatchesAnyPattern(
		a, b,
		"FT*******",
		"F**T*****",
		"F***T****",
	)
}

// Contains returns true if and only if geometry A contains geometry B.
// Formally, the following two conditions must hold:
//
// 1. No points of B lies on the exterior of geometry A. That is, B must only
// be in the exterior or boundary of A.
//
// 2 .At least one point of the interior of B lies on the interior of A. That
// is, they can't *only* intersect at their boundaries.
func Contains(a, b Geometry) (bool, error) {
	return relateMatchesAnyPattern(a, b, "T*****FF*")
}

// Covers returns true if and only if geometry A covers geometry B. Formally,
// the following two conditions must hold:
//
// 1. No points of B lies on the exterior of geometry A. That is, B must only
// be in the exterior or boundary of A.
//
// 2. At least one point of B lies on A (either its interior or boundary).
func Covers(a, b Geometry) (bool, error) {
	return relateMatchesAnyPattern(
		a, b,
		"T*****FF*",
		"*T****FF*",
		"***T**FF*",
		"****T*FF*",
	)
}

// Within returns true if and only if geometry A is completely within geometry
// B. Formally, the following two conditions must hold:
//
// 1. No points of A lies on the exterior of geometry B. That is, A must only
// be in the exterior or boundary of B.
//
// 2.At least one point of the interior of A lies on the interior of B. That
// is, they can't *only* intersect at their boundaries.
func Within(a, b Geometry) (bool, error) {
	return relateMatchesAnyPattern(a, b, "T*F**F***")
}

// CoveredBy returns true if and only if geometry A is covered by geometry B.
// Formally, the following two conditions must hold:
//
// 1. No points of A lies on the exterior of geometry B. That is, A must only
// be in the exterior or boundary of B.
//
// 2. At least one point of A lies on B (either its interior or boundary).
func CoveredBy(a, b Geometry) (bool, error) {
	return relateMatchesAnyPattern(
		a, b,
		"T*F**F***",
		"*TF**F***",
		"**FT*F***",
		"**F*TF***",
	)
}

// Crosses returns true if and only if geometry A and B cross each other.
// Formally, the following conditions must hold:
//
// 1. The geometries must have some but not all interior points in common.
//
// 2. The dimensionality of the intersection must be less than the maximum
// dimension of the input geometries.
//
// 3. The intersection must not equal either of the input geometries.
func Crosses(a, b Geometry) (bool, error) {
	dimA := a.Dimension()
	dimB := b.Dimension()
	switch {
	case dimA < dimB: // Point/Line, Point/Area, Line/Area
		return relateMatchesAnyPattern(a, b, "T*T******")
	case dimA > dimB: // Line/Point, Area/Point, Area/Line
		return relateMatchesAnyPattern(a, b, "T*****T**")
	case dimA == 1 && dimB == 1: // Line/Line
		return relateMatchesAnyPattern(a, b, "0********")
	default:
		return false, nil
	}
}

// Overlaps returns true if and only if geometry A and B overlap with each
// other. Formally, the following conditions must hold:
//
// 1. The geometries must have the same dimension.
//
// 2. The geometries must have some but not all points in common.
//
// 3. The intersection of the geometries must have the same dimension as the
// geometries themselves.
func Overlaps(a, b Geometry) (bool, error) {
	dimA := a.Dimension()
	dimB := b.Dimension()
	switch {
	case (dimA == 0 && dimB == 0) || (dimA == 2 && dimB == 2):
		return relateMatchesAnyPattern(a, b, "T*T***T**")
	case (dimA == 1 && dimB == 1):
		return relateMatchesAnyPattern(a, b, "1*T***T**")
	default:
		return false, nil
	}
}
