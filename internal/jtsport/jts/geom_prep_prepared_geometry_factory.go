package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

type GeomPrep_PreparedGeometryFactory struct {
}

func GeomPrep_PreparedGeometryFactory_Prepare(geom *Geom_Geometry) GeomPrep_PreparedGeometry {
	return GeomPrep_NewPreparedGeometryFactory().Create(geom)
}

func GeomPrep_NewPreparedGeometryFactory() *GeomPrep_PreparedGeometryFactory {
	return &GeomPrep_PreparedGeometryFactory{}
}

func (f *GeomPrep_PreparedGeometryFactory) Create(geom *Geom_Geometry) GeomPrep_PreparedGeometry {
	if java.InstanceOf[Geom_Polygonal](geom) {
		return GeomPrep_NewPreparedPolygon(java.GetLeaf(geom).(Geom_Polygonal))
	}
	if java.InstanceOf[Geom_Lineal](geom) {
		return GeomPrep_NewPreparedLineString(java.GetLeaf(geom).(Geom_Lineal))
	}
	if java.InstanceOf[Geom_Puntal](geom) {
		return GeomPrep_NewPreparedPoint(java.GetLeaf(geom).(Geom_Puntal))
	}

	return GeomPrep_NewBasicPreparedGeometry(geom)
}
