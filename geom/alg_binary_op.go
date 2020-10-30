package geom

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
		return Geometry{}, err
	}
	return overlay.extractGeometry(include)
}

func createOverlay(a, b Geometry) (*doublyConnectedEdgeList, error) {
	a, b = reNodeGeometries(a, b)
	dcelA, err := newDCELFromGeometry(a, inputAMask)
	if err != nil {
		return nil, err
	}
	dcelB, err := newDCELFromGeometry(b, inputBMask)
	if err != nil {
		return nil, err
	}
	dcelA.overlay(dcelB)
	return dcelA, nil
}

func containsLinearElement(g Geometry) bool {
	switch g.Type() {
	case TypeLineString, TypeMultiLineString:
		return true
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			g := gc.GeometryN(i)
			if containsLinearElement(g) {
				return true
			}
		}
	}
	return false
}
