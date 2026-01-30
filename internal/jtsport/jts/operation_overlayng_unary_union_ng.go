package jts

// OperationOverlayng_UnaryUnionNG unions a geometry or collection of geometries
// in an efficient way, using OverlayNG to ensure robust computation.
//
// This class is most useful for performing UnaryUnion using a fixed-precision
// model. For unary union using floating precision, OverlayNGRobust_Union should
// be used.

// OperationOverlayng_UnaryUnionNG_UnionGeom unions a geometry (which is often a
// collection) using a given precision model.
func OperationOverlayng_UnaryUnionNG_UnionGeom(geom *Geom_Geometry, pm *Geom_PrecisionModel) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromGeometry(geom)
	op.SetUnionFunction(operationOverlayng_UnaryUnionNG_createUnionStrategy(pm))
	return op.Union()
}

// OperationOverlayng_UnaryUnionNG_UnionCollection unions a collection of
// geometries using a given precision model.
func OperationOverlayng_UnaryUnionNG_UnionCollection(geoms []*Geom_Geometry, pm *Geom_PrecisionModel) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromCollection(geoms)
	op.SetUnionFunction(operationOverlayng_UnaryUnionNG_createUnionStrategy(pm))
	return op.Union()
}

// OperationOverlayng_UnaryUnionNG_UnionCollectionWithFactory unions a collection
// of geometries using a given precision model.
func OperationOverlayng_UnaryUnionNG_UnionCollectionWithFactory(geoms []*Geom_Geometry, geomFact *Geom_GeometryFactory, pm *Geom_PrecisionModel) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromCollectionWithFactory(geoms, geomFact)
	op.SetUnionFunction(operationOverlayng_UnaryUnionNG_createUnionStrategy(pm))
	return op.Union()
}

func operationOverlayng_UnaryUnionNG_createUnionStrategy(pm *Geom_PrecisionModel) OperationUnion_UnionStrategy {
	return &unaryUnionNGUnionStrategy{pm: pm}
}

type unaryUnionNGUnionStrategy struct {
	pm *Geom_PrecisionModel
}

func (s *unaryUnionNGUnionStrategy) Union(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlayng_OverlayNG_Overlay(g0, g1, OperationOverlayng_OverlayNG_UNION, s.pm)
}

func (s *unaryUnionNGUnionStrategy) IsFloatingPrecision() bool {
	return OperationOverlayng_OverlayUtil_IsFloating(s.pm)
}
