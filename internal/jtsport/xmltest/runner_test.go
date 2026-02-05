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

							tst.Run()
							if tst.GetException() != nil {
								t.Fatalf("error: %v", tst.GetException())
							}
							test.True(t, tst.IsPassed())
						})
					}
				})
			}
		})
	}
}

func isUnsupportedOp(opName string) bool {
	// Operations that are still stubbed (not yet ported).
	unsupported := []string{
		"convexhull",       // algorithm/ConvexHull not ported
		"densify",          // densify/Densifier not ported
		"getcentroid",      // algorithm/Centroid not ported
		"getinteriorpoint", // algorithm/InteriorPoint not ported
		"minclearance",     // precision/MinimumClearance not ported
		"minclearanceline", // precision/MinimumClearance not ported
		"polygonize",       // operation/polygonize/Polygonizer not ported
		"simplifydp",       // simplify/DouglasPeuckerSimplifier not ported
		"simplifytp",       // simplify/TopologyPreservingSimplifier not ported
	}
	for _, u := range unsupported {
		if opName == u {
			return true
		}
	}
	return false
}
