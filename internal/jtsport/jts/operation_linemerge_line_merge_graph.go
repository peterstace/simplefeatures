package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationLinemerge_LineMergeGraph is a planar graph of edges that is analyzed to sew the edges together.
// The marked flag on Edges and Nodes indicates whether they have been logically deleted from the graph.
type OperationLinemerge_LineMergeGraph struct {
	*Planargraph_PlanarGraph
	child java.Polymorphic
}

func (g *OperationLinemerge_LineMergeGraph) GetChild() java.Polymorphic {
	return g.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (g *OperationLinemerge_LineMergeGraph) GetParent() java.Polymorphic {
	return g.Planargraph_PlanarGraph
}

// OperationLinemerge_NewLineMergeGraph constructs a new, empty LineMergeGraph.
func OperationLinemerge_NewLineMergeGraph() *OperationLinemerge_LineMergeGraph {
	pg := Planargraph_NewPlanarGraph()
	g := &OperationLinemerge_LineMergeGraph{
		Planargraph_PlanarGraph: pg,
	}
	pg.child = g
	return g
}

// AddEdge adds an Edge, DirectedEdges, and Nodes for the given LineString representation
// of an edge. Empty lines or lines with all coordinates equal are not added.
func (g *OperationLinemerge_LineMergeGraph) AddEdge(lineString *Geom_LineString) {
	if lineString.IsEmpty() {
		return
	}

	coordinates := Geom_CoordinateArrays_RemoveRepeatedPoints(lineString.GetCoordinates())

	// Don't add lines with all coordinates equal.
	if len(coordinates) <= 1 {
		return
	}

	startCoordinate := coordinates[0]
	endCoordinate := coordinates[len(coordinates)-1]
	startNode := g.getNode(startCoordinate)
	endNode := g.getNode(endCoordinate)

	directedEdge0 := OperationLinemerge_NewLineMergeDirectedEdge(startNode, endNode, coordinates[1], true)
	directedEdge1 := OperationLinemerge_NewLineMergeDirectedEdge(endNode, startNode, coordinates[len(coordinates)-2], false)

	edge := OperationLinemerge_NewLineMergeEdge(lineString)
	edge.SetDirectedEdges(directedEdge0.Planargraph_DirectedEdge, directedEdge1.Planargraph_DirectedEdge)
	g.addEdge(edge.Planargraph_Edge)
}

// getNode returns the Node at the given coordinate, creating a new one if it does not exist.
func (g *OperationLinemerge_LineMergeGraph) getNode(coordinate *Geom_Coordinate) *Planargraph_Node {
	node := g.FindNode(coordinate)
	if node == nil {
		node = Planargraph_NewNode(coordinate)
		g.addNode(node)
	}
	return node
}
