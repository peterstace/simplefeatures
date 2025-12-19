package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestPointLocator_Box(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// POLYGON ((0 0, 0 20, 20 20, 20 0, 0 0))
	shell := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 20),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(20, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	})
	poly := factory.CreatePolygonFromLinearRing(shell)

	pl := jts.Algorithm_NewPointLocator()
	loc := pl.Locate(jts.Geom_NewCoordinateWithXY(10, 10), poly.Geom_Geometry)
	if loc != jts.Geom_Location_Interior {
		t.Errorf("expected Interior, got %d", loc)
	}
}

func TestPointLocator_ComplexRing(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// POLYGON ((-40 80, -40 -80, 20 0, 20 -100, 40 40, 80 -80, 100 80, 140 -20, 120 140, 40 180, 60 40, 0 120, -20 -20, -40 80))
	shell := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(-40, 80),
		jts.Geom_NewCoordinateWithXY(-40, -80),
		jts.Geom_NewCoordinateWithXY(20, 0),
		jts.Geom_NewCoordinateWithXY(20, -100),
		jts.Geom_NewCoordinateWithXY(40, 40),
		jts.Geom_NewCoordinateWithXY(80, -80),
		jts.Geom_NewCoordinateWithXY(100, 80),
		jts.Geom_NewCoordinateWithXY(140, -20),
		jts.Geom_NewCoordinateWithXY(120, 140),
		jts.Geom_NewCoordinateWithXY(40, 180),
		jts.Geom_NewCoordinateWithXY(60, 40),
		jts.Geom_NewCoordinateWithXY(0, 120),
		jts.Geom_NewCoordinateWithXY(-20, -20),
		jts.Geom_NewCoordinateWithXY(-40, 80),
	})
	poly := factory.CreatePolygonFromLinearRing(shell)

	pl := jts.Algorithm_NewPointLocator()
	loc := pl.Locate(jts.Geom_NewCoordinateWithXY(0, 0), poly.Geom_Geometry)
	if loc != jts.Geom_Location_Interior {
		t.Errorf("expected Interior, got %d", loc)
	}
}

func TestPointLocator_LinearRingLineString(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// GEOMETRYCOLLECTION( LINESTRING(0 0, 10 10), LINEARRING(10 10, 10 20, 20 10, 10 10))
	ls := factory.CreateLineStringFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
	})
	ring := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(10, 20),
		jts.Geom_NewCoordinateWithXY(20, 10),
		jts.Geom_NewCoordinateWithXY(10, 10),
	})
	gc := factory.CreateGeometryCollectionFromGeometries([]*jts.Geom_Geometry{ls.Geom_Geometry, ring.Geom_Geometry})

	// Point (0, 0) is an endpoint of the linestring, so it's on the boundary.
	pl := jts.Algorithm_NewPointLocator()
	loc := pl.Locate(jts.Geom_NewCoordinateWithXY(0, 0), gc.Geom_Geometry)
	if loc != jts.Geom_Location_Boundary {
		t.Errorf("expected Boundary, got %d", loc)
	}
}

func TestPointLocator_PointInsideLinearRing(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// LINEARRING(10 10, 10 20, 20 10, 10 10)
	ring := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(10, 20),
		jts.Geom_NewCoordinateWithXY(20, 10),
		jts.Geom_NewCoordinateWithXY(10, 10),
	})

	// Points inside a LinearRing are EXTERIOR (rings don't enclose area).
	pl := jts.Algorithm_NewPointLocator()
	loc := pl.Locate(jts.Geom_NewCoordinateWithXY(11, 11), ring.Geom_Geometry)
	if loc != jts.Geom_Location_Exterior {
		t.Errorf("expected Exterior, got %d", loc)
	}
}

func TestPointLocator_Polygon(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// POLYGON ((70 340, 430 50, 70 50, 70 340))
	shell := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(70, 340),
		jts.Geom_NewCoordinateWithXY(430, 50),
		jts.Geom_NewCoordinateWithXY(70, 50),
		jts.Geom_NewCoordinateWithXY(70, 340),
	})
	poly := factory.CreatePolygonFromLinearRing(shell)

	pl := jts.Algorithm_NewPointLocator()

	tests := []struct {
		name     string
		pt       *jts.Geom_Coordinate
		expected int
	}{
		{"exterior", jts.Geom_NewCoordinateWithXY(420, 340), jts.Geom_Location_Exterior},
		{"boundary 1", jts.Geom_NewCoordinateWithXY(350, 50), jts.Geom_Location_Boundary},
		{"boundary 2", jts.Geom_NewCoordinateWithXY(410, 50), jts.Geom_Location_Boundary},
		{"interior", jts.Geom_NewCoordinateWithXY(190, 150), jts.Geom_Location_Interior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := pl.Locate(tt.pt, poly.Geom_Geometry)
			if loc != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, loc)
			}
		})
	}
}

func TestPointLocator_RingBoundaryNodeRule(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// LINEARRING(10 10, 10 20, 20 10, 10 10)
	ring := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(10, 20),
		jts.Geom_NewCoordinateWithXY(20, 10),
		jts.Geom_NewCoordinateWithXY(10, 10),
	})

	pt := jts.Geom_NewCoordinateWithXY(10, 10)

	tests := []struct {
		name     string
		rule     jts.Algorithm_BoundaryNodeRule
		expected int
	}{
		// Mod2 rule: closed ring has 2 boundary touches at endpoint -> 2 % 2 = 0, so interior.
		{"Mod2", jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, jts.Geom_Location_Interior},
		// Endpoint rule: any endpoint is boundary.
		{"Endpoint", jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, jts.Geom_Location_Boundary},
		// Monovalent rule: only valency 1 is boundary.
		{"Monovalent", jts.Algorithm_BoundaryNodeRule_MONOVALENT_ENDPOINT_BOUNDARY_RULE, jts.Geom_Location_Interior},
		// Multivalent rule: valency > 1 is boundary.
		{"Multivalent", jts.Algorithm_BoundaryNodeRule_MULTIVALENT_ENDPOINT_BOUNDARY_RULE, jts.Geom_Location_Boundary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := jts.Algorithm_NewPointLocatorWithBoundaryRule(tt.rule)
			loc := pl.Locate(pt, ring.Geom_Geometry)
			if loc != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, loc)
			}
		})
	}
}

func TestPointLocator_Intersects(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// POLYGON ((0 0, 0 20, 20 20, 20 0, 0 0))
	shell := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 20),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(20, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	})
	poly := factory.CreatePolygonFromLinearRing(shell)

	pl := jts.Algorithm_NewPointLocator()

	// Interior point - should intersect.
	if !pl.Intersects(jts.Geom_NewCoordinateWithXY(10, 10), poly.Geom_Geometry) {
		t.Error("expected interior point to intersect")
	}

	// Boundary point - should intersect.
	if !pl.Intersects(jts.Geom_NewCoordinateWithXY(0, 10), poly.Geom_Geometry) {
		t.Error("expected boundary point to intersect")
	}

	// Exterior point - should not intersect.
	if pl.Intersects(jts.Geom_NewCoordinateWithXY(30, 10), poly.Geom_Geometry) {
		t.Error("expected exterior point not to intersect")
	}
}

func TestPointLocator_EmptyGeometry(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	emptyPoly := factory.CreatePolygon()

	pl := jts.Algorithm_NewPointLocator()
	loc := pl.Locate(jts.Geom_NewCoordinateWithXY(0, 0), emptyPoly.Geom_Geometry)
	if loc != jts.Geom_Location_Exterior {
		t.Errorf("expected Exterior for empty geometry, got %d", loc)
	}
}

func TestPointLocator_Point(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	pt := factory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(5, 5))

	pl := jts.Algorithm_NewPointLocator()

	// Same point.
	loc := pl.Locate(jts.Geom_NewCoordinateWithXY(5, 5), pt.Geom_Geometry)
	if loc != jts.Geom_Location_Interior {
		t.Errorf("expected Interior for same point, got %d", loc)
	}

	// Different point.
	loc = pl.Locate(jts.Geom_NewCoordinateWithXY(10, 10), pt.Geom_Geometry)
	if loc != jts.Geom_Location_Exterior {
		t.Errorf("expected Exterior for different point, got %d", loc)
	}
}

func TestPointLocator_LineString(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// LINESTRING (0 0, 10 10, 20 0)
	ls := factory.CreateLineStringFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(20, 0),
	})

	pl := jts.Algorithm_NewPointLocator()

	tests := []struct {
		name     string
		pt       *jts.Geom_Coordinate
		expected int
	}{
		{"start point", jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_Location_Boundary},
		{"end point", jts.Geom_NewCoordinateWithXY(20, 0), jts.Geom_Location_Boundary},
		{"middle vertex", jts.Geom_NewCoordinateWithXY(10, 10), jts.Geom_Location_Interior},
		{"on segment", jts.Geom_NewCoordinateWithXY(5, 5), jts.Geom_Location_Interior},
		{"exterior", jts.Geom_NewCoordinateWithXY(100, 100), jts.Geom_Location_Exterior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := pl.Locate(tt.pt, ls.Geom_Geometry)
			if loc != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, loc)
			}
		})
	}
}

func TestPointLocator_ClosedLineString(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// Closed LINESTRING (forms a triangle).
	ls := factory.CreateLineStringFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(20, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	})

	pl := jts.Algorithm_NewPointLocator()

	// For a closed line with Mod2 rule, the endpoint touches boundary twice,
	// so 2 % 2 = 0 means it's interior, not boundary.
	loc := pl.Locate(jts.Geom_NewCoordinateWithXY(0, 0), ls.Geom_Geometry)
	if loc != jts.Geom_Location_Interior {
		t.Errorf("expected Interior for closed linestring endpoint with Mod2 rule, got %d", loc)
	}
}

func TestPointLocator_PolygonWithHole(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()

	// Outer shell: 0,0 to 30,30.
	shell := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 30),
		jts.Geom_NewCoordinateWithXY(30, 30),
		jts.Geom_NewCoordinateWithXY(30, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	})

	// Inner hole: 10,10 to 20,20.
	hole := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(10, 20),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(20, 10),
		jts.Geom_NewCoordinateWithXY(10, 10),
	})

	poly := factory.CreatePolygonWithLinearRingAndHoles(shell, []*jts.Geom_LinearRing{hole})

	pl := jts.Algorithm_NewPointLocator()

	tests := []struct {
		name     string
		pt       *jts.Geom_Coordinate
		expected int
	}{
		{"in shell, outside hole", jts.Geom_NewCoordinateWithXY(5, 5), jts.Geom_Location_Interior},
		{"in hole", jts.Geom_NewCoordinateWithXY(15, 15), jts.Geom_Location_Exterior},
		{"on hole boundary", jts.Geom_NewCoordinateWithXY(10, 15), jts.Geom_Location_Boundary},
		{"on shell boundary", jts.Geom_NewCoordinateWithXY(0, 15), jts.Geom_Location_Boundary},
		{"outside shell", jts.Geom_NewCoordinateWithXY(50, 50), jts.Geom_Location_Exterior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := pl.Locate(tt.pt, poly.Geom_Geometry)
			if loc != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, loc)
			}
		})
	}
}
