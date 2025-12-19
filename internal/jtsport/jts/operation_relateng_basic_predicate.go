package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// The base class for relate topological predicates with a boolean value.
// Implements tri-state logic for the predicate value, to detect when the
// final value has been determined.

const operationRelateng_BasicPredicate_UNKNOWN = -1
const operationRelateng_BasicPredicate_FALSE = 0
const operationRelateng_BasicPredicate_TRUE = 1

func operationRelateng_BasicPredicate_isKnownValue(value int) bool {
	return value > operationRelateng_BasicPredicate_UNKNOWN
}

func operationRelateng_BasicPredicate_toBoolean(value int) bool {
	return value == operationRelateng_BasicPredicate_TRUE
}

func operationRelateng_BasicPredicate_toValue(val bool) int {
	if val {
		return operationRelateng_BasicPredicate_TRUE
	}
	return operationRelateng_BasicPredicate_FALSE
}

// OperationRelateng_BasicPredicate_IsIntersection tests if two geometries
// intersect based on an interaction at given locations.
func OperationRelateng_BasicPredicate_IsIntersection(locA, locB int) bool {
	// i.e. some location on both geometries intersects.
	return locA != Geom_Location_Exterior && locB != Geom_Location_Exterior
}

var _ OperationRelateng_TopologyPredicate = (*OperationRelateng_BasicPredicate)(nil)

type OperationRelateng_BasicPredicate struct {
	child java.Polymorphic
	value int
}

func (p *OperationRelateng_BasicPredicate) IsOperationRelateng_TopologyPredicate() {}

func (p *OperationRelateng_BasicPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *OperationRelateng_BasicPredicate) GetParent() java.Polymorphic {
	return nil
}

func OperationRelateng_NewBasicPredicate() *OperationRelateng_BasicPredicate {
	return &OperationRelateng_BasicPredicate{
		value: operationRelateng_BasicPredicate_UNKNOWN,
	}
}

// Name dispatcher - abstract, must be overridden.
func (p *OperationRelateng_BasicPredicate) Name() string {
	if impl, ok := java.GetLeaf(p).(interface{ Name_BODY() string }); ok {
		return impl.Name_BODY()
	}
	panic("abstract method Name called")
}

// RequireSelfNoding dispatcher.
func (p *OperationRelateng_BasicPredicate) RequireSelfNoding() bool {
	if impl, ok := java.GetLeaf(p).(interface{ RequireSelfNoding_BODY() bool }); ok {
		return impl.RequireSelfNoding_BODY()
	}
	return p.RequireSelfNoding_BODY()
}

func (p *OperationRelateng_BasicPredicate) RequireSelfNoding_BODY() bool {
	return true
}

// RequireInteraction dispatcher.
func (p *OperationRelateng_BasicPredicate) RequireInteraction() bool {
	if impl, ok := java.GetLeaf(p).(interface{ RequireInteraction_BODY() bool }); ok {
		return impl.RequireInteraction_BODY()
	}
	return p.RequireInteraction_BODY()
}

func (p *OperationRelateng_BasicPredicate) RequireInteraction_BODY() bool {
	return true
}

// RequireCovers dispatcher.
func (p *OperationRelateng_BasicPredicate) RequireCovers(isSourceA bool) bool {
	if impl, ok := java.GetLeaf(p).(interface{ RequireCovers_BODY(bool) bool }); ok {
		return impl.RequireCovers_BODY(isSourceA)
	}
	return p.RequireCovers_BODY(isSourceA)
}

func (p *OperationRelateng_BasicPredicate) RequireCovers_BODY(isSourceA bool) bool {
	return false
}

// RequireExteriorCheck dispatcher.
func (p *OperationRelateng_BasicPredicate) RequireExteriorCheck(isSourceA bool) bool {
	if impl, ok := java.GetLeaf(p).(interface{ RequireExteriorCheck_BODY(bool) bool }); ok {
		return impl.RequireExteriorCheck_BODY(isSourceA)
	}
	return p.RequireExteriorCheck_BODY(isSourceA)
}

func (p *OperationRelateng_BasicPredicate) RequireExteriorCheck_BODY(isSourceA bool) bool {
	return true
}

// InitDim dispatcher.
func (p *OperationRelateng_BasicPredicate) InitDim(dimA, dimB int) {
	if impl, ok := java.GetLeaf(p).(interface{ InitDim_BODY(int, int) }); ok {
		impl.InitDim_BODY(dimA, dimB)
		return
	}
	p.InitDim_BODY(dimA, dimB)
}

func (p *OperationRelateng_BasicPredicate) InitDim_BODY(dimA, dimB int) {
	// default if dimensions provide no information.
}

// InitEnv dispatcher.
func (p *OperationRelateng_BasicPredicate) InitEnv(envA, envB *Geom_Envelope) {
	if impl, ok := java.GetLeaf(p).(interface {
		InitEnv_BODY(*Geom_Envelope, *Geom_Envelope)
	}); ok {
		impl.InitEnv_BODY(envA, envB)
		return
	}
	p.InitEnv_BODY(envA, envB)
}

func (p *OperationRelateng_BasicPredicate) InitEnv_BODY(envA, envB *Geom_Envelope) {
	// default if envelopes provide no information.
}

// UpdateDimension dispatcher - abstract, must be overridden.
func (p *OperationRelateng_BasicPredicate) UpdateDimension(locA, locB, dimension int) {
	if impl, ok := java.GetLeaf(p).(interface{ UpdateDimension_BODY(int, int, int) }); ok {
		impl.UpdateDimension_BODY(locA, locB, dimension)
		return
	}
	panic("abstract method UpdateDimension called")
}

// Finish dispatcher - abstract, must be overridden.
func (p *OperationRelateng_BasicPredicate) Finish() {
	if impl, ok := java.GetLeaf(p).(interface{ Finish_BODY() }); ok {
		impl.Finish_BODY()
		return
	}
	panic("abstract method Finish called")
}

// IsKnown dispatcher.
func (p *OperationRelateng_BasicPredicate) IsKnown() bool {
	if impl, ok := java.GetLeaf(p).(interface{ IsKnown_BODY() bool }); ok {
		return impl.IsKnown_BODY()
	}
	return p.IsKnown_BODY()
}

func (p *OperationRelateng_BasicPredicate) IsKnown_BODY() bool {
	return operationRelateng_BasicPredicate_isKnownValue(p.value)
}

// Value dispatcher.
func (p *OperationRelateng_BasicPredicate) Value() bool {
	if impl, ok := java.GetLeaf(p).(interface{ Value_BODY() bool }); ok {
		return impl.Value_BODY()
	}
	return p.Value_BODY()
}

func (p *OperationRelateng_BasicPredicate) Value_BODY() bool {
	return operationRelateng_BasicPredicate_toBoolean(p.value)
}

// SetValue updates the predicate value to the given state if it is currently
// unknown.
func (p *OperationRelateng_BasicPredicate) SetValue(val bool) {
	// don't change already-known value.
	if p.IsKnown() {
		return
	}
	p.value = operationRelateng_BasicPredicate_toValue(val)
}

// SetValueInt sets the predicate value to the given int state if it is
// currently unknown.
func (p *OperationRelateng_BasicPredicate) SetValueInt(val int) {
	// don't change already-known value.
	if p.IsKnown() {
		return
	}
	p.value = val
}

// SetValueIf sets the value if the condition is true.
func (p *OperationRelateng_BasicPredicate) SetValueIf(value bool, cond bool) {
	if cond {
		p.SetValue(value)
	}
}

// Require sets the value to false if the condition is not met.
func (p *OperationRelateng_BasicPredicate) Require(cond bool) {
	if !cond {
		p.SetValue(false)
	}
}

// RequireCoversEnv sets the value to false if envelope a does not cover
// envelope b.
func (p *OperationRelateng_BasicPredicate) RequireCoversEnv(a, b *Geom_Envelope) {
	p.Require(a.CoversEnvelope(b))
}
