package jts

import "math"

// OperationOverlayng_OverlayUtil provides utility methods for overlay
// processing.

const (
	operationOverlayng_OverlayUtil_SAFE_ENV_BUFFER_FACTOR   = 0.1
	operationOverlayng_OverlayUtil_SAFE_ENV_GRID_FACTOR     = 3
	operationOverlayng_OverlayUtil_AREA_HEURISTIC_TOLERANCE = 0.1
)

// OperationOverlayng_OverlayUtil_IsFloating is a null-handling wrapper for
// PrecisionModel.IsFloating().
func OperationOverlayng_OverlayUtil_IsFloating(pm *Geom_PrecisionModel) bool {
	if pm == nil {
		return true
	}
	return pm.IsFloating()
}

// OperationOverlayng_OverlayUtil_ClippingEnvelope computes a clipping envelope
// for overlay input geometries. The clipping envelope encloses all geometry
// line segments which might participate in the overlay, with a buffer to
// account for numerical precision.
func OperationOverlayng_OverlayUtil_ClippingEnvelope(opCode int, inputGeom *OperationOverlayng_InputGeometry, pm *Geom_PrecisionModel) *Geom_Envelope {
	resultEnv := operationOverlayng_OverlayUtil_resultEnvelope(opCode, inputGeom, pm)
	if resultEnv == nil {
		return nil
	}

	clipEnv := OperationOverlayng_RobustClipEnvelopeComputer_GetEnvelope(
		inputGeom.GetGeometry(0),
		inputGeom.GetGeometry(1),
		resultEnv,
	)

	safeEnv := operationOverlayng_OverlayUtil_safeEnv(clipEnv, pm)
	return safeEnv
}

// resultEnvelope computes an envelope which covers the extent of the result of
// a given overlay operation for given inputs.
func operationOverlayng_OverlayUtil_resultEnvelope(opCode int, inputGeom *OperationOverlayng_InputGeometry, pm *Geom_PrecisionModel) *Geom_Envelope {
	var overlapEnv *Geom_Envelope
	switch opCode {
	case OperationOverlayng_OverlayNG_INTERSECTION:
		// Use safe envelopes for intersection to ensure they contain rounded
		// coordinates.
		envA := operationOverlayng_OverlayUtil_safeEnv(inputGeom.GetEnvelope(0), pm)
		envB := operationOverlayng_OverlayUtil_safeEnv(inputGeom.GetEnvelope(1), pm)
		overlapEnv = envA.Intersection(envB)
	case OperationOverlayng_OverlayNG_DIFFERENCE:
		overlapEnv = operationOverlayng_OverlayUtil_safeEnv(inputGeom.GetEnvelope(0), pm)
	}
	// Return nil for UNION and SYMDIFFERENCE to indicate no clipping.
	return overlapEnv
}

// safeEnv determines a safe geometry envelope for clipping, taking into
// account the precision model being used.
func operationOverlayng_OverlayUtil_safeEnv(env *Geom_Envelope, pm *Geom_PrecisionModel) *Geom_Envelope {
	envExpandDist := operationOverlayng_OverlayUtil_safeExpandDistance(env, pm)
	safeEnv := env.Copy()
	safeEnv.ExpandBy(envExpandDist)
	return safeEnv
}

func operationOverlayng_OverlayUtil_safeExpandDistance(env *Geom_Envelope, pm *Geom_PrecisionModel) float64 {
	var envExpandDist float64
	if OperationOverlayng_OverlayUtil_IsFloating(pm) {
		// If PM is FLOAT then there is no scale factor, so add 10%.
		minSize := math.Min(env.GetHeight(), env.GetWidth())
		// Heuristic to ensure zero-width envelopes don't cause total clipping.
		if minSize <= 0.0 {
			minSize = math.Max(env.GetHeight(), env.GetWidth())
		}
		envExpandDist = operationOverlayng_OverlayUtil_SAFE_ENV_BUFFER_FACTOR * minSize
	} else {
		// If PM is fixed, add a small multiple of the grid size.
		gridSize := 1.0 / pm.GetScale()
		envExpandDist = float64(operationOverlayng_OverlayUtil_SAFE_ENV_GRID_FACTOR) * gridSize
	}
	return envExpandDist
}

// OperationOverlayng_OverlayUtil_IsEmptyResult tests if the result can be
// determined to be empty based on simple properties of the input geometries.
func OperationOverlayng_OverlayUtil_IsEmptyResult(opCode int, a, b *Geom_Geometry, pm *Geom_PrecisionModel) bool {
	switch opCode {
	case OperationOverlayng_OverlayNG_INTERSECTION:
		if OperationOverlayng_OverlayUtil_IsEnvDisjoint(a, b, pm) {
			return true
		}
	case OperationOverlayng_OverlayNG_DIFFERENCE:
		if operationOverlayng_OverlayUtil_isEmpty(a) {
			return true
		}
	case OperationOverlayng_OverlayNG_UNION, OperationOverlayng_OverlayNG_SYMDIFFERENCE:
		if operationOverlayng_OverlayUtil_isEmpty(a) && operationOverlayng_OverlayUtil_isEmpty(b) {
			return true
		}
	}
	return false
}

func operationOverlayng_OverlayUtil_isEmpty(geom *Geom_Geometry) bool {
	return geom == nil || geom.IsEmpty()
}

// OperationOverlayng_OverlayUtil_IsEnvDisjoint tests if the geometry envelopes
// are disjoint, or empty.
func OperationOverlayng_OverlayUtil_IsEnvDisjoint(a, b *Geom_Geometry, pm *Geom_PrecisionModel) bool {
	if operationOverlayng_OverlayUtil_isEmpty(a) || operationOverlayng_OverlayUtil_isEmpty(b) {
		return true
	}
	if OperationOverlayng_OverlayUtil_IsFloating(pm) {
		return a.GetEnvelopeInternal().Disjoint(b.GetEnvelopeInternal())
	}
	return operationOverlayng_OverlayUtil_isDisjoint(a.GetEnvelopeInternal(), b.GetEnvelopeInternal(), pm)
}

// isDisjoint tests for disjoint envelopes adjusting for rounding caused by a
// fixed precision model.
func operationOverlayng_OverlayUtil_isDisjoint(envA, envB *Geom_Envelope, pm *Geom_PrecisionModel) bool {
	if pm.MakePrecise(envB.GetMinX()) > pm.MakePrecise(envA.GetMaxX()) {
		return true
	}
	if pm.MakePrecise(envB.GetMaxX()) < pm.MakePrecise(envA.GetMinX()) {
		return true
	}
	if pm.MakePrecise(envB.GetMinY()) > pm.MakePrecise(envA.GetMaxY()) {
		return true
	}
	if pm.MakePrecise(envB.GetMaxY()) < pm.MakePrecise(envA.GetMinY()) {
		return true
	}
	return false
}

// OperationOverlayng_OverlayUtil_CreateEmptyResult creates an empty result
// geometry of the appropriate dimension.
func OperationOverlayng_OverlayUtil_CreateEmptyResult(dim int, geomFact *Geom_GeometryFactory) *Geom_Geometry {
	var result *Geom_Geometry
	switch dim {
	case 0:
		result = geomFact.CreatePoint().Geom_Geometry
	case 1:
		result = geomFact.CreateLineString().Geom_Geometry
	case 2:
		result = geomFact.CreatePolygon().Geom_Geometry
	case -1:
		result = geomFact.CreateGeometryCollectionFromGeometries([]*Geom_Geometry{}).Geom_Geometry
	default:
		Util_Assert_ShouldNeverReachHereWithMessage("Unable to determine overlay result geometry dimension")
	}
	return result
}

// OperationOverlayng_OverlayUtil_ResultDimension computes the dimension of the
// result of applying the given operation to inputs with the given dimensions.
func OperationOverlayng_OverlayUtil_ResultDimension(opCode, dim0, dim1 int) int {
	resultDimension := -1
	switch opCode {
	case OperationOverlayng_OverlayNG_INTERSECTION:
		if dim0 < dim1 {
			resultDimension = dim0
		} else {
			resultDimension = dim1
		}
	case OperationOverlayng_OverlayNG_UNION:
		if dim0 > dim1 {
			resultDimension = dim0
		} else {
			resultDimension = dim1
		}
	case OperationOverlayng_OverlayNG_DIFFERENCE:
		resultDimension = dim0
	case OperationOverlayng_OverlayNG_SYMDIFFERENCE:
		// SymDiff = Union( Diff(A, B), Diff(B, A) ) and Union has the dimension
		// of the highest-dimension argument.
		if dim0 > dim1 {
			resultDimension = dim0
		} else {
			resultDimension = dim1
		}
	}
	return resultDimension
}

// OperationOverlayng_OverlayUtil_CreateResultGeometry creates an overlay result
// geometry for homogeneous or mixed components.
func OperationOverlayng_OverlayUtil_CreateResultGeometry(resultPolyList []*Geom_Polygon, resultLineList []*Geom_LineString, resultPointList []*Geom_Point, geometryFactory *Geom_GeometryFactory) *Geom_Geometry {
	geomList := make([]*Geom_Geometry, 0)

	// Element geometries of the result are always in the order A,L,P.
	for _, poly := range resultPolyList {
		geomList = append(geomList, poly.Geom_Geometry)
	}
	for _, line := range resultLineList {
		geomList = append(geomList, line.Geom_Geometry)
	}
	for _, pt := range resultPointList {
		geomList = append(geomList, pt.Geom_Geometry)
	}

	// Build the most specific geometry possible.
	return geometryFactory.BuildGeometry(geomList)
}

// OperationOverlayng_OverlayUtil_ToLines converts the overlay graph to lines
// for debugging purposes.
func OperationOverlayng_OverlayUtil_ToLines(graph *OperationOverlayng_OverlayGraph, isOutputEdges bool, geomFact *Geom_GeometryFactory) *Geom_Geometry {
	lines := make([]*Geom_Geometry, 0)
	for _, edge := range graph.GetEdges() {
		includeEdge := isOutputEdges || edge.IsInResultArea()
		if !includeEdge {
			continue
		}
		pts := edge.GetCoordinatesOriented()
		line := geomFact.CreateLineStringFromCoordinates(pts)
		line.SetUserData(operationOverlayng_OverlayUtil_labelForResult(edge))
		lines = append(lines, line.Geom_Geometry)
	}
	return geomFact.BuildGeometry(lines)
}

func operationOverlayng_OverlayUtil_labelForResult(edge *OperationOverlayng_OverlayEdge) string {
	result := edge.GetLabel().ToStringWithDirection(edge.IsForward())
	if edge.IsInResultArea() {
		result += " Res"
	}
	return result
}

// OperationOverlayng_OverlayUtil_RoundPoint rounds the key point if precision
// model is fixed.
func OperationOverlayng_OverlayUtil_RoundPoint(pt *Geom_Point, pm *Geom_PrecisionModel) *Geom_Coordinate {
	if pt.IsEmpty() {
		return nil
	}
	return OperationOverlayng_OverlayUtil_Round(pt.GetCoordinate(), pm)
}

// OperationOverlayng_OverlayUtil_Round rounds a coordinate if precision model
// is fixed.
func OperationOverlayng_OverlayUtil_Round(p *Geom_Coordinate, pm *Geom_PrecisionModel) *Geom_Coordinate {
	if !OperationOverlayng_OverlayUtil_IsFloating(pm) {
		pRound := p.Copy()
		pm.MakePreciseCoordinate(pRound)
		return pRound
	}
	return p
}

// OperationOverlayng_OverlayUtil_IsResultAreaConsistent is a heuristic check
// for overlay result correctness comparing the areas of the input and result.
func OperationOverlayng_OverlayUtil_IsResultAreaConsistent(geom0, geom1 *Geom_Geometry, opCode int, result *Geom_Geometry) bool {
	if geom0 == nil || geom1 == nil {
		return true
	}

	if result.GetDimension() < 2 {
		return true
	}

	areaResult := result.GetArea()
	areaA := geom0.GetArea()
	areaB := geom1.GetArea()

	isConsistent := true
	switch opCode {
	case OperationOverlayng_OverlayNG_INTERSECTION:
		isConsistent = operationOverlayng_OverlayUtil_isLess(areaResult, areaA, operationOverlayng_OverlayUtil_AREA_HEURISTIC_TOLERANCE) &&
			operationOverlayng_OverlayUtil_isLess(areaResult, areaB, operationOverlayng_OverlayUtil_AREA_HEURISTIC_TOLERANCE)
	case OperationOverlayng_OverlayNG_DIFFERENCE:
		isConsistent = operationOverlayng_OverlayUtil_isDifferenceAreaConsistent(areaA, areaB, areaResult, operationOverlayng_OverlayUtil_AREA_HEURISTIC_TOLERANCE)
	case OperationOverlayng_OverlayNG_SYMDIFFERENCE:
		isConsistent = operationOverlayng_OverlayUtil_isLess(areaResult, areaA+areaB, operationOverlayng_OverlayUtil_AREA_HEURISTIC_TOLERANCE)
	case OperationOverlayng_OverlayNG_UNION:
		isConsistent = operationOverlayng_OverlayUtil_isLess(areaA, areaResult, operationOverlayng_OverlayUtil_AREA_HEURISTIC_TOLERANCE) &&
			operationOverlayng_OverlayUtil_isLess(areaB, areaResult, operationOverlayng_OverlayUtil_AREA_HEURISTIC_TOLERANCE) &&
			operationOverlayng_OverlayUtil_isGreater(areaResult, areaA-areaB, operationOverlayng_OverlayUtil_AREA_HEURISTIC_TOLERANCE)
	}
	return isConsistent
}

func operationOverlayng_OverlayUtil_isDifferenceAreaConsistent(areaA, areaB, areaResult, tolFrac float64) bool {
	if !operationOverlayng_OverlayUtil_isLess(areaResult, areaA, tolFrac) {
		return false
	}
	areaDiffMin := areaA - areaB - tolFrac*areaA
	return areaResult > areaDiffMin
}

func operationOverlayng_OverlayUtil_isLess(v1, v2, tol float64) bool {
	return v1 <= v2*(1+tol)
}

func operationOverlayng_OverlayUtil_isGreater(v1, v2, tol float64) bool {
	return v1 >= v2*(1-tol)
}
