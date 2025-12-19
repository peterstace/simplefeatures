package jts

var _ JtstestTestrunner_Result = (*JtstestTestrunner_BooleanResult)(nil)

type JtstestTestrunner_BooleanResult struct {
	result bool
}

func JtstestTestrunner_NewBooleanResult(result bool) *JtstestTestrunner_BooleanResult {
	return &JtstestTestrunner_BooleanResult{result: result}
}

func JtstestTestrunner_NewBooleanResultFromBoolean(result bool) *JtstestTestrunner_BooleanResult {
	return JtstestTestrunner_NewBooleanResult(result)
}

func (r *JtstestTestrunner_BooleanResult) IsJtstestTestrunner_Result() {}

func (r *JtstestTestrunner_BooleanResult) EqualsResult(other JtstestTestrunner_Result, tolerance float64) bool {
	otherBooleanResult, ok := other.(*JtstestTestrunner_BooleanResult)
	if !ok {
		return false
	}
	return r.result == otherBooleanResult.result
}

func (r *JtstestTestrunner_BooleanResult) ToFormattedString() string {
	return r.ToShortString()
}

func (r *JtstestTestrunner_BooleanResult) ToLongString() string {
	return r.ToShortString()
}

func (r *JtstestTestrunner_BooleanResult) ToShortString() string {
	if r.result {
		return "true"
	}
	return "false"
}
