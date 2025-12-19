package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestGeometryCollectionIteratorGeometryCollection(t *testing.T) {
	// Build: GEOMETRYCOLLECTION (GEOMETRYCOLLECTION (POINT (10 10)))
	factory := jts.Geom_NewGeometryFactoryDefault()

	point := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	innerGC := factory.CreateGeometryCollectionFromGeometries([]*jts.Geom_Geometry{point.Geom_Geometry})
	outerGC := factory.CreateGeometryCollectionFromGeometries([]*jts.Geom_Geometry{innerGC.Geom_Geometry})

	it := jts.Geom_NewGeometryCollectionIterator(outerGC.Geom_Geometry)

	// First element should be the outer GeometryCollection.
	if !it.HasNext() {
		t.Fatal("expected HasNext() to be true")
	}
	elem := it.Next()
	if !java.InstanceOf[*jts.Geom_GeometryCollection](elem) {
		t.Errorf("expected GeometryCollection, got %T", java.GetLeaf(elem))
	}

	// Second element should be the inner GeometryCollection.
	if !it.HasNext() {
		t.Fatal("expected HasNext() to be true")
	}
	elem = it.Next()
	if !java.InstanceOf[*jts.Geom_GeometryCollection](elem) {
		t.Errorf("expected GeometryCollection, got %T", java.GetLeaf(elem))
	}

	// Third element should be the Point.
	if !it.HasNext() {
		t.Fatal("expected HasNext() to be true")
	}
	elem = it.Next()
	if !java.InstanceOf[*jts.Geom_Point](elem) {
		t.Errorf("expected Point, got %T", java.GetLeaf(elem))
	}

	// No more elements.
	if it.HasNext() {
		t.Error("expected HasNext() to be false")
	}
}

func TestGeometryCollectionIteratorAtomic(t *testing.T) {
	// Build: POLYGON ((1 9, 9 9, 9 1, 1 1, 1 9))
	factory := jts.Geom_NewGeometryFactoryDefault()

	coords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(1, 9),
		jts.Geom_NewCoordinateWithXY(9, 9),
		jts.Geom_NewCoordinateWithXY(9, 1),
		jts.Geom_NewCoordinateWithXY(1, 1),
		jts.Geom_NewCoordinateWithXY(1, 9),
	}
	polygon := factory.CreatePolygonFromCoordinates(coords)

	it := jts.Geom_NewGeometryCollectionIterator(polygon.Geom_Geometry)

	// First element should be the Polygon itself.
	if !it.HasNext() {
		t.Fatal("expected HasNext() to be true")
	}
	elem := it.Next()
	if !java.InstanceOf[*jts.Geom_Polygon](elem) {
		t.Errorf("expected Polygon, got %T", java.GetLeaf(elem))
	}

	// No more elements.
	if it.HasNext() {
		t.Error("expected HasNext() to be false")
	}
}
