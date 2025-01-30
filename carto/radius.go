package carto

const (
	// WGS84EllipsoidEquatorialRadiusM is the radius of the WGS84 ellipsoid at
	// the equator.
	WGS84EllipsoidEquatorialRadiusM = 6378137.0

	// WGS84EllipsoidPolarRadiusM is the radius of the WGS84 ellipsoid at the poles.
	WGS84EllipsoidPolarRadiusM = 6356752.314245

	// WGS84EllipsoidMeanRadiusM is the mean radius of the WGS84 ellipsoid.
	WGS84EllipsoidMeanRadiusM = (2*WGS84EllipsoidEquatorialRadiusM + WGS84EllipsoidPolarRadiusM) / 3
)
