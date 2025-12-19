package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Edgegraph_HalfEdge represents a directed component of an edge in an
// EdgeGraph. HalfEdges link vertices whose locations are defined by
// Coordinates. HalfEdges start at an origin vertex, and terminate at a
// destination vertex. HalfEdges always occur in symmetric pairs, with the
// Sym() method giving access to the oppositely-oriented component. HalfEdges
// and the methods on them form an edge algebra, which can be used to traverse
// and query the topology of the graph formed by the edges.
//
// To support graphs where the edges are sequences of coordinates each edge may
// also have a direction point supplied. This is used to determine the ordering
// of the edges around the origin. HalfEdges with the same origin are ordered
// so that the ring of edges formed by them is oriented CCW.
//
// By design HalfEdges carry minimal information about the actual usage of the
// graph they represent. They can be subclassed to carry more information if
// required.
//
// HalfEdges form a complete and consistent data structure by themselves, but
// an EdgeGraph is useful to allow retrieving edges by vertex and edge
// location, as well as ensuring edges are created and linked appropriately.
type Edgegraph_HalfEdge struct {
	child java.Polymorphic
	orig  *Geom_Coordinate
	sym   *Edgegraph_HalfEdge
	next  *Edgegraph_HalfEdge
}

// GetChild returns the child type for polymorphism support.
func (he *Edgegraph_HalfEdge) GetChild() java.Polymorphic {
	return he.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (he *Edgegraph_HalfEdge) GetParent() java.Polymorphic {
	return nil
}

// Edgegraph_HalfEdge_Create creates a HalfEdge pair representing an edge
// between two vertices located at coordinates p0 and p1.
func Edgegraph_HalfEdge_Create(p0, p1 *Geom_Coordinate) *Edgegraph_HalfEdge {
	e0 := Edgegraph_NewHalfEdge(p0)
	e1 := Edgegraph_NewHalfEdge(p1)
	e0.Link(e1)
	return e0
}

// Edgegraph_NewHalfEdge creates a half-edge originating from a given
// coordinate.
func Edgegraph_NewHalfEdge(orig *Geom_Coordinate) *Edgegraph_HalfEdge {
	return &Edgegraph_HalfEdge{
		orig: orig,
	}
}

// Link links this edge with its sym (opposite) edge. This must be done for
// each pair of edges created.
func (he *Edgegraph_HalfEdge) Link(sym *Edgegraph_HalfEdge) {
	he.setSym(sym)
	sym.setSym(he)
	// set next ptrs for a single segment
	he.setNext(sym)
	sym.setNext(he)
}

// Orig gets the origin coordinate of this edge.
func (he *Edgegraph_HalfEdge) Orig() *Geom_Coordinate {
	return he.orig
}

// Dest gets the destination coordinate of this edge.
func (he *Edgegraph_HalfEdge) Dest() *Geom_Coordinate {
	return he.sym.orig
}

// DirectionX returns the X component of the direction vector.
func (he *Edgegraph_HalfEdge) DirectionX() float64 {
	return he.DirectionPt().GetX() - he.orig.GetX()
}

// DirectionY returns the Y component of the direction vector.
func (he *Edgegraph_HalfEdge) DirectionY() float64 {
	return he.DirectionPt().GetY() - he.orig.GetY()
}

// DirectionPt gets the direction point of this edge. In the base case this is
// the dest coordinate of the edge. Subclasses may override to allow a HalfEdge
// to represent an edge with more than two coordinates.
func (he *Edgegraph_HalfEdge) DirectionPt() *Geom_Coordinate {
	if impl, ok := java.GetLeaf(he).(interface{ DirectionPt_BODY() *Geom_Coordinate }); ok {
		return impl.DirectionPt_BODY()
	}
	return he.DirectionPt_BODY()
}

// DirectionPt_BODY is the default implementation.
func (he *Edgegraph_HalfEdge) DirectionPt_BODY() *Geom_Coordinate {
	return he.Dest()
}

// Sym gets the symmetric pair edge of this edge.
func (he *Edgegraph_HalfEdge) Sym() *Edgegraph_HalfEdge {
	return he.sym
}

func (he *Edgegraph_HalfEdge) setSym(e *Edgegraph_HalfEdge) {
	he.sym = e
}

func (he *Edgegraph_HalfEdge) setNext(e *Edgegraph_HalfEdge) {
	he.next = e
}

// Next gets the next edge CCW around the destination vertex of this edge,
// originating at that vertex. If the destination vertex has degree 1 then this
// is the sym edge.
func (he *Edgegraph_HalfEdge) Next() *Edgegraph_HalfEdge {
	return he.next
}

// Prev gets the previous edge CW around the origin vertex of this edge, with
// that vertex being its destination. It is always true that e.Next().Prev() == e.
// Note that this requires a scan of the origin edges, so may not be efficient
// for some uses.
func (he *Edgegraph_HalfEdge) Prev() *Edgegraph_HalfEdge {
	curr := he
	prev := he
	for {
		prev = curr
		curr = curr.ONext()
		if curr == he {
			break
		}
	}
	return prev.sym
}

// ONext gets the next edge CCW around the origin of this edge, with the same
// origin. If the origin vertex has degree 1 then this is the edge itself.
// e.ONext() is equal to e.Sym().Next().
func (he *Edgegraph_HalfEdge) ONext() *Edgegraph_HalfEdge {
	return he.sym.next
}

// Find finds the edge starting at the origin of this edge with the given dest
// vertex, if any.
func (he *Edgegraph_HalfEdge) Find(dest *Geom_Coordinate) *Edgegraph_HalfEdge {
	oNext := he
	for {
		if oNext == nil {
			return nil
		}
		if oNext.Dest().Equals2D(dest) {
			return oNext
		}
		oNext = oNext.ONext()
		if oNext == he {
			break
		}
	}
	return nil
}

// Equals tests whether this edge has the given orig and dest vertices.
func (he *Edgegraph_HalfEdge) Equals(p0, p1 *Geom_Coordinate) bool {
	return he.orig.Equals2D(p0) && he.sym.orig.Equals(p1)
}

// Insert inserts an edge into the ring of edges around the origin vertex of
// this edge, ensuring that the edges remain ordered CCW. The inserted edge
// must have the same origin as this edge.
func (he *Edgegraph_HalfEdge) Insert(eAdd *Edgegraph_HalfEdge) {
	// If this is only edge at origin, insert it after this
	if he.ONext() == he {
		// set linkage so ring is correct
		he.insertAfter(eAdd)
		return
	}

	// Scan edges until insertion point is found
	ePrev := he.insertionEdge(eAdd)
	ePrev.insertAfter(eAdd)
}

// insertionEdge finds the insertion edge for an edge being added to this
// origin, ensuring that the star of edges around the origin remains fully CCW.
func (he *Edgegraph_HalfEdge) insertionEdge(eAdd *Edgegraph_HalfEdge) *Edgegraph_HalfEdge {
	ePrev := he
	for {
		eNext := ePrev.ONext()
		// Case 1: General case, with eNext higher than ePrev.
		// Insert edge here if it lies between ePrev and eNext.
		if eNext.CompareTo(ePrev) > 0 &&
			eAdd.CompareTo(ePrev) >= 0 &&
			eAdd.CompareTo(eNext) <= 0 {
			return ePrev
		}
		// Case 2: Origin-crossing case, indicated by eNext <= ePrev.
		// Insert edge here if it lies in the gap between ePrev and eNext
		// across the origin.
		if eNext.CompareTo(ePrev) <= 0 &&
			(eAdd.CompareTo(eNext) <= 0 || eAdd.CompareTo(ePrev) >= 0) {
			return ePrev
		}
		ePrev = eNext
		if ePrev == he {
			break
		}
	}
	Util_Assert_ShouldNeverReachHereWithMessage("insertion edge not found")
	return nil
}

// insertAfter inserts an edge with the same origin after this one. Assumes
// that the inserted edge is in the correct position around the ring.
func (he *Edgegraph_HalfEdge) insertAfter(e *Edgegraph_HalfEdge) {
	Util_Assert_Equals(he.orig, e.Orig())
	save := he.ONext()
	he.sym.setNext(e)
	e.sym.setNext(save)
}

// IsEdgesSorted tests whether the edges around the origin are sorted
// correctly. Note that edges must be strictly increasing, which implies no two
// edges can have the same direction point.
func (he *Edgegraph_HalfEdge) IsEdgesSorted() bool {
	// find lowest edge at origin
	lowest := he.findLowest()
	e := lowest
	// check that all edges are sorted
	for {
		eNext := e.ONext()
		if eNext == lowest {
			break
		}
		isSorted := eNext.CompareTo(e) > 0
		if !isSorted {
			return false
		}
		e = eNext
		if e == lowest {
			break
		}
	}
	return true
}

// findLowest finds the lowest edge around the origin, using the standard edge
// ordering.
func (he *Edgegraph_HalfEdge) findLowest() *Edgegraph_HalfEdge {
	lowest := he
	e := he.ONext()
	for e != he {
		if e.CompareTo(lowest) < 0 {
			lowest = e
		}
		e = e.ONext()
	}
	return lowest
}

// CompareTo compares edges which originate at the same vertex based on the
// angle they make at their origin vertex with the positive X-axis. This allows
// sorting edges around their origin vertex in CCW order.
func (he *Edgegraph_HalfEdge) CompareTo(e *Edgegraph_HalfEdge) int {
	comp := he.CompareAngularDirection(e)
	return comp
}

// CompareAngularDirection implements the total order relation where the angle
// of edge a is greater than the angle of edge b, where the angle of an edge is
// the angle made by the first segment of the edge with the positive x-axis.
// When applied to a list of edges originating at the same point, this produces
// a CCW ordering of the edges around the point.
func (he *Edgegraph_HalfEdge) CompareAngularDirection(e *Edgegraph_HalfEdge) int {
	dx := he.DirectionX()
	dy := he.DirectionY()
	dx2 := e.DirectionX()
	dy2 := e.DirectionY()

	// same vector
	if dx == dx2 && dy == dy2 {
		return 0
	}

	quadrant := Geom_Quadrant_QuadrantFromDeltas(dx, dy)
	quadrant2 := Geom_Quadrant_QuadrantFromDeltas(dx2, dy2)

	// If the direction vectors are in different quadrants, that determines the
	// ordering
	if quadrant > quadrant2 {
		return 1
	}
	if quadrant < quadrant2 {
		return -1
	}

	// vectors are in the same quadrant
	// Check relative orientation of direction vectors
	// this is > e if it is CCW of e
	dir1 := he.DirectionPt()
	dir2 := e.DirectionPt()
	return Algorithm_Orientation_Index(e.Orig(), dir2, dir1)
}

// String provides a string representation of a HalfEdge.
func (he *Edgegraph_HalfEdge) String() string {
	return "HE(" + he.orig.String() + ", " + he.sym.orig.String() + ")"
}

// ToStringNode provides a string representation of the edges around the origin
// node of this edge. Uses the subclass representation for each edge.
func (he *Edgegraph_HalfEdge) ToStringNode() string {
	orig := he.Orig()
	_ = he.Dest()
	sb := "Node( " + orig.String() + " )\n"
	e := he
	for {
		sb += "  -> " + e.String() + "\n"
		e = e.ONext()
		if e == he {
			break
		}
	}
	return sb
}

// Degree computes the degree of the origin vertex. The degree is the number of
// edges originating from the vertex.
func (he *Edgegraph_HalfEdge) Degree() int {
	degree := 0
	e := he
	for {
		degree++
		e = e.ONext()
		if e == he {
			break
		}
	}
	return degree
}

// PrevNode finds the first node previous to this edge, if any. A node has
// degree <> 2. If no such node exists (i.e. the edge is part of a ring) then
// null is returned.
func (he *Edgegraph_HalfEdge) PrevNode() *Edgegraph_HalfEdge {
	e := he
	for e.Degree() == 2 {
		e = e.Prev()
		if e == he {
			return nil
		}
	}
	return e
}
