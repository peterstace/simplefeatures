package geom

import "fmt"

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
	if !a.IsPolygon() || !b.IsPolygon() {
		// TODO: support all other input geometry types.
		panic(fmt.Sprintf("binary op not implemented for types %s and %s", a.Type(), b.Type()))
	}
	polyA := a.AsPolygon()
	polyB := b.AsPolygon()

	dcelA := newDCELFromPolygon(polyA, inputAValue|inputAPresent)
	dcelB := newDCELFromPolygon(polyB, inputBValue|inputBPresent)
	dcelA.reNodeGraph(polyB)
	dcelB.reNodeGraph(polyA)

	dcelA.overlay(dcelB)
	return dcelA.toGeometry(include)
}
