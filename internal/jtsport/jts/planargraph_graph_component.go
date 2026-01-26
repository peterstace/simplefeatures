package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Planargraph_GraphComponent_SetVisitedIterator sets the Visited state for all
// GraphComponents in a slice.
func Planargraph_GraphComponent_SetVisitedIterator(components []*Planargraph_GraphComponent, visited bool) {
	for _, comp := range components {
		comp.SetVisited(visited)
	}
}

// Planargraph_GraphComponent_SetMarkedIterator sets the Marked state for all
// GraphComponents in a slice.
func Planargraph_GraphComponent_SetMarkedIterator(components []*Planargraph_GraphComponent, marked bool) {
	for _, comp := range components {
		comp.SetMarked(marked)
	}
}

// Planargraph_GraphComponent_GetComponentWithVisitedState finds the first
// GraphComponent in a slice which has the specified visited state.
// Returns nil if none found.
func Planargraph_GraphComponent_GetComponentWithVisitedState(components []*Planargraph_GraphComponent, visitedState bool) *Planargraph_GraphComponent {
	for _, comp := range components {
		if comp.IsVisited() == visitedState {
			return comp
		}
	}
	return nil
}

// Planargraph_GraphComponent is the base class for all graph component classes.
// Maintains flags of use in generic graph algorithms.
// Provides two flags:
//   - marked - typically this is used to indicate a state that persists
//     for the course of the graph's lifetime. For instance, it can be
//     used to indicate that a component has been logically deleted from the graph.
//   - visited - this is used to indicate that a component has been processed
//     or visited by a single graph algorithm. For instance, a breadth-first traversal of the
//     graph might use this to indicate that a node has already been traversed.
//     The visited flag may be set and cleared many times during the lifetime of a graph.
//
// Graph components support storing user context data. This will typically be
// used by client algorithms which use planar graphs.
type Planargraph_GraphComponent struct {
	child     java.Polymorphic
	isMarked  bool
	isVisited bool
	data      any
}

func (g *Planargraph_GraphComponent) GetChild() java.Polymorphic {
	return g.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (g *Planargraph_GraphComponent) GetParent() java.Polymorphic {
	return nil
}

// IsVisited tests if a component has been visited during the course of a graph algorithm.
func (g *Planargraph_GraphComponent) IsVisited() bool {
	return g.isVisited
}

// SetVisited sets the visited flag for this component.
func (g *Planargraph_GraphComponent) SetVisited(isVisited bool) {
	g.isVisited = isVisited
}

// IsMarked tests if a component has been marked at some point during the processing
// involving this graph.
func (g *Planargraph_GraphComponent) IsMarked() bool {
	return g.isMarked
}

// SetMarked sets the marked flag for this component.
func (g *Planargraph_GraphComponent) SetMarked(isMarked bool) {
	g.isMarked = isMarked
}

// SetContext sets the user-defined data for this component.
func (g *Planargraph_GraphComponent) SetContext(data any) {
	g.data = data
}

// GetContext gets the user-defined data for this component.
func (g *Planargraph_GraphComponent) GetContext() any {
	return g.data
}

// SetData sets the user-defined data for this component.
func (g *Planargraph_GraphComponent) SetData(data any) {
	g.data = data
}

// GetData gets the user-defined data for this component.
func (g *Planargraph_GraphComponent) GetData() any {
	return g.data
}

// IsRemoved tests whether this component has been removed from its containing graph.
// This is an abstract method that must be implemented by child types.
func (g *Planargraph_GraphComponent) IsRemoved() bool {
	type isRemovedImpl interface {
		IsRemoved_BODY() bool
	}
	if impl, ok := java.GetLeaf(g).(isRemovedImpl); ok {
		return impl.IsRemoved_BODY()
	}
	panic("abstract method Planargraph_GraphComponent.IsRemoved called")
}
