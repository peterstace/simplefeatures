# Prepared Geometries Porting Plan

21 files need to be ported (17 implementation + 4 test). All transitive
dependencies are already ported except for 3 noding files listed below.

## Porting Order

Port in this order (one file per session). Dependencies are listed first.

### Phase 1: Missing noding dependencies

No corresponding JTS test files exist for these.

| # | Java file                                      | Go file                                            |
|---|------------------------------------------------|----------------------------------------------------|
| 1 | `noding/SegmentIntersectionDetector.java`      | `noding_segment_intersection_detector.go`          |
| 2 | `noding/SegmentStringUtil.java`                | `noding_segment_string_util.go`                    |
| 3 | `noding/FastSegmentSetIntersectionFinder.java` | `noding_fast_segment_set_intersection_finder.go`   |

### Phase 2: Core prepared geometry types

| #  | Java file                                              | Go file                                                    |
|----|--------------------------------------------------------|------------------------------------------------------------|
| 4  | `geom/prep/PreparedGeometry.java`                      | `geom_prep_prepared_geometry.go`                           |
| 5  | `geom/prep/BasicPreparedGeometry.java`                 | `geom_prep_basic_prepared_geometry.go`                     |
| 6  | `geom/prep/PreparedPoint.java`                         | `geom_prep_prepared_point.go`                              |
| 7  | `geom/prep/PreparedLineString.java`                    | `geom_prep_prepared_line_string.go`                        |
| 8  | `geom/prep/PreparedLineStringIntersects.java`          | `geom_prep_prepared_line_string_intersects.go`             |
| 9  | `geom/prep/PreparedPolygon.java`                       | `geom_prep_prepared_polygon.go`                            |
| 10 | `geom/prep/PreparedPolygonPredicate.java`              | `geom_prep_prepared_polygon_predicate.go`                  |
| 11 | `geom/prep/AbstractPreparedPolygonContains.java`       | `geom_prep_abstract_prepared_polygon_contains.go`          |
| 12 | `geom/prep/PreparedPolygonContains.java`               | `geom_prep_prepared_polygon_contains.go`                   |
| 13 | `geom/prep/PreparedPolygonContainsProperly.java`       | `geom_prep_prepared_polygon_contains_properly.go`          |
| 14 | `geom/prep/PreparedPolygonCovers.java`                 | `geom_prep_prepared_polygon_covers.go`                     |
| 15 | `geom/prep/PreparedPolygonIntersects.java`             | `geom_prep_prepared_polygon_intersects.go`                 |
| 16 | `geom/prep/PreparedGeometryFactory.java`               | `geom_prep_prepared_geometry_factory.go`                   |

### Phase 3: XML test runner support

The 3 XML test files (`TestPreparedPolygonPredicate.xml`,
`TestPreparedPointPredicate.xml`, `TestPreparedPredicatesWithGeometryCollection.xml`)
are already in `xmltest/testdata/general/` but specify
`<geometryOperation>org.locationtech.jtstest.geomop.PreparedGeometryOperation</geometryOperation>`.
The test reader's `getInstance` method (in `jtstest_testrunner_test_reader.go`)
currently returns `nil` for all class lookups, so these tests silently fall back
to standard `Geometry` methods instead of exercising `PreparedGeometry`. Porting
the file below and wiring it into `getInstance` will enable them.

| #  | Java file                                              | Go file                                                    |
|----|--------------------------------------------------------|------------------------------------------------------------|
| 17 | `geomop/PreparedGeometryOperation.java`                | `jtstest_geomop_prepared_geometry_operation.go`            |

### Phase 4: Test files

| #  | Java file                                              | Go file                                                    |
|----|--------------------------------------------------------|------------------------------------------------------------|
| 18 | `geom/prep/StressTestHarness.java`                     | `geom_prep_stress_test_harness.go`                         |
| 19 | `geom/prep/PreparedGeometryTest.java`                  | `geom_prep_prepared_geometry_test.go`                      |
| 20 | `geom/prep/PreparedPolygonIntersectsStressTest.java`   | `geom_prep_prepared_polygon_intersects_stress_test.go`     |
| 21 | `geom/prep/PreparedPolygonPredicateStressTest.java`    | `geom_prep_prepared_polygon_predicate_stress_test.go`      |

## Already-ported dependencies

All transitive dependencies are already ported:

- `algorithm/locate/IndexedPointInAreaLocator.java`
- `algorithm/locate/PointOnGeometryLocator.java`
- `algorithm/locate/SimplePointInAreaLocator.java`
- `algorithm/PointLocator.java`
- `algorithm/LineIntersector.java`
- `algorithm/RobustLineIntersector.java`
- `operation/predicate/RectangleContains.java`
- `operation/predicate/RectangleIntersects.java`
- `geom/util/ComponentCoordinateExtracter.java`
- `geom/util/LinearComponentExtracter.java`
- `geom/Polygonal.java`
- `geom/Lineal.java`
- `geom/Puntal.java`
- `noding/MCIndexSegmentSetMutualIntersector.java`
- `noding/SegmentSetMutualIntersector.java`
- All core geom types (Coordinate, Geometry, LineString, etc.)

## Not needed

- `geom/prep/package-info.java` (documentation only)
- 6 perf test files in `test/jts/perf/geom/prep/` (benchmarks)
- 1 example file (not core library)
