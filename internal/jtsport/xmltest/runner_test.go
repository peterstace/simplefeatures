// This file is a Go test harness that exercises the ported JTS XML test runner
// classes. It does not correspond to any specific Java file in JTS - it is
// roughly equivalent to running JTSTestRunnerCmd from the command line.

package xmltest_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/test"
)

func TestXMLTestSuite(t *testing.T) {
	testXMLDirs := []string{
		"testdata/general",
		"testdata/validate",
	}

	var xmlFiles []string
	for _, dir := range testXMLDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.xml"))
		test.NoErr(t, err)
		xmlFiles = append(xmlFiles, files...)
	}
	test.True(t, len(xmlFiles) > 0)

	reader := jts.JtstestTestrunner_NewTestReader()
	for _, xmlFile := range xmlFiles {
		testRun := reader.CreateTestRun(xmlFile, 0)
		test.True(t, testRun != nil)

		fileName := filepath.Base(xmlFile)
		t.Run(fileName, func(t *testing.T) {
			for _, testCase := range testRun.GetTestCases() {
				t.Run(fmt.Sprintf("Case%d", testCase.GetCaseIndex()), func(t *testing.T) {
					for _, tst := range testCase.GetTests() {
						t.Run(fmt.Sprintf("Test%d_%s", tst.GetTestIndex(), tst.GetOperation()), func(t *testing.T) {
							opName := strings.ToLower(tst.GetOperation())

							// Skip unsupported operations.
							if isUnsupportedOp(opName) {
								t.Skip("unsupported operation")
							}

							// Run test with panic recovery.
							passed, err, panicked := runTestWithRecovery(tst)
							if panicked != nil {
								t.Fatalf("PANIC: %v", panicked)
							}

							if err != nil {
								t.Fatalf("error: %v", err)
							}

							if !passed {
								t.Fatal("test failed")
							}
						})
					}
				})
			}
		})
	}
}

func runTestWithRecovery(test *jts.JtstestTestrunner_Test) (passed bool, err error, panicked any) { //nolint:revive,stylecheck
	defer func() {
		if r := recover(); r != nil {
			panicked = r
		}
	}()

	test.Run()
	if test.GetException() != nil {
		return false, fmt.Errorf("%w", test.GetException()), nil
	}
	return test.IsPassed(), nil, nil
}

func isUnsupportedOp(opName string) bool {
	unsupported := []string{
		"buffer",
		"buffermitredjoin",
		"convexhull",
		"densify",
		"distance",
		"getcentroid",
		"getinteriorpoint",
		"getlength",
		"isvalid",
		"iswithindistance",
		"minclearance",
		"minclearanceline",
		"polygonize",
		"simplifydp",
		"simplifytp",
	}
	for _, u := range unsupported {
		if opName == u {
			return true
		}
	}
	return false
}
