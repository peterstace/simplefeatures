package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestEnvelopeEverything(t *testing.T) {
	e1 := Geom_NewEnvelope()
	junit.AssertTrue(t, e1.IsNull())
	junit.AssertEquals(t, 0.0, e1.GetWidth())
	junit.AssertEquals(t, 0.0, e1.GetHeight())
	e1.ExpandToIncludeXY(100, 101)
	e1.ExpandToIncludeXY(200, 202)
	e1.ExpandToIncludeXY(150, 151)
	junit.AssertEquals(t, 200.0, e1.GetMaxX())
	junit.AssertEquals(t, 202.0, e1.GetMaxY())
	junit.AssertEquals(t, 100.0, e1.GetMinX())
	junit.AssertEquals(t, 101.0, e1.GetMinY())
	junit.AssertTrue(t, e1.ContainsXY(120, 120))
	junit.AssertTrue(t, e1.ContainsXY(120, 101))
	junit.AssertTrue(t, !e1.ContainsXY(120, 100))
	junit.AssertEquals(t, 101.0, e1.GetHeight())
	junit.AssertEquals(t, 100.0, e1.GetWidth())
	junit.AssertTrue(t, !e1.IsNull())

	e2 := Geom_NewEnvelopeFromXY(499, 500, 500, 501)
	junit.AssertTrue(t, !e1.ContainsEnvelope(e2))
	junit.AssertTrue(t, !e1.IntersectsEnvelope(e2))
	e1.ExpandToIncludeEnvelope(e2)
	junit.AssertTrue(t, e1.ContainsEnvelope(e2))
	junit.AssertTrue(t, e1.IntersectsEnvelope(e2))
	junit.AssertEquals(t, 500.0, e1.GetMaxX())
	junit.AssertEquals(t, 501.0, e1.GetMaxY())
	junit.AssertEquals(t, 100.0, e1.GetMinX())
	junit.AssertEquals(t, 101.0, e1.GetMinY())

	e3 := Geom_NewEnvelopeFromXY(300, 700, 300, 700)
	junit.AssertTrue(t, !e1.ContainsEnvelope(e3))
	junit.AssertTrue(t, e1.IntersectsEnvelope(e3))

	e4 := Geom_NewEnvelopeFromXY(300, 301, 300, 301)
	junit.AssertTrue(t, e1.ContainsEnvelope(e4))
	junit.AssertTrue(t, e1.IntersectsEnvelope(e4))
}

func TestEnvelopeIntersects(t *testing.T) {
	checkIntersectsPermuted(t, 1, 1, 2, 2, 2, 2, 3, 3, true)
	checkIntersectsPermuted(t, 1, 1, 2, 2, 3, 3, 4, 4, false)
}

func TestEnvelopeIntersectsEmpty(t *testing.T) {
	junit.AssertTrue(t, !Geom_NewEnvelopeFromXY(-5, 5, -5, 5).IntersectsEnvelope(Geom_NewEnvelope()))
	junit.AssertTrue(t, !Geom_NewEnvelope().IntersectsEnvelope(Geom_NewEnvelopeFromXY(-5, 5, -5, 5)))
	junit.AssertTrue(t, !Geom_NewEnvelope().IntersectsEnvelope(Geom_NewEnvelopeFromXY(100, 101, 100, 101)))
	junit.AssertTrue(t, !Geom_NewEnvelopeFromXY(100, 101, 100, 101).IntersectsEnvelope(Geom_NewEnvelope()))
}

func TestEnvelopeDisjointEmpty(t *testing.T) {
	junit.AssertTrue(t, Geom_NewEnvelopeFromXY(-5, 5, -5, 5).Disjoint(Geom_NewEnvelope()))
	junit.AssertTrue(t, Geom_NewEnvelope().Disjoint(Geom_NewEnvelopeFromXY(-5, 5, -5, 5)))
	junit.AssertTrue(t, Geom_NewEnvelope().Disjoint(Geom_NewEnvelopeFromXY(100, 101, 100, 101)))
	junit.AssertTrue(t, Geom_NewEnvelopeFromXY(100, 101, 100, 101).Disjoint(Geom_NewEnvelope()))
}

func TestEnvelopeContainsEmpty(t *testing.T) {
	junit.AssertTrue(t, !Geom_NewEnvelopeFromXY(-5, 5, -5, 5).ContainsEnvelope(Geom_NewEnvelope()))
	junit.AssertTrue(t, !Geom_NewEnvelope().ContainsEnvelope(Geom_NewEnvelopeFromXY(-5, 5, -5, 5)))
	junit.AssertTrue(t, !Geom_NewEnvelope().ContainsEnvelope(Geom_NewEnvelopeFromXY(100, 101, 100, 101)))
	junit.AssertTrue(t, !Geom_NewEnvelopeFromXY(100, 101, 100, 101).ContainsEnvelope(Geom_NewEnvelope()))
}

func TestEnvelopeExpandToIncludeEmpty(t *testing.T) {
	junit.AssertEqualsDeep(t, Geom_NewEnvelopeFromXY(-5, 5, -5, 5), expandToInclude(Geom_NewEnvelopeFromXY(-5, 5, -5, 5), Geom_NewEnvelope()))
	junit.AssertEqualsDeep(t, Geom_NewEnvelopeFromXY(-5, 5, -5, 5), expandToInclude(Geom_NewEnvelope(), Geom_NewEnvelopeFromXY(-5, 5, -5, 5)))
	junit.AssertEqualsDeep(t, Geom_NewEnvelopeFromXY(100, 101, 100, 101), expandToInclude(Geom_NewEnvelope(), Geom_NewEnvelopeFromXY(100, 101, 100, 101)))
	junit.AssertEqualsDeep(t, Geom_NewEnvelopeFromXY(100, 101, 100, 101), expandToInclude(Geom_NewEnvelopeFromXY(100, 101, 100, 101), Geom_NewEnvelope()))
}

func expandToInclude(a, b *Geom_Envelope) *Geom_Envelope {
	a.ExpandToIncludeEnvelope(b)
	return a
}

func TestEnvelopeEmpty(t *testing.T) {
	junit.AssertEquals(t, 0.0, Geom_NewEnvelope().GetHeight())
	junit.AssertEquals(t, 0.0, Geom_NewEnvelope().GetWidth())
	junit.AssertEqualsDeep(t, Geom_NewEnvelope(), Geom_NewEnvelope())
	e := Geom_NewEnvelopeFromXY(100, 101, 100, 101)
	e.InitFromEnvelope(Geom_NewEnvelope())
	junit.AssertEqualsDeep(t, Geom_NewEnvelope(), e)
}

func TestEnvelopeSetToNull(t *testing.T) {
	e1 := Geom_NewEnvelope()
	junit.AssertTrue(t, e1.IsNull())
	e1.ExpandToIncludeXY(5, 5)
	junit.AssertTrue(t, !e1.IsNull())
	e1.SetToNull()
	junit.AssertTrue(t, e1.IsNull())
}

func TestEnvelopeEquals(t *testing.T) {
	e1 := Geom_NewEnvelopeFromXY(1, 2, 3, 4)
	e2 := Geom_NewEnvelopeFromXY(1, 2, 3, 4)
	junit.AssertEqualsDeep(t, e1, e2)
	junit.AssertEquals(t, e1.HashCode(), e2.HashCode())

	e3 := Geom_NewEnvelopeFromXY(1, 2, 3, 5)
	junit.AssertTrue(t, !e1.Equals(e3))
	junit.AssertTrue(t, e1.HashCode() != e3.HashCode())
	e1.SetToNull()
	junit.AssertTrue(t, !e1.Equals(e2))
	junit.AssertTrue(t, e1.HashCode() != e2.HashCode())
	e2.SetToNull()
	junit.AssertEqualsDeep(t, e1, e2)
	junit.AssertEquals(t, e1.HashCode(), e2.HashCode())
}

func TestEnvelopeEquals2(t *testing.T) {
	junit.AssertTrue(t, Geom_NewEnvelope().Equals(Geom_NewEnvelope()))
	junit.AssertTrue(t, Geom_NewEnvelopeFromXY(1, 2, 1, 2).Equals(Geom_NewEnvelopeFromXY(1, 2, 1, 2)))
	junit.AssertTrue(t, !Geom_NewEnvelopeFromXY(1, 2, 1.5, 2).Equals(Geom_NewEnvelopeFromXY(1, 2, 1, 2)))
}

func TestEnvelopeCopyConstructor(t *testing.T) {
	e1 := Geom_NewEnvelopeFromXY(1, 2, 3, 4)
	e2 := Geom_NewEnvelopeFromEnvelope(e1)
	junit.AssertEquals(t, 1.0, e2.GetMinX())
	junit.AssertEquals(t, 2.0, e2.GetMaxX())
	junit.AssertEquals(t, 3.0, e2.GetMinY())
	junit.AssertEquals(t, 4.0, e2.GetMaxY())
}

func TestEnvelopeCopy(t *testing.T) {
	e1 := Geom_NewEnvelopeFromXY(1, 2, 3, 4)
	e2 := e1.Copy()
	junit.AssertEquals(t, 1.0, e2.GetMinX())
	junit.AssertEquals(t, 2.0, e2.GetMaxX())
	junit.AssertEquals(t, 3.0, e2.GetMinY())
	junit.AssertEquals(t, 4.0, e2.GetMaxY())

	eNull := Geom_NewEnvelope()
	eNullCopy := eNull.Copy()
	junit.AssertTrue(t, eNullCopy.IsNull())
}

func TestEnvelopeMetrics(t *testing.T) {
	env := Geom_NewEnvelopeFromXY(0, 4, 0, 3)
	junit.AssertEquals(t, 4.0, env.GetWidth())
	junit.AssertEquals(t, 3.0, env.GetHeight())
	junit.AssertEquals(t, 5.0, env.GetDiameter())
}

func TestEnvelopeEmptyMetrics(t *testing.T) {
	env := Geom_NewEnvelope()
	junit.AssertEquals(t, 0.0, env.GetWidth())
	junit.AssertEquals(t, 0.0, env.GetHeight())
	junit.AssertEquals(t, 0.0, env.GetDiameter())
}

func TestEnvelopeCompareTo(t *testing.T) {
	checkCompareTo(t, 0, Geom_NewEnvelope(), Geom_NewEnvelope())
	checkCompareTo(t, 0, Geom_NewEnvelopeFromXY(1, 2, 1, 2), Geom_NewEnvelopeFromXY(1, 2, 1, 2))
	checkCompareTo(t, 1, Geom_NewEnvelopeFromXY(2, 3, 1, 2), Geom_NewEnvelopeFromXY(1, 2, 1, 2))
	checkCompareTo(t, -1, Geom_NewEnvelopeFromXY(1, 2, 1, 2), Geom_NewEnvelopeFromXY(2, 3, 1, 2))
	checkCompareTo(t, 1, Geom_NewEnvelopeFromXY(1, 2, 1, 3), Geom_NewEnvelopeFromXY(1, 2, 1, 2))
	checkCompareTo(t, 1, Geom_NewEnvelopeFromXY(2, 3, 1, 3), Geom_NewEnvelopeFromXY(1, 3, 1, 2))
}

func checkCompareTo(t *testing.T, expected int, env1, env2 *Geom_Envelope) {
	junit.AssertTrue(t, expected == env1.CompareTo(env2))
	junit.AssertTrue(t, -expected == env2.CompareTo(env1))
}

func checkIntersectsPermuted(t *testing.T, a1x, a1y, a2x, a2y, b1x, b1y, b2x, b2y float64, expected bool) {
	checkIntersects(t, a1x, a1y, a2x, a2y, b1x, b1y, b2x, b2y, expected)
	checkIntersects(t, a1x, a2y, a2x, a1y, b1x, b1y, b2x, b2y, expected)
	checkIntersects(t, a1x, a1y, a2x, a2y, b1x, b2y, b2x, b1y, expected)
	checkIntersects(t, a1x, a2y, a2x, a1y, b1x, b2y, b2x, b1y, expected)
}

func checkIntersects(t *testing.T, a1x, a1y, a2x, a2y, b1x, b1y, b2x, b2y float64, expected bool) {
	a := Geom_NewEnvelopeFromXY(a1x, a2x, a1y, a2y)
	b := Geom_NewEnvelopeFromXY(b1x, b2x, b1y, b2y)
	junit.AssertEquals(t, expected, a.IntersectsEnvelope(b))
	junit.AssertEquals(t, expected, !a.Disjoint(b))

	a1 := Geom_NewCoordinateWithXY(a1x, a1y)
	a2 := Geom_NewCoordinateWithXY(a2x, a2y)
	b1 := Geom_NewCoordinateWithXY(b1x, b1y)
	b2 := Geom_NewCoordinateWithXY(b2x, b2y)
	junit.AssertEquals(t, expected, Geom_Envelope_IntersectsEnvelopeEnvelope(a1, a2, b1, b2))

	junit.AssertEquals(t, expected, a.IntersectsCoordinates(b1, b2))
}

func TestEnvelopeAsGeometry(t *testing.T) {
	precisionModel := Geom_NewPrecisionModelWithScale(1)
	geometryFactory := Geom_NewGeometryFactoryWithPrecisionModel(precisionModel)

	junit.AssertTrue(t, geometryFactory.CreatePointFromCoordinate(nil).GetEnvelope().IsEmpty())

	g := geometryFactory.CreatePointFromCoordinate(Geom_NewCoordinateWithXY(5, 6)).GetEnvelope()
	junit.AssertTrue(t, !g.IsEmpty())
	junit.AssertTrue(t, g.GetChild() != nil)

	reader := Io_NewWKTReaderWithFactory(geometryFactory)
	l, err := reader.Read("LINESTRING(10 10, 20 20, 30 40)")
	if err != nil {
		junit.Fail(t, "failed to read linestring")
	}
	g2 := l.GetEnvelope()
	junit.AssertTrue(t, !g2.IsEmpty())
	junit.AssertTrue(t, g2.GetChild() != nil)
}

func TestEnvelopeGeometryFactoryCreateEnvelope(t *testing.T) {
	precisionModel := Geom_NewPrecisionModelWithScale(1)
	geometryFactory := Geom_NewGeometryFactoryWithPrecisionModel(precisionModel)
	reader := Io_NewWKTReaderWithFactory(geometryFactory)

	checkExpectedEnvelopeGeometry(t, reader, geometryFactory, "POINT (0 0)", "POINT (0 0)")
	checkExpectedEnvelopeGeometry(t, reader, geometryFactory, "POINT (100 13)", "POINT (100 13)")
	checkExpectedEnvelopeGeometry(t, reader, geometryFactory, "LINESTRING (0 0, 0 10)", "LINESTRING (0 0, 0 10)")
	checkExpectedEnvelopeGeometry(t, reader, geometryFactory, "LINESTRING (0 0, 10 0)", "LINESTRING (0 0, 10 0)")

	poly10 := "POLYGON ((0 10, 10 10, 10 0, 0 0, 0 10))"
	checkExpectedEnvelopeGeometry(t, reader, geometryFactory, poly10, poly10)

	checkExpectedEnvelopeGeometry(t, reader, geometryFactory, "LINESTRING (0 0, 10 10)", poly10)
	checkExpectedEnvelopeGeometry(t, reader, geometryFactory, "POLYGON ((5 10, 10 6, 5 0, 0 6, 5 10))", poly10)
}

func checkExpectedEnvelopeGeometry(t *testing.T, reader *Io_WKTReader, factory *Geom_GeometryFactory, wktInput, wktExpected string) {
	input, err := reader.Read(wktInput)
	if err != nil {
		junit.Fail(t, "failed to read input")
	}
	expected, err := reader.Read(wktExpected)
	if err != nil {
		junit.Fail(t, "failed to read expected")
	}

	env := input.GetEnvelopeInternal()
	actual := factory.ToGeometry(env)
	junit.AssertTrue(t, actual.EqualsNorm(expected))
}
