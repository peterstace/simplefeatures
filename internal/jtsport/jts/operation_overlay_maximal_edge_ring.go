package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationOverlay_MaximalEdgeRing is a ring of DirectedEdges which may contain
// nodes of degree > 2. A MaximalEdgeRing may represent two different spatial
// entities:
//   - a single polygon possibly containing inversions (if the ring is oriented CW)
//   - a single hole possibly containing exversions (if the ring is oriented CCW)
//
// If the MaximalEdgeRing represents a polygon, the interior of the polygon is
// strongly connected.
//
// These are the form of rings used to define polygons under some spatial data
// models. However, under the OGC SFS model, MinimalEdgeRings are required. A
// MaximalEdgeRing can be converted to a list of MinimalEdgeRings using the
// BuildMinimalRings method.
type OperationOverlay_MaximalEdgeRing struct {
	*Geomgraph_EdgeRing
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mer *OperationOverlay_MaximalEdgeRing) GetChild() java.Polymorphic {
	return mer.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mer *OperationOverlay_MaximalEdgeRing) GetParent() java.Polymorphic {
	return mer.Geomgraph_EdgeRing
}

// OperationOverlay_NewMaximalEdgeRing creates a new MaximalEdgeRing.
func OperationOverlay_NewMaximalEdgeRing(start *Geomgraph_DirectedEdge, geometryFactory *Geom_GeometryFactory) *OperationOverlay_MaximalEdgeRing {
	er := geomgraph_NewEdgeRingBase(geometryFactory)
	mer := &OperationOverlay_MaximalEdgeRing{
		Geomgraph_EdgeRing: er,
	}
	er.child = mer
	geomgraph_InitEdgeRing(er, start)
	return mer
}

// GetNext_BODY returns the next DirectedEdge in the ring.
func (mer *OperationOverlay_MaximalEdgeRing) GetNext_BODY(de *Geomgraph_DirectedEdge) *Geomgraph_DirectedEdge {
	return de.GetNext()
}

// SetEdgeRing_BODY sets the edge ring for the given DirectedEdge.
func (mer *OperationOverlay_MaximalEdgeRing) SetEdgeRing_BODY(de *Geomgraph_DirectedEdge, er *Geomgraph_EdgeRing) {
	de.SetEdgeRing(er)
}

// LinkDirectedEdgesForMinimalEdgeRings links the DirectedEdges at each node in
// this EdgeRing to form MinimalEdgeRings.
func (mer *OperationOverlay_MaximalEdgeRing) LinkDirectedEdgesForMinimalEdgeRings() {
	de := mer.startDe
	for {
		node := de.GetNode()
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		des.LinkMinimalDirectedEdges(mer.Geomgraph_EdgeRing)
		de = de.GetNext()
		if de == mer.startDe {
			break
		}
	}
}

// BuildMinimalRings builds the list of MinimalEdgeRings for this MaximalEdgeRing.
func (mer *OperationOverlay_MaximalEdgeRing) BuildMinimalRings() []*Geomgraph_EdgeRing {
	var minEdgeRings []*Geomgraph_EdgeRing
	de := mer.startDe
	for {
		if de.GetMinEdgeRing() == nil {
			minEr := OperationOverlay_NewMinimalEdgeRing(de, mer.geometryFactory)
			minEdgeRings = append(minEdgeRings, minEr.Geomgraph_EdgeRing)
		}
		de = de.GetNext()
		if de == mer.startDe {
			break
		}
	}
	return minEdgeRings
}
