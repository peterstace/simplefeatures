package jts

// Algorithm_BoundaryNodeRule is an interface for rules which determine whether
// node points which are in boundaries of Lineal geometry components are in the
// boundary of the parent geometry collection. The SFS specifies a single kind
// of boundary node rule, the Mod2BoundaryNodeRule rule. However, other kinds of
// Boundary Node Rules are appropriate in specific situations (for instance,
// linear network topology usually follows the EndPointBoundaryNodeRule).
//
// Some JTS operations allow the BoundaryNodeRule to be specified, and respect
// the supplied rule when computing the results of the operation.
type Algorithm_BoundaryNodeRule interface {
	// IsInBoundary tests whether a point that lies in boundaryCount geometry
	// component boundaries is considered to form part of the boundary of the
	// parent geometry.
	IsInBoundary(boundaryCount int) bool
}

// Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE is the Mod-2 Boundary Node Rule
// (which is the rule specified in the OGC SFS).
var Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE Algorithm_BoundaryNodeRule = &Algorithm_Mod2BoundaryNodeRule{}

// Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE is the Endpoint Boundary
// Node Rule.
var Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE Algorithm_BoundaryNodeRule = &Algorithm_EndPointBoundaryNodeRule{}

// Algorithm_BoundaryNodeRule_MULTIVALENT_ENDPOINT_BOUNDARY_RULE is the
// MultiValent Endpoint Boundary Node Rule.
var Algorithm_BoundaryNodeRule_MULTIVALENT_ENDPOINT_BOUNDARY_RULE Algorithm_BoundaryNodeRule = &Algorithm_MultiValentEndPointBoundaryNodeRule{}

// Algorithm_BoundaryNodeRule_MONOVALENT_ENDPOINT_BOUNDARY_RULE is the Monovalent
// Endpoint Boundary Node Rule.
var Algorithm_BoundaryNodeRule_MONOVALENT_ENDPOINT_BOUNDARY_RULE Algorithm_BoundaryNodeRule = &Algorithm_MonoValentEndPointBoundaryNodeRule{}

// Algorithm_BoundaryNodeRule_OGC_SFS_BOUNDARY_RULE is the Boundary Node Rule
// specified by the OGC Simple Features Specification, which is the same as the
// Mod-2 rule.
var Algorithm_BoundaryNodeRule_OGC_SFS_BOUNDARY_RULE = Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE

// Algorithm_Mod2BoundaryNodeRule is a BoundaryNodeRule which specifies that
// points are in the boundary of a lineal geometry iff the point lies on the
// boundary of an odd number of components. Under this rule LinearRings and
// closed LineStrings have an empty boundary.
//
// This is the rule specified by the OGC SFS, and is the default rule used in
// JTS.
type Algorithm_Mod2BoundaryNodeRule struct{}

// IsInBoundary implements Algorithm_BoundaryNodeRule.
func (r *Algorithm_Mod2BoundaryNodeRule) IsInBoundary(boundaryCount int) bool {
	// The "Mod-2 Rule".
	return boundaryCount%2 == 1
}

// Algorithm_EndPointBoundaryNodeRule is a BoundaryNodeRule which specifies that
// any points which are endpoints of lineal components are in the boundary of
// the parent geometry. This corresponds to the "intuitive" topological
// definition of boundary. Under this rule LinearRings have a non-empty boundary
// (the common endpoint of the underlying LineString).
//
// This rule is useful when dealing with linear networks. For example, it can be
// used to check whether linear networks are correctly noded. The usual network
// topology constraint is that linear segments may touch only at endpoints. In
// the case of a segment touching a closed segment (ring) at one point, the Mod2
// rule cannot distinguish between the permitted case of touching at the node
// point and the invalid case of touching at some other interior (non-node)
// point. The EndPoint rule does distinguish between these cases, so is more
// appropriate for use.
type Algorithm_EndPointBoundaryNodeRule struct{}

// IsInBoundary implements Algorithm_BoundaryNodeRule.
func (r *Algorithm_EndPointBoundaryNodeRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount > 0
}

// Algorithm_MultiValentEndPointBoundaryNodeRule is a BoundaryNodeRule which
// determines that only endpoints with valency greater than 1 are on the
// boundary. This corresponds to the boundary of a MultiLineString being all the
// "attached" endpoints, but not the "unattached" ones.
type Algorithm_MultiValentEndPointBoundaryNodeRule struct{}

// IsInBoundary implements Algorithm_BoundaryNodeRule.
func (r *Algorithm_MultiValentEndPointBoundaryNodeRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount > 1
}

// Algorithm_MonoValentEndPointBoundaryNodeRule is a BoundaryNodeRule which
// determines that only endpoints with valency of exactly 1 are on the boundary.
// This corresponds to the boundary of a MultiLineString being all the
// "unattached" endpoints.
type Algorithm_MonoValentEndPointBoundaryNodeRule struct{}

// IsInBoundary implements Algorithm_BoundaryNodeRule.
func (r *Algorithm_MonoValentEndPointBoundaryNodeRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount == 1
}
