package jts

// OperationBufferValidate_BufferCurveMaximumDistanceFinder finds the approximate maximum distance
// from a buffer curve to the originating geometry.
// This is similar to the Discrete Oriented Hausdorff distance
// from the buffer curve to the input.
//
// The approximate maximum distance is determined by testing
// all vertices in the buffer curve, as well
// as midpoints of the curve segments.
// Due to the way buffer curves are constructed, this
// should be a very close approximation.
type OperationBufferValidate_BufferCurveMaximumDistanceFinder struct {
	inputGeom *Geom_Geometry
	maxPtDist *OperationBufferValidate_PointPairDistance
}

// OperationBufferValidate_NewBufferCurveMaximumDistanceFinder creates a new BufferCurveMaximumDistanceFinder.
func OperationBufferValidate_NewBufferCurveMaximumDistanceFinder(inputGeom *Geom_Geometry) *OperationBufferValidate_BufferCurveMaximumDistanceFinder {
	return &OperationBufferValidate_BufferCurveMaximumDistanceFinder{
		inputGeom: inputGeom,
		maxPtDist: OperationBufferValidate_NewPointPairDistance(),
	}
}

// FindDistance finds the maximum distance from the buffer curve to the input geometry.
func (f *OperationBufferValidate_BufferCurveMaximumDistanceFinder) FindDistance(bufferCurve *Geom_Geometry) float64 {
	f.computeMaxVertexDistance(bufferCurve)
	f.computeMaxMidpointDistance(bufferCurve)
	return f.maxPtDist.GetDistance()
}

// GetDistancePoints gets the point pair containing the points that have the computed distance.
func (f *OperationBufferValidate_BufferCurveMaximumDistanceFinder) GetDistancePoints() *OperationBufferValidate_PointPairDistance {
	return f.maxPtDist
}

func (f *OperationBufferValidate_BufferCurveMaximumDistanceFinder) computeMaxVertexDistance(curve *Geom_Geometry) {
	distFilter := operationBufferValidate_newMaxPointDistanceFilter(f.inputGeom)
	curve.ApplyCoordinateFilter(distFilter)
	f.maxPtDist.SetMaximumFromPointPairDistance(distFilter.getMaxPointDistance())
}

func (f *OperationBufferValidate_BufferCurveMaximumDistanceFinder) computeMaxMidpointDistance(curve *Geom_Geometry) {
	distFilter := operationBufferValidate_newMaxMidpointDistanceFilter(f.inputGeom)
	curve.ApplyCoordinateSequenceFilter(distFilter)
	f.maxPtDist.SetMaximumFromPointPairDistance(distFilter.getMaxPointDistance())
}

// operationBufferValidate_MaxPointDistanceFilter is a filter to compute the maximum distance
// from all vertices of a geometry to another geometry.
type operationBufferValidate_MaxPointDistanceFilter struct {
	maxPtDist *OperationBufferValidate_PointPairDistance
	minPtDist *OperationBufferValidate_PointPairDistance
	geom      *Geom_Geometry
}

var _ Geom_CoordinateFilter = (*operationBufferValidate_MaxPointDistanceFilter)(nil)

func operationBufferValidate_newMaxPointDistanceFilter(geom *Geom_Geometry) *operationBufferValidate_MaxPointDistanceFilter {
	return &operationBufferValidate_MaxPointDistanceFilter{
		maxPtDist: OperationBufferValidate_NewPointPairDistance(),
		minPtDist: OperationBufferValidate_NewPointPairDistance(),
		geom:      geom,
	}
}

func (f *operationBufferValidate_MaxPointDistanceFilter) IsGeom_CoordinateFilter() {}

func (f *operationBufferValidate_MaxPointDistanceFilter) Filter(pt *Geom_Coordinate) {
	f.minPtDist.Initialize()
	OperationBufferValidate_DistanceToPointFinder_ComputeDistanceGeometry(f.geom, pt, f.minPtDist)
	f.maxPtDist.SetMaximumFromPointPairDistance(f.minPtDist)
}

func (f *operationBufferValidate_MaxPointDistanceFilter) getMaxPointDistance() *OperationBufferValidate_PointPairDistance {
	return f.maxPtDist
}

// operationBufferValidate_MaxMidpointDistanceFilter is a filter to compute the maximum distance
// from segment midpoints of a geometry to another geometry.
type operationBufferValidate_MaxMidpointDistanceFilter struct {
	maxPtDist *OperationBufferValidate_PointPairDistance
	minPtDist *OperationBufferValidate_PointPairDistance
	geom      *Geom_Geometry
}

var _ Geom_CoordinateSequenceFilter = (*operationBufferValidate_MaxMidpointDistanceFilter)(nil)

func operationBufferValidate_newMaxMidpointDistanceFilter(geom *Geom_Geometry) *operationBufferValidate_MaxMidpointDistanceFilter {
	return &operationBufferValidate_MaxMidpointDistanceFilter{
		maxPtDist: OperationBufferValidate_NewPointPairDistance(),
		minPtDist: OperationBufferValidate_NewPointPairDistance(),
		geom:      geom,
	}
}

func (f *operationBufferValidate_MaxMidpointDistanceFilter) IsGeom_CoordinateSequenceFilter() {}

func (f *operationBufferValidate_MaxMidpointDistanceFilter) Filter(seq Geom_CoordinateSequence, index int) {
	if index == 0 {
		return
	}

	p0 := seq.GetCoordinate(index - 1)
	p1 := seq.GetCoordinate(index)
	midPt := Geom_NewCoordinateWithXY(
		(p0.X+p1.X)/2,
		(p0.Y+p1.Y)/2)

	f.minPtDist.Initialize()
	OperationBufferValidate_DistanceToPointFinder_ComputeDistanceGeometry(f.geom, midPt, f.minPtDist)
	f.maxPtDist.SetMaximumFromPointPairDistance(f.minPtDist)
}

func (f *operationBufferValidate_MaxMidpointDistanceFilter) IsGeometryChanged() bool {
	return false
}

func (f *operationBufferValidate_MaxMidpointDistanceFilter) IsDone() bool {
	return false
}

func (f *operationBufferValidate_MaxMidpointDistanceFilter) getMaxPointDistance() *OperationBufferValidate_PointPairDistance {
	return f.maxPtDist
}
