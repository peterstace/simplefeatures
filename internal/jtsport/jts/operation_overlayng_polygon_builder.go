package jts

// OperationOverlayng_PolygonBuilder builds polygons from overlay edges.
type OperationOverlayng_PolygonBuilder struct {
	geometryFactory    *Geom_GeometryFactory
	shellList          []*OperationOverlayng_OverlayEdgeRing
	freeHoleList       []*OperationOverlayng_OverlayEdgeRing
	isEnforcePolygonal bool
}

// OperationOverlayng_NewPolygonBuilder creates a new PolygonBuilder.
func OperationOverlayng_NewPolygonBuilder(resultAreaEdges []*OperationOverlayng_OverlayEdge, geomFact *Geom_GeometryFactory) *OperationOverlayng_PolygonBuilder {
	return OperationOverlayng_NewPolygonBuilderWithEnforcePolygonal(resultAreaEdges, geomFact, true)
}

// OperationOverlayng_NewPolygonBuilderWithEnforcePolygonal creates a new
// PolygonBuilder with control over polygonal enforcement.
func OperationOverlayng_NewPolygonBuilderWithEnforcePolygonal(resultAreaEdges []*OperationOverlayng_OverlayEdge, geomFact *Geom_GeometryFactory, isEnforcePolygonal bool) *OperationOverlayng_PolygonBuilder {
	pb := &OperationOverlayng_PolygonBuilder{
		geometryFactory:    geomFact,
		shellList:          make([]*OperationOverlayng_OverlayEdgeRing, 0),
		freeHoleList:       make([]*OperationOverlayng_OverlayEdgeRing, 0),
		isEnforcePolygonal: isEnforcePolygonal,
	}
	pb.buildRings(resultAreaEdges)
	return pb
}

// GetPolygons returns the polygons built from the overlay edges.
func (pb *OperationOverlayng_PolygonBuilder) GetPolygons() []*Geom_Polygon {
	return pb.computePolygons(pb.shellList)
}

// GetShellRings returns the shell rings.
func (pb *OperationOverlayng_PolygonBuilder) GetShellRings() []*OperationOverlayng_OverlayEdgeRing {
	return pb.shellList
}

func (pb *OperationOverlayng_PolygonBuilder) computePolygons(shellList []*OperationOverlayng_OverlayEdgeRing) []*Geom_Polygon {
	resultPolyList := make([]*Geom_Polygon, 0, len(shellList))
	for _, er := range shellList {
		poly := er.ToPolygon(pb.geometryFactory)
		resultPolyList = append(resultPolyList, poly)
	}
	return resultPolyList
}

func (pb *OperationOverlayng_PolygonBuilder) buildRings(resultAreaEdges []*OperationOverlayng_OverlayEdge) {
	pb.linkResultAreaEdgesMax(resultAreaEdges)
	maxRings := operationOverlayng_PolygonBuilder_buildMaximalRings(resultAreaEdges)
	pb.buildMinimalRings(maxRings)
	pb.placeFreeHoles(pb.shellList, pb.freeHoleList)
}

func (pb *OperationOverlayng_PolygonBuilder) linkResultAreaEdgesMax(resultEdges []*OperationOverlayng_OverlayEdge) {
	for _, edge := range resultEdges {
		OperationOverlayng_MaximalEdgeRing_LinkResultAreaMaxRingAtNode(edge)
	}
}

func operationOverlayng_PolygonBuilder_buildMaximalRings(edges []*OperationOverlayng_OverlayEdge) []*OperationOverlayng_MaximalEdgeRing {
	edgeRings := make([]*OperationOverlayng_MaximalEdgeRing, 0)
	for _, e := range edges {
		if e.IsInResultArea() && e.GetLabel().IsBoundaryEither() {
			if e.GetEdgeRingMax() == nil {
				er := OperationOverlayng_NewMaximalEdgeRing(e)
				edgeRings = append(edgeRings, er)
			}
		}
	}
	return edgeRings
}

func (pb *OperationOverlayng_PolygonBuilder) buildMinimalRings(maxRings []*OperationOverlayng_MaximalEdgeRing) {
	for _, erMax := range maxRings {
		minRings := erMax.BuildMinimalRings(pb.geometryFactory)
		pb.assignShellsAndHoles(minRings)
	}
}

func (pb *OperationOverlayng_PolygonBuilder) assignShellsAndHoles(minRings []*OperationOverlayng_OverlayEdgeRing) {
	// Two situations may occur:
	// - the rings are a shell and some holes
	// - rings are a set of holes
	// This code identifies the situation and places the rings appropriately.
	shell := pb.findSingleShell(minRings)
	if shell != nil {
		operationOverlayng_PolygonBuilder_assignHoles(shell, minRings)
		pb.shellList = append(pb.shellList, shell)
	} else {
		// All rings are holes; their shell will be found later.
		pb.freeHoleList = append(pb.freeHoleList, minRings...)
	}
}

// findSingleShell finds the single shell, if any, out of a list of minimal
// rings derived from a maximal ring. The other possibility is that they are a
// set of (connected) holes, in which case no shell will be found.
func (pb *OperationOverlayng_PolygonBuilder) findSingleShell(edgeRings []*OperationOverlayng_OverlayEdgeRing) *OperationOverlayng_OverlayEdgeRing {
	shellCount := 0
	var shell *OperationOverlayng_OverlayEdgeRing
	for _, er := range edgeRings {
		if !er.IsHole() {
			shell = er
			shellCount++
		}
	}
	Util_Assert_IsTrueWithMessage(shellCount <= 1, "found two shells in EdgeRing list")
	return shell
}

// assignHoles assigns holes to a shell. For the set of minimal rings
// comprising a maximal ring, this assigns the holes to the shell known to
// contain them.
func operationOverlayng_PolygonBuilder_assignHoles(shell *OperationOverlayng_OverlayEdgeRing, edgeRings []*OperationOverlayng_OverlayEdgeRing) {
	for _, er := range edgeRings {
		if er.IsHole() {
			er.SetShell(shell)
		}
	}
}

// placeFreeHoles places holes that have not yet been assigned to a shell.
// These "free" holes should all be properly contained in their parent shells,
// so it is safe to use the findEdgeRingContaining method.
func (pb *OperationOverlayng_PolygonBuilder) placeFreeHoles(shellList []*OperationOverlayng_OverlayEdgeRing, freeHoleList []*OperationOverlayng_OverlayEdgeRing) {
	for _, hole := range freeHoleList {
		// Only place this hole if it doesn't yet have a shell.
		if hole.GetShell() == nil {
			shell := hole.FindEdgeRingContaining(shellList)
			// Only when building a polygon-valid result.
			if pb.isEnforcePolygonal && shell == nil {
				panic(Geom_NewTopologyExceptionWithCoordinate("unable to assign free hole to a shell", hole.GetCoordinate()))
			}
			hole.SetShell(shell)
		}
	}
}
