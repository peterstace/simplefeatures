package geom

// Union returns a geometry that represents the parts that are common to either
// geometry A or geometry B (or both).
func Union(a, b Geometry) Geometry {
	return binaryOp(a, b, selectUnion)
}

// Intersection returns a geometry that represents the parts that are common to
// both geometry A and geometry B.
func Intersection(a, b Geometry) Geometry {
	return binaryOp(a, b, selectIntersection)
}

// Difference returns a geometry that represents the parts of input geometry A
// that do not intersect with any parts of input geometry B.
func Difference(a, b Geometry) Geometry {
	return binaryOp(a, b, selectDifference)
}

// SymmetricDifference returns a geometry that represents the parts of geometry
// A and B that do not intersect with each other.
func SymmetricDifference(a, b Geometry) Geometry {
	return binaryOp(a, b, selectSymmetricDifference)
}

func binaryOp(a, b Geometry, include func(uint8) bool) Geometry {
	dcelA := newDCELFromGeometry(a, inputAMask)
	dcelB := newDCELFromGeometry(b, inputBMask)

	var linesA []line
	switch {
	case a.IsPolygon():
		linesA = a.AsPolygon().Boundary().asLines()
	case a.IsMultiPolygon():
		linesA = a.AsMultiPolygon().Boundary().asLines()
	default:
		panic("not supported")
	}

	var linesB []line
	switch {
	case b.IsPolygon():
		linesB = b.AsPolygon().Boundary().asLines()
	case b.IsMultiPolygon():
		linesB = b.AsMultiPolygon().Boundary().asLines()
	default:
		panic("not supported")
	}

	dcelA.reNodeGraph(linesB)
	dcelB.reNodeGraph(linesA)

	dcelA.overlay(dcelB)
	return dcelA.toGeometry(include)
}
