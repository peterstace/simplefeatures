package jts

import "fmt"

// OperationRelateng_TopologyPredicateTracer_Trace creates a new predicate
// tracing the evaluation of a given predicate.
func OperationRelateng_TopologyPredicateTracer_Trace(pred OperationRelateng_TopologyPredicate) OperationRelateng_TopologyPredicate {
	return &operationRelateng_PredicateTracer{
		pred: pred,
	}
}

// operationRelateng_PredicateTracer wraps a TopologyPredicate and traces its
// evaluation for debugging.
type operationRelateng_PredicateTracer struct {
	pred OperationRelateng_TopologyPredicate
}

func (pt *operationRelateng_PredicateTracer) IsOperationRelateng_TopologyPredicate() {}

// Name returns the name of the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) Name() string {
	return pt.pred.Name()
}

// RequireSelfNoding delegates to the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) RequireSelfNoding() bool {
	return pt.pred.RequireSelfNoding()
}

// RequireInteraction delegates to the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) RequireInteraction() bool {
	return pt.pred.RequireInteraction()
}

// RequireCovers delegates to the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) RequireCovers(isSourceA bool) bool {
	return pt.pred.RequireCovers(isSourceA)
}

// RequireExteriorCheck delegates to the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) RequireExteriorCheck(isSourceA bool) bool {
	return pt.pred.RequireExteriorCheck(isSourceA)
}

// InitDim initializes with dimensions and traces the result.
func (pt *operationRelateng_PredicateTracer) InitDim(dimA, dimB int) {
	pt.pred.InitDim(dimA, dimB)
	pt.checkValue("dimensions")
}

// InitEnv initializes with envelopes and traces the result.
func (pt *operationRelateng_PredicateTracer) InitEnv(envA, envB *Geom_Envelope) {
	pt.pred.InitEnv(envA, envB)
	pt.checkValue("envelopes")
}

// UpdateDimension updates the dimension and traces the result.
func (pt *operationRelateng_PredicateTracer) UpdateDimension(locA, locB, dimension int) {
	desc := fmt.Sprintf("A:%c/B:%c -> %d",
		Geom_Location_ToLocationSymbol(locA),
		Geom_Location_ToLocationSymbol(locB),
		dimension)
	ind := ""
	isChanged := pt.isDimChanged(locA, locB, dimension)
	if isChanged {
		ind = " <<< "
	}
	fmt.Println(desc + ind)
	pt.pred.UpdateDimension(locA, locB, dimension)
	if isChanged {
		pt.checkValue("IM entry")
	}
}

func (pt *operationRelateng_PredicateTracer) isDimChanged(locA, locB, dimension int) bool {
	if implPred, ok := pt.pred.(interface {
		IsDimChanged(int, int, int) bool
	}); ok {
		return implPred.IsDimChanged(locA, locB, dimension)
	}
	return false
}

func (pt *operationRelateng_PredicateTracer) checkValue(source string) {
	if pt.pred.IsKnown() {
		fmt.Printf("%s = %v based on %s\n", pt.pred.Name(), pt.pred.Value(), source)
	}
}

// Finish delegates to the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) Finish() {
	pt.pred.Finish()
}

// IsKnown delegates to the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) IsKnown() bool {
	return pt.pred.IsKnown()
}

// Value delegates to the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) Value() bool {
	return pt.pred.Value()
}

// String returns the string representation of the wrapped predicate.
func (pt *operationRelateng_PredicateTracer) String() string {
	return fmt.Sprintf("%v", pt.pred)
}
