package geom

import (
	"errors"
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
	return binaryOp(a, b, selectUnion)
}

// Intersection returns a geometry that represents the parts that are common to
// both geometry A and geometry B. An error may be returned in pathological
// cases of numerical degeneracy. GeometryCollections are not supported.
func Intersection(a, b Geometry) (Geometry, error) {
	insertOrGetCount = 0
	appendNewNodeCount = 0
	appendLinesInsertOrGetCount = 0
	reNodeLineStringInsertOrGetCount = 0
	defer func() {
		fmt.Println("insertOrGetCount", insertOrGetCount)
		fmt.Println("appendNewNodeCount", appendNewNodeCount)
		fmt.Println("appendLinesInsertOrGetCount", appendLinesInsertOrGetCount)
		fmt.Println("reNodeLineStringInsertOrGetCount", reNodeLineStringInsertOrGetCount)
	}()
	if a.IsEmpty() || b.IsEmpty() {
		return Geometry{}, nil
	}
	return binaryOp(a, b, selectIntersection)
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
	return binaryOp(a, b, selectDifference)
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
	return binaryOp(a, b, selectSymmetricDifference)
}

func binaryOp(a, b Geometry, include func(uint8) bool) (Geometry, error) {
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

func createOverlay(a, b Geometry) (*doublyConnectedEdgeList, error) {
	if a.IsGeometryCollection() || b.IsGeometryCollection() {
		return nil, errors.New("GeometryCollection argument not supported")
	}

	aGhost := connectGeometry(a)
	bGhost := connectGeometry(b)
	joinGhost := connectGeometries(a, b)
	ghosts := mergeMultiLineStrings([]MultiLineString{aGhost, bGhost, joinGhost.AsMultiLineString()})

	a, b, ghosts, err := reNodeGeometries(a, b, ghosts)
	if err != nil {
		return nil, err
	}

	interactionPoints := findInteractionPoints([]Geometry{a, b, ghosts.AsGeometry()})

	dcelA := newDCELFromGeometry(a, ghosts, inputAMask, interactionPoints)
	dcelB := newDCELFromGeometry(b, ghosts, inputBMask, interactionPoints)

	dcelA.overlay(dcelB)
	return dcelA, nil
}
