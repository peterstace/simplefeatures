package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Compile-time interface check.
var _ AlgorithmLocate_PointOnGeometryLocator = (*AlgorithmLocate_SimplePointInAreaLocator)(nil)

// AlgorithmLocate_SimplePointInAreaLocator computes the location of points
// relative to a Polygonal Geometry, using a simple O(n) algorithm.
//
// The algorithm used reports if a point lies in the interior, exterior, or
// exactly on the boundary of the Geometry.
//
// Instance methods are provided to implement the interface
// PointOnGeometryLocator. However, they provide no performance advantage over
// the class methods.
//
// This algorithm is suitable for use in cases where only a few points will be
// tested. If many points will be tested, IndexedPointInAreaLocator may provide
// better performance.

// AlgorithmLocate_SimplePointInAreaLocator_Locate determines the Location of a
// point in an areal Geometry. The return value is one of:
//   - Geom_Location_Interior if the point is in the geometry interior
//   - Geom_Location_Boundary if the point lies exactly on the boundary
//   - Geom_Location_Exterior if the point is outside the geometry
func AlgorithmLocate_SimplePointInAreaLocator_Locate(p *Geom_Coordinate, geom *Geom_Geometry) int {
	if geom.IsEmpty() {
		return Geom_Location_Exterior
	}
	// Do a fast check against the geometry envelope first.
	if !geom.GetEnvelopeInternal().IntersectsCoordinate(p) {
		return Geom_Location_Exterior
	}
	return algorithmLocate_SimplePointInAreaLocator_locateInGeometry(p, geom)
}

// AlgorithmLocate_SimplePointInAreaLocator_IsContained determines whether a
// point is contained in a Geometry, or lies on its boundary. This is a
// convenience method for Location.EXTERIOR != locate(p, geom).
func AlgorithmLocate_SimplePointInAreaLocator_IsContained(p *Geom_Coordinate, geom *Geom_Geometry) bool {
	return Geom_Location_Exterior != AlgorithmLocate_SimplePointInAreaLocator_Locate(p, geom)
}

func algorithmLocate_SimplePointInAreaLocator_locateInGeometry(p *Geom_Coordinate, geom *Geom_Geometry) int {
	if java.InstanceOf[*Geom_Polygon](geom) {
		return AlgorithmLocate_SimplePointInAreaLocator_LocatePointInPolygon(p, java.Cast[*Geom_Polygon](geom))
	}

	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		geomi := Geom_NewGeometryCollectionIterator(geom)
		for geomi.HasNext() {
			g2 := geomi.Next()
			if g2 != geom {
				loc := algorithmLocate_SimplePointInAreaLocator_locateInGeometry(p, g2)
				if loc != Geom_Location_Exterior {
					return loc
				}
			}
		}
	}
	return Geom_Location_Exterior
}

// AlgorithmLocate_SimplePointInAreaLocator_LocatePointInPolygon determines the
// Location of a point in a Polygon. The return value is one of:
//   - Geom_Location_Interior if the point is in the geometry interior
//   - Geom_Location_Boundary if the point lies exactly on the boundary
//   - Geom_Location_Exterior if the point is outside the geometry
//
// This method is provided for backwards compatibility only. Use Locate instead.
func AlgorithmLocate_SimplePointInAreaLocator_LocatePointInPolygon(p *Geom_Coordinate, poly *Geom_Polygon) int {
	if poly.IsEmpty() {
		return Geom_Location_Exterior
	}
	shell := poly.GetExteriorRing()
	shellLoc := algorithmLocate_SimplePointInAreaLocator_locatePointInRing(p, shell)
	if shellLoc != Geom_Location_Interior {
		return shellLoc
	}

	// Now test if the point lies in or on the holes.
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		hole := poly.GetInteriorRingN(i)
		holeLoc := algorithmLocate_SimplePointInAreaLocator_locatePointInRing(p, hole)
		if holeLoc == Geom_Location_Boundary {
			return Geom_Location_Boundary
		}
		if holeLoc == Geom_Location_Interior {
			return Geom_Location_Exterior
		}
		// If in EXTERIOR of this hole keep checking the other ones.
	}
	// If not in any hole must be inside polygon.
	return Geom_Location_Interior
}

// AlgorithmLocate_SimplePointInAreaLocator_ContainsPointInPolygon determines
// whether a point lies in a Polygon. If the point lies on the polygon boundary
// it is considered to be inside.
func AlgorithmLocate_SimplePointInAreaLocator_ContainsPointInPolygon(p *Geom_Coordinate, poly *Geom_Polygon) bool {
	return Geom_Location_Exterior != AlgorithmLocate_SimplePointInAreaLocator_LocatePointInPolygon(p, poly)
}

// locatePointInRing determines whether a point lies in a LinearRing, using
// the ring envelope to short-circuit if possible.
func algorithmLocate_SimplePointInAreaLocator_locatePointInRing(p *Geom_Coordinate, ring *Geom_LinearRing) int {
	// Short-circuit if point is not in ring envelope.
	if !ring.GetEnvelopeInternal().IntersectsCoordinate(p) {
		return Geom_Location_Exterior
	}
	return Algorithm_PointLocation_LocateInRing(p, ring.GetCoordinates())
}

type AlgorithmLocate_SimplePointInAreaLocator struct {
	geom *Geom_Geometry
}

// IsAlgorithmLocate_PointOnGeometryLocator is a marker method for the interface.
func (s *AlgorithmLocate_SimplePointInAreaLocator) IsAlgorithmLocate_PointOnGeometryLocator() {}

// AlgorithmLocate_NewSimplePointInAreaLocator creates an instance of a
// point-in-area locator, using the provided areal geometry.
func AlgorithmLocate_NewSimplePointInAreaLocator(geom *Geom_Geometry) *AlgorithmLocate_SimplePointInAreaLocator {
	return &AlgorithmLocate_SimplePointInAreaLocator{
		geom: geom,
	}
}

// Locate determines the Location of a point in an areal Geometry. The
// return value is one of:
//   - Geom_Location_Interior if the point is in the geometry interior
//   - Geom_Location_Boundary if the point lies exactly on the boundary
//   - Geom_Location_Exterior if the point is outside the geometry
func (s *AlgorithmLocate_SimplePointInAreaLocator) Locate(p *Geom_Coordinate) int {
	return AlgorithmLocate_SimplePointInAreaLocator_Locate(p, s.geom)
}
