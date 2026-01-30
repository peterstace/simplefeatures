package jts

import "sort"

// Planargraph_DirectedEdgeStar is a sorted collection of DirectedEdges which leave a Node
// in a PlanarGraph.
type Planargraph_DirectedEdgeStar struct {
	outEdges []*Planargraph_DirectedEdge
	sorted   bool
}

// Planargraph_NewDirectedEdgeStar constructs a DirectedEdgeStar with no edges.
func Planargraph_NewDirectedEdgeStar() *Planargraph_DirectedEdgeStar {
	return &Planargraph_DirectedEdgeStar{
		outEdges: make([]*Planargraph_DirectedEdge, 0),
		sorted:   false,
	}
}

// Add adds a new member to this DirectedEdgeStar.
func (des *Planargraph_DirectedEdgeStar) Add(de *Planargraph_DirectedEdge) {
	des.outEdges = append(des.outEdges, de)
	des.sorted = false
}

// Remove drops a member of this DirectedEdgeStar.
func (des *Planargraph_DirectedEdgeStar) Remove(de *Planargraph_DirectedEdge) {
	for i, e := range des.outEdges {
		if e == de {
			des.outEdges = append(des.outEdges[:i], des.outEdges[i+1:]...)
			return
		}
	}
}

// Iterator returns the DirectedEdges, in ascending order by angle with the positive x-axis.
func (des *Planargraph_DirectedEdgeStar) Iterator() []*Planargraph_DirectedEdge {
	des.sortEdges()
	return des.outEdges
}

// GetDegree returns the number of edges around the Node associated with this DirectedEdgeStar.
func (des *Planargraph_DirectedEdgeStar) GetDegree() int {
	return len(des.outEdges)
}

// GetCoordinate returns the coordinate for the node at which this star is based.
func (des *Planargraph_DirectedEdgeStar) GetCoordinate() *Geom_Coordinate {
	des.sortEdges()
	if len(des.outEdges) == 0 {
		return nil
	}
	return des.outEdges[0].GetCoordinate()
}

// GetEdges returns the DirectedEdges, in ascending order by angle with the positive x-axis.
func (des *Planargraph_DirectedEdgeStar) GetEdges() []*Planargraph_DirectedEdge {
	des.sortEdges()
	return des.outEdges
}

func (des *Planargraph_DirectedEdgeStar) sortEdges() {
	if !des.sorted {
		sort.Slice(des.outEdges, func(i, j int) bool {
			return des.outEdges[i].CompareTo(des.outEdges[j]) < 0
		})
		des.sorted = true
	}
}

// GetIndexByEdge returns the zero-based index of the given Edge, after sorting in ascending order
// by angle with the positive x-axis.
func (des *Planargraph_DirectedEdgeStar) GetIndexByEdge(edge *Planargraph_Edge) int {
	des.sortEdges()
	for i, de := range des.outEdges {
		if de.GetEdge() == edge {
			return i
		}
	}
	return -1
}

// GetIndexByDirectedEdge returns the zero-based index of the given DirectedEdge, after sorting in ascending order
// by angle with the positive x-axis.
func (des *Planargraph_DirectedEdgeStar) GetIndexByDirectedEdge(dirEdge *Planargraph_DirectedEdge) int {
	des.sortEdges()
	for i, de := range des.outEdges {
		if de == dirEdge {
			return i
		}
	}
	return -1
}

// GetIndexMod returns value of i modulo the number of edges in this DirectedEdgeStar
// (i.e. the remainder when i is divided by the number of edges).
func (des *Planargraph_DirectedEdgeStar) GetIndexMod(i int) int {
	modi := i % len(des.outEdges)
	if modi < 0 {
		modi += len(des.outEdges)
	}
	return modi
}

// GetNextEdge returns the DirectedEdge on the left-hand (CCW)
// side of the given DirectedEdge (which must be a member of this DirectedEdgeStar).
func (des *Planargraph_DirectedEdgeStar) GetNextEdge(dirEdge *Planargraph_DirectedEdge) *Planargraph_DirectedEdge {
	i := des.GetIndexByDirectedEdge(dirEdge)
	return des.outEdges[des.GetIndexMod(i+1)]
}

// GetNextCWEdge returns the DirectedEdge on the right-hand (CW)
// side of the given DirectedEdge (which must be a member of this DirectedEdgeStar).
func (des *Planargraph_DirectedEdgeStar) GetNextCWEdge(dirEdge *Planargraph_DirectedEdge) *Planargraph_DirectedEdge {
	i := des.GetIndexByDirectedEdge(dirEdge)
	return des.outEdges[des.GetIndexMod(i-1)]
}
