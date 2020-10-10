package geom

// Union returns a geometry that represents the parts from either geometry A or
// geometry B (or both).
func Union(a, b Geometry) Geometry {
	return binaryOp(a, b, selectUnion)
}

// Intersection returns a geometry that represents the parts that are common to
// both geometry A and geometry B.
func Intersection(a, b Geometry) Geometry {
	return binaryOp(a, b, selectIntersection)
}

// Difference returns a geometry that represents the parts of input geometry A
// that are not part of input geometry B.
func Difference(a, b Geometry) Geometry {
	return binaryOp(a, b, selectDifference)
}

// SymmetricDifference returns a geometry that represents the parts of geometry
// A and B that are not in common.
func SymmetricDifference(a, b Geometry) Geometry {
	return binaryOp(a, b, selectSymmetricDifference)
}

func binaryOp(a, b Geometry, include func(uint8) bool) Geometry {
	overlay := createOverlay(a, b)
	return overlay.extractGeometry(include)
}

func createOverlay(a, b Geometry) *doublyConnectedEdgeList {
	cutA := newCutSet(a)
	cutB := newCutSet(b)

	// Re-node any linear elements with themselves, since they may be
	// self-intersecting.
	if containsLinearElement(a) {
		a = reNodeGeometry(a, cutA)
		cutA = newCutSet(a)
	}
	if containsLinearElement(b) {
		b = reNodeGeometry(b, cutB)
		cutB = newCutSet(b)
	}

	a = reNodeGeometry(a, cutB)
	b = reNodeGeometry(b, cutA)

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
