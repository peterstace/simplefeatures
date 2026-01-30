package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelate_RelateOp_Relate computes the IntersectionMatrix for the
// spatial relationship between two Geometries, using the default (OGC SFS)
// Boundary Node Rule.
func OperationRelate_RelateOp_Relate(a, b *Geom_Geometry) *Geom_IntersectionMatrix {
	relOp := OperationRelate_NewRelateOp(a, b)
	return relOp.GetIntersectionMatrix()
}

// OperationRelate_RelateOp_RelateWithBoundaryNodeRule computes the
// IntersectionMatrix for the spatial relationship between two Geometries using
// a specified Boundary Node Rule.
func OperationRelate_RelateOp_RelateWithBoundaryNodeRule(a, b *Geom_Geometry, boundaryNodeRule Algorithm_BoundaryNodeRule) *Geom_IntersectionMatrix {
	relOp := OperationRelate_NewRelateOpWithBoundaryNodeRule(a, b, boundaryNodeRule)
	return relOp.GetIntersectionMatrix()
}

// OperationRelate_RelateOp implements the SFS relate() generalized spatial
// predicate on two Geometries.
//
// The class supports specifying a custom BoundaryNodeRule to be used during the
// relate computation.
//
// If named spatial predicates are used on the result IntersectionMatrix of the
// RelateOp, the result may or not be affected by the choice of
// BoundaryNodeRule, depending on the exact nature of the pattern. For instance,
// IsIntersects() is insensitive to the choice of BoundaryNodeRule, whereas
// IsTouches(int, int) is affected by the rule chosen.
//
// Note: custom Boundary Node Rules do not (currently) affect the results of
// other Geometry methods (such as GetBoundary). The results of these methods
// may not be consistent with the relationship computed by a custom Boundary
// Node Rule.
type OperationRelate_RelateOp struct {
	*Operation_GeometryGraphOperation
	child java.Polymorphic

	relate *OperationRelate_RelateComputer
}

// GetChild returns the immediate child in the type hierarchy chain.
func (ro *OperationRelate_RelateOp) GetChild() java.Polymorphic {
	return ro.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (ro *OperationRelate_RelateOp) GetParent() java.Polymorphic {
	return ro.Operation_GeometryGraphOperation
}

// OperationRelate_NewRelateOp creates a new Relate operation, using the default
// (OGC SFS) Boundary Node Rule.
func OperationRelate_NewRelateOp(g0, g1 *Geom_Geometry) *OperationRelate_RelateOp {
	ggo := Operation_NewGeometryGraphOperation(g0, g1)
	ro := &OperationRelate_RelateOp{
		Operation_GeometryGraphOperation: ggo,
	}
	ggo.child = ro
	ro.relate = OperationRelate_NewRelateComputer(ro.arg)
	return ro
}

// OperationRelate_NewRelateOpWithBoundaryNodeRule creates a new Relate
// operation with a specified Boundary Node Rule.
func OperationRelate_NewRelateOpWithBoundaryNodeRule(g0, g1 *Geom_Geometry, boundaryNodeRule Algorithm_BoundaryNodeRule) *OperationRelate_RelateOp {
	ggo := Operation_NewGeometryGraphOperationWithBoundaryNodeRule(g0, g1, boundaryNodeRule)
	ro := &OperationRelate_RelateOp{
		Operation_GeometryGraphOperation: ggo,
	}
	ggo.child = ro
	ro.relate = OperationRelate_NewRelateComputer(ro.arg)
	return ro
}

// GetIntersectionMatrix gets the IntersectionMatrix for the spatial
// relationship between the input geometries.
func (ro *OperationRelate_RelateOp) GetIntersectionMatrix() *Geom_IntersectionMatrix {
	return ro.relate.ComputeIM()
}
