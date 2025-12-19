package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_ShortCircuitedGeometryVisitor is a visitor to Geometry components,
// which allows short-circuiting when a defined condition holds.
type GeomUtil_ShortCircuitedGeometryVisitor struct {
	isDone bool
}

func (v *GeomUtil_ShortCircuitedGeometryVisitor) ApplyTo(geom *Geom_Geometry, impl geomUtil_ShortCircuitedGeometryVisitorImpl) {
	for i := 0; i < geom.GetNumGeometries() && !v.isDone; i++ {
		element := geom.GetGeometryN(i)
		if !java.InstanceOf[*Geom_GeometryCollection](element) {
			impl.Visit(element)
			if impl.IsDone() {
				v.isDone = true
				return
			}
		} else {
			v.ApplyTo(element, impl)
		}
	}
}

// geomUtil_ShortCircuitedGeometryVisitorImpl is the interface that concrete
// visitors must implement.
type geomUtil_ShortCircuitedGeometryVisitorImpl interface {
	Visit(element *Geom_Geometry)
	IsDone() bool
}
