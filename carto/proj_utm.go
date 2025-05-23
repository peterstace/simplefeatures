package carto

import (
	"fmt"

	"github.com/peterstace/simplefeatures/geom"
)

type UTM struct {
	zone          int
	falseNorthing float64
}

// NewUTMFromLocation creates a UTM projection using the appropriate UTM zone
// and hemisphere for the given location.
func NewUTMFromLocation(lonlat geom.XY) (*UTM, error) {
	zoneAndHemi, err := lookupUTMZoneAndHemisphere(lonlat)
	if err != nil {
		return nil, err
	}
	var falseNorthing float64
	if !zoneAndHemi.north {
		falseNorthing = utm_falseNorthingForSouthernHemisphere
	}
	return &UTM{zoneAndHemi.utmZone, falseNorthing}, nil

}

// NewUTMFromEPSGCode creates a UTM projection using the given EPSG code for a
// WGS84 UTM Projection (codes 32601-32660 for northern hemisphere and
// 32701-32760 for southern hemisphere).
func NewUTMFromEPSGCode(code int) (*UTM, error) {
	// EPSG codes for UTM are in the range 32601 to 32660 for northern
	// hemisphere and 32701 to 32760 for southern hemisphere.
	code -= 32600 // Normalize to 1-60 and 101-160.
	if code >= 1 && code <= 60 {
		// Northern hemisphere.
		return &UTM{code, 0}, nil
	}
	if code >= 101 && code <= 160 {
		// Southern hemisphere.
		return &UTM{code - 100, utm_falseNorthingForSouthernHemisphere}, nil
	}
	return nil, fmt.Errorf("EPSG code %d not a WGS84 UTM Zone (must be in range [32601, 32660] or [32701, 32760])", code)
}

// NewUTMFromZoneAndHemisphere creates a UTM projection using the given UTM
// zone and hemisphere designator. The zone must be in the range 1-60 and the
// hemisphere must be "N" or "S".
func NewUTMFromZoneAndHemisphere(zone int, hemisphere string) (*UTM, error) {
	if zone < 1 || zone > 60 {
		return nil, fmt.Errorf("zone %d out of range (must be between 1 and 60, both inclusive)", zone)
	}
	switch hemisphere {
	case "S":
		return &UTM{zone, utm_falseNorthingForSouthernHemisphere}, nil
	case "N":
		return &UTM{zone, 0}, nil
	default:
		return nil, fmt.Errorf("hemisphere %q not recognized (must be either 'N' or 'S')", hemisphere)
	}
}

type utmZoneAndHemisphere struct {
	utmZone int
	north   bool
}

func lookupUTMZoneAndHemisphere(lonlat geom.XY) (utmZoneAndHemisphere, error) {
	λ := lonlat.X
	φ := lonlat.Y
	var zero utmZoneAndHemisphere
	if φ < -80 || φ > 84 {
		return zero, fmt.Errorf("latitude %v out of range (must be between -80 and 84, both inclusive)", φ)
	}
	if λ < -180 || λ > 180 {
		return zero, fmt.Errorf("longitude %v out of range (must be between -180 and 180, both inclusive)", λ)
	}

	// Norway exception:
	if λ >= 3 && λ < 12 && φ >= 56 && φ <= 64 {
		return utmZoneAndHemisphere{32, true}, nil
	}

	// Svalbard exception:
	if φ >= 72 && λ >= 0 && λ < 42 {
		if λ < 9 {
			return utmZoneAndHemisphere{31, true}, nil
		}
		if λ < 21 {
			return utmZoneAndHemisphere{33, true}, nil
		}
		if λ < 33 {
			return utmZoneAndHemisphere{35, true}, nil
		}
		return utmZoneAndHemisphere{37, true}, nil
	}

	z := int((λ + 180) / 6)
	return utmZoneAndHemisphere{z, φ >= 0}, nil
}

// Projection formulas taken from:
//
// Map projections: A working manual (Professional Paper 1395) by John P.
// Snyder.
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

const (
	utm_a  = 6378137
	utm_k0 = 0.9996
	utm_e2 = 0.00669438

	utm_falseEasting                       = 500e3
	utm_falseNorthingForSouthernHemisphere = 10e6
)

func (u *UTM) Forward(lonlat geom.XY) geom.XY {
	var (
		λ  = dtor(lonlat.X)
		φ  = dtor(lonlat.Y)
		λ0 = u.centralMeridian()
		N0 = u.falseNorthing
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

func (u *UTM) Recurse(xy geom.XY) geom.XY {
	var (
		y = xy.Y - u.falseNorthing
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
