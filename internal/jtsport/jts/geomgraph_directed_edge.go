package jts

import (
	"io"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_DirectedEdge_DepthFactor computes the factor for the change in
// depth when moving from one location to another. E.g. if crossing from the
// EXTERIOR to the INTERIOR the depth increases, so the factor is 1.
func Geomgraph_DirectedEdge_DepthFactor(currLocation, nextLocation int) int {
	if currLocation == Geom_Location_Exterior && nextLocation == Geom_Location_Interior {
		return 1
	} else if currLocation == Geom_Location_Interior && nextLocation == Geom_Location_Exterior {
		return -1
	}
	return 0
}

// Geomgraph_DirectedEdge represents a directed edge in a topology graph.
type Geomgraph_DirectedEdge struct {
	*Geomgraph_EdgeEnd
	child java.Polymorphic

	isForward  bool
	isInResult bool
	isVisited  bool

	sym         *Geomgraph_DirectedEdge // The symmetric edge.
	next        *Geomgraph_DirectedEdge // The next edge in the edge ring for the polygon containing this edge.
	nextMin     *Geomgraph_DirectedEdge // The next edge in the MinimalEdgeRing that contains this edge.
	edgeRing    *Geomgraph_EdgeRing     // The EdgeRing that this edge is part of.
	minEdgeRing *Geomgraph_EdgeRing     // The MinimalEdgeRing that this edge is part of.

	// The depth of each side (position) of this edge.
	// The 0 element of the array is never used.
	depth [3]int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (de *Geomgraph_DirectedEdge) GetChild() java.Polymorphic {
	return de.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (de *Geomgraph_DirectedEdge) GetParent() java.Polymorphic {
	return de.Geomgraph_EdgeEnd
}

// Geomgraph_NewDirectedEdge creates a new DirectedEdge from an edge and direction.
func Geomgraph_NewDirectedEdge(edge *Geomgraph_Edge, isForward bool) *Geomgraph_DirectedEdge {
	ee := Geomgraph_NewEdgeEndFromEdge(edge)
	de := &Geomgraph_DirectedEdge{
		Geomgraph_EdgeEnd: ee,
		isForward:        isForward,
		depth:            [3]int{0, -999, -999},
	}
	ee.child = de
	if isForward {
		de.init(edge.GetCoordinateAtIndex(0), edge.GetCoordinateAtIndex(1))
	} else {
		n := edge.GetNumPoints() - 1
		de.init(edge.GetCoordinateAtIndex(n), edge.GetCoordinateAtIndex(n-1))
	}
	de.computeDirectedLabel()
	return de
}

// GetEdge returns the parent edge of this directed edge.
func (de *Geomgraph_DirectedEdge) GetEdge() *Geomgraph_Edge {
	return de.edge
}

// SetInResult sets whether this edge is in the result.
func (de *Geomgraph_DirectedEdge) SetInResult(isInResult bool) {
	de.isInResult = isInResult
}

// IsInResult returns true if this edge is in the result.
func (de *Geomgraph_DirectedEdge) IsInResult() bool {
	return de.isInResult
}

// IsVisited returns true if this edge has been visited.
func (de *Geomgraph_DirectedEdge) IsVisited() bool {
	return de.isVisited
}

// SetVisited sets whether this edge has been visited.
func (de *Geomgraph_DirectedEdge) SetVisited(isVisited bool) {
	de.isVisited = isVisited
}

// SetEdgeRing sets the EdgeRing that this edge is part of.
func (de *Geomgraph_DirectedEdge) SetEdgeRing(edgeRing *Geomgraph_EdgeRing) {
	de.edgeRing = edgeRing
}

// GetEdgeRing returns the EdgeRing that this edge is part of.
func (de *Geomgraph_DirectedEdge) GetEdgeRing() *Geomgraph_EdgeRing {
	return de.edgeRing
}

// SetMinEdgeRing sets the MinimalEdgeRing that this edge is part of.
func (de *Geomgraph_DirectedEdge) SetMinEdgeRing(minEdgeRing *Geomgraph_EdgeRing) {
	de.minEdgeRing = minEdgeRing
}

// GetMinEdgeRing returns the MinimalEdgeRing that this edge is part of.
func (de *Geomgraph_DirectedEdge) GetMinEdgeRing() *Geomgraph_EdgeRing {
	return de.minEdgeRing
}

// GetDepth returns the depth for a given position.
func (de *Geomgraph_DirectedEdge) GetDepth(position int) int {
	return de.depth[position]
}

// SetDepth sets the depth for a position. You may also use SetEdgeDepths to
// update depth and opposite depth together.
func (de *Geomgraph_DirectedEdge) SetDepth(position, depthVal int) {
	if de.depth[position] != -999 {
		if de.depth[position] != depthVal {
			panic(Geom_NewTopologyExceptionWithCoordinate("assigned depths do not match", de.GetCoordinate()))
		}
	}
	de.depth[position] = depthVal
}

// GetDepthDelta returns the depth delta for this edge.
func (de *Geomgraph_DirectedEdge) GetDepthDelta() int {
	depthDelta := de.edge.GetDepthDelta()
	if !de.isForward {
		depthDelta = -depthDelta
	}
	return depthDelta
}

// SetVisitedEdge marks both DirectedEdges attached to a given Edge. This is
// used for edges corresponding to lines, which will only appear oriented in
// a single direction in the result.
func (de *Geomgraph_DirectedEdge) SetVisitedEdge(isVisited bool) {
	de.SetVisited(isVisited)
	de.sym.SetVisited(isVisited)
}

// GetSym returns the DirectedEdge for the same Edge but in the opposite direction.
func (de *Geomgraph_DirectedEdge) GetSym() *Geomgraph_DirectedEdge {
	return de.sym
}

// IsForward returns true if this edge is in the forward direction.
func (de *Geomgraph_DirectedEdge) IsForward() bool {
	return de.isForward
}

// SetSym sets the symmetric DirectedEdge.
func (de *Geomgraph_DirectedEdge) SetSym(sym *Geomgraph_DirectedEdge) {
	de.sym = sym
}

// GetNext returns the next edge in the edge ring.
func (de *Geomgraph_DirectedEdge) GetNext() *Geomgraph_DirectedEdge {
	return de.next
}

// SetNext sets the next edge in the edge ring.
func (de *Geomgraph_DirectedEdge) SetNext(next *Geomgraph_DirectedEdge) {
	de.next = next
}

// GetNextMin returns the next edge in the MinimalEdgeRing.
func (de *Geomgraph_DirectedEdge) GetNextMin() *Geomgraph_DirectedEdge {
	return de.nextMin
}

// SetNextMin sets the next edge in the MinimalEdgeRing.
func (de *Geomgraph_DirectedEdge) SetNextMin(nextMin *Geomgraph_DirectedEdge) {
	de.nextMin = nextMin
}

// IsLineEdge returns true if this edge is a line edge (not an area edge).
// An edge is a line edge if at least one of the labels is a line label, and
// any labels which are not line labels have all Locations = EXTERIOR.
func (de *Geomgraph_DirectedEdge) IsLineEdge() bool {
	isLine := de.label.IsLine(0) || de.label.IsLine(1)
	isExteriorIfArea0 := !de.label.IsAreaAt(0) || de.label.AllPositionsEqual(0, Geom_Location_Exterior)
	isExteriorIfArea1 := !de.label.IsAreaAt(1) || de.label.AllPositionsEqual(1, Geom_Location_Exterior)
	return isLine && isExteriorIfArea0 && isExteriorIfArea1
}

// IsInteriorAreaEdge returns true if this is an interior Area edge. An edge is
// an interior area edge if its label is an Area label for both Geometries and
// for each Geometry both sides are in the interior.
func (de *Geomgraph_DirectedEdge) IsInteriorAreaEdge() bool {
	isInteriorAreaEdge := true
	for i := 0; i < 2; i++ {
		if !(de.label.IsAreaAt(i) &&
			de.label.GetLocation(i, Geom_Position_Left) == Geom_Location_Interior &&
			de.label.GetLocation(i, Geom_Position_Right) == Geom_Location_Interior) {
			isInteriorAreaEdge = false
		}
	}
	return isInteriorAreaEdge
}

// computeDirectedLabel computes the label in the appropriate orientation for
// this DirEdge.
func (de *Geomgraph_DirectedEdge) computeDirectedLabel() {
	de.label = Geomgraph_NewLabelFromLabel(de.edge.GetLabel())
	if !de.isForward {
		de.label.Flip()
	}
}

// SetEdgeDepths sets both edge depths. One depth for a given side is provided.
// The other is computed depending on the Location transition and the
// depthDelta of the edge.
func (de *Geomgraph_DirectedEdge) SetEdgeDepths(position, depth int) {
	// Get the depth transition delta from R to L for this directed Edge.
	depthDelta := de.GetEdge().GetDepthDelta()
	if !de.isForward {
		depthDelta = -depthDelta
	}

	// If moving from L to R instead of R to L must change sign of delta.
	directionFactor := 1
	if position == Geom_Position_Left {
		directionFactor = -1
	}

	oppositePos := Geom_Position_Opposite(position)
	delta := depthDelta * directionFactor
	oppositeDepth := depth + delta
	de.SetDepth(position, depth)
	de.SetDepth(oppositePos, oppositeDepth)
}

// Print writes a representation of this DirectedEdge to the given writer.
func (de *Geomgraph_DirectedEdge) Print(out io.Writer) {
	de.Geomgraph_EdgeEnd.Print(out)
	io.WriteString(out, " "+itoa(de.depth[Geom_Position_Left])+"/"+itoa(de.depth[Geom_Position_Right]))
	io.WriteString(out, " ("+itoa(de.GetDepthDelta())+")")
	if de.isInResult {
		io.WriteString(out, " inResult")
	}
}

// PrintEdge writes a full representation including the edge coordinates.
func (de *Geomgraph_DirectedEdge) PrintEdge(out io.Writer) {
	de.Print(out)
	io.WriteString(out, " ")
	if de.isForward {
		de.edge.Print(out)
	} else {
		de.edge.PrintReverse(out)
	}
}

// itoa is a helper to convert int to string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
