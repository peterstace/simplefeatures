package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Factory functions for creating predicate instances for evaluating OGC-standard
// named topological relationships. Predicates can be evaluated for geometries
// using RelateNG.

// OperationRelateng_RelatePredicate_Intersects creates a predicate to determine
// whether two geometries intersect.
func OperationRelateng_RelatePredicate_Intersects() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewBasicPredicate()
	pred := &operationRelateng_IntersectsPredicate{
		OperationRelateng_BasicPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_IntersectsPredicate struct {
	*OperationRelateng_BasicPredicate
	child java.Polymorphic
}

func (p *operationRelateng_IntersectsPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_IntersectsPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_BasicPredicate
}

func (p *operationRelateng_IntersectsPredicate) Name_BODY() string { return "intersects" }

func (p *operationRelateng_IntersectsPredicate) RequireSelfNoding_BODY() bool {
	// self-noding is not required to check for a simple interaction.
	return false
}

func (p *operationRelateng_IntersectsPredicate) RequireExteriorCheck_BODY(isSourceA bool) bool {
	// intersects only requires testing interaction.
	return false
}

func (p *operationRelateng_IntersectsPredicate) InitEnv_BODY(envA, envB *Geom_Envelope) {
	p.Require(envA.IntersectsEnvelope(envB))
}

func (p *operationRelateng_IntersectsPredicate) UpdateDimension_BODY(locA, locB, dimension int) {
	p.SetValueIf(true, OperationRelateng_BasicPredicate_IsIntersection(locA, locB))
}

func (p *operationRelateng_IntersectsPredicate) Finish_BODY() {
	// if no intersecting locations were found.
	p.SetValue(false)
}

// OperationRelateng_RelatePredicate_Disjoint creates a predicate to determine
// whether two geometries are disjoint.
func OperationRelateng_RelatePredicate_Disjoint() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewBasicPredicate()
	pred := &operationRelateng_DisjointPredicate{
		OperationRelateng_BasicPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_DisjointPredicate struct {
	*OperationRelateng_BasicPredicate
	child java.Polymorphic
}

func (p *operationRelateng_DisjointPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_DisjointPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_BasicPredicate
}

func (p *operationRelateng_DisjointPredicate) Name_BODY() string { return "disjoint" }

func (p *operationRelateng_DisjointPredicate) RequireSelfNoding_BODY() bool {
	// self-noding is not required to check for a simple interaction.
	return false
}

func (p *operationRelateng_DisjointPredicate) RequireInteraction_BODY() bool {
	// ensure entire matrix is computed.
	return false
}

func (p *operationRelateng_DisjointPredicate) RequireExteriorCheck_BODY(isSourceA bool) bool {
	// disjoint only requires testing interaction.
	return false
}

func (p *operationRelateng_DisjointPredicate) InitEnv_BODY(envA, envB *Geom_Envelope) {
	p.SetValueIf(true, envA.Disjoint(envB))
}

func (p *operationRelateng_DisjointPredicate) UpdateDimension_BODY(locA, locB, dimension int) {
	p.SetValueIf(false, OperationRelateng_BasicPredicate_IsIntersection(locA, locB))
}

func (p *operationRelateng_DisjointPredicate) Finish_BODY() {
	// if no intersecting locations were found.
	p.SetValue(true)
}

// OperationRelateng_RelatePredicate_Contains creates a predicate to determine
// whether a geometry contains another geometry.
func OperationRelateng_RelatePredicate_Contains() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &operationRelateng_ContainsPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_ContainsPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *operationRelateng_ContainsPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_ContainsPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func (p *operationRelateng_ContainsPredicate) Name_BODY() string { return "contains" }

func (p *operationRelateng_ContainsPredicate) RequireCovers_BODY(isSourceA bool) bool {
	return isSourceA == OperationRelateng_RelateGeometry_GEOM_A
}

func (p *operationRelateng_ContainsPredicate) RequireExteriorCheck_BODY(isSourceA bool) bool {
	// only need to check B against Exterior of A.
	return isSourceA == OperationRelateng_RelateGeometry_GEOM_B
}

func (p *operationRelateng_ContainsPredicate) InitDim_BODY(dimA, dimB int) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(dimA, dimB)
	p.Require(OperationRelateng_IMPredicate_IsDimsCompatibleWithCovers(dimA, dimB))
}

func (p *operationRelateng_ContainsPredicate) InitEnv_BODY(envA, envB *Geom_Envelope) {
	p.RequireCoversEnv(envA, envB)
}

func (p *operationRelateng_ContainsPredicate) IsDetermined_BODY() bool {
	return p.IntersectsExteriorOf(OperationRelateng_RelateGeometry_GEOM_A)
}

func (p *operationRelateng_ContainsPredicate) ValueIM_BODY() bool {
	return p.intMatrix.IsContains()
}

// OperationRelateng_RelatePredicate_Within creates a predicate to determine
// whether a geometry is within another geometry.
func OperationRelateng_RelatePredicate_Within() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &operationRelateng_WithinPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_WithinPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *operationRelateng_WithinPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_WithinPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func (p *operationRelateng_WithinPredicate) Name_BODY() string { return "within" }

func (p *operationRelateng_WithinPredicate) RequireCovers_BODY(isSourceA bool) bool {
	return isSourceA == OperationRelateng_RelateGeometry_GEOM_B
}

func (p *operationRelateng_WithinPredicate) RequireExteriorCheck_BODY(isSourceA bool) bool {
	// only need to check A against Exterior of B.
	return isSourceA == OperationRelateng_RelateGeometry_GEOM_A
}

func (p *operationRelateng_WithinPredicate) InitDim_BODY(dimA, dimB int) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(dimA, dimB)
	p.Require(OperationRelateng_IMPredicate_IsDimsCompatibleWithCovers(dimB, dimA))
}

func (p *operationRelateng_WithinPredicate) InitEnv_BODY(envA, envB *Geom_Envelope) {
	p.RequireCoversEnv(envB, envA)
}

func (p *operationRelateng_WithinPredicate) IsDetermined_BODY() bool {
	return p.IntersectsExteriorOf(OperationRelateng_RelateGeometry_GEOM_B)
}

func (p *operationRelateng_WithinPredicate) ValueIM_BODY() bool {
	return p.intMatrix.IsWithin()
}

// OperationRelateng_RelatePredicate_Covers creates a predicate to determine
// whether a geometry covers another geometry.
func OperationRelateng_RelatePredicate_Covers() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &operationRelateng_CoversPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_CoversPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *operationRelateng_CoversPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_CoversPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func (p *operationRelateng_CoversPredicate) Name_BODY() string { return "covers" }

func (p *operationRelateng_CoversPredicate) RequireCovers_BODY(isSourceA bool) bool {
	return isSourceA == OperationRelateng_RelateGeometry_GEOM_A
}

func (p *operationRelateng_CoversPredicate) RequireExteriorCheck_BODY(isSourceA bool) bool {
	// only need to check B against Exterior of A.
	return isSourceA == OperationRelateng_RelateGeometry_GEOM_B
}

func (p *operationRelateng_CoversPredicate) InitDim_BODY(dimA, dimB int) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(dimA, dimB)
	p.Require(OperationRelateng_IMPredicate_IsDimsCompatibleWithCovers(dimA, dimB))
}

func (p *operationRelateng_CoversPredicate) InitEnv_BODY(envA, envB *Geom_Envelope) {
	p.RequireCoversEnv(envA, envB)
}

func (p *operationRelateng_CoversPredicate) IsDetermined_BODY() bool {
	return p.IntersectsExteriorOf(OperationRelateng_RelateGeometry_GEOM_A)
}

func (p *operationRelateng_CoversPredicate) ValueIM_BODY() bool {
	return p.intMatrix.IsCovers()
}

// OperationRelateng_RelatePredicate_CoveredBy creates a predicate to determine
// whether a geometry is covered by another geometry.
func OperationRelateng_RelatePredicate_CoveredBy() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &operationRelateng_CoveredByPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_CoveredByPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *operationRelateng_CoveredByPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_CoveredByPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func (p *operationRelateng_CoveredByPredicate) Name_BODY() string { return "coveredBy" }

func (p *operationRelateng_CoveredByPredicate) RequireCovers_BODY(isSourceA bool) bool {
	return isSourceA == OperationRelateng_RelateGeometry_GEOM_B
}

func (p *operationRelateng_CoveredByPredicate) RequireExteriorCheck_BODY(isSourceA bool) bool {
	// only need to check A against Exterior of B.
	return isSourceA == OperationRelateng_RelateGeometry_GEOM_A
}

func (p *operationRelateng_CoveredByPredicate) InitDim_BODY(dimA, dimB int) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(dimA, dimB)
	p.Require(OperationRelateng_IMPredicate_IsDimsCompatibleWithCovers(dimB, dimA))
}

func (p *operationRelateng_CoveredByPredicate) InitEnv_BODY(envA, envB *Geom_Envelope) {
	p.RequireCoversEnv(envB, envA)
}

func (p *operationRelateng_CoveredByPredicate) IsDetermined_BODY() bool {
	return p.IntersectsExteriorOf(OperationRelateng_RelateGeometry_GEOM_B)
}

func (p *operationRelateng_CoveredByPredicate) ValueIM_BODY() bool {
	return p.intMatrix.IsCoveredBy()
}

// OperationRelateng_RelatePredicate_Crosses creates a predicate to determine
// whether a geometry crosses another geometry.
func OperationRelateng_RelatePredicate_Crosses() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &operationRelateng_CrossesPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_CrossesPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *operationRelateng_CrossesPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_CrossesPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func (p *operationRelateng_CrossesPredicate) Name_BODY() string { return "crosses" }

func (p *operationRelateng_CrossesPredicate) InitDim_BODY(dimA, dimB int) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(dimA, dimB)
	isBothPointsOrAreas := (dimA == Geom_Dimension_P && dimB == Geom_Dimension_P) ||
		(dimA == Geom_Dimension_A && dimB == Geom_Dimension_A)
	p.Require(!isBothPointsOrAreas)
}

func (p *operationRelateng_CrossesPredicate) IsDetermined_BODY() bool {
	if p.dimA == Geom_Dimension_L && p.dimB == Geom_Dimension_L {
		// L/L interaction can only be dim = P.
		if p.GetDimension(Geom_Location_Interior, Geom_Location_Interior) > Geom_Dimension_P {
			return true
		}
	} else if p.dimA < p.dimB {
		if p.IsIntersects(Geom_Location_Interior, Geom_Location_Interior) &&
			p.IsIntersects(Geom_Location_Interior, Geom_Location_Exterior) {
			return true
		}
	} else if p.dimA > p.dimB {
		if p.IsIntersects(Geom_Location_Interior, Geom_Location_Interior) &&
			p.IsIntersects(Geom_Location_Exterior, Geom_Location_Interior) {
			return true
		}
	}
	return false
}

func (p *operationRelateng_CrossesPredicate) ValueIM_BODY() bool {
	return p.intMatrix.IsCrosses(p.dimA, p.dimB)
}

// OperationRelateng_RelatePredicate_EqualsTopo creates a predicate to determine
// whether two geometries are topologically equal.
func OperationRelateng_RelatePredicate_EqualsTopo() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &operationRelateng_EqualsTopoPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_EqualsTopoPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *operationRelateng_EqualsTopoPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_EqualsTopoPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func (p *operationRelateng_EqualsTopoPredicate) Name_BODY() string { return "equals" }

func (p *operationRelateng_EqualsTopoPredicate) InitDim_BODY(dimA, dimB int) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(dimA, dimB)
	p.Require(dimA == dimB)
}

func (p *operationRelateng_EqualsTopoPredicate) InitEnv_BODY(envA, envB *Geom_Envelope) {
	p.Require(envA.Equals(envB))
}

func (p *operationRelateng_EqualsTopoPredicate) IsDetermined_BODY() bool {
	isEitherExteriorIntersects :=
		p.IsIntersects(Geom_Location_Interior, Geom_Location_Exterior) ||
			p.IsIntersects(Geom_Location_Boundary, Geom_Location_Exterior) ||
			p.IsIntersects(Geom_Location_Exterior, Geom_Location_Interior) ||
			p.IsIntersects(Geom_Location_Exterior, Geom_Location_Boundary)
	return isEitherExteriorIntersects
}

func (p *operationRelateng_EqualsTopoPredicate) ValueIM_BODY() bool {
	return p.intMatrix.IsEquals(p.dimA, p.dimB)
}

// OperationRelateng_RelatePredicate_Overlaps creates a predicate to determine
// whether a geometry overlaps another geometry.
func OperationRelateng_RelatePredicate_Overlaps() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &operationRelateng_OverlapsPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_OverlapsPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *operationRelateng_OverlapsPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_OverlapsPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func (p *operationRelateng_OverlapsPredicate) Name_BODY() string { return "overlaps" }

func (p *operationRelateng_OverlapsPredicate) InitDim_BODY(dimA, dimB int) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(dimA, dimB)
	p.Require(dimA == dimB)
}

func (p *operationRelateng_OverlapsPredicate) IsDetermined_BODY() bool {
	if p.dimA == Geom_Dimension_A || p.dimA == Geom_Dimension_P {
		if p.IsIntersects(Geom_Location_Interior, Geom_Location_Interior) &&
			p.IsIntersects(Geom_Location_Interior, Geom_Location_Exterior) &&
			p.IsIntersects(Geom_Location_Exterior, Geom_Location_Interior) {
			return true
		}
	}
	if p.dimA == Geom_Dimension_L {
		if p.IsDimension(Geom_Location_Interior, Geom_Location_Interior, Geom_Dimension_L) &&
			p.IsIntersects(Geom_Location_Interior, Geom_Location_Exterior) &&
			p.IsIntersects(Geom_Location_Exterior, Geom_Location_Interior) {
			return true
		}
	}
	return false
}

func (p *operationRelateng_OverlapsPredicate) ValueIM_BODY() bool {
	return p.intMatrix.IsOverlaps(p.dimA, p.dimB)
}

// OperationRelateng_RelatePredicate_Touches creates a predicate to determine
// whether a geometry touches another geometry.
func OperationRelateng_RelatePredicate_Touches() OperationRelateng_TopologyPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &operationRelateng_TouchesPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

type operationRelateng_TouchesPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *operationRelateng_TouchesPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *operationRelateng_TouchesPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func (p *operationRelateng_TouchesPredicate) Name_BODY() string { return "touches" }

func (p *operationRelateng_TouchesPredicate) InitDim_BODY(dimA, dimB int) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(dimA, dimB)
	// Points have only interiors, so cannot touch.
	isBothPoints := dimA == 0 && dimB == 0
	p.Require(!isBothPoints)
}

func (p *operationRelateng_TouchesPredicate) IsDetermined_BODY() bool {
	// for touches interiors cannot intersect.
	isInteriorsIntersects := p.IsIntersects(Geom_Location_Interior, Geom_Location_Interior)
	return isInteriorsIntersects
}

func (p *operationRelateng_TouchesPredicate) ValueIM_BODY() bool {
	return p.intMatrix.IsTouches(p.dimA, p.dimB)
}

// OperationRelateng_RelatePredicate_Matches creates a predicate that matches a
// DE-9IM matrix pattern.
func OperationRelateng_RelatePredicate_Matches(imPattern string) OperationRelateng_TopologyPredicate {
	return OperationRelateng_NewIMPatternMatcher(imPattern)
}
