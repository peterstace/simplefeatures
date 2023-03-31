package main

import (
	"database/sql"
	"math"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/peterstace/simplefeatures/geom"
)

// BatchPostGIS is a DB access type allowing batch based interactions with a
// PostGIS database.
type BatchPostGIS struct {
	db *sql.DB
}

// UnaryResult holds the result of unary (single input) PostGIS operations.
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
	IsRing     sql.NullBool
	Length     float64
	Area       float64
	Centroid   geom.Geometry
	Reverse    geom.Geometry
	Type       string
	Force2D    geom.Geometry
	Force3DZ   geom.Geometry
	Force3DM   geom.Geometry
	Force4D    geom.Geometry
	ForceCW    geom.Geometry
	ForceCCW   geom.Geometry
	Dump       []geom.Geometry
}

const (
	postgisTypePrefix = "ST_"
)

// Unary runs a batch of unary operations on a geometry.
func (p BatchPostGIS) Unary(g geom.Geometry) (UnaryResult, error) {
	// Retry several times because requests to PostGIS can intermittently fail.
	// Requests to PostGIS could fail for many reasons. For example, a previous
	// test case could cause PostGIS to segfault, causing it to restart (so it
	// may not yet be ready to accept connections by the time this test
	// executes).
	var i int
	for {
		u, err := p.tryUnary(g)
		if err == nil {
			return u, nil
		}

		i++
		if i >= 10 {
			return UnaryResult{}, err
		}

		time.Sleep(time.Millisecond * time.Duration(math.Pow(2, float64(i))))
	}
}

func (p BatchPostGIS) tryUnary(g geom.Geometry) (UnaryResult, error) {
	// WKB and WKB forms returned from PostGIS don't _always_ give the same
	// result (usually differences around empty geometries). In the case of
	// boundary, convex hull, and reverse, they are different enough that it's
	// advantageous to use the WKT form.
	var boundaryWKT sql.NullString
	var convexHullWKT string
	var reverseWKT string
	var dumpWKTs []string

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
			ELSE ST_AsText(ST_Boundary(ST_GeomFromText($3)))
		END,

		ST_AsText(ST_ConvexHull(ST_GeomFromText($3))),

		-- IsRing only defined for LineStrings.
		CASE
			WHEN ST_GeometryType(ST_GeomFromWKB($1)) != 'ST_LineString'
			THEN NULL
			ELSE ST_IsRing(ST_GeomFromWKB($1))
		END,

		ST_Length(ST_GeomFromWKB($1)),
		ST_Area(ST_GeomFromWKB($1)),
		ST_AsBinary(ST_Centroid(ST_GeomFromWKB($1))),
		ST_AsText(ST_Reverse(ST_GeomFromText($3))),
		ST_GeometryType(ST_GeomFromWKB($1)),

		ST_AsBinary(ST_Force2D(ST_GeomFromWKB($1))),
		ST_AsBinary(ST_Force3DZ(ST_GeomFromWKB($1))),
		ST_AsBinary(ST_Force3DM(ST_GeomFromWKB($1))),
		ST_AsBinary(ST_Force4D(ST_GeomFromWKB($1))),

		ST_AsBinary(ST_ForcePolygonCW(ST_GeomFromWKB($1))),
		ST_AsBinary(ST_ForcePolygonCCW(ST_GeomFromWKB($1))),

		array_agg(ST_AsText(geom)) FROM ST_Dump(ST_GeomFromWKB($1))

		`, g, isNestedGeometryCollection(g), g.AsText(),
	).Scan(
		&result.AsText,
		&result.AsBinary,
		&result.AsGeoJSON,
		&result.IsEmpty,
		&result.Dimension,
		&result.Envelope,
		&result.IsSimple,
		&boundaryWKT,
		&convexHullWKT,
		&result.IsRing,
		&result.Length,
		&result.Area,
		&result.Centroid,
		&reverseWKT,
		&result.Type,
		&result.Force2D,
		&result.Force3DZ,
		&result.Force3DM,
		&result.Force4D,
		&result.ForceCW,
		&result.ForceCCW,
		pq.Array(&dumpWKTs),
	)
	if err != nil {
		return result, err
	}

	if boundaryWKT.Valid {
		result.Boundary.Valid = true
		result.Boundary.Geometry, err = geom.UnmarshalWKT(boundaryWKT.String)
		if err != nil {
			return result, err
		}
	}

	result.ConvexHull, err = geom.UnmarshalWKT(convexHullWKT)
	if err != nil {
		return result, err
	}

	result.Reverse, err = geom.UnmarshalWKT(reverseWKT)
	if err != nil {
		return result, err
	}

	// remove ST_ prefix that ST_GeometryType returned, since we don't want ST_ prefix in our type
	result.Type = strings.TrimPrefix(result.Type, postgisTypePrefix)

	for _, wkt := range dumpWKTs {
		dumpGeom, err := geom.UnmarshalWKT(wkt)
		if err != nil {
			return result, err
		}
		result.Dump = append(result.Dump, dumpGeom)
	}
	return result, nil
}

func isNestedGeometryCollection(g geom.Geometry) bool {
	if !g.IsGeometryCollection() {
		return false
	}
	gc := g.MustAsGeometryCollection()
	for i := 0; i < gc.NumGeometries(); i++ {
		if gc.GeometryN(i).IsGeometryCollection() {
			return true
		}
	}
	return false
}
