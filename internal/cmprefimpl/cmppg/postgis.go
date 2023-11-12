package main

import (
	"database/sql"
)

// PostGIS is a DB access type allowing non-batch based interactions
// with a PostGIS database.
type PostGIS struct {
	db *sql.DB
}

// WKTIsValidWithReason checks if a WKT is valid, and if not gives the reason.
func (p PostGIS) WKTIsValidWithReason(wkt string) (bool, string) {
	var isValid bool
	var reason string
	err := p.db.QueryRow(`
		SELECT
			ST_IsValid(ST_GeomFromText($1)),
			ST_IsValidReason(ST_GeomFromText($1))`,
		wkt,
	).Scan(&isValid, &reason)
	if err != nil {
		// It's not possible to distinguish between problems with the geometry
		// and problems with the database except by string-matching. It's
		// better to just report all errors, even if this means there will be
		// some false errors in the case of connectivity problems (or similar).
		return false, err.Error()
	}
	return isValid, reason
}

// WKBIsValidWithReason checks if a WKB is valid and, if not, gives the reason.
// It's unable to differentiate between errors due to problems with the WKB and
// general problems with the database.
func (p PostGIS) WKBIsValidWithReason(wkb string) (bool, string) {
	var isValid bool
	err := p.db.QueryRow(`SELECT ST_IsValid(ST_GeomFromWKB(decode($1, 'hex')))`, wkb).Scan(&isValid)
	if err != nil {
		return false, err.Error()
	}
	var reason string
	err = p.db.QueryRow(`SELECT ST_IsValidReason(ST_GeomFromWKB(decode($1, 'hex')))`, wkb).Scan(&reason)
	if err != nil {
		return false, err.Error()
	}
	return isValid, reason
}

// WKBIsValidWithReason checks if a GeoJSON value is valid and, if not, gives
// the reason.  It's unable to differentiate between errors due to problems
// with the GeoJSON value and general problems with the database.
func (p PostGIS) GeoJSONIsValidWithReason(geojson string) (bool, string) {
	var isValid bool
	err := p.db.QueryRow(`SELECT ST_IsValid(ST_GeomFromGeoJSON($1))`, geojson).Scan(&isValid)
	if err != nil {
		return false, err.Error()
	}
	var reason string
	err = p.db.QueryRow(`SELECT ST_IsValidReason(ST_GeomFromGeoJSON($1))`, geojson).Scan(&reason)
	if err != nil {
		return false, err.Error()
	}
	return isValid, reason
}
