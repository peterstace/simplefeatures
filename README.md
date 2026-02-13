# Simple Features

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
- [Native Go Packages](#native-go-packages)
  - [Geom](#geom)
  - [Carto](#carto)
  - [R-Tree](#r-tree)
- [C-Wrapper Packages](#c-wrapper-packages)
  - [GEOS](#geos)
  - [PROJ](#proj)

## Native Go Packages

The Simple Features library contains several packages that are implemented
purely in Go, and do not require CGO to use. These are `geom`, `carto`, and
`rtree`.

### Geom

[Package
`geom`](https://pkg.go.dev/github.com/peterstace/simplefeatures/geom?tab=doc)
provides types (`Point`, `MultiPoint`, `LineString`, `MultiLineString`,
`Polygon`, `MultiPolygon`, `GeometryCollection`) representing each geometry, as
well as algorithms that operate with/on them.

The following operations are supported:

- **Serialisation:** WKT, WKB, GeoJSON, TWKB.
- **Overlay operations:** union, intersection, difference, symmetric difference, unary union.
- **Relate operations:** DE-9IM matrix, equals, intersects, disjoint, contains, covered by, covers, overlaps, touches, within, crosses.
- **Measurements:** area, length, distance.
- **Analysis:** centroid, convex hull, envelope, point on surface, boundary, minimum area bounding rectangle, minimum width bounding rectangle.
- **Transformations:** simplify, densify, snap to grid, reverse, force clockwise, force counter-clockwise, affine transformation.
- **Buffer:** buffer with configurable end cap styles, join styles, and single-sided mode.
- **Prepared geometries:** preprocess a geometry for efficient repeated spatial predicate evaluation.
- **Comparison:** exact equals.
- **Linear Interpolation:** interpolate point, interpolate evenly spaced points.
- **Properties:** is simple, is empty, dimension, is clockwise, is counter-clockwise, validate.
- **Envelope operations:** intersects, contains, covers, distance, expand to include, center, width, height, area, bounding diagonal.

The overlay, relate, buffer, and prepared geometry operations are powered by a Go port of
[JTS](https://locationtech.github.io/jts/). This means that it's using robust
and battle tested algorithms that are common to JTS and its derivates (such as
GEOS).

### Carto

[Package
`carto`](https://pkg.go.dev/github.com/peterstace/simplefeatures/carto?tab=doc)
provides cartographic map projections and related functionality. It can be used
as a lite replacement for PROJ for simple cartographic projections.

The following projections are supported:

- **Conic:** Lambert conformal conic, Albers equal area conic, equidistant
  conic.
- **Cylindrical:** equirectangular, web mercator, Lambert cylindrical equal
  area.
- **Azimuthal:** orthographic, azimuthal equidistant.
- **Pseudocylindrical:** sinusoidal.
- **UTM:** All 60 UTM projections.

### R-Tree

[Package
`rtree`](https://pkg.go.dev/github.com/peterstace/simplefeatures/rtree?tab=doc)
provides an in-memory R-Tree data structure, which can be used for fast and
efficient spatial searches.

## C-Wrapper Packages

The Simple Features library also contains several packages that depend on C
libraries (via CGO). They are segregated into separate Go packages so users who
don't need that functionality aren't required to install the relevant C
dependencies.

### GEOS

[Package
`geos`](https://pkg.go.dev/github.com/peterstace/simplefeatures/geos?tab=doc)
is a thin CGO wrapper around the [GEOS](https://www.osgeo.org/projects/geos/)
library. It provides access to GEOS functionality that has not yet been
implemented in the `geom` package natively in Go.

To install the GEOS C library:

```sh
# Debian/Ubuntu
apt-get install libgeos-dev

# Alpine
apk add geos-dev

# macOS
brew install geos
```

### PROJ

[Package
`proj`](https://pkg.go.dev/github.com/peterstace/simplefeatures/proj?tab=doc)
is a thin CGO wrapper around the [PROJ](https://proj.org/) library. It provides
an exhaustive set of cartographic projections and coordinate transformations.

To install the PROJ C library:

```sh
# Debian/Ubuntu
apt-get install libproj-dev

# Alpine
apk add proj-dev

# macOS
brew install proj
```
