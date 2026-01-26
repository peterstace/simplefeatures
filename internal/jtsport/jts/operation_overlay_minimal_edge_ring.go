package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationOverlay_MinimalEdgeRing is a ring of Edges with the property that no
// node has degree greater than 2. These are the form of rings required to
// represent polygons under the OGC SFS spatial data model.
type OperationOverlay_MinimalEdgeRing struct {
	*Geomgraph_EdgeRing
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mer *OperationOverlay_MinimalEdgeRing) GetChild() java.Polymorphic {
	return mer.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mer *OperationOverlay_MinimalEdgeRing) GetParent() java.Polymorphic {
	return mer.Geomgraph_EdgeRing
}

// OperationOverlay_NewMinimalEdgeRing creates a new MinimalEdgeRing.
func OperationOverlay_NewMinimalEdgeRing(start *Geomgraph_DirectedEdge, geometryFactory *Geom_GeometryFactory) *OperationOverlay_MinimalEdgeRing {
	er := geomgraph_NewEdgeRingBase(geometryFactory)
	mer := &OperationOverlay_MinimalEdgeRing{
		Geomgraph_EdgeRing: er,
	}
	er.child = mer
	geomgraph_InitEdgeRing(er, start)
	return mer
}

// GetNext_BODY returns the next DirectedEdge in the minimal ring.
func (mer *OperationOverlay_MinimalEdgeRing) GetNext_BODY(de *Geomgraph_DirectedEdge) *Geomgraph_DirectedEdge {
	return de.GetNextMin()
}

// SetEdgeRing_BODY sets the minimal edge ring for the given DirectedEdge.
func (mer *OperationOverlay_MinimalEdgeRing) SetEdgeRing_BODY(de *Geomgraph_DirectedEdge, er *Geomgraph_EdgeRing) {
	de.SetMinEdgeRing(er)
}
