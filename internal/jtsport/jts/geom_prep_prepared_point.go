package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

type GeomPrep_PreparedPoint struct {
	*GeomPrep_BasicPreparedGeometry
}

func (p *GeomPrep_PreparedPoint) GetChild() java.Polymorphic { return nil }

func GeomPrep_NewPreparedPoint(point Geom_Puntal) *GeomPrep_PreparedPoint {
	base := GeomPrep_NewBasicPreparedGeometry(java.Cast[*Geom_Geometry](point.(java.Polymorphic)))
	pp := &GeomPrep_PreparedPoint{GeomPrep_BasicPreparedGeometry: base}
	base.child = pp
	return pp
}

func (p *GeomPrep_PreparedPoint) Intersects_BODY(g *Geom_Geometry) bool {
	if !p.envelopesIntersect(g) {
		return false
	}

	return p.IsAnyTargetComponentInTest(g)
}
