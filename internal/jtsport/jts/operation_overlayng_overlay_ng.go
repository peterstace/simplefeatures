package jts

// OperationOverlayng_OverlayNG constants for overlay operations.
const (
	OperationOverlayng_OverlayNG_INTERSECTION  = 1
	OperationOverlayng_OverlayNG_UNION         = 2
	OperationOverlayng_OverlayNG_DIFFERENCE    = 3
	OperationOverlayng_OverlayNG_SYMDIFFERENCE = 4
)

// OperationOverlayng_OverlayNG_STRICT_MODE_DEFAULT is the default setting for
// strict mode. The original JTS overlay semantics used non-strict result
// semantics, including:
//   - An Intersection result can be mixed-dimension, due to inclusion of
//     intersection components of all dimensions
//   - Results can include lines caused by Area topology collapse
const OperationOverlayng_OverlayNG_STRICT_MODE_DEFAULT = false

// OperationOverlayng_OverlayNG_IsResultOfOp tests whether a point with given
// Locations relative to two geometries would be contained in the result of
// overlaying the geometries using a given overlay operation. This is used to
// determine whether components computed during the overlay process should be
// included in the result geometry.
//
// The method handles arguments of Location.NONE correctly.
func OperationOverlayng_OverlayNG_IsResultOfOp(overlayOpCode, loc0, loc1 int) bool {
	if loc0 == Geom_Location_Boundary {
		loc0 = Geom_Location_Interior
	}
	if loc1 == Geom_Location_Boundary {
		loc1 = Geom_Location_Interior
	}
	switch overlayOpCode {
	case OperationOverlayng_OverlayNG_INTERSECTION:
		return loc0 == Geom_Location_Interior && loc1 == Geom_Location_Interior
	case OperationOverlayng_OverlayNG_UNION:
		return loc0 == Geom_Location_Interior || loc1 == Geom_Location_Interior
	case OperationOverlayng_OverlayNG_DIFFERENCE:
		return loc0 == Geom_Location_Interior && loc1 != Geom_Location_Interior
	case OperationOverlayng_OverlayNG_SYMDIFFERENCE:
		return (loc0 == Geom_Location_Interior && loc1 != Geom_Location_Interior) ||
			(loc0 != Geom_Location_Interior && loc1 == Geom_Location_Interior)
	}
	return false
}

// OperationOverlayng_OverlayNG computes the geometric overlay of two Geometrys,
// using an explicit precision model to allow robust computation.
//
// The overlay can be used to determine any of the following set-theoretic
// operations (boolean combinations) of the geometries:
//   - INTERSECTION - all points which lie in both geometries
//   - UNION - all points which lie in at least one geometry
//   - DIFFERENCE - all points which lie in the first geometry but not the second
//   - SYMDIFFERENCE - all points which lie in one geometry but not both
//
// Input geometries may have different dimension. Input collections must be
// homogeneous (all elements must have the same dimension). Inputs may be simple
// GeometryCollections. A GeometryCollection is simple if it can be flattened
// into a valid Multi-geometry; i.e. it is homogeneous and does not contain any
// overlapping Polygons.
//
// The precision model used for the computation can be supplied independent of
// the precision model of the input geometry. The main use for this is to allow
// using a fixed precision for geometry with a floating precision model. This
// does two things: ensures robust computation; and forces the output to be
// validly rounded to the precision model.
//
// For fixed precision models noding is performed using a SnapRoundingNoder.
// This provides robust computation (as long as precision is limited to around
// 13 decimal digits).
//
// For floating precision an MCIndexNoder is used. This is not fully robust, so
// can sometimes result in TopologyExceptions being thrown. For robust
// full-precision overlay see OverlayNGRobust.
//
// A custom Noder can be supplied. This allows using a more performant noding
// strategy in specific cases, for instance in CoverageUnion.
//
// Note: If a SnappingNoder is used it is best to specify a fairly small snap
// tolerance, since the intersection clipping optimization can interact with the
// snapping to alter the result.
//
// Optionally the overlay computation can process using strict mode (via
// SetStrictMode). In strict mode result semantics are:
//   - Lines and Points resulting from topology collapses are not included in
//     the result
//   - Result geometry is homogeneous for the INTERSECTION and DIFFERENCE
//     operations.
//   - Result geometry is homogeneous for the UNION and SYMDIFFERENCE operations
//     if the inputs have the same dimension
//
// Strict mode has the following benefits:
//   - Results are simpler
//   - Overlay operations are chainable without needing to remove
//     lower-dimension elements
//
// The original JTS overlay semantics corresponds to non-strict mode.
//
// If a robustness error occurs, a TopologyException is thrown. These are
// usually caused by numerical rounding causing the noding output to not be
// fully noded. For robust computation with full-precision OverlayNGRobust can
// be used.
type OperationOverlayng_OverlayNG struct {
	opCode              int
	inputGeom           *OperationOverlayng_InputGeometry
	geomFact            *Geom_GeometryFactory
	pm                  *Geom_PrecisionModel
	noder               Noding_Noder
	isStrictMode        bool
	isOptimized         bool
	isAreaResultOnly    bool
	isOutputEdges       bool
	isOutputResultEdges bool
	isOutputNodedEdges  bool
}

// OperationOverlayng_NewOverlayNGWithPM creates an overlay operation on the
// given geometries, with a defined precision model. The noding strategy is
// determined by the precision model.
func OperationOverlayng_NewOverlayNGWithPM(geom0, geom1 *Geom_Geometry, pm *Geom_PrecisionModel, opCode int) *OperationOverlayng_OverlayNG {
	return &OperationOverlayng_OverlayNG{
		pm:           pm,
		opCode:       opCode,
		geomFact:     geom0.GetFactory(),
		inputGeom:    OperationOverlayng_NewInputGeometry(geom0, geom1),
		isStrictMode: OperationOverlayng_OverlayNG_STRICT_MODE_DEFAULT,
		isOptimized:  true,
	}
}

// OperationOverlayng_NewOverlayNG creates an overlay operation on the given
// geometries using the precision model of the geometries.
//
// The noder is chosen according to the precision model specified.
//   - For FIXED a snap-rounding noder is used, and the computation is robust.
//   - For FLOATING a non-snapping noder is used, and this computation may not
//     be robust. If errors occur a TopologyException is thrown.
func OperationOverlayng_NewOverlayNG(geom0, geom1 *Geom_Geometry, opCode int) *OperationOverlayng_OverlayNG {
	return OperationOverlayng_NewOverlayNGWithPM(geom0, geom1, geom0.GetFactory().GetPrecisionModel(), opCode)
}

// OperationOverlayng_NewOverlayNGUnary creates a union of a single geometry
// with a given precision model.
func OperationOverlayng_NewOverlayNGUnary(geom *Geom_Geometry, pm *Geom_PrecisionModel) *OperationOverlayng_OverlayNG {
	return OperationOverlayng_NewOverlayNGWithPM(geom, nil, pm, OperationOverlayng_OverlayNG_UNION)
}

// OperationOverlayng_OverlayNG_Overlay computes an overlay operation for the
// given geometry operands, with the noding strategy determined by the precision
// model.
func OperationOverlayng_OverlayNG_Overlay(geom0, geom1 *Geom_Geometry, opCode int, pm *Geom_PrecisionModel) *Geom_Geometry {
	ov := OperationOverlayng_NewOverlayNGWithPM(geom0, geom1, pm, opCode)
	return ov.GetResult()
}

// OperationOverlayng_OverlayNG_OverlayWithNoder computes an overlay operation on
// the given geometry operands, using a supplied Noder.
func OperationOverlayng_OverlayNG_OverlayWithNoder(geom0, geom1 *Geom_Geometry, opCode int, pm *Geom_PrecisionModel, noder Noding_Noder) *Geom_Geometry {
	ov := OperationOverlayng_NewOverlayNGWithPM(geom0, geom1, pm, opCode)
	ov.SetNoder(noder)
	return ov.GetResult()
}

// OperationOverlayng_OverlayNG_OverlayWithNoderOnly computes an overlay
// operation on the given geometry operands, using a supplied Noder.
func OperationOverlayng_OverlayNG_OverlayWithNoderOnly(geom0, geom1 *Geom_Geometry, opCode int, noder Noding_Noder) *Geom_Geometry {
	ov := OperationOverlayng_NewOverlayNGWithPM(geom0, geom1, nil, opCode)
	ov.SetNoder(noder)
	return ov.GetResult()
}

// OperationOverlayng_OverlayNG_OverlayDefault computes an overlay operation on
// the given geometry operands, using the precision model of the geometry and an
// appropriate noder.
//
// The noder is chosen according to the precision model specified.
//   - For FIXED a snap-rounding noder is used, and the computation is robust.
//   - For FLOATING a non-snapping noder is used, and this computation may not
//     be robust. If errors occur a TopologyException is thrown.
func OperationOverlayng_OverlayNG_OverlayDefault(geom0, geom1 *Geom_Geometry, opCode int) *Geom_Geometry {
	ov := OperationOverlayng_NewOverlayNG(geom0, geom1, opCode)
	return ov.GetResult()
}

// OperationOverlayng_OverlayNG_UnionGeom computes a union operation on the given
// geometry, with the supplied precision model.
//
// The input must be a valid geometry. Collections must be homogeneous.
//
// To union an overlapping set of polygons in a more performant way use
// UnaryUnionNG. To union a polygon coverage or linear network in a more
// performant way, use CoverageUnion.
func OperationOverlayng_OverlayNG_UnionGeom(geom *Geom_Geometry, pm *Geom_PrecisionModel) *Geom_Geometry {
	ov := OperationOverlayng_NewOverlayNGUnary(geom, pm)
	return ov.GetResult()
}

// OperationOverlayng_OverlayNG_UnionGeomWithNoder computes a union of a single
// geometry using a custom noder.
//
// The primary use of this is to support coverage union. Because of this the
// overlay is performed using strict mode.
func OperationOverlayng_OverlayNG_UnionGeomWithNoder(geom *Geom_Geometry, pm *Geom_PrecisionModel, noder Noding_Noder) *Geom_Geometry {
	ov := OperationOverlayng_NewOverlayNGUnary(geom, pm)
	ov.SetNoder(noder)
	ov.SetStrictMode(true)
	return ov.GetResult()
}

// SetStrictMode sets whether the overlay results are computed according to
// strict mode semantics.
//   - Lines resulting from topology collapse are not included
//   - Result geometry is homogeneous for the INTERSECTION and DIFFERENCE
//     operations.
//   - Result geometry is homogeneous for the UNION and SYMDIFFERENCE operations
//     if the inputs have the same dimension
func (ov *OperationOverlayng_OverlayNG) SetStrictMode(isStrictMode bool) {
	ov.isStrictMode = isStrictMode
}

// SetOptimized sets whether overlay processing optimizations are enabled. It
// may be useful to disable optimizations for testing purposes. Default is TRUE
// (optimization enabled).
func (ov *OperationOverlayng_OverlayNG) SetOptimized(isOptimized bool) {
	ov.isOptimized = isOptimized
}

// SetAreaResultOnly sets whether the result can contain only Polygon
// components. This is used if it is known that the result must be an (possibly
// empty) area.
func (ov *OperationOverlayng_OverlayNG) SetAreaResultOnly(isAreaResultOnly bool) {
	ov.isAreaResultOnly = isAreaResultOnly
}

// SetOutputEdges enables outputting edges (for testing).
func (ov *OperationOverlayng_OverlayNG) SetOutputEdges(isOutputEdges bool) {
	ov.isOutputEdges = isOutputEdges
}

// SetOutputNodedEdges enables outputting noded edges (for testing).
func (ov *OperationOverlayng_OverlayNG) SetOutputNodedEdges(isOutputNodedEdges bool) {
	ov.isOutputEdges = true
	ov.isOutputNodedEdges = isOutputNodedEdges
}

// SetOutputResultEdges enables outputting result edges (for testing).
func (ov *OperationOverlayng_OverlayNG) SetOutputResultEdges(isOutputResultEdges bool) {
	ov.isOutputResultEdges = isOutputResultEdges
}

// SetNoder sets the noder to use for computing the overlay.
func (ov *OperationOverlayng_OverlayNG) SetNoder(noder Noding_Noder) {
	ov.noder = noder
}

// GetResult gets the result of the overlay operation.
func (ov *OperationOverlayng_OverlayNG) GetResult() *Geom_Geometry {
	// Handle empty inputs which determine result.
	if OperationOverlayng_OverlayUtil_IsEmptyResult(ov.opCode,
		ov.inputGeom.GetGeometry(0),
		ov.inputGeom.GetGeometry(1),
		ov.pm) {
		return ov.createEmptyResult()
	}

	// The elevation model is only computed if the input geometries have Z values.
	elevModel := OperationOverlayng_ElevationModel_Create(ov.inputGeom.GetGeometry(0), ov.inputGeom.GetGeometry(1))
	var result *Geom_Geometry
	if ov.inputGeom.IsAllPoints() {
		// Handle Point-Point inputs.
		result = OperationOverlayng_OverlayPoints_Overlay(ov.opCode, ov.inputGeom.GetGeometry(0), ov.inputGeom.GetGeometry(1), ov.pm)
	} else if !ov.inputGeom.IsSingle() && ov.inputGeom.HasPoints() {
		// Handle Point-nonPoint inputs.
		result = OperationOverlayng_OverlayMixedPoints_Overlay(ov.opCode, ov.inputGeom.GetGeometry(0), ov.inputGeom.GetGeometry(1), ov.pm)
	} else {
		// Handle case where both inputs are formed of edges (Lines and Polygons).
		result = ov.computeEdgeOverlay()
	}
	// This is a no-op if the elevation model was not computed due to Z not present.
	elevModel.PopulateZ(result)
	return result
}

func (ov *OperationOverlayng_OverlayNG) computeEdgeOverlay() *Geom_Geometry {
	edges := ov.nodeEdges()

	graph := ov.buildGraph(edges)

	if ov.isOutputNodedEdges {
		return OperationOverlayng_OverlayUtil_ToLines(graph, ov.isOutputEdges, ov.geomFact)
	}

	ov.labelGraph(graph)

	if ov.isOutputEdges || ov.isOutputResultEdges {
		return OperationOverlayng_OverlayUtil_ToLines(graph, ov.isOutputEdges, ov.geomFact)
	}

	result := ov.extractResult(ov.opCode, graph)

	// Heuristic check on result area. Catches cases where noding causes vertex
	// to move and make topology graph area "invert".
	if OperationOverlayng_OverlayUtil_IsFloating(ov.pm) {
		isAreaConsistent := OperationOverlayng_OverlayUtil_IsResultAreaConsistent(ov.inputGeom.GetGeometry(0), ov.inputGeom.GetGeometry(1), ov.opCode, result)
		if !isAreaConsistent {
			panic(Geom_NewTopologyException("Result area inconsistent with overlay operation"))
		}
	}
	return result
}

func (ov *OperationOverlayng_OverlayNG) nodeEdges() []*OperationOverlayng_Edge {
	// Node the edges, using whatever noder is being used.
	nodingBuilder := OperationOverlayng_NewEdgeNodingBuilder(ov.pm, ov.noder)

	// Optimize Intersection and Difference by clipping to the result extent,
	// if enabled.
	if ov.isOptimized {
		clipEnv := OperationOverlayng_OverlayUtil_ClippingEnvelope(ov.opCode, ov.inputGeom, ov.pm)
		if clipEnv != nil {
			nodingBuilder.SetClipEnvelope(clipEnv)
		}
	}

	mergedEdges := nodingBuilder.Build(
		ov.inputGeom.GetGeometry(0),
		ov.inputGeom.GetGeometry(1))

	// Record if an input geometry has collapsed. This is used to avoid trying
	// to locate disconnected edges against a geometry which has collapsed
	// completely.
	ov.inputGeom.SetCollapsed(0, !nodingBuilder.HasEdgesFor(0))
	ov.inputGeom.SetCollapsed(1, !nodingBuilder.HasEdgesFor(1))

	return mergedEdges
}

func (ov *OperationOverlayng_OverlayNG) buildGraph(edges []*OperationOverlayng_Edge) *OperationOverlayng_OverlayGraph {
	graph := OperationOverlayng_NewOverlayGraph()
	for _, e := range edges {
		graph.AddEdge(e.GetCoordinates(), e.CreateLabel())
	}
	return graph
}

func (ov *OperationOverlayng_OverlayNG) labelGraph(graph *OperationOverlayng_OverlayGraph) {
	labeller := OperationOverlayng_NewOverlayLabeller(graph, ov.inputGeom)
	labeller.ComputeLabelling()
	labeller.MarkResultAreaEdges(ov.opCode)
	labeller.UnmarkDuplicateEdgesFromResultArea()
}

// extractResult extracts the result geometry components from the fully
// labelled topology graph.
//
// This method implements the semantic that the result of an intersection
// operation is homogeneous with highest dimension. In other words, if an
// intersection has components of a given dimension no lower-dimension
// components are output. For example, if two polygons intersect in an area, no
// linestrings or points are included in the result, even if portions of the
// input do meet in lines or points. This semantic choice makes more sense for
// typical usage, in which only the highest dimension components are of
// interest.
func (ov *OperationOverlayng_OverlayNG) extractResult(opCode int, graph *OperationOverlayng_OverlayGraph) *Geom_Geometry {
	isAllowMixedIntResult := !ov.isStrictMode

	// Build polygons.
	resultAreaEdges := graph.GetResultAreaEdges()
	polyBuilder := OperationOverlayng_NewPolygonBuilder(resultAreaEdges, ov.geomFact)
	resultPolyList := polyBuilder.GetPolygons()
	hasResultAreaComponents := len(resultPolyList) > 0

	var resultLineList []*Geom_LineString
	var resultPointList []*Geom_Point

	if !ov.isAreaResultOnly {
		// Build lines.
		allowResultLines := !hasResultAreaComponents ||
			isAllowMixedIntResult ||
			opCode == OperationOverlayng_OverlayNG_SYMDIFFERENCE ||
			opCode == OperationOverlayng_OverlayNG_UNION
		if allowResultLines {
			lineBuilder := OperationOverlayng_NewLineBuilder(ov.inputGeom, graph, hasResultAreaComponents, opCode, ov.geomFact)
			lineBuilder.SetStrictMode(ov.isStrictMode)
			resultLineList = lineBuilder.GetLines()
		}
		// Operations with point inputs are handled elsewhere. Only an
		// Intersection op can produce point results from non-point inputs.
		hasResultComponents := hasResultAreaComponents || len(resultLineList) > 0
		allowResultPoints := !hasResultComponents || isAllowMixedIntResult
		if opCode == OperationOverlayng_OverlayNG_INTERSECTION && allowResultPoints {
			pointBuilder := OperationOverlayng_NewIntersectionPointBuilder(graph, ov.geomFact)
			pointBuilder.SetStrictMode(ov.isStrictMode)
			resultPointList = pointBuilder.GetPoints()
		}
	}

	if operationOverlayng_OverlayNG_isEmpty(resultPolyList) &&
		operationOverlayng_OverlayNG_isEmptyLines(resultLineList) &&
		operationOverlayng_OverlayNG_isEmptyPoints(resultPointList) {
		return ov.createEmptyResult()
	}

	return OperationOverlayng_OverlayUtil_CreateResultGeometry(resultPolyList, resultLineList, resultPointList, ov.geomFact)
}

func operationOverlayng_OverlayNG_isEmpty(list []*Geom_Polygon) bool {
	return list == nil || len(list) == 0
}

func operationOverlayng_OverlayNG_isEmptyLines(list []*Geom_LineString) bool {
	return list == nil || len(list) == 0
}

func operationOverlayng_OverlayNG_isEmptyPoints(list []*Geom_Point) bool {
	return list == nil || len(list) == 0
}

func (ov *OperationOverlayng_OverlayNG) createEmptyResult() *Geom_Geometry {
	return OperationOverlayng_OverlayUtil_CreateEmptyResult(
		OperationOverlayng_OverlayUtil_ResultDimension(ov.opCode,
			ov.inputGeom.GetDimension(0),
			ov.inputGeom.GetDimension(1)),
		ov.geomFact)
}

// OperationOverlayng_OverlayNG_IsResultOfOpPoint tests whether a point with a
// given topological OverlayLabel relative to two geometries is contained in the
// result of overlaying the geometries using a given overlay operation.
//
// The method handles arguments of Location.NONE correctly.
func OperationOverlayng_OverlayNG_IsResultOfOpPoint(label *OperationOverlayng_OverlayLabel, opCode int) bool {
	loc0 := label.GetLocationIndex(0)
	loc1 := label.GetLocationIndex(1)
	return OperationOverlayng_OverlayNG_IsResultOfOp(opCode, loc0, loc1)
}

// OperationOverlayng_OverlayNG_Union computes a union operation on the given
// geometry, with the supplied precision model. This is an alias for
// UnionGeom for backwards compatibility.
func OperationOverlayng_OverlayNG_Union(geom *Geom_Geometry, pm *Geom_PrecisionModel) *Geom_Geometry {
	return OperationOverlayng_OverlayNG_UnionGeom(geom, pm)
}
