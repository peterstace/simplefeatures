package jts

import "math"

// AlgorithmDistance_DiscreteHausdorffDistance is an algorithm for computing a distance metric
// which is an approximation to the Hausdorff Distance
// based on a discretization of the input Geometry.
// The algorithm computes the Hausdorff distance restricted to discrete points
// for one of the geometries.
// The points can be either the vertices of the geometries (the default),
// or the geometries with line segments densified by a given fraction.
// Also determines two points of the Geometries which are separated by the computed distance.
//
// This algorithm is an approximation to the standard Hausdorff distance.
// Specifically,
//
//	for all geometries a, b:    DHD(a, b) <= HD(a, b)
//
// The approximation can be made as close as needed by densifying the input geometries.
// In the limit, this value will approach the true Hausdorff distance:
//
//	DHD(A, B, densifyFactor) -> HD(A, B) as densifyFactor -> 0.0
//
// The default approximation is exact or close enough for a large subset of useful cases.
// Examples of these are:
//   - computing distance between Linestrings that are roughly parallel to each other,
//     and roughly equal in length. This occurs in matching linear networks.
//   - Testing similarity of geometries.
//
// An example where the default approximation is not close is:
//
//	A = LINESTRING (0 0, 100 0, 10 100, 10 100)
//	B = LINESTRING (0 100, 0 10, 80 10)
//
//	DHD(A, B) = 22.360679774997898
//	HD(A, B) ~= 47.8
type AlgorithmDistance_DiscreteHausdorffDistance struct {
	g0     *Geom_Geometry
	g1     *Geom_Geometry
	ptDist *AlgorithmDistance_PointPairDistance
	// Value of 0.0 indicates that no densification should take place
	densifyFrac float64
}

// AlgorithmDistance_DiscreteHausdorffDistance_Distance computes the Discrete Hausdorff Distance
// of two Geometries.
func AlgorithmDistance_DiscreteHausdorffDistance_Distance(g0, g1 *Geom_Geometry) float64 {
	dist := AlgorithmDistance_NewDiscreteHausdorffDistance(g0, g1)
	return dist.Distance()
}

// AlgorithmDistance_DiscreteHausdorffDistance_DistanceWithDensifyFrac computes the Discrete Hausdorff Distance
// of two Geometries with a densify fraction.
func AlgorithmDistance_DiscreteHausdorffDistance_DistanceWithDensifyFrac(g0, g1 *Geom_Geometry, densifyFrac float64) float64 {
	dist := AlgorithmDistance_NewDiscreteHausdorffDistance(g0, g1)
	dist.SetDensifyFraction(densifyFrac)
	return dist.Distance()
}

// AlgorithmDistance_NewDiscreteHausdorffDistance creates a new DiscreteHausdorffDistance.
func AlgorithmDistance_NewDiscreteHausdorffDistance(g0, g1 *Geom_Geometry) *AlgorithmDistance_DiscreteHausdorffDistance {
	return &AlgorithmDistance_DiscreteHausdorffDistance{
		g0:     g0,
		g1:     g1,
		ptDist: AlgorithmDistance_NewPointPairDistance(),
	}
}

// SetDensifyFraction sets the fraction by which to densify each segment.
// Each segment will be (virtually) split into a number of equal-length
// subsegments, whose fraction of the total length is closest
// to the given fraction.
func (dhd *AlgorithmDistance_DiscreteHausdorffDistance) SetDensifyFraction(densifyFrac float64) {
	if densifyFrac > 1.0 || densifyFrac <= 0.0 {
		panic("Fraction is not in range (0.0 - 1.0]")
	}
	dhd.densifyFrac = densifyFrac
}

// Distance computes the Discrete Hausdorff Distance.
func (dhd *AlgorithmDistance_DiscreteHausdorffDistance) Distance() float64 {
	dhd.compute(dhd.g0, dhd.g1)
	return dhd.ptDist.GetDistance()
}

// OrientedDistance computes the oriented Discrete Hausdorff Distance.
func (dhd *AlgorithmDistance_DiscreteHausdorffDistance) OrientedDistance() float64 {
	dhd.computeOrientedDistance(dhd.g0, dhd.g1, dhd.ptDist)
	return dhd.ptDist.GetDistance()
}

// GetCoordinates returns the coordinates of the points that are the computed distance apart.
func (dhd *AlgorithmDistance_DiscreteHausdorffDistance) GetCoordinates() []*Geom_Coordinate {
	return dhd.ptDist.GetCoordinates()
}

func (dhd *AlgorithmDistance_DiscreteHausdorffDistance) compute(g0, g1 *Geom_Geometry) {
	dhd.computeOrientedDistance(g0, g1, dhd.ptDist)
	dhd.computeOrientedDistance(g1, g0, dhd.ptDist)
}

func (dhd *AlgorithmDistance_DiscreteHausdorffDistance) computeOrientedDistance(discreteGeom, geom *Geom_Geometry, ptDist *AlgorithmDistance_PointPairDistance) {
	distFilter := algorithmDistance_newMaxPointDistanceFilter(geom)
	discreteGeom.ApplyCoordinateFilter(distFilter)
	ptDist.SetMaximumFromPointPairDistance(distFilter.getMaxPointDistance())

	if dhd.densifyFrac > 0 {
		fracFilter := algorithmDistance_newMaxDensifiedByFractionDistanceFilter(geom, dhd.densifyFrac)
		discreteGeom.ApplyCoordinateSequenceFilter(fracFilter)
		ptDist.SetMaximumFromPointPairDistance(fracFilter.getMaxPointDistance())
	}
}

// algorithmDistance_maxPointDistanceFilter is a filter to compute the maximum distance
// from all coordinates to a geometry.
type algorithmDistance_maxPointDistanceFilter struct {
	maxPtDist *AlgorithmDistance_PointPairDistance
	minPtDist *AlgorithmDistance_PointPairDistance
	geom      *Geom_Geometry
}

var _ Geom_CoordinateFilter = (*algorithmDistance_maxPointDistanceFilter)(nil)

func algorithmDistance_newMaxPointDistanceFilter(geom *Geom_Geometry) *algorithmDistance_maxPointDistanceFilter {
	return &algorithmDistance_maxPointDistanceFilter{
		maxPtDist: AlgorithmDistance_NewPointPairDistance(),
		minPtDist: AlgorithmDistance_NewPointPairDistance(),
		geom:      geom,
	}
}

func (f *algorithmDistance_maxPointDistanceFilter) IsGeom_CoordinateFilter() {}

func (f *algorithmDistance_maxPointDistanceFilter) Filter(pt *Geom_Coordinate) {
	f.minPtDist.Initialize()
	AlgorithmDistance_DistanceToPoint_ComputeDistanceGeometry(f.geom, pt, f.minPtDist)
	f.maxPtDist.SetMaximumFromPointPairDistance(f.minPtDist)
}

func (f *algorithmDistance_maxPointDistanceFilter) getMaxPointDistance() *AlgorithmDistance_PointPairDistance {
	return f.maxPtDist
}

// algorithmDistance_maxDensifiedByFractionDistanceFilter is a filter to compute the maximum distance
// from densified segments to a geometry.
type algorithmDistance_maxDensifiedByFractionDistanceFilter struct {
	maxPtDist  *AlgorithmDistance_PointPairDistance
	minPtDist  *AlgorithmDistance_PointPairDistance
	geom       *Geom_Geometry
	numSubSegs int
}

var _ Geom_CoordinateSequenceFilter = (*algorithmDistance_maxDensifiedByFractionDistanceFilter)(nil)

func algorithmDistance_newMaxDensifiedByFractionDistanceFilter(geom *Geom_Geometry, fraction float64) *algorithmDistance_maxDensifiedByFractionDistanceFilter {
	return &algorithmDistance_maxDensifiedByFractionDistanceFilter{
		maxPtDist:  AlgorithmDistance_NewPointPairDistance(),
		minPtDist:  AlgorithmDistance_NewPointPairDistance(),
		geom:       geom,
		numSubSegs: int(math.Round(1.0 / fraction)),
	}
}

func (f *algorithmDistance_maxDensifiedByFractionDistanceFilter) IsGeom_CoordinateSequenceFilter() {}

func (f *algorithmDistance_maxDensifiedByFractionDistanceFilter) Filter(seq Geom_CoordinateSequence, index int) {
	// This logic also handles skipping Point geometries
	if index == 0 {
		return
	}

	p0 := seq.GetCoordinate(index - 1)
	p1 := seq.GetCoordinate(index)

	delx := (p1.X - p0.X) / float64(f.numSubSegs)
	dely := (p1.Y - p0.Y) / float64(f.numSubSegs)

	for i := 0; i < f.numSubSegs; i++ {
		x := p0.X + float64(i)*delx
		y := p0.Y + float64(i)*dely
		pt := Geom_NewCoordinateWithXY(x, y)
		f.minPtDist.Initialize()
		AlgorithmDistance_DistanceToPoint_ComputeDistanceGeometry(f.geom, pt, f.minPtDist)
		f.maxPtDist.SetMaximumFromPointPairDistance(f.minPtDist)
	}
}

func (f *algorithmDistance_maxDensifiedByFractionDistanceFilter) IsGeometryChanged() bool {
	return false
}

func (f *algorithmDistance_maxDensifiedByFractionDistanceFilter) IsDone() bool {
	return false
}

func (f *algorithmDistance_maxDensifiedByFractionDistanceFilter) getMaxPointDistance() *AlgorithmDistance_PointPairDistance {
	return f.maxPtDist
}
