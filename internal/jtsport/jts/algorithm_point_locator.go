package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Algorithm_PointLocator computes the topological (Location) of a single point
// to a Geometry. A BoundaryNodeRule may be specified to control the evaluation
// of whether the point lies on the boundary or not. The default rule is to use
// the SFS Boundary Determination Rule.
//
// Notes:
//   - LinearRings do not enclose any area - points inside the ring are still in
//     the EXTERIOR of the ring.
//
// Instances of this class are not reentrant.
type Algorithm_PointLocator struct {
	// Default is to use OGC SFS rule.
	boundaryRule Algorithm_BoundaryNodeRule
	// True if the point lies in or on any Geometry element.
	isIn bool
	// The number of sub-elements whose boundaries the point lies in.
	numBoundaries int
}

// Algorithm_NewPointLocator creates a new PointLocator using the OGC SFS boundary
// rule.
func Algorithm_NewPointLocator() *Algorithm_PointLocator {
	return &Algorithm_PointLocator{
		boundaryRule: Algorithm_BoundaryNodeRule_OGC_SFS_BOUNDARY_RULE,
	}
}

// Algorithm_NewPointLocatorWithBoundaryRule creates a new PointLocator using the
// specified boundary rule.
func Algorithm_NewPointLocatorWithBoundaryRule(boundaryRule Algorithm_BoundaryNodeRule) *Algorithm_PointLocator {
	if boundaryRule == nil {
		panic("Rule must be non-null")
	}
	return &Algorithm_PointLocator{
		boundaryRule: boundaryRule,
	}
}

// Intersects is a convenience method to test a point for intersection with a
// Geometry.
func (pl *Algorithm_PointLocator) Intersects(p *Geom_Coordinate, geom *Geom_Geometry) bool {
	return pl.Locate(p, geom) != Geom_Location_Exterior
}

// Locate computes the topological relationship (Location) of a single point to
// a Geometry. It handles both single-element and multi-element Geometries. The
// algorithm for multi-part Geometries takes into account the SFS Boundary
// Determination Rule.
func (pl *Algorithm_PointLocator) Locate(p *Geom_Coordinate, geom *Geom_Geometry) int {
	if geom.IsEmpty() {
		return Geom_Location_Exterior
	}

	if java.InstanceOf[*Geom_LineString](geom) {
		return pl.locateOnLineString(p, java.Cast[*Geom_LineString](geom))
	} else if java.InstanceOf[*Geom_Polygon](geom) {
		return pl.locateInPolygon(p, java.Cast[*Geom_Polygon](geom))
	}

	pl.isIn = false
	pl.numBoundaries = 0
	pl.computeLocation(p, geom)
	if pl.boundaryRule.IsInBoundary(pl.numBoundaries) {
		return Geom_Location_Boundary
	}
	if pl.numBoundaries > 0 || pl.isIn {
		return Geom_Location_Interior
	}
	return Geom_Location_Exterior
}

func (pl *Algorithm_PointLocator) computeLocation(p *Geom_Coordinate, geom *Geom_Geometry) {
	if geom.IsEmpty() {
		return
	}

	if java.InstanceOf[*Geom_Point](geom) {
		pl.updateLocationInfo(pl.locateOnPoint(p, java.Cast[*Geom_Point](geom)))
	}
	if java.InstanceOf[*Geom_LineString](geom) {
		pl.updateLocationInfo(pl.locateOnLineString(p, java.Cast[*Geom_LineString](geom)))
	} else if java.InstanceOf[*Geom_Polygon](geom) {
		pl.updateLocationInfo(pl.locateInPolygon(p, java.Cast[*Geom_Polygon](geom)))
	} else if java.InstanceOf[*Geom_MultiLineString](geom) {
		mls := java.Cast[*Geom_MultiLineString](geom)
		for i := 0; i < mls.GetNumGeometries(); i++ {
			l := java.Cast[*Geom_LineString](mls.GetGeometryN(i))
			pl.updateLocationInfo(pl.locateOnLineString(p, l))
		}
	} else if java.InstanceOf[*Geom_MultiPolygon](geom) {
		mpoly := java.Cast[*Geom_MultiPolygon](geom)
		for i := 0; i < mpoly.GetNumGeometries(); i++ {
			poly := java.Cast[*Geom_Polygon](mpoly.GetGeometryN(i))
			pl.updateLocationInfo(pl.locateInPolygon(p, poly))
		}
	} else if java.InstanceOf[*Geom_GeometryCollection](geom) {
		geomi := Geom_NewGeometryCollectionIterator(geom)
		for geomi.HasNext() {
			g2 := geomi.Next()
			if g2 != geom {
				pl.computeLocation(p, g2)
			}
		}
	}
}

func (pl *Algorithm_PointLocator) updateLocationInfo(loc int) {
	if loc == Geom_Location_Interior {
		pl.isIn = true
	}
	if loc == Geom_Location_Boundary {
		pl.numBoundaries++
	}
}

func (pl *Algorithm_PointLocator) locateOnPoint(p *Geom_Coordinate, pt *Geom_Point) int {
	// No point in doing envelope test, since equality test is just as fast.
	ptCoord := pt.GetCoordinate()
	if ptCoord.Equals2D(p) {
		return Geom_Location_Interior
	}
	return Geom_Location_Exterior
}

func (pl *Algorithm_PointLocator) locateOnLineString(p *Geom_Coordinate, l *Geom_LineString) int {
	// Bounding-box check.
	if !l.GetEnvelopeInternal().IntersectsCoordinate(p) {
		return Geom_Location_Exterior
	}

	seq := l.GetCoordinateSequence()
	if p.Equals(seq.GetCoordinate(0)) || p.Equals(seq.GetCoordinate(seq.Size()-1)) {
		boundaryCount := 1
		if l.IsClosed() {
			boundaryCount = 2
		}
		if pl.boundaryRule.IsInBoundary(boundaryCount) {
			return Geom_Location_Boundary
		}
		return Geom_Location_Interior
	}
	if Algorithm_PointLocation_IsOnLineSeq(p, seq) {
		return Geom_Location_Interior
	}
	return Geom_Location_Exterior
}

func (pl *Algorithm_PointLocator) locateInPolygonRing(p *Geom_Coordinate, ring *Geom_LinearRing) int {
	// Bounding-box check.
	if !ring.GetEnvelopeInternal().IntersectsCoordinate(p) {
		return Geom_Location_Exterior
	}
	return Algorithm_PointLocation_LocateInRing(p, ring.GetCoordinates())
}

func (pl *Algorithm_PointLocator) locateInPolygon(p *Geom_Coordinate, poly *Geom_Polygon) int {
	if poly.IsEmpty() {
		return Geom_Location_Exterior
	}

	shell := poly.GetExteriorRing()

	shellLoc := pl.locateInPolygonRing(p, shell)
	if shellLoc == Geom_Location_Exterior {
		return Geom_Location_Exterior
	}
	if shellLoc == Geom_Location_Boundary {
		return Geom_Location_Boundary
	}
	// Now test if the point lies in or on the holes.
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		hole := poly.GetInteriorRingN(i)
		holeLoc := pl.locateInPolygonRing(p, hole)
		if holeLoc == Geom_Location_Interior {
			return Geom_Location_Exterior
		}
		if holeLoc == Geom_Location_Boundary {
			return Geom_Location_Boundary
		}
	}
	return Geom_Location_Interior
}
