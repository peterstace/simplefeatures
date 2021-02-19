package geom

import (
	"fmt"
)

// Union returns a geometry that represents the parts from either geometry A or
// geometry B (or both). An error may be returned in pathological cases of
// numerical degeneracy. GeometryCollections are not supported.
func Union(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() && b.IsEmpty() {
		return Geometry{}, nil
	}
	if a.IsEmpty() {
		return b, nil
	}
	if b.IsEmpty() {
		return a, nil
	}
	return setOp(a, b, selectUnion)
}

// Intersection returns a geometry that represents the parts that are common to
// both geometry A and geometry B. An error may be returned in pathological
// cases of numerical degeneracy. GeometryCollections are not supported.
func Intersection(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() || b.IsEmpty() {
		return Geometry{}, nil
	}
	return setOp(a, b, selectIntersection)
}

// Difference returns a geometry that represents the parts of input geometry A
// that are not part of input geometry B. An error may be returned in cases of
// pathological cases of numerical degeneracy. GeometryCollections are not
// supported.
func Difference(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() {
		return Geometry{}, nil
	}
	if b.IsEmpty() {
		return a, nil
	}
	return setOp(a, b, selectDifference)
}

// SymmetricDifference returns a geometry that represents the parts of geometry
// A and B that are not in common. An error may be returned in pathological
// cases of numerical degeneracy. GeometryCollections are not supported.
func SymmetricDifference(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() && b.IsEmpty() {
		return Geometry{}, nil
	}
	if a.IsEmpty() {
		return b, nil
	}
	if b.IsEmpty() {
		return a, nil
	}
	return setOp(a, b, selectSymmetricDifference)
}

func setOp(a, b Geometry, include func([2]label) bool) (Geometry, error) {
	overlay, err := createOverlay(a, b)
	if err != nil {
		return Geometry{}, fmt.Errorf("internal error creating overlay: %v", err)
	}

	g, err := overlay.extractGeometry(include)
	if err != nil {
		return Geometry{}, fmt.Errorf("internal error extracting geometry: %v", err)
	}
	return g, nil
}
