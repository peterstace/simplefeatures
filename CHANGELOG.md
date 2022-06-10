# Changelog

## v0.39.0

2022-06-10

__Special thanks to David McLeish and Albert Teoh for contributing to this
release.__

- Add support for `GeometryCollection`s in `Union`, `Intersection`,
  `Difference`, `SymmetricDifference`, and `Relate`.

## v0.38.0

2022-05-27

__Special thanks to Sameera Perera and Albert Teoh for contributing to this release.__

- Add initial linear referencing methods to `LineString`. The initial methods
  are `InterpolatePoint` and `InterpolateEvenlySpacedPoints`.

- Fixes a bug in the `TransformXY` method where empty `MultiPoint` and
  `MultiLineString`s would have their coordinates type downgraded to XY.

- Add a new `DumpRings` method to the `Polygon` type, which gives the rings of
  the polygon as a slice of `LineString`s.

- Uses `unsafe.Slice` for internal WKB conversions. This increases the minimum
  Go version required to use simplefeatures from 1.14 to 1.17.

## v0.37.0

2022-03-29

__Special thanks to Lachlan Patrick and Albert Teoh for contributing to this release.__

- Improves performance of `ForceCW` and `ForceCCW` methods by eliminating
  unneeded memory allocations.

- Adds full support for TWKB (Tiny Well Known Binary) as a serialisation format.

- Fixes a vet warning affecting Go 1.18 relating to printf verbs in tests.

- Fixes a bug in `ExactEquals` that incorrectly compares empty points of unequal
  coordinate type as being equal.

## v0.36.0

2022-01-24

__Special thanks to Lachlan Patrick and Albert Teoh for contributing to this release.__

- Eliminates redundant calls to the optional user supplied transform func during
  area calculations.

- Adds `IsCW` and `IsCCW` methods, which check if geometries have consistent
  clockwise or counterclockwise winding orientation.

## v0.35.0

2021-11-23

__Special thanks to Albert Teoh and Sameera Perera for contributing to this release.__

- Fixes spelling of "Marshaller" when referring to the interface defined in the
  `encoding/json` package.

- Adds `UnmarshalJSON` methods to each concrete geometry type
  (`GeometryCollection`, `Point`, `MultiPoint`, `LineString`,
  `MultiLineString`, `Polygon`, `MultiPolygon`). This causes these types to
  implement the `encoding/json.Unmarshaler` interface. GeoJSON can now be
  unmarshalled directly into a concrete geometry type.

- Uses the `%w` verb for wrapping errors internally. Note that simplefeatures
  does not yet currently expose any sentinel errors or error types.

- **Breaking change**: Changes the `Simplify` package level function to become
  a method on the `Geometry` type. Users upgrading can just change function
  invocations that look like `simp, err := geom.Simplify(g, tolerance)` to
  method invocations that look like `simp, err := g.Simplify(tolerance)`.

- Adds `Simplify` methods to the concrete geometry types `LineString`,
  `MultiLineString`, `Polygon`, `MultiPolygon`, and `GeometryCollection`. These
  methods may be used if one of these concrete geometries is to be simplified,
  rather than converting to a `Geometry`, calling `Simplify`, then converting
  back to the concrete geometry type.

- Fixes a bug in Simplify where invalid interior rings would be omitted rather
  than producing an error.

- Adds a wrapper in the `geos` package for the `GEOSMakeValid_r` function
  (exposed as `MakeValid`).

## v0.34.0

2021-11-02

__Special thanks to Albert Teoh for contributing to this release.__

- **Breaking change**: Renames the `AsFoo` methods of the Geometry type to
  `MustAsFoo` (where `Foo` is a concrete geometry type such as `Point`). This
  follows the go convention that methods and functions prefixed with Must may
  panic if preconditions are not met. Note that there's no change in behaviour
  here, it's simply a rename (these methods previously panicked). Users may
  resolve this breaking change by just updating the names of any `AsFoo`
  methods they are calling to `MustAsFoo`.

- **Breaking change**: Adds new methods named `AsFoo` to the Geometry type.
  These methods have the signature `AsFoo() (Foo, bool)`. The boolean return
  value indicates if the conversion was successful or not. These methods are
  useful because they allow concrete geometries to be extracted from a Geometry
  value, with the concrete type for the `Is` and `As` call only specified once.
  Users now just have to call `AsFoo`, and can then check the flag. This helps
  to eliminate the class of bugs there the type specified with `IsFoo`
  erroneously differs from the type specified by `AsFoo`.

## v0.33.1

2021-10-14

__Special thanks to Albert Teoh for contributing to this release.__

- Adds a new method `MinMaxXYs (XY, XY, bool)` to the `Envelope` type. The
  first two return values are the minimum and maximum XY values in the
  envelope, and the third return value indicates whether or not the first two
  are defined (they are only defined for non-empty envelopes).

## v0.33.0

2021-10-11

__Special thanks to Albert Teoh for contributing to this release.__

- **Breaking change**: The `Envelope` type can now be an empty envelope.
  Previously, it was only able to represent a rectangle with some area, a
  horizontal or vertical line, or a single point. Its `AsGeometry` returns
  an empty `GeometryCollection` in the case where it's empty. The result of
  `AsGeometry` is unchanged for non-empty envelopes.

- **Breaking change**: The `NewEnvelope` function signature has changed. It now
  accepts a slice of `geom.XY` as the sole argument. The behaviour of the
  function is the same as before, except that if no XY values are provided then
  an empty envelope is returned without error.

- **Breaking change**: The `Envelope` type's `EnvelopeFromGeoms` method has
  been removed. To replicate the behaviour of this method, users can construct
  a `GeometryCollection` and call its `Envelope` method.

- **Breaking change**: The `Envelope` type's `Min`, `Max`, and `Center` methods
  now return `Point`s rather than `XY`s. When the envelope is empty, `Min`,
  `Max`, and `Center` return empty points.

- **Breaking change**: The `Envelope` type's `Distance` method now returns
  `(float64, bool)` rather than `float64`. The returned boolean is only true if
  the distance between the two envelopes is defined (i.e. when they are both
  non-empty).

- **Breaking change**: The `Envelope` method on the `Geometry`,
  `GeometryCollection`, `Point`, `LineString`, `Polygon`, `MultiPoint`,
  `MultiLineString`, and `MultiPolygon` types now return `Envelope` instead of
  `(Envelope, bool)`. The empty vs non-empty status is encoded inside the
  envelope instead of via an explicit boolean.

- The `Envelope` type now has `IsEmpty`, `IsPoint`, `IsLine`, and
  `IsRectanagle` methods. These correspond to the 4 possible envelope
  categories.

## v0.32.0

2021-09-08

__Special thanks to Albert Teoh for contributing to this release.__

- **Breaking change**: Consolidates `MultiPoint` constructors and simplifies
  `MultiPoint` internal representation. Removes the `BitSet` type, previously
  used for `MultiPoint` construction. Removes the `NewMultiPointFromPoints` and
  `NewMultiPointWithEmptyMask` functions. Modifies the `NewMultiPoint` function
  to accept a slice of `Point`s rather than a `Sequence`.

- **Breaking change**: Consolidates `Point` construction. Removes the
  `NewPointFromXY` function. It is replaced by a new `AsPoint` method on the
  `XY` type.

- Refactors internal test helpers.

- Adds linting to CI using `golangci-lint`.

- **Breaking change**: Renames geometry constructors for consistency.
  `NewPolygonFromRings` is renamed to `NewPolygon`.
  `NewMultiLineStringFromLineStrings` is renamed to `NewMultiLineString`.
  `NewMultiPolygonFromPolygons` is renamed to `NewMultiPolygon`.

- **Breaking change**: Adds checks for anomalous `float64` values (NaN and +/-
  infinity) during geometry construction.

	- The `NewPoint` function now returns `(Point, error)` rather than `Point`.
	  The returned error is non-nil when the inputs contain anomalous values.

	- The `NewLineString` function's signature doesn't change, but now returns
	  a non-nil error if the input `Sequence` contains anomalous values.

	- The `OmitInvalid` constructor option now has implications when
	  constructing `Point` and `MultiPoint` types.

	- The `NewEnvelope` function now returns `(Envelope, error)` rather than
	  `Envelope`. The returned error is non-nil when when the input XYs contain
	  anomalous values.

	- The `Envelope` type's `ExtendToIncludePoint` method is renamed to
	  `ExtendToIncludeXY` (better matching its argument type). It now returns
	  `(Envelope, erorr)` rather than `Envelope`. The returned error is non-nil
	  if the inputs contain any anomalous values.

	- The `Envelope` type's `ExpandBy` method is removed due to its limited
	  utility and complex interactions with anomalous values.

## v0.31.0

2021-08-09

__Special thanks to Albert Teoh for contributing to this release.__

- Fixes some minor linting (and other similar) issues identified by Go Report
  Card.

- Adds a new `DumpCoordinates` method to geometry types. This method returns a
  `Sequence` containing all of the control points that define the geometry.

- Adds a new `Summary` method to all geometry types. This method gives a short
  and human readable summary of geometry values. The summary includes the
  geometry type, coordinates type, and component cardinalities where
  appropriate (e.g. number of rings in a polygon).

- Adds a new `String` method to all geometry types, implementing the
  `fmt.Stringer` interface. The method returns the same string as that returned
  by the `Summary` method.

- Adds a new `NumRings` method to the `Polygon` type. This method gives the
  total number of rings that make the polygon.

## v0.30.0

2021-07-18

- Adds `Dump` methods to `Geometry`, `MultiPoint`, `MultiLineString`,
  `MultiPolygon`, and `GeometryCollection` types. These methods break down
  composite geometries into their constituent non-multi type parts (i.e.
  `Points`, `LineStrings`, and `Polygons`) and return them as a slice.

- Fixes a bug in the `BitSet` data structure. This data structure is used to
  specify which Points within a MultiPoint are empty during manual
  construction.

## v0.29.0

2021-07-04

- Modifies error string formatting to be more consistent. Errors should treated
  opaquely by users, so this change is only cosmetic.

- Modifies the GEOS wrapper functions for testing spatial relationships between
  geometries to use direct pass through to their GEOS equivalents. These
  functions are `Equals`, `Disjoint`, `Touches`, `Contains`, `Covers`,
  `Intersects`, `Within`, `CoveredBy`, `Crosses`, and `Overlaps` (all in the
  `geos` package). Previously, these functions passed through to
  `GEOSRelatePattern_r` (which calculates a DE-9IM Intersection Matrix), and
  the intersection relationship was calculated in Go afterwards. Using a more
  direct passthrough allows the GEOS library to make better optimisations in
  same case.

## v0.28.1

2021-05-14

- Modifies the `Simplify` function to use an iterative rather than recursive
  approach. This prevents crashes for complex geometries.

## v0.28.0

2021-05-14

- Reimplements the `ConvexHull` operation using a more robust and numerically
  stable algorithm. The new algorithm is the Monotone Chain algorithm (also
  known as Andrew's Algorithm).

- Adds a new `Simplify` function that simplifies geometries using the Ramer
  Douglas Peucker algorithm.

## v0.27.0

2021-04-11

- Changes the `Boundary` method to use GEOS style behaviour rather than PostGIS
  style behaviour for empty geometries. This means that `Boundary` now always
  returns a consistent geometry type.

- Adds back the RTree `Insert` and `Delete` methods. These were previously
  removed in v0.25.1.

- Adds `Scan` methods to each concrete geometry type, so that that they
  implement the `database/sql.Scanner` interface.

## v0.26.0

2021-02-19

- Adds a `Relate` top level function, which calculates the DE-9IM matrix
  describing the relationship between two geometries.

- Adds named spatial predicate top level functions, implemented in terms of
  DE-9IM matches. These are `Equals`, `Intersects`, `Disjoint`, `Contains`,
  `CoveredBy`, `Covers`, `Overlaps`, `Touches`, `Within`, and `Crosses`.

## v0.25.1

2021-01-31

- Fixes a noding bug in DCEL operations (`Intersection`, `Union`, `Difference`,
  and `SymmetricDifference`) that occasionally caused a panic.

- Changes the strategy for joining together geometries in DCEL operations using
  "ghost lines". Geometries were previously joining using a naive radial line
  approach, but are now joining using a euclidean minimum spanning tree. This
  greatly reduces the number of additional crossing control points introduced
  by DCEL operations.

## v0.25.0

2020-12-31

- Internal refactor for WKT unmarshalling to simplify error handling.

- Performance improvements for Polygon validation (quicker 'ring in ring' check).

- Performance improvements for the LineString IsSimple method.

- Performance improvements for MultiPoint/MultiPoint intersection check.

- Adds a `Nearest` method to `RTree`, which finds the nearest entry to a given
  box.

- Performance improvements to the `Distance` and `Intersection` functions,
  based on conditionally swapping the order of the inputs.

- Adds a `Count` method to `RTree`, which gives the number of entries in the
  tree.

- Performance improvements to the MultiLineString IsSimple method.

- Performance improvements to set operations (Union, Intersection, Difference,
  SymmetricDifference).

## v0.24.0

2020-12-07

- More optimisations for DCEL operations (`Intersection`, `Union`,
  `Difference`, and `SymmetricDifference`). The improvements result in a 25%
  speed improvement.

- **Breaking change**: `Intersects` is now a top level function rather than a
  method. This is to make it consistent with other operations that act on two
  geometries.

- **Breaking change**: `ExactEquals` is now a top level function rather than a
  method. This is to make it consistent with other operations that act on two
  geometries. It was also renamed from `EqualsExact` to `ExactEquals`. This is
  inconsistent with the corresponding function in GEOS, but reads more
  fluently.

## v0.23.0

2020-12-02

- Improvements for DCEL operations (`Intersection`, `Union`, `Difference`, and
  `SymmetricDifference`). Some improvements simplify data structures and
  algorithms, and other improvements increase performance (~2x speedup).

- Fixes a compiler warning in the `geos` package.

- Internal refactor of WKB error handling to mirror the error handling strategy
  used for WKT.

## v0.22.0

2020-10-30

- Add `Intersection`, `Union`, `Difference`, and `SymmetricDifference`
  operations.

## v0.21.0

2020-10-30

- Add a `Distance` function that calculates the shortest Euclidean distance
  between two geometries.

## v0.20.0

2020-08-10

- Add area calculation options (there are initially 2 options). The first
  option causes the area calculation to return the signed area. This replaces
  the `SignedArea` methods, and so is a breaking change. The second area allows
  the geometries to be transformed inline with the area calculation.

- Add `ForceCW` and `ForceCCW` methods. These methods force areal geometries to
  have consistent winding order (orientation).

- Fix a bug in the convex hull algorithm that causes a crash in some rare
  scenarios involving almost collinear points.

- Add GEOS Buffer option wrappers. The following options are now wrapped:
  
    - The number of line segments used to represent curved parts of buffered
      geometries.

    - End-cap style (round, flat, square).

    - Join style (round, mitre, bevel).

- Add a new constructor option `OmitInvalid`. This option causes invalid
  geometries to be replaced with empty geometries upon construction rather than
  giving an error.

## v0.19.0

2020-06-27

- Fix a bug where constructor options were ignored in GeoJSON unmarshalling.

- Performance improvements to geometry validations and the Intersects
  operation (due to improvements to point-in-ring operations and RTree bulk
loading).

## v0.18.0

2020-05-30

- Improve R-Tree delete operation performance.

- Fix a bug in MultiPolygon validation.

## v0.17.0

2020-05-17

- Improve the performance of R-Tree operations (with flow on improvements to
  many algorithms, including geometry validation).

- Add a Delete method to the R-Tree implementation.

- Improve the numerical stability of the centroid calculation algorithm.

- Add a method to the R-Tree to find the boxes nearest to another box.

- Add a wrapper for the GEOS Relate function (which returns a DE-9IM code).

## v0.16.0

2020-05-08

- Add wrappers for the GEOS Difference and Symmetric Difference algorithms.

- Implement the Point On Surface algorithm, which finds a point on the interior
  of a Polygon or Polygon. This algorithm is extended to also work with point
and linear geometries.

- Improve performance of WKB marshalling and unmarshalling.

- Alters the `UnmarshalWKT` function to accept a `string` rather than an
  `io.Reader`. A new function `UnmarshalWKTFromReader` has been added that
accepts a reader. This makes the WKT interface more consistent with the WKB
interface.

## v0.15.0

2020-04-27

- Allow geometry constructor options to be passed to GEOS operations that
  produce geometries.

- Improve performance for MultiPolygon validation in cases where the child
  polygons touch at many points.

## v0.14.0

2020-04-20

- Adds an R-Tree data structure (new package,
  `github.com/peterestace/simplefeatures/rtree`). The implementation follows
the approach outlined in [R-Trees - A Dynamic Index Structure For Spatial
Searching](http://www-db.deis.unibo.it/courses/SI-LS/papers/Gut84.pdf).

- Improves some nasty worst-case performance behaviour for Polygon and
  MultiPolygon validation.

## v0.13.0

2020-04-13

- Removes the `Line` type. This was done in order to better align
  simplefeatures with the OGC Simple Feature Access specification, and allows
many edge cases in the code to be removed. Users previously using the `Line`
type should use the `LineString` type instead.

- Adds an explicit `GeometryType` type, which represents one of the 7 types of
  geometry (Point, LineString, Polygon, MultiPoint, MultiLineString,
MultiPolygon, and GeometryCollection). The `Type` method now returns a
`GeometryType` rather than a `string`.

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
