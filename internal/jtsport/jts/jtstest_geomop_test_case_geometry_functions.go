package jts

import "math"

// JtstestGeomop_TestCaseGeometryFunctions provides geometry functions which
// augment the existing methods on Geometry, for use in XML Test files.
// This is the default used in the TestRunner, and thus all the operations
// in this class should be named differently to the Geometry methods
// (otherwise they will shadow the real Geometry methods).
//
// Ported from org.locationtech.jtstest.geomop.TestCaseGeometryFunctions.

// JtstestGeomop_TestCaseGeometryFunctions_BufferMitredJoin computes a buffer
// with mitred join style.
func JtstestGeomop_TestCaseGeometryFunctions_BufferMitredJoin(g *Geom_Geometry, distance float64) *Geom_Geometry {
	bufParams := OperationBuffer_NewBufferParameters()
	bufParams.SetJoinStyle(OperationBuffer_BufferParameters_JOIN_MITRE)
	return OperationBuffer_BufferOp_BufferOpWithParams(g, distance, bufParams)
}

// JtstestGeomop_TestCaseGeometryFunctions_Densify densifies a geometry.
func JtstestGeomop_TestCaseGeometryFunctions_Densify(g *Geom_Geometry, distance float64) *Geom_Geometry {
	return Densify_Densifier_Densify(g, distance)
}

// JtstestGeomop_TestCaseGeometryFunctions_MinClearance computes the minimum
// clearance distance of a geometry.
func JtstestGeomop_TestCaseGeometryFunctions_MinClearance(g *Geom_Geometry) float64 {
	return Precision_MinimumClearance_GetDistance(g)
}

// JtstestGeomop_TestCaseGeometryFunctions_MinClearanceLine computes the minimum
// clearance line of a geometry.
func JtstestGeomop_TestCaseGeometryFunctions_MinClearanceLine(g *Geom_Geometry) *Geom_Geometry {
	return Precision_MinimumClearance_GetLine(g)
}

func jtstestGeomop_TestCaseGeometryFunctions_polygonize(g *Geom_Geometry, extractOnlyPolygonal bool) *Geom_Geometry {
	lines := GeomUtil_LinearComponentExtracter_GetLines(g)
	polygonizer := OperationPolygonize_NewPolygonizer(extractOnlyPolygonal)
	polygonizer.AddCollection(lines)
	return polygonizer.GetGeometry()
}

// JtstestGeomop_TestCaseGeometryFunctions_Polygonize polygonizes a geometry.
func JtstestGeomop_TestCaseGeometryFunctions_Polygonize(g *Geom_Geometry) *Geom_Geometry {
	return jtstestGeomop_TestCaseGeometryFunctions_polygonize(g, false)
}

// JtstestGeomop_TestCaseGeometryFunctions_PolygonizeValidPolygonal polygonizes
// a geometry, extracting only valid polygonal results.
func JtstestGeomop_TestCaseGeometryFunctions_PolygonizeValidPolygonal(g *Geom_Geometry) *Geom_Geometry {
	return jtstestGeomop_TestCaseGeometryFunctions_polygonize(g, true)
}

// JtstestGeomop_TestCaseGeometryFunctions_SimplifyDP simplifies a geometry
// using Douglas-Peucker algorithm.
func JtstestGeomop_TestCaseGeometryFunctions_SimplifyDP(g *Geom_Geometry, distance float64) *Geom_Geometry {
	return Simplify_DouglasPeuckerSimplifier_Simplify(g, distance)
}

// JtstestGeomop_TestCaseGeometryFunctions_SimplifyTP simplifies a geometry
// using topology-preserving algorithm.
func JtstestGeomop_TestCaseGeometryFunctions_SimplifyTP(g *Geom_Geometry, distance float64) *Geom_Geometry {
	return Simplify_TopologyPreservingSimplifier_Simplify(g, distance)
}

// JtstestGeomop_TestCaseGeometryFunctions_ReducePrecision reduces the precision
// of a geometry.
func JtstestGeomop_TestCaseGeometryFunctions_ReducePrecision(g *Geom_Geometry, scaleFactor float64) *Geom_Geometry {
	return Precision_GeometryPrecisionReducer_Reduce(g, Geom_NewPrecisionModelWithScale(scaleFactor))
}

// JtstestGeomop_TestCaseGeometryFunctions_IntersectionNG computes the
// intersection using OverlayNG.
func JtstestGeomop_TestCaseGeometryFunctions_IntersectionNG(geom0, geom1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlayng_OverlayNG_Overlay(geom0, geom1, OperationOverlayng_OverlayNG_INTERSECTION, nil)
}

// JtstestGeomop_TestCaseGeometryFunctions_UnionNG computes the union using
// OverlayNG.
func JtstestGeomop_TestCaseGeometryFunctions_UnionNG(geom0, geom1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlayng_OverlayNG_Overlay(geom0, geom1, OperationOverlayng_OverlayNG_UNION, nil)
}

// JtstestGeomop_TestCaseGeometryFunctions_DifferenceNG computes the difference
// using OverlayNG.
func JtstestGeomop_TestCaseGeometryFunctions_DifferenceNG(geom0, geom1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlayng_OverlayNG_Overlay(geom0, geom1, OperationOverlayng_OverlayNG_DIFFERENCE, nil)
}

// JtstestGeomop_TestCaseGeometryFunctions_SymDifferenceNG computes the
// symmetric difference using OverlayNG.
func JtstestGeomop_TestCaseGeometryFunctions_SymDifferenceNG(geom0, geom1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlayng_OverlayNG_Overlay(geom0, geom1, OperationOverlayng_OverlayNG_SYMDIFFERENCE, nil)
}

// JtstestGeomop_TestCaseGeometryFunctions_IntersectionSR computes the
// intersection using OverlayNG with a specified precision scale.
func JtstestGeomop_TestCaseGeometryFunctions_IntersectionSR(geom0, geom1 *Geom_Geometry, scale float64) *Geom_Geometry {
	pm := Geom_NewPrecisionModelWithScale(scale)
	return OperationOverlayng_OverlayNG_Overlay(geom0, geom1, OperationOverlayng_OverlayNG_INTERSECTION, pm)
}

// JtstestGeomop_TestCaseGeometryFunctions_UnionSR computes the union using
// OverlayNG with a specified precision scale.
func JtstestGeomop_TestCaseGeometryFunctions_UnionSR(geom0, geom1 *Geom_Geometry, scale float64) *Geom_Geometry {
	pm := Geom_NewPrecisionModelWithScale(scale)
	return OperationOverlayng_OverlayNG_Overlay(geom0, geom1, OperationOverlayng_OverlayNG_UNION, pm)
}

// JtstestGeomop_TestCaseGeometryFunctions_DifferenceSR computes the difference
// using OverlayNG with a specified precision scale.
func JtstestGeomop_TestCaseGeometryFunctions_DifferenceSR(geom0, geom1 *Geom_Geometry, scale float64) *Geom_Geometry {
	pm := Geom_NewPrecisionModelWithScale(scale)
	return OperationOverlayng_OverlayNG_Overlay(geom0, geom1, OperationOverlayng_OverlayNG_DIFFERENCE, pm)
}

// JtstestGeomop_TestCaseGeometryFunctions_SymDifferenceSR computes the
// symmetric difference using OverlayNG with a specified precision scale.
func JtstestGeomop_TestCaseGeometryFunctions_SymDifferenceSR(geom0, geom1 *Geom_Geometry, scale float64) *Geom_Geometry {
	pm := Geom_NewPrecisionModelWithScale(scale)
	return OperationOverlayng_OverlayNG_Overlay(geom0, geom1, OperationOverlayng_OverlayNG_SYMDIFFERENCE, pm)
}

// JtstestGeomop_TestCaseGeometryFunctions_UnionArea computes the area of the
// union of a geometry.
func JtstestGeomop_TestCaseGeometryFunctions_UnionArea(geom *Geom_Geometry) float64 {
	return geom.UnionSelf().GetArea()
}

// JtstestGeomop_TestCaseGeometryFunctions_UnionLength computes the length of
// the union of a geometry.
func JtstestGeomop_TestCaseGeometryFunctions_UnionLength(geom *Geom_Geometry) float64 {
	return geom.UnionSelf().GetLength()
}

// JtstestGeomop_TestCaseGeometryFunctions_OverlayAreaTest tests if the overlay
// operations satisfy area identity equations.
func JtstestGeomop_TestCaseGeometryFunctions_OverlayAreaTest(a, b *Geom_Geometry) bool {
	areaDelta := jtstestGeomop_TestCaseGeometryFunctions_areaDelta(a, b)
	return areaDelta < 1e-6
}

// jtstestGeomop_TestCaseGeometryFunctions_areaDelta computes the maximum area
// delta value resulting from identity equations over the overlay operations.
// The delta value is normalized to the total area of the geometries.
// If the overlay operations are computed correctly the area delta is expected
// to be very small (e.g. < 1e-6).
func jtstestGeomop_TestCaseGeometryFunctions_areaDelta(a, b *Geom_Geometry) float64 {
	areaA := 0.0
	if a != nil {
		areaA = a.GetArea()
	}
	areaB := 0.0
	if b != nil {
		areaB = b.GetArea()
	}

	// If an input is non-polygonal delta is 0.
	if areaA == 0 || areaB == 0 {
		return 0
	}

	areaU := a.Union(b).GetArea()
	areaI := a.Intersection(b).GetArea()
	areaDab := a.Difference(b).GetArea()
	areaDba := b.Difference(a).GetArea()
	areaSD := a.SymDifference(b).GetArea()

	maxDelta := 0.0

	// & : intersection
	// - : difference
	// + : union
	// ^ : symdifference

	// A = ( A & B ) + ( A - B )
	delta := math.Abs(areaA - areaI - areaDab)
	if delta > maxDelta {
		maxDelta = delta
	}

	// B = ( A & B ) + ( B - A )
	delta = math.Abs(areaB - areaI - areaDba)
	if delta > maxDelta {
		maxDelta = delta
	}

	// ( A ^ B ) = ( A - B ) + ( B - A )
	delta = math.Abs(areaDab + areaDba - areaSD)
	if delta > maxDelta {
		maxDelta = delta
	}

	// ( A + B ) = ( A & B ) + ( A ^ B )
	delta = math.Abs(areaI + areaSD - areaU)
	if delta > maxDelta {
		maxDelta = delta
	}

	// ( A + B ) = ( A & B ) + ( A - B ) + ( A - B )
	delta = math.Abs(areaU - areaI - areaDab - areaDba)
	if delta > maxDelta {
		maxDelta = delta
	}

	// Normalize the area delta value.
	return maxDelta / (areaA + areaB)
}
