package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

type GeomPrep_PreparedPolygonContainsProperly struct {
	*GeomPrep_PreparedPolygonPredicate
}

func (c *GeomPrep_PreparedPolygonContainsProperly) GetChild() java.Polymorphic  { return nil }
func (c *GeomPrep_PreparedPolygonContainsProperly) GetParent() java.Polymorphic { return nil }

func GeomPrep_PreparedPolygonContainsProperly_ContainsProperly(prep *GeomPrep_PreparedPolygon, geom *Geom_Geometry) bool {
	polyInt := geomPrep_NewPreparedPolygonContainsProperly(prep)
	return polyInt.ContainsProperly(geom)
}

func geomPrep_NewPreparedPolygonContainsProperly(prepPoly *GeomPrep_PreparedPolygon) *GeomPrep_PreparedPolygonContainsProperly {
	base := geomPrep_NewPreparedPolygonPredicate(prepPoly)
	c := &GeomPrep_PreparedPolygonContainsProperly{
		GeomPrep_PreparedPolygonPredicate: base,
	}
	base.child = c
	return c
}

func (c *GeomPrep_PreparedPolygonContainsProperly) ContainsProperly(geom *Geom_Geometry) bool {
	isAllInPrepGeomAreaInterior := c.isAllTestComponentsInTargetInterior(geom)
	if !isAllInPrepGeomAreaInterior {
		return false
	}

	lineSegStr := Noding_SegmentStringUtil_ExtractSegmentStrings(geom)
	segsIntersect := c.prepPoly.GetIntersectionFinder().Intersects(lineSegStr)
	if segsIntersect {
		return false
	}

	if java.InstanceOf[Geom_Polygonal](geom) {
		// TODO: generalize this to handle GeometryCollections
		isTargetGeomInTestArea := c.isAnyTargetComponentInAreaTest(geom, c.prepPoly.GetRepresentativePoints())
		if isTargetGeomInTestArea {
			return false
		}
	}

	return true
}
