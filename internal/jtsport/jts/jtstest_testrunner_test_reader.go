package jts

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	jtstestTestrunner_testReader_TAG_geometryOperation = "geometryOperation"
	jtstestTestrunner_testReader_TAG_resultMatcher     = "resultMatcher"
)

// JtstestTestrunner_TestReader reads XML test files and creates TestRun objects.
type JtstestTestrunner_TestReader struct {
	parsingProblems []string
	geometryFactory *Geom_GeometryFactory
	wktorbReader    *JtstestUtilIo_WKTOrWKBReader
	tolerance       float64
	geomOp          JtstestGeomop_GeometryOperation
	resultMatcher   JtstestTestrunner_ResultMatcher
}

func JtstestTestrunner_NewTestReader() *JtstestTestrunner_TestReader {
	return &JtstestTestrunner_TestReader{}
}

func (r *JtstestTestrunner_TestReader) getGeometryOperation() JtstestGeomop_GeometryOperation {
	if r.geomOp == nil {
		return JtstestGeomop_NewGeometryMethodOperation()
	}
	return r.geomOp
}

func (r *JtstestTestrunner_TestReader) isBooleanFunction(name string) bool {
	return r.getGeometryOperation().GetReturnType(name) == "boolean"
}

func (r *JtstestTestrunner_TestReader) isIntegerFunction(name string) bool {
	return r.getGeometryOperation().GetReturnType(name) == "int"
}

func (r *JtstestTestrunner_TestReader) isDoubleFunction(name string) bool {
	return r.getGeometryOperation().GetReturnType(name) == "double"
}

func (r *JtstestTestrunner_TestReader) isGeometryFunction(name string) bool {
	return r.getGeometryOperation().GetReturnType(name) == "geometry"
}

func (r *JtstestTestrunner_TestReader) GetParsingProblems() []string {
	return r.parsingProblems
}

func (r *JtstestTestrunner_TestReader) ClearParsingProblems() {
	r.parsingProblems = nil
}

// XML element types for parsing.
type xmlRun struct {
	XMLName        xml.Name           `xml:"run"`
	Desc           string             `xml:"desc"`
	Workspace      *xmlWorkspace      `xml:"workspace"`
	Tolerance      *string            `xml:"tolerance"`
	PrecisionModel *xmlPrecisionModel `xml:"precisionModel"`
	GeomOperation  *string            `xml:"geometryOperation"`
	ResultMatcher  *string            `xml:"resultMatcher"`
	Cases          []xmlCase          `xml:"case"`
}

type xmlWorkspace struct {
	Dir string `xml:"dir,attr"`
}

type xmlPrecisionModel struct {
	Type  string `xml:"type,attr"`
	Scale string `xml:"scale,attr"`
}

type xmlCase struct {
	Desc  string    `xml:"desc"`
	A     *xmlGeom  `xml:"a"`
	B     *xmlGeom  `xml:"b"`
	Tests []xmlTest `xml:"test"`
}

type xmlGeom struct {
	File string `xml:"file,attr"`
	WKT  string `xml:",chardata"`
}

type xmlTest struct {
	Desc string `xml:"desc"`
	Op   xmlOp  `xml:"op"`
}

type xmlOp struct {
	Name    string `xml:"name,attr"`
	Arg1    string `xml:"arg1,attr"`
	Arg2    string `xml:"arg2,attr"`
	Arg3    string `xml:"arg3,attr"`
	Pattern string `xml:"pattern,attr"`
	Result  string `xml:",chardata"`
}

func (r *JtstestTestrunner_TestReader) CreateTestRun(testFile string, runIndex int) *JtstestTestrunner_TestRun {
	data, err := os.ReadFile(testFile)
	if err != nil {
		r.parsingProblems = append(r.parsingProblems,
			fmt.Sprintf("An exception occurred while parsing %s: %v", testFile, err))
		return nil
	}

	var runElement xmlRun
	if err := xml.Unmarshal(data, &runElement); err != nil {
		r.parsingProblems = append(r.parsingProblems,
			fmt.Sprintf("An exception occurred while parsing %s: %v", testFile, err))
		return nil
	}

	testRun, err := r.parseTestRun(&runElement, testFile, runIndex)
	if err != nil {
		r.parsingProblems = append(r.parsingProblems,
			fmt.Sprintf("An exception occurred while parsing %s: %v", testFile, err))
		return nil
	}

	return testRun
}

func (r *JtstestTestrunner_TestReader) parseTestRun(runElement *xmlRun, testFile string, runIndex int) (*JtstestTestrunner_TestRun, error) {
	// Parse workspace.
	var workspace string
	if runElement.Workspace != nil {
		workspace = runElement.Workspace.Dir
		if workspace != "" {
			info, err := os.Stat(workspace)
			if err != nil {
				return nil, &JtstestTestrunner_TestParseException{
					message: fmt.Sprintf("<workspace> does not exist: %s", workspace),
				}
			}
			if !info.IsDir() {
				return nil, &JtstestTestrunner_TestParseException{
					message: fmt.Sprintf("<workspace> is not a directory: %s", workspace),
				}
			}
		}
	}

	// Parse tolerance.
	r.tolerance = r.parseTolerance(runElement)

	// Parse geometry operation.
	r.geomOp = r.parseGeometryOperation(runElement)

	// Parse result matcher.
	r.resultMatcher = r.parseResultMatcher(runElement)

	// Parse precision model.
	precisionModel := r.parsePrecisionModel(runElement.PrecisionModel)

	// Build TestRun.
	testRun := JtstestTestrunner_NewTestRun(
		runElement.Desc,
		runIndex,
		precisionModel,
		r.geomOp,
		r.resultMatcher,
		testFile,
	)
	testRun.SetWorkspace(workspace)

	if len(runElement.Cases) == 0 {
		return nil, &JtstestTestrunner_TestParseException{message: "Missing <case> in <run>"}
	}

	// Parse test cases.
	testCases, err := r.parseTestCases(runElement.Cases, testFile, testRun, r.tolerance)
	if err != nil {
		return nil, err
	}

	for _, testCase := range testCases {
		testRun.AddTestCase(testCase)
	}

	return testRun, nil
}

func (r *JtstestTestrunner_TestReader) parsePrecisionModel(pm *xmlPrecisionModel) *Geom_PrecisionModel {
	if pm == nil {
		return Geom_NewPrecisionModel()
	}
	if pm.Scale != "" {
		scale, err := strconv.ParseFloat(pm.Scale, 64)
		if err != nil {
			return Geom_NewPrecisionModel()
		}
		return Geom_NewPrecisionModelWithScale(scale)
	}
	if strings.EqualFold(pm.Type, "FIXED") {
		return Geom_NewPrecisionModel()
	}
	return Geom_NewPrecisionModel()
}

func (r *JtstestTestrunner_TestReader) parseGeometryOperation(runElement *xmlRun) JtstestGeomop_GeometryOperation {
	if runElement.GeomOperation == nil {
		return nil
	}
	goClass := strings.TrimSpace(*runElement.GeomOperation)
	geomOp := r.getInstance(goClass, "GeometryOperation")
	if geomOp == nil {
		r.parsingProblems = append(r.parsingProblems,
			fmt.Sprintf("Could not create instance of GeometryOperation from class %s", goClass))
		return nil
	}
	return geomOp.(JtstestGeomop_GeometryOperation)
}

func (r *JtstestTestrunner_TestReader) parseResultMatcher(runElement *xmlRun) JtstestTestrunner_ResultMatcher {
	if runElement.ResultMatcher == nil {
		return nil
	}
	goClass := strings.TrimSpace(*runElement.ResultMatcher)
	resultMatcher := r.getInstance(goClass, "ResultMatcher")
	if resultMatcher == nil {
		r.parsingProblems = append(r.parsingProblems,
			fmt.Sprintf("Could not create instance of ResultMatcher from class %s", goClass))
		return nil
	}
	return resultMatcher.(JtstestTestrunner_ResultMatcher)
}

func (r *JtstestTestrunner_TestReader) getInstance(classname string, baseClass string) any {
	// TRANSLITERATION NOTE: Go doesn't support dynamic class loading like Java's reflection.
	// Instead, we use explicit type mapping for known classes.
	switch classname {
	case "org.locationtech.jtstest.geomop.PreparedGeometryOperation":
		return JtstestGeomop_NewPreparedGeometryOperation()
	}
	return nil
}

func (r *JtstestTestrunner_TestReader) parseTolerance(runElement *xmlRun) float64 {
	tolerance := 0.0
	if runElement.Tolerance != nil {
		tol, err := strconv.ParseFloat(strings.TrimSpace(*runElement.Tolerance), 64)
		if err != nil {
			r.parsingProblems = append(r.parsingProblems,
				fmt.Sprintf("Could not parse tolerance from string: %s", *runElement.Tolerance))
			return 0.0
		}
		tolerance = tol
	}
	return tolerance
}

func (r *JtstestTestrunner_TestReader) createPrecisionModel(precisionModelElement *xmlPrecisionModel) (*Geom_PrecisionModel, error) {
	if precisionModelElement.Scale == "" {
		return nil, &JtstestTestrunner_TestParseException{
			message: "Missing scale attribute in <precisionModel>",
		}
	}
	scale, err := strconv.ParseFloat(precisionModelElement.Scale, 64)
	if err != nil {
		return nil, &JtstestTestrunner_TestParseException{
			message: fmt.Sprintf("Could not convert scale attribute to double: %s", precisionModelElement.Scale),
		}
	}
	return Geom_NewPrecisionModelWithScale(scale), nil
}

func (r *JtstestTestrunner_TestReader) parseTestCases(
	caseElements []xmlCase,
	testFile string,
	testRun *JtstestTestrunner_TestRun,
	tolerance float64,
) ([]*JtstestTestrunner_TestCase, error) {
	r.geometryFactory = Geom_NewGeometryFactoryWithPrecisionModel(testRun.GetPrecisionModel())
	r.wktorbReader = JtstestUtilIo_NewWKTOrWKBReaderWithFactory(r.geometryFactory)

	var testCases []*JtstestTestrunner_TestCase
	for caseIndex, caseElement := range caseElements {
		caseNum := caseIndex + 1
		tc, err := r.parseTestCase(&caseElement, testFile, testRun, caseNum, tolerance)
		if err != nil {
			r.parsingProblems = append(r.parsingProblems,
				fmt.Sprintf("An exception occurred while parsing <case> %d in %s: %v", caseNum, testFile, err))
			continue
		}
		testCases = append(testCases, tc)
	}
	return testCases, nil
}

func (r *JtstestTestrunner_TestReader) parseTestCase(
	caseElement *xmlCase,
	testFile string,
	testRun *JtstestTestrunner_TestRun,
	caseIndex int,
	tolerance float64,
) (*JtstestTestrunner_TestCase, error) {
	aWktFile := r.wktFile(caseElement.A, testRun)
	bWktFile := r.wktFile(caseElement.B, testRun)

	a, err := r.readGeometry(caseElement.A, r.absoluteWktFile(aWktFile, testRun))
	if err != nil {
		return nil, err
	}
	b, err := r.readGeometry(caseElement.B, r.absoluteWktFile(bWktFile, testRun))
	if err != nil {
		return nil, err
	}

	testCase := JtstestTestrunner_NewTestCase(
		caseElement.Desc,
		a,
		b,
		aWktFile,
		bWktFile,
		testRun,
		caseIndex,
		0, // Line number not tracked in Go XML parser.
	)

	tests, err := r.parseTests(caseElement.Tests, caseIndex, testFile, testCase, tolerance)
	if err != nil {
		return nil, err
	}

	for _, test := range tests {
		testCase.Add(test)
	}

	return testCase, nil
}

func (r *JtstestTestrunner_TestReader) parseTests(
	testElements []xmlTest,
	caseIndex int,
	testFile string,
	testCase *JtstestTestrunner_TestCase,
	tolerance float64,
) ([]*JtstestTestrunner_Test, error) {
	var tests []*JtstestTestrunner_Test
	for testIndex, testElement := range testElements {
		testNum := testIndex + 1
		test, err := r.parseTest(&testElement, testCase, testNum, tolerance)
		if err != nil {
			r.parsingProblems = append(r.parsingProblems,
				fmt.Sprintf("An exception occurred while parsing <test> %d in <case> %d in %s: %v",
					testNum, caseIndex, testFile, err))
			continue
		}
		tests = append(tests, test)
	}
	return tests, nil
}

func (r *JtstestTestrunner_TestReader) parseTest(
	testElement *xmlTest,
	testCase *JtstestTestrunner_TestCase,
	testIndex int,
	tolerance float64,
) (*JtstestTestrunner_Test, error) {
	opElement := &testElement.Op
	if opElement.Name == "" {
		return nil, &JtstestTestrunner_TestParseException{message: "Missing name attribute in <op>"}
	}

	arg1 := opElement.Arg1
	if arg1 == "" {
		arg1 = "A"
	}

	arg2 := strings.TrimSpace(opElement.Arg2)
	arg3 := strings.TrimSpace(opElement.Arg3)

	// Handle relate pattern.
	if arg3 == "" && strings.EqualFold(opElement.Name, "relate") {
		arg3 = strings.TrimSpace(opElement.Pattern)
	}

	var arguments []string
	if arg2 != "" {
		arguments = append(arguments, arg2)
	}
	if arg3 != "" {
		arguments = append(arguments, arg3)
	}

	result, err := r.toResult(strings.TrimSpace(opElement.Result), strings.TrimSpace(opElement.Name), testCase.GetTestRun())
	if err != nil {
		return nil, err
	}

	test := JtstestTestrunner_NewTest(
		testCase,
		testIndex,
		testElement.Desc,
		strings.TrimSpace(opElement.Name),
		strings.TrimSpace(arg1),
		arguments,
		result,
		tolerance,
	)

	return test, nil
}

func (r *JtstestTestrunner_TestReader) toResult(value, name string, testRun *JtstestTestrunner_TestRun) (JtstestTestrunner_Result, error) {
	if value == "" {
		return nil, nil
	}
	if r.isBooleanFunction(name) {
		return r.toBooleanResult(value)
	}
	if r.isIntegerFunction(name) {
		return r.toIntegerResult(value)
	}
	if r.isDoubleFunction(name) {
		return r.toDoubleResult(value)
	}
	if r.isGeometryFunction(name) {
		return r.toGeometryResult(value, testRun)
	}
	return nil, nil
}

func (r *JtstestTestrunner_TestReader) toBooleanResult(value string) (JtstestTestrunner_Result, error) {
	if strings.EqualFold(value, "true") {
		return JtstestTestrunner_NewBooleanResult(true), nil
	}
	if strings.EqualFold(value, "false") {
		return JtstestTestrunner_NewBooleanResult(false), nil
	}
	return nil, &JtstestTestrunner_TestParseException{
		message: fmt.Sprintf("Expected 'true' or 'false' but encountered '%s'", value),
	}
}

func (r *JtstestTestrunner_TestReader) toDoubleResult(value string) (JtstestTestrunner_Result, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, &JtstestTestrunner_TestParseException{
			message: fmt.Sprintf("Expected double but encountered '%s'", value),
		}
	}
	return JtstestTestrunner_NewDoubleResult(f), nil
}

func (r *JtstestTestrunner_TestReader) toIntegerResult(value string) (JtstestTestrunner_Result, error) {
	i, err := strconv.Atoi(value)
	if err != nil {
		return nil, &JtstestTestrunner_TestParseException{
			message: fmt.Sprintf("Expected integer but encountered '%s'", value),
		}
	}
	return JtstestTestrunner_NewIntegerResult(i), nil
}

func (r *JtstestTestrunner_TestReader) toGeometryResult(value string, testRun *JtstestTestrunner_TestRun) (JtstestTestrunner_Result, error) {
	geometryFactory := Geom_NewGeometryFactoryWithPrecisionModel(testRun.GetPrecisionModel())
	wktorbReader := JtstestUtilIo_NewWKTOrWKBReaderWithFactory(geometryFactory)
	geom, err := wktorbReader.Read(value)
	if err != nil {
		return nil, err
	}
	return JtstestTestrunner_NewGeometryResult(geom), nil
}

func (r *JtstestTestrunner_TestReader) wktFile(geomElement *xmlGeom, testRun *JtstestTestrunner_TestRun) string {
	if geomElement == nil {
		return ""
	}
	return strings.TrimSpace(geomElement.File)
}

func (r *JtstestTestrunner_TestReader) readGeometry(geomElement *xmlGeom, wktFile string) (*Geom_Geometry, error) {
	var geomText string
	if wktFile != "" {
		wktList, err := jtstestTestrunner_testReader_getContents(wktFile)
		if err != nil {
			return nil, err
		}
		geomText = r.toString(wktList)
	} else {
		if geomElement == nil {
			return nil, nil
		}
		geomText = strings.TrimSpace(geomElement.WKT)
	}
	return r.wktorbReader.Read(geomText)
	// TRANSLITERATION NOTE: Java has commented code for WKB support:
	// if (isHex(geomText, 6))
	//   return wkbReader.read(WKBReader.hexToBytes(geomText));
	// return wktReader.read(geomText);
}

func (r *JtstestTestrunner_TestReader) toString(stringList []string) string {
	result := ""
	for _, line := range stringList {
		result += line + "\n"
	}
	return result
}

func (r *JtstestTestrunner_TestReader) absoluteWktFile(wktFile string, testRun *JtstestTestrunner_TestRun) string {
	if wktFile == "" {
		return ""
	}
	if filepath.IsAbs(wktFile) {
		return wktFile
	}
	var dir string
	if testRun.GetWorkspace() != "" {
		dir = testRun.GetWorkspace()
	} else {
		dir = filepath.Dir(testRun.GetTestFile())
	}
	return filepath.Join(dir, filepath.Base(wktFile))
}

func jtstestTestrunner_testReader_getContents(textFileName string) ([]string, error) {
	data, err := os.ReadFile(textFileName)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	return lines, nil
}
