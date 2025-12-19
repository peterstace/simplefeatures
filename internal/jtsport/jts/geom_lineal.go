package jts

// Geom_Lineal identifies Geometry subclasses which are 1-dimensional and have
// components which are LineStrings.
type Geom_Lineal interface {
	IsLineal()
}
