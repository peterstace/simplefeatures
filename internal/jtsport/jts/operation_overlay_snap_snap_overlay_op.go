package jts

// OperationOverlaySnap_SnapOverlayOp performs an overlay operation using
// snapping and enhanced precision to improve the robustness of the result. This
// class always uses snapping. This is less performant than the standard JTS
// overlay code, and may even introduce errors which were not present in the
// original data. For this reason, this class should only be used if the
// standard overlay code fails to produce a correct result.
type OperationOverlaySnap_SnapOverlayOp struct {
	geom          []*Geom_Geometry
	snapTolerance float64
	cbr           *Precision_CommonBitsRemover
}

// OperationOverlaySnap_SnapOverlayOp_OverlayOp computes an overlay operation
// using snapping.
func OperationOverlaySnap_SnapOverlayOp_OverlayOp(g0, g1 *Geom_Geometry, opCode int) *Geom_Geometry {
	op := OperationOverlaySnap_NewSnapOverlayOp(g0, g1)
	return op.GetResultGeometry(opCode)
}

// OperationOverlaySnap_SnapOverlayOp_Intersection computes the intersection
// using snapping.
func OperationOverlaySnap_SnapOverlayOp_Intersection(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlaySnap_SnapOverlayOp_OverlayOp(g0, g1, OperationOverlay_OverlayOp_Intersection)
}

// OperationOverlaySnap_SnapOverlayOp_Union computes the union using snapping.
func OperationOverlaySnap_SnapOverlayOp_Union(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlaySnap_SnapOverlayOp_OverlayOp(g0, g1, OperationOverlay_OverlayOp_Union)
}

// OperationOverlaySnap_SnapOverlayOp_Difference computes the difference using
// snapping.
func OperationOverlaySnap_SnapOverlayOp_Difference(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlaySnap_SnapOverlayOp_OverlayOp(g0, g1, OperationOverlay_OverlayOp_Difference)
}

// OperationOverlaySnap_SnapOverlayOp_SymDifference computes the symmetric
// difference using snapping.
func OperationOverlaySnap_SnapOverlayOp_SymDifference(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlaySnap_SnapOverlayOp_OverlayOp(g0, g1, OperationOverlay_OverlayOp_SymDifference)
}

// OperationOverlaySnap_NewSnapOverlayOp creates a new SnapOverlayOp.
func OperationOverlaySnap_NewSnapOverlayOp(g1, g2 *Geom_Geometry) *OperationOverlaySnap_SnapOverlayOp {
	op := &OperationOverlaySnap_SnapOverlayOp{
		geom: make([]*Geom_Geometry, 2),
	}
	op.geom[0] = g1
	op.geom[1] = g2
	op.computeSnapTolerance()
	return op
}

func (soo *OperationOverlaySnap_SnapOverlayOp) computeSnapTolerance() {
	soo.snapTolerance = OperationOverlaySnap_GeometrySnapper_ComputeOverlaySnapToleranceFromTwo(soo.geom[0], soo.geom[1])
}

// GetResultGeometry computes the overlay result geometry.
func (soo *OperationOverlaySnap_SnapOverlayOp) GetResultGeometry(opCode int) *Geom_Geometry {
	prepGeom := soo.snap(soo.geom)
	result := OperationOverlay_OverlayOp_OverlayOp(prepGeom[0], prepGeom[1], opCode)
	return soo.prepareResult(result)
}

func (soo *OperationOverlaySnap_SnapOverlayOp) selfSnap(geom *Geom_Geometry) *Geom_Geometry {
	snapper0 := OperationOverlaySnap_NewGeometrySnapper(geom)
	return snapper0.SnapTo(geom, soo.snapTolerance)
}

func (soo *OperationOverlaySnap_SnapOverlayOp) snap(geom []*Geom_Geometry) []*Geom_Geometry {
	remGeom := soo.removeCommonBits(geom)

	snapGeom := OperationOverlaySnap_GeometrySnapper_Snap(remGeom[0], remGeom[1], soo.snapTolerance)
	return snapGeom
}

func (soo *OperationOverlaySnap_SnapOverlayOp) prepareResult(geom *Geom_Geometry) *Geom_Geometry {
	soo.cbr.AddCommonBits(geom)
	return geom
}

func (soo *OperationOverlaySnap_SnapOverlayOp) removeCommonBits(geom []*Geom_Geometry) []*Geom_Geometry {
	soo.cbr = Precision_NewCommonBitsRemover()
	soo.cbr.Add(geom[0])
	soo.cbr.Add(geom[1])
	remGeom := make([]*Geom_Geometry, 2)
	remGeom[0] = soo.cbr.RemoveCommonBits(geom[0].Copy())
	remGeom[1] = soo.cbr.RemoveCommonBits(geom[1].Copy())
	return remGeom
}
