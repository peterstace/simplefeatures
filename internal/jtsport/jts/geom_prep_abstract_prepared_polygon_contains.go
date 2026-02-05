package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

type GeomPrep_AbstractPreparedPolygonContains struct {
	*GeomPrep_PreparedPolygonPredicate
	child                      java.Polymorphic
	requireSomePointInInterior bool

	// information about geometric situation
	hasSegmentIntersection   bool
	hasProperIntersection    bool
	hasNonProperIntersection bool
}

func (a *GeomPrep_AbstractPreparedPolygonContains) GetChild() java.Polymorphic  { return a.child }
func (a *GeomPrep_AbstractPreparedPolygonContains) GetParent() java.Polymorphic { return nil }

func geomPrep_NewAbstractPreparedPolygonContains(prepPoly *GeomPrep_PreparedPolygon) *GeomPrep_AbstractPreparedPolygonContains {
	base := geomPrep_NewPreparedPolygonPredicate(prepPoly)
	a := &GeomPrep_AbstractPreparedPolygonContains{
		GeomPrep_PreparedPolygonPredicate: base,
		requireSomePointInInterior:        true,
	}
	base.child = a
	return a
}

func (a *GeomPrep_AbstractPreparedPolygonContains) eval(geom *Geom_Geometry) bool {
	if geom.GetDimension() == 0 {
		return a.evalPoints(geom)
	}

	isAllInTargetArea := a.isAllTestComponentsInTarget(geom)
	if !isAllInTargetArea {
		return false
	}

	properIntersectionImpliesNotContained := a.isProperIntersectionImpliesNotContainedSituation(geom)

	// find all intersection types which exist
	a.findAndClassifyIntersections(geom)

	if properIntersectionImpliesNotContained && a.hasProperIntersection {
		return false
	}

	if a.hasSegmentIntersection && !a.hasNonProperIntersection {
		return false
	}

	if a.hasSegmentIntersection {
		return a.fullTopologicalPredicate(geom)
	}

	if java.InstanceOf[Geom_Polygonal](geom) {
		// TODO: generalize this to handle GeometryCollections
		isTargetInTestArea := a.isAnyTargetComponentInAreaTest(geom, a.prepPoly.GetRepresentativePoints())
		if isTargetInTestArea {
			return false
		}
	}
	return true
}

func (a *GeomPrep_AbstractPreparedPolygonContains) evalPoints(geom *Geom_Geometry) bool {
	isAllInTargetArea := a.isAllTestPointsInTarget(geom)
	if !isAllInTargetArea {
		return false
	}

	if a.requireSomePointInInterior {
		isAnyInTargetInterior := a.isAnyTestPointInTargetInterior(geom)
		return isAnyInTargetInterior
	}
	return true
}

func (a *GeomPrep_AbstractPreparedPolygonContains) isProperIntersectionImpliesNotContainedSituation(testGeom *Geom_Geometry) bool {
	if java.InstanceOf[Geom_Polygonal](testGeom) {
		return true
	}
	if a.isSingleShell(a.prepPoly.GetGeometry()) {
		return true
	}
	return false
}

func (a *GeomPrep_AbstractPreparedPolygonContains) isSingleShell(geom *Geom_Geometry) bool {
	// handles single-element MultiPolygons, as well as Polygons
	if geom.GetNumGeometries() != 1 {
		return false
	}

	poly := java.Cast[*Geom_Polygon](geom.GetGeometryN(0))
	numHoles := poly.GetNumInteriorRing()
	if numHoles == 0 {
		return true
	}
	return false
}

func (a *GeomPrep_AbstractPreparedPolygonContains) findAndClassifyIntersections(geom *Geom_Geometry) {
	lineSegStr := Noding_SegmentStringUtil_ExtractSegmentStrings(geom)

	intDetector := Noding_NewSegmentIntersectionDetector()
	intDetector.SetFindAllIntersectionTypes(true)
	a.prepPoly.GetIntersectionFinder().IntersectsWithDetector(lineSegStr, intDetector)

	a.hasSegmentIntersection = intDetector.HasIntersection()
	a.hasProperIntersection = intDetector.HasProperIntersection()
	a.hasNonProperIntersection = intDetector.HasNonProperIntersection()
}

func (a *GeomPrep_AbstractPreparedPolygonContains) fullTopologicalPredicate(geom *Geom_Geometry) bool {
	if impl, ok := java.GetLeaf(a).(interface {
		FullTopologicalPredicate_BODY(*Geom_Geometry) bool
	}); ok {
		return impl.FullTopologicalPredicate_BODY(geom)
	}
	panic("abstract method called")
}
