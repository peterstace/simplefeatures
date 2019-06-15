# Simple Features

Go Implementation of the OpenGIS Simple Features Specification.

It is based on the [Simple Features Access - Part 1: Common
Architecture](http://www.opengeospatial.org/standards/sfa).

## Feature Checklist

| Type               | Assertions | Dimension | GeometryType | SRID | Envelope | AsText | AsBinary | IsEmpty | IsSimple | Is3D | IsMeasured | Boundary |
| ---                | ---        | ---       | ---          | ---  | ---      | ---    | ---      | ---     | ---      | ---  | ---        | ---      |
| Point              | ✅         | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ⚠️        | ❌   | ❌         | ❌       |
| Line               | ✅         | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ⚠️        | ❌   | ❌         | ❌       |
| LineString         | ✅         | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ✅       | ❌   | ❌         | ❌       |
| LinearRing         | ✅         | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ⚠️        | ❌   | ❌         | ❌       |
| Polygon            | ⚠️          | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ⚠️        | ❌   | ❌         | ❌       |
| MultiPoint         | ✅         | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ⚠️        | ❌   | ❌         | ❌       |
| MultiLineString    | ✅         | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ⚠️        | ❌   | ❌         | ❌       |
| MultiPolygon       | ⚠️          | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ⚠️        | ❌   | ❌         | ❌       |
| GeometryCollection | ✅         | ✅        | ❌           | ❌   | ❌       | ✅     | ❌       | ✅      | ⚠️        | ❌   | ❌         | ❌       |
                                 
### Spatial Relationships

| Type Combination                      | Equals | Disjoin | Intersects | Touches | Crosses | Within | Contains | Overlaps | Relate |
| ---                                   | ---    | ---     | ---        | ---     | ---     | ---    | ---      | ---      | ---    |
| Point/Point                           | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/Line                            | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/LineString                      | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/LinearRing                      | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/Polygon                         | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/MultiPoint                      | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/MultiLineString                 | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/MultiPolygon                    | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Point/GeometryCollection              | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/Line                             | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/LineString                       | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/LinearRing                       | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/Polygon                          | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/MultiPoint                       | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/MultiLineString                  | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/MultiPolygon                     | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Line/GeometryCollection               | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/LineString                 | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/LinearRing                 | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/Polygon                    | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/MultiPoint                 | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/MultiLineString            | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/MultiPolygon               | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LineString/GeometryCollection         | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/LinearRing                 | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/Polygon                    | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/MultiPoint                 | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/MultiLineString            | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/MultiPolygon               | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| LinearRing/GeometryCollection         | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/Polygon                       | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/MultiPoint                    | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/MultiLineString               | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/MultiPolygon                  | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| Polygon/GeometryCollection            | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPoint/MultiPoint                 | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPoint/MultiLineString            | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPoint/MultiPolygon               | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPoint/GeometryCollection         | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultLineString/MultiLineString        | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultLineString/MultiPolygon           | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultLineString/GeometryCollection     | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPolygon/MultiPolygon             | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| MultiPolygon/GeometryCollection       | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |
| GeometryCollection/GeometryCollection | ❌     | ❌      | ❌         | ❌      | ❌      | ❌     | ❌       | ❌       | ❌     |

### Measures on Geometry

| Type               | LocateAlong | LocateBetween |
| ---                | ---         | ---           |
| Point              | ❌          | ❌            |
| Line               | ❌          | ❌            |
| LineString         | ❌          | ❌            |
| LinearRing         | ❌          | ❌            |
| Polygon            | ❌          | ❌            |
| MultiPoint         | ❌          | ❌            |
| MultiLineString    | ❌          | ❌            |
| MultiPolygon       | ❌          | ❌            |
| GeometryCollection | ❌          | ❌            |

### Spatial Analysis

| Type Combination                      | Distance | Buffer | ConvexHull | Intersection | Union | Difference | SymDifference |
| ---                                   | ---      | ---    | ---        | ---          | ---   | ---        | ---           |
| Point/Point                           | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/Line                            | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/LineString                      | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/LinearRing                      | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/Polygon                         | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| Point/MultiPoint                      | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
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
| LineString/LineString                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/LinearRing                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/Polygon                    | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/MultiPoint                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/MultiLineString            | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/MultiPolygon               | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LineString/GeometryCollection         | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| LinearRing/LinearRing                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
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
| MultiPoint/MultiPoint                 | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPoint/MultiLineString            | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPoint/MultiPolygon               | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPoint/GeometryCollection         | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultLineString/MultiLineString        | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultLineString/MultiPolygon           | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultLineString/GeometryCollection     | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPolygon/MultiPolygon             | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| MultiPolygon/GeometryCollection       | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |
| GeometryCollection/GeometryCollection | ❌       | ❌     | ❌         | ⚠️            | ❌    | ❌         | ❌            |

### Type Specific Methods

#### Geometry Collection

| Type               | NumGeometries | GeometryN |
| ---                | ---           | ---       |
| GeometryCollection | ❌            | ❌        |

#### Point

| Type  | X   | Y   | Z   | M   |
| ---   | --- | --- | --- | --- |
| Point | ❌  | ❌  | ❌  | ❌  |

#### Curve (Line, LineString, LinearRing)

| Type               | Length | StartPoint | EndPoint | IsClosed | IsRing | NumPoints | PointN |
| ---                | ---    | ---        | ---      | ---      | ---    | ---       | ---    |
| Line               | ❌     | ❌         | ❌       | ❌       | ❌     | ❌        | ❌     |
| LineString         | ❌     | ❌         | ❌       | ❌       | ❌     | ❌        | ❌     |
| LinearRing         | ❌     | ❌         | ❌       | ❌       | ❌     | ❌        | ❌     |

#### MultiCurve (MultiLineString)

| Type            | IsClosed | Length |
| ---             | ---      | ---    |
| MultiLineString | ❌       | ❌     |

#### Surface (Polygon)

| Type    | Area | Centroid | PointOnSurface | ExteriorRing | NumInteriorRing | InteriorRingN |
| ---     | ---  | ---      | ---            | ---          | ---             | ---           |
| Polygon | ❌   | ❌       | ❌             | ❌           | ❌              | ❌            |

#### MultiSurface (MultiPolygon)

| Type         | Area | Centroid | PointOnSurface |
| ---          | ---  | ---      | ---            |
| MultiPolygon | ❌   | ❌       | ❌             |
