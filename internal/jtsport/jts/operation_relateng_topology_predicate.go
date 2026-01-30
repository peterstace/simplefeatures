package jts

// OperationRelateng_TopologyPredicate is the API for strategy classes implementing
// spatial predicates based on the DE-9IM topology model.
// Predicate values for specific geometry pairs can be evaluated by RelateNG.
type OperationRelateng_TopologyPredicate interface {
	IsOperationRelateng_TopologyPredicate()

	// Name gets the name of the predicate.
	Name() string

	// RequireSelfNoding reports whether this predicate requires self-noding for
	// geometries which contain crossing edges (for example, LineStrings, or
	// GeometryCollections containing lines or polygons which may self-intersect).
	// Self-noding ensures that intersections are computed consistently in cases
	// which contain self-crossings and mutual crossings.
	//
	// Most predicates require this, but it can be avoided for simple intersection
	// detection (such as in Intersects() and Disjoint()). Avoiding self-noding
	// improves performance for polygonal inputs.
	RequireSelfNoding() bool

	// RequireInteraction reports whether this predicate requires interaction
	// between the input geometries. This is the case if:
	//   IM[I, I] >= 0 or IM[I, B] >= 0 or IM[B, I] >= 0 or IM[B, B] >= 0
	// This allows a fast result if the envelopes of the geometries are disjoint.
	RequireInteraction() bool

	// RequireCovers reports whether this predicate requires that the source
	// cover the target. This is the case if:
	//   IM[Ext(Src), Int(Tgt)] = F and IM[Ext(Src), Bdy(Tgt)] = F
	// If true, this allows a fast result if the source envelope does not cover
	// the target envelope.
	RequireCovers(isSourceA bool) bool

	// RequireExteriorCheck reports whether this predicate requires checking if
	// the source input intersects the Exterior of the target input. This is the
	// case if:
	//   IM[Int(Src), Ext(Tgt)] >= 0 or IM[Bdy(Src), Ext(Tgt)] >= 0
	// If false, this may permit a faster result in some geometric situations.
	RequireExteriorCheck(isSourceA bool) bool

	// InitDim initializes the predicate for a specific geometric case.
	// This may allow the predicate result to become known if it can be
	// inferred from the dimensions.
	InitDim(dimA, dimB int)

	// InitEnv initializes the predicate for a specific geometric case.
	// This may allow the predicate result to become known if it can be
	// inferred from the envelopes.
	InitEnv(envA, envB *Geom_Envelope)

	// UpdateDimension updates the entry in the DE-9IM intersection matrix
	// for given Locations in the input geometries.
	//
	// If this method is called with a Dimension value which is less than
	// the current value for the matrix entry, the implementing class should
	// avoid changing the entry if this would cause information loss.
	UpdateDimension(locA, locB, dimension int)

	// Finish indicates that the value of the predicate can be finalized
	// based on its current state.
	Finish()

	// IsKnown tests if the predicate value is known.
	IsKnown() bool

	// Value gets the current value of the predicate result.
	// The value is only valid if IsKnown() is true.
	Value() bool
}
