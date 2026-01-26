package jts

import (
	"io"
	"sort"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_EdgeIntersectionList is a list of edge intersections along an
// Edge. Implements splitting an edge with intersections into multiple
// resultant edges.
type Geomgraph_EdgeIntersectionList struct {
	child   java.Polymorphic
	nodeMap []*Geomgraph_EdgeIntersection
	edge    *Geomgraph_Edge
}

// GetChild returns the immediate child in the type hierarchy chain.
func (eil *Geomgraph_EdgeIntersectionList) GetChild() java.Polymorphic {
	return eil.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (eil *Geomgraph_EdgeIntersectionList) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewEdgeIntersectionList creates a new EdgeIntersectionList for the
// given edge.
func Geomgraph_NewEdgeIntersectionList(edge *Geomgraph_Edge) *Geomgraph_EdgeIntersectionList {
	return &Geomgraph_EdgeIntersectionList{
		edge: edge,
	}
}

// Add adds an intersection into the list, if it isn't already there. The input
// segmentIndex and dist are expected to be normalized. Returns the
// EdgeIntersection found or added.
func (eil *Geomgraph_EdgeIntersectionList) Add(intPt *Geom_Coordinate, segmentIndex int, dist float64) *Geomgraph_EdgeIntersection {
	// Check if already exists.
	for _, ei := range eil.nodeMap {
		if ei.SegmentIndex == segmentIndex && ei.Dist == dist {
			return ei
		}
	}
	// Add new intersection.
	eiNew := Geomgraph_NewEdgeIntersection(intPt, segmentIndex, dist)
	eil.nodeMap = append(eil.nodeMap, eiNew)
	// Keep sorted by segment index and distance.
	sort.Slice(eil.nodeMap, func(i, j int) bool {
		return eil.nodeMap[i].CompareTo(eil.nodeMap[j]) < 0
	})
	return eiNew
}

// Iterator returns the list of EdgeIntersections.
func (eil *Geomgraph_EdgeIntersectionList) Iterator() []*Geomgraph_EdgeIntersection {
	return eil.nodeMap
}

// IsIntersection tests if the given point is an edge intersection.
func (eil *Geomgraph_EdgeIntersectionList) IsIntersection(pt *Geom_Coordinate) bool {
	for _, ei := range eil.nodeMap {
		if ei.Coord.Equals(pt) {
			return true
		}
	}
	return false
}

// AddEndpoints adds entries for the first and last points of the edge to the
// list.
func (eil *Geomgraph_EdgeIntersectionList) AddEndpoints() {
	maxSegIndex := eil.edge.GetNumPoints() - 1
	eil.Add(eil.edge.GetCoordinateAtIndex(0), 0, 0.0)
	eil.Add(eil.edge.GetCoordinateAtIndex(maxSegIndex), maxSegIndex, 0.0)
}

// AddSplitEdges creates new edges for all the edges that the intersections in
// this list split the parent edge into. Adds the edges to the input list (this
// is so a single list can be used to accumulate all split edges for a
// Geometry).
func (eil *Geomgraph_EdgeIntersectionList) AddSplitEdges(edgeList *[]*Geomgraph_Edge) {
	// Ensure that the list has entries for the first and last point of the
	// edge.
	eil.AddEndpoints()

	intersections := eil.Iterator()
	// There should always be at least two entries in the list.
	if len(intersections) < 2 {
		return
	}

	eiPrev := intersections[0]
	for i := 1; i < len(intersections); i++ {
		ei := intersections[i]
		newEdge := eil.createSplitEdge(eiPrev, ei)
		*edgeList = append(*edgeList, newEdge)
		eiPrev = ei
	}
}

// createSplitEdge creates a new "split edge" with the section of points
// between (and including) the two intersections. The label for the new edge is
// the same as the label for the parent edge.
func (eil *Geomgraph_EdgeIntersectionList) createSplitEdge(ei0, ei1 *Geomgraph_EdgeIntersection) *Geomgraph_Edge {
	npts := ei1.SegmentIndex - ei0.SegmentIndex + 2

	lastSegStartPt := eil.edge.GetCoordinateAtIndex(ei1.SegmentIndex)
	// If the last intersection point is not equal to its segment start pt, add
	// it to the points list as well. (This check is needed because the distance
	// metric is not totally reliable!) The check for point equality is 2D only
	// - Z values are ignored.
	useIntPt1 := ei1.Dist > 0.0 || !ei1.Coord.Equals2D(lastSegStartPt)
	if !useIntPt1 {
		npts--
	}

	pts := make([]*Geom_Coordinate, npts)
	ipt := 0
	pts[ipt] = ei0.Coord.Copy()
	ipt++
	for i := ei0.SegmentIndex + 1; i <= ei1.SegmentIndex; i++ {
		pts[ipt] = eil.edge.GetCoordinateAtIndex(i)
		ipt++
	}
	if useIntPt1 {
		pts[ipt] = ei1.Coord
	}
	return geomgraph_NewEdgeWithLabel(pts, Geomgraph_NewLabelFromLabel(eil.edge.GetLabel()))
}

// geomgraph_NewEdgeWithLabel is a temporary helper to create edges with
// labels. Will be replaced when Edge is fully ported.
func geomgraph_NewEdgeWithLabel(pts []*Geom_Coordinate, label *Geomgraph_Label) *Geomgraph_Edge {
	gc := Geomgraph_NewGraphComponent()
	edge := &Geomgraph_Edge{
		Geomgraph_GraphComponent: gc,
		pts:                     pts,
		depth:                   Geomgraph_NewDepth(),
	}
	gc.child = edge
	gc.label = label
	edge.eiList = Geomgraph_NewEdgeIntersectionList(edge)
	return edge
}

// Print writes the intersections to the given writer.
func (eil *Geomgraph_EdgeIntersectionList) Print(out io.Writer) {
	io.WriteString(out, "Intersections:\n")
	for _, ei := range eil.nodeMap {
		ei.Print(out)
	}
}
