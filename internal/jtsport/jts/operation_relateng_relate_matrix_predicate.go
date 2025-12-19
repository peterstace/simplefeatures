package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Evaluates the full relate IntersectionMatrix.

type OperationRelateng_RelateMatrixPredicate struct {
	*OperationRelateng_IMPredicate
	child java.Polymorphic
}

func (p *OperationRelateng_RelateMatrixPredicate) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *OperationRelateng_RelateMatrixPredicate) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func OperationRelateng_NewRelateMatrixPredicate() *OperationRelateng_RelateMatrixPredicate {
	base := OperationRelateng_NewIMPredicate()
	pred := &OperationRelateng_RelateMatrixPredicate{
		OperationRelateng_IMPredicate: base,
	}
	base.child = pred
	return pred
}

func (p *OperationRelateng_RelateMatrixPredicate) Name_BODY() string {
	return "relateMatrix"
}

func (p *OperationRelateng_RelateMatrixPredicate) RequireInteraction_BODY() bool {
	// ensure entire matrix is computed.
	return false
}

func (p *OperationRelateng_RelateMatrixPredicate) IsDetermined_BODY() bool {
	// ensure entire matrix is computed.
	return false
}

func (p *OperationRelateng_RelateMatrixPredicate) ValueIM_BODY() bool {
	// indicates full matrix is being evaluated.
	return false
}

// GetIM gets the current state of the IM matrix (which may only be partially
// complete).
func (p *OperationRelateng_RelateMatrixPredicate) GetIM() *Geom_IntersectionMatrix {
	return p.intMatrix
}
