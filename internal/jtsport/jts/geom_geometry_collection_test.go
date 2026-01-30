package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestGeometryCollectionGetDimension(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// Empty collection has dimension FALSE (-1).
	emptyGC := factory.CreateGeometryCollection()
	if got := emptyGC.GetDimension(); got != jts.Geom_Dimension_False {
		t.Errorf("empty GeometryCollection.GetDimension() = %d, want %d", got, jts.Geom_Dimension_False)
	}

	// Collection with just points has dimension 0.
	pt1 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pt2 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(30, 30))
	pointsGC := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt1.Geom_Geometry, pt2.Geom_Geometry}, factory)
	if got := pointsGC.GetDimension(); got != 0 {
		t.Errorf("points-only GeometryCollection.GetDimension() = %d, want %d", got, 0)
	}

	// Collection with points and linestring has dimension 1.
	coords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(15, 15),
		jts.Geom_NewCoordinateWithXY(20, 20),
	}
	ls := factory.CreateLineStringFromCoordinates(coords)
	mixedGC := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt1.Geom_Geometry, pt2.Geom_Geometry, ls.Geom_Geometry}, factory)
	if got := mixedGC.GetDimension(); got != 1 {
		t.Errorf("points+line GeometryCollection.GetDimension() = %d, want %d", got, 1)
	}

	// Collection with polygon has dimension 2.
	shellCoords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(0, 10),
		jts.Geom_NewCoordinateWithXY(0, 0),
	}
	shell := factory.CreateLinearRingFromCoordinates(shellCoords)
	poly := factory.CreatePolygonFromLinearRing(shell)
	polyGC := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt1.Geom_Geometry, ls.Geom_Geometry, poly.Geom_Geometry}, factory)
	if got := polyGC.GetDimension(); got != 2 {
		t.Errorf("points+line+polygon GeometryCollection.GetDimension() = %d, want %d", got, 2)
	}
}

func TestGeometryCollectionGetCoordinates(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	pt1 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pt2 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(30, 30))
	coords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(15, 15),
		jts.Geom_NewCoordinateWithXY(20, 20),
	}
	ls := factory.CreateLineStringFromCoordinates(coords)
	gc := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt1.Geom_Geometry, pt2.Geom_Geometry, ls.Geom_Geometry}, factory)

	if got := gc.GetNumPoints(); got != 4 {
		t.Errorf("GeometryCollection.GetNumPoints() = %d, want %d", got, 4)
	}

	coordinates := gc.GetCoordinates()
	if len(coordinates) != 4 {
		t.Fatalf("len(GetCoordinates()) = %d, want %d", len(coordinates), 4)
	}

	expected0 := jts.Geom_NewCoordinateWithXY(10, 10)
	if !coordinates[0].Equals2D(expected0) {
		t.Errorf("coordinates[0] = (%v, %v), want (%v, %v)", coordinates[0].GetX(), coordinates[0].GetY(), expected0.GetX(), expected0.GetY())
	}

	expected3 := jts.Geom_NewCoordinateWithXY(20, 20)
	if !coordinates[3].Equals2D(expected3) {
		t.Errorf("coordinates[3] = (%v, %v), want (%v, %v)", coordinates[3].GetX(), coordinates[3].GetY(), expected3.GetX(), expected3.GetY())
	}
}

func TestGeometryCollectionIsEmpty(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// Empty collection is empty.
	emptyGC := factory.CreateGeometryCollection()
	if !emptyGC.IsEmpty() {
		t.Error("empty GeometryCollection.IsEmpty() = false, want true")
	}

	// Collection with empty point is still empty.
	emptyPt := factory.CreatePoint()
	gcWithEmptyPt := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{emptyPt.Geom_Geometry}, factory)
	if !gcWithEmptyPt.IsEmpty() {
		t.Error("GeometryCollection with empty point IsEmpty() = false, want true")
	}

	// Collection with non-empty point is not empty.
	pt := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	gcWithPt := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt.Geom_Geometry}, factory)
	if gcWithPt.IsEmpty() {
		t.Error("GeometryCollection with point IsEmpty() = true, want false")
	}
}

func TestGeometryCollectionGetGeometryType(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	gc := factory.CreateGeometryCollection()
	if got := gc.GetGeometryType(); got != "GeometryCollection" {
		t.Errorf("GeometryCollection.GetGeometryType() = %q, want %q", got, "GeometryCollection")
	}
}

func TestGeometryCollectionGetNumGeometries(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	emptyGC := factory.CreateGeometryCollection()
	if got := emptyGC.GetNumGeometries(); got != 0 {
		t.Errorf("empty GeometryCollection.GetNumGeometries() = %d, want %d", got, 0)
	}

	pt1 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pt2 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(20, 20))
	gc := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt1.Geom_Geometry, pt2.Geom_Geometry}, factory)
	if got := gc.GetNumGeometries(); got != 2 {
		t.Errorf("GeometryCollection.GetNumGeometries() = %d, want %d", got, 2)
	}
}

func TestGeometryCollectionGetGeometryN(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	pt1 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pt2 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(20, 20))
	gc := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt1.Geom_Geometry, pt2.Geom_Geometry}, factory)

	geom0 := gc.GetGeometryN(0)
	if geom0 != pt1.Geom_Geometry {
		t.Error("GetGeometryN(0) did not return first geometry")
	}

	geom1 := gc.GetGeometryN(1)
	if geom1 != pt2.Geom_Geometry {
		t.Error("GetGeometryN(1) did not return second geometry")
	}
}

func TestGeometryCollectionCopy(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	pt1 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pt2 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(20, 20))
	gc := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt1.Geom_Geometry, pt2.Geom_Geometry}, factory)

	gcCopy := gc.Copy()
	if gcCopy == gc.Geom_Geometry {
		t.Error("Copy() returned same instance")
	}

	gcCopyCast := java.Cast[*jts.Geom_GeometryCollection](gcCopy)
	if gcCopyCast.GetNumGeometries() != 2 {
		t.Errorf("Copy().GetNumGeometries() = %d, want %d", gcCopyCast.GetNumGeometries(), 2)
	}

	if !gc.Geom_Geometry.EqualsExact(gcCopy) {
		t.Error("Copy() is not equal to original")
	}
}

func TestGeometryCollectionEqualsExact(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	pt1 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pt2 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(20, 20))
	gc1 := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt1.Geom_Geometry, pt2.Geom_Geometry}, factory)

	pt3 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pt4 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(20, 20))
	gc2 := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt3.Geom_Geometry, pt4.Geom_Geometry}, factory)

	if !gc1.Geom_Geometry.EqualsExact(gc2.Geom_Geometry) {
		t.Error("identical GeometryCollections are not EqualsExact")
	}

	pt5 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pt6 := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(30, 30))
	gc3 := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt5.Geom_Geometry, pt6.Geom_Geometry}, factory)

	if gc1.Geom_Geometry.EqualsExact(gc3.Geom_Geometry) {
		t.Error("different GeometryCollections are EqualsExact")
	}
}

func TestGeometryCollectionHasDimension(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// Point-only collection.
	pt := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	pointGC := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt.Geom_Geometry}, factory)
	if !pointGC.HasDimension(0) {
		t.Error("point collection HasDimension(0) = false, want true")
	}
	if pointGC.HasDimension(1) {
		t.Error("point collection HasDimension(1) = true, want false")
	}
	if pointGC.HasDimension(2) {
		t.Error("point collection HasDimension(2) = true, want false")
	}

	// Line-only collection.
	coords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
	}
	ls := factory.CreateLineStringFromCoordinates(coords)
	lineGC := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{ls.Geom_Geometry}, factory)
	if lineGC.HasDimension(0) {
		t.Error("line collection HasDimension(0) = true, want false")
	}
	if !lineGC.HasDimension(1) {
		t.Error("line collection HasDimension(1) = false, want true")
	}
	if lineGC.HasDimension(2) {
		t.Error("line collection HasDimension(2) = true, want false")
	}

	// Polygon-only collection.
	shellCoords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(0, 10),
		jts.Geom_NewCoordinateWithXY(0, 0),
	}
	shell := factory.CreateLinearRingFromCoordinates(shellCoords)
	poly := factory.CreatePolygonFromLinearRing(shell)
	polyGC := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{poly.Geom_Geometry}, factory)
	if polyGC.HasDimension(0) {
		t.Error("polygon collection HasDimension(0) = true, want false")
	}
	if polyGC.HasDimension(1) {
		t.Error("polygon collection HasDimension(1) = true, want false")
	}
	if !polyGC.HasDimension(2) {
		t.Error("polygon collection HasDimension(2) = false, want true")
	}

	// Mixed collection.
	mixedGC := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt.Geom_Geometry, ls.Geom_Geometry, poly.Geom_Geometry}, factory)
	if !mixedGC.HasDimension(0) {
		t.Error("mixed collection HasDimension(0) = false, want true")
	}
	if !mixedGC.HasDimension(1) {
		t.Error("mixed collection HasDimension(1) = false, want true")
	}
	if !mixedGC.HasDimension(2) {
		t.Error("mixed collection HasDimension(2) = false, want true")
	}
}

func TestGeometryCollectionGetArea(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// Polygon with area 100.
	shellCoords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(0, 10),
		jts.Geom_NewCoordinateWithXY(0, 0),
	}
	shell := factory.CreateLinearRingFromCoordinates(shellCoords)
	poly := factory.CreatePolygonFromLinearRing(shell)
	gc := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{poly.Geom_Geometry}, factory)

	if got := gc.GetArea(); got != 100.0 {
		t.Errorf("GeometryCollection.GetArea() = %v, want %v", got, 100.0)
	}
}

// TestGeometryCollectionGetLength is skipped because LineString.GetLength
// depends on algorithm/Length which is not yet ported.

func TestGeometryCollectionReverse(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	coords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 0),
		jts.Geom_NewCoordinateWithXY(20, 0),
	}
	ls := factory.CreateLineStringFromCoordinates(coords)
	gc := jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{ls.Geom_Geometry}, factory)

	reversed := gc.Reverse()
	reversedGC := java.Cast[*jts.Geom_GeometryCollection](reversed)
	reversedLS := java.Cast[*jts.Geom_LineString](reversedGC.GetGeometryN(0))

	reversedCoords := reversedLS.GetCoordinates()
	if len(reversedCoords) != 3 {
		t.Fatalf("reversed linestring has %d coords, want 3", len(reversedCoords))
	}
	if reversedCoords[0].GetX() != 20.0 || reversedCoords[0].GetY() != 0.0 {
		t.Errorf("reversed coords[0] = (%v, %v), want (20, 0)", reversedCoords[0].GetX(), reversedCoords[0].GetY())
	}
	if reversedCoords[2].GetX() != 0.0 || reversedCoords[2].GetY() != 0.0 {
		t.Errorf("reversed coords[2] = (%v, %v), want (0, 0)", reversedCoords[2].GetX(), reversedCoords[2].GetY())
	}
}

func TestGeometryCollectionNilElementsPanics(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	pt := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil element, got none")
		}
	}()

	jts.Geom_NewGeometryCollection([]*jts.Geom_Geometry{pt.Geom_Geometry, nil}, factory)
}
