package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelate_EdgeEndBundleStar is an ordered list of EdgeEndBundles around
// a RelateNode. They are maintained in CCW order (starting with the positive
// x-axis) around the node for efficient lookup and topology building.
type OperationRelate_EdgeEndBundleStar struct {
	*Geomgraph_EdgeEndStar
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (eebs *OperationRelate_EdgeEndBundleStar) GetChild() java.Polymorphic {
	return eebs.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (eebs *OperationRelate_EdgeEndBundleStar) GetParent() java.Polymorphic {
	return eebs.Geomgraph_EdgeEndStar
}

// OperationRelate_NewEdgeEndBundleStar creates a new empty EdgeEndBundleStar.
func OperationRelate_NewEdgeEndBundleStar() *OperationRelate_EdgeEndBundleStar {
	ees := Geomgraph_NewEdgeEndStar()
	eebs := &OperationRelate_EdgeEndBundleStar{
		Geomgraph_EdgeEndStar: ees,
	}
	ees.child = eebs
	return eebs
}

// Insert_BODY inserts an EdgeEnd in order in the list. If there is an existing
// EdgeEndBundle which is parallel, the EdgeEnd is added to the bundle.
// Otherwise, a new EdgeEndBundle is created to contain the EdgeEnd.
func (eebs *OperationRelate_EdgeEndBundleStar) Insert_BODY(e *Geomgraph_EdgeEnd) {
	eb := eebs.findExistingBundle(e)
	if eb == nil {
		eb = OperationRelate_NewEdgeEndBundle(e)
		eebs.InsertEdgeEnd(eb.Geomgraph_EdgeEnd)
	} else {
		eb.Insert(e)
	}
}

// findExistingBundle finds an existing EdgeEndBundle that is parallel to the
// given EdgeEnd, or returns nil if none exists.
func (eebs *OperationRelate_EdgeEndBundleStar) findExistingBundle(e *Geomgraph_EdgeEnd) *OperationRelate_EdgeEndBundle {
	for _, edge := range eebs.edgeMap {
		if edge.CompareTo(e) == 0 {
			// Found parallel edge - return the bundle.
			if bundle, ok := edge.GetChild().(*OperationRelate_EdgeEndBundle); ok {
				return bundle
			}
		}
	}
	return nil
}

// UpdateIM updates the IM with the contribution for the EdgeEndBundles around
// the node.
func (eebs *OperationRelate_EdgeEndBundleStar) UpdateIM(im *Geom_IntersectionMatrix) {
	for _, e := range eebs.GetEdges() {
		if bundle, ok := e.GetChild().(*OperationRelate_EdgeEndBundle); ok {
			bundle.UpdateIM(im)
		}
	}
}
