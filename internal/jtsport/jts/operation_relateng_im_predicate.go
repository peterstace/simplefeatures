package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// A base class for predicates which are determined using entries in an
// IntersectionMatrix.

// OperationRelateng_IMPredicate_IsDimsCompatibleWithCovers tests if the
// dimensions are compatible for a covers relationship.
func OperationRelateng_IMPredicate_IsDimsCompatibleWithCovers(dim0, dim1 int) bool {
	// allow Points coveredBy zero-length Lines.
	if dim0 == Geom_Dimension_P && dim1 == Geom_Dimension_L {
		return true
	}
	return dim0 >= dim1
}

const operationRelateng_IMPredicate_DIM_UNKNOWN = Geom_Dimension_DontCare

type OperationRelateng_IMPredicate struct {
	*OperationRelateng_BasicPredicate
	child     java.Polymorphic
	dimA      int
	dimB      int
	intMatrix *Geom_IntersectionMatrix
}

func (p *OperationRelateng_IMPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *OperationRelateng_IMPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_BasicPredicate
}

func OperationRelateng_NewIMPredicate() *OperationRelateng_IMPredicate {
	base := OperationRelateng_NewBasicPredicate()
	im := Geom_NewIntersectionMatrix()
	// E/E is always dim = 2.
	im.Set(Geom_Location_Exterior, Geom_Location_Exterior, Geom_Dimension_A)
	pred := &OperationRelateng_IMPredicate{
		OperationRelateng_BasicPredicate: base,
		intMatrix:                       im,
	}
	base.child = pred
	return pred
}

func (p *OperationRelateng_IMPredicate) GetDimA() int {
	return p.dimA
}

func (p *OperationRelateng_IMPredicate) GetDimB() int {
	return p.dimB
}

func (p *OperationRelateng_IMPredicate) GetIntMatrix() *Geom_IntersectionMatrix {
	return p.intMatrix
}

// InitDim_BODY overrides BasicPredicate.
func (p *OperationRelateng_IMPredicate) InitDim_BODY(dimA, dimB int) {
	p.dimA = dimA
	p.dimB = dimB
}

// UpdateDimension_BODY overrides BasicPredicate.
func (p *OperationRelateng_IMPredicate) UpdateDimension_BODY(locA, locB, dimension int) {
	// only record an increased dimension value.
	if p.IsDimChanged(locA, locB, dimension) {
		p.intMatrix.Set(locA, locB, dimension)
		// set value if predicate value can be known.
		if p.IsDetermined() {
			p.SetValue(p.ValueIM())
		}
	}
}

// IsDimChanged tests if the dimension at the given locations would change.
func (p *OperationRelateng_IMPredicate) IsDimChanged(locA, locB, dimension int) bool {
	return dimension > p.intMatrix.Get(locA, locB)
}

// IsDetermined dispatcher - abstract, must be overridden.
// Tests whether predicate evaluation can be short-circuited due to the current
// state of the matrix providing enough information to determine the predicate
// value. If this value is true then ValueIM() must provide the correct result.
func (p *OperationRelateng_IMPredicate) IsDetermined() bool {
	if impl, ok := java.GetLeaf(p).(interface{ IsDetermined_BODY() bool }); ok {
		return impl.IsDetermined_BODY()
	}
	panic("abstract method IsDetermined called")
}

// IntersectsExteriorOf tests whether the exterior of the specified input
// geometry is intersected by any part of the other input.
func (p *OperationRelateng_IMPredicate) IntersectsExteriorOf(isA bool) bool {
	if isA {
		return p.IsIntersects(Geom_Location_Exterior, Geom_Location_Interior) ||
			p.IsIntersects(Geom_Location_Exterior, Geom_Location_Boundary)
	}
	return p.IsIntersects(Geom_Location_Interior, Geom_Location_Exterior) ||
		p.IsIntersects(Geom_Location_Boundary, Geom_Location_Exterior)
}

// IsIntersects tests if the matrix entry at the given locations indicates an
// intersection.
func (p *OperationRelateng_IMPredicate) IsIntersects(locA, locB int) bool {
	return p.intMatrix.Get(locA, locB) >= Geom_Dimension_P
}

// IsKnownEntry tests if the matrix entry at the given locations is known.
func (p *OperationRelateng_IMPredicate) IsKnownEntry(locA, locB int) bool {
	return p.intMatrix.Get(locA, locB) != operationRelateng_IMPredicate_DIM_UNKNOWN
}

// IsDimension tests if the matrix entry at the given locations equals the
// given dimension.
func (p *OperationRelateng_IMPredicate) IsDimension(locA, locB, dimension int) bool {
	return p.intMatrix.Get(locA, locB) == dimension
}

// GetDimension gets the dimension at the given locations.
func (p *OperationRelateng_IMPredicate) GetDimension(locA, locB int) int {
	return p.intMatrix.Get(locA, locB)
}

// Finish_BODY overrides BasicPredicate - sets the final value based on the
// state of the IM.
func (p *OperationRelateng_IMPredicate) Finish_BODY() {
	p.SetValue(p.ValueIM())
}

// ValueIM dispatcher - abstract, must be overridden.
// Gets the value of the predicate according to the current intersection
// matrix state.
func (p *OperationRelateng_IMPredicate) ValueIM() bool {
	if impl, ok := java.GetLeaf(p).(interface{ ValueIM_BODY() bool }); ok {
		return impl.ValueIM_BODY()
	}
	panic("abstract method ValueIM called")
}

func (p *OperationRelateng_IMPredicate) String() string {
	return p.Name() + ": " + p.intMatrix.String()
}
