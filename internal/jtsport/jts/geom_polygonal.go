package jts

// Geom_Polygonal identifies Geometry subclasses which are 2-dimensional and have
// components which have Lineal boundaries.
type Geom_Polygonal interface {
	IsPolygonal()
}
