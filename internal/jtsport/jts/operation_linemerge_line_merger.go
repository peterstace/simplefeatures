package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationLinemerge_LineMerger merges a collection of linear components to form maximal-length linestrings.
//
// Merging stops at nodes of degree 1 or degree 3 or more.
// In other words, all nodes of degree 2 are merged together.
// The exception is in the case of an isolated loop, which only has degree-2 nodes.
// In this case one of the nodes is chosen as a starting point.
//
// The direction of each merged LineString will be that of the majority of the LineStrings
// from which it was derived.
//
// Any dimension of Geometry is handled - the constituent linework is extracted to
// form the edges. The edges must be correctly noded; that is, they must only meet
// at their endpoints. The LineMerger will accept non-noded input but will not merge
// non-noded edges.
//
// Input lines which are empty or contain only a single unique coordinate are not included
// in the merging.
type OperationLinemerge_LineMerger struct {
	graph             *OperationLinemerge_LineMergeGraph
	mergedLineStrings []*Geom_LineString
	factory           *Geom_GeometryFactory
	edgeStrings       []*OperationLinemerge_EdgeString
}

// OperationLinemerge_NewLineMerger creates a new line merger.
func OperationLinemerge_NewLineMerger() *OperationLinemerge_LineMerger {
	return &OperationLinemerge_LineMerger{
		graph: OperationLinemerge_NewLineMergeGraph(),
	}
}

// AddGeometry adds a Geometry to be processed. May be called multiple times.
// Any dimension of Geometry may be added; the constituent linework will be extracted.
func (lm *OperationLinemerge_LineMerger) AddGeometry(geometry *Geom_Geometry) {
	filter := operationLinemerge_NewLineMergerFilter(lm)
	geometry.Apply(filter)
}

// AddGeometries adds a collection of Geometries to be processed. May be called multiple times.
// Any dimension of Geometry may be added; the constituent linework will be extracted.
func (lm *OperationLinemerge_LineMerger) AddGeometries(geometries []*Geom_Geometry) {
	lm.mergedLineStrings = nil
	for _, geometry := range geometries {
		lm.AddGeometry(geometry)
	}
}

// addLineString adds a LineString to the graph.
func (lm *OperationLinemerge_LineMerger) addLineString(lineString *Geom_LineString) {
	if lm.factory == nil {
		lm.factory = lineString.GetFactory()
	}
	lm.graph.AddEdge(lineString)
}

// merge performs the merge operation.
func (lm *OperationLinemerge_LineMerger) merge() {
	if lm.mergedLineStrings != nil {
		return
	}

	// Reset marks (this allows incremental processing).
	lm.setMarkedOnNodes(false)
	lm.setMarkedOnEdges(false)

	lm.edgeStrings = make([]*OperationLinemerge_EdgeString, 0)
	lm.buildEdgeStringsForObviousStartNodes()
	lm.buildEdgeStringsForIsolatedLoops()
	lm.mergedLineStrings = make([]*Geom_LineString, 0)
	for _, edgeString := range lm.edgeStrings {
		lm.mergedLineStrings = append(lm.mergedLineStrings, edgeString.ToLineString())
	}
}

// setMarkedOnNodes sets the marked flag on all nodes.
func (lm *OperationLinemerge_LineMerger) setMarkedOnNodes(marked bool) {
	nodes := lm.graph.GetNodes()
	components := make([]*Planargraph_GraphComponent, len(nodes))
	for i, node := range nodes {
		components[i] = node.Planargraph_GraphComponent
	}
	Planargraph_GraphComponent_SetMarkedIterator(components, marked)
}

// setMarkedOnEdges sets the marked flag on all edges.
func (lm *OperationLinemerge_LineMerger) setMarkedOnEdges(marked bool) {
	edges := lm.graph.GetEdges()
	components := make([]*Planargraph_GraphComponent, len(edges))
	for i, edge := range edges {
		components[i] = edge.Planargraph_GraphComponent
	}
	Planargraph_GraphComponent_SetMarkedIterator(components, marked)
}

// buildEdgeStringsForObviousStartNodes builds edge strings starting from obvious start nodes.
func (lm *OperationLinemerge_LineMerger) buildEdgeStringsForObviousStartNodes() {
	lm.buildEdgeStringsForNonDegree2Nodes()
}

// buildEdgeStringsForIsolatedLoops builds edge strings for isolated loops.
func (lm *OperationLinemerge_LineMerger) buildEdgeStringsForIsolatedLoops() {
	lm.buildEdgeStringsForUnprocessedNodes()
}

// buildEdgeStringsForUnprocessedNodes builds edge strings for nodes that haven't been processed.
func (lm *OperationLinemerge_LineMerger) buildEdgeStringsForUnprocessedNodes() {
	for _, node := range lm.graph.GetNodes() {
		if !node.IsMarked() {
			Util_Assert_IsTrue(node.GetDegree() == 2)
			lm.buildEdgeStringsStartingAt(node)
			node.SetMarked(true)
		}
	}
}

// buildEdgeStringsForNonDegree2Nodes builds edge strings starting from non-degree-2 nodes.
func (lm *OperationLinemerge_LineMerger) buildEdgeStringsForNonDegree2Nodes() {
	for _, node := range lm.graph.GetNodes() {
		if node.GetDegree() != 2 {
			lm.buildEdgeStringsStartingAt(node)
			node.SetMarked(true)
		}
	}
}

// buildEdgeStringsStartingAt builds edge strings starting at the given node.
func (lm *OperationLinemerge_LineMerger) buildEdgeStringsStartingAt(node *Planargraph_Node) {
	for _, directedEdge := range node.GetOutEdges().GetEdges() {
		if directedEdge.GetEdge().IsMarked() {
			continue
		}
		lineMergeDE := directedEdge.GetChild().(*OperationLinemerge_LineMergeDirectedEdge)
		lm.edgeStrings = append(lm.edgeStrings, lm.buildEdgeStringStartingWith(lineMergeDE))
	}
}

// buildEdgeStringStartingWith builds an edge string starting with the given directed edge.
func (lm *OperationLinemerge_LineMerger) buildEdgeStringStartingWith(start *OperationLinemerge_LineMergeDirectedEdge) *OperationLinemerge_EdgeString {
	edgeString := OperationLinemerge_NewEdgeString(lm.factory)
	current := start
	for {
		edgeString.Add(current)
		current.GetEdge().SetMarked(true)
		current = current.GetNext()
		if current == nil || current == start {
			break
		}
	}
	return edgeString
}

// GetMergedLineStrings gets the LineStrings created by the merging process.
func (lm *OperationLinemerge_LineMerger) GetMergedLineStrings() []*Geom_LineString {
	lm.merge()
	return lm.mergedLineStrings
}

// operationLinemerge_LineMergerFilter is a filter that extracts LineStrings from a geometry.
type operationLinemerge_LineMergerFilter struct {
	lm *OperationLinemerge_LineMerger
}

var _ Geom_GeometryComponentFilter = (*operationLinemerge_LineMergerFilter)(nil)

func (f *operationLinemerge_LineMergerFilter) IsGeom_GeometryComponentFilter() {}

func operationLinemerge_NewLineMergerFilter(lm *OperationLinemerge_LineMerger) *operationLinemerge_LineMergerFilter {
	return &operationLinemerge_LineMergerFilter{lm: lm}
}

func (f *operationLinemerge_LineMergerFilter) Filter(geom *Geom_Geometry) {
	if java.InstanceOf[*Geom_LineString](geom) {
		f.lm.addLineString(java.Cast[*Geom_LineString](geom))
	}
}
