package jts

// OperationOverlayng_OverlayMixedPoints computes an overlay where one input is
// Point(s) and one is not. This class supports overlay being used as an
// efficient way to find points within or outside a polygon.
type OperationOverlayng_OverlayMixedPoints struct {
	opCode            int
	pm                *Geom_PrecisionModel
	geomPoint         *Geom_Geometry
	geomNonPointInput *Geom_Geometry
	geometryFactory   *Geom_GeometryFactory
	isPointRHS        bool
	geomNonPoint      *Geom_Geometry
	geomNonPointDim   int
	locator           AlgorithmLocate_PointOnGeometryLocator
	resultDim         int
}

// OperationOverlayng_OverlayMixedPoints_Overlay performs the overlay operation
// on mixed point/non-point geometries.
func OperationOverlayng_OverlayMixedPoints_Overlay(opCode int, geom0, geom1 *Geom_Geometry, pm *Geom_PrecisionModel) *Geom_Geometry {
	overlay := OperationOverlayng_NewOverlayMixedPoints(opCode, geom0, geom1, pm)
	return overlay.GetResult()
}

// OperationOverlayng_NewOverlayMixedPoints creates a new OverlayMixedPoints.
func OperationOverlayng_NewOverlayMixedPoints(opCode int, geom0, geom1 *Geom_Geometry, pm *Geom_PrecisionModel) *OperationOverlayng_OverlayMixedPoints {
	omp := &OperationOverlayng_OverlayMixedPoints{
		opCode:          opCode,
		pm:              pm,
		geometryFactory: geom0.GetFactory(),
		resultDim:       OperationOverlayng_OverlayUtil_ResultDimension(opCode, geom0.GetDimension(), geom1.GetDimension()),
	}

	// Name the dimensional geometries.
	if geom0.GetDimension() == 0 {
		omp.geomPoint = geom0
		omp.geomNonPointInput = geom1
		omp.isPointRHS = false
	} else {
		omp.geomPoint = geom1
		omp.geomNonPointInput = geom0
		omp.isPointRHS = true
	}
	return omp
}

// GetResult returns the result of the overlay operation.
func (omp *OperationOverlayng_OverlayMixedPoints) GetResult() *Geom_Geometry {
	// Reduce precision of non-point input, if required.
	omp.geomNonPoint = omp.prepareNonPoint(omp.geomNonPointInput)
	omp.geomNonPointDim = omp.geomNonPoint.GetDimension()
	omp.locator = omp.createLocator(omp.geomNonPoint)

	coords := operationOverlayng_OverlayMixedPoints_extractCoordinates(omp.geomPoint, omp.pm)

	switch omp.opCode {
	case OperationOverlayng_OverlayNG_INTERSECTION:
		return omp.computeIntersection(coords)
	case OperationOverlayng_OverlayNG_UNION, OperationOverlayng_OverlayNG_SYMDIFFERENCE:
		// UNION and SYMDIFFERENCE have same output.
		return omp.computeUnion(coords)
	case OperationOverlayng_OverlayNG_DIFFERENCE:
		return omp.computeDifference(coords)
	}
	Util_Assert_ShouldNeverReachHereWithMessage("Unknown overlay op code")
	return nil
}

func (omp *OperationOverlayng_OverlayMixedPoints) createLocator(geomNonPoint *Geom_Geometry) AlgorithmLocate_PointOnGeometryLocator {
	if omp.geomNonPointDim == 2 {
		return AlgorithmLocate_NewIndexedPointInAreaLocator(geomNonPoint)
	}
	return OperationOverlayng_NewIndexedPointOnLineLocator(geomNonPoint)
}

func (omp *OperationOverlayng_OverlayMixedPoints) prepareNonPoint(geomInput *Geom_Geometry) *Geom_Geometry {
	// If non-point not in output no need to node it.
	if omp.resultDim == 0 {
		return geomInput
	}

	// Node and round the non-point geometry for output.
	// NOTE: This calls OverlayNG.union which is stubbed for now.
	return OperationOverlayng_OverlayNG_Union(omp.geomNonPointInput, omp.pm)
}

func (omp *OperationOverlayng_OverlayMixedPoints) computeIntersection(coords []*Geom_Coordinate) *Geom_Geometry {
	return omp.createPointResult(omp.findPoints(true, coords))
}

func (omp *OperationOverlayng_OverlayMixedPoints) computeUnion(coords []*Geom_Coordinate) *Geom_Geometry {
	resultPointList := omp.findPoints(false, coords)
	var resultLineList []*Geom_LineString
	if omp.geomNonPointDim == 1 {
		resultLineList = operationOverlayng_OverlayMixedPoints_extractLines(omp.geomNonPoint)
	}
	var resultPolyList []*Geom_Polygon
	if omp.geomNonPointDim == 2 {
		resultPolyList = operationOverlayng_OverlayMixedPoints_extractPolygons(omp.geomNonPoint)
	}

	return OperationOverlayng_OverlayUtil_CreateResultGeometry(resultPolyList, resultLineList, resultPointList, omp.geometryFactory)
}

func (omp *OperationOverlayng_OverlayMixedPoints) computeDifference(coords []*Geom_Coordinate) *Geom_Geometry {
	if omp.isPointRHS {
		return omp.copyNonPoint()
	}
	return omp.createPointResult(omp.findPoints(false, coords))
}

func (omp *OperationOverlayng_OverlayMixedPoints) createPointResult(points []*Geom_Point) *Geom_Geometry {
	if len(points) == 0 {
		return omp.geometryFactory.CreateEmpty(0)
	} else if len(points) == 1 {
		return points[0].Geom_Geometry
	}
	return omp.geometryFactory.CreateMultiPointFromPoints(points).Geom_Geometry
}

func (omp *OperationOverlayng_OverlayMixedPoints) findPoints(isCovered bool, coords []*Geom_Coordinate) []*Geom_Point {
	resultCoords := make(map[string]*Geom_Coordinate)
	for _, coord := range coords {
		if omp.hasLocation(isCovered, coord) {
			// Copy coordinate to avoid aliasing.
			key := coord.String()
			if _, exists := resultCoords[key]; !exists {
				resultCoords[key] = coord.Copy()
			}
		}
	}
	return omp.createPoints(resultCoords)
}

func (omp *OperationOverlayng_OverlayMixedPoints) createPoints(coords map[string]*Geom_Coordinate) []*Geom_Point {
	points := make([]*Geom_Point, 0, len(coords))
	for _, coord := range coords {
		point := omp.geometryFactory.CreatePointFromCoordinate(coord)
		points = append(points, point)
	}
	return points
}

func (omp *OperationOverlayng_OverlayMixedPoints) hasLocation(isCovered bool, coord *Geom_Coordinate) bool {
	isExterior := Geom_Location_Exterior == omp.locator.Locate(coord)
	if isCovered {
		return !isExterior
	}
	return isExterior
}

// copyNonPoint copies the non-point input geometry if not already done by
// precision reduction process.
func (omp *OperationOverlayng_OverlayMixedPoints) copyNonPoint() *Geom_Geometry {
	if omp.geomNonPointInput != omp.geomNonPoint {
		return omp.geomNonPoint
	}
	return omp.geomNonPoint.Copy()
}

func operationOverlayng_OverlayMixedPoints_extractCoordinates(points *Geom_Geometry, pm *Geom_PrecisionModel) []*Geom_Coordinate {
	coords := Geom_NewCoordinateList()
	filter := newExtractCoordinatesFilter(coords, pm)
	points.ApplyCoordinateFilter(filter)
	return coords.ToCoordinateArray()
}

type extractCoordinatesFilter struct {
	coords *Geom_CoordinateList
	pm     *Geom_PrecisionModel
}

var _ Geom_CoordinateFilter = (*extractCoordinatesFilter)(nil)

func (f *extractCoordinatesFilter) IsGeom_CoordinateFilter() {}

func newExtractCoordinatesFilter(coords *Geom_CoordinateList, pm *Geom_PrecisionModel) *extractCoordinatesFilter {
	return &extractCoordinatesFilter{
		coords: coords,
		pm:     pm,
	}
}

func (f *extractCoordinatesFilter) Filter(coord *Geom_Coordinate) {
	p := OperationOverlayng_OverlayUtil_Round(coord, f.pm)
	f.coords.AddCoordinate(p, false)
}

func operationOverlayng_OverlayMixedPoints_extractPolygons(geom *Geom_Geometry) []*Geom_Polygon {
	list := make([]*Geom_Polygon, 0)
	for i := 0; i < geom.GetNumGeometries(); i++ {
		g := geom.GetGeometryN(i)
		if poly, ok := g.GetChild().(*Geom_Polygon); ok {
			if !poly.IsEmpty() {
				list = append(list, poly)
			}
		}
	}
	return list
}

func operationOverlayng_OverlayMixedPoints_extractLines(geom *Geom_Geometry) []*Geom_LineString {
	list := make([]*Geom_LineString, 0)
	for i := 0; i < geom.GetNumGeometries(); i++ {
		g := geom.GetGeometryN(i)
		if line, ok := g.GetChild().(*Geom_LineString); ok {
			if !line.IsEmpty() {
				list = append(list, line)
			}
		}
	}
	return list
}
