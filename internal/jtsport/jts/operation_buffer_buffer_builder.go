package jts

import "sort"

// operationBuffer_BufferBuilder builds the buffer geometry for a given input geometry and precision model.
// Allows setting the level of approximation for circular arcs,
// and the precision model in which to carry out the computation.
//
// When computing buffers in floating point double-precision
// it can happen that the process of iterated noding can fail to converge (terminate).
// In this case a TopologyException will be thrown.
// Retrying the computation in a fixed precision
// can produce more robust results.
type operationBuffer_BufferBuilder struct {
	bufParams *OperationBuffer_BufferParameters

	workingPrecisionModel *Geom_PrecisionModel
	workingNoder          Noding_Noder
	geomFact              *Geom_GeometryFactory
	graph                 *Geomgraph_PlanarGraph
	edgeList              *Geomgraph_EdgeList

	isInvertOrientation bool
}

// operationBuffer_newBufferBuilder creates a new BufferBuilder using the given parameters.
func operationBuffer_newBufferBuilder(bufParams *OperationBuffer_BufferParameters) *operationBuffer_BufferBuilder {
	return &operationBuffer_BufferBuilder{
		bufParams: bufParams,
		edgeList:  Geomgraph_NewEdgeList(),
	}
}

// operationBuffer_bufferBuilder_depthDelta computes the change in depth as an edge is crossed from R to L.
func operationBuffer_bufferBuilder_depthDelta(label *Geomgraph_Label) int {
	lLoc := label.GetLocation(0, Geom_Position_Left)
	rLoc := label.GetLocation(0, Geom_Position_Right)
	if lLoc == Geom_Location_Interior && rLoc == Geom_Location_Exterior {
		return 1
	} else if lLoc == Geom_Location_Exterior && rLoc == Geom_Location_Interior {
		return -1
	}
	return 0
}

// SetWorkingPrecisionModel sets the precision model to use during the curve computation and noding,
// if it is different to the precision model of the Geometry.
// If the precision model is less than the precision of the Geometry precision model,
// the Geometry must have previously been rounded to that precision.
func (bb *operationBuffer_BufferBuilder) SetWorkingPrecisionModel(pm *Geom_PrecisionModel) {
	bb.workingPrecisionModel = pm
}

// SetNoder sets the Noder to use during noding.
// This allows choosing fast but non-robust noding, or slower but robust noding.
func (bb *operationBuffer_BufferBuilder) SetNoder(noder Noding_Noder) {
	bb.workingNoder = noder
}

// SetInvertOrientation sets whether the offset curve is generated
// using the inverted orientation of input rings.
// This allows generating a buffer(0) polygon from the smaller lobes
// of self-crossing rings.
func (bb *operationBuffer_BufferBuilder) SetInvertOrientation(isInvertOrientation bool) {
	bb.isInvertOrientation = isInvertOrientation
}

// Buffer computes the buffer for a geometry.
func (bb *operationBuffer_BufferBuilder) Buffer(g *Geom_Geometry, distance float64) *Geom_Geometry {
	precisionModel := bb.workingPrecisionModel
	if precisionModel == nil {
		precisionModel = g.GetPrecisionModel()
	}

	// factory must be the same as the one used by the input
	bb.geomFact = g.GetFactory()

	curveSetBuilder := OperationBuffer_NewBufferCurveSetBuilder(g, distance, precisionModel, bb.bufParams)
	curveSetBuilder.SetInvertOrientation(bb.isInvertOrientation)

	bufferSegStrList := curveSetBuilder.GetCurves()

	// short-circuit test
	if len(bufferSegStrList) <= 0 {
		return bb.createEmptyResultGeometry()
	}

	// Currently only zero-distance buffers are validated,
	// to avoid reducing performance for other buffers.
	// This fixes some noding failure cases found via GeometryFixer
	// (see JTS-852).
	isNodingValidated := distance == 0.0
	bb.computeNodedEdges(bufferSegStrList, precisionModel, isNodingValidated)

	bb.graph = Geomgraph_NewPlanarGraph(OperationOverlay_NewOverlayNodeFactory().Geomgraph_NodeFactory)
	bb.graph.AddEdges(bb.edgeList.GetEdges())

	subgraphList := bb.createSubgraphs(bb.graph)
	polyBuilder := OperationOverlay_NewPolygonBuilder(bb.geomFact)
	bb.buildSubgraphs(subgraphList, polyBuilder)
	resultPolyList := polyBuilder.GetPolygons()

	// just in case...
	if len(resultPolyList) <= 0 {
		return bb.createEmptyResultGeometry()
	}

	// Convert polygons to geometries for BuildGeometry
	geomList := make([]*Geom_Geometry, len(resultPolyList))
	for i, poly := range resultPolyList {
		geomList[i] = poly.Geom_Geometry
	}
	resultGeom := bb.geomFact.BuildGeometry(geomList)
	return resultGeom
}

func (bb *operationBuffer_BufferBuilder) getNoder(precisionModel *Geom_PrecisionModel) Noding_Noder {
	if bb.workingNoder != nil {
		return bb.workingNoder
	}

	// otherwise use a fast (but non-robust) noder
	noder := Noding_NewMCIndexNoder()
	li := Algorithm_NewRobustLineIntersector()
	li.SetPrecisionModel(precisionModel)
	noder.SetSegmentIntersector(Noding_NewIntersectionAdder(li.Algorithm_LineIntersector))
	return noder
}

func (bb *operationBuffer_BufferBuilder) computeNodedEdges(bufferSegStrList []Noding_SegmentString, precisionModel *Geom_PrecisionModel, isNodingValidated bool) {
	noder := bb.getNoder(precisionModel)
	noder.ComputeNodes(bufferSegStrList)
	nodedSegStrings := noder.GetNodedSubstrings()

	if isNodingValidated {
		nv := Noding_NewFastNodingValidator(nodedSegStrings)
		nv.CheckValid()
	}

	for _, segStr := range nodedSegStrings {
		// Discard edges which have zero length,
		// since they carry no information and cause problems with topology building
		pts := segStr.GetCoordinates()
		if len(pts) == 2 && pts[0].Equals2D(pts[1]) {
			continue
		}

		oldLabel := segStr.GetData().(*Geomgraph_Label)
		edge := Geomgraph_NewEdge(segStr.GetCoordinates(), Geomgraph_NewLabelFromLabel(oldLabel))
		bb.insertUniqueEdge(edge)
	}
}

// insertUniqueEdge inserts edges, checking to see if an identical edge already exists.
// If so, the edge is not inserted, but its label is merged
// with the existing edge.
func (bb *operationBuffer_BufferBuilder) insertUniqueEdge(e *Geomgraph_Edge) {
	// fast lookup
	existingEdge := bb.edgeList.FindEqualEdge(e)

	// If an identical edge already exists, simply update its label
	if existingEdge != nil {
		existingLabel := existingEdge.GetLabel()

		labelToMerge := e.GetLabel()
		// check if new edge is in reverse direction to existing edge
		// if so, must flip the label before merging it
		if !existingEdge.IsPointwiseEqual(e) {
			labelToMerge = Geomgraph_NewLabelFromLabel(e.GetLabel())
			labelToMerge.Flip()
		}
		existingLabel.Merge(labelToMerge)

		// compute new depth delta of sum of edges
		mergeDelta := operationBuffer_bufferBuilder_depthDelta(labelToMerge)
		existingDelta := existingEdge.GetDepthDelta()
		newDelta := existingDelta + mergeDelta
		existingEdge.SetDepthDelta(newDelta)
	} else { // no matching existing edge was found
		// add this new edge to the list of edges in this graph
		bb.edgeList.Add(e)
		e.SetDepthDelta(operationBuffer_bufferBuilder_depthDelta(e.GetLabel()))
	}
}

func (bb *operationBuffer_BufferBuilder) createSubgraphs(graph *Geomgraph_PlanarGraph) []*OperationBuffer_BufferSubgraph {
	subgraphList := make([]*OperationBuffer_BufferSubgraph, 0)
	for _, node := range graph.GetNodes() {
		if !node.IsVisited() {
			subgraph := OperationBuffer_NewBufferSubgraph()
			subgraph.Create(node)
			subgraphList = append(subgraphList, subgraph)
		}
	}
	// Sort the subgraphs in descending order of their rightmost coordinate.
	// This ensures that when the Polygons for the subgraphs are built,
	// subgraphs for shells will have been built before the subgraphs for
	// any holes they contain.
	sort.Slice(subgraphList, func(i, j int) bool {
		return subgraphList[i].CompareTo(subgraphList[j]) > 0
	})
	return subgraphList
}

// buildSubgraphs completes the building of the input subgraphs by depth-labelling them,
// and adds them to the PolygonBuilder.
// The subgraph list must be sorted in rightmost-coordinate order.
func (bb *operationBuffer_BufferBuilder) buildSubgraphs(subgraphList []*OperationBuffer_BufferSubgraph, polyBuilder *OperationOverlay_PolygonBuilder) {
	processedGraphs := make([]*OperationBuffer_BufferSubgraph, 0)
	for _, subgraph := range subgraphList {
		p := subgraph.GetRightmostCoordinate()
		locater := operationBuffer_newSubgraphDepthLocater(processedGraphs)
		outsideDepth := locater.GetDepth(p)
		subgraph.ComputeDepth(outsideDepth)
		subgraph.FindResultEdges()
		processedGraphs = append(processedGraphs, subgraph)
		// Convert DirectedEdges to EdgeEnds for PolygonBuilder.Add
		dirEdges := subgraph.GetDirectedEdges()
		edgeEnds := make([]*Geomgraph_EdgeEnd, len(dirEdges))
		for i, de := range dirEdges {
			edgeEnds[i] = de.Geomgraph_EdgeEnd
		}
		polyBuilder.Add(edgeEnds, subgraph.GetNodes())
	}
}

// TRANSLITERATION NOTE: This method is dead code in Java but included for 1-1 correspondence.
// The Java version takes an Iterator<SegmentString>; we use a slice here for simplicity
// since this is dead code. If this were live code, we'd need to model the Iterator properly.
func operationBuffer_bufferBuilder_convertSegStrings(segStrings []Noding_SegmentString) *Geom_Geometry {
	fact := Geom_NewGeometryFactoryDefault()
	lines := make([]*Geom_Geometry, 0)
	for _, ss := range segStrings {
		line := fact.CreateLineStringFromCoordinates(ss.GetCoordinates())
		lines = append(lines, line.Geom_Geometry)
	}
	return fact.BuildGeometry(lines)
}

// createEmptyResultGeometry gets the standard result for an empty buffer.
// Since buffer always returns a polygonal result,
// this is chosen to be an empty polygon.
func (bb *operationBuffer_BufferBuilder) createEmptyResultGeometry() *Geom_Geometry {
	emptyGeom := bb.geomFact.CreatePolygon()
	return emptyGeom.Geom_Geometry
}
