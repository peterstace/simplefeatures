package jts

import (
	"strconv"
	"strings"
)

// JtstestTestrunner_Test represents a single test for two geometries.
type JtstestTestrunner_Test struct {
	description    string
	operation      string
	expectedResult JtstestTestrunner_Result
	testIndex      int
	geometryIndex  string
	arguments      []string
	testCase       *JtstestTestrunner_TestCase
	passed         bool
	tolerance      float64

	// Cache for actual computed result.
	targetGeometry *Geom_Geometry
	operationArgs  []any
	isRun          bool
	actualResult   JtstestTestrunner_Result
	exception      error
}

// JtstestTestrunner_NewTest creates a Test with the given description. The
// given operation (e.g. "equals") will be performed, the expected result of
// which is expectedResult.
func JtstestTestrunner_NewTest(
	testCase *JtstestTestrunner_TestCase,
	testIndex int,
	description string,
	operation string,
	geometryIndex string,
	arguments []string,
	expectedResult JtstestTestrunner_Result,
	tolerance float64,
) *JtstestTestrunner_Test {
	args := make([]string, len(arguments))
	copy(args, arguments)
	return &JtstestTestrunner_Test{
		testCase:       testCase,
		testIndex:      testIndex,
		description:    description,
		operation:      operation,
		geometryIndex:  geometryIndex,
		arguments:      args,
		expectedResult: expectedResult,
		tolerance:      tolerance,
	}
}

func (t *JtstestTestrunner_Test) SetResult(result JtstestTestrunner_Result) {
	t.expectedResult = result
}

func (t *JtstestTestrunner_Test) SetArgument(i int, value string) {
	t.arguments[i] = value
}

func (t *JtstestTestrunner_Test) GetDescription() string {
	return t.description
}

func (t *JtstestTestrunner_Test) GetGeometryIndex() string {
	return t.geometryIndex
}

func (t *JtstestTestrunner_Test) GetExpectedResult() JtstestTestrunner_Result {
	return t.expectedResult
}

func (t *JtstestTestrunner_Test) HasExpectedResult() bool {
	return t.expectedResult != nil
}

func (t *JtstestTestrunner_Test) GetOperation() string {
	return t.operation
}

func (t *JtstestTestrunner_Test) GetTestIndex() int {
	return t.testIndex
}

func (t *JtstestTestrunner_Test) GetArgument(i int) string {
	return t.arguments[i]
}

func (t *JtstestTestrunner_Test) GetArgumentCount() int {
	return len(t.arguments)
}

func (t *JtstestTestrunner_Test) IsPassed() bool {
	return t.passed
}

func (t *JtstestTestrunner_Test) GetException() error {
	return t.exception
}

func (t *JtstestTestrunner_Test) GetTestCase() *JtstestTestrunner_TestCase {
	return t.testCase
}

func (t *JtstestTestrunner_Test) RemoveArgument(i int) {
	t.arguments = append(t.arguments[:i], t.arguments[i+1:]...)
}

func (t *JtstestTestrunner_Test) Run() {
	t.exception = nil
	passed, err := t.computePassed()
	if err != nil {
		t.exception = err
	} else {
		t.passed = passed
	}
}

func (t *JtstestTestrunner_Test) IsRun() bool {
	return t.isRun
}

func (t *JtstestTestrunner_Test) computePassed() (bool, error) {
	actualResult, err := t.GetActualResult()
	if err != nil {
		return false, err
	}
	if !t.HasExpectedResult() {
		return true, nil
	}
	matcher := t.testCase.GetTestRun().GetResultMatcher()
	return matcher.IsMatch(
		t.targetGeometry,
		t.operation,
		t.operationArgs,
		actualResult,
		t.expectedResult,
		t.tolerance,
	), nil
}

func (t *JtstestTestrunner_Test) isExpectedResultGeometryValid() bool {
	if geomResult, ok := t.expectedResult.(*JtstestTestrunner_GeometryResult); ok {
		expectedGeom := geomResult.GetGeometry()
		return expectedGeom.IsValid()
	}
	return true
}

// GetActualResult computes the actual result and caches the result value.
func (t *JtstestTestrunner_Test) GetActualResult() (JtstestTestrunner_Result, error) {
	if t.isRun {
		return t.actualResult, nil
	}
	t.isRun = true
	if strings.EqualFold(t.geometryIndex, "A") {
		t.targetGeometry = t.testCase.GetGeometryA()
	} else {
		t.targetGeometry = t.testCase.GetGeometryB()
	}
	t.operationArgs = t.convertArgs(t.arguments)
	op := t.getGeometryOperation()
	result, err := op.Invoke(t.operation, t.targetGeometry, t.operationArgs)
	if err != nil {
		return nil, err
	}
	t.actualResult = result
	return t.actualResult, nil
}

func (t *JtstestTestrunner_Test) getGeometryOperation() JtstestGeomop_GeometryOperation {
	return t.testCase.GetTestRun().GetGeometryOperation()
}

func (t *JtstestTestrunner_Test) ToXml() string {
	xml := ""
	xml += "<test>" + jtstestUtil_stringUtil_newLine
	if t.description != "" {
		xml += "  <desc>" + jtstestUtil_StringUtil_EscapeHTML(t.description) + "</desc>" + jtstestUtil_stringUtil_newLine
	}
	xml += "  <op name=\"" + t.operation + "\""
	xml += " arg1=\"" + t.geometryIndex + "\""
	j := 2
	for _, argument := range t.arguments {
		xml += " arg" + strconv.Itoa(j) + "=\"" + argument + "\""
		j++
	}
	xml += ">" + jtstestUtil_stringUtil_newLine
	xml += jtstestUtil_StringUtil_Indent(t.expectedResult.ToFormattedString(), 4) + jtstestUtil_stringUtil_newLine
	xml += "  </op>" + jtstestUtil_stringUtil_newLine
	xml += "</test>" + jtstestUtil_stringUtil_newLine
	return xml
}

func (t *JtstestTestrunner_Test) convertArgs(argStr []string) []any {
	args := make([]any, len(argStr))
	for i := range argStr {
		args[i] = t.convertArgToGeomOrString(argStr[i])
	}
	return args
}

func (t *JtstestTestrunner_Test) convertArgToGeomOrString(argStr string) any {
	if strings.EqualFold(argStr, "null") {
		return nil
	}
	if strings.EqualFold(argStr, "A") {
		return t.testCase.GetGeometryA()
	}
	if strings.EqualFold(argStr, "B") {
		return t.testCase.GetGeometryB()
	}
	return argStr
}
