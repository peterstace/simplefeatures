package carto

import (
	"fmt"
	"strconv"

	"github.com/peterstace/simplefeatures/geom"
)

// UTM allows projecting (longitude, latitude) coordinates to (x, y) pairs via
// a UTM (Universal Transverse Mercator) projection.
//
// There are 60 UTM zones, each covering 6 degrees of longitude. There is a
// different set of projections for the northern and southern hemispheres,
// resulting in 120 distinct UTM projections in total.
//
// UTM projections are:
//   - Configured by a UTM zone and hemisphere designator.
//   - Conformal, but not equal area or equidistant. Despite not being equal
//     area or equidistant, area and distance are only degrade a small amount
//     within the bounds of a particular UTM zone.
type UTM struct {
	zone  int
	north bool // true for northern hemisphere, false for southern hemisphere.
}

// Code returns the UTM code for this projection, which is the zone (with leading
// zero to pad to 2 digits) and an N or S designator. E.g. "06N" and "56S".
func (u *UTM) Code() string {
	hemisphere := "N"
	if !u.north {
		hemisphere = "S"
	}
	return fmt.Sprintf("%02d%s", u.zone, hemisphere)
}

type invalidUTMCodeError struct {
	code string
}

func (e invalidUTMCodeError) Error() string {
	return fmt.Sprintf(
		"invalid UTM code '%s': must be 3 characters long, with the first "+
			"two being digits (01-60) and the last being 'N' or 'S'", e.code)
}

// NewUTMFromCode creates a [UTM] projection using the given UTM zone and
// hemisphere code. Codes are composed the zone (with leading zero to pad to 2
// digits) and an N or S designator. E.g. "06N" and "56S".
func NewUTMFromCode(code string) (*UTM, error) {
	if len(code) != 3 {
		return nil, invalidUTMCodeError{code}
	}

	zone, err := strconv.Atoi(code[:2])
	if err != nil {
		return nil, invalidUTMCodeError{code}
	}
	if zone < 1 || zone > 60 {
		return nil, invalidUTMCodeError{code}
	}

	switch hemisphere := code[2:3]; hemisphere {
	case "S":
		return &UTM{zone, false}, nil
	case "N":
		return &UTM{zone, true}, nil
	default:
		return nil, invalidUTMCodeError{code}
	}
}

// NewUTMFromLocation creates a [UTM] projection using the appropriate [UTM] zone
// and hemisphere for the given location.
func NewUTMFromLocation(lonlat geom.XY) (*UTM, error) {
	λ := lonlat.X
	φ := lonlat.Y
	if φ < -80 || φ > 84 {
		return nil, fmt.Errorf("latitude %v out of range (must be between -80 and 84, both inclusive)", φ)
	}
	if λ < -180 || λ > 180 {
		return nil, fmt.Errorf("longitude %v out of range (must be between -180 and 180, both inclusive)", λ)
	}

	// Norway exception:
	if λ >= 3 && λ <= 6 && φ >= 56 && φ <= 64 {
		return &UTM{32, true}, nil
	}

	// Svalbard exception:
	if φ >= 72 && λ >= 0 && λ < 42 {
		if λ < 9 {
			return &UTM{31, true}, nil
		}
		if λ < 21 {
			return &UTM{33, true}, nil
		}
		if λ < 33 {
			return &UTM{35, true}, nil
		}
		return &UTM{37, true}, nil
	}

	z := int((λ+180)/6) + 1
	if z > 60 {
		// Handle the edge case where λ is exactly 180 degrees.
		z = 60
	}
	return &UTM{z, φ >= 0}, nil
}

// Projection formulas taken from:
//
// Map projections: A working manual (Professional Paper 1395) by John P. Snyder.
//
// https://doi.org/10.3133/pp1395

func (u *UTM) centralMeridian() float64 {
	var (
		zone    = float64(u.zone)
		degrees = (zone-1)*6 - 180 + 3
		radians = dtor(degrees)
	)
	return radians
}

func (u *UTM) falseNorthing() float64 {
	if u.north {
		return 0
	}
	return 10e6
}

//nolint:revive,stylecheck // Underscores used to retain as close as possible Snyder's notation.
const (
	utm_a            = 6378137
	utm_k0           = 0.9996
	utm_e2           = 0.00669438
	utm_falseEasting = 500e3
)

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (easting, northing) pair expressed in meters.
func (u *UTM) Forward(lonlat geom.XY) geom.XY {
	var (
		λ  = dtor(lonlat.X)
		φ  = dtor(lonlat.Y)
		λ0 = u.centralMeridian()
		N0 = u.falseNorthing()
	)
	const (
		a  = utm_a
		k0 = utm_k0
		e2 = utm_e2
		E0 = utm_falseEasting
	)
	var (
		ê2 = e2 / (1 - e2) // Called e'² in Snyder.
		N  = a / sqrt(1-e2*sq(sin(φ)))
		T  = sq(tan(φ))
		C  = ê2 * sq(cos(φ))
		A  = (λ - λ0) * cos(φ)
	)
	var (
		e4 = e2 * e2
		e6 = e4 * e2
		J1 = (1 - e2/4 - 3*e4/64 - 5*e6/256) * φ
		J2 = (3*e2/8 + 3*e4/32 + 45*e6/1024) * sin(2*φ)
		J3 = (15*e4/256 + 45*e6/1024) * sin(4*φ)
		J4 = (35 * e6 / 3072) * sin(6*φ)
		M  = a * (J1 - J2 + J3 - J4)
	)
	var (
		A2 = A * A
		A3 = A2 * A
		A4 = A3 * A
		A5 = A4 * A
		A6 = A5 * A
		T2 = T * T
	)
	var (
		Q1 = (1 - T + C) * A3 / 6
		Q2 = (5 - 18*T + T2 + 72*C - 58*ê2) * A5 / 120
		x  = E0 + k0*N*(A+Q1+Q2)
	)
	var (
		Q3 = (5 - T + 9*C + 4*C*C) * A4 / 24
		Q4 = (61 - 58*T + T2 + 600*C - 330*ê2) * A6 / 720
		y  = N0 + k0*(M+N*tan(φ)*(A2/2+Q3+Q4))
	)
	return geom.XY{X: x, Y: y}
}

// Reverse converts a projected (easting, northing) pair expressed in meters to
// a (longitude, latitude) pair expressed in degrees.
func (u *UTM) Reverse(xy geom.XY) geom.XY {
	var (
		y = xy.Y - u.falseNorthing()
		x = xy.X - utm_falseEasting
	)

	const (
		a  = utm_a
		k0 = utm_k0
		e2 = utm_e2
		e4 = e2 * e2
		e6 = e4 * e2
	)
	var (
		ε = (1 - sqrt(1-e2)) / (1 + sqrt(1-e2)) // Called e₁ in Snyder.
		M = y / k0
		μ = M / (a * (1 - e2/4 - 3*e4/64 - 5*e6/256))
	)

	var (
		ε2 = ε * ε
		ε3 = ε2 * ε
		ε4 = ε3 * ε
		J1 = 3*ε/2 - 27*ε3/32
		J2 = 21*ε2/16 - 55*ε4/32
		J3 = 151 * ε3 / 96
		J4 = 1097 * ε4 / 512
		φ0 = μ + J1*sin(2*μ) + J2*sin(4*μ) + J3*sin(6*μ) + J4*sin(8*μ)
	)

	var (
		ê2 = e2 / (1 - e2) // Called e'² in Snyder.
		C1 = ê2 * sq(cos(φ0))
		T1 = sq(tan(φ0))
		N1 = a / sqrt(1-e2*sq(sin(φ0)))
		R1 = a * (1 - e2) / pow(1-e2*sq(sin(φ0)), 1.5)
		D  = x / (N1 * k0)
	)
	var (
		D2 = D * D
		D3 = D2 * D
		D4 = D3 * D
		D5 = D4 * D
		D6 = D5 * D
	)
	var (
		Q1 = N1 * tan(φ0) / R1
		Q2 = (5 + 3*T1 + 10*C1 - 4*C1*C1 - 9*ê2) * D4 / 24
		Q3 = (61 + 90*T1 + 298*C1 + 45*T1*T1 - 3*C1*C1 - 252*ê2) * D6 / 720
		φ  = φ0 - Q1*(D2/2-Q2+Q3)
	)
	var (
		Q4 = (1 + 2*T1 + C1) * D3 / 6
		Q5 = (5 - 2*C1 + 28*T1 - 3*C1*C1 + 8*ê2 + 24*T1*T1) * D5 / 120
		λ0 = u.centralMeridian()
		λ  = λ0 + (D-Q4+Q5)/cos(φ0)
	)

	return rtodxy(λ, φ)
}
