package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelateng_RelateNG computes the value of topological predicates
// between two geometries based on the Dimensionally-Extended 9-Intersection
// Model (DE-9IM).
//
// The RelateNG algorithm has the following capabilities:
//   - Efficient short-circuited evaluation of topological predicates
//     (including matching custom DE-9IM matrix patterns)
//   - Optimized repeated evaluation of predicates against a single geometry
//     via cached spatial indexes (AKA "prepared mode")
//   - Robust computation (only point-local topology is required,
//     so invalid geometry topology does not cause failures)
//   - GeometryCollection inputs containing mixed types and overlapping
//     polygons are supported, using union semantics.
//   - Zero-length LineStrings are treated as being topologically identical to
//     Points.
//   - Support for BoundaryNodeRules.
//
// RelateNG operates in 2D only; it ignores any Z ordinates.
type OperationRelateng_RelateNG struct {
	boundaryNodeRule Algorithm_BoundaryNodeRule
	geomA            *OperationRelateng_RelateGeometry
	edgeMutualInt    *Noding_MCIndexSegmentSetMutualIntersector
}

// OperationRelateng_RelateNG_Relate tests whether the topological relationship
// between two geometries satisfies a topological predicate.
func OperationRelateng_RelateNG_Relate(a, b *Geom_Geometry, pred OperationRelateng_TopologyPredicate) bool {
	rng := operationRelateng_newRelateNG(a, false)
	return rng.EvaluatePredicate(b, pred)
}

// OperationRelateng_RelateNG_RelateWithRule tests whether the topological
// relationship between two geometries satisfies a topological predicate,
// using a given BoundaryNodeRule.
func OperationRelateng_RelateNG_RelateWithRule(a, b *Geom_Geometry, pred OperationRelateng_TopologyPredicate, bnRule Algorithm_BoundaryNodeRule) bool {
	rng := operationRelateng_newRelateNGWithRule(a, false, bnRule)
	return rng.EvaluatePredicate(b, pred)
}

// OperationRelateng_RelateNG_RelatePattern tests whether the topological
// relationship to a geometry matches a DE-9IM matrix pattern.
func OperationRelateng_RelateNG_RelatePattern(a, b *Geom_Geometry, imPattern string) bool {
	rng := operationRelateng_newRelateNG(a, false)
	return rng.EvaluatePattern(b, imPattern)
}

// OperationRelateng_RelateNG_RelateMatrix computes the DE-9IM matrix for the
// topological relationship between two geometries.
func OperationRelateng_RelateNG_RelateMatrix(a, b *Geom_Geometry) *Geom_IntersectionMatrix {
	rng := operationRelateng_newRelateNG(a, false)
	return rng.Evaluate(b)
}

// OperationRelateng_RelateNG_RelateMatrixWithRule computes the DE-9IM matrix for
// the topological relationship between two geometries.
func OperationRelateng_RelateNG_RelateMatrixWithRule(a, b *Geom_Geometry, bnRule Algorithm_BoundaryNodeRule) *Geom_IntersectionMatrix {
	rng := operationRelateng_newRelateNGWithRule(a, false, bnRule)
	return rng.Evaluate(b)
}

// OperationRelateng_RelateNG_Prepare creates a prepared RelateNG instance to
// optimize the evaluation of relationships against a single geometry.
func OperationRelateng_RelateNG_Prepare(a *Geom_Geometry) *OperationRelateng_RelateNG {
	return operationRelateng_newRelateNG(a, true)
}

// OperationRelateng_RelateNG_PrepareWithRule creates a prepared RelateNG
// instance to optimize the computation of predicates against a single geometry,
// using a given BoundaryNodeRule.
func OperationRelateng_RelateNG_PrepareWithRule(a *Geom_Geometry, bnRule Algorithm_BoundaryNodeRule) *OperationRelateng_RelateNG {
	return operationRelateng_newRelateNGWithRule(a, true, bnRule)
}

func operationRelateng_newRelateNG(inputA *Geom_Geometry, isPrepared bool) *OperationRelateng_RelateNG {
	return operationRelateng_newRelateNGWithRule(inputA, isPrepared, Algorithm_BoundaryNodeRule_OGC_SFS_BOUNDARY_RULE)
}

func operationRelateng_newRelateNGWithRule(inputA *Geom_Geometry, isPrepared bool, bnRule Algorithm_BoundaryNodeRule) *OperationRelateng_RelateNG {
	return &OperationRelateng_RelateNG{
		boundaryNodeRule: bnRule,
		geomA:            OperationRelateng_NewRelateGeometryWithOptions(inputA, isPrepared, bnRule),
	}
}

// Evaluate computes the DE-9IM matrix for the topological relationship to a
// geometry.
func (rng *OperationRelateng_RelateNG) Evaluate(b *Geom_Geometry) *Geom_IntersectionMatrix {
	rel := OperationRelateng_NewRelateMatrixPredicate()
	rng.EvaluatePredicate(b, rel)
	return rel.GetIM()
}

// EvaluatePattern tests whether the topological relationship to a geometry
// matches a DE-9IM matrix pattern.
func (rng *OperationRelateng_RelateNG) EvaluatePattern(b *Geom_Geometry, imPattern string) bool {
	return rng.EvaluatePredicate(b, OperationRelateng_RelatePredicate_Matches(imPattern))
}

// EvaluatePredicate tests whether the topological relationship to a geometry
// satisfies a topology predicate.
func (rng *OperationRelateng_RelateNG) EvaluatePredicate(b *Geom_Geometry, predicate OperationRelateng_TopologyPredicate) bool {
	// Fast envelope checks.
	if !rng.hasRequiredEnvelopeInteraction(b, predicate) {
		return false
	}

	geomB := OperationRelateng_NewRelateGeometryWithRule(b, rng.boundaryNodeRule)

	if rng.geomA.IsEmpty() && geomB.IsEmpty() {
		return rng.finishValue(predicate)
	}
	dimA := rng.geomA.GetDimensionReal()
	dimB := geomB.GetDimensionReal()

	// Check if predicate is determined by dimension or envelope.
	predicate.InitDim(dimA, dimB)
	if predicate.IsKnown() {
		return rng.finishValue(predicate)
	}

	predicate.InitEnv(rng.geomA.GetEnvelope(), geomB.GetEnvelope())
	if predicate.IsKnown() {
		return rng.finishValue(predicate)
	}

	topoComputer := OperationRelateng_NewTopologyComputer(predicate, rng.geomA, geomB)

	// Optimized P/P evaluation.
	if dimA == Geom_Dimension_P && dimB == Geom_Dimension_P {
		rng.computePP(geomB, topoComputer)
		topoComputer.Finish()
		return topoComputer.GetResult()
	}

	// Test points against (potentially) indexed geometry first.
	rng.computeAtPoints(geomB, OperationRelateng_RelateGeometry_GEOM_B, rng.geomA, topoComputer)
	if topoComputer.IsResultKnown() {
		return topoComputer.GetResult()
	}
	rng.computeAtPoints(rng.geomA, OperationRelateng_RelateGeometry_GEOM_A, geomB, topoComputer)
	if topoComputer.IsResultKnown() {
		return topoComputer.GetResult()
	}

	if rng.geomA.HasEdges() && geomB.HasEdges() {
		rng.computeAtEdges(geomB, topoComputer)
	}

	// After all processing, set remaining unknown values in IM.
	topoComputer.Finish()
	return topoComputer.GetResult()
}

func (rng *OperationRelateng_RelateNG) hasRequiredEnvelopeInteraction(b *Geom_Geometry, predicate OperationRelateng_TopologyPredicate) bool {
	envB := b.GetEnvelopeInternal()
	isInteracts := false
	if predicate.RequireCovers(OperationRelateng_RelateGeometry_GEOM_A) {
		if !rng.geomA.GetEnvelope().CoversEnvelope(envB) {
			return false
		}
		isInteracts = true
	} else if predicate.RequireCovers(OperationRelateng_RelateGeometry_GEOM_B) {
		if !envB.CoversEnvelope(rng.geomA.GetEnvelope()) {
			return false
		}
		isInteracts = true
	}
	if !isInteracts &&
		predicate.RequireInteraction() &&
		!rng.geomA.GetEnvelope().IntersectsEnvelope(envB) {
		return false
	}
	return true
}

func (rng *OperationRelateng_RelateNG) finishValue(predicate OperationRelateng_TopologyPredicate) bool {
	predicate.Finish()
	return predicate.Value()
}

// computePP is an optimized algorithm for evaluating P/P cases. It tests one
// point set against the other.
func (rng *OperationRelateng_RelateNG) computePP(geomB *OperationRelateng_RelateGeometry, topoComputer *OperationRelateng_TopologyComputer) {
	ptsA := rng.geomA.GetUniquePoints()
	ptsB := geomB.GetUniquePoints()

	numBinA := 0
	for ptB := range ptsB {
		if ptsA[ptB] {
			numBinA++
			topoComputer.AddPointOnPointInterior(Geom_NewCoordinateWithXY(ptB.x, ptB.y))
		} else {
			topoComputer.AddPointOnPointExterior(OperationRelateng_RelateGeometry_GEOM_B, Geom_NewCoordinateWithXY(ptB.x, ptB.y))
		}
		if topoComputer.IsResultKnown() {
			return
		}
	}
	// If number of matched B points is less than size of A, there must be at
	// least one A point in the exterior of B.
	if numBinA < len(ptsA) {
		topoComputer.AddPointOnPointExterior(OperationRelateng_RelateGeometry_GEOM_A, nil)
	}
}

func (rng *OperationRelateng_RelateNG) computeAtPoints(geom *OperationRelateng_RelateGeometry, isA bool,
	geomTarget *OperationRelateng_RelateGeometry, topoComputer *OperationRelateng_TopologyComputer) {

	isResultKnown := rng.computePoints(geom, isA, geomTarget, topoComputer)
	if isResultKnown {
		return
	}

	// Performance optimization: only check points against target if it has
	// areas OR if the predicate requires checking for exterior interaction.
	// In particular, this avoids testing line ends against lines for the
	// intersects predicate (since these are checked during segment/segment
	// intersection checking anyway). Checking points against areas is
	// necessary, since the input linework is disjoint if one input lies wholly
	// inside an area, so segment intersection checking is not sufficient.
	checkDisjointPoints := geomTarget.HasDimension(Geom_Dimension_A) ||
		topoComputer.IsExteriorCheckRequired(isA)
	if !checkDisjointPoints {
		return
	}

	isResultKnown = rng.computeLineEnds(geom, isA, geomTarget, topoComputer)
	if isResultKnown {
		return
	}

	rng.computeAreaVertex(geom, isA, geomTarget, topoComputer)
}

func (rng *OperationRelateng_RelateNG) computePoints(geom *OperationRelateng_RelateGeometry, isA bool, geomTarget *OperationRelateng_RelateGeometry,
	topoComputer *OperationRelateng_TopologyComputer) bool {
	if !geom.HasDimension(Geom_Dimension_P) {
		return false
	}

	points := geom.GetEffectivePoints()
	for _, point := range points {
		if point.IsEmpty() {
			continue
		}

		pt := point.GetCoordinate()
		rng.computePoint(isA, pt, geomTarget, topoComputer)
		if topoComputer.IsResultKnown() {
			return true
		}
	}
	return false
}

func (rng *OperationRelateng_RelateNG) computePoint(isA bool, pt *Geom_Coordinate, geomTarget *OperationRelateng_RelateGeometry, topoComputer *OperationRelateng_TopologyComputer) {
	locDimTarget := geomTarget.LocateWithDim(pt)
	locTarget := OperationRelateng_DimensionLocation_Location(locDimTarget)
	dimTarget := OperationRelateng_DimensionLocation_DimensionWithExterior(locDimTarget, topoComputer.GetDimension(!isA))
	topoComputer.AddPointOnGeometry(isA, locTarget, dimTarget, pt)
}

func (rng *OperationRelateng_RelateNG) computeLineEnds(geom *OperationRelateng_RelateGeometry, isA bool, geomTarget *OperationRelateng_RelateGeometry,
	topoComputer *OperationRelateng_TopologyComputer) bool {
	if !geom.HasDimension(Geom_Dimension_L) {
		return false
	}

	hasExteriorIntersection := false
	geomi := Geom_NewGeometryCollectionIterator(geom.GetGeometry())
	for geomi.HasNext() {
		elem := geomi.Next()
		if elem.IsEmpty() {
			continue
		}

		if java.InstanceOf[*Geom_LineString](elem) {
			line := java.Cast[*Geom_LineString](elem)
			// Once an intersection with target exterior is recorded, skip
			// further known-exterior points.
			if hasExteriorIntersection &&
				elem.GetEnvelopeInternal().Disjoint(geomTarget.GetEnvelope()) {
				continue
			}
			e0 := line.GetCoordinateN(0)
			var hasExt bool
			hasExt = rng.computeLineEnd(geom, isA, e0, geomTarget, topoComputer)
			hasExteriorIntersection = hasExteriorIntersection || hasExt
			if topoComputer.IsResultKnown() {
				return true
			}

			if !line.IsClosed() {
				e1 := line.GetCoordinateN(line.GetNumPoints() - 1)
				hasExt = rng.computeLineEnd(geom, isA, e1, geomTarget, topoComputer)
				hasExteriorIntersection = hasExteriorIntersection || hasExt
				if topoComputer.IsResultKnown() {
					return true
				}
			}
		}
	}
	return false
}

// computeLineEnd computes the topology of a line endpoint. Also reports if
// the line end is in the exterior of the target geometry, to optimize testing
// multiple exterior endpoints.
func (rng *OperationRelateng_RelateNG) computeLineEnd(geom *OperationRelateng_RelateGeometry, isA bool, pt *Geom_Coordinate,
	geomTarget *OperationRelateng_RelateGeometry, topoComputer *OperationRelateng_TopologyComputer) bool {
	locDimLineEnd := geom.LocateLineEndWithDim(pt)
	dimLineEnd := OperationRelateng_DimensionLocation_DimensionWithExterior(locDimLineEnd, topoComputer.GetDimension(isA))
	// Skip line ends which are in a GC area.
	if dimLineEnd != Geom_Dimension_L {
		return false
	}
	locLineEnd := OperationRelateng_DimensionLocation_Location(locDimLineEnd)

	locDimTarget := geomTarget.LocateWithDim(pt)
	locTarget := OperationRelateng_DimensionLocation_Location(locDimTarget)
	dimTarget := OperationRelateng_DimensionLocation_DimensionWithExterior(locDimTarget, topoComputer.GetDimension(!isA))
	topoComputer.AddLineEndOnGeometry(isA, locLineEnd, locTarget, dimTarget, pt)
	return locTarget == Geom_Location_Exterior
}

func (rng *OperationRelateng_RelateNG) computeAreaVertex(geom *OperationRelateng_RelateGeometry, isA bool, geomTarget *OperationRelateng_RelateGeometry, topoComputer *OperationRelateng_TopologyComputer) bool {
	if !geom.HasDimension(Geom_Dimension_A) {
		return false
	}
	// Evaluate for line and area targets only, since points are handled in the
	// reverse direction.
	if geomTarget.GetDimension() < Geom_Dimension_L {
		return false
	}

	hasExteriorIntersection := false
	geomi := Geom_NewGeometryCollectionIterator(geom.GetGeometry())
	for geomi.HasNext() {
		elem := geomi.Next()
		if elem.IsEmpty() {
			continue
		}

		if java.InstanceOf[*Geom_Polygon](elem) {
			poly := java.Cast[*Geom_Polygon](elem)
			// Once an intersection with target exterior is recorded, skip
			// further known-exterior points.
			if hasExteriorIntersection &&
				elem.GetEnvelopeInternal().Disjoint(geomTarget.GetEnvelope()) {
				continue
			}
			var hasExt bool
			hasExt = rng.computeAreaVertexRing(geom, isA, poly.GetExteriorRing(), geomTarget, topoComputer)
			hasExteriorIntersection = hasExteriorIntersection || hasExt
			if topoComputer.IsResultKnown() {
				return true
			}
			for j := 0; j < poly.GetNumInteriorRing(); j++ {
				hasExt = rng.computeAreaVertexRing(geom, isA, poly.GetInteriorRingN(j), geomTarget, topoComputer)
				hasExteriorIntersection = hasExteriorIntersection || hasExt
				if topoComputer.IsResultKnown() {
					return true
				}
			}
		}
	}
	return false
}

func (rng *OperationRelateng_RelateNG) computeAreaVertexRing(geom *OperationRelateng_RelateGeometry, isA bool, ring *Geom_LinearRing, geomTarget *OperationRelateng_RelateGeometry, topoComputer *OperationRelateng_TopologyComputer) bool {
	pt := ring.GetCoordinate()

	locArea := geom.LocateAreaVertex(pt)
	locDimTarget := geomTarget.LocateWithDim(pt)
	locTarget := OperationRelateng_DimensionLocation_Location(locDimTarget)
	dimTarget := OperationRelateng_DimensionLocation_DimensionWithExterior(locDimTarget, topoComputer.GetDimension(!isA))
	topoComputer.AddAreaVertex(isA, locArea, locTarget, dimTarget, pt)
	return locTarget == Geom_Location_Exterior
}

func (rng *OperationRelateng_RelateNG) computeAtEdges(geomB *OperationRelateng_RelateGeometry, topoComputer *OperationRelateng_TopologyComputer) {
	envInt := rng.geomA.GetEnvelope().Intersection(geomB.GetEnvelope())
	if envInt.IsNull() {
		return
	}

	edgesB := geomB.ExtractSegmentStrings(OperationRelateng_RelateGeometry_GEOM_B, envInt)
	intersector := OperationRelateng_NewEdgeSegmentIntersector(topoComputer)

	if topoComputer.IsSelfNodingRequired() {
		rng.computeEdgesAll(edgesB, envInt, intersector)
	} else {
		rng.computeEdgesMutual(edgesB, envInt, intersector)
	}
	if topoComputer.IsResultKnown() {
		return
	}

	topoComputer.EvaluateNodes()
}

func (rng *OperationRelateng_RelateNG) computeEdgesAll(edgesB []*OperationRelateng_RelateSegmentString, envInt *Geom_Envelope, intersector *OperationRelateng_EdgeSegmentIntersector) {
	edgesA := rng.geomA.ExtractSegmentStrings(OperationRelateng_RelateGeometry_GEOM_A, envInt)

	edgeInt := OperationRelateng_NewEdgeSetIntersector(edgesA, edgesB, envInt)
	edgeInt.Process(intersector)
}

func (rng *OperationRelateng_RelateNG) computeEdgesMutual(edgesB []*OperationRelateng_RelateSegmentString, envInt *Geom_Envelope, intersector *OperationRelateng_EdgeSegmentIntersector) {
	// In prepared mode the A edge index is reused.
	if rng.edgeMutualInt == nil {
		var envExtract *Geom_Envelope
		if !rng.geomA.IsPrepared() {
			envExtract = envInt
		}
		edgesA := rng.geomA.ExtractSegmentStrings(OperationRelateng_RelateGeometry_GEOM_A, envExtract)
		rng.edgeMutualInt = Noding_NewMCIndexSegmentSetMutualIntersectorWithEnvelope(rng.toSegmentStrings(edgesA), envExtract)
	}

	rng.edgeMutualInt.Process(rng.toSegmentStrings(edgesB), intersector)
}

func (rng *OperationRelateng_RelateNG) toSegmentStrings(edges []*OperationRelateng_RelateSegmentString) []Noding_SegmentString {
	result := make([]Noding_SegmentString, len(edges))
	for i, e := range edges {
		result[i] = e
	}
	return result
}
