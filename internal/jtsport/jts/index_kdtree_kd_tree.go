package jts

// IndexKdtree_KdNodeVisitor is a visitor for nodes in a KdTree.
type IndexKdtree_KdNodeVisitor interface {
	Visit(node *IndexKdtree_KdNode)
}

// IndexKdtree_KdTree is an implementation of a KD-Tree over two dimensions
// (X and Y). KD-trees provide fast range searching and fast lookup for point
// data. The tree is built dynamically by inserting points. The tree supports
// queries by range and for point equality.
//
// This implementation supports detecting and snapping points which are closer
// than a given distance tolerance. If the same point (up to tolerance) is
// inserted more than once, it is snapped to the existing node. When an
// inserted point is snapped to a node then a new node is not created but the
// count of the existing node is incremented.
type IndexKdtree_KdTree struct {
	root          *IndexKdtree_KdNode
	numberOfNodes int64
	tolerance     float64
}

// IndexKdtree_NewKdTree creates a new instance of a KdTree with a snapping
// tolerance of 0.0. (I.e. distinct points will not be snapped.)
func IndexKdtree_NewKdTree() *IndexKdtree_KdTree {
	return IndexKdtree_NewKdTreeWithTolerance(0.0)
}

// IndexKdtree_NewKdTreeWithTolerance creates a new instance of a KdTree,
// specifying a snapping distance tolerance. Points which lie closer than the
// tolerance to a point already in the tree will be treated as identical to the
// existing point.
func IndexKdtree_NewKdTreeWithTolerance(tolerance float64) *IndexKdtree_KdTree {
	return &IndexKdtree_KdTree{
		tolerance: tolerance,
	}
}

// IndexKdtree_KdTree_ToCoordinates converts a collection of KdNodes to an
// array of Coordinates.
func IndexKdtree_KdTree_ToCoordinates(kdnodes []*IndexKdtree_KdNode) []*Geom_Coordinate {
	return IndexKdtree_KdTree_ToCoordinatesIncludeRepeated(kdnodes, false)
}

// IndexKdtree_KdTree_ToCoordinatesIncludeRepeated converts a collection of
// KdNodes to an array of Coordinates, specifying whether repeated nodes should
// be represented by multiple coordinates.
func IndexKdtree_KdTree_ToCoordinatesIncludeRepeated(kdnodes []*IndexKdtree_KdNode, includeRepeated bool) []*Geom_Coordinate {
	coords := make([]*Geom_Coordinate, 0)
	for _, node := range kdnodes {
		count := 1
		if includeRepeated {
			count = node.GetCount()
		}
		for i := 0; i < count; i++ {
			coords = append(coords, node.GetCoordinate())
		}
	}
	return coords
}

// GetRoot gets the root node of this tree.
func (t *IndexKdtree_KdTree) GetRoot() *IndexKdtree_KdNode {
	return t.root
}

// IsEmpty tests whether the index contains any items.
func (t *IndexKdtree_KdTree) IsEmpty() bool {
	return t.root == nil
}

// Insert inserts a new point in the kd-tree, with no data.
func (t *IndexKdtree_KdTree) Insert(p *Geom_Coordinate) *IndexKdtree_KdNode {
	return t.InsertWithData(p, nil)
}

// InsertWithData inserts a new point into the kd-tree.
func (t *IndexKdtree_KdTree) InsertWithData(p *Geom_Coordinate, data any) *IndexKdtree_KdNode {
	if t.root == nil {
		t.root = IndexKdtree_NewKdNode(p, data)
		return t.root
	}

	// Check if the point is already in the tree, up to tolerance.
	// If tolerance is zero, this phase of the insertion can be skipped.
	if t.tolerance > 0 {
		matchNode := t.findBestMatchNode(p)
		if matchNode != nil {
			// Point already in index - increment counter.
			matchNode.Increment()
			return matchNode
		}
	}

	return t.insertExact(p, data)
}

// findBestMatchNode finds the node in the tree which is the best match for a
// point being inserted. The match is made deterministic by returning the
// lowest of any nodes which lie the same distance from the point.
func (t *IndexKdtree_KdTree) findBestMatchNode(p *Geom_Coordinate) *IndexKdtree_KdNode {
	visitor := &indexKdtree_bestMatchVisitor{
		p:         p,
		tolerance: t.tolerance,
	}
	t.QueryEnvelopeVisitor(visitor.queryEnvelope(), visitor)
	return visitor.matchNode
}

type indexKdtree_bestMatchVisitor struct {
	tolerance float64
	matchNode *IndexKdtree_KdNode
	matchDist float64
	p         *Geom_Coordinate
}

func (v *indexKdtree_bestMatchVisitor) queryEnvelope() *Geom_Envelope {
	queryEnv := Geom_NewEnvelopeFromCoordinate(v.p)
	queryEnv.ExpandBy(v.tolerance)
	return queryEnv
}

func (v *indexKdtree_bestMatchVisitor) Visit(node *IndexKdtree_KdNode) {
	dist := v.p.Distance(node.GetCoordinate())
	isInTolerance := dist <= v.tolerance
	if !isInTolerance {
		return
	}
	update := false
	if v.matchNode == nil ||
		dist < v.matchDist ||
		// If distances are the same, record the lesser coordinate.
		(v.matchNode != nil && dist == v.matchDist &&
			node.GetCoordinate().CompareTo(v.matchNode.GetCoordinate()) < 1) {
		update = true
	}
	if update {
		v.matchNode = node
		v.matchDist = dist
	}
}

// insertExact inserts a point known to be beyond the distance tolerance of
// any existing node. The point is inserted at the bottom of the exact
// splitting path, so that tree shape is deterministic.
func (t *IndexKdtree_KdTree) insertExact(p *Geom_Coordinate, data any) *IndexKdtree_KdNode {
	currentNode := t.root
	var leafNode *IndexKdtree_KdNode
	isXLevel := true
	isLessThan := true

	// Traverse the tree, first cutting the plane left-right (by X ordinate)
	// then top-bottom (by Y ordinate).
	for currentNode != nil {
		isInTolerance := p.Distance(currentNode.GetCoordinate()) <= t.tolerance

		// Check if point is already in tree (up to tolerance) and if so simply
		// return existing node.
		if isInTolerance {
			currentNode.Increment()
			return currentNode
		}

		splitValue := currentNode.SplitValue(isXLevel)
		if isXLevel {
			isLessThan = p.GetX() < splitValue
		} else {
			isLessThan = p.GetY() < splitValue
		}
		leafNode = currentNode
		if isLessThan {
			currentNode = currentNode.GetLeft()
		} else {
			currentNode = currentNode.GetRight()
		}

		isXLevel = !isXLevel
	}

	// No node found, add new leaf node to tree.
	t.numberOfNodes++
	node := IndexKdtree_NewKdNode(p, data)
	if isLessThan {
		leafNode.SetLeft(node)
	} else {
		leafNode.SetRight(node)
	}
	return node
}

// QueryEnvelopeVisitor performs a range search of the points in the index and
// visits all nodes found.
func (t *IndexKdtree_KdTree) QueryEnvelopeVisitor(queryEnv *Geom_Envelope, visitor IndexKdtree_KdNodeVisitor) {
	type queryStackFrame struct {
		node     *IndexKdtree_KdNode
		isXLevel bool
	}

	queryStack := make([]queryStackFrame, 0)
	currentNode := t.root
	isXLevel := true

	// Search is computed via in-order traversal.
	for {
		if currentNode != nil {
			queryStack = append(queryStack, queryStackFrame{node: currentNode, isXLevel: isXLevel})

			searchLeft := currentNode.IsRangeOverLeft(isXLevel, queryEnv)
			if searchLeft {
				currentNode = currentNode.GetLeft()
				if currentNode != nil {
					isXLevel = !isXLevel
				}
			} else {
				currentNode = nil
			}
		} else if len(queryStack) > 0 {
			// currentNode is empty, so pop stack.
			frame := queryStack[len(queryStack)-1]
			queryStack = queryStack[:len(queryStack)-1]
			currentNode = frame.node
			isXLevel = frame.isXLevel

			// Check if search matches current node.
			if queryEnv.ContainsCoordinate(currentNode.GetCoordinate()) {
				visitor.Visit(currentNode)
			}

			searchRight := currentNode.IsRangeOverRight(isXLevel, queryEnv)
			if searchRight {
				currentNode = currentNode.GetRight()
				if currentNode != nil {
					isXLevel = !isXLevel
				}
			} else {
				currentNode = nil
			}
		} else {
			// Stack is empty and no current node.
			return
		}
	}
}

// QueryEnvelope performs a range search of the points in the index.
func (t *IndexKdtree_KdTree) QueryEnvelope(queryEnv *Geom_Envelope) []*IndexKdtree_KdNode {
	result := make([]*IndexKdtree_KdNode, 0)
	t.QueryEnvelopeVisitor(queryEnv, &indexKdtree_listVisitor{result: &result})
	return result
}

type indexKdtree_listVisitor struct {
	result *[]*IndexKdtree_KdNode
}

func (v *indexKdtree_listVisitor) Visit(node *IndexKdtree_KdNode) {
	*v.result = append(*v.result, node)
}

// QueryPoint searches for a given point in the index and returns its node if
// found.
func (t *IndexKdtree_KdTree) QueryPoint(queryPt *Geom_Coordinate) *IndexKdtree_KdNode {
	currentNode := t.root
	isXLevel := true

	for currentNode != nil {
		if currentNode.GetCoordinate().Equals2D(queryPt) {
			return currentNode
		}

		searchLeft := currentNode.IsPointOnLeft(isXLevel, queryPt)
		if searchLeft {
			currentNode = currentNode.GetLeft()
		} else {
			currentNode = currentNode.GetRight()
		}
		isXLevel = !isXLevel
	}
	// Point not found.
	return nil
}

// Depth computes the depth of the tree.
func (t *IndexKdtree_KdTree) Depth() int {
	return t.depthNode(t.root)
}

func (t *IndexKdtree_KdTree) depthNode(currentNode *IndexKdtree_KdNode) int {
	if currentNode == nil {
		return 0
	}

	dL := t.depthNode(currentNode.GetLeft())
	dR := t.depthNode(currentNode.GetRight())
	if dL > dR {
		return 1 + dL
	}
	return 1 + dR
}

// Size computes the size (number of items) in the tree.
func (t *IndexKdtree_KdTree) Size() int {
	return t.sizeNode(t.root)
}

func (t *IndexKdtree_KdTree) sizeNode(currentNode *IndexKdtree_KdNode) int {
	if currentNode == nil {
		return 0
	}

	sizeL := t.sizeNode(currentNode.GetLeft())
	sizeR := t.sizeNode(currentNode.GetRight())
	return 1 + sizeL + sizeR
}
