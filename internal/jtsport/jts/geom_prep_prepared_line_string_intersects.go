package jts

func GeomPrep_PreparedLineStringIntersects_Intersects(prep *GeomPrep_PreparedLineString, geom *Geom_Geometry) bool {
	op := GeomPrep_NewPreparedLineStringIntersects(prep)
	return op.Intersects(geom)
}

type GeomPrep_PreparedLineStringIntersects struct {
	prepLine *GeomPrep_PreparedLineString
}

func GeomPrep_NewPreparedLineStringIntersects(prepLine *GeomPrep_PreparedLineString) *GeomPrep_PreparedLineStringIntersects {
	return &GeomPrep_PreparedLineStringIntersects{prepLine: prepLine}
}

func (op *GeomPrep_PreparedLineStringIntersects) Intersects(geom *Geom_Geometry) bool {
	// If any segments intersect, obviously intersects = true
	lineSegStr := Noding_SegmentStringUtil_ExtractSegmentStrings(geom)
	// only request intersection finder if there are segments (ie NOT for point inputs)
	if len(lineSegStr) > 0 {
		segsIntersect := op.prepLine.GetIntersectionFinder().Intersects(lineSegStr)
		if segsIntersect {
			return true
		}
	}

	// For L/A case, need to check for proper inclusion of the target in the test
	if geom.GetDimension() == 2 && op.prepLine.IsAnyTargetComponentInTest(geom) {
		return true
	}

	// For L/P case, need to check if any points lie on line(s)
	if geom.HasDimension(0) {
		return op.isAnyTestPointInTarget(geom)
	}

	return false
}

func (op *GeomPrep_PreparedLineStringIntersects) isAnyTestPointInTarget(testGeom *Geom_Geometry) bool {
	// This could be optimized by using the segment index on the lineal target.
	// However, it seems like the L/P case would be pretty rare in practice.
	locator := Algorithm_NewPointLocator()
	coords := GeomUtil_ComponentCoordinateExtracter_GetCoordinates(testGeom)
	for _, p := range coords {
		if locator.Intersects(p, op.prepLine.GetGeometry()) {
			return true
		}
	}
	return false
}
