package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

type GeomPrep_PreparedPolygonPredicate struct {
	child              java.Polymorphic
	prepPoly           *GeomPrep_PreparedPolygon
	targetPointLocator AlgorithmLocate_PointOnGeometryLocator
}

func (p *GeomPrep_PreparedPolygonPredicate) GetChild() java.Polymorphic { return p.child }

func geomPrep_NewPreparedPolygonPredicate(prepPoly *GeomPrep_PreparedPolygon) *GeomPrep_PreparedPolygonPredicate {
	return &GeomPrep_PreparedPolygonPredicate{
		prepPoly:           prepPoly,
		targetPointLocator: prepPoly.GetPointLocator(),
	}
}

func (p *GeomPrep_PreparedPolygonPredicate) isAllTestComponentsInTarget(testGeom *Geom_Geometry) bool {
	coords := GeomUtil_ComponentCoordinateExtracter_GetCoordinates(testGeom)
	for _, c := range coords {
		loc := p.targetPointLocator.Locate(c)
		if loc == Geom_Location_Exterior {
			return false
		}
	}
	return true
}

func (p *GeomPrep_PreparedPolygonPredicate) isAllTestComponentsInTargetInterior(testGeom *Geom_Geometry) bool {
	coords := GeomUtil_ComponentCoordinateExtracter_GetCoordinates(testGeom)
	for _, c := range coords {
		loc := p.targetPointLocator.Locate(c)
		if loc != Geom_Location_Interior {
			return false
		}
	}
	return true
}

func (p *GeomPrep_PreparedPolygonPredicate) isAnyTestComponentInTarget(testGeom *Geom_Geometry) bool {
	coords := GeomUtil_ComponentCoordinateExtracter_GetCoordinates(testGeom)
	for _, c := range coords {
		loc := p.targetPointLocator.Locate(c)
		if loc != Geom_Location_Exterior {
			return true
		}
	}
	return false
}

func (p *GeomPrep_PreparedPolygonPredicate) isAllTestPointsInTarget(testGeom *Geom_Geometry) bool {
	for i := 0; i < testGeom.GetNumGeometries(); i++ {
		pt := java.Cast[*Geom_Point](testGeom.GetGeometryN(i))
		c := pt.GetCoordinate()
		loc := p.targetPointLocator.Locate(c)
		if loc == Geom_Location_Exterior {
			return false
		}
	}
	return true
}

func (p *GeomPrep_PreparedPolygonPredicate) isAnyTestPointInTargetInterior(testGeom *Geom_Geometry) bool {
	for i := 0; i < testGeom.GetNumGeometries(); i++ {
		pt := java.Cast[*Geom_Point](testGeom.GetGeometryN(i))
		c := pt.GetCoordinate()
		loc := p.targetPointLocator.Locate(c)
		if loc == Geom_Location_Interior {
			return true
		}
	}
	return false
}

func (p *GeomPrep_PreparedPolygonPredicate) isAnyTargetComponentInAreaTest(testGeom *Geom_Geometry, targetRepPts []*Geom_Coordinate) bool {
	piaLoc := AlgorithmLocate_NewSimplePointInAreaLocator(testGeom)
	for _, c := range targetRepPts {
		loc := piaLoc.Locate(c)
		if loc != Geom_Location_Exterior {
			return true
		}
	}
	return false
}
