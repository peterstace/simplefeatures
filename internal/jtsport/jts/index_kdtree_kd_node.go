package jts

// IndexKdtree_KdNode represents a node of a KdTree, which represents one or
// more points in the same location.
type IndexKdtree_KdNode struct {
	p     *Geom_Coordinate
	data  any
	left  *IndexKdtree_KdNode
	right *IndexKdtree_KdNode
	count int
}

// IndexKdtree_NewKdNodeFromXY creates a new KdNode from x,y coordinates.
func IndexKdtree_NewKdNodeFromXY(x, y float64, data any) *IndexKdtree_KdNode {
	return &IndexKdtree_KdNode{
		p:     Geom_NewCoordinateWithXY(x, y),
		data:  data,
		count: 1,
	}
}

// IndexKdtree_NewKdNode creates a new KdNode from a coordinate.
func IndexKdtree_NewKdNode(p *Geom_Coordinate, data any) *IndexKdtree_KdNode {
	return &IndexKdtree_KdNode{
		p:     Geom_NewCoordinateFromCoordinate(p),
		data:  data,
		count: 1,
	}
}

// GetX returns the X coordinate of the node.
func (n *IndexKdtree_KdNode) GetX() float64 {
	return n.p.GetX()
}

// GetY returns the Y coordinate of the node.
func (n *IndexKdtree_KdNode) GetY() float64 {
	return n.p.GetY()
}

// SplitValue gets the split value at a node, depending on whether the node
// splits on X or Y. The X (or Y) ordinates of all points in the left subtree
// are less than the split value, and those in the right subtree are greater
// than or equal to the split value.
func (n *IndexKdtree_KdNode) SplitValue(isSplitOnX bool) float64 {
	if isSplitOnX {
		return n.p.GetX()
	}
	return n.p.GetY()
}

// GetCoordinate returns the location of this node.
func (n *IndexKdtree_KdNode) GetCoordinate() *Geom_Coordinate {
	return n.p
}

// GetData gets the user data object associated with this node.
func (n *IndexKdtree_KdNode) GetData() any {
	return n.data
}

// GetLeft returns the left node of the tree.
func (n *IndexKdtree_KdNode) GetLeft() *IndexKdtree_KdNode {
	return n.left
}

// GetRight returns the right node of the tree.
func (n *IndexKdtree_KdNode) GetRight() *IndexKdtree_KdNode {
	return n.right
}

// Increment increments the count of points at this location.
func (n *IndexKdtree_KdNode) Increment() {
	n.count++
}

// GetCount returns the number of inserted points that are coincident at this
// location.
func (n *IndexKdtree_KdNode) GetCount() int {
	return n.count
}

// IsRepeated tests whether more than one point with this value have been
// inserted (up to the tolerance).
func (n *IndexKdtree_KdNode) IsRepeated() bool {
	return n.count > 1
}

// SetLeft sets the left node value.
func (n *IndexKdtree_KdNode) SetLeft(left *IndexKdtree_KdNode) {
	n.left = left
}

// SetRight sets the right node value.
func (n *IndexKdtree_KdNode) SetRight(right *IndexKdtree_KdNode) {
	n.right = right
}

// IsRangeOverLeft tests whether the node's left subtree may contain values in
// a given range envelope.
func (n *IndexKdtree_KdNode) IsRangeOverLeft(isSplitOnX bool, env *Geom_Envelope) bool {
	var envMin float64
	if isSplitOnX {
		envMin = env.GetMinX()
	} else {
		envMin = env.GetMinY()
	}
	splitValue := n.SplitValue(isSplitOnX)
	return envMin < splitValue
}

// IsRangeOverRight tests whether the node's right subtree may contain values
// in a given range envelope.
func (n *IndexKdtree_KdNode) IsRangeOverRight(isSplitOnX bool, env *Geom_Envelope) bool {
	var envMax float64
	if isSplitOnX {
		envMax = env.GetMaxX()
	} else {
		envMax = env.GetMaxY()
	}
	splitValue := n.SplitValue(isSplitOnX)
	return splitValue <= envMax
}

// IsPointOnLeft tests whether a point is strictly to the left of the splitting
// plane for this node. If so it may be in the left subtree of this node,
// otherwise, the point may be in the right subtree. The point is to the left
// if its X (or Y) ordinate is less than the split value.
func (n *IndexKdtree_KdNode) IsPointOnLeft(isSplitOnX bool, pt *Geom_Coordinate) bool {
	var ptOrdinate float64
	if isSplitOnX {
		ptOrdinate = pt.GetX()
	} else {
		ptOrdinate = pt.GetY()
	}
	splitValue := n.SplitValue(isSplitOnX)
	return ptOrdinate < splitValue
}
