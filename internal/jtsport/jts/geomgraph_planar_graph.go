package jts

import (
	"io"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_PlanarGraph_LinkResultDirectedEdges links the DirectedEdges at the
// nodes of a collection that are in the result.
func Geomgraph_PlanarGraph_LinkResultDirectedEdges(nodes []*Geomgraph_Node) {
	for _, node := range nodes {
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		des.LinkResultDirectedEdges()
	}
}

// Geomgraph_PlanarGraph is a graph that models a given Geometry.
type Geomgraph_PlanarGraph struct {
	child java.Polymorphic

	edges       []*Geomgraph_Edge
	nodes       *Geomgraph_NodeMap
	edgeEndList []*Geomgraph_EdgeEnd
}

// GetChild returns the immediate child in the type hierarchy chain.
func (pg *Geomgraph_PlanarGraph) GetChild() java.Polymorphic {
	return pg.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (pg *Geomgraph_PlanarGraph) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewPlanarGraph creates a new PlanarGraph with the given NodeFactory.
func Geomgraph_NewPlanarGraph(nodeFact *Geomgraph_NodeFactory) *Geomgraph_PlanarGraph {
	return &Geomgraph_PlanarGraph{
		edges:       make([]*Geomgraph_Edge, 0),
		nodes:       Geomgraph_NewNodeMap(nodeFact),
		edgeEndList: make([]*Geomgraph_EdgeEnd, 0),
	}
}

// Geomgraph_NewPlanarGraphDefault creates a new PlanarGraph with the default NodeFactory.
func Geomgraph_NewPlanarGraphDefault() *Geomgraph_PlanarGraph {
	return Geomgraph_NewPlanarGraph(Geomgraph_NewNodeFactory())
}

// GetEdgeIterator returns an iterator over the edges.
func (pg *Geomgraph_PlanarGraph) GetEdgeIterator() []*Geomgraph_Edge {
	return pg.edges
}

// GetEdgeEnds returns the edge ends.
func (pg *Geomgraph_PlanarGraph) GetEdgeEnds() []*Geomgraph_EdgeEnd {
	return pg.edgeEndList
}

// IsBoundaryNode returns true if the coordinate is on the boundary of the
// geometry with the given index.
func (pg *Geomgraph_PlanarGraph) IsBoundaryNode(geomIndex int, coord *Geom_Coordinate) bool {
	node := pg.nodes.Find(coord)
	if node == nil {
		return false
	}
	label := node.GetLabel()
	if label != nil && label.GetLocationOn(geomIndex) == Geom_Location_Boundary {
		return true
	}
	return false
}

// InsertEdge inserts an edge into the graph.
func (pg *Geomgraph_PlanarGraph) InsertEdge(e *Geomgraph_Edge) {
	pg.edges = append(pg.edges, e)
}

// Add adds an EdgeEnd to the graph.
func (pg *Geomgraph_PlanarGraph) Add(e *Geomgraph_EdgeEnd) {
	pg.nodes.Add(e)
	pg.edgeEndList = append(pg.edgeEndList, e)
}

// GetNodeIterator returns an iterator over the nodes.
func (pg *Geomgraph_PlanarGraph) GetNodeIterator() []*Geomgraph_Node {
	return pg.nodes.Values()
}

// GetNodes returns the collection of nodes.
func (pg *Geomgraph_PlanarGraph) GetNodes() []*Geomgraph_Node {
	return pg.nodes.Values()
}

// AddNode adds a node to the graph.
func (pg *Geomgraph_PlanarGraph) AddNode(node *Geomgraph_Node) *Geomgraph_Node {
	return pg.nodes.AddNode(node)
}

// AddNodeFromCoord adds a node at the given coordinate.
func (pg *Geomgraph_PlanarGraph) AddNodeFromCoord(coord *Geom_Coordinate) *Geomgraph_Node {
	return pg.nodes.AddNodeFromCoord(coord)
}

// Find returns the node at the given coordinate, or nil if not found.
func (pg *Geomgraph_PlanarGraph) Find(coord *Geom_Coordinate) *Geomgraph_Node {
	return pg.nodes.Find(coord)
}

// AddEdges adds a set of edges to the graph. For each edge two DirectedEdges
// will be created. DirectedEdges are NOT linked by this method.
func (pg *Geomgraph_PlanarGraph) AddEdges(edgesToAdd []*Geomgraph_Edge) {
	// Create all the nodes for the edges.
	for _, e := range edgesToAdd {
		pg.edges = append(pg.edges, e)

		de1 := Geomgraph_NewDirectedEdge(e, true)
		de2 := Geomgraph_NewDirectedEdge(e, false)
		de1.SetSym(de2)
		de2.SetSym(de1)

		pg.Add(de1.Geomgraph_EdgeEnd)
		pg.Add(de2.Geomgraph_EdgeEnd)
	}
}

// LinkResultDirectedEdges links the DirectedEdges at the nodes of the graph.
func (pg *Geomgraph_PlanarGraph) LinkResultDirectedEdges() {
	for _, node := range pg.nodes.Values() {
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		des.LinkResultDirectedEdges()
	}
}

// LinkAllDirectedEdges links all the DirectedEdges at the nodes of the graph.
func (pg *Geomgraph_PlanarGraph) LinkAllDirectedEdges() {
	for _, node := range pg.nodes.Values() {
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		des.LinkAllDirectedEdges()
	}
}

// FindEdgeEnd returns the EdgeEnd which has edge e as its base edge.
func (pg *Geomgraph_PlanarGraph) FindEdgeEnd(e *Geomgraph_Edge) *Geomgraph_EdgeEnd {
	for _, ee := range pg.GetEdgeEnds() {
		if ee.GetEdge() == e {
			return ee
		}
	}
	return nil
}

// FindEdge returns the edge whose first two coordinates are p0 and p1.
func (pg *Geomgraph_PlanarGraph) FindEdge(p0, p1 *Geom_Coordinate) *Geomgraph_Edge {
	for _, e := range pg.edges {
		eCoord := e.GetCoordinates()
		if p0.Equals(eCoord[0]) && p1.Equals(eCoord[1]) {
			return e
		}
	}
	return nil
}

// FindEdgeInSameDirection returns the edge which starts at p0 and whose first
// segment is parallel to p1.
func (pg *Geomgraph_PlanarGraph) FindEdgeInSameDirection(p0, p1 *Geom_Coordinate) *Geomgraph_Edge {
	for _, e := range pg.edges {
		eCoord := e.GetCoordinates()
		if pg.matchInSameDirection(p0, p1, eCoord[0], eCoord[1]) {
			return e
		}
		if pg.matchInSameDirection(p0, p1, eCoord[len(eCoord)-1], eCoord[len(eCoord)-2]) {
			return e
		}
	}
	return nil
}

// matchInSameDirection checks if coordinate pairs define line segments lying
// in the same direction. E.g. the segments are parallel and in the same
// quadrant (as opposed to parallel and opposite!).
func (pg *Geomgraph_PlanarGraph) matchInSameDirection(p0, p1, ep0, ep1 *Geom_Coordinate) bool {
	if !p0.Equals(ep0) {
		return false
	}
	if Algorithm_Orientation_Index(p0, p1, ep1) == Algorithm_Orientation_Collinear &&
		Geom_Quadrant_QuadrantFromCoords(p0, p1) == Geom_Quadrant_QuadrantFromCoords(ep0, ep1) {
		return true
	}
	return false
}

// PrintEdges writes the edges to the given writer.
func (pg *Geomgraph_PlanarGraph) PrintEdges(out io.Writer) {
	io.WriteString(out, "Edges:\n")
	for i, e := range pg.edges {
		io.WriteString(out, "edge "+itoa(i)+":\n")
		e.Print(out)
		e.GetEdgeIntersectionList().Print(out)
	}
}
