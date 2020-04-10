# Changelog

## v0.12.0

2020-04-11

- Removes redundant PostGIS reference implementation tests, improving CI speed.

- Overhauls WKB and WKT interfaces. These now return `[]byte` and `string`
  respectively, rather than writing to a supplied `io.Writer`.

- Removes the `IsValid` method. If users want to know if a geometry is valid or
  not, they should check for an error when geometries are constructed.

- Unexports the partially implemented `Intersection` method. It will be
  reexported once the feature is complete.

- Fixes a memory related bug in the `github.com/peterstace/simplefeatures/geos`
  package.

- Adds a wrapper for the GEOS simplify function.

- Simplifies the `github.com/peterstace/simplefeatures/geos` package by not
  exporting the 'handle' concept. The package now just exposes standalone
functions.

## v0.11.0

2020-04-05

- Adds a new package `github.com/peterstace/simplefeatures/geos` that wraps the
  [GEOS](https://github.com/libgeos/geos) library. The following functions are
wrapped: Equals, Disjoint, Touches, Contains, Covers, Intersects, Within,
CoveredBy, Crosses, Overlaps, Union, Intersection, and Buffer.

## v0.10.1

2020-03-24

- Adds documentation comments to all exported symbols.

## v0.10.0

2020-03-20

- Adds support to geometries for Z (3D) and M (Measure) values. This includes
  many breaking changes, primarily around geometry constructor functions.

## v0.9.0:

2020-03-03

__Special thanks to Frank Sun and Peter Jones for contributing to this release.__

- Fixes a bug in the intersects predicate between `MultiLineString` and
  `MultiPolygon`.

- Removes the `EmptySet` type. The `Point`, `LineString`, and `Polygon` types
  can now represent empty geometries.

- Adds the `GeometryType` method, which returns a string representation of the
  geometry type (e.g. returns `"LineString"` for the `LineString` geometries).

- Adds support to store empty `Point` geometries within `MultiPoint`
  geometries.

- Simplifies geometry constructor options by removing the
  `DisableExpensiveValidations` option.

- Removes the `Equals` predicate (which was only implemented for a small
  proportion of geometry type pairs).

- Modifies the `TransformXY` methods to return their own concrete geometry
  type.

## v0.8.0

2020-02-20

__Special thanks to Lachlan Patrick for contributing to this release.__

- Adds a `Length` implementation for the `Geometry` type.

- Modifies the `Boundary` for concrete geometry types to return a concrete
  geometry type (rather than the generic `Geometry` type).

- Modifies the Fuzz Tests to use batching.

- Fixes a bug in the `IsSimple` method for `MultiLineString`.

- Adds `Centroid` implementations for all geometry types.

- Adds a new suite of reference implementation tests using `libgeos`.

## v0.7.0

2020-01-24

- Fixes a deficiency where `LineString` would not retain coincident adjacent
  points.

- Adds two new methods to `LineString`. `NumLines` gives the number of `Line`
  segments making up the `LineString`, and `LineN` allows access to those
`Line` segments.

- Reduces the memory required to store a `LineString`.

## v0.6.0

2020-01-21

__Special thanks to Lachlan Patrick for contributing to this release.__

- Adds `Reverse` methods, which reverses the order of each geometry's control
  points.

- Adds `SignedArea` methods, which calculate the signed area of geometries. The
  signed area takes into consideration winding order, and produces either a
negative or positive result for non-empty areal geometries.

## v0.5.0

2020-01-17

- Fixes a bug where polygons with nested rings would erroneously be reported as
  valid.

- Performance improvements have been made to `Polygon` and `MultiPolygon`
  validation. The algorithms used now have sub-quadratic time complexity, and
memory allocations have been significantly reduced.

## v0.4.0

2020-01-07

- The `Geometry` interface has been replaced with a concrete type also named
  `Geometry`. This new type holds exactly one geometry value (one of
`EmptySet`, `Point`, `Line`, `LineString`, `Polygon`, `MultiPoint`,
`MultiLineString`, `MultiPolygon`, `GeometryCollection`. The `AnyGeometry` type
has been removed (`Geometry` can be used instead).

## v0.3.0

2020-01-07

- A Linesweep algorithm is now used for the `Intersects() bool` implementation
  between line types (`Line`, `LineString`, `MultiLineString`). This reduces
the computational complexity from quadratic time to linearithmic time.

## v0.2.0

2019-12-24

- The `Intersects` method is now implemented for all geometry pairs. The method
  signature has been changed to no longer return an error (errors were only
returned for unimplemented geometry pairs).

## v0.1.0

2019-12-02

__Special thanks to Lachlan Patrick and Den Tsou for contributing to this release.__

Initial tagged version.
