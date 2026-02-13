package jts

import (
	"sync"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

type GeomPrep_PreparedPolygon struct {
	*GeomPrep_BasicPreparedGeometry
	isRectangle  bool
	segIntFinder *Noding_FastSegmentSetIntersectionFinder
	pia          AlgorithmLocate_PointOnGeometryLocator
	// TRANSLITERATION NOTE: sync.Mutex replaces Java's synchronized keyword.
	mu sync.Mutex
}

func GeomPrep_NewPreparedPolygon(poly Geom_Polygonal) *GeomPrep_PreparedPolygon {
	base := GeomPrep_NewBasicPreparedGeometry(java.Cast[*Geom_Geometry](poly.(java.Polymorphic)))
	pp := &GeomPrep_PreparedPolygon{
		GeomPrep_BasicPreparedGeometry: base,
		isRectangle:                    base.GetGeometry().IsRectangle(),
	}
	base.child = pp
	return pp
}

func (p *GeomPrep_PreparedPolygon) GetChild() java.Polymorphic { return nil }

// TRANSLITERATION NOTE: sync.Mutex replaces Java's synchronized keyword.
func (p *GeomPrep_PreparedPolygon) GetIntersectionFinder() *Noding_FastSegmentSetIntersectionFinder {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.segIntFinder == nil {
		p.segIntFinder = Noding_NewFastSegmentSetIntersectionFinder(Noding_SegmentStringUtil_ExtractSegmentStrings(p.GetGeometry()))
	}
	return p.segIntFinder
}

// TRANSLITERATION NOTE: sync.Mutex replaces Java's synchronized keyword.
func (p *GeomPrep_PreparedPolygon) GetPointLocator() AlgorithmLocate_PointOnGeometryLocator {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pia == nil {
		p.pia = AlgorithmLocate_NewIndexedPointInAreaLocator(p.GetGeometry())
	}
	return p.pia
}

func (p *GeomPrep_PreparedPolygon) Intersects_BODY(g *Geom_Geometry) bool {
	// envelope test
	if !p.envelopesIntersect(g) {
		return false
	}

	// optimization for rectangles
	if p.isRectangle {
		return OperationPredicate_RectangleIntersects_Intersects(java.Cast[*Geom_Polygon](p.GetGeometry()), g)
	}

	return GeomPrep_PreparedPolygonIntersects_Intersects(p, g)
}

func (p *GeomPrep_PreparedPolygon) Contains_BODY(g *Geom_Geometry) bool {
	// short-circuit test
	if !p.envelopeCovers(g) {
		return false
	}

	// optimization for rectangles
	if p.isRectangle {
		return OperationPredicate_RectangleContains_Contains(java.Cast[*Geom_Polygon](p.GetGeometry()), g)
	}

	return GeomPrep_PreparedPolygonContains_Contains(p, g)
}

func (p *GeomPrep_PreparedPolygon) ContainsProperly_BODY(g *Geom_Geometry) bool {
	// short-circuit test
	if !p.envelopeCovers(g) {
		return false
	}
	return GeomPrep_PreparedPolygonContainsProperly_ContainsProperly(p, g)
}

func (p *GeomPrep_PreparedPolygon) Covers_BODY(g *Geom_Geometry) bool {
	// short-circuit test
	if !p.envelopeCovers(g) {
		return false
	}
	// optimization for rectangle arguments
	if p.isRectangle {
		return true
	}
	return GeomPrep_PreparedPolygonCovers_Covers(p, g)
}
