package jts

// OperationUnion_UnionInteracting is experimental code to union MultiPolygons
// with processing limited to the elements which actually interact.
//
// Not currently used, since it doesn't seem to offer much of a performance
// advantage.
type OperationUnion_UnionInteracting struct {
	geomFactory *Geom_GeometryFactory
	g0          *Geom_Geometry
	g1          *Geom_Geometry
	interacts0  []bool
	interacts1  []bool
}

// OperationUnion_UnionInteracting_Union unions two geometries.
func OperationUnion_UnionInteracting_Union(g0, g1 *Geom_Geometry) *Geom_Geometry {
	uue := OperationUnion_NewUnionInteracting(g0, g1)
	return uue.Union()
}

// OperationUnion_NewUnionInteracting creates a new UnionInteracting instance.
func OperationUnion_NewUnionInteracting(g0, g1 *Geom_Geometry) *OperationUnion_UnionInteracting {
	return &OperationUnion_UnionInteracting{
		g0:          g0,
		g1:          g1,
		geomFactory: g0.GetFactory(),
		interacts0:  make([]bool, g0.GetNumGeometries()),
		interacts1:  make([]bool, g1.GetNumGeometries()),
	}
}

// Union performs the union operation.
func (ui *OperationUnion_UnionInteracting) Union() *Geom_Geometry {
	ui.computeInteracting()

	// Check for all interacting or none interacting!
	int0 := ui.extractElements(ui.g0, ui.interacts0, true)
	int1 := ui.extractElements(ui.g1, ui.interacts1, true)

	union := int0.Union(int1)

	disjoint0 := ui.extractElements(ui.g0, ui.interacts0, false)
	disjoint1 := ui.extractElements(ui.g1, ui.interacts1, false)

	overallUnion := GeomUtil_GeometryCombiner_Combine3(union, disjoint0, disjoint1)

	return overallUnion
}

func (ui *OperationUnion_UnionInteracting) computeInteracting() {
	for i := 0; i < ui.g0.GetNumGeometries(); i++ {
		elem := ui.g0.GetGeometryN(i)
		ui.interacts0[i] = ui.computeInteractingElem(elem)
	}
}

func (ui *OperationUnion_UnionInteracting) computeInteractingElem(elem0 *Geom_Geometry) bool {
	interactsWithAny := false
	for i := 0; i < ui.g1.GetNumGeometries(); i++ {
		elem1 := ui.g1.GetGeometryN(i)
		interacts := elem1.GetEnvelopeInternal().IntersectsEnvelope(elem0.GetEnvelopeInternal())
		if interacts {
			ui.interacts1[i] = true
		}
		if interacts {
			interactsWithAny = true
		}
	}
	return interactsWithAny
}

func (ui *OperationUnion_UnionInteracting) extractElements(geom *Geom_Geometry, interacts []bool, isInteracting bool) *Geom_Geometry {
	extractedGeoms := make([]*Geom_Geometry, 0)
	for i := 0; i < geom.GetNumGeometries(); i++ {
		elem := geom.GetGeometryN(i)
		if interacts[i] == isInteracting {
			extractedGeoms = append(extractedGeoms, elem)
		}
	}
	return ui.geomFactory.BuildGeometry(extractedGeoms)
}
