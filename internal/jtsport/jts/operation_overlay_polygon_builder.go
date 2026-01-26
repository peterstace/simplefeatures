package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationOverlay_PolygonBuilder forms Polygons out of a graph of
// DirectedEdges. The edges to use are marked as being in the result Area.
type OperationOverlay_PolygonBuilder struct {
	child java.Polymorphic

	geometryFactory *Geom_GeometryFactory
	shellList       []*Geomgraph_EdgeRing
}

// GetChild returns the immediate child in the type hierarchy chain.
func (pb *OperationOverlay_PolygonBuilder) GetChild() java.Polymorphic {
	return pb.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (pb *OperationOverlay_PolygonBuilder) GetParent() java.Polymorphic {
	return nil
}

// OperationOverlay_NewPolygonBuilder creates a new PolygonBuilder.
func OperationOverlay_NewPolygonBuilder(geometryFactory *Geom_GeometryFactory) *OperationOverlay_PolygonBuilder {
	return &OperationOverlay_PolygonBuilder{
		geometryFactory: geometryFactory,
	}
}

// AddFromGraph adds a complete graph. The graph is assumed to contain one or
// more polygons, possibly with holes.
func (pb *OperationOverlay_PolygonBuilder) AddFromGraph(graph *Geomgraph_PlanarGraph) {
	pb.Add(graph.GetEdgeEnds(), graph.GetNodes())
}

// Add adds a set of edges and nodes, which form a graph. The graph is assumed
// to contain one or more polygons, possibly with holes.
func (pb *OperationOverlay_PolygonBuilder) Add(dirEdges []*Geomgraph_EdgeEnd, nodes []*Geomgraph_Node) {
	Geomgraph_PlanarGraph_LinkResultDirectedEdges(nodes)
	maxEdgeRings := pb.buildMaximalEdgeRings(dirEdges)
	var freeHoleList []*Geomgraph_EdgeRing
	edgeRings := pb.buildMinimalEdgeRings(maxEdgeRings, &pb.shellList, &freeHoleList)
	pb.sortShellsAndHoles(edgeRings, &pb.shellList, &freeHoleList)
	pb.placeFreeHoles(pb.shellList, freeHoleList)
}

// GetPolygons returns the list of polygons built.
func (pb *OperationOverlay_PolygonBuilder) GetPolygons() []*Geom_Polygon {
	return pb.computePolygons(pb.shellList)
}

// buildMaximalEdgeRings builds MaximalEdgeRings for all DirectedEdges in
// result.
func (pb *OperationOverlay_PolygonBuilder) buildMaximalEdgeRings(dirEdges []*Geomgraph_EdgeEnd) []*OperationOverlay_MaximalEdgeRing {
	var maxEdgeRings []*OperationOverlay_MaximalEdgeRing
	for _, ee := range dirEdges {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		if de.IsInResult() && de.GetLabel().IsArea() {
			// If this edge has not yet been processed.
			if de.GetEdgeRing() == nil {
				er := OperationOverlay_NewMaximalEdgeRing(de, pb.geometryFactory)
				maxEdgeRings = append(maxEdgeRings, er)
				er.SetInResult()
			}
		}
	}
	return maxEdgeRings
}

func (pb *OperationOverlay_PolygonBuilder) buildMinimalEdgeRings(maxEdgeRings []*OperationOverlay_MaximalEdgeRing, shellList *[]*Geomgraph_EdgeRing, freeHoleList *[]*Geomgraph_EdgeRing) []*Geomgraph_EdgeRing {
	var edgeRings []*Geomgraph_EdgeRing
	for _, er := range maxEdgeRings {
		if er.GetMaxNodeDegree() > 2 {
			er.LinkDirectedEdgesForMinimalEdgeRings()
			minEdgeRings := er.BuildMinimalRings()
			// At this point we can go ahead and attempt to place holes, if
			// this EdgeRing is a polygon.
			shell := pb.findShell(minEdgeRings)
			if shell != nil {
				pb.placePolygonHoles(shell, minEdgeRings)
				*shellList = append(*shellList, shell)
			} else {
				*freeHoleList = append(*freeHoleList, minEdgeRings...)
			}
		} else {
			edgeRings = append(edgeRings, er.Geomgraph_EdgeRing)
		}
	}
	return edgeRings
}

// findShell takes a list of MinimalEdgeRings derived from a MaximalEdgeRing,
// and tests whether they form a Polygon. This is the case if there is a single
// shell in the list. In this case the shell is returned. The other possibility
// is that they are a series of connected holes, in which case no shell is
// returned.
func (pb *OperationOverlay_PolygonBuilder) findShell(minEdgeRings []*Geomgraph_EdgeRing) *Geomgraph_EdgeRing {
	shellCount := 0
	var shell *Geomgraph_EdgeRing
	for _, er := range minEdgeRings {
		if !er.IsHole() {
			shell = er
			shellCount++
		}
	}
	Util_Assert_IsTrueWithMessage(shellCount <= 1, "found two shells in MinimalEdgeRing list")
	return shell
}

// placePolygonHoles assigns the holes for a Polygon (formed from a list of
// MinimalEdgeRings) to its shell.
func (pb *OperationOverlay_PolygonBuilder) placePolygonHoles(shell *Geomgraph_EdgeRing, minEdgeRings []*Geomgraph_EdgeRing) {
	for _, er := range minEdgeRings {
		if er.IsHole() {
			er.SetShell(shell)
		}
	}
}

// sortShellsAndHoles determines for all rings in the input list whether the
// ring is a shell or a hole and adds it to the appropriate list. Due to the
// way the DirectedEdges were linked, a ring is a shell if it is oriented CW, a
// hole otherwise.
func (pb *OperationOverlay_PolygonBuilder) sortShellsAndHoles(edgeRings []*Geomgraph_EdgeRing, shellList *[]*Geomgraph_EdgeRing, freeHoleList *[]*Geomgraph_EdgeRing) {
	for _, er := range edgeRings {
		if er.IsHole() {
			*freeHoleList = append(*freeHoleList, er)
		} else {
			*shellList = append(*shellList, er)
		}
	}
}

// placeFreeHoles finds a containing shell for all holes which have not yet
// been assigned to a shell. These "free" holes should all be properly
// contained in their parent shells, so it is safe to use the
// findEdgeRingContaining method.
func (pb *OperationOverlay_PolygonBuilder) placeFreeHoles(shellList []*Geomgraph_EdgeRing, freeHoleList []*Geomgraph_EdgeRing) {
	for _, hole := range freeHoleList {
		// Only place this hole if it doesn't yet have a shell.
		if hole.GetShell() == nil {
			shell := OperationOverlay_PolygonBuilder_FindEdgeRingContaining(hole, shellList)
			if shell == nil {
				panic(Geom_NewTopologyExceptionWithCoordinate("unable to assign hole to a shell", hole.GetCoordinate(0)))
			}
			hole.SetShell(shell)
		}
	}
}

// OperationOverlay_PolygonBuilder_FindEdgeRingContaining finds the innermost
// enclosing shell EdgeRing containing the argument EdgeRing, if any. The
// innermost enclosing ring is the smallest enclosing ring. The algorithm used
// depends on the fact that ring A contains ring B if envelope(ring A) contains
// envelope(ring B). This routine is only safe to use if the chosen point of
// the hole is known to be properly contained in a shell (which is guaranteed
// to be the case if the hole does not touch its shell).
func OperationOverlay_PolygonBuilder_FindEdgeRingContaining(testEr *Geomgraph_EdgeRing, shellList []*Geomgraph_EdgeRing) *Geomgraph_EdgeRing {
	testRing := testEr.GetLinearRing()
	testEnv := testRing.GetEnvelopeInternal()
	testPt := testRing.GetCoordinateN(0)

	var minShell *Geomgraph_EdgeRing
	var minShellEnv *Geom_Envelope
	for _, tryShell := range shellList {
		tryShellRing := tryShell.GetLinearRing()
		tryShellEnv := tryShellRing.GetEnvelopeInternal()
		// The hole envelope cannot equal the shell envelope.
		// (Also guards against testing rings against themselves.)
		if tryShellEnv.Equals(testEnv) {
			continue
		}
		// Hole must be contained in shell.
		if !tryShellEnv.ContainsEnvelope(testEnv) {
			continue
		}

		testPt = Geom_CoordinateArrays_PtNotInList(testRing.GetCoordinates(), tryShellRing.GetCoordinates())
		isContained := false
		if Algorithm_PointLocation_IsInRing(testPt, tryShellRing.GetCoordinates()) {
			isContained = true
		}

		// Check if this new containing ring is smaller than the current
		// minimum ring.
		if isContained {
			if minShell == nil || minShellEnv.ContainsEnvelope(tryShellEnv) {
				minShell = tryShell
				minShellEnv = minShell.GetLinearRing().GetEnvelopeInternal()
			}
		}
	}
	return minShell
}

func (pb *OperationOverlay_PolygonBuilder) computePolygons(shellList []*Geomgraph_EdgeRing) []*Geom_Polygon {
	var resultPolyList []*Geom_Polygon
	// Add Polygons for all shells.
	for _, er := range shellList {
		poly := er.ToPolygon(pb.geometryFactory)
		resultPolyList = append(resultPolyList, poly)
	}
	return resultPolyList
}
