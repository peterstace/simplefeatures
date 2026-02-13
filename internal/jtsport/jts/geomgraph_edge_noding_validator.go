package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geomgraph_EdgeNodingValidator validates that a collection of Edges is
// correctly noded. Throws an appropriate exception if a noding error is found.
// Uses Noding_FastNodingValidator to perform the validation.
type Geomgraph_EdgeNodingValidator struct {
	child java.Polymorphic
	nv    *Noding_FastNodingValidator
}

// GetChild returns the immediate child in the type hierarchy chain.
func (env *Geomgraph_EdgeNodingValidator) GetChild() java.Polymorphic {
	return env.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (env *Geomgraph_EdgeNodingValidator) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_EdgeNodingValidator_CheckValid checks whether the supplied Edges
// are correctly noded. Throws a TopologyException if they are not.
func Geomgraph_EdgeNodingValidator_CheckValid(edges []*Geomgraph_Edge) {
	validator := Geomgraph_NewEdgeNodingValidator(edges)
	validator.CheckValid()
}

// Geomgraph_EdgeNodingValidator_ToSegmentStrings converts Edges to
// SegmentStrings.
func Geomgraph_EdgeNodingValidator_ToSegmentStrings(edges []*Geomgraph_Edge) []*Noding_BasicSegmentString {
	segStrings := make([]*Noding_BasicSegmentString, len(edges))
	for i, e := range edges {
		segStrings[i] = Noding_NewBasicSegmentString(e.GetCoordinates(), e)
	}
	return segStrings
}

// Geomgraph_NewEdgeNodingValidator creates a new validator for the given
// collection of Edges.
func Geomgraph_NewEdgeNodingValidator(edges []*Geomgraph_Edge) *Geomgraph_EdgeNodingValidator {
	bssSlice := Geomgraph_EdgeNodingValidator_ToSegmentStrings(edges)
	segStrings := make([]Noding_SegmentString, len(bssSlice))
	for i, bss := range bssSlice {
		segStrings[i] = bss
	}
	return &Geomgraph_EdgeNodingValidator{
		nv: Noding_NewFastNodingValidator(segStrings),
	}
}

// CheckValid checks whether the supplied edges are correctly noded.
// Panics with a TopologyException if they are not.
func (env *Geomgraph_EdgeNodingValidator) CheckValid() {
	env.nv.CheckValid()
}
