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

	linesA := geometryToLines(a)
	linesB := geometryToLines(b)
	dcelA.reNodeGraph(linesB)
	dcelB.reNodeGraph(linesA)

	dcelA.overlay(dcelB)
	return dcelA.extractGeometry(include)
}

func geometryToLines(g Geometry) []line {
	switch g.Type() {
	case TypePolygon:
		return g.AsPolygon().Boundary().asLines()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().Boundary().asLines()
	default:
		panic("not supported")
	}
}
