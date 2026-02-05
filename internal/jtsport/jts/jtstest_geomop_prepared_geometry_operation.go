package jts

var _ JtstestGeomop_GeometryOperation = (*JtstestGeomop_PreparedGeometryOperation)(nil)

type JtstestGeomop_PreparedGeometryOperation struct {
	chainOp *JtstestGeomop_GeometryMethodOperation
}

func JtstestGeomop_NewPreparedGeometryOperation() *JtstestGeomop_PreparedGeometryOperation {
	return &JtstestGeomop_PreparedGeometryOperation{
		chainOp: JtstestGeomop_NewGeometryMethodOperation(),
	}
}

func (op *JtstestGeomop_PreparedGeometryOperation) IsJtstestGeomop_GeometryOperation() {}

func (op *JtstestGeomop_PreparedGeometryOperation) GetReturnType(opName string) string {
	if jtstestGeomop_PreparedGeometryOperation_isPreparedOp(opName) {
		return "boolean"
	}
	return op.chainOp.GetReturnType(opName)
}

func JtstestGeomop_NewPreparedGeometryOperationWithChainOp(chainOp *JtstestGeomop_GeometryMethodOperation) *JtstestGeomop_PreparedGeometryOperation {
	return &JtstestGeomop_PreparedGeometryOperation{
		chainOp: chainOp,
	}
}

func jtstestGeomop_PreparedGeometryOperation_isPreparedOp(opName string) bool {
	if opName == "intersects" {
		return true
	}
	if opName == "contains" {
		return true
	}
	if opName == "containsProperly" {
		return true
	}
	if opName == "covers" {
		return true
	}
	return false
}

func (op *JtstestGeomop_PreparedGeometryOperation) Invoke(opName string, geometry *Geom_Geometry, args []any) (JtstestTestrunner_Result, error) {
	if !jtstestGeomop_PreparedGeometryOperation_isPreparedOp(opName) {
		return op.chainOp.Invoke(opName, geometry, args)
	}
	return op.invokePreparedOp(opName, geometry, args), nil
}

func (op *JtstestGeomop_PreparedGeometryOperation) invokePreparedOp(opName string, geometry *Geom_Geometry, args []any) JtstestTestrunner_Result {
	g2 := args[0].(*Geom_Geometry)
	if opName == "intersects" {
		return JtstestTestrunner_NewBooleanResult(jtstestGeomop_PreparedGeometryOp_intersects(geometry, g2))
	}
	if opName == "contains" {
		return JtstestTestrunner_NewBooleanResult(jtstestGeomop_PreparedGeometryOp_contains(geometry, g2))
	}
	if opName == "containsProperly" {
		return JtstestTestrunner_NewBooleanResult(jtstestGeomop_PreparedGeometryOp_containsProperly(geometry, g2))
	}
	if opName == "covers" {
		return JtstestTestrunner_NewBooleanResult(jtstestGeomop_PreparedGeometryOp_covers(geometry, g2))
	}
	return nil
}

// Inner class PreparedGeometryOp

func jtstestGeomop_PreparedGeometryOp_intersects(g1 *Geom_Geometry, g2 *Geom_Geometry) bool {
	prepGeom := GeomPrep_PreparedGeometryFactory_Prepare(g1)
	return prepGeom.Intersects(g2)
}

func jtstestGeomop_PreparedGeometryOp_contains(g1 *Geom_Geometry, g2 *Geom_Geometry) bool {
	prepGeom := GeomPrep_PreparedGeometryFactory_Prepare(g1)
	return prepGeom.Contains(g2)
}

func jtstestGeomop_PreparedGeometryOp_containsProperly(g1 *Geom_Geometry, g2 *Geom_Geometry) bool {
	prepGeom := GeomPrep_PreparedGeometryFactory_Prepare(g1)
	return prepGeom.ContainsProperly(g2)
}

func jtstestGeomop_PreparedGeometryOp_covers(g1 *Geom_Geometry, g2 *Geom_Geometry) bool {
	prepGeom := GeomPrep_PreparedGeometryFactory_Prepare(g1)
	return prepGeom.Covers(g2)
}
