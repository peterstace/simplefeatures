package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelateng_AdjacentEdgeLocator determines the location for a point
// which is known to lie on at least one edge of a set of polygons. This
// provides the union-semantics for determining point location in a
// GeometryCollection, which may have polygons with adjacent edges which are
// effectively in the interior of the geometry. Note that it is also possible to
// have adjacent edges which lie on the boundary of the geometry (e.g. a polygon
// contained within another polygon with adjacent edges).
type OperationRelateng_AdjacentEdgeLocator struct {
	ringList [][]*Geom_Coordinate
}

// OperationRelateng_NewAdjacentEdgeLocator creates a new AdjacentEdgeLocator
// for the given geometry.
func OperationRelateng_NewAdjacentEdgeLocator(geom *Geom_Geometry) *OperationRelateng_AdjacentEdgeLocator {
	ael := &OperationRelateng_AdjacentEdgeLocator{}
	ael.init(geom)
	return ael
}

// Locate locates a point that is known to lie on a polygon edge.
func (ael *OperationRelateng_AdjacentEdgeLocator) Locate(p *Geom_Coordinate) int {
	sections := OperationRelateng_NewNodeSections(p)
	for _, ring := range ael.ringList {
		ael.addSections(p, ring, sections)
	}
	node := sections.CreateNode()
	if node.HasExteriorEdge(true) {
		return Geom_Location_Boundary
	}
	return Geom_Location_Interior
}

func (ael *OperationRelateng_AdjacentEdgeLocator) addSections(p *Geom_Coordinate, ring []*Geom_Coordinate, sections *OperationRelateng_NodeSections) {
	for i := 0; i < len(ring)-1; i++ {
		p0 := ring[i]
		pnext := ring[i+1]

		if p.Equals2D(pnext) {
			// Segment final point is assigned to next segment.
			continue
		} else if p.Equals2D(p0) {
			iprev := i - 1
			if i == 0 {
				iprev = len(ring) - 2
			}
			pprev := ring[iprev]
			sections.AddNodeSection(ael.createSection(p, pprev, pnext))
		} else if Algorithm_PointLocation_IsOnSegment(p, p0, pnext) {
			sections.AddNodeSection(ael.createSection(p, p0, pnext))
		}
	}
}

func (ael *OperationRelateng_AdjacentEdgeLocator) createSection(p, prev, next *Geom_Coordinate) *OperationRelateng_NodeSection {
	// Note: the Java code has debug logging here for zero-length segments.
	return OperationRelateng_NewNodeSection(true, Geom_Dimension_A, 1, 0, nil, false, prev, p, next)
}

func (ael *OperationRelateng_AdjacentEdgeLocator) init(geom *Geom_Geometry) {
	if geom.IsEmpty() {
		return
	}
	ael.ringList = make([][]*Geom_Coordinate, 0)
	ael.addRings(geom)
}

func (ael *OperationRelateng_AdjacentEdgeLocator) addRings(geom *Geom_Geometry) {
	if java.InstanceOf[*Geom_Polygon](geom) {
		poly := java.Cast[*Geom_Polygon](geom)
		shell := poly.GetExteriorRing()
		ael.addRing(shell, true)
		for i := 0; i < poly.GetNumInteriorRing(); i++ {
			hole := poly.GetInteriorRingN(i)
			ael.addRing(hole, false)
		}
	} else if java.InstanceOf[*Geom_GeometryCollection](geom) {
		// Recurse through collections.
		for i := 0; i < geom.GetNumGeometries(); i++ {
			ael.addRings(geom.GetGeometryN(i))
		}
	}
}

func (ael *OperationRelateng_AdjacentEdgeLocator) addRing(ring *Geom_LinearRing, requireCW bool) {
	// TODO: remove repeated points?
	pts := OperationRelateng_RelateGeometry_Orient(ring.GetCoordinates(), requireCW)
	ael.ringList = append(ael.ringList, pts)
}
