package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// TRANSLITERATION NOTE: Java main() method (JUnit TestRunner entry point) not
// ported - Go uses `go test`.

type geomPrep_PreparedPolygonPredicateStressTest struct {
	testFailed bool
}

// TRANSLITERATION NOTE: Java constructor
// PreparedPolygonPredicateStressTest(String name) not ported - JUnit TestCase
// infrastructure not needed in Go.

func TestPreparedPolygonPredicateStress(t *testing.T) {
	st := &geomPrep_PreparedPolygonPredicateStressTest{}
	st.test()
}

func (st *geomPrep_PreparedPolygonPredicateStressTest) test() {
	tester := geomPrep_newPreparedPolygonPredicateStressTest_PredicateStressTester(st)
	tester.Run(1000)
}

type geomPrep_PreparedPolygonPredicateStressTest_PredicateStressTester struct {
	*GeomPrep_StressTestHarness
	child java.Polymorphic
	outer *geomPrep_PreparedPolygonPredicateStressTest
}

func (p *geomPrep_PreparedPolygonPredicateStressTest_PredicateStressTester) GetChild() java.Polymorphic {
	return p.child
}

func (p *geomPrep_PreparedPolygonPredicateStressTest_PredicateStressTester) GetParent() java.Polymorphic {
	return p.GeomPrep_StressTestHarness
}

func geomPrep_newPreparedPolygonPredicateStressTest_PredicateStressTester(outer *geomPrep_PreparedPolygonPredicateStressTest) *geomPrep_PreparedPolygonPredicateStressTest_PredicateStressTester {
	h := GeomPrep_NewStressTestHarness()
	pt := &geomPrep_PreparedPolygonPredicateStressTest_PredicateStressTester{
		GeomPrep_StressTestHarness: h,
		outer:                      outer,
	}
	h.child = pt
	return pt
}

func (p *geomPrep_PreparedPolygonPredicateStressTest_PredicateStressTester) CheckResult_BODY(target, test *Geom_Geometry) bool {
	if !p.outer.checkIntersects(target, test) {
		return false
	}
	if !p.outer.checkContains(target, test) {
		return false
	}
	return true
}

func (st *geomPrep_PreparedPolygonPredicateStressTest) checkContains(target, test *Geom_Geometry) bool {
	expectedResult := target.Contains(test)

	pgFact := GeomPrep_NewPreparedGeometryFactory()
	prepGeom := pgFact.Create(target)

	prepResult := prepGeom.Contains(test)

	if prepResult != expectedResult {
		return false
	}
	return true
}

func (st *geomPrep_PreparedPolygonPredicateStressTest) checkIntersects(target, test *Geom_Geometry) bool {
	expectedResult := target.Intersects(test)

	pgFact := GeomPrep_NewPreparedGeometryFactory()
	prepGeom := pgFact.Create(target)

	prepResult := prepGeom.Intersects(test)

	if prepResult != expectedResult {
		return false
	}
	return true
}
