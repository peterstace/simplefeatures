# Simple Features

[![Build Status](https://travis-ci.com/peterstace/simplefeatures.svg?token=ueRpGt4cSSnk321nW8xG&branch=master)](https://travis-ci.com/peterstace/simplefeatures)
[![Documentation](https://godoc.org/github.com/peterstace/simplefeatures?status.svg)](http://godoc.org/github.com/peterstace/simplefeatures)

Go Implementation of the OpenGIS Simple Features Specification.

It is based on the [Simple Features Access - Part 1: Common
Architecture](http://www.opengeospatial.org/standards/sfa).

#### Features Not Planned Yet

- SRIDs
- Z/M Values

### Running Tests

Unit tests can be run in the usual Go way:

```
go test ./...
```

To run the integration tests as well, use the `docker-compose.yml` file:

```
docker-compose up --abort-on-container-exit
```

### Feature Checklist

| Type               | Assertions | Dimension | Envelope | AsText | AsBinary | IsEmpty | IsSimple | Boundary |
| ---                | ---        | ---       | ---      | ---    | ---      | ---     | ---      | ---      |
| Empty              | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| Point              | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| Line               | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| LineString         | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| LinearRing         | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| Polygon            | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| MultiPoint         | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| MultiLineString    | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| MultiPolygon       | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | ✅       | ✅       |
| GeometryCollection | ✅         | ✅        | ✅       | ✅     | ❌       | ✅      | N/A      | ✅       |
                                 
### Type Specific Methods

#### Accessors

| Accessor Method  | Point | Line | LineString | LinearRing | Polygon | MultiPoint | MultiLineString | MultiPolygon | GeometryCollection |
| ---              | ---   | ---  | ---        | ---        | ---     | ---        | ---             | ---          | ---                |
| XY               | ✅    |      |            |            |         |            |                 |              |                    |
| NumPoints        |       | ✅   | ✅         | ✅         |         | ✅         |                 |              |                    |
| PointN           |       | ✅   | ✅         | ✅         |         | ✅         |                 |              |                    |
| StartPoint       |       | ✅   | ✅         | ✅         |         |            |                 |              |                    |
| EndPoint         |       | ✅   | ✅         | ✅         |         |            |                 |              |                    |
| ExteriorRing     |       |      |            |            | ✅      |            |                 |              |                    |
| NumInteriorRings |       |      |            |            | ✅      |            |                 |              |                    |
| InteriorRingN    |       |      |            |            | ✅      |            |                 |              |                    |
| NumLineStrings   |       |      |            |            |         |            | ✅              |              |                    |
| LineStringN      |       |      |            |            |         |            | ✅              |              |                    |
| NumPolygons      |       |      |            |            |         |            |                 | ✅           |                    |
| PolygonN         |       |      |            |            |         |            |                 | ✅           |                    |
| NumGeometries    |       |      |            |            |         |            |                 |              | ✅                 |
| GeometryN        |       |      |            |            |         |            |                 |              | ✅                 |

#### Calculations

| Calc Method    | Point | Line | LineString | LinearRing | Polygon | MultiPoint | MultiLineString | MultiPolygon | GeometryCollection |
| ---            | ---   | ---  | ---        | ---        | ---     | ---        | ---             | ---          | ---                |
| Length         |       | ❌   | ❌         | ❌         |         |            | ❌              |              |                    |
| IsClosed       |       | ❌   | ✅         | ❌         |         |            | ❌              |              |                    |
| IsRing         |       | ❌   | ❌         | ❌         |         |            |                 |              |                    |
| Area           |       |      |            |            | ❌      |            |                 | ❌           |                    |
| Centroid       |       |      |            |            | ❌      |            |                 | ❌           |                    |
| PointOnSurface |       |      |            |            | ❌      |            |                 | ❌           |                    |

### Spatial Relationships

| Type Combination                      | Equals | Disjoin | Intersects | Touches | Crosses | Within | Contains | Overlaps | Relate |
| ---                                   | ---    | ---     | ---        | ---     | ---     | ---    | ---      | ---      | ---    |
| Empty/Empty                           | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/Point                           | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/Line                            | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/LineString                      | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/LinearRing                      | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/Polygon                         | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/MultiPoint                      | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/MultiLineString                 | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/MultiPolygon                    | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Empty/GeometryCollection              | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/Point                           | ✅     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/Line                            | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/LineString                      | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/LinearRing                      | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/Polygon                         | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/MultiPoint                      | ✅     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/MultiLineString                 | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/MultiPolygon                    | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/GeometryCollection              | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/Line                             | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/LineString                       | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/LinearRing                       | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/Polygon                          | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/MultiPoint                       | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/MultiLineString                  | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/MultiPolygon                     | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/GeometryCollection               | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/LineString                 | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/LinearRing                 | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/Polygon                    | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/MultiPoint                 | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/MultiLineString            | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/MultiPolygon               | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/GeometryCollection         | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/LinearRing                 | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/Polygon                    | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/MultiPoint                 | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/MultiLineString            | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/MultiPolygon               | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/GeometryCollection         | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/Polygon                       | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/MultiPoint                    | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/MultiLineString               | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/MultiPolygon                  | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/GeometryCollection            | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPoint/MultiPoint                 | ✅     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPoint/MultiLineString            | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPoint/MultiPolygon               | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPoint/GeometryCollection         | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultLineString/MultiLineString        | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultLineString/MultiPolygon           | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultLineString/GeometryCollection     | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPolygon/MultiPolygon             | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPolygon/GeometryCollection       | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| GeometryCollection/GeometryCollection | ⚠️      | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |

### Spatial Analysis

| Type Combination                      | Distance | Buffer | ConvexHull | Intersection | Union | Difference | SymDifference |
| ---                                   | ---      | ---    | ---        | ---          | ---   | ---        | ---           |
| Empty/Empty                           | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/Point                           | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/Line                            | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/LineString                      | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/LinearRing                      | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/Polygon                         | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/MultiPoint                      | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/MultiLineString                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/MultiPolygon                    | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Empty/GeometryCollection              | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/Point                           | ❌       | ❌     | ❌         | ✅           | ❌    | ❌         | ❌            |
| Point/Line                            | ❌       | ❌     | ❌         | ✅           | ❌    | ❌         | ❌            |
| Point/LineString                      | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/LinearRing                      | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/Polygon                         | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/MultiPoint                      | ❌       | ❌     | ❌         | ✅           | ❌    | ❌         | ❌            |
| Point/MultiLineString                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/MultiPolygon                    | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/GeometryCollection              | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Line/Line                             | ❌       | ❌     | ❌         | ✅           | ❌    | ❌         | ❌            |
| Line/LineString                       | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Line/LinearRing                       | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Line/Polygon                          | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Line/MultiPoint                       | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Line/MultiLineString                  | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Line/MultiPolygon                     | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Line/GeometryCollection               | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/LineString                 | ❌       | ❌     | ❌         | ✅           | ❌    | ❌         | ❌            |
| LineString/LinearRing                 | ❌       | ❌     | ❌         | ✅           | ❌    | ❌         | ❌            |
| LineString/Polygon                    | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/MultiPoint                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/MultiLineString            | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/MultiPolygon               | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/GeometryCollection         | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LinearRing/LinearRing                 | ❌       | ❌     | ❌         | ✅           | ❌    | ❌         | ❌            |
| LinearRing/Polygon                    | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LinearRing/MultiPoint                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LinearRing/MultiLineString            | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LinearRing/MultiPolygon               | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LinearRing/GeometryCollection         | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Polygon/Polygon                       | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Polygon/MultiPoint                    | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Polygon/MultiLineString               | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Polygon/MultiPolygon                  | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Polygon/GeometryCollection            | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPoint/MultiPoint                 | ❌       | ❌     | ❌         | ✅           | ❌    | ❌         | ❌            |
| MultiPoint/MultiLineString            | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPoint/MultiPolygon               | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPoint/GeometryCollection         | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultLineString/MultiLineString        | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultLineString/MultiPolygon           | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultLineString/GeometryCollection     | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPolygon/MultiPolygon             | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPolygon/GeometryCollection       | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| GeometryCollection/GeometryCollection | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
