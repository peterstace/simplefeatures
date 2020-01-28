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
	AsGeoJSON  []byte
	IsEmpty    bool
	Dimension  int
	Envelope   geom.Geometry
	IsSimple   bool
	Boundary   geom.Geometry
	ConvexHull geom.Geometry
	IsValid    bool
	IsRing     bool
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
		ST_AsGeoJSON(ST_GeomFromWKB($1)),
		ST_IsEmpty(ST_GeomFromWKB($1)),
		ST_Dimension(ST_GeomFromWKB($1)),
		ST_AsBinary(ST_Envelope(ST_GeomFromWKB($1))),
		ST_IsSimple(ST_GeomFromWKB($1)),
		ST_AsBinary(ST_Boundary(ST_GeomFromWKB($1))),
		ST_AsBinary(ST_ConvexHull(ST_GeomFromWKB($1))),
		ST_IsValid(ST_GeomFromWKB($1)),
		ST_GeometryType(ST_GeomFromWKB($1)) = 'ST_LineString' AND ST_IsRing(ST_GeomFromWKB($1)), -- ST_IsRing returns an error whenever it gets anything other than an ST_LineString
		ST_Length(ST_GeomFromWKB($1)),
		ST_Area(ST_GeomFromWKB($1)),
		ST_AsBinary(ST_Centroid(ST_GeomFromWKB($1))),
		ST_AsBinary(ST_Reverse(ST_GeomFromWKB($1)))
		`, g,
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
