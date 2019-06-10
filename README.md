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

### Interfaces

#### Geometry Methods

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

#### Curve

#### Surface

#### MultiCurve

#### MultiSurface

## TODO

- Implement type construction constraints
- 2D, Z, M, ZM points.
- Spatial Reference Systems.
- Type properties
- Predicates
- Operators
- WKB
