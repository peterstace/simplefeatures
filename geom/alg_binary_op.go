package geom

// Union returns a geometry that represents the parts from either geometry A or
// geometry B (or both). An error may be returned in pathological cases of
// numerical degeneracy.
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
// cases of numerical degeneracy.
func Intersection(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() || b.IsEmpty() {
		return Geometry{}, nil
	}
	return binaryOp(a, b, selectIntersection)
}

// Difference returns a geometry that represents the parts of input geometry A
// that are not part of input geometry B. An error may be returned in cases of
// pathological cases of numerical degeneracy.
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
// cases of numerical degeneracy.
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
	overlay := createOverlay(a, b)
	return overlay.extractGeometry(include)
}

func createOverlay(a, b Geometry) *doublyConnectedEdgeList {
	a, b = reNodeGeometries(a, b)
	dcelA := newDCELFromGeometry(a, inputAMask)
	dcelB := newDCELFromGeometry(b, inputBMask)
	dcelA.overlay(dcelB)
	return dcelA
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
