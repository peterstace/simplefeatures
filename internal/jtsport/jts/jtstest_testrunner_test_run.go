package jts

// JtstestTestrunner_TestRun represents a collection of test cases read from a
// single XML test file.
type JtstestTestrunner_TestRun struct {
	testCaseIndexToRun int
	description        string
	testCases          []*JtstestTestrunner_TestCase
	precisionModel     *Geom_PrecisionModel
	geomOp             JtstestGeomop_GeometryOperation
	resultMatcher      JtstestTestrunner_ResultMatcher
	runIndex           int
	testFile           string
	workspace          string
}

func JtstestTestrunner_NewTestRun(
	description string,
	runIndex int,
	precisionModel *Geom_PrecisionModel,
	geomOp JtstestGeomop_GeometryOperation,
	resultMatcher JtstestTestrunner_ResultMatcher,
	testFile string,
) *JtstestTestrunner_TestRun {
	return &JtstestTestrunner_TestRun{
		testCaseIndexToRun: -1,
		description:        description,
		runIndex:           runIndex,
		precisionModel:     precisionModel,
		geomOp:             geomOp,
		resultMatcher:      resultMatcher,
		testFile:           testFile,
	}
}

func (r *JtstestTestrunner_TestRun) SetWorkspace(workspace string) {
	r.workspace = workspace
}

func (r *JtstestTestrunner_TestRun) SetTestCaseIndexToRun(testCaseIndexToRun int) {
	r.testCaseIndexToRun = testCaseIndexToRun
}

func (r *JtstestTestrunner_TestRun) GetWorkspace() string {
	return r.workspace
}

func (r *JtstestTestrunner_TestRun) GetTestCount() int {
	count := 0
	for _, testCase := range r.testCases {
		count += testCase.GetTestCount()
	}
	return count
}

func (r *JtstestTestrunner_TestRun) GetDescription() string {
	return r.description
}

func (r *JtstestTestrunner_TestRun) GetRunIndex() int {
	return r.runIndex
}

func (r *JtstestTestrunner_TestRun) GetPrecisionModel() *Geom_PrecisionModel {
	return r.precisionModel
}

func (r *JtstestTestrunner_TestRun) GetGeometryOperation() JtstestGeomop_GeometryOperation {
	// In Go port, we don't have JTSTestRunnerCmd, so just return the stored op.
	if r.geomOp == nil {
		return JtstestGeomop_NewGeometryMethodOperation()
	}
	return r.geomOp
}

func (r *JtstestTestrunner_TestRun) GetResultMatcher() JtstestTestrunner_ResultMatcher {
	// In Go port, we don't have JTSTestRunnerCmd, so just return the stored matcher.
	if r.resultMatcher == nil {
		return JtstestTestrunner_NewBufferResultMatcher()
	}
	return r.resultMatcher
}

func (r *JtstestTestrunner_TestRun) GetTestCases() []*JtstestTestrunner_TestCase {
	return r.testCases
}

func (r *JtstestTestrunner_TestRun) GetTestFile() string {
	return r.testFile
}

func (r *JtstestTestrunner_TestRun) GetTestFileName() string {
	if r.testFile == "" {
		return ""
	}
	// Extract just the filename from the path.
	for i := len(r.testFile) - 1; i >= 0; i-- {
		if r.testFile[i] == '/' || r.testFile[i] == '\\' {
			return r.testFile[i+1:]
		}
	}
	return r.testFile
}

func (r *JtstestTestrunner_TestRun) AddTestCase(testCase *JtstestTestrunner_TestCase) {
	r.testCases = append(r.testCases, testCase)
}

func (r *JtstestTestrunner_TestRun) Run() {
	for _, testCase := range r.testCases {
		if r.testCaseIndexToRun < 0 || testCase.GetCaseIndex() == r.testCaseIndexToRun {
			testCase.Run()
		}
	}
}
