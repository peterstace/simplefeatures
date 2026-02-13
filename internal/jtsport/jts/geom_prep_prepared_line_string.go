package jts

import (
	"sync"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

type GeomPrep_PreparedLineString struct {
	*GeomPrep_BasicPreparedGeometry
	segIntFinder *Noding_FastSegmentSetIntersectionFinder
	// TRANSLITERATION NOTE: sync.Mutex replaces Java's synchronized keyword.
	mu sync.Mutex
}

func GeomPrep_NewPreparedLineString(line Geom_Lineal) *GeomPrep_PreparedLineString {
	base := GeomPrep_NewBasicPreparedGeometry(java.Cast[*Geom_Geometry](line.(java.Polymorphic)))
	pls := &GeomPrep_PreparedLineString{GeomPrep_BasicPreparedGeometry: base}
	base.child = pls
	return pls
}

func (p *GeomPrep_PreparedLineString) GetChild() java.Polymorphic { return nil }

// TRANSLITERATION NOTE: sync.Mutex replaces Java's synchronized keyword.
func (p *GeomPrep_PreparedLineString) GetIntersectionFinder() *Noding_FastSegmentSetIntersectionFinder {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.segIntFinder == nil {
		p.segIntFinder = Noding_NewFastSegmentSetIntersectionFinder(Noding_SegmentStringUtil_ExtractSegmentStrings(p.GetGeometry()))
	}
	return p.segIntFinder
}

func (p *GeomPrep_PreparedLineString) Intersects_BODY(g *Geom_Geometry) bool {
	if !p.envelopesIntersect(g) {
		return false
	}
	return GeomPrep_PreparedLineStringIntersects_Intersects(p, g)
}
