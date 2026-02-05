package jts

type GeomPrep_PreparedGeometry interface {
	IsGeomPrep_PreparedGeometry()
	GetGeometry() *Geom_Geometry
	Contains(geom *Geom_Geometry) bool
	ContainsProperly(geom *Geom_Geometry) bool
	CoveredBy(geom *Geom_Geometry) bool
	Covers(geom *Geom_Geometry) bool
	Crosses(geom *Geom_Geometry) bool
	Disjoint(geom *Geom_Geometry) bool
	Intersects(geom *Geom_Geometry) bool
	Overlaps(geom *Geom_Geometry) bool
	Touches(geom *Geom_Geometry) bool
	Within(geom *Geom_Geometry) bool
}
