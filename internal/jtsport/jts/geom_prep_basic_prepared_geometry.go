package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

var _ GeomPrep_PreparedGeometry = (*GeomPrep_BasicPreparedGeometry)(nil)

type GeomPrep_BasicPreparedGeometry struct {
	child             java.Polymorphic
	baseGeom          *Geom_Geometry
	representativePts []*Geom_Coordinate
}

func (b *GeomPrep_BasicPreparedGeometry) IsGeomPrep_PreparedGeometry() {}

func (b *GeomPrep_BasicPreparedGeometry) GetChild() java.Polymorphic { return b.child }

func (b *GeomPrep_BasicPreparedGeometry) GetParent() java.Polymorphic { return nil }

func GeomPrep_NewBasicPreparedGeometry(geom *Geom_Geometry) *GeomPrep_BasicPreparedGeometry {
	return &GeomPrep_BasicPreparedGeometry{
		baseGeom:          geom,
		representativePts: GeomUtil_ComponentCoordinateExtracter_GetCoordinates(geom),
	}
}

func (b *GeomPrep_BasicPreparedGeometry) GetGeometry() *Geom_Geometry { return b.baseGeom }

func (b *GeomPrep_BasicPreparedGeometry) GetRepresentativePoints() []*Geom_Coordinate {
	//TODO wrap in unmodifiable?
	return b.representativePts
}

func (b *GeomPrep_BasicPreparedGeometry) IsAnyTargetComponentInTest(testGeom *Geom_Geometry) bool {
	locator := Algorithm_NewPointLocator()
	for _, p := range b.representativePts {
		if locator.Intersects(p, testGeom) {
			return true
		}
	}
	return false
}

func (b *GeomPrep_BasicPreparedGeometry) envelopesIntersect(g *Geom_Geometry) bool {
	if !b.baseGeom.GetEnvelopeInternal().IntersectsEnvelope(g.GetEnvelopeInternal()) {
		return false
	}
	return true
}

func (b *GeomPrep_BasicPreparedGeometry) envelopeCovers(g *Geom_Geometry) bool {
	if !b.baseGeom.GetEnvelopeInternal().CoversEnvelope(g.GetEnvelopeInternal()) {
		return false
	}
	return true
}

// Contains dispatcher - overridden in PreparedPolygon.
func (b *GeomPrep_BasicPreparedGeometry) Contains(g *Geom_Geometry) bool {
	if impl, ok := java.GetLeaf(b).(interface{ Contains_BODY(*Geom_Geometry) bool }); ok {
		return impl.Contains_BODY(g)
	}
	return b.Contains_BODY(g)
}

func (b *GeomPrep_BasicPreparedGeometry) Contains_BODY(g *Geom_Geometry) bool {
	return b.baseGeom.Contains(g)
}

// ContainsProperly dispatcher - overridden in PreparedPolygon.
func (b *GeomPrep_BasicPreparedGeometry) ContainsProperly(g *Geom_Geometry) bool {
	if impl, ok := java.GetLeaf(b).(interface{ ContainsProperly_BODY(*Geom_Geometry) bool }); ok {
		return impl.ContainsProperly_BODY(g)
	}
	return b.ContainsProperly_BODY(g)
}

func (b *GeomPrep_BasicPreparedGeometry) ContainsProperly_BODY(g *Geom_Geometry) bool {
	// since raw relate is used, provide some optimizations

	// short-circuit test
	if !b.baseGeom.GetEnvelopeInternal().ContainsEnvelope(g.GetEnvelopeInternal()) {
		return false
	}

	// otherwise, compute using relate mask
	return b.baseGeom.Relate(g, "T**FF*FF*")
}

func (b *GeomPrep_BasicPreparedGeometry) CoveredBy(g *Geom_Geometry) bool {
	return b.baseGeom.CoveredBy(g)
}

// Covers dispatcher - overridden in PreparedPolygon.
func (b *GeomPrep_BasicPreparedGeometry) Covers(g *Geom_Geometry) bool {
	if impl, ok := java.GetLeaf(b).(interface{ Covers_BODY(*Geom_Geometry) bool }); ok {
		return impl.Covers_BODY(g)
	}
	return b.Covers_BODY(g)
}

func (b *GeomPrep_BasicPreparedGeometry) Covers_BODY(g *Geom_Geometry) bool {
	return b.baseGeom.Covers(g)
}

func (b *GeomPrep_BasicPreparedGeometry) Crosses(g *Geom_Geometry) bool {
	return b.baseGeom.Crosses(g)
}

func (b *GeomPrep_BasicPreparedGeometry) Disjoint(g *Geom_Geometry) bool {
	return !b.Intersects(g)
}

// Intersects dispatcher - overridden in PreparedPoint, PreparedLineString, PreparedPolygon.
func (b *GeomPrep_BasicPreparedGeometry) Intersects(g *Geom_Geometry) bool {
	if impl, ok := java.GetLeaf(b).(interface{ Intersects_BODY(*Geom_Geometry) bool }); ok {
		return impl.Intersects_BODY(g)
	}
	return b.Intersects_BODY(g)
}

func (b *GeomPrep_BasicPreparedGeometry) Intersects_BODY(g *Geom_Geometry) bool {
	return b.baseGeom.Intersects(g)
}

func (b *GeomPrep_BasicPreparedGeometry) Overlaps(g *Geom_Geometry) bool {
	return b.baseGeom.Overlaps(g)
}

func (b *GeomPrep_BasicPreparedGeometry) Touches(g *Geom_Geometry) bool {
	return b.baseGeom.Touches(g)
}

func (b *GeomPrep_BasicPreparedGeometry) Within(g *Geom_Geometry) bool {
	return b.baseGeom.Within(g)
}

func (b *GeomPrep_BasicPreparedGeometry) String() string {
	return b.baseGeom.String()
}
