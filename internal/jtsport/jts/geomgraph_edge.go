package jts

import (
	"fmt"
	"io"
	"strings"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_Edge_UpdateIM updates an IM from the label for an edge. Handles
// edges from both L and A geometries.
func Geomgraph_Edge_UpdateIM(label *Geomgraph_Label, im *Geom_IntersectionMatrix) {
	im.SetAtLeastIfValid(label.GetLocation(0, Geom_Position_On), label.GetLocation(1, Geom_Position_On), 1)
	if label.IsArea() {
		im.SetAtLeastIfValid(label.GetLocation(0, Geom_Position_Left), label.GetLocation(1, Geom_Position_Left), 2)
		im.SetAtLeastIfValid(label.GetLocation(0, Geom_Position_Right), label.GetLocation(1, Geom_Position_Right), 2)
	}
}

// Geomgraph_Edge represents an edge in a topology graph.
type Geomgraph_Edge struct {
	*Geomgraph_GraphComponent
	child java.Polymorphic

	pts        []*Geom_Coordinate
	env        *Geom_Envelope
	eiList     *Geomgraph_EdgeIntersectionList
	name       string
	mce        *GeomgraphIndex_MonotoneChainEdge
	isIsolated bool
	depth      *Geomgraph_Depth
	depthDelta int // The change in area depth from the R to L side of this edge.
}

// GetChild returns the immediate child in the type hierarchy chain.
func (e *Geomgraph_Edge) GetChild() java.Polymorphic {
	return e.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (e *Geomgraph_Edge) GetParent() java.Polymorphic {
	return e.Geomgraph_GraphComponent
}

// Geomgraph_NewEdge creates a new Edge with the given coordinates and label.
func Geomgraph_NewEdge(pts []*Geom_Coordinate, label *Geomgraph_Label) *Geomgraph_Edge {
	gc := Geomgraph_NewGraphComponent()
	edge := &Geomgraph_Edge{
		Geomgraph_GraphComponent: gc,
		pts:                      pts,
		isIsolated:               true,
		depth:                    Geomgraph_NewDepth(),
	}
	gc.child = edge
	gc.label = label
	edge.eiList = Geomgraph_NewEdgeIntersectionList(edge)
	return edge
}

// Geomgraph_NewEdgeFromCoords creates a new Edge with the given coordinates.
func Geomgraph_NewEdgeFromCoords(pts []*Geom_Coordinate) *Geomgraph_Edge {
	return Geomgraph_NewEdge(pts, nil)
}

// GetNumPoints returns the number of points in this edge.
func (e *Geomgraph_Edge) GetNumPoints() int {
	return len(e.pts)
}

// SetName sets the name of this edge.
func (e *Geomgraph_Edge) SetName(name string) {
	e.name = name
}

// GetCoordinates returns all coordinates of this edge.
func (e *Geomgraph_Edge) GetCoordinates() []*Geom_Coordinate {
	return e.pts
}

// GetCoordinateAtIndex returns the coordinate at the given index.
func (e *Geomgraph_Edge) GetCoordinateAtIndex(i int) *Geom_Coordinate {
	return e.pts[i]
}

// GetCoordinate_BODY returns the first coordinate of this edge (or nil if empty).
func (e *Geomgraph_Edge) GetCoordinate_BODY() *Geom_Coordinate {
	if len(e.pts) > 0 {
		return e.pts[0]
	}
	return nil
}

// GetEnvelope returns the envelope of this edge.
func (e *Geomgraph_Edge) GetEnvelope() *Geom_Envelope {
	// Compute envelope lazily.
	if e.env == nil {
		e.env = Geom_NewEnvelope()
		for _, pt := range e.pts {
			e.env.ExpandToIncludeCoordinate(pt)
		}
	}
	return e.env
}

// GetDepth returns the depth of this edge.
func (e *Geomgraph_Edge) GetDepth() *Geomgraph_Depth {
	return e.depth
}

// GetDepthDelta returns the change in depth as an edge is crossed from R to L.
func (e *Geomgraph_Edge) GetDepthDelta() int {
	return e.depthDelta
}

// SetDepthDelta sets the change in depth as an edge is crossed from R to L.
func (e *Geomgraph_Edge) SetDepthDelta(depthDelta int) {
	e.depthDelta = depthDelta
}

// GetMaximumSegmentIndex returns the maximum segment index.
func (e *Geomgraph_Edge) GetMaximumSegmentIndex() int {
	return len(e.pts) - 1
}

// GetEdgeIntersectionList returns the EdgeIntersectionList for this edge.
func (e *Geomgraph_Edge) GetEdgeIntersectionList() *Geomgraph_EdgeIntersectionList {
	return e.eiList
}

// GetMonotoneChainEdge returns the MonotoneChainEdge for this edge.
func (e *Geomgraph_Edge) GetMonotoneChainEdge() *GeomgraphIndex_MonotoneChainEdge {
	if e.mce == nil {
		e.mce = GeomgraphIndex_NewMonotoneChainEdge(e)
	}
	return e.mce
}

// IsClosed returns true if the edge is closed (first point equals last point).
func (e *Geomgraph_Edge) IsClosed() bool {
	return e.pts[0].Equals(e.pts[len(e.pts)-1])
}

// IsCollapsed returns true if this edge is collapsed. An Edge is collapsed if
// it is an Area edge and it consists of two segments which are equal and
// opposite (eg a zero-width V).
func (e *Geomgraph_Edge) IsCollapsed() bool {
	if !e.label.IsArea() {
		return false
	}
	if len(e.pts) != 3 {
		return false
	}
	if e.pts[0].Equals(e.pts[2]) {
		return true
	}
	return false
}

// GetCollapsedEdge returns a collapsed version of this edge.
func (e *Geomgraph_Edge) GetCollapsedEdge() *Geomgraph_Edge {
	newPts := []*Geom_Coordinate{e.pts[0], e.pts[1]}
	return Geomgraph_NewEdge(newPts, Geomgraph_Label_ToLineLabel(e.label))
}

// SetIsolated sets whether this edge is isolated.
func (e *Geomgraph_Edge) SetIsolated(isIsolated bool) {
	e.isIsolated = isIsolated
}

// IsIsolated_BODY returns true if this edge is isolated.
func (e *Geomgraph_Edge) IsIsolated_BODY() bool {
	return e.isIsolated
}

// AddIntersections adds EdgeIntersections for one or both intersections found
// for a segment of an edge to the edge intersection list.
func (e *Geomgraph_Edge) AddIntersections(li *Algorithm_LineIntersector, segmentIndex, geomIndex int) {
	for i := 0; i < li.GetIntersectionNum(); i++ {
		e.AddIntersection(li, segmentIndex, geomIndex, i)
	}
}

// AddIntersection adds an EdgeIntersection for intersection intIndex. An
// intersection that falls exactly on a vertex of the edge is normalized to use
// the higher of the two possible segmentIndexes.
func (e *Geomgraph_Edge) AddIntersection(li *Algorithm_LineIntersector, segmentIndex, geomIndex, intIndex int) {
	intPt := li.GetIntersection(intIndex).Copy()
	normalizedSegmentIndex := segmentIndex
	dist := li.GetEdgeDistance(geomIndex, intIndex)

	// Normalize the intersection point location.
	nextSegIndex := normalizedSegmentIndex + 1
	if nextSegIndex < len(e.pts) {
		nextPt := e.pts[nextSegIndex]

		// Normalize segment index if intPt falls on vertex. The check for point
		// equality is 2D only - Z values are ignored.
		if intPt.Equals2D(nextPt) {
			normalizedSegmentIndex = nextSegIndex
			dist = 0.0
		}
	}
	// Add the intersection point to edge intersection list.
	e.eiList.Add(intPt, normalizedSegmentIndex, dist)
}

// ComputeIM_BODY updates the IM with the contribution for this component.
func (e *Geomgraph_Edge) ComputeIM_BODY(im *Geom_IntersectionMatrix) {
	Geomgraph_Edge_UpdateIM(e.label, im)
}

// Equals checks if this edge equals another object. An edge equals another iff
// the coordinates of e1 are the same or the reverse of the coordinates in e2.
func (e *Geomgraph_Edge) Equals(other *Geomgraph_Edge) bool {
	if len(e.pts) != len(other.pts) {
		return false
	}

	isEqualForward := true
	isEqualReverse := true
	iRev := len(e.pts)
	for i := range e.pts {
		if !e.pts[i].Equals2D(other.pts[i]) {
			isEqualForward = false
		}
		iRev--
		if !e.pts[i].Equals2D(other.pts[iRev]) {
			isEqualReverse = false
		}
		if !isEqualForward && !isEqualReverse {
			return false
		}
	}
	return true
}

// IsPointwiseEqual checks if coordinate sequences of the Edges are identical.
func (e *Geomgraph_Edge) IsPointwiseEqual(other *Geomgraph_Edge) bool {
	if len(e.pts) != len(other.pts) {
		return false
	}
	for i := range e.pts {
		if !e.pts[i].Equals2D(other.pts[i]) {
			return false
		}
	}
	return true
}

// String returns a string representation of this edge.
func (e *Geomgraph_Edge) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("edge %s: ", e.name))
	builder.WriteString("LINESTRING (")
	for i, pt := range e.pts {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf("%v %v", pt.GetX(), pt.GetY()))
	}
	builder.WriteString(fmt.Sprintf(")  %v %d", e.label, e.depthDelta))
	return builder.String()
}

// Print writes a representation of this edge to the given writer.
func (e *Geomgraph_Edge) Print(out io.Writer) {
	io.WriteString(out, "edge "+e.name+": ")
	io.WriteString(out, "LINESTRING (")
	for i, pt := range e.pts {
		if i > 0 {
			io.WriteString(out, ",")
		}
		fmt.Fprintf(out, "%v %v", pt.GetX(), pt.GetY())
	}
	fmt.Fprintf(out, ")  %v %d", e.label, e.depthDelta)
}

// PrintReverse writes a representation of this edge in reverse order.
func (e *Geomgraph_Edge) PrintReverse(out io.Writer) {
	io.WriteString(out, "edge "+e.name+": ")
	for i := len(e.pts) - 1; i >= 0; i-- {
		fmt.Fprintf(out, "%v ", e.pts[i])
	}
	io.WriteString(out, "\n")
}
