# Simple Features

[![Documentation](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/peterstace/simplefeatures/geom?tab=doc)
[![Build Status](https://github.com/peterstace/simplefeatures/workflows/build/badge.svg)](https://github.com/peterstace/simplefeatures/actions)
[![Go Report
Card](https://goreportcard.com/badge/github.com/peterstace/simplefeatures)](https://goreportcard.com/report/github.com/peterstace/simplefeatures)
[![Coverage
Status](https://coveralls.io/repos/github/peterstace/simplefeatures/badge.svg?branch=master)](https://coveralls.io/github/peterstace/simplefeatures?branch=master)

Simple Features is a 2D geometry library that provides Go types that model
geometries, as well as algorithms that operate on them.

It's a pure Go Implementation of the OpenGIS Consortium's Simple
Feature Access Specification (which can be found
[here](http://www.opengeospatial.org/standards/sfa)). This is the same
specification that [GEOS](https://trac.osgeo.org/geos),
[JTS](https://locationtech.github.io/jts/), and [PostGIS](https://postgis.net/)
implement, so the Simple Features API will be familiar to developers who have
used those libraries before.

**Table of Contents:**
- [Supported Features](#supported-features)
- [GEOS Wrapper](#geos-wrapper)
- [PROJ Wrapper](#proj-wrapper)
- [Examples](#examples)

### Package Structure

TODO

### Native Go Supported Features

- **All 7 OGC Geometry Types:** `Point`, `MultiPoint`, `LineString`, `MultiLineString`, `Polygon`, `MultiPolygon`, `GeometryCollection`.
- **Overlay operations:** union, intersection, difference, symmetric difference, unary union.
- **Relate operations:** DE-9IM matrix, equals, intersects, disjoint, contains, covered by, covers, overlaps, touches, within, crosses.
- **Envelope operations:** intersects, contains, covers, distance, expand to include, center, width, height, area, bounding diagonal.
- **Measurements:** area, length, distance.
- **Analysis:** centroid, convex hull, envelope, point on surface, boundary, minimum area bounding rectangle, minimum width bounding rectangle.
- **Transformations:** simplify, densify, snap to grid, reverse, force clockwise, force counter-clockwise, affine transformation.
- **Comparison:** exact equals.
- **Linear Interpolation:** interpolate point, interpolate evenly spaced points.
- **Properties:** is simple, is empty, dimension, is clockwise, is counter-clockwise, validate.
- **Serialisation:** WKT, WKB, GeoJSON, TWKB.
- **Map projections:** equirectangular, web mercator, orthographic, sinusoidal,
  UTM, Lambert cylindrical equal area, Lambert conformal conic, Albers equal
  area conic, equidistant conic, azimuthal equidistant.

### GEOS Wrapper

A [GEOS](https://www.osgeo.org/projects/geos/) CGO wrapper is also provided,
giving access to functionality not yet implemented natively in Go. The [wrapper
is implemented in a separate
package](https://pkg.go.dev/github.com/peterstace/simplefeatures/geos?tab=doc),
meaning that library users who don't need this additional functionality don't
need to expose themselves to CGO.

### PROJ Wrapper

A [PROJ](https://proj.org/) CGO wrapper is also provided, giving access to a
vast array of transformations between various coordinate reference systems.
The [wrapper is implemented in a separate
package](https://pkg.go.dev/github.com/peterstace/simplefeatures/proj?tab=doc),
meaning that library users who don't need this additional functionality don't
need to expose themselves to CGO.

### JTS Port

TODO

### Examples

TODO: Do we need this section?

The following examples show some common operations (errors are omitted for
brevity).

#### WKT

Encoding and decoding WKT:

```go
// Unmarshal from WKT
input := "POLYGON((0 0,0 1,1 1,1 0,0 0))"
g, _ := geom.UnmarshalWKT(input)

// Marshal to WKT
output := g.AsText()
fmt.Println(output) // Prints: POLYGON((0 0,0 1,1 1,1 0,0 0))
```

#### WKB

Encoding and decoding WKB directly:

```go
// Marshal as WKB
coords := geom.Coordinates{XY: geom.XY{1.5, 2.5}}
pt := geom.NewPoint(coords)
wkb := pt.AsBinary()
fmt.Println(wkb) // Prints: [1 1 0 0 0 0 0 0 0 0 0 248 63 0 0 0 0 0 0 4 64]

// Unmarshal from WKB
fromWKB, _ := geom.UnmarshalWKB(wkb)
fmt.Println(fromWKB.AsText()) // POINT(1.5 2.5)
```

Encoding and decoding WKB for integration with PostGIS:

```go
db, _ := sql.Open("postgres", "postgres://...")

db.Exec(`
    CREATE TABLE my_table (
        my_geom geometry(geometry, 4326),
        population double precision
    )`,
)

// Insert our geometry and population data into PostGIS via WKB.
coords := geom.Coordinates{XY: geom.XY{-74.0, 40.7}}
nyc := geom.NewPoint(coords)
db.Exec(`
    INSERT INTO my_table
    (my_geom, population)
    VALUES (ST_GeomFromWKB($1, 4326), $2)`,
    nyc, 8.4e6,
)

// Get the geometry and population data back out of PostGIS via WKB.
var location geom.Geometry
var population float64
db.QueryRow(`
    SELECT ST_AsBinary(my_geom), population
    FROM my_table LIMIT 1`,
).Scan(&location, &population)
fmt.Println(location.AsText(), population) // Prints: POINT(-74 40.7) 8.4e+06
```

#### GeoJSON

Encoding and decoding GeoJSON directly:

```go
// Unmarshal geometry from GeoJSON.
raw := `{"type":"Point","coordinates":[-74.0,40.7]}`
var g geom.Geometry
json.NewDecoder(strings.NewReader(raw)).Decode(&g)
fmt.Println(g.AsText()) // Prints: POINT(-74 40.7)

// Marshal back to GeoJSON.
enc := json.NewEncoder(os.Stdout)
enc.Encode(g) // Prints: {"type":"Point","coordinates":[-74,40.7]}
```

Geometries can also be part of larger structs:

```go
type CityPopulation struct {
    Location   geom.Geometry `json:"loc"`
    Population int           `json:"pop"`
}

// Unmarshal geometry from GeoJSON.
raw := `{"loc":{"type":"Point","coordinates":[-74.0,40.7]},"pop":8400000}`
var v CityPopulation
json.NewDecoder(strings.NewReader(raw)).Decode(&v)
fmt.Println(v.Location.AsText()) // Prints: POINT(-74 40.7)
fmt.Println(v.Population)        // Prints: 8400000

// Marshal back to GeoJSON.
enc := json.NewEncoder(os.Stdout)
enc.Encode(v) // Prints: {"loc":{"type":"Point","coordinates":[-74,40.7]},"pop":8400000}
```
