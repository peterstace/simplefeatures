package jts

import (
	"math"
	"sort"
)

// planargraph_CoordKey is a map key for coordinates that handles NaN values correctly.
// In Go, NaN != NaN, so using Geom_Coordinate directly as a map key fails when
// coordinates have NaN values (e.g., in the Z ordinate). This key type uses the
// bit representation of floats, where NaN values are normalized to a consistent
// pattern, similar to Java's Double.doubleToLongBits().
type planargraph_CoordKey struct {
	xBits, yBits, zBits uint64
}

// planargraph_coordToKey converts a Coordinate to a map key that handles NaN correctly.
func planargraph_coordToKey(c *Geom_Coordinate) planargraph_CoordKey {
	return planargraph_CoordKey{
		xBits: planargraph_normalizeNaN(c.X),
		yBits: planargraph_normalizeNaN(c.Y),
		zBits: planargraph_normalizeNaN(c.Z),
	}
}

// planargraph_normalizeNaN converts a float64 to its bit representation,
// normalizing all NaN values to a single canonical NaN bit pattern.
// This mimics Java's Double.doubleToLongBits() behavior.
func planargraph_normalizeNaN(v float64) uint64 {
	if math.IsNaN(v) {
		// Use a canonical NaN bit pattern (same as Java's canonical NaN).
		return 0x7ff8000000000000
	}
	return math.Float64bits(v)
}

// Planargraph_NodeMap is a map of Nodes, indexed by the coordinate of the node.
type Planargraph_NodeMap struct {
	nodeMap map[planargraph_CoordKey]*Planargraph_Node
}

// Planargraph_NewNodeMap constructs a NodeMap without any Nodes.
func Planargraph_NewNodeMap() *Planargraph_NodeMap {
	return &Planargraph_NodeMap{
		nodeMap: make(map[planargraph_CoordKey]*Planargraph_Node),
	}
}

// Add adds a node to the map, replacing any that is already at that location.
// Returns the added node.
func (nm *Planargraph_NodeMap) Add(n *Planargraph_Node) *Planargraph_Node {
	key := planargraph_coordToKey(n.GetCoordinate())
	nm.nodeMap[key] = n
	return n
}

// Remove removes the Node at the given location, and returns it (or nil if no Node was there).
func (nm *Planargraph_NodeMap) Remove(pt *Geom_Coordinate) *Planargraph_Node {
	key := planargraph_coordToKey(pt)
	node := nm.nodeMap[key]
	delete(nm.nodeMap, key)
	return node
}

// Find returns the Node at the given location, or nil if no Node was there.
func (nm *Planargraph_NodeMap) Find(coord *Geom_Coordinate) *Planargraph_Node {
	key := planargraph_coordToKey(coord)
	return nm.nodeMap[key]
}

// iterator returns an Iterator over the Nodes in this NodeMap, sorted in ascending order
// by angle with the positive x-axis.
func (nm *Planargraph_NodeMap) iterator() []*Planargraph_Node {
	return nm.Values()
}

// Values returns the Nodes in this NodeMap, sorted in ascending order
// by angle with the positive x-axis.
func (nm *Planargraph_NodeMap) Values() []*Planargraph_Node {
	nodes := make([]*Planargraph_Node, 0, len(nm.nodeMap))
	for _, node := range nm.nodeMap {
		nodes = append(nodes, node)
	}
	// Sort by coordinate for deterministic ordering (similar to Java TreeMap).
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetCoordinate().CompareTo(nodes[j].GetCoordinate()) < 0
	})
	return nodes
}
