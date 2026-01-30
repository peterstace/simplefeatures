package jts

import (
	"fmt"
	"io"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationRelate_EdgeEndBundle is a collection of EdgeEnds which obey the
// following invariant: They originate at the same node and have the same
// direction.
type OperationRelate_EdgeEndBundle struct {
	*Geomgraph_EdgeEnd
	child    java.Polymorphic
	edgeEnds []*Geomgraph_EdgeEnd
}

// GetChild returns the immediate child in the type hierarchy chain.
func (eeb *OperationRelate_EdgeEndBundle) GetChild() java.Polymorphic {
	return eeb.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (eeb *OperationRelate_EdgeEndBundle) GetParent() java.Polymorphic {
	return eeb.Geomgraph_EdgeEnd
}

// OperationRelate_NewEdgeEndBundle creates a new EdgeEndBundle from an EdgeEnd.
func OperationRelate_NewEdgeEndBundle(e *Geomgraph_EdgeEnd) *OperationRelate_EdgeEndBundle {
	return OperationRelate_NewEdgeEndBundleWithBoundaryNodeRule(nil, e)
}

// OperationRelate_NewEdgeEndBundleWithBoundaryNodeRule creates a new
// EdgeEndBundle from an EdgeEnd with a BoundaryNodeRule.
func OperationRelate_NewEdgeEndBundleWithBoundaryNodeRule(boundaryNodeRule Algorithm_BoundaryNodeRule, e *Geomgraph_EdgeEnd) *OperationRelate_EdgeEndBundle {
	edgeEnd := Geomgraph_NewEdgeEndWithLabel(e.GetEdge(), e.GetCoordinate(), e.GetDirectedCoordinate(), Geomgraph_NewLabelFromLabel(e.GetLabel()))
	eeb := &OperationRelate_EdgeEndBundle{
		Geomgraph_EdgeEnd: edgeEnd,
	}
	edgeEnd.child = eeb
	eeb.Insert(e)
	return eeb
}

// GetLabel returns the label for this bundle.
func (eeb *OperationRelate_EdgeEndBundle) GetLabel() *Geomgraph_Label {
	return eeb.label
}

// Iterator returns an iterator over the EdgeEnds in this bundle.
func (eeb *OperationRelate_EdgeEndBundle) Iterator() []*Geomgraph_EdgeEnd {
	return eeb.edgeEnds
}

// GetEdgeEnds returns the EdgeEnds in this bundle.
func (eeb *OperationRelate_EdgeEndBundle) GetEdgeEnds() []*Geomgraph_EdgeEnd {
	return eeb.edgeEnds
}

// Insert adds an EdgeEnd to this bundle.
func (eeb *OperationRelate_EdgeEndBundle) Insert(e *Geomgraph_EdgeEnd) {
	// Assert: start point is the same.
	// Assert: direction is the same.
	eeb.edgeEnds = append(eeb.edgeEnds, e)
}

// ComputeLabel_BODY computes the overall edge label for the set of edges in
// this EdgeEndBundle. It essentially merges the ON and side labels for each
// edge. These labels must be compatible.
func (eeb *OperationRelate_EdgeEndBundle) ComputeLabel_BODY(boundaryNodeRule Algorithm_BoundaryNodeRule) {
	// Create the label. If any of the edges belong to areas, the label must be
	// an area label.
	isArea := false
	for _, e := range eeb.edgeEnds {
		if e.GetLabel().IsArea() {
			isArea = true
		}
	}
	if isArea {
		eeb.label = Geomgraph_NewLabelOnLeftRight(Geom_Location_None, Geom_Location_None, Geom_Location_None)
	} else {
		eeb.label = Geomgraph_NewLabelOn(Geom_Location_None)
	}

	// Compute the On label, and the side labels if present.
	for i := 0; i < 2; i++ {
		eeb.computeLabelOn(i, boundaryNodeRule)
		if isArea {
			eeb.computeLabelSides(i)
		}
	}
}

// computeLabelOn computes the overall ON location for the list of EdgeEnds.
// (This is essentially equivalent to computing the self-overlay of a single
// Geometry.)
//
// EdgeEnds can be either on the boundary (e.g. Polygon edge) OR in the interior
// (e.g. segment of a LineString) of their parent Geometry.
//
// In addition, GeometryCollections use a BoundaryNodeRule to determine whether
// a segment is on the boundary or not.
//
// Finally, in GeometryCollections it can occur that an edge is both on the
// boundary and in the interior (e.g. a LineString segment lying on top of a
// Polygon edge.) In this case the Boundary is given precedence.
//
// These observations result in the following rules for computing the ON
// location:
//   - if there are an odd number of Bdy edges, the attribute is Bdy
//   - if there are an even number >= 2 of Bdy edges, the attribute is Int
//   - if there are any Int edges, the attribute is Int
//   - otherwise, the attribute is NULL.
func (eeb *OperationRelate_EdgeEndBundle) computeLabelOn(geomIndex int, boundaryNodeRule Algorithm_BoundaryNodeRule) {
	// Compute the ON location value.
	boundaryCount := 0
	foundInterior := false

	for _, e := range eeb.edgeEnds {
		loc := e.GetLabel().GetLocationOn(geomIndex)
		if loc == Geom_Location_Boundary {
			boundaryCount++
		}
		if loc == Geom_Location_Interior {
			foundInterior = true
		}
	}
	loc := Geom_Location_None
	if foundInterior {
		loc = Geom_Location_Interior
	}
	if boundaryCount > 0 {
		loc = Geomgraph_GeometryGraph_DetermineBoundary(boundaryNodeRule, boundaryCount)
	}
	eeb.label.SetLocationOn(geomIndex, loc)
}

// computeLabelSides computes the labelling for each side.
func (eeb *OperationRelate_EdgeEndBundle) computeLabelSides(geomIndex int) {
	eeb.computeLabelSide(geomIndex, Geom_Position_Left)
	eeb.computeLabelSide(geomIndex, Geom_Position_Right)
}

// computeLabelSide computes the summary label for a side.
//
// The algorithm is:
//
//	FOR all edges
//	  IF any edge's location is INTERIOR for the side, side location = INTERIOR
//	  ELSE IF there is at least one EXTERIOR attribute, side location = EXTERIOR
//	  ELSE side location = NULL
//
// Note that it is possible for two sides to have apparently contradictory
// information i.e. one edge side may indicate that it is in the interior of a
// geometry, while another edge side may indicate the exterior of the same
// geometry. This is not an incompatibility - GeometryCollections may contain
// two Polygons that touch along an edge. This is the reason for
// Interior-primacy rule above - it results in the summary label having the
// Geometry interior on both sides.
func (eeb *OperationRelate_EdgeEndBundle) computeLabelSide(geomIndex, side int) {
	for _, e := range eeb.edgeEnds {
		if e.GetLabel().IsArea() {
			loc := e.GetLabel().GetLocation(geomIndex, side)
			if loc == Geom_Location_Interior {
				eeb.label.SetLocation(geomIndex, side, Geom_Location_Interior)
				return
			} else if loc == Geom_Location_Exterior {
				eeb.label.SetLocation(geomIndex, side, Geom_Location_Exterior)
			}
		}
	}
}

// UpdateIM updates the IM with the contribution for the computed label for the
// EdgeEnds.
func (eeb *OperationRelate_EdgeEndBundle) UpdateIM(im *Geom_IntersectionMatrix) {
	Geomgraph_Edge_UpdateIM(eeb.label, im)
}

// Print writes a representation of this EdgeEndBundle to the given writer.
func (eeb *OperationRelate_EdgeEndBundle) Print(out io.Writer) {
	fmt.Fprintf(out, "EdgeEndBundle--> Label: %v\n", eeb.label)
	for _, ee := range eeb.edgeEnds {
		ee.Print(out)
		fmt.Fprintln(out)
	}
}
