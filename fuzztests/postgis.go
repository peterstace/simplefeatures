package main

import (
	"database/sql"
	"strings"
	"testing"
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
