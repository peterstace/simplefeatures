package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

type GeomPrep_PreparedPolygonIntersects struct {
	*GeomPrep_PreparedPolygonPredicate
}

func (c *GeomPrep_PreparedPolygonIntersects) GetChild() java.Polymorphic  { return nil }
func (c *GeomPrep_PreparedPolygonIntersects) GetParent() java.Polymorphic { return nil }

func GeomPrep_PreparedPolygonIntersects_Intersects(prep *GeomPrep_PreparedPolygon, geom *Geom_Geometry) bool {
	polyInt := geomPrep_NewPreparedPolygonIntersects(prep)
	return polyInt.Intersects(geom)
}

func geomPrep_NewPreparedPolygonIntersects(prepPoly *GeomPrep_PreparedPolygon) *GeomPrep_PreparedPolygonIntersects {
	base := geomPrep_NewPreparedPolygonPredicate(prepPoly)
	c := &GeomPrep_PreparedPolygonIntersects{
		GeomPrep_PreparedPolygonPredicate: base,
	}
	base.child = c
	return c
}

func (c *GeomPrep_PreparedPolygonIntersects) Intersects(geom *Geom_Geometry) bool {
	// Do point-in-poly tests first, since they are cheaper and may result in a
	// quick positive result.
	//
	// If a point of any test components lie in target, result is true
	isInPrepGeomArea := c.isAnyTestComponentInTarget(geom)
	if isInPrepGeomArea {
		return true
	}
	// If input contains only points, then at
	// this point it is known that none of them are contained in the target
	if geom.GetDimension() == 0 {
		return false
	}
	// If any segments intersect, result is true
	lineSegStr := Noding_SegmentStringUtil_ExtractSegmentStrings(geom)
	// only request intersection finder if there are segments
	// (i.e. NOT for point inputs)
	if len(lineSegStr) > 0 {
		segsIntersect := c.prepPoly.GetIntersectionFinder().Intersects(lineSegStr)
		if segsIntersect {
			return true
		}
	}

	// If the test has dimension = 2 as well, it is necessary to test for proper
	// inclusion of the target. Since no segments intersect, it is sufficient to
	// test representative points.
	if geom.GetDimension() == 2 {
		// TODO: generalize this to handle GeometryCollections
		isPrepGeomInArea := c.isAnyTargetComponentInAreaTest(geom, c.prepPoly.GetRepresentativePoints())
		if isPrepGeomInArea {
			return true
		}
	}

	return false
}
