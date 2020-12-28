# Simple Features

[![Documentation](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/peterstace/simplefeatures/geom?tab=doc)
[![Build Status](https://github.com/peterstace/simplefeatures/workflows/build/badge.svg)](https://github.com/peterstace/simplefeatures/actions)
[![Go Report
Card](https://goreportcard.com/badge/github.com/peterstace/simplefeatures)](https://goreportcard.com/report/github.com/peterstace/simplefeatures)

Simple Features is a 2D geometry library. It provides Go types that model
geometries, as well as algorithms that operate on them.

Simple Features is a pure Go Implementation of the OpenGIS Consortium's Simple
Feature Access Specification (which can be found
[here](http://www.opengeospatial.org/standards/sfa)). This is the same
specification that [GEOS](https://trac.osgeo.org/geos),
[JTS](https://locationtech.github.io/jts/), and [PostGIS](https://postgis.net/)
implement, so the Simple Features API will be familiar to developers who have
used those libraries before.

**TODO:** Table of contents

### Geometry Types

<table>

<thead>
<tr>
<th>Type</th>
<th>Example</th>
<th>Description</th>
</tr>
</thead>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#Point">Point</a></td>
<td><a href="https://commons.wikimedia.org/wiki/File:SFA_Point.svg"><img width=51 height=51 src="https://upload.wikimedia.org/wikipedia/commons/c/c2/SFA_Point.svg"></a></td>
<td>Point is a single location in space.</td>
</tr>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#MultiPoint">MultiPoint</a></td>
<td><a href="https://commons.wikimedia.org/wiki/File:SFA_MultiPoint.svg"><img width=51 height=51  src="https://upload.wikimedia.org/wikipedia/commons/d/d6/SFA_MultiPoint.svg"></a></td>
<td>MultiPoint is collection of points in space.</td>
</tr>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#LineString">LineString</a></td>
<td><a href="https://commons.wikimedia.org/wiki/File:SFA_LineString.svg"><img width=51 height=51  src="https://upload.wikimedia.org/wikipedia/commons/b/b9/SFA_LineString.svg"></a></td>
<td>LineString is curve defined by linear interpolation between a set of
control points.</td>
</tr>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#MultiLineString">MultiLineString</a></td>
<td><a href="https://commons.wikimedia.org/wiki/File:SFA_MultiLineString.svg"><img width=51 height=51  src="https://upload.wikimedia.org/wikipedia/commons/8/86/SFA_MultiLineString.svg"></a></td>
<td>MultiLineString is a collection of LineStrings.</td>
</tr>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#Polygon">Polygon</a></td>
<td><a href="https://commons.wikimedia.org/wiki/File:SFA_Polygon.svg"><img width=51 height=51  src="https://upload.wikimedia.org/wikipedia/commons/5/55/SFA_Polygon_with_hole.svg"></a></td>
<td>Polygon is a planar surface geometry that bounds some area. It may have holes.</td>
</tr>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#MultiPolygon">MultiPolygon</a></td>
<td><a href="https://commons.wikimedia.org/wiki/File:SFA_MultiPolygon.svg"><img width=51 height=51  src="https://upload.wikimedia.org/wikipedia/commons/3/3b/SFA_MultiPolygon_with_hole.svg"></a></td>
<td>Polygon is collection of Polygons (with some constraints on how the Polygons interact with each other).</td>
</tr>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#GeometryCollection">GeometryCollection</a></td>
<td><a href="https://commons.wikimedia.org/wiki/File:SFA_GeometryCollection.svg"><img width=51 height=51  src="https://upload.wikimedia.org/wikipedia/commons/1/1d/SFA_GeometryCollection.svg"></a></td>
<td>GeometryCollection is an unconstrained collection of geometries.</td>
</tr>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#Geometry">Geometry</a></td>
<td><a href="https://commons.wikimedia.org/wiki/File:SFA_Polygon.svg"><img width=51 height=51  src="https://upload.wikimedia.org/wikipedia/commons/5/55/SFA_Polygon_with_hole.svg"></a></td>
<td>Geometry holds any type of geometry (Point, MultiPoint, LineString, MultiLineString, Polygon, MultiPolygon, or GeometryCollection). It's the type that the Simple Features library uses when it needs to represent geometries in a generic way.</td>
</tr>

<tr>
<td><a href="https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#Envelope">Envelope</a></td>
<td><img src="./.hidden/assets/envelope.svg"></td>
<td>Envelope is an axis aligned bounding box typically used to describe the spatial extent of other geometric entities.</td>
</tr>

</table>

### Marshalling and Unmarshalling

Simple features supports the following external geometry representation formats:

| Format  | Example                                                              | Description                                                                                                                                                                                                                                                                                                                                                        |
| ---     | ---                                                                  | ---                                                                                                                                                                                                                                                                                                                                                                |
| WKT     | `POLYGON((0 0,0 1,1 1,1 0,0 0))`                                     | [Well Known Text](https://en.wikipedia.org/wiki/Well-known_text_representation_of_geometry) is a human readable format for storing geometries. It's often the lowest common denominator geometry format, and is useful for integration with other GIS applications.                                                                                                |
| WKB     | `<binary>`                                                           | [Well Known Binary](https://en.wikipedia.org/wiki/Well-known_text_representation_of_geometry#Well-known_binary) is a machine readable format that is efficient for computers to use (both from a processing and storage space perspective). WKB is a good choice for transferring geometries to and from PostGIS and other databases that support geometric types. |
| GeoJSON | `{"type":"Polygon","coordinates":[[[0,0],[0,1],[1,1],[1,0],[0,0]]]}` | [GeoJSON](https://en.wikipedia.org/wiki/GeoJSON) represents geometries in a similar way to WKB, but is based on the JSON format. This makes it ideal to use with web APIs or other situations where JSON would normally be used.                                                                                                                                   |

### Geometry Algorithms

The following algorithms are supported:

| Algorithm                                                                         | Description                                         |
| ---                                                                               | ---                                                 |
| [Distance](https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#Distance) | Finds the shortest distance between two geometries. |

**TODO:** Add other algorithms to this table.

### GEOS Wrapper

A [GEOS](https://www.osgeo.org/projects/geos/) CGO wrapper is also provided,
giving access to functionality not yet implemented natively in Go. The [wrapper
is implemented in a separate
package](https://pkg.go.dev/github.com/peterstace/simplefeatures/geos?tab=doc),
meaning that library users who don't need this additional functionality don't
need to expose themselves to CGO.

### Examples

The following examples show some common operations (errors are omitted for
brevity).

#### WKT

Encoding and decoding WKT:

```
// Unmarshal from WKT
input := "POLYGON((0 0,0 1,1 1,1 0,0 0))"
g, _ := geom.UnmarshalWKT(input)

// Marshal to WKT
output := g.AsText()
fmt.Println(output) // Prints: POLYGON((0 0,0 1,1 1,1 0,0 0))
```

#### WKB

Encoding and decoding WKB directly:

```
// Marshal as WKB
pt := geom.NewPointFromXY(geom.XY{1.5, 2.5})
wkb := pt.AsBinary()
fmt.Println(wkb) // Prints: [1 1 0 0 0 0 0 0 0 0 0 248 63 0 0 0 0 0 0 4 64]

// Unmarshal from WKB
fromWKB, _ := geom.UnmarshalWKB(wkb)
fmt.Println(fromWKB.AsText()) // POINT(1.5 2.5)
```

Encoding and decoding WKB for integration with PostGIS:

```
db, _ := sql.Open("postgres", "postgres://...")

db.Exec(`
    CREATE TABLE my_table (
        my_geom geometry(geometry, 4326),
        population double precision
    )`,
)

// Insert our geometry and population data into PostGIS via WKB.
nyc := geom.NewPointFromXY(geom.XY{-74.0, 40.7})
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

```
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

```
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

### FAQs

**Q:** Why create Simple Features when the GEOS library already exists?

**A:** TODO
