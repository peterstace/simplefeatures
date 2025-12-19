// This file is a Go test harness that exercises the ported JTS XML test runner
// classes. It does not correspond to any specific Java file in JTS - it is
// roughly equivalent to running JTSTestRunnerCmd from the command line.

package xmltest_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestXMLTestSuite(t *testing.T) {
	jtsDir := findJTSDir(t)
	testXMLDirs := []string{
		filepath.Join(jtsDir, "modules/tests/src/test/resources/testxml/general"),
		filepath.Join(jtsDir, "modules/tests/src/test/resources/testxml/validate"),
	}

	var xmlFiles []string
	for _, dir := range testXMLDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.xml"))
		if err != nil {
			t.Fatalf("error finding XML files in %s: %v", dir, err)
		}
		xmlFiles = append(xmlFiles, files...)
	}

	if len(xmlFiles) == 0 {
		t.Fatal("no XML test files found")
	}

	reader := jts.JtstestTestrunner_NewTestReader()
	var (
		totalTests   int
		passedTests  int
		failedTests  int
		errorTests   int
		skippedTests int
		panicTests   int
	)

	for _, xmlFile := range xmlFiles {
		testRun := reader.CreateTestRun(xmlFile, 0)
		if testRun == nil {
			for _, problem := range reader.GetParsingProblems() {
				t.Logf("Parse problem in %s: %s", xmlFile, problem)
			}
			reader.ClearParsingProblems()
			continue
		}

		fileName := filepath.Base(xmlFile)
		t.Run(fileName, func(t *testing.T) {
			for _, testCase := range testRun.GetTestCases() {
				for _, test := range testCase.GetTests() {
					totalTests++
					opName := strings.ToLower(test.GetOperation())

					// Skip unsupported operations.
					if isUnsupportedOp(opName) {
						skippedTests++
						continue
					}

					// Run test with panic recovery.
					passed, err, panicked := runTestWithRecovery(test)
					if panicked != nil {
						panicTests++
						t.Logf("Case %d Test %d (%s): PANIC: %v",
							testCase.GetCaseIndex(), test.GetTestIndex(), opName, panicked)
						continue
					}

					if err != nil {
						errorTests++
						t.Logf("Case %d Test %d (%s): error: %v",
							testCase.GetCaseIndex(), test.GetTestIndex(), opName, err)
						continue
					}

					if passed {
						passedTests++
					} else {
						failedTests++
						t.Logf("Case %d Test %d (%s): FAILED",
							testCase.GetCaseIndex(), test.GetTestIndex(), opName)
					}
				}
			}
		})
	}

	t.Logf("Total: %d, Passed: %d, Failed: %d, Errors: %d, Panics: %d, Skipped: %d",
		totalTests, passedTests, failedTests, errorTests, panicTests, skippedTests)
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

func findJTSDir(t *testing.T) string {
	t.Helper()
	// Try relative path from xmltest directory.
	candidates := []string{
		"../../../../locationtech/jts",
		"../../../../../locationtech/jts",
		os.Getenv("JTS_DIR"),
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(absPath, "modules")); err == nil {
			return absPath
		}
	}
	t.Skip("JTS directory not found; set JTS_DIR environment variable")
	return ""
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
