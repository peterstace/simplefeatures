package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationOverlayng_OverlayPoints performs an overlay operation on inputs
// which are both point geometries.
//
// Semantics are:
//   - Points are rounded to the precision model if provided
//   - Points with identical XY values are merged to a single point
//   - Extended ordinate values are preserved in the output, apart from merging
//   - An empty result is returned as POINT EMPTY
type OperationOverlayng_OverlayPoints struct {
	opCode          int
	geom0           *Geom_Geometry
	geom1           *Geom_Geometry
	pm              *Geom_PrecisionModel
	geometryFactory *Geom_GeometryFactory
	resultList      []*Geom_Point
}

// OperationOverlayng_OverlayPoints_Overlay performs an overlay operation on
// inputs which are both point geometries.
func OperationOverlayng_OverlayPoints_Overlay(opCode int, geom0, geom1 *Geom_Geometry, pm *Geom_PrecisionModel) *Geom_Geometry {
	overlay := OperationOverlayng_NewOverlayPoints(opCode, geom0, geom1, pm)
	return overlay.GetResult()
}

// OperationOverlayng_NewOverlayPoints creates an instance of an overlay
// operation on inputs which are both point geometries.
func OperationOverlayng_NewOverlayPoints(opCode int, geom0, geom1 *Geom_Geometry, pm *Geom_PrecisionModel) *OperationOverlayng_OverlayPoints {
	return &OperationOverlayng_OverlayPoints{
		opCode:          opCode,
		geom0:           geom0,
		geom1:           geom1,
		pm:              pm,
		geometryFactory: geom0.GetFactory(),
	}
}

// GetResult gets the result of the overlay.
func (op *OperationOverlayng_OverlayPoints) GetResult() *Geom_Geometry {
	map0 := op.buildPointMap(op.geom0)
	map1 := op.buildPointMap(op.geom1)

	op.resultList = make([]*Geom_Point, 0)
	switch op.opCode {
	case OperationOverlayng_OverlayNG_INTERSECTION:
		op.computeIntersection(map0, map1)
	case OperationOverlayng_OverlayNG_UNION:
		op.computeUnion(map0, map1)
	case OperationOverlayng_OverlayNG_DIFFERENCE:
		op.computeDifference(map0, map1)
	case OperationOverlayng_OverlayNG_SYMDIFFERENCE:
		op.computeDifference(map0, map1)
		op.computeDifference(map1, map0)
	}
	if len(op.resultList) == 0 {
		return OperationOverlayng_OverlayUtil_CreateEmptyResult(0, op.geometryFactory)
	}

	// Convert points to geometries for BuildGeometry.
	geomList := make([]*Geom_Geometry, len(op.resultList))
	for i, pt := range op.resultList {
		geomList[i] = pt.Geom_Geometry
	}
	return op.geometryFactory.BuildGeometry(geomList)
}

func (op *OperationOverlayng_OverlayPoints) computeIntersection(map0, map1 map[string]*Geom_Point) {
	for _, key := range java.SortedKeysString(map0) {
		pt := map0[key]
		if _, exists := map1[key]; exists {
			op.resultList = append(op.resultList, op.copyPoint(pt))
		}
	}
}

func (op *OperationOverlayng_OverlayPoints) computeDifference(map0, map1 map[string]*Geom_Point) {
	for _, key := range java.SortedKeysString(map0) {
		pt := map0[key]
		if _, exists := map1[key]; !exists {
			op.resultList = append(op.resultList, op.copyPoint(pt))
		}
	}
}

func (op *OperationOverlayng_OverlayPoints) computeUnion(map0, map1 map[string]*Geom_Point) {
	// Copy all A points.
	for _, key := range java.SortedKeysString(map0) {
		op.resultList = append(op.resultList, op.copyPoint(map0[key]))
	}

	for _, key := range java.SortedKeysString(map1) {
		pt := map1[key]
		if _, exists := map0[key]; !exists {
			op.resultList = append(op.resultList, op.copyPoint(pt))
		}
	}
}

func (op *OperationOverlayng_OverlayPoints) copyPoint(pt *Geom_Point) *Geom_Point {
	// If pm is floating, the point coordinate is not changed.
	if OperationOverlayng_OverlayUtil_IsFloating(op.pm) {
		copied := pt.Geom_Geometry.Copy()
		return copied.GetChild().(*Geom_Point)
	}

	// pm is fixed. Round off X&Y ordinates, copy other ordinates unchanged.
	seq := pt.GetCoordinateSequence()
	seq2 := seq.Copy()
	seq2.SetOrdinate(0, Geom_Coordinate_X, op.pm.MakePrecise(seq.GetX(0)))
	seq2.SetOrdinate(0, Geom_Coordinate_Y, op.pm.MakePrecise(seq.GetY(0)))
	return op.geometryFactory.CreatePointFromCoordinateSequence(seq2)
}

func (op *OperationOverlayng_OverlayPoints) buildPointMap(geoms *Geom_Geometry) map[string]*Geom_Point {
	pointMap := make(map[string]*Geom_Point)
	filter := newBuildPointMapFilter(pointMap, op.pm)
	geoms.Apply(filter)
	return pointMap
}

type buildPointMapFilter struct {
	pointMap map[string]*Geom_Point
	pm       *Geom_PrecisionModel
}

var _ Geom_GeometryComponentFilter = (*buildPointMapFilter)(nil)

func (f *buildPointMapFilter) IsGeom_GeometryComponentFilter() {}

func newBuildPointMapFilter(pointMap map[string]*Geom_Point, pm *Geom_PrecisionModel) *buildPointMapFilter {
	return &buildPointMapFilter{
		pointMap: pointMap,
		pm:       pm,
	}
}

func (f *buildPointMapFilter) Filter(geom *Geom_Geometry) {
	pt, ok := geom.GetChild().(*Geom_Point)
	if !ok {
		return
	}
	if pt.IsEmpty() {
		return
	}

	p := operationOverlayng_OverlayPoints_roundCoord(pt, f.pm)
	// Only add first occurrence of a point. This provides the merging semantics
	// of overlay.
	key := p.String()
	if _, exists := f.pointMap[key]; !exists {
		f.pointMap[key] = pt
	}
}

// roundCoord rounds the key point if precision model is fixed.
func operationOverlayng_OverlayPoints_roundCoord(pt *Geom_Point, pm *Geom_PrecisionModel) *Geom_Coordinate {
	p := pt.GetCoordinate()
	if OperationOverlayng_OverlayUtil_IsFloating(pm) {
		return p
	}
	p2 := p.Copy()
	pm.MakePreciseCoordinate(p2)
	return p2
}
