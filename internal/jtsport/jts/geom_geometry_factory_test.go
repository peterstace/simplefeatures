package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func gfGetFactory() *jts.Geom_GeometryFactory {
	pm := jts.Geom_NewPrecisionModel()
	return jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(pm, 0)
}

func gfGetReader() *jts.Io_WKTReader {
	return jts.Io_NewWKTReaderWithFactory(gfGetFactory())
}

func TestGeometryFactoryCreateGeometry(t *testing.T) {
	tests := []string{
		"POINT EMPTY",
		"POINT ( 10 20 )",
		"LINESTRING EMPTY",
		"LINESTRING(0 0, 10 10)",
		"MULTILINESTRING ((50 100, 100 200), (100 100, 150 200))",
		"POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))",
		"MULTIPOLYGON (((100 200, 200 200, 200 100, 100 100, 100 200)), ((300 200, 400 200, 400 100, 300 100, 300 200)))",
		"GEOMETRYCOLLECTION (POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200)), LINESTRING (250 100, 350 200), POINT (350 150))",
	}
	factory := gfGetFactory()
	reader := gfGetReader()
	for _, wkt := range tests {
		g, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", wkt, err)
		}
		g2 := factory.CreateGeometry(g)
		if !g.EqualsExact(g2) {
			t.Errorf("createGeometry failed for %q", wkt)
		}
	}
}

func TestGeometryFactoryCreateEmpty(t *testing.T) {
	factory := gfGetFactory()

	checkEmpty := func(geom *jts.Geom_Geometry) {
		t.Helper()
		if !geom.IsEmpty() {
			t.Error("expected empty geometry")
		}
	}

	checkEmpty(factory.CreateEmpty(0))
	checkEmpty(factory.CreateEmpty(1))
	checkEmpty(factory.CreateEmpty(2))

	checkEmpty(factory.CreatePoint().Geom_Geometry)
	checkEmpty(factory.CreateLineString().Geom_Geometry)
	checkEmpty(factory.CreatePolygon().Geom_Geometry)

	checkEmpty(factory.CreateMultiPoint().Geom_Geometry)
	checkEmpty(factory.CreateMultiLineString().Geom_Geometry)
	checkEmpty(factory.CreateMultiPolygon().Geom_Geometry)
	checkEmpty(factory.CreateGeometryCollection().Geom_Geometry)
}

func TestGeometryFactoryCreateEmptyTypes(t *testing.T) {
	factory := gfGetFactory()

	// Check that CreateEmpty returns correct types.
	e0 := factory.CreateEmpty(0)
	if _, ok := java.GetLeaf(e0).(*jts.Geom_Point); !ok {
		t.Error("expected Point for dimension 0")
	}

	e1 := factory.CreateEmpty(1)
	if _, ok := java.GetLeaf(e1).(*jts.Geom_LineString); !ok {
		t.Error("expected LineString for dimension 1")
	}

	e2 := factory.CreateEmpty(2)
	if _, ok := java.GetLeaf(e2).(*jts.Geom_Polygon); !ok {
		t.Error("expected Polygon for dimension 2")
	}

	// Check that Create* methods return correct types.
	if !java.InstanceOf[*jts.Geom_Point](factory.CreatePoint()) {
		t.Error("expected Point from CreatePoint")
	}
	if !java.InstanceOf[*jts.Geom_LineString](factory.CreateLineString()) {
		t.Error("expected LineString from CreateLineString")
	}
	if !java.InstanceOf[*jts.Geom_Polygon](factory.CreatePolygon()) {
		t.Error("expected Polygon from CreatePolygon")
	}
	if !java.InstanceOf[*jts.Geom_MultiPoint](factory.CreateMultiPoint()) {
		t.Error("expected MultiPoint from CreateMultiPoint")
	}
	if !java.InstanceOf[*jts.Geom_MultiLineString](factory.CreateMultiLineString()) {
		t.Error("expected MultiLineString from CreateMultiLineString")
	}
	if !java.InstanceOf[*jts.Geom_MultiPolygon](factory.CreateMultiPolygon()) {
		t.Error("expected MultiPolygon from CreateMultiPolygon")
	}
	if !java.InstanceOf[*jts.Geom_GeometryCollection](factory.CreateGeometryCollection()) {
		t.Error("expected GeometryCollection from CreateGeometryCollection")
	}
}

func TestGeometryFactoryDeepCopy(t *testing.T) {
	reader := gfGetReader()
	factory := gfGetFactory()

	g, err := reader.Read("POINT ( 10 10 )")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	pt := java.Cast[*jts.Geom_Point](g)
	g2 := factory.CreateGeometry(g)
	pt.GetCoordinateSequence().SetOrdinate(0, 0, 99)
	if g.EqualsExact(g2) {
		t.Error("expected deep copy - geometries should not be equal after modification")
	}
}

func TestGeometryFactoryMultiPointCS(t *testing.T) {
	// Use CoordinateArraySequenceFactory (PackedCoordinateSequenceFactory is not ported).
	gf := jts.Geom_NewGeometryFactoryWithCoordinateSequenceFactory(
		jts.GeomImpl_CoordinateArraySequenceFactory_Instance(),
	)
	// Create a 4D (XYZM) coordinate sequence with 1 point.
	mpSeq := gf.GetCoordinateSequenceFactory().CreateWithSizeAndDimensionAndMeasures(1, 4, 1)
	mpSeq.SetOrdinate(0, 0, 50)
	mpSeq.SetOrdinate(0, 1, -2)
	mpSeq.SetOrdinate(0, 2, 10)
	mpSeq.SetOrdinate(0, 3, 20)

	mp := gf.CreateMultiPointFromCoordinateSequence(mpSeq)
	pt := java.Cast[*jts.Geom_Point](mp.GetGeometryN(0))
	pSeq := pt.GetCoordinateSequence()

	if pSeq.GetDimension() != 4 {
		t.Errorf("expected dimension 4, got %d", pSeq.GetDimension())
	}
	for i := 0; i < 4; i++ {
		if mpSeq.GetOrdinate(0, i) != pSeq.GetOrdinate(0, i) {
			t.Errorf("ordinate %d: expected %v, got %v", i, mpSeq.GetOrdinate(0, i), pSeq.GetOrdinate(0, i))
		}
	}
}

func TestGeometryFactoryCopyGeometryWithNonDefaultDimension(t *testing.T) {
	gf := jts.Geom_NewGeometryFactoryWithCoordinateSequenceFactory(
		jts.GeomImpl_CoordinateArraySequenceFactory_Instance(),
	)
	mpSeq := gf.GetCoordinateSequenceFactory().CreateWithSizeAndDimension(1, 2)
	mpSeq.SetOrdinate(0, 0, 50)
	mpSeq.SetOrdinate(0, 1, -2)

	g := gf.CreatePointFromCoordinateSequence(mpSeq)
	pSeq := g.GetCoordinateSequence()
	if pSeq.GetDimension() != 2 {
		t.Errorf("expected dimension 2, got %d", pSeq.GetDimension())
	}

	defaultFactory := gfGetFactory()
	g2 := java.Cast[*jts.Geom_Point](defaultFactory.CreateGeometry(g.Geom_Geometry))
	if g2.GetCoordinateSequence().GetDimension() != 2 {
		t.Errorf("expected copied geometry to have dimension 2, got %d", g2.GetCoordinateSequence().GetDimension())
	}
}
