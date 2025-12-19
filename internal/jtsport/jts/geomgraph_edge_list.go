package jts

import (
	"fmt"
	"strings"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_EdgeList is a list of Edges. It supports locating edges that are
// pointwise equals to a target edge.
type Geomgraph_EdgeList struct {
	child java.Polymorphic
	edges []*Geomgraph_Edge
	// ocaMap is an index of the edges, for fast lookup. Keys are OCA strings.
	ocaMap map[string]*Geomgraph_Edge
}

// GetChild returns the immediate child in the type hierarchy chain.
func (el *Geomgraph_EdgeList) GetChild() java.Polymorphic {
	return el.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (el *Geomgraph_EdgeList) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewEdgeList creates a new EdgeList.
func Geomgraph_NewEdgeList() *Geomgraph_EdgeList {
	return &Geomgraph_EdgeList{
		ocaMap: make(map[string]*Geomgraph_Edge),
	}
}

// Add inserts an edge into the list.
func (el *Geomgraph_EdgeList) Add(e *Geomgraph_Edge) {
	el.edges = append(el.edges, e)
	oca := geomgraph_EdgeList_makeOCA(e.GetCoordinates())
	el.ocaMap[oca] = e
}

// AddAll adds all edges from a collection.
func (el *Geomgraph_EdgeList) AddAll(edges []*Geomgraph_Edge) {
	for _, e := range edges {
		el.Add(e)
	}
}

// Clear removes all edges from the list.
func (el *Geomgraph_EdgeList) Clear() {
	el.edges = nil
	el.ocaMap = make(map[string]*Geomgraph_Edge)
}

// GetEdges returns the list of edges.
func (el *Geomgraph_EdgeList) GetEdges() []*Geomgraph_Edge {
	return el.edges
}

// FindEqualEdge returns an edge equal to e if one is already in the list,
// otherwise returns nil.
func (el *Geomgraph_EdgeList) FindEqualEdge(e *Geomgraph_Edge) *Geomgraph_Edge {
	oca := geomgraph_EdgeList_makeOCA(e.GetCoordinates())
	// Will return nil if no edge matches.
	return el.ocaMap[oca]
}

// Get returns the edge at the given index.
func (el *Geomgraph_EdgeList) Get(i int) *Geomgraph_Edge {
	return el.edges[i]
}

// FindEdgeIndex returns the index of the edge e if it's in the list, otherwise
// -1.
func (el *Geomgraph_EdgeList) FindEdgeIndex(e *Geomgraph_Edge) int {
	for i, edge := range el.edges {
		if edge.Equals(e) {
			return i
		}
	}
	return -1
}

// String returns a string representation of this EdgeList.
func (el *Geomgraph_EdgeList) String() string {
	var buf strings.Builder
	buf.WriteString("MULTILINESTRING ( ")
	for j, e := range el.edges {
		if j > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("(")
		pts := e.GetCoordinates()
		for i, pt := range pts {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(fmt.Sprintf("%v %v", pt.GetX(), pt.GetY()))
		}
		buf.WriteString(")\n")
	}
	buf.WriteString(")  ")
	return buf.String()
}

// geomgraph_EdgeList_makeOCA creates an orientation-independent key for a
// coordinate array. This is similar to OrientedCoordinateArray in JTS.
func geomgraph_EdgeList_makeOCA(pts []*Geom_Coordinate) string {
	if len(pts) == 0 {
		return ""
	}
	// Determine canonical orientation.
	orientation := geomgraph_EdgeList_increasingDirection(pts) == 1

	// Build a string key from the coordinates in canonical order.
	var buf strings.Builder
	if orientation {
		for _, pt := range pts {
			buf.WriteString(fmt.Sprintf("%.15g,%.15g;", pt.GetX(), pt.GetY()))
		}
	} else {
		for i := len(pts) - 1; i >= 0; i-- {
			buf.WriteString(fmt.Sprintf("%.15g,%.15g;", pts[i].GetX(), pts[i].GetY()))
		}
	}
	return buf.String()
}

// geomgraph_EdgeList_increasingDirection returns 1 if the coordinate array
// should be read forward, -1 if backward (based on lexicographic comparison).
func geomgraph_EdgeList_increasingDirection(pts []*Geom_Coordinate) int {
	for i := 0; i < len(pts)/2; i++ {
		j := len(pts) - 1 - i
		// Skip equal points on both ends.
		comp := pts[i].CompareTo(pts[j])
		if comp != 0 {
			return comp
		}
	}
	// Array must be a palindrome - defined to be in positive direction.
	return 1
}
