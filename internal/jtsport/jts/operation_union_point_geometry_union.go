package jts

import (
	"sort"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationUnion_PointGeometryUnion computes the union of a Puntal geometry
// with another arbitrary Geometry. Does not copy any component geometries.
type OperationUnion_PointGeometryUnion struct {
	pointGeom *Geom_Geometry
	otherGeom *Geom_Geometry
	geomFact  *Geom_GeometryFactory
}

// OperationUnion_PointGeometryUnion_Union computes the union of a puntal
// geometry with another geometry.
func OperationUnion_PointGeometryUnion_Union(pointGeom Geom_Puntal, otherGeom *Geom_Geometry) *Geom_Geometry {
	unioner := OperationUnion_NewPointGeometryUnion(pointGeom, otherGeom)
	return unioner.Union()
}

// OperationUnion_NewPointGeometryUnion creates a new PointGeometryUnion.
func OperationUnion_NewPointGeometryUnion(pointGeom Geom_Puntal, otherGeom *Geom_Geometry) *OperationUnion_PointGeometryUnion {
	// Get the base Geom_Geometry from the Puntal.
	var pg *Geom_Geometry
	switch p := pointGeom.(type) {
	case *Geom_Point:
		pg = p.Geom_Geometry
	case *Geom_MultiPoint:
		pg = p.Geom_GeometryCollection.Geom_Geometry
	}
	return &OperationUnion_PointGeometryUnion{
		pointGeom: pg,
		otherGeom: otherGeom,
		geomFact:  otherGeom.GetFactory(),
	}
}

// Union computes the union of the point geometry with the other geometry.
func (pgu *OperationUnion_PointGeometryUnion) Union() *Geom_Geometry {
	locater := Algorithm_NewPointLocator()
	// Use a map to eliminate duplicates, as required for union.
	exteriorCoords := make(map[coordKey]*Geom_Coordinate)

	for i := 0; i < pgu.pointGeom.GetNumGeometries(); i++ {
		point := java.GetLeaf(pgu.pointGeom.GetGeometryN(i)).(*Geom_Point)
		coord := point.GetCoordinate()
		loc := locater.Locate(coord, pgu.otherGeom)
		if loc == Geom_Location_Exterior {
			key := coordKey{x: coord.GetX(), y: coord.GetY()}
			exteriorCoords[key] = coord
		}
	}

	// If no points are in exterior, return the other geom.
	if len(exteriorCoords) == 0 {
		return pgu.otherGeom
	}

	// Convert map values to sorted slice (TreeSet equivalent).
	coords := make([]*Geom_Coordinate, 0, len(exteriorCoords))
	for _, c := range exteriorCoords {
		coords = append(coords, c)
	}
	sort.Slice(coords, func(i, j int) bool {
		return coords[i].CompareTo(coords[j]) < 0
	})

	// Make a puntal geometry of appropriate size.
	var ptComp *Geom_Geometry
	if len(coords) == 1 {
		ptComp = pgu.geomFact.CreatePointFromCoordinate(coords[0]).Geom_Geometry
	} else {
		ptComp = pgu.geomFact.CreateMultiPointFromCoords(coords).Geom_GeometryCollection.Geom_Geometry
	}

	// Add point component to the other geometry.
	return GeomUtil_GeometryCombiner_Combine2(ptComp, pgu.otherGeom)
}

// coordKey is used as a map key for coordinate deduplication.
type coordKey struct {
	x, y float64
}
