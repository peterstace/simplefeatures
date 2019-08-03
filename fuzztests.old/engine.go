package main

import (
	"bytes"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
)

// GeometryEngine describes all of the operations that a geometry engine should
// be able to perform. This is an abstraction such that the simplefeatures
// library and PostGIS can be used via the same interface.
type GeometryEngine interface {
	ValidateWKT(wkt string) error
	ValidateWKB(hex string) error
	ValidateGeoJSON(geojson string) error

	AsText(geom.Geometry) (string, error)
}

type SimpleFeaturesEngine struct{}

func (SimpleFeaturesEngine) ValidateWKT(wkt string) error {
	_, err := geom.UnmarshalWKT(strings.NewReader(wkt))
	return err
}

func hexStringToBytes(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errors.New("hex string must have even length")
	}
	var buf []byte
	for i := 0; i < len(s); i += 2 {
		x, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return nil, err
		}
		buf = append(buf, byte(x))
	}
	return buf, nil
}

func (SimpleFeaturesEngine) ValidateWKB(hex string) error {
	hexBytes, err := hexStringToBytes(hex)
	if err != nil {
		return err
	}
	_, err = geom.UnmarshalWKB(bytes.NewReader(hexBytes))
	return err
}

func (SimpleFeaturesEngine) ValidateGeoJSON(geojson string) error {
	_, err := geom.UnmarshalGeoJSON([]byte(geojson))
	return err
}

func (SimpleFeaturesEngine) AsText(g geom.Geometry) (string, error) {
	return g.AsText(), nil
}

type PostgisEngine struct {
	db *sql.DB
}

func (p *PostgisEngine) validate(sqlSnippet, geometry string) error {
	var isValid bool
	if err := p.db.QueryRow(`
		SELECT ST_IsValid(`+sqlSnippet+`)`,
		geometry,
	).Scan(&isValid); err != nil {
		return err
	}
	if isValid {
		return nil
	}
	var reason string
	if err := p.db.QueryRow(`
		SELECT ST_IsValidReason(`+sqlSnippet+`)`,
		geometry,
	).Scan(&reason); err != nil {
		return err
	}
	return errors.New(reason)
}

func (p *PostgisEngine) ValidateWKT(wkt string) error {
	// The simple feature library accepts LINEARRING WKTs. However, postgis
	// doesn't accept them. A workaround for this is to just substitute
	// LINEARRING for LINESTRING. However, this will give a false negative if
	// the corpus contains a LINEARRING WKT that isn't closed (and thus won't
	// be accepted by simple features).
	wkt = strings.ReplaceAll(wkt, "LINEARRING", "LINESTRING")
	return p.validate(`ST_GeomFromText($1)`, wkt)
}

func (p *PostgisEngine) ValidateWKB(wkb string) error {
	return p.validate(`ST_GeomFromWKB(decode($1, 'hex'))`, wkb)
}

func (p *PostgisEngine) ValidateGeoJSON(geojson string) error {
	return p.validate(`ST_GeomFromGeoJSON($1)`, geojson)
}

func (p *PostgisEngine) AsText(g geom.Geometry) (string, error) {
	var wkt string
	err := p.db.QueryRow(`SELECT ST_AsText($1)`, g).Scan(&wkt)
	return wkt, err
}
