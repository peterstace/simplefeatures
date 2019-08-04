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
	return p.isValidWithReason(t, `ST_GeomFromText($1)`, wkt)
}

func (p PostGIS) isValidWithReason(t *testing.T, sqlSnippet, geometry string) (bool, string) {
	var isValid bool
	err := p.db.QueryRow(`SELECT ST_IsValid(`+sqlSnippet+`)`, geometry).Scan(&isValid)
	if err != nil && strings.Contains(err.Error(), "parse error") {
		isValid = false
		err = nil
	}
	if err != nil {
		t.Fatalf("postgis error: %v", err)
	}

	var reason string
	err = p.db.QueryRow(`SELECT ST_IsValidReason(`+sqlSnippet+`)`, geometry).Scan(&reason)
	if err != nil && strings.Contains(err.Error(), "parse error") {
		reason = err.Error()
		err = nil
	}
	if err != nil {
		t.Fatalf("postgis error: %v", err)
	}
	return isValid, reason
}
