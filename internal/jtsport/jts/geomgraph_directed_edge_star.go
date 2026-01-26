package jts

import (
	"io"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const (
	geomgraph_DirectedEdgeStar_ScanningForIncoming = 1
	geomgraph_DirectedEdgeStar_LinkingToOutgoing   = 2
)

// Geomgraph_DirectedEdgeStar is an ordered list of outgoing DirectedEdges
// around a node. It supports labelling the edges as well as linking the edges
// to form both MaximalEdgeRings and MinimalEdgeRings.
type Geomgraph_DirectedEdgeStar struct {
	*Geomgraph_EdgeEndStar
	child java.Polymorphic

	// A list of all outgoing edges in the result, in CCW order.
	resultAreaEdgeList []*Geomgraph_DirectedEdge
	label              *Geomgraph_Label
}

// GetChild returns the immediate child in the type hierarchy chain.
func (des *Geomgraph_DirectedEdgeStar) GetChild() java.Polymorphic {
	return des.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (des *Geomgraph_DirectedEdgeStar) GetParent() java.Polymorphic {
	return des.Geomgraph_EdgeEndStar
}

// Geomgraph_NewDirectedEdgeStar creates a new DirectedEdgeStar.
func Geomgraph_NewDirectedEdgeStar() *Geomgraph_DirectedEdgeStar {
	ees := Geomgraph_NewEdgeEndStar()
	des := &Geomgraph_DirectedEdgeStar{
		Geomgraph_EdgeEndStar: ees,
	}
	ees.child = des
	return des
}

// Insert_BODY inserts a directed edge into the list.
func (des *Geomgraph_DirectedEdgeStar) Insert_BODY(ee *Geomgraph_EdgeEnd) {
	de := java.Cast[*Geomgraph_DirectedEdge](ee)
	des.InsertEdgeEnd(ee)
	_ = de // Used to verify cast succeeded.
}

// GetLabel returns the label for this DirectedEdgeStar.
func (des *Geomgraph_DirectedEdgeStar) GetLabel() *Geomgraph_Label {
	return des.label
}

// GetOutgoingDegree returns the number of edges in the result.
func (des *Geomgraph_DirectedEdgeStar) GetOutgoingDegree() int {
	degree := 0
	for _, ee := range des.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		if de.IsInResult() {
			degree++
		}
	}
	return degree
}

// GetOutgoingDegreeForEdgeRing returns the number of edges in the given EdgeRing.
func (des *Geomgraph_DirectedEdgeStar) GetOutgoingDegreeForEdgeRing(er *Geomgraph_EdgeRing) int {
	degree := 0
	for _, ee := range des.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		if de.GetEdgeRing() == er {
			degree++
		}
	}
	return degree
}

// GetRightmostEdge returns the rightmost edge in this star.
func (des *Geomgraph_DirectedEdgeStar) GetRightmostEdge() *Geomgraph_DirectedEdge {
	edges := des.GetEdges()
	size := len(edges)
	if size < 1 {
		return nil
	}
	de0 := java.Cast[*Geomgraph_DirectedEdge](edges[0])
	if size == 1 {
		return de0
	}
	deLast := java.Cast[*Geomgraph_DirectedEdge](edges[size-1])

	quad0 := de0.GetQuadrant()
	quad1 := deLast.GetQuadrant()
	if Geom_Quadrant_IsNorthern(quad0) && Geom_Quadrant_IsNorthern(quad1) {
		return de0
	} else if !Geom_Quadrant_IsNorthern(quad0) && !Geom_Quadrant_IsNorthern(quad1) {
		return deLast
	} else {
		// Edges are in different hemispheres - make sure we return one that is non-horizontal.
		if de0.GetDy() != 0 {
			return de0
		} else if deLast.GetDy() != 0 {
			return deLast
		}
	}
	Util_Assert_ShouldNeverReachHereWithMessage("found two horizontal edges incident on node")
	return nil
}

// ComputeLabelling_BODY computes the labelling for all dirEdges in this star,
// as well as the overall labelling.
func (des *Geomgraph_DirectedEdgeStar) ComputeLabelling_BODY(geom []*Geomgraph_GeometryGraph) {
	des.Geomgraph_EdgeEndStar.ComputeLabelling_BODY(geom)

	// Determine the overall labelling for this DirectedEdgeStar
	// (i.e. for the node it is based at).
	des.label = Geomgraph_NewLabelOn(Geom_Location_None)
	for _, ee := range des.GetEdges() {
		e := ee.GetEdge()
		eLabel := e.GetLabel()
		for i := 0; i < 2; i++ {
			eLoc := eLabel.GetLocationOn(i)
			if eLoc == Geom_Location_Interior || eLoc == Geom_Location_Boundary {
				des.label.SetLocationOn(i, Geom_Location_Interior)
			}
		}
	}
}

// MergeSymLabels merges the label from the sym dirEdge into the label for
// each dirEdge in the star.
func (des *Geomgraph_DirectedEdgeStar) MergeSymLabels() {
	for _, ee := range des.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		label := de.GetLabel()
		label.Merge(de.GetSym().GetLabel())
	}
}

// UpdateLabelling updates incomplete dirEdge labels from the labelling for
// the node.
func (des *Geomgraph_DirectedEdgeStar) UpdateLabelling(nodeLabel *Geomgraph_Label) {
	for _, ee := range des.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		label := de.GetLabel()
		label.SetAllLocationsIfNull(0, nodeLabel.GetLocationOn(0))
		label.SetAllLocationsIfNull(1, nodeLabel.GetLocationOn(1))
	}
}

func (des *Geomgraph_DirectedEdgeStar) getResultAreaEdges() []*Geomgraph_DirectedEdge {
	if des.resultAreaEdgeList != nil {
		return des.resultAreaEdgeList
	}
	des.resultAreaEdgeList = make([]*Geomgraph_DirectedEdge, 0)
	for _, ee := range des.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		if de.IsInResult() || de.GetSym().IsInResult() {
			des.resultAreaEdgeList = append(des.resultAreaEdgeList, de)
		}
	}
	return des.resultAreaEdgeList
}

// LinkResultDirectedEdges traverses the star of DirectedEdges, linking the
// included edges together. To link two dirEdges, the next pointer for an
// incoming dirEdge is set to the next outgoing edge.
//
// DirEdges are only linked if:
//   - they belong to an area (i.e. they have sides)
//   - they are marked as being in the result
//
// Edges are linked in CCW order (the order they are stored). This means that
// rings have their face on the Right (in other words, the topological location
// of the face is given by the RHS label of the DirectedEdge).
//
// PRECONDITION: No pair of dirEdges are both marked as being in the result.
func (des *Geomgraph_DirectedEdgeStar) LinkResultDirectedEdges() {
	// Make sure edges are copied to resultAreaEdges list.
	des.getResultAreaEdges()
	// Find first area edge (if any) to start linking at.
	var firstOut *Geomgraph_DirectedEdge
	var incoming *Geomgraph_DirectedEdge
	state := geomgraph_DirectedEdgeStar_ScanningForIncoming
	// Link edges in CCW order.
	for i := 0; i < len(des.resultAreaEdgeList); i++ {
		nextOut := des.resultAreaEdgeList[i]
		nextIn := nextOut.GetSym()

		// Skip de's that we're not interested in.
		if !nextOut.GetLabel().IsArea() {
			continue
		}

		// Record first outgoing edge, in order to link the last incoming edge.
		if firstOut == nil && nextOut.IsInResult() {
			firstOut = nextOut
		}

		switch state {
		case geomgraph_DirectedEdgeStar_ScanningForIncoming:
			if !nextIn.IsInResult() {
				continue
			}
			incoming = nextIn
			state = geomgraph_DirectedEdgeStar_LinkingToOutgoing
		case geomgraph_DirectedEdgeStar_LinkingToOutgoing:
			if !nextOut.IsInResult() {
				continue
			}
			incoming.SetNext(nextOut)
			state = geomgraph_DirectedEdgeStar_ScanningForIncoming
		}
	}
	if state == geomgraph_DirectedEdgeStar_LinkingToOutgoing {
		if firstOut == nil {
			panic(Geom_NewTopologyExceptionWithCoordinate("no outgoing dirEdge found", des.GetCoordinate()))
		}
		Util_Assert_IsTrueWithMessage(firstOut.IsInResult(), "unable to link last incoming dirEdge")
		incoming.SetNext(firstOut)
	}
}

// LinkMinimalDirectedEdges links MinimalEdgeRings around the star.
func (des *Geomgraph_DirectedEdgeStar) LinkMinimalDirectedEdges(er *Geomgraph_EdgeRing) {
	// Find first area edge (if any) to start linking at.
	var firstOut *Geomgraph_DirectedEdge
	var incoming *Geomgraph_DirectedEdge
	state := geomgraph_DirectedEdgeStar_ScanningForIncoming
	// Link edges in CW order.
	for i := len(des.resultAreaEdgeList) - 1; i >= 0; i-- {
		nextOut := des.resultAreaEdgeList[i]
		nextIn := nextOut.GetSym()

		// Record first outgoing edge, in order to link the last incoming edge.
		if firstOut == nil && nextOut.GetEdgeRing() == er {
			firstOut = nextOut
		}

		switch state {
		case geomgraph_DirectedEdgeStar_ScanningForIncoming:
			if nextIn.GetEdgeRing() != er {
				continue
			}
			incoming = nextIn
			state = geomgraph_DirectedEdgeStar_LinkingToOutgoing
		case geomgraph_DirectedEdgeStar_LinkingToOutgoing:
			if nextOut.GetEdgeRing() != er {
				continue
			}
			incoming.SetNextMin(nextOut)
			state = geomgraph_DirectedEdgeStar_ScanningForIncoming
		}
	}
	if state == geomgraph_DirectedEdgeStar_LinkingToOutgoing {
		Util_Assert_IsTrueWithMessage(firstOut != nil, "found null for first outgoing dirEdge")
		Util_Assert_IsTrueWithMessage(firstOut.GetEdgeRing() == er, "unable to link last incoming dirEdge")
		incoming.SetNextMin(firstOut)
	}
}

// LinkAllDirectedEdges links all DirectedEdges around the star.
func (des *Geomgraph_DirectedEdgeStar) LinkAllDirectedEdges() {
	des.GetEdges()
	// Find first area edge (if any) to start linking at.
	var prevOut *Geomgraph_DirectedEdge
	var firstIn *Geomgraph_DirectedEdge
	// Link edges in CW order.
	for i := len(des.edgeList) - 1; i >= 0; i-- {
		nextOut := java.Cast[*Geomgraph_DirectedEdge](des.edgeList[i])
		nextIn := nextOut.GetSym()
		if firstIn == nil {
			firstIn = nextIn
		}
		if prevOut != nil {
			nextIn.SetNext(prevOut)
		}
		// Record outgoing edge, in order to link the last incoming edge.
		prevOut = nextOut
	}
	firstIn.SetNext(prevOut)
}

// FindCoveredLineEdges traverses the star of edges, maintaining the current
// location in the result area at this node (if any). If any L edges are found
// in the interior of the result, mark them as covered.
func (des *Geomgraph_DirectedEdgeStar) FindCoveredLineEdges() {
	// Since edges are stored in CCW order around the node, as we move around
	// the ring we move from the right to the left side of the edge.

	// Find first DirectedEdge of result area (if any).
	// The interior of the result is on the RHS of the edge, so the start
	// location will be:
	// - INTERIOR if the edge is outgoing
	// - EXTERIOR if the edge is incoming
	startLoc := Geom_Location_None
	for _, ee := range des.GetEdges() {
		nextOut := java.Cast[*Geomgraph_DirectedEdge](ee)
		nextIn := nextOut.GetSym()
		if !nextOut.IsLineEdge() {
			if nextOut.IsInResult() {
				startLoc = Geom_Location_Interior
				break
			}
			if nextIn.IsInResult() {
				startLoc = Geom_Location_Exterior
				break
			}
		}
	}
	// No A edges found, so can't determine if L edges are covered or not.
	if startLoc == Geom_Location_None {
		return
	}

	// Move around ring, keeping track of the current location (Interior or
	// Exterior) for the result area. If L edges are found, mark them as
	// covered if they are in the interior.
	currLoc := startLoc
	for _, ee := range des.GetEdges() {
		nextOut := java.Cast[*Geomgraph_DirectedEdge](ee)
		nextIn := nextOut.GetSym()
		if nextOut.IsLineEdge() {
			nextOut.GetEdge().SetCovered(currLoc == Geom_Location_Interior)
		} else {
			// Edge is an Area edge.
			if nextOut.IsInResult() {
				currLoc = Geom_Location_Exterior
			}
			if nextIn.IsInResult() {
				currLoc = Geom_Location_Interior
			}
		}
	}
}

// ComputeDepths computes the depths from a starting DirectedEdge.
func (des *Geomgraph_DirectedEdgeStar) ComputeDepths(de *Geomgraph_DirectedEdge) {
	edgeIndex := des.FindIndex(de.Geomgraph_EdgeEnd)
	startDepth := de.GetDepth(Geom_Position_Left)
	targetLastDepth := de.GetDepth(Geom_Position_Right)
	// Compute the depths from this edge up to the end of the edge array.
	nextDepth := des.computeDepths(edgeIndex+1, len(des.edgeList), startDepth)
	// Compute the depths for the initial part of the array.
	lastDepth := des.computeDepths(0, edgeIndex, nextDepth)
	if lastDepth != targetLastDepth {
		panic(Geom_NewTopologyExceptionWithCoordinate("depth mismatch at", de.GetCoordinate()))
	}
}

// computeDepths computes the DirectedEdge depths for a subsequence of the edge array.
// Returns the last depth assigned (from the R side of the last edge visited).
func (des *Geomgraph_DirectedEdgeStar) computeDepths(startIndex, endIndex, startDepth int) int {
	currDepth := startDepth
	for i := startIndex; i < endIndex; i++ {
		nextDe := java.Cast[*Geomgraph_DirectedEdge](des.edgeList[i])
		nextDe.SetEdgeDepths(Geom_Position_Right, currDepth)
		currDepth = nextDe.GetDepth(Geom_Position_Left)
	}
	return currDepth
}

// Print writes a representation of this DirectedEdgeStar to the given writer.
func (des *Geomgraph_DirectedEdgeStar) Print(out io.Writer) {
	io.WriteString(out, "DirectedEdgeStar: "+des.GetCoordinate().String()+"\n")
	for _, ee := range des.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		io.WriteString(out, "out ")
		de.Print(out)
		io.WriteString(out, "\n")
		io.WriteString(out, "in ")
		de.GetSym().Print(out)
		io.WriteString(out, "\n")
	}
}
