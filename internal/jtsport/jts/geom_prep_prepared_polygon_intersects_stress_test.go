package jts

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const geomPrep_PreparedPolygonIntersectsStressTest_MAX_ITER = 10000

var geomPrep_preparedPolygonIntersectsStressTest_pm = Geom_NewPrecisionModel()
var geomPrep_preparedPolygonIntersectsStressTest_fact = Geom_NewGeometryFactoryWithPrecisionModelAndSRID(geomPrep_preparedPolygonIntersectsStressTest_pm, 0)
var geomPrep_preparedPolygonIntersectsStressTest_wktRdr = Io_NewWKTReaderWithFactory(geomPrep_preparedPolygonIntersectsStressTest_fact)
var geomPrep_preparedPolygonIntersectsStressTest_wktWriter = Io_NewWKTWriter()

// TRANSLITERATION NOTE: Java main() method (JUnit TestRunner entry point) not
// ported - Go uses `go test`.

type geomPrep_PreparedPolygonIntersectsStressTest struct {
	testFailed bool
}

// TRANSLITERATION NOTE: Java constructor
// PreparedPolygonIntersectsStressTest(String name) not ported - JUnit TestCase
// infrastructure not needed in Go.

func TestPreparedPolygonIntersectsStress(t *testing.T) {
	st := &geomPrep_PreparedPolygonIntersectsStressTest{}
	st.test()
}

func (st *geomPrep_PreparedPolygonIntersectsStressTest) test() {
	st.run(1000)
}

func (st *geomPrep_PreparedPolygonIntersectsStressTest) run(nPts int) {
	//  	Geometry poly = createCircle(new Coordinate(0, 0), 100, nPts);
	poly := st.createSineStar(Geom_NewCoordinateWithXY(0, 0), 100, nPts)
	//System.out.println(poly);

	//System.out.println();
	//System.out.println("Running with " + nPts + " points");
	st.testWithGeometry(poly)
}

func (st *geomPrep_PreparedPolygonIntersectsStressTest) createCircle(origin *Geom_Coordinate, size float64, nPts int) *Geom_Geometry {
	gsf := Util_NewGeometricShapeFactory()
	gsf.SetCentre(origin)
	gsf.SetSize(size)
	gsf.SetNumPoints(nPts)
	circle := gsf.CreateCircle()
	// Polygon gRect = gsf.createRectangle();
	// Geometry g = gRect.getExteriorRing();
	return circle.Geom_Geometry
}

func (st *geomPrep_PreparedPolygonIntersectsStressTest) createSineStar(origin *Geom_Coordinate, size float64, nPts int) *Geom_Geometry {
	gsf := GeomUtil_NewSineStarFactory()
	gsf.SetCentre(origin)
	gsf.SetSize(size)
	gsf.SetNumPoints(nPts)
	gsf.SetArmLengthRatio(0.1)
	gsf.SetNumArms(20)
	poly := gsf.CreateSineStar()
	return poly
}

func (st *geomPrep_PreparedPolygonIntersectsStressTest) createTestLineFromEnvelope(env *Geom_Envelope, size float64, nPts int) *Geom_LineString {
	width := env.GetWidth()
	xOffset := width * rand.Float64()
	yOffset := env.GetHeight() * rand.Float64()
	basePt := Geom_NewCoordinateWithXY(
		env.GetMinX()+xOffset,
		env.GetMinY()+yOffset)
	line := st.createTestLineFromCoordinate(basePt, size, nPts)
	return line
}

func (st *geomPrep_PreparedPolygonIntersectsStressTest) createTestLineFromCoordinate(base *Geom_Coordinate, size float64, nPts int) *Geom_LineString {
	gsf := Util_NewGeometricShapeFactory()
	gsf.SetCentre(base)
	gsf.SetSize(size)
	gsf.SetNumPoints(nPts)
	circle := gsf.CreateCircle()
	//    System.out.println(circle);
	return java.Cast[*Geom_LineString](circle.Geom_Geometry.GetBoundary())
}

func (st *geomPrep_PreparedPolygonIntersectsStressTest) testWithGeometry(g *Geom_Geometry) {
	count := 0
	for count < geomPrep_PreparedPolygonIntersectsStressTest_MAX_ITER {
		count++
		line := st.createTestLineFromEnvelope(g.GetEnvelopeInternal(), 10, 20)

		//      System.out.println("Test # " + count);
		//  		System.out.println(line);
		st.testResultsEqual(g, line)
	}
}

func (st *geomPrep_PreparedPolygonIntersectsStressTest) testResultsEqual(g *Geom_Geometry, line *Geom_LineString) {
	slowIntersects := g.Intersects(line.Geom_Geometry)

	pgFact := GeomPrep_NewPreparedGeometryFactory()
	prepGeom := pgFact.Create(g)

	fastIntersects := prepGeom.Intersects(line.Geom_Geometry)

	if slowIntersects != fastIntersects {
		fmt.Println(line)
		fmt.Printf("Slow = %v, Fast = %v\n", slowIntersects, fastIntersects)
		panic("Different results found for intersects() !")
	}
}
