package main

import (
	"database/sql"

	"github.com/peterstace/simplefeatures/geom"
)

type PG struct {
	db *sql.DB
}

type UnaryResult struct {
	AsText     string
	AsBinary   []byte
	AsGeoJSON  sql.NullString
	IsEmpty    bool
	Dimension  int
	Envelope   geom.Geometry
	IsSimple   sql.NullBool
	Boundary   geom.NullGeometry
	ConvexHull geom.Geometry
	IsValid    bool
	IsRing     sql.NullBool
	Length     float64
	Area       float64
	Cetroid    geom.Geometry
	Reverse    geom.Geometry
}

func (p PG) Unary(g geom.Geometry) (UnaryResult, error) {
	var result UnaryResult
	err := p.db.QueryRow(`
		SELECT

		ST_AsText(ST_GeomFromWKB($1)),
		ST_AsBinary(ST_GeomFromWKB($1)),

		-- PostGIS cannot convert to geojson in the case where it has
		-- nested geometry collections. That seems to be based on the
		-- following section of https://tools.ietf.org/html/rfc7946:
		--
		-- To maximize interoperability, implementations SHOULD avoid
		-- nested GeometryCollections.  Furthermore, GeometryCollections
		-- composed of a single part or a number of parts of a single type
		-- SHOULD be avoided when that single part or a single object of
		-- multipart type (MultiPoint, MultiLineString, or MultiPolygon)
		-- could be used instead.
		CASE
			WHEN $2 -- nested GeometryCollection
			THEN NULL
			ELSE ST_AsGeoJSON(ST_GeomFromWKB($1))
		END,

		ST_IsEmpty(ST_GeomFromWKB($1)),
		ST_Dimension(ST_GeomFromWKB($1)),
		ST_AsBinary(ST_Envelope(ST_GeomFromWKB($1))),

		-- Simplicity is not defined for GeometryCollections.
		CASE
			WHEN ST_GeometryType(ST_GeomFromWKB($1)) = 'ST_GeometryCollection'
			THEN NULL
			ELSE ST_IsSimple(ST_GeomFromWKB($1))
		END,

		-- Boundary not defined for GeometryCollections.
		CASE
			WHEN ST_GeometryType(ST_GeomFromWKB($1)) = 'ST_GeometryCollection'
			THEN NULL
			ELSE ST_AsBinary(ST_Boundary(ST_GeomFromWKB($1)))
		END,

		ST_AsBinary(ST_ConvexHull(ST_GeomFromWKB($1))),
		ST_IsValid(ST_GeomFromWKB($1)),

		-- IsRing only defined for LineStrings.
		CASE
			WHEN ST_GeometryType(ST_GeomFromWKB($1)) != 'ST_LineString'
			THEN NULL
			ELSE ST_IsRing(ST_GeomFromWKB($1))
		END,

		ST_Length(ST_GeomFromWKB($1)),
		ST_Area(ST_GeomFromWKB($1)),
		ST_AsBinary(ST_Centroid(ST_GeomFromWKB($1))),
		ST_AsBinary(ST_Reverse(ST_GeomFromWKB($1)))
		`, g, isNestedGeometryCollection(g),
	).Scan(
		&result.AsText,
		&result.AsBinary,
		&result.AsGeoJSON,
		&result.IsEmpty,
		&result.Dimension,
		&result.Envelope,
		&result.IsSimple,
		&result.Boundary,
		&result.ConvexHull,
		&result.IsValid,
		&result.IsRing,
		&result.Length,
		&result.Area,
		&result.Cetroid,
		&result.Reverse,
	)
	return result, err
}

func isNestedGeometryCollection(g geom.Geometry) bool {
	if !g.IsGeometryCollection() {
		return false
	}
	gc := g.AsGeometryCollection()
	for i := 0; i < gc.NumGeometries(); i++ {
		if gc.GeometryN(i).IsGeometryCollection() {
			return true
		}
	}
	return false
}
