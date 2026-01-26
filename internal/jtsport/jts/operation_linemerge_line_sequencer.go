package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationLinemerge_LineSequencer builds a sequence from a set of LineStrings so that
// they are ordered end to end.
// A sequence is a complete non-repeating list of the linear
// components of the input. Each linestring is oriented
// so that identical endpoints are adjacent in the list.
//
// A typical use case is to convert a set of
// unoriented geometric links from a linear network
// (e.g. such as block faces on a bus route)
// into a continuous oriented path through the network.
//
// The input linestrings may form one or more connected sets.
// The input linestrings should be correctly noded, or the results may
// not be what is expected.
// The computed output is a single MultiLineString containing the ordered
// linestrings in the sequence.
//
// The sequencing employs the classic Eulerian path graph algorithm.
// Since Eulerian paths are not uniquely determined,
// further rules are used to make the computed sequence preserve as much as possible
// of the input ordering.
// Within a connected subset of lines, the ordering rules are:
//   - If there is degree-1 node which is the start node of an linestring, use that node as the start of the sequence
//   - If there is a degree-1 node which is the end node of an linestring, use that node as the end of the sequence
//   - If the sequence has no degree-1 nodes, use any node as the start
//
// Note that not all arrangements of lines can be sequenced.
// For a connected set of edges in a graph,
// Euler's Theorem states that there is a sequence containing each edge once
// if and only if there are no more than 2 nodes of odd degree.
// If it is not possible to find a sequence, the IsSequenceable method
// will return false.
type OperationLinemerge_LineSequencer struct {
	graph             *OperationLinemerge_LineMergeGraph
	factory           *Geom_GeometryFactory
	lineCount         int
	isRun             bool
	sequencedGeometry *Geom_Geometry
	isSequenceable    bool
}

// OperationLinemerge_NewLineSequencer creates a new LineSequencer.
func OperationLinemerge_NewLineSequencer() *OperationLinemerge_LineSequencer {
	return &OperationLinemerge_LineSequencer{
		graph:   OperationLinemerge_NewLineMergeGraph(),
		factory: Geom_NewGeometryFactoryDefault(),
	}
}

// OperationLinemerge_LineSequencer_Sequence sequences the given geometry.
func OperationLinemerge_LineSequencer_Sequence(geom *Geom_Geometry) *Geom_Geometry {
	sequencer := OperationLinemerge_NewLineSequencer()
	sequencer.AddGeometry(geom)
	return sequencer.GetSequencedLineStrings()
}

// OperationLinemerge_LineSequencer_IsSequenced tests whether a Geometry is sequenced correctly.
// LineStrings are trivially sequenced.
// MultiLineStrings are checked for correct sequencing.
// Otherwise, isSequenced is defined to be true for geometries that are not lineal.
func OperationLinemerge_LineSequencer_IsSequenced(geom *Geom_Geometry) bool {
	if !java.InstanceOf[*Geom_MultiLineString](geom) {
		return true
	}
	mls := java.Cast[*Geom_MultiLineString](geom)

	// The nodes in all subgraphs which have been completely scanned.
	prevSubgraphNodes := make(map[string]bool)

	var lastNode *Geom_Coordinate
	currNodes := make([]*Geom_Coordinate, 0)

	for i := 0; i < mls.GetNumGeometries(); i++ {
		line := java.Cast[*Geom_LineString](mls.GetGeometryN(i))
		startNode := line.GetCoordinateN(0)
		endNode := line.GetCoordinateN(line.GetNumPoints() - 1)

		// If this linestring is connected to a previous subgraph, geom is not sequenced.
		startKey := operationLinemerge_coordKey(startNode)
		endKey := operationLinemerge_coordKey(endNode)
		if prevSubgraphNodes[startKey] {
			return false
		}
		if prevSubgraphNodes[endKey] {
			return false
		}

		if lastNode != nil {
			if !startNode.Equals(lastNode) {
				// Start new connected sequence.
				for _, n := range currNodes {
					prevSubgraphNodes[operationLinemerge_coordKey(n)] = true
				}
				currNodes = currNodes[:0]
			}
		}
		currNodes = append(currNodes, startNode, endNode)
		lastNode = endNode
	}
	return true
}

// operationLinemerge_coordKey returns a string key for a coordinate for use in a map.
func operationLinemerge_coordKey(c *Geom_Coordinate) string {
	// Use a simple string representation for coordinate equality.
	return c.String()
}

// AddGeometries adds a collection of Geometries to be sequenced.
// May be called multiple times.
// Any dimension of Geometry may be added; the constituent linework will be extracted.
func (ls *OperationLinemerge_LineSequencer) AddGeometries(geometries []*Geom_Geometry) {
	for _, geometry := range geometries {
		ls.AddGeometry(geometry)
	}
}

// AddGeometry adds a Geometry to be sequenced.
// May be called multiple times.
// Any dimension of Geometry may be added; the constituent linework will be extracted.
func (ls *OperationLinemerge_LineSequencer) AddGeometry(geometry *Geom_Geometry) {
	filter := operationLinemerge_NewLineSequencerFilter(ls)
	geometry.Apply(filter)
}

// addLine adds a LineString to the graph.
func (ls *OperationLinemerge_LineSequencer) addLine(lineString *Geom_LineString) {
	if ls.factory == nil {
		ls.factory = lineString.GetFactory()
	}
	ls.graph.AddEdge(lineString)
	ls.lineCount++
}

// IsSequenceable tests whether the arrangement of linestrings has a valid sequence.
func (ls *OperationLinemerge_LineSequencer) IsSequenceable() bool {
	ls.computeSequence()
	return ls.isSequenceable
}

// GetSequencedLineStrings returns the LineString or MultiLineString
// built by the sequencing process, if one exists.
func (ls *OperationLinemerge_LineSequencer) GetSequencedLineStrings() *Geom_Geometry {
	ls.computeSequence()
	return ls.sequencedGeometry
}

func (ls *OperationLinemerge_LineSequencer) computeSequence() {
	if ls.isRun {
		return
	}
	ls.isRun = true

	sequences := ls.findSequences()
	if sequences == nil {
		return
	}

	ls.sequencedGeometry = ls.buildSequencedGeometry(sequences)
	ls.isSequenceable = true

	finalLineCount := ls.sequencedGeometry.GetNumGeometries()
	Util_Assert_IsTrueWithMessage(ls.lineCount == finalLineCount, "Lines were missing from result")
	isLineString := java.InstanceOf[*Geom_LineString](ls.sequencedGeometry)
	isMultiLineString := java.InstanceOf[*Geom_MultiLineString](ls.sequencedGeometry)
	Util_Assert_IsTrueWithMessage(isLineString || isMultiLineString, "Result is not lineal")
}

func (ls *OperationLinemerge_LineSequencer) findSequences() [][]*Planargraph_DirectedEdge {
	var sequences [][]*Planargraph_DirectedEdge
	csFinder := PlanargraphAlgorithm_NewConnectedSubgraphFinder(ls.graph.Planargraph_PlanarGraph)
	subgraphs := csFinder.GetConnectedSubgraphs()
	for _, subgraph := range subgraphs {
		if ls.hasSequence(subgraph) {
			seq := ls.findSequence(subgraph)
			sequences = append(sequences, seq)
		} else {
			// If any subgraph cannot be sequenced, abort.
			return nil
		}
	}
	return sequences
}

// hasSequence tests whether a complete unique path exists in a graph using Euler's Theorem.
func (ls *OperationLinemerge_LineSequencer) hasSequence(graph *Planargraph_Subgraph) bool {
	oddDegreeCount := 0
	for _, node := range graph.GetNodes() {
		if node.GetDegree()%2 == 1 {
			oddDegreeCount++
		}
	}
	return oddDegreeCount <= 2
}

func (ls *OperationLinemerge_LineSequencer) findSequence(graph *Planargraph_Subgraph) []*Planargraph_DirectedEdge {
	// Set visited to false on all edges.
	for _, edge := range graph.GetEdges() {
		edge.SetVisited(false)
	}

	startNode := operationLinemerge_findLowestDegreeNode(graph)
	edges := startNode.GetOutEdges().GetEdges()
	startDE := edges[0]
	startDESym := startDE.GetSym()

	seq := make([]*Planargraph_DirectedEdge, 0)
	ls.addReverseSubpath(startDESym, &seq, false)

	// Process the sequence backwards.
	for i := len(seq) - 1; i >= 0; i-- {
		prev := seq[i]
		unvisitedOutDE := operationLinemerge_findUnvisitedBestOrientedDE(prev.GetFromNode())
		if unvisitedOutDE != nil {
			ls.addReverseSubpathAt(unvisitedOutDE.GetSym(), &seq, i+1, true)
		}
	}

	// At this point, we have a valid sequence of graph DirectedEdges, but it
	// is not necessarily appropriately oriented relative to the underlying geometry.
	orientedSeq := ls.orient(seq)
	return orientedSeq
}

// operationLinemerge_findUnvisitedBestOrientedDE finds a DirectedEdge for an unvisited edge (if any),
// choosing the dirEdge which preserves orientation, if possible.
func operationLinemerge_findUnvisitedBestOrientedDE(node *Planargraph_Node) *Planargraph_DirectedEdge {
	var wellOrientedDE *Planargraph_DirectedEdge
	var unvisitedDE *Planargraph_DirectedEdge
	for _, de := range node.GetOutEdges().GetEdges() {
		if !de.GetEdge().IsVisited() {
			unvisitedDE = de
			if de.GetEdgeDirection() {
				wellOrientedDE = de
			}
		}
	}
	if wellOrientedDE != nil {
		return wellOrientedDE
	}
	return unvisitedDE
}

func (ls *OperationLinemerge_LineSequencer) addReverseSubpath(de *Planargraph_DirectedEdge, seq *[]*Planargraph_DirectedEdge, expectedClosed bool) {
	ls.addReverseSubpathAt(de, seq, len(*seq), expectedClosed)
}

func (ls *OperationLinemerge_LineSequencer) addReverseSubpathAt(de *Planargraph_DirectedEdge, seq *[]*Planargraph_DirectedEdge, insertPos int, expectedClosed bool) {
	// Trace an unvisited path *backwards* from this de.
	endNode := de.GetToNode()

	var fromNode *Planargraph_Node
	insertions := make([]*Planargraph_DirectedEdge, 0)
	for {
		insertions = append(insertions, de.GetSym())
		de.GetEdge().SetVisited(true)
		fromNode = de.GetFromNode()
		unvisitedOutDE := operationLinemerge_findUnvisitedBestOrientedDE(fromNode)
		// This must terminate, since we are continually marking edges as visited.
		if unvisitedOutDE == nil {
			break
		}
		de = unvisitedOutDE.GetSym()
	}

	// Insert the collected edges at the insertion position.
	newSeq := make([]*Planargraph_DirectedEdge, 0, len(*seq)+len(insertions))
	newSeq = append(newSeq, (*seq)[:insertPos]...)
	newSeq = append(newSeq, insertions...)
	newSeq = append(newSeq, (*seq)[insertPos:]...)
	*seq = newSeq

	if expectedClosed {
		// The path should end at the toNode of this de, otherwise we have an error.
		Util_Assert_IsTrueWithMessage(fromNode == endNode, "path not contiguous")
	}
}

func operationLinemerge_findLowestDegreeNode(graph *Planargraph_Subgraph) *Planargraph_Node {
	minDegree := math.MaxInt
	var minDegreeNode *Planargraph_Node
	for _, node := range graph.GetNodes() {
		if minDegreeNode == nil || node.GetDegree() < minDegree {
			minDegree = node.GetDegree()
			minDegreeNode = node
		}
	}
	return minDegreeNode
}

// orient computes a version of the sequence which is optimally
// oriented relative to the underlying geometry.
func (ls *OperationLinemerge_LineSequencer) orient(seq []*Planargraph_DirectedEdge) []*Planargraph_DirectedEdge {
	startEdge := seq[0]
	endEdge := seq[len(seq)-1]
	startNode := startEdge.GetFromNode()
	endNode := endEdge.GetToNode()

	flipSeq := false
	hasDegree1Node := startNode.GetDegree() == 1 || endNode.GetDegree() == 1

	if hasDegree1Node {
		hasObviousStartNode := false

		// Test end edge before start edge, to make result stable
		// (ie. if both are good starts, pick the actual start).
		if endEdge.GetToNode().GetDegree() == 1 && !endEdge.GetEdgeDirection() {
			hasObviousStartNode = true
			flipSeq = true
		}
		if startEdge.GetFromNode().GetDegree() == 1 && startEdge.GetEdgeDirection() {
			hasObviousStartNode = true
			flipSeq = false
		}

		// Since there is no obvious start node, use any node of degree 1.
		if !hasObviousStartNode {
			// Check if the start node should actually be the end node.
			if startEdge.GetFromNode().GetDegree() == 1 {
				flipSeq = true
			}
			// If the end node is of degree 1, it is properly the end node.
		}
	}

	// If there is no degree 1 node, just use the sequence as is.
	// (Could insert heuristic of taking direction of majority of lines as overall direction.)

	if flipSeq {
		return ls.reverseSequence(seq)
	}
	return seq
}

// reverseSequence reverses the sequence.
// This requires reversing the order of the dirEdges, and flipping each dirEdge as well.
func (ls *OperationLinemerge_LineSequencer) reverseSequence(seq []*Planargraph_DirectedEdge) []*Planargraph_DirectedEdge {
	newSeq := make([]*Planargraph_DirectedEdge, len(seq))
	for i, de := range seq {
		newSeq[len(seq)-1-i] = de.GetSym()
	}
	return newSeq
}

// buildSequencedGeometry builds a geometry (LineString or MultiLineString) representing the sequence.
func (ls *OperationLinemerge_LineSequencer) buildSequencedGeometry(sequences [][]*Planargraph_DirectedEdge) *Geom_Geometry {
	var lines []*Geom_Geometry

	for _, seq := range sequences {
		for _, de := range seq {
			lineMergeEdge := java.Cast[*OperationLinemerge_LineMergeEdge](de.GetEdge())
			line := lineMergeEdge.GetLine()

			lineToAdd := line
			if !de.GetEdgeDirection() && !line.IsClosed() {
				lineToAdd = operationLinemerge_reverseLine(line)
			}

			lines = append(lines, lineToAdd.Geom_Geometry)
		}
	}

	if len(lines) == 0 {
		return ls.factory.CreateMultiLineStringFromLineStrings([]*Geom_LineString{}).Geom_GeometryCollection.Geom_Geometry
	}
	return ls.factory.BuildGeometry(lines)
}

func operationLinemerge_reverseLine(line *Geom_LineString) *Geom_LineString {
	pts := line.GetCoordinates()
	revPts := make([]*Geom_Coordinate, len(pts))
	length := len(pts)
	for i := 0; i < length; i++ {
		revPts[length-1-i] = Geom_NewCoordinateFromCoordinate(pts[i])
	}
	return line.GetFactory().CreateLineStringFromCoordinates(revPts)
}

// operationLinemerge_LineSequencerFilter is a filter that extracts LineStrings from a geometry.
type operationLinemerge_LineSequencerFilter struct {
	ls *OperationLinemerge_LineSequencer
}

var _ Geom_GeometryComponentFilter = (*operationLinemerge_LineSequencerFilter)(nil)

func (f *operationLinemerge_LineSequencerFilter) IsGeom_GeometryComponentFilter() {}

func operationLinemerge_NewLineSequencerFilter(ls *OperationLinemerge_LineSequencer) *operationLinemerge_LineSequencerFilter {
	return &operationLinemerge_LineSequencerFilter{ls: ls}
}

func (f *operationLinemerge_LineSequencerFilter) Filter(geom *Geom_Geometry) {
	if java.InstanceOf[*Geom_LineString](geom) {
		f.ls.addLine(java.Cast[*Geom_LineString](geom))
	}
}
