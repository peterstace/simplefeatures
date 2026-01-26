package jts

import (
	"sort"
)

// Noding_SegmentNodeList is a list of the SegmentNodes present along a noded
// SegmentString.
type Noding_SegmentNodeList struct {
	nodes []*Noding_SegmentNode
	edge  *Noding_NodedSegmentString // The parent edge.
}

// Noding_NewSegmentNodeList creates a new SegmentNodeList for the given edge.
func Noding_NewSegmentNodeList(edge *Noding_NodedSegmentString) *Noding_SegmentNodeList {
	return &Noding_SegmentNodeList{
		edge: edge,
	}
}

// Size gets the number of nodes in the list.
func (snl *Noding_SegmentNodeList) Size() int {
	return len(snl.nodes)
}

// GetEdge gets the parent edge.
func (snl *Noding_SegmentNodeList) GetEdge() *Noding_NodedSegmentString {
	return snl.edge
}

// Add adds an intersection into the list, if it isn't already there. The input
// segmentIndex and dist are expected to be normalized.
//
// Returns the SegmentNode found or added.
func (snl *Noding_SegmentNodeList) Add(intPt *Geom_Coordinate, segmentIndex int) *Noding_SegmentNode {
	eiNew := Noding_NewSegmentNode(snl.edge, intPt, segmentIndex, snl.edge.GetSegmentOctant(segmentIndex))

	// Binary search to find insertion point or existing node.
	idx := sort.Search(len(snl.nodes), func(i int) bool {
		return snl.nodes[i].CompareTo(eiNew) >= 0
	})
	found := idx < len(snl.nodes) && snl.nodes[idx].CompareTo(eiNew) == 0

	if found {
		ei := snl.nodes[idx]
		// Debugging sanity check.
		if !ei.Coord.Equals2D(intPt) {
			panic("Found equal nodes with different coordinates")
		}
		return ei
	}

	// Node does not exist, so insert it.
	snl.nodes = append(snl.nodes, nil)
	copy(snl.nodes[idx+1:], snl.nodes[idx:])
	snl.nodes[idx] = eiNew
	return eiNew
}

// Iterator returns the nodes in order. Iterates through all SegmentNodes.
func (snl *Noding_SegmentNodeList) Iterator() []*Noding_SegmentNode {
	return snl.nodes
}

// addEndpoints adds nodes for the first and last points of the edge.
func (snl *Noding_SegmentNodeList) addEndpoints() {
	maxSegIndex := snl.edge.Size() - 1
	snl.Add(snl.edge.GetCoordinate(0), 0)
	snl.Add(snl.edge.GetCoordinate(maxSegIndex), maxSegIndex)
}

// addCollapsedNodes adds nodes for any collapsed edge pairs. Collapsed edge
// pairs can be caused by inserted nodes, or they can be pre-existing in the
// edge vertex list. In order to provide the correct fully noded semantics, the
// vertex at the base of a collapsed pair must also be added as a node.
func (snl *Noding_SegmentNodeList) addCollapsedNodes() {
	var collapsedVertexIndexes []int

	snl.findCollapsesFromInsertedNodes(&collapsedVertexIndexes)
	snl.findCollapsesFromExistingVertices(&collapsedVertexIndexes)

	// Node the collapses.
	for _, vertexIndex := range collapsedVertexIndexes {
		snl.Add(snl.edge.GetCoordinate(vertexIndex), vertexIndex)
	}
}

// findCollapsesFromExistingVertices adds nodes for any collapsed edge pairs
// which are pre-existing in the vertex list.
func (snl *Noding_SegmentNodeList) findCollapsesFromExistingVertices(collapsedVertexIndexes *[]int) {
	for i := 0; i < snl.edge.Size()-2; i++ {
		p0 := snl.edge.GetCoordinate(i)
		p1 := snl.edge.GetCoordinate(i + 1)
		p2 := snl.edge.GetCoordinate(i + 2)
		if p0.Equals2D(p2) {
			// Add base of collapse as node.
			*collapsedVertexIndexes = append(*collapsedVertexIndexes, i+1)
		}
		_ = p1 // Not used but accessed in Java for reference.
	}
}

// findCollapsesFromInsertedNodes adds nodes for any collapsed edge pairs
// caused by inserted nodes. Collapsed edge pairs occur when the same
// coordinate is inserted as a node both before and after an existing edge
// vertex. To provide the correct fully noded semantics, the vertex must be
// added as a node as well.
func (snl *Noding_SegmentNodeList) findCollapsesFromInsertedNodes(collapsedVertexIndexes *[]int) {
	// There should always be at least two entries in the list, since the
	// endpoints are nodes.
	if len(snl.nodes) < 2 {
		return
	}

	eiPrev := snl.nodes[0]
	for i := 1; i < len(snl.nodes); i++ {
		ei := snl.nodes[i]
		collapsedVertexIndex := snl.findCollapseIndex(eiPrev, ei)
		if collapsedVertexIndex >= 0 {
			*collapsedVertexIndexes = append(*collapsedVertexIndexes, collapsedVertexIndex)
		}
		eiPrev = ei
	}
}

func (snl *Noding_SegmentNodeList) findCollapseIndex(ei0, ei1 *Noding_SegmentNode) int {
	// Only looking for equal nodes.
	if !ei0.Coord.Equals2D(ei1.Coord) {
		return -1
	}

	numVerticesBetween := ei1.SegmentIndex - ei0.SegmentIndex
	if !ei1.IsInterior() {
		numVerticesBetween--
	}

	// If there is a single vertex between the two equal nodes, this is a
	// collapse.
	if numVerticesBetween == 1 {
		return ei0.SegmentIndex + 1
	}
	return -1
}

// AddSplitEdges creates new edges for all the edges that the intersections in
// this list split the parent edge into. Adds the edges to the provided argument
// list (this is so a single list can be used to accumulate all split edges for
// a set of SegmentStrings).
func (snl *Noding_SegmentNodeList) AddSplitEdges(edgeList *[]*Noding_NodedSegmentString) {
	// Ensure that the list has entries for the first and last point of the
	// edge.
	snl.addEndpoints()
	snl.addCollapsedNodes()

	// There should always be at least two entries in the list, since the
	// endpoints are nodes.
	if len(snl.nodes) < 2 {
		return
	}

	eiPrev := snl.nodes[0]
	for i := 1; i < len(snl.nodes); i++ {
		ei := snl.nodes[i]
		newEdge := snl.createSplitEdge(eiPrev, ei)
		*edgeList = append(*edgeList, newEdge)
		eiPrev = ei
	}
}

// createSplitEdge creates a new "split edge" with the section of points between
// (and including) the two intersections. The label for the new edge is the same
// as the label for the parent edge.
func (snl *Noding_SegmentNodeList) createSplitEdge(ei0, ei1 *Noding_SegmentNode) *Noding_NodedSegmentString {
	pts := snl.createSplitEdgePts(ei0, ei1)
	return Noding_NewNodedSegmentString(pts, snl.edge.GetData())
}

// createSplitEdgePts extracts the points for a split edge running between two
// nodes. The extracted points should contain no duplicate points. There should
// always be at least two points extracted (which will be the given nodes).
func (snl *Noding_SegmentNodeList) createSplitEdgePts(ei0, ei1 *Noding_SegmentNode) []*Geom_Coordinate {
	npts := ei1.SegmentIndex - ei0.SegmentIndex + 2

	// If only two points in split edge they must be the node points.
	if npts == 2 {
		return []*Geom_Coordinate{
			Geom_NewCoordinateFromCoordinate(ei0.Coord),
			Geom_NewCoordinateFromCoordinate(ei1.Coord),
		}
	}

	lastSegStartPt := snl.edge.GetCoordinate(ei1.SegmentIndex)
	// If the last intersection point is not equal to its segment start pt, add
	// it to the points list as well. This check is needed because the distance
	// metric is not totally reliable!
	//
	// Also ensure that the created edge always has at least 2 points.
	//
	// The check for point equality is 2D only - Z values are ignored.
	useIntPt1 := ei1.IsInterior() || !ei1.Coord.Equals2D(lastSegStartPt)
	if !useIntPt1 {
		npts--
	}

	pts := make([]*Geom_Coordinate, npts)
	ipt := 0
	pts[ipt] = ei0.Coord.Copy()
	ipt++
	for i := ei0.SegmentIndex + 1; i <= ei1.SegmentIndex; i++ {
		pts[ipt] = snl.edge.GetCoordinate(i)
		ipt++
	}
	if useIntPt1 {
		pts[ipt] = ei1.Coord.Copy()
	}
	return pts
}

// GetSplitCoordinates gets the list of coordinates for the fully noded segment
// string, including all original segment string vertices and vertices
// introduced by nodes in this list. Repeated coordinates are collapsed.
func (snl *Noding_SegmentNodeList) GetSplitCoordinates() []*Geom_Coordinate {
	coordList := Geom_NewCoordinateList()
	// Ensure that the list has entries for the first and last point of the
	// edge.
	snl.addEndpoints()

	// There should always be at least two entries in the list, since the
	// endpoints are nodes.
	if len(snl.nodes) < 2 {
		return coordList.ToCoordinateArray()
	}

	eiPrev := snl.nodes[0]
	for i := 1; i < len(snl.nodes); i++ {
		ei := snl.nodes[i]
		snl.addEdgeCoordinates(eiPrev, ei, coordList)
		eiPrev = ei
	}
	return coordList.ToCoordinateArray()
}

func (snl *Noding_SegmentNodeList) addEdgeCoordinates(ei0, ei1 *Noding_SegmentNode, coordList *Geom_CoordinateList) {
	pts := snl.createSplitEdgePts(ei0, ei1)
	coordList.AddCoordinates(pts, false)
}

// INCOMPLETE inner class - dead code preserved for 1-1 correspondence.
type noding_NodeVertexIterator struct {
	nodeList     *Noding_SegmentNodeList
	edge         *Noding_NodedSegmentString
	nodeIt       []*Noding_SegmentNode
	nodeItIndex  int
	currNode     *Noding_SegmentNode
	nextNode     *Noding_SegmentNode
	currSegIndex int
}

func noding_newNodeVertexIterator(nodeList *Noding_SegmentNodeList) *noding_NodeVertexIterator {
	nvi := &noding_NodeVertexIterator{
		nodeList: nodeList,
		edge:     nodeList.GetEdge(),
		nodeIt:   nodeList.Iterator(),
	}
	nvi.readNextNode()
	return nvi
}

func (nvi *noding_NodeVertexIterator) hasNext() bool {
	return nvi.nextNode != nil
}

func (nvi *noding_NodeVertexIterator) next() any {
	if nvi.currNode == nil {
		nvi.currNode = nvi.nextNode
		nvi.currSegIndex = nvi.currNode.SegmentIndex
		nvi.readNextNode()
		return nvi.currNode
	}
	// Check for trying to read too far.
	if nvi.nextNode == nil {
		return nil
	}

	if nvi.nextNode.SegmentIndex == nvi.currNode.SegmentIndex {
		nvi.currNode = nvi.nextNode
		nvi.currSegIndex = nvi.currNode.SegmentIndex
		nvi.readNextNode()
		return nvi.currNode
	}

	if nvi.nextNode.SegmentIndex > nvi.currNode.SegmentIndex {
		// Incomplete implementation in Java source.
	}
	return nil
}

func (nvi *noding_NodeVertexIterator) readNextNode() {
	if nvi.nodeItIndex < len(nvi.nodeIt) {
		nvi.nextNode = nvi.nodeIt[nvi.nodeItIndex]
		nvi.nodeItIndex++
	} else {
		nvi.nextNode = nil
	}
}
