package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geomgraph_GraphComponent is the parent class for the objects that form a
// graph. Each GraphComponent can carry a Label.
type Geomgraph_GraphComponent struct {
	child java.Polymorphic

	label *Geomgraph_Label

	// isInResult indicates if this component has already been included in the
	// result.
	isInResult   bool
	isCovered    bool
	isCoveredSet bool
	isVisited    bool
}

// GetChild returns the immediate child in the type hierarchy chain.
func (gc *Geomgraph_GraphComponent) GetChild() java.Polymorphic {
	return gc.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (gc *Geomgraph_GraphComponent) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewGraphComponent creates a new GraphComponent.
func Geomgraph_NewGraphComponent() *Geomgraph_GraphComponent {
	return &Geomgraph_GraphComponent{}
}

// Geomgraph_NewGraphComponentWithLabel creates a new GraphComponent with the
// given label.
func Geomgraph_NewGraphComponentWithLabel(label *Geomgraph_Label) *Geomgraph_GraphComponent {
	return &Geomgraph_GraphComponent{
		label: label,
	}
}

// GetLabel returns the label for this GraphComponent.
func (gc *Geomgraph_GraphComponent) GetLabel() *Geomgraph_Label {
	return gc.label
}

// SetLabel sets the label for this GraphComponent.
func (gc *Geomgraph_GraphComponent) SetLabel(label *Geomgraph_Label) {
	gc.label = label
}

// SetInResult sets whether this component is in the result.
func (gc *Geomgraph_GraphComponent) SetInResult(isInResult bool) {
	gc.isInResult = isInResult
}

// IsInResult returns true if this component is in the result.
func (gc *Geomgraph_GraphComponent) IsInResult() bool {
	return gc.isInResult
}

// SetCovered sets whether this component is covered.
func (gc *Geomgraph_GraphComponent) SetCovered(isCovered bool) {
	gc.isCovered = isCovered
	gc.isCoveredSet = true
}

// IsCovered returns true if this component is covered.
func (gc *Geomgraph_GraphComponent) IsCovered() bool {
	return gc.isCovered
}

// IsCoveredSet returns true if the covered flag has been set.
func (gc *Geomgraph_GraphComponent) IsCoveredSet() bool {
	return gc.isCoveredSet
}

// IsVisited returns true if this component has been visited.
func (gc *Geomgraph_GraphComponent) IsVisited() bool {
	return gc.isVisited
}

// SetVisited sets whether this component has been visited.
func (gc *Geomgraph_GraphComponent) SetVisited(isVisited bool) {
	gc.isVisited = isVisited
}

// GetCoordinate returns a coordinate in this component (or nil if there are
// none). This is an abstract method that must be implemented by subtypes.
func (gc *Geomgraph_GraphComponent) GetCoordinate() *Geom_Coordinate {
	if impl, ok := java.GetLeaf(gc).(interface{ GetCoordinate_BODY() *Geom_Coordinate }); ok {
		return impl.GetCoordinate_BODY()
	}
	panic("abstract method Geomgraph_GraphComponent.GetCoordinate called")
}

// ComputeIM computes the contribution to an IM for this component. This is an
// abstract method that must be implemented by subtypes.
func (gc *Geomgraph_GraphComponent) ComputeIM(im *Geom_IntersectionMatrix) {
	if impl, ok := java.GetLeaf(gc).(interface {
		ComputeIM_BODY(*Geom_IntersectionMatrix)
	}); ok {
		impl.ComputeIM_BODY(im)
		return
	}
	panic("abstract method Geomgraph_GraphComponent.ComputeIM called")
}

// IsIsolated returns true if this is an isolated component. An isolated
// component is one that does not intersect or touch any other component. This
// is the case if the label has valid locations for only a single Geometry.
// This is an abstract method that must be implemented by subtypes.
func (gc *Geomgraph_GraphComponent) IsIsolated() bool {
	if impl, ok := java.GetLeaf(gc).(interface{ IsIsolated_BODY() bool }); ok {
		return impl.IsIsolated_BODY()
	}
	panic("abstract method Geomgraph_GraphComponent.IsIsolated called")
}

// UpdateIM updates the IM with the contribution for this component. A
// component only contributes if it has a labelling for both parent geometries.
func (gc *Geomgraph_GraphComponent) UpdateIM(im *Geom_IntersectionMatrix) {
	Util_Assert_IsTrueWithMessage(gc.label.GetGeometryCount() >= 2, "found partial label")
	gc.ComputeIM(im)
}
