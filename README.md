# Simple Features

Go Implementation of the OpenGIS Simple Features Specification.

It is based on the [Simple Features Access - Part 1: Common
Architecture](http://www.opengeospatial.org/standards/sfa).

## Feature Checklist

### Types

- [x] Point
- [x] LineString
- [x] Line
- [x] LinearRing
- [/] Polygon
- [x] MultiPoint
- [x] MultiLineString
- [/] MultiPolygon (non-trivial)
- [x] GeometryCollection
- [ ] TIN
- [ ] Triangle
- [ ] PolyhedralSurface

### Geometry Attributes

- [ ] Dimension
- [ ] GeometryType
- [ ] SRID
- [ ] Envelope
- [x] AsText
- [ ] AsBinary
- [ ] IsEmpty
- [/] IsSimple
- [ ] Is3D
- [ ] IsMeasured
- [ ] Boundary

### Spatial Relationships

- [ ] Equals
- [ ] Disjoint
- [ ] Intersects
- [ ] Touches
- [ ] Crosses
- [ ] Within
- [ ] Contains
- [ ] Overlaps
- [ ] Relate
- [ ] LocateAlong
- [ ] LocateBetween

### Spatial Analysis

- [ ] Distance
- [ ] Buffer
- [ ] ConvexHull
- [ ] Intersection
- [ ] Union
- [ ] Difference
- [ ] SymDifference

### Type Specific Methods

TODO

## TODO

- Implement type construction constraints
- 2D, Z, M, ZM points.
- Spatial Reference Systems.
- Type properties
- Predicates
- Operators
- WKB
