# Test Helper Consolidation

This document describes the current state of test helper consolidation into the
`internal/test` package, and identifies remaining work. Remove sections from
this document as they are completed.

## Current state of `internal/test`

The `internal/test` package provides the following helpers:

| Helper                | Description                                     |
| ---                   | ---                                             |
| `FromWKT`             | Parse WKT string into `geom.Geometry`           |
| `FromGeoJSON`         | Parse GeoJSON string into `geom.Geometry`       |
| `ReadFile`            | Read a file, failing on error                   |
| `Eq[T]`               | Assert two comparable values are equal          |
| `GT[T]`               | Assert a value is greater than another          |
| `LT[T]`               | Assert a value is less than another             |
| `True`                | Assert a condition is true                      |
| `False`               | Assert a condition is false                     |
| `NoErr`               | Assert no error                                 |
| `Err`                 | Assert an error occurred                        |
| `ErrIs`               | Assert `errors.Is` match                        |
| `ErrAs`               | Assert `errors.As` match                        |
| `Panics`              | Assert a function panics                        |
| `ExactEquals`         | Assert two geometries are exactly equal         |
| `NotExactEquals`      | Assert two geometries are not exactly equal     |
| `ExactEqualsWKT`      | Assert a geometry exactly equals a WKT string   |
| `DeepEqual`           | Assert `reflect.DeepEqual`                      |
| `NotDeepEqual`        | Assert not `reflect.DeepEqual`                  |
| `ApproxEqual`         | Assert two float64 values are approximately eq. |
| `ImageToPNG`          | Encode an image to PNG bytes                    |

## Remaining helpers in the geom package

The `geom/util_test.go` file was deleted as part of the initial consolidation,
but several helpers remain scattered across other geom test files.

### Used in a single geom test file

These are file-local but could be consolidated for consistency:

| Helper                              | File                            |
| ---                                 | ---                             |
| `expectDumpEqWKT`                   | `geom/alg_dump_test.go`         |
| `upcastPoints/LineStrings/Polygons` | `geom/alg_dump_test.go`         |
| `expectSequenceEq`                  | `geom/dump_coordinates_test.go` |
| `xyCoords`                          | `geom/accessor_test.go`         |
| `xy`                                | `geom/validation_test.go`       |
| `checkSequence`                     | `geom/type_sequence_test.go`    |
| `regularPolygon`                    | `geom/perf_test.go`             |

### Excluded from consolidation

The following are trivial local utilities, not reusable test helpers:

- `maxInt`, `minInt` in `geom/alg_overlay_test.go`
- `minMax` in `geom/twkb_test.go`
- `newSimpleDisjointSet` in `geom/alg_disjoint_set_internal_test.go`

## Helpers in other packages

### carto

File: `carto/projections_test.go`

| Helper                    | Notes                                           |
| ---                       | ---                                             |
| `xy`                      | Trivial `geom.XY` constructor                   |
| `expectXYWithinTolerance` | Similar to `test.ApproxEqual` but for `geom.XY` |

### Duplicated `regularPolygon`

The `regularPolygon` helper is defined independently in three places:

- `geom/perf_test.go`
- `internal/perf/util_test.go`
- `internal/rawgeos/benchmark_internal_test.go`

### Not candidates for consolidation

The following packages have domain-specific test helpers that are tightly
coupled to their package internals:

- **rtree**: `testBulkLoad`, `testPopulations`, `checkSearch`,
  `checkInvariants`, `randomBox`, `checkNearest`, `checkPrioritySearch` — these
  operate on internal rtree types.
- **internal/cmprefimpl/cmppg**: `setupDB`, `loadStringsFromFile`,
  `convertToGeometries`, `isMultiPointWithEmptyPoint`, `hasLargeCoordinates` —
  these are specific to PostGIS comparison testing.
