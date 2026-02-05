package jts

import "math"

// OperationBufferValidate_PointPairDistance contains a pair of points and the distance between them.
// Provides methods to update with a new point pair with
// either maximum or minimum distance.
type OperationBufferValidate_PointPairDistance struct {
	pt       [2]*Geom_Coordinate
	distance float64
	isNull   bool
}

// OperationBufferValidate_NewPointPairDistance creates a new PointPairDistance.
func OperationBufferValidate_NewPointPairDistance() *OperationBufferValidate_PointPairDistance {
	return &OperationBufferValidate_PointPairDistance{
		pt:       [2]*Geom_Coordinate{Geom_NewCoordinate(), Geom_NewCoordinate()},
		distance: math.NaN(),
		isNull:   true,
	}
}

// Initialize initializes this PointPairDistance.
func (ppd *OperationBufferValidate_PointPairDistance) Initialize() {
	ppd.isNull = true
}

// InitializeWithCoordinates initializes the points, computing the distance between them.
func (ppd *OperationBufferValidate_PointPairDistance) InitializeWithCoordinates(p0, p1 *Geom_Coordinate) {
	ppd.pt[0].SetCoordinate(p0)
	ppd.pt[1].SetCoordinate(p1)
	ppd.distance = p0.Distance(p1)
	ppd.isNull = false
}

// initializeWithCoordinatesAndDistance initializes the points, avoiding recomputing the distance.
func (ppd *OperationBufferValidate_PointPairDistance) initializeWithCoordinatesAndDistance(p0, p1 *Geom_Coordinate, distance float64) {
	ppd.pt[0].SetCoordinate(p0)
	ppd.pt[1].SetCoordinate(p1)
	ppd.distance = distance
	ppd.isNull = false
}

// GetDistance gets the distance between the paired points.
func (ppd *OperationBufferValidate_PointPairDistance) GetDistance() float64 {
	return ppd.distance
}

// GetCoordinates gets the paired points.
func (ppd *OperationBufferValidate_PointPairDistance) GetCoordinates() []*Geom_Coordinate {
	return ppd.pt[:]
}

// GetCoordinate gets one of the paired points.
func (ppd *OperationBufferValidate_PointPairDistance) GetCoordinate(i int) *Geom_Coordinate {
	return ppd.pt[i]
}

// SetMaximumFromPointPairDistance sets this to the maximum distance found.
func (ppd *OperationBufferValidate_PointPairDistance) SetMaximumFromPointPairDistance(ptDist *OperationBufferValidate_PointPairDistance) {
	ppd.SetMaximum(ptDist.pt[0], ptDist.pt[1])
}

// SetMaximum sets this to the maximum distance found.
func (ppd *OperationBufferValidate_PointPairDistance) SetMaximum(p0, p1 *Geom_Coordinate) {
	if ppd.isNull {
		ppd.InitializeWithCoordinates(p0, p1)
		return
	}
	dist := p0.Distance(p1)
	if dist > ppd.distance {
		ppd.initializeWithCoordinatesAndDistance(p0, p1, dist)
	}
}

// SetMinimumFromPointPairDistance sets this to the minimum distance found.
func (ppd *OperationBufferValidate_PointPairDistance) SetMinimumFromPointPairDistance(ptDist *OperationBufferValidate_PointPairDistance) {
	ppd.SetMinimum(ptDist.pt[0], ptDist.pt[1])
}

// SetMinimum sets this to the minimum distance found.
func (ppd *OperationBufferValidate_PointPairDistance) SetMinimum(p0, p1 *Geom_Coordinate) {
	if ppd.isNull {
		ppd.InitializeWithCoordinates(p0, p1)
		return
	}
	dist := p0.Distance(p1)
	if dist < ppd.distance {
		ppd.initializeWithCoordinatesAndDistance(p0, p1, dist)
	}
}
