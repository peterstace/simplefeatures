package jts

import (
	"sort"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationOverlayng_OverlayEdge represents a directed edge in an overlay
// graph.
type OperationOverlayng_OverlayEdge struct {
	*Edgegraph_HalfEdge
	child java.Polymorphic

	pts       []*Geom_Coordinate
	direction bool
	dirPt     *Geom_Coordinate
	label     *OperationOverlayng_OverlayLabel

	isInResultArea bool
	isInResultLine bool
	isVisited      bool

	nextResultEdge    *OperationOverlayng_OverlayEdge
	edgeRing          *OperationOverlayng_OverlayEdgeRing
	maxEdgeRing       *OperationOverlayng_MaximalEdgeRing
	nextResultMaxEdge *OperationOverlayng_OverlayEdge
}

// GetChild returns the child type for polymorphism support.
func (oe *OperationOverlayng_OverlayEdge) GetChild() java.Polymorphic {
	return oe.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (oe *OperationOverlayng_OverlayEdge) GetParent() java.Polymorphic {
	return oe.Edgegraph_HalfEdge
}

// OperationOverlayng_OverlayEdge_CreateEdge creates a single OverlayEdge.
func OperationOverlayng_OverlayEdge_CreateEdge(pts []*Geom_Coordinate, lbl *OperationOverlayng_OverlayLabel, direction bool) *OperationOverlayng_OverlayEdge {
	var origin, dirPt *Geom_Coordinate
	if direction {
		origin = pts[0]
		dirPt = pts[1]
	} else {
		ilast := len(pts) - 1
		origin = pts[ilast]
		dirPt = pts[ilast-1]
	}
	return OperationOverlayng_NewOverlayEdge(origin, dirPt, direction, lbl, pts)
}

// OperationOverlayng_OverlayEdge_CreateEdgePair creates a pair of OverlayEdges
// with opposite directions.
func OperationOverlayng_OverlayEdge_CreateEdgePair(pts []*Geom_Coordinate, lbl *OperationOverlayng_OverlayLabel) *OperationOverlayng_OverlayEdge {
	e0 := OperationOverlayng_OverlayEdge_CreateEdge(pts, lbl, true)
	e1 := OperationOverlayng_OverlayEdge_CreateEdge(pts, lbl, false)
	e0.linkOverlayEdges(e1)
	return e0
}

// linkOverlayEdges links two overlay edges as symmetric pairs.
func (oe *OperationOverlayng_OverlayEdge) linkOverlayEdges(sym *OperationOverlayng_OverlayEdge) {
	// Link at the HalfEdge level
	oe.Edgegraph_HalfEdge.Link(sym.Edgegraph_HalfEdge)
}

// OperationOverlayng_OverlayEdge_NodeComparator returns a function that sorts
// OverlayEdges by their origin Coordinates.
func OperationOverlayng_OverlayEdge_NodeComparator() func(e1, e2 *OperationOverlayng_OverlayEdge) int {
	return func(e1, e2 *OperationOverlayng_OverlayEdge) int {
		return e1.Orig().CompareTo(e2.Orig())
	}
}

// OperationOverlayng_OverlayEdge_SortByNode sorts a slice of OverlayEdges by
// their origin coordinates.
func OperationOverlayng_OverlayEdge_SortByNode(edges []*OperationOverlayng_OverlayEdge) {
	cmp := OperationOverlayng_OverlayEdge_NodeComparator()
	sort.Slice(edges, func(i, j int) bool {
		return cmp(edges[i], edges[j]) < 0
	})
}

// OperationOverlayng_NewOverlayEdge creates a new OverlayEdge.
func OperationOverlayng_NewOverlayEdge(orig, dirPt *Geom_Coordinate, direction bool, label *OperationOverlayng_OverlayLabel, pts []*Geom_Coordinate) *OperationOverlayng_OverlayEdge {
	halfEdge := Edgegraph_NewHalfEdge(orig)
	oe := &OperationOverlayng_OverlayEdge{
		Edgegraph_HalfEdge: halfEdge,
		dirPt:             dirPt,
		direction:         direction,
		pts:               pts,
		label:             label,
	}
	halfEdge.child = oe
	return oe
}

// IsForward returns true if this edge has the forward direction.
func (oe *OperationOverlayng_OverlayEdge) IsForward() bool {
	return oe.direction
}

// DirectionPt_BODY overrides the base class DirectionPt.
func (oe *OperationOverlayng_OverlayEdge) DirectionPt_BODY() *Geom_Coordinate {
	return oe.dirPt
}

// GetLabel returns the label for this edge.
func (oe *OperationOverlayng_OverlayEdge) GetLabel() *OperationOverlayng_OverlayLabel {
	return oe.label
}

// GetLocation returns the location for a given index and position.
func (oe *OperationOverlayng_OverlayEdge) GetLocation(index, position int) int {
	return oe.label.GetLocation(index, position, oe.direction)
}

// GetCoordinate returns the origin coordinate.
func (oe *OperationOverlayng_OverlayEdge) GetCoordinate() *Geom_Coordinate {
	return oe.Orig()
}

// GetCoordinates returns the coordinates of this edge.
func (oe *OperationOverlayng_OverlayEdge) GetCoordinates() []*Geom_Coordinate {
	return oe.pts
}

// GetCoordinatesOriented returns the coordinates in the direction of this edge.
func (oe *OperationOverlayng_OverlayEdge) GetCoordinatesOriented() []*Geom_Coordinate {
	if oe.direction {
		return oe.pts
	}
	cpy := make([]*Geom_Coordinate, len(oe.pts))
	copy(cpy, oe.pts)
	Geom_CoordinateArrays_Reverse(cpy)
	return cpy
}

// AddCoordinates adds the coordinates of this edge to the given list, in the
// direction of the edge. Duplicate coordinates are removed (which means that
// this is safe to use for a path of connected edges in the topology graph).
func (oe *OperationOverlayng_OverlayEdge) AddCoordinates(coords *Geom_CoordinateList) {
	isFirstEdge := coords.Size() > 0
	if oe.direction {
		startIndex := 1
		if isFirstEdge {
			startIndex = 0
		}
		for i := startIndex; i < len(oe.pts); i++ {
			coords.AddCoordinate(oe.pts[i], false)
		}
	} else {
		// is backward
		startIndex := len(oe.pts) - 2
		if isFirstEdge {
			startIndex = len(oe.pts) - 1
		}
		for i := startIndex; i >= 0; i-- {
			coords.AddCoordinate(oe.pts[i], false)
		}
	}
}

// SymOE gets the symmetric pair edge of this edge as an OverlayEdge.
func (oe *OperationOverlayng_OverlayEdge) SymOE() *OperationOverlayng_OverlayEdge {
	sym := oe.Sym()
	if sym == nil {
		return nil
	}
	return sym.child.(*OperationOverlayng_OverlayEdge)
}

// ONextOE gets the next edge CCW around the origin of this edge, with the same
// origin, as an OverlayEdge.
func (oe *OperationOverlayng_OverlayEdge) ONextOE() *OperationOverlayng_OverlayEdge {
	oNext := oe.ONext()
	if oNext == nil {
		return nil
	}
	return oNext.child.(*OperationOverlayng_OverlayEdge)
}

// NextOE gets the next edge CCW around the destination vertex of this edge as
// an OverlayEdge.
func (oe *OperationOverlayng_OverlayEdge) NextOE() *OperationOverlayng_OverlayEdge {
	next := oe.Next()
	if next == nil {
		return nil
	}
	return next.child.(*OperationOverlayng_OverlayEdge)
}

// PrevOE gets the previous edge CW around the origin vertex of this edge as an
// OverlayEdge.
func (oe *OperationOverlayng_OverlayEdge) PrevOE() *OperationOverlayng_OverlayEdge {
	prev := oe.Prev()
	if prev == nil {
		return nil
	}
	return prev.child.(*OperationOverlayng_OverlayEdge)
}

// IsInResultArea returns whether this edge is in the result area.
func (oe *OperationOverlayng_OverlayEdge) IsInResultArea() bool {
	return oe.isInResultArea
}

// IsInResultAreaBoth returns whether both this edge and its symmetric edge are
// in the result area.
func (oe *OperationOverlayng_OverlayEdge) IsInResultAreaBoth() bool {
	return oe.isInResultArea && oe.SymOE().isInResultArea
}

// UnmarkFromResultAreaBoth unmarks both this edge and its symmetric edge from
// the result area.
func (oe *OperationOverlayng_OverlayEdge) UnmarkFromResultAreaBoth() {
	oe.isInResultArea = false
	oe.SymOE().isInResultArea = false
}

// MarkInResultArea marks this edge as in the result area.
func (oe *OperationOverlayng_OverlayEdge) MarkInResultArea() {
	oe.isInResultArea = true
}

// MarkInResultAreaBoth marks both this edge and its symmetric edge as in the
// result area.
func (oe *OperationOverlayng_OverlayEdge) MarkInResultAreaBoth() {
	oe.isInResultArea = true
	oe.SymOE().isInResultArea = true
}

// IsInResultLine returns whether this edge is in the result line.
func (oe *OperationOverlayng_OverlayEdge) IsInResultLine() bool {
	return oe.isInResultLine
}

// MarkInResultLine marks this edge and its symmetric edge as in the result line.
func (oe *OperationOverlayng_OverlayEdge) MarkInResultLine() {
	oe.isInResultLine = true
	oe.SymOE().isInResultLine = true
}

// IsInResult returns whether this edge is in the result (either area or line).
func (oe *OperationOverlayng_OverlayEdge) IsInResult() bool {
	return oe.isInResultArea || oe.isInResultLine
}

// IsInResultEither returns whether either this edge or its symmetric edge is
// in the result.
func (oe *OperationOverlayng_OverlayEdge) IsInResultEither() bool {
	return oe.IsInResult() || oe.SymOE().IsInResult()
}

// SetNextResult sets the next result edge.
func (oe *OperationOverlayng_OverlayEdge) SetNextResult(e *OperationOverlayng_OverlayEdge) {
	oe.nextResultEdge = e
}

// NextResult returns the next result edge.
func (oe *OperationOverlayng_OverlayEdge) NextResult() *OperationOverlayng_OverlayEdge {
	return oe.nextResultEdge
}

// IsResultLinked returns whether this edge has a next result edge linked.
func (oe *OperationOverlayng_OverlayEdge) IsResultLinked() bool {
	return oe.nextResultEdge != nil
}

// SetNextResultMax sets the next result max edge.
func (oe *OperationOverlayng_OverlayEdge) SetNextResultMax(e *OperationOverlayng_OverlayEdge) {
	oe.nextResultMaxEdge = e
}

// NextResultMax returns the next result max edge.
func (oe *OperationOverlayng_OverlayEdge) NextResultMax() *OperationOverlayng_OverlayEdge {
	return oe.nextResultMaxEdge
}

// IsResultMaxLinked returns whether this edge has a next result max edge linked.
func (oe *OperationOverlayng_OverlayEdge) IsResultMaxLinked() bool {
	return oe.nextResultMaxEdge != nil
}

// IsVisited returns whether this edge has been visited.
func (oe *OperationOverlayng_OverlayEdge) IsVisited() bool {
	return oe.isVisited
}

func (oe *OperationOverlayng_OverlayEdge) markVisited() {
	oe.isVisited = true
}

// MarkVisitedBoth marks both this edge and its symmetric edge as visited.
func (oe *OperationOverlayng_OverlayEdge) MarkVisitedBoth() {
	oe.markVisited()
	oe.SymOE().markVisited()
}

// SetEdgeRing sets the edge ring for this edge.
func (oe *OperationOverlayng_OverlayEdge) SetEdgeRing(edgeRing *OperationOverlayng_OverlayEdgeRing) {
	oe.edgeRing = edgeRing
}

// GetEdgeRing returns the edge ring for this edge.
func (oe *OperationOverlayng_OverlayEdge) GetEdgeRing() *OperationOverlayng_OverlayEdgeRing {
	return oe.edgeRing
}

// GetEdgeRingMax returns the maximal edge ring for this edge.
func (oe *OperationOverlayng_OverlayEdge) GetEdgeRingMax() *OperationOverlayng_MaximalEdgeRing {
	return oe.maxEdgeRing
}

// SetEdgeRingMax sets the maximal edge ring for this edge.
func (oe *OperationOverlayng_OverlayEdge) SetEdgeRingMax(maximalEdgeRing *OperationOverlayng_MaximalEdgeRing) {
	oe.maxEdgeRing = maximalEdgeRing
}

// String returns a string representation of this edge.
func (oe *OperationOverlayng_OverlayEdge) String() string {
	orig := oe.Orig()
	dest := oe.Dest()
	dirPtStr := ""
	if len(oe.pts) > 2 {
		dirPtStr = ", " + IO_WKTWriter_Format(oe.DirectionPt())
	}

	return "OE( " + IO_WKTWriter_Format(orig) +
		dirPtStr +
		" .. " + IO_WKTWriter_Format(dest) +
		" ) " +
		oe.label.ToStringWithDirection(oe.direction) +
		oe.resultSymbol() +
		" / Sym: " + oe.SymOE().GetLabel().ToStringWithDirection(oe.SymOE().direction) +
		oe.SymOE().resultSymbol()
}

func (oe *OperationOverlayng_OverlayEdge) resultSymbol() string {
	if oe.isInResultArea {
		return " resA"
	}
	if oe.isInResultLine {
		return " resL"
	}
	return ""
}
