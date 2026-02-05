package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

type GeomPrep_PreparedPolygonContains struct {
	*GeomPrep_AbstractPreparedPolygonContains
}

func (c *GeomPrep_PreparedPolygonContains) GetChild() java.Polymorphic { return nil }

func GeomPrep_PreparedPolygonContains_Contains(prep *GeomPrep_PreparedPolygon, geom *Geom_Geometry) bool {
	polyInt := geomPrep_NewPreparedPolygonContains(prep)
	return polyInt.Contains(geom)
}

func geomPrep_NewPreparedPolygonContains(prepPoly *GeomPrep_PreparedPolygon) *GeomPrep_PreparedPolygonContains {
	base := geomPrep_NewAbstractPreparedPolygonContains(prepPoly)
	c := &GeomPrep_PreparedPolygonContains{
		GeomPrep_AbstractPreparedPolygonContains: base,
	}
	base.child = c
	return c
}

func (c *GeomPrep_PreparedPolygonContains) Contains(geom *Geom_Geometry) bool {
	return c.eval(geom)
}

func (c *GeomPrep_PreparedPolygonContains) FullTopologicalPredicate_BODY(geom *Geom_Geometry) bool {
	isContained := c.prepPoly.GetGeometry().Contains(geom)
	return isContained
}
