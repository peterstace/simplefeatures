package jts

// OperationUnion_UnionStrategy is a strategy interface that adapts UnaryUnion
// to different kinds of overlay algorithms.
type OperationUnion_UnionStrategy interface {
	// Union computes the union of two geometries.
	// This method may panic with a Geom_TopologyException if one is encountered.
	Union(g0, g1 *Geom_Geometry) *Geom_Geometry

	// IsFloatingPrecision indicates whether the union function operates using
	// a floating (full) precision model.
	// If this is the case, then the unary union code can make use of the
	// OverlapUnion performance optimization, and perhaps other optimizations
	// as well. Otherwise, the union result extent may not be the same as the
	// extent of the inputs, which prevents using some optimizations.
	IsFloatingPrecision() bool
}
