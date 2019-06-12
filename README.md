# Simple Features

Go Implementation of the OpenGIS Simple Features Specification.

It is based on the [Simple Features Access - Part 1: Common
Architecture](http://www.opengeospatial.org/standards/sfa).

## Feature Checklist

| Type                       | Geometry Assertions |
| ---                        | ---                 |
| Point                      | ✅                  |
| Line                       | ✅                  |
| LineString                 | ✅                  |
| LinearRing                 | ✅                  |
| Polygon                    | ⚠️                   |
| MultiPoint                 | ✅                  |
| MultiLineString            | ✅                  |
| MultiPolygon (non-trivial) | ⚠️                   |
| GeometryCollection         | ✅                  |
                                 
### Geometry Attributes

- [ ] Dimension
- [ ] GeometryType
- [ ] SRID
- [ ] Envelope
- [x] AsText
- [ ] AsBinary
- [ ] IsEmpty
- [ ] IsSimple
- [ ] Is3D
- [ ] IsMeasured
- [ ] Boundary

### Spatial Relationships

Matrix:

- [ ] Equals
- [ ] Disjoint
- [ ] Intersects
- [ ] Touches
- [ ] Crosses
- [ ] Within
- [ ] Contains
- [ ] Overlaps
- [ ] Relate

### Measures on Geometry

- [ ] LocateAlong
- [ ] LocateBetween

### Spatial Analysis

Matrix

- [ ] Distance
- [ ] Buffer
- [ ] ConvexHull
- [ ] Intersection (done for Line:Line)
- [ ] Union
- [ ] Difference
- [ ] SymDifference

### Type Specific Methods

#### Geometry Collection

- [ ] NumGeometries
- [ ] GeometryN

#### Point

- [ ] X
- [ ] Y
- [ ] Z
- [ ] M

#### Curve (Line, LineString, LinearRing)

- [ ] Length
- [ ] StartPoint
- [ ] EndPoint
- [ ] IsClosed
- [ ] IsRing
- [ ] NumPoints
- [ ] PointN

#### MultiCurve (MultiLineString)

- [ ] IsClosed
- [ ] Length

#### Surface (Polygon)

- [ ] Area
- [ ] Centroid
- [ ] PointOnSurface
- [ ] ExteriorRing
- [ ] NumInteriorRing
- [ ] InteriorRingN

#### MultiSurface (MultiPolygon)

- [ ] Area
- [ ] Centroid
- [ ] PointOnSurface
