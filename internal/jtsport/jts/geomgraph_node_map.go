package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geomgraph_NodeMap is a map of nodes, indexed by the coordinate of the node.
type Geomgraph_NodeMap struct {
	child    java.Polymorphic
	nodeMap  map[geomgraph_NodeMap_CoordKey]*Geomgraph_Node
	nodeFact *Geomgraph_NodeFactory
}

// geomgraph_NodeMap_CoordKey is used as a map key for coordinates.
type geomgraph_NodeMap_CoordKey struct {
	x, y float64
}

func geomgraph_NodeMap_makeKey(c *Geom_Coordinate) geomgraph_NodeMap_CoordKey {
	return geomgraph_NodeMap_CoordKey{x: c.GetX(), y: c.GetY()}
}

// GetChild returns the immediate child in the type hierarchy chain.
func (nm *Geomgraph_NodeMap) GetChild() java.Polymorphic {
	return nm.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (nm *Geomgraph_NodeMap) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewNodeMap creates a new NodeMap with the given NodeFactory.
func Geomgraph_NewNodeMap(nodeFact *Geomgraph_NodeFactory) *Geomgraph_NodeMap {
	return &Geomgraph_NodeMap{
		nodeMap:  make(map[geomgraph_NodeMap_CoordKey]*Geomgraph_Node),
		nodeFact: nodeFact,
	}
}

// AddNodeFromCoord adds a node for the given coordinate. This method expects
// that a node has a coordinate value.
func (nm *Geomgraph_NodeMap) AddNodeFromCoord(coord *Geom_Coordinate) *Geomgraph_Node {
	key := geomgraph_NodeMap_makeKey(coord)
	node, exists := nm.nodeMap[key]
	if !exists {
		node = nm.nodeFact.CreateNode(coord)
		nm.nodeMap[key] = node
	}
	return node
}

// AddNode adds a node to the map, merging labels if a node already exists at
// that coordinate.
func (nm *Geomgraph_NodeMap) AddNode(n *Geomgraph_Node) *Geomgraph_Node {
	key := geomgraph_NodeMap_makeKey(n.GetCoordinate_BODY())
	node, exists := nm.nodeMap[key]
	if !exists {
		nm.nodeMap[key] = n
		return n
	}
	node.MergeLabel(n)
	return node
}

// Add adds a node for the start point of this EdgeEnd (if one does not already
// exist in this map). Adds the EdgeEnd to the (possibly new) node.
func (nm *Geomgraph_NodeMap) Add(e *Geomgraph_EdgeEnd) {
	p := e.p0
	n := nm.AddNodeFromCoord(p)
	n.Add(e)
}

// Find returns the node for the given coordinate, or nil if not found.
func (nm *Geomgraph_NodeMap) Find(coord *Geom_Coordinate) *Geomgraph_Node {
	key := geomgraph_NodeMap_makeKey(coord)
	return nm.nodeMap[key]
}

// Values returns all nodes in the map.
func (nm *Geomgraph_NodeMap) Values() []*Geomgraph_Node {
	result := make([]*Geomgraph_Node, 0, len(nm.nodeMap))
	for _, node := range nm.nodeMap {
		result = append(result, node)
	}
	return result
}

// GetBoundaryNodes returns all nodes that are on the boundary for the given
// geometry index.
func (nm *Geomgraph_NodeMap) GetBoundaryNodes(geomIndex int) []*Geomgraph_Node {
	var bdyNodes []*Geomgraph_Node
	for _, node := range nm.nodeMap {
		if node.GetLabel().GetLocationOn(geomIndex) == Geom_Location_Boundary {
			bdyNodes = append(bdyNodes, node)
		}
	}
	return bdyNodes
}
