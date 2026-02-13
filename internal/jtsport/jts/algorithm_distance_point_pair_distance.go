package jts

import "math"

// AlgorithmDistance_PointPairDistance contains a pair of points and the distance between them.
// Provides methods to update with a new point pair with
// either maximum or minimum distance.
type AlgorithmDistance_PointPairDistance struct {
	pt       [2]*Geom_Coordinate
	distance float64
	isNull   bool
}

// AlgorithmDistance_NewPointPairDistance creates an instance of this class.
func AlgorithmDistance_NewPointPairDistance() *AlgorithmDistance_PointPairDistance {
	return &AlgorithmDistance_PointPairDistance{
		pt:       [2]*Geom_Coordinate{Geom_NewCoordinate(), Geom_NewCoordinate()},
		distance: math.NaN(),
		isNull:   true,
	}
}

// Initialize initializes this instance.
func (ppd *AlgorithmDistance_PointPairDistance) Initialize() {
	ppd.isNull = true
}

// InitializeWithCoordinates initializes the points, computing the distance between them.
func (ppd *AlgorithmDistance_PointPairDistance) InitializeWithCoordinates(p0, p1 *Geom_Coordinate) {
	ppd.InitializeWithCoordinatesAndDistance(p0, p1, p0.Distance(p1))
}

// InitializeWithCoordinatesAndDistance initializes the points, avoiding recomputing the distance.
func (ppd *AlgorithmDistance_PointPairDistance) InitializeWithCoordinatesAndDistance(p0, p1 *Geom_Coordinate, distance float64) {
	ppd.pt[0].SetCoordinate(p0)
	ppd.pt[1].SetCoordinate(p1)
	ppd.distance = distance
	ppd.isNull = false
}

// GetDistance gets the distance between the paired points.
func (ppd *AlgorithmDistance_PointPairDistance) GetDistance() float64 {
	return ppd.distance
}

// GetCoordinates gets the paired points.
func (ppd *AlgorithmDistance_PointPairDistance) GetCoordinates() []*Geom_Coordinate {
	return ppd.pt[:]
}

// GetCoordinate gets one of the paired points.
func (ppd *AlgorithmDistance_PointPairDistance) GetCoordinate(i int) *Geom_Coordinate {
	return ppd.pt[i]
}

// SetMaximumFromPointPairDistance sets this to the maximum distance found.
func (ppd *AlgorithmDistance_PointPairDistance) SetMaximumFromPointPairDistance(ptDist *AlgorithmDistance_PointPairDistance) {
	ppd.SetMaximum(ptDist.pt[0], ptDist.pt[1])
}

// SetMaximum sets this to the maximum distance found.
func (ppd *AlgorithmDistance_PointPairDistance) SetMaximum(p0, p1 *Geom_Coordinate) {
	if ppd.isNull {
		ppd.InitializeWithCoordinates(p0, p1)
		return
	}
	dist := p0.Distance(p1)
	if dist > ppd.distance {
		ppd.InitializeWithCoordinatesAndDistance(p0, p1, dist)
	}
}

// SetMinimumFromPointPairDistance sets this to the minimum distance found.
func (ppd *AlgorithmDistance_PointPairDistance) SetMinimumFromPointPairDistance(ptDist *AlgorithmDistance_PointPairDistance) {
	ppd.SetMinimum(ptDist.pt[0], ptDist.pt[1])
}

// SetMinimum sets this to the minimum distance found.
func (ppd *AlgorithmDistance_PointPairDistance) SetMinimum(p0, p1 *Geom_Coordinate) {
	if ppd.isNull {
		ppd.InitializeWithCoordinates(p0, p1)
		return
	}
	dist := p0.Distance(p1)
	if dist < ppd.distance {
		ppd.InitializeWithCoordinatesAndDistance(p0, p1, dist)
	}
}

// String returns a string representation of this PointPairDistance.
func (ppd *AlgorithmDistance_PointPairDistance) String() string {
	return Io_WKTWriter_ToLineStringFromTwoCoords(ppd.pt[0], ppd.pt[1])
}
