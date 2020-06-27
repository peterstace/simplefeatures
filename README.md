# Simple Features

[![Build Status](https://github.com/peterstace/simplefeatures/workflows/build/badge.svg)](https://github.com/peterstace/simplefeatures/actions)
[![Documentation](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/peterstace/simplefeatures/geom?tab=doc)

Simple Features is a pure Go Implementation of the OpenGIS Simple Feature Access
Specification (which can be found
[here](http://www.opengeospatial.org/standards/sfa)).

The specification describes a common access and storage model for 2-dimensional
geometries. This is the same access and storage model used by libraries such as
[GEOS](https://trac.osgeo.org/geos),
[JTS](https://locationtech.github.io/jts/), and
[PostGIS](https://postgis.net/).

Simple Features also provides spatial analysis algorithms that operate on
2-dimensional geometries.

### Changelog

The changelog can be found [here](CHANGELOG.md).

### Library Features (native Go)

- Marshalling/unmarshalling (WKT, WKB, GeoJSON).

- 3D and Measure coordinates.

- Spatial analysis (geometry validation, boundary calculation, envelopes,
  convex hull, equality, is simple, intersects, length, closed, ring, area,
centroid, point on surface).

- Geometry manipulation (reverse, pointwise transform, force coordinates
  types).

### GEOS Wrapper

A [GEOS](https://www.osgeo.org/projects/geos/) CGO wrapper is also provided,
giving access to functionality not yet implemented natively in Go. The [wrapper
is implemented in a separate
package](https://pkg.go.dev/github.com/peterstace/simplefeatures/geos?tab=doc),
meaning that library users who don't need this additional functionality don't
need to expose themselves to CGO.
