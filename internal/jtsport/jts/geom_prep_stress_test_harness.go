package jts

import (
	"math/rand"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const geomPrep_StressTestHarness_MAX_ITER = 10000

var geomPrep_stressTestHarness_pm = Geom_NewPrecisionModel()
var geomPrep_stressTestHarness_fact = Geom_NewGeometryFactoryWithPrecisionModelAndSRID(geomPrep_stressTestHarness_pm, 0)
var geomPrep_stressTestHarness_wktRdr = Io_NewWKTReaderWithFactory(geomPrep_stressTestHarness_fact)
var geomPrep_stressTestHarness_wktWriter = Io_NewWKTWriter()

type GeomPrep_StressTestHarness struct {
	child        java.Polymorphic
	numTargetPts int
}

func (h *GeomPrep_StressTestHarness) GetChild() java.Polymorphic  { return h.child }
func (h *GeomPrep_StressTestHarness) GetParent() java.Polymorphic { return nil }

func GeomPrep_NewStressTestHarness() *GeomPrep_StressTestHarness {
	return &GeomPrep_StressTestHarness{
		numTargetPts: 1000,
	}
}

func (h *GeomPrep_StressTestHarness) SetTargetSize(nPts int) {
	h.numTargetPts = nPts
}

func (h *GeomPrep_StressTestHarness) Run(nIter int) {
	//System.out.println("Running " + nIter + " tests");
	//  Geometry poly = createCircle(new Coordinate(0, 0), 100, nPts);
	poly := h.createSineStar(Geom_NewCoordinateWithXY(0, 0), 100, h.numTargetPts)
	//System.out.println(poly);

	//System.out.println();
	//System.out.println("Running with " + nPts + " points");
	h.RunWithTarget(nIter, poly)
}

func (h *GeomPrep_StressTestHarness) createCircle(origin *Geom_Coordinate, size float64, nPts int) *Geom_Geometry {
	gsf := Util_NewGeometricShapeFactory()
	gsf.SetCentre(origin)
	gsf.SetSize(size)
	gsf.SetNumPoints(nPts)
	circle := gsf.CreateCircle()
	// Polygon gRect = gsf.createRectangle();
	// Geometry g = gRect.getExteriorRing();
	return circle.Geom_Geometry
}

func (h *GeomPrep_StressTestHarness) createSineStar(origin *Geom_Coordinate, size float64, nPts int) *Geom_Geometry {
	gsf := GeomUtil_NewSineStarFactory()
	gsf.SetCentre(origin)
	gsf.SetSize(size)
	gsf.SetNumPoints(nPts)
	gsf.SetArmLengthRatio(0.1)
	gsf.SetNumArms(20)
	poly := gsf.CreateSineStar()
	return poly
}

func (h *GeomPrep_StressTestHarness) createRandomTestGeometry(env *Geom_Envelope, size float64, nPts int) *Geom_Geometry {
	width := env.GetWidth()
	xOffset := width * rand.Float64()
	yOffset := env.GetHeight() * rand.Float64()
	basePt := Geom_NewCoordinateWithXY(
		env.GetMinX()+xOffset,
		env.GetMinY()+yOffset)
	test := h.createTestCircle(basePt, size, nPts)
	if java.InstanceOf[*Geom_Polygon](test) && rand.Float64() > 0.5 {
		test = test.GetBoundary()
	}
	return test
}

func (h *GeomPrep_StressTestHarness) createTestCircle(base *Geom_Coordinate, size float64, nPts int) *Geom_Geometry {
	gsf := Util_NewGeometricShapeFactory()
	gsf.SetCentre(base)
	gsf.SetSize(size)
	gsf.SetNumPoints(nPts)
	circle := gsf.CreateCircle()
	//    System.out.println(circle);
	return circle.Geom_Geometry
}

func (h *GeomPrep_StressTestHarness) RunWithTarget(nIter int, target *Geom_Geometry) {
	count := 0
	for count < nIter {
		count++
		test := h.createRandomTestGeometry(target.GetEnvelopeInternal(), 10, 20)

		//      System.out.println("Test # " + count);
		//  		System.out.println(line);
		//  		System.out.println("Test[" + count + "] " + target.getClass() + "/" + test.getClass());
		isResultCorrect := h.CheckResult(target, test)
		if !isResultCorrect {
			panic("Invalid result found")
		}
	}
}

func (h *GeomPrep_StressTestHarness) CheckResult(target, test *Geom_Geometry) bool {
	if impl, ok := java.GetLeaf(h).(interface {
		CheckResult_BODY(*Geom_Geometry, *Geom_Geometry) bool
	}); ok {
		return impl.CheckResult_BODY(target, test)
	}
	panic("abstract method called")
}
