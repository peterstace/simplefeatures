package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

type GeomPrep_PreparedPolygonCovers struct {
	*GeomPrep_AbstractPreparedPolygonContains
}

func (c *GeomPrep_PreparedPolygonCovers) GetChild() java.Polymorphic { return nil }

func GeomPrep_PreparedPolygonCovers_Covers(prep *GeomPrep_PreparedPolygon, geom *Geom_Geometry) bool {
	polyInt := geomPrep_NewPreparedPolygonCovers(prep)
	return polyInt.Covers(geom)
}

func geomPrep_NewPreparedPolygonCovers(prepPoly *GeomPrep_PreparedPolygon) *GeomPrep_PreparedPolygonCovers {
	base := geomPrep_NewAbstractPreparedPolygonContains(prepPoly)
	c := &GeomPrep_PreparedPolygonCovers{
		GeomPrep_AbstractPreparedPolygonContains: base,
	}
	base.requireSomePointInInterior = false
	base.child = c
	return c
}

func (c *GeomPrep_PreparedPolygonCovers) Covers(geom *Geom_Geometry) bool {
	return c.eval(geom)
}

func (c *GeomPrep_PreparedPolygonCovers) FullTopologicalPredicate_BODY(geom *Geom_Geometry) bool {
	result := c.prepPoly.GetGeometry().Covers(geom)
	return result
}
