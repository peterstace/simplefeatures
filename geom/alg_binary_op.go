package geom

// Union returns a geometry that represents the parts from either geometry A or
// geometry B (or both).
func Union(a, b Geometry) Geometry {
	if a.IsEmpty() && b.IsEmpty() {
		return Geometry{}
	}
	if a.IsEmpty() {
		return b
	}
	if b.IsEmpty() {
		return a
	}
	return binaryOp(a, b, selectUnion)
}

// Intersection returns a geometry that represents the parts that are common to
// both geometry A and geometry B.
func Intersection(a, b Geometry) Geometry {
	if a.IsEmpty() || b.IsEmpty() {
		return Geometry{}
	}
	return binaryOp(a, b, selectIntersection)
}

// Difference returns a geometry that represents the parts of input geometry A
// that are not part of input geometry B.
func Difference(a, b Geometry) Geometry {
	if a.IsEmpty() {
		return Geometry{}
	}
	if b.IsEmpty() {
		return a
	}
	return binaryOp(a, b, selectDifference)
}

// SymmetricDifference returns a geometry that represents the parts of geometry
// A and B that are not in common.
func SymmetricDifference(a, b Geometry) Geometry {
	if a.IsEmpty() && b.IsEmpty() {
		return Geometry{}
	}
	if a.IsEmpty() {
		return b
	}
	if b.IsEmpty() {
		return a
	}
	return binaryOp(a, b, selectSymmetricDifference)
}

func binaryOp(a, b Geometry, include func(uint8) bool) Geometry {
	overlay := createOverlay(a, b)
	return overlay.extractGeometry(include)
}

func createOverlay(a, b Geometry) *doublyConnectedEdgeList {
	cut := newCutSet(a, b)
	a = reNodeGeometry(a, cut)
	b = reNodeGeometry(b, cut)

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
