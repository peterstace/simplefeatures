package jts

import "math"

// IndexStrtree_GeometryItemDistance is an ItemDistance function for items which
// are Geometrys, using the Geometry.Distance method.
//
// To make this distance function suitable for using to query a single index
// tree, the distance metric is anti-reflexive. That is, if the two arguments
// are the same Geometry object, the distance returned is math.MaxFloat64.
type IndexStrtree_GeometryItemDistance struct{}

var _ IndexStrtree_ItemDistance = (*IndexStrtree_GeometryItemDistance)(nil)

func (gid *IndexStrtree_GeometryItemDistance) IsIndexStrtree_ItemDistance() {}

// IndexStrtree_NewGeometryItemDistance creates a new GeometryItemDistance.
func IndexStrtree_NewGeometryItemDistance() *IndexStrtree_GeometryItemDistance {
	return &IndexStrtree_GeometryItemDistance{}
}

// Distance computes the distance between two Geometry items, using the
// Geometry.Distance method.
func (gid *IndexStrtree_GeometryItemDistance) Distance(item1, item2 *IndexStrtree_ItemBoundable) float64 {
	if item1 == item2 {
		return math.MaxFloat64
	}
	g1 := item1.GetItem().(*Geom_Geometry)
	g2 := item2.GetItem().(*Geom_Geometry)
	return g1.Distance(g2)
}
