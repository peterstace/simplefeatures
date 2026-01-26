package jts

// JtstestTestrunner_TestCase represents a set of tests for two geometries.
type JtstestTestrunner_TestCase struct {
	description string
	a           *Geom_Geometry
	b           *Geom_Geometry
	tests       []*JtstestTestrunner_Test
	testRun     *JtstestTestrunner_TestRun
	caseIndex   int
	lineNumber  int
	aWktFile    string
	bWktFile    string
	isRun       bool
}

// JtstestTestrunner_NewTestCase creates a TestCase with the given description.
// The tests will be applied to a and b.
func JtstestTestrunner_NewTestCase(
	description string,
	a *Geom_Geometry,
	b *Geom_Geometry,
	aWktFile string,
	bWktFile string,
	testRun *JtstestTestrunner_TestRun,
	caseIndex int,
	lineNumber int,
) *JtstestTestrunner_TestCase {
	return &JtstestTestrunner_TestCase{
		description: description,
		a:           a,
		b:           b,
		aWktFile:    aWktFile,
		bWktFile:    bWktFile,
		testRun:     testRun,
		caseIndex:   caseIndex,
		lineNumber:  lineNumber,
	}
}

func (tc *JtstestTestrunner_TestCase) GetLineNumber() int {
	return tc.lineNumber
}

func (tc *JtstestTestrunner_TestCase) SetGeometryA(a *Geom_Geometry) {
	tc.aWktFile = ""
	tc.a = a
}

func (tc *JtstestTestrunner_TestCase) SetGeometryB(b *Geom_Geometry) {
	tc.bWktFile = ""
	tc.b = b
}

func (tc *JtstestTestrunner_TestCase) SetDescription(description string) {
	tc.description = description
}

func (tc *JtstestTestrunner_TestCase) IsRun() bool {
	return tc.isRun
}

func (tc *JtstestTestrunner_TestCase) GetGeometryA() *Geom_Geometry {
	return tc.a
}

func (tc *JtstestTestrunner_TestCase) GetGeometryB() *Geom_Geometry {
	return tc.b
}

func (tc *JtstestTestrunner_TestCase) GetTestCount() int {
	return len(tc.tests)
}

func (tc *JtstestTestrunner_TestCase) GetTests() []*JtstestTestrunner_Test {
	return tc.tests
}

func (tc *JtstestTestrunner_TestCase) GetTestRun() *JtstestTestrunner_TestRun {
	return tc.testRun
}

func (tc *JtstestTestrunner_TestCase) GetCaseIndex() int {
	return tc.caseIndex
}

func (tc *JtstestTestrunner_TestCase) GetDescription() string {
	return tc.description
}

func (tc *JtstestTestrunner_TestCase) Add(test *JtstestTestrunner_Test) {
	tc.tests = append(tc.tests, test)
}

func (tc *JtstestTestrunner_TestCase) Remove(test *JtstestTestrunner_Test) {
	for i, t := range tc.tests {
		if t == test {
			tc.tests = append(tc.tests[:i], tc.tests[i+1:]...)
			return
		}
	}
}

func (tc *JtstestTestrunner_TestCase) Run() {
	tc.isRun = true
	for _, test := range tc.tests {
		test.Run()
	}
}

func (tc *JtstestTestrunner_TestCase) ToXml() string {
	xml := ""
	xml += "<case>\n"
	if tc.description != "" {
		xml += "  <desc>" + JtstestUtil_StringUtil_EscapeHTML(tc.description) + "</desc>\n"
	}
	xml += tc.xml("a", tc.a, tc.aWktFile) + "\n"
	xml += tc.xml("b", tc.b, tc.bWktFile)
	for _, test := range tc.tests {
		xml += test.ToXml()
	}
	xml += "</case>\n"
	return xml
}

func (tc *JtstestTestrunner_TestCase) xml(id string, g *Geom_Geometry, wktFile string) string {
	if g == nil {
		return ""
	}
	if wktFile != "" {
		return "  <" + id + " file=\"" + wktFile + "\"/>"
	}
	xml := ""
	xml += "  <" + id + ">\n"
	writer := Io_NewWKTWriter()
	xml += JtstestUtil_StringUtil_Indent(writer.WriteFormatted(g), 4) + "\n"
	xml += "  </" + id + ">\n"
	return xml
}
