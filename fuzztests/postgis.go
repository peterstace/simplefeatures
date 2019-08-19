package main

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

type PostGIS struct {
	db *sql.DB
}

func (p PostGIS) WKTIsValidWithReason(t *testing.T, wkt string) (bool, string) {
	var isValid bool
	err := p.db.QueryRow(`SELECT ST_IsValid(ST_GeomFromText($1))`, wkt).Scan(&isValid)
	if err != nil && strings.Contains(err.Error(), "parse error") {
		isValid = false
		err = nil
	}
	if err != nil {
		t.Fatalf("postgis error: %v", err)
	}

	var reason string
	err = p.db.QueryRow(`SELECT ST_IsValidReason(ST_GeomFromText($1))`, wkt).Scan(&reason)
	if err != nil && strings.Contains(err.Error(), "parse error") {
		reason = err.Error()
		err = nil
	}
	if err != nil {
		t.Fatalf("postgis error: %v", err)
	}
	return isValid, reason
}

func (p PostGIS) WKBIsValidWithReason(t *testing.T, wkb string) (bool, string) {
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

func (p PostGIS) GeoJSONIsValidWithReason(t *testing.T, geojson string) (bool, string) {
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

func (p PostGIS) AsText(t *testing.T, g geom.Geometry) string {
	var asText string
	if err := p.db.QueryRow(`SELECT ST_AsText($1)`, g).Scan(&asText); err != nil {
		t.Fatalf("pg error: %v", err)
	}
	return asText
}

func (p PostGIS) AsBinary(t *testing.T, g geom.Geometry) []byte {
	var asBinary []byte
	if err := p.db.QueryRow(`SELECT ST_AsBinary($1::geometry)`, g).Scan(&asBinary); err != nil {
		t.Fatalf("pg error: %v", err)
	}
	return asBinary
}

func (p PostGIS) AsGeoJSON(t *testing.T, g geom.Geometry) []byte {
	var geojson []byte
	if err := p.db.QueryRow(`SELECT ST_AsGeoJSON($1::geometry)`, g).Scan(&geojson); err != nil {
		t.Fatalf("pg error: %v", err)
	}
	return geojson
}

func (p PostGIS) IsEmpty(t *testing.T, g geom.Geometry) bool {
	var empty bool
	if err := p.db.QueryRow(`
		SELECT ST_IsEmpty(ST_GeomFromText($1))`, g,
	).Scan(&empty); err != nil {
		t.Fatalf("pg error: %v", err)
	}
	return empty
}
