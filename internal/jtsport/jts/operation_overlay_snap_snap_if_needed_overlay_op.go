package jts

// OperationOverlaySnap_SnapIfNeededOverlayOp performs an overlay operation using
// snapping and enhanced precision to improve the robustness of the result. This
// class only uses snapping if an error is detected when running the standard
// JTS overlay code. Errors detected include thrown exceptions (in particular,
// TopologyException) and invalid overlay computations.
type OperationOverlaySnap_SnapIfNeededOverlayOp struct {
	geom []*Geom_Geometry
}

// OperationOverlaySnap_SnapIfNeededOverlayOp_OverlayOp computes an overlay
// operation, using snapping if the standard overlay fails.
func OperationOverlaySnap_SnapIfNeededOverlayOp_OverlayOp(g0, g1 *Geom_Geometry, opCode int) *Geom_Geometry {
	op := OperationOverlaySnap_NewSnapIfNeededOverlayOp(g0, g1)
	return op.GetResultGeometry(opCode)
}

// OperationOverlaySnap_SnapIfNeededOverlayOp_Intersection computes the
// intersection, using snapping if the standard overlay fails.
func OperationOverlaySnap_SnapIfNeededOverlayOp_Intersection(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlaySnap_SnapIfNeededOverlayOp_OverlayOp(g0, g1, OperationOverlay_OverlayOp_Intersection)
}

// OperationOverlaySnap_SnapIfNeededOverlayOp_Union computes the union, using
// snapping if the standard overlay fails.
func OperationOverlaySnap_SnapIfNeededOverlayOp_Union(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlaySnap_SnapIfNeededOverlayOp_OverlayOp(g0, g1, OperationOverlay_OverlayOp_Union)
}

// OperationOverlaySnap_SnapIfNeededOverlayOp_Difference computes the difference,
// using snapping if the standard overlay fails.
func OperationOverlaySnap_SnapIfNeededOverlayOp_Difference(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlaySnap_SnapIfNeededOverlayOp_OverlayOp(g0, g1, OperationOverlay_OverlayOp_Difference)
}

// OperationOverlaySnap_SnapIfNeededOverlayOp_SymDifference computes the symmetric
// difference, using snapping if the standard overlay fails.
func OperationOverlaySnap_SnapIfNeededOverlayOp_SymDifference(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlaySnap_SnapIfNeededOverlayOp_OverlayOp(g0, g1, OperationOverlay_OverlayOp_SymDifference)
}

// OperationOverlaySnap_NewSnapIfNeededOverlayOp creates a new
// SnapIfNeededOverlayOp.
func OperationOverlaySnap_NewSnapIfNeededOverlayOp(g1, g2 *Geom_Geometry) *OperationOverlaySnap_SnapIfNeededOverlayOp {
	op := &OperationOverlaySnap_SnapIfNeededOverlayOp{
		geom: make([]*Geom_Geometry, 2),
	}
	op.geom[0] = g1
	op.geom[1] = g2
	return op
}

// GetResultGeometry computes the overlay result geometry.
func (sinoo *OperationOverlaySnap_SnapIfNeededOverlayOp) GetResultGeometry(opCode int) *Geom_Geometry {
	var result *Geom_Geometry
	isSuccess := false
	var savedException any

	// Try basic operation with input geometries.
	func() {
		defer func() {
			if r := recover(); r != nil {
				savedException = r
			}
		}()
		result = OperationOverlay_OverlayOp_OverlayOp(sinoo.geom[0], sinoo.geom[1], opCode)
		isValid := true
		// Not needed if noding validation is used.
		if isValid {
			isSuccess = true
		}
	}()

	if !isSuccess {
		// This may still throw an exception.
		// If so, throw the original exception since it has the input coordinates.
		func() {
			defer func() {
				if r := recover(); r != nil {
					panic(savedException)
				}
			}()
			result = OperationOverlaySnap_SnapOverlayOp_OverlayOp(sinoo.geom[0], sinoo.geom[1], opCode)
		}()
	}
	return result
}
