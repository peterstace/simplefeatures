package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// A predicate that matches a DE-9IM pattern.

type OperationRelateng_IMPatternMatcher struct {
	*OperationRelateng_IMPredicate
	child         java.Polymorphic
	imPattern     string
	patternMatrix *Geom_IntersectionMatrix
}

func (p *OperationRelateng_IMPatternMatcher) GetChild() java.Polymorphic { return p.child }

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *OperationRelateng_IMPatternMatcher) GetParent() java.Polymorphic {
	return p.OperationRelateng_IMPredicate
}

func OperationRelateng_NewIMPatternMatcher(imPattern string) *OperationRelateng_IMPatternMatcher {
	base := OperationRelateng_NewIMPredicate()
	pm := Geom_NewIntersectionMatrixWithElements(imPattern)
	matcher := &OperationRelateng_IMPatternMatcher{
		OperationRelateng_IMPredicate: base,
		imPattern:                    imPattern,
		patternMatrix:                pm,
	}
	base.child = matcher
	return matcher
}

func (p *OperationRelateng_IMPatternMatcher) Name_BODY() string {
	return "IMPattern"
}

func (p *OperationRelateng_IMPatternMatcher) InitEnv_BODY(envA, envB *Geom_Envelope) {
	p.OperationRelateng_IMPredicate.InitDim_BODY(p.dimA, p.dimB)
	// if pattern specifies any non-E/non-E interaction, envelopes must not be disjoint.
	requiresInteraction := operationRelateng_IMPatternMatcher_requireInteraction(p.patternMatrix)
	isDisjoint := envA.Disjoint(envB)
	p.SetValueIf(false, requiresInteraction && isDisjoint)
}

func (p *OperationRelateng_IMPatternMatcher) RequireInteraction_BODY() bool {
	return operationRelateng_IMPatternMatcher_requireInteraction(p.patternMatrix)
}

func operationRelateng_IMPatternMatcher_requireInteraction(im *Geom_IntersectionMatrix) bool {
	return operationRelateng_IMPatternMatcher_isInteraction(im.Get(Geom_Location_Interior, Geom_Location_Interior)) ||
		operationRelateng_IMPatternMatcher_isInteraction(im.Get(Geom_Location_Interior, Geom_Location_Boundary)) ||
		operationRelateng_IMPatternMatcher_isInteraction(im.Get(Geom_Location_Boundary, Geom_Location_Interior)) ||
		operationRelateng_IMPatternMatcher_isInteraction(im.Get(Geom_Location_Boundary, Geom_Location_Boundary))
}

func operationRelateng_IMPatternMatcher_isInteraction(imDim int) bool {
	return imDim == Geom_Dimension_True || imDim >= Geom_Dimension_P
}

func (p *OperationRelateng_IMPatternMatcher) IsDetermined_BODY() bool {
	// Matrix entries only increase in dimension as topology is computed.
	// The predicate can be short-circuited (as false) if any computed entry
	// is greater than the mask value.
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			patternEntry := p.patternMatrix.Get(i, j)
			if patternEntry == Geom_Dimension_DontCare {
				continue
			}
			matrixVal := p.GetDimension(i, j)
			// mask entry TRUE requires a known matrix entry.
			if patternEntry == Geom_Dimension_True {
				if matrixVal < 0 {
					return false
				}
			} else if matrixVal > patternEntry {
				// result is known (false) if matrix entry has exceeded mask.
				return true
			}
		}
	}
	return false
}

func (p *OperationRelateng_IMPatternMatcher) ValueIM_BODY() bool {
	return p.intMatrix.MatchesPattern(p.imPattern)
}

func (p *OperationRelateng_IMPatternMatcher) String() string {
	return p.Name_BODY() + "(" + p.imPattern + ")"
}
