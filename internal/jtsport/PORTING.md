# JTS Porting Progress

## Completed Phases (1-15)

| Phase | Description                              | Files | Status                           |
| ----- | ---------------------------------------- | ----- | -------------------------------- |
| 1     | Foundation (util/, math/)                | 4     | Done                             |
| 2     | Core Geometry Types (geom/, geom/impl/)  | 35    | Done                             |
| 3     | Algorithms (algorithm/)                  | 17    | Done                             |
| 4     | Spatial Index (index/, strtree/, chain/) | 19    | Done                             |
| 5     | Point Location (intervalrtree/, locate/) | 8     | Done                             |
| 6     | Geometry Utilities (geom/util/)          | 11    | Done                             |
| 7     | Geometry Graph (geomgraph/)              | 30    | Done                             |
| 8     | Overlay Operations (operation/)          | 35    | Done                             |
| 9     | I/O (io/)                                | 6     | Done                             |
| 10    | Relate Operations (operation/relate/)    | 8     | Done                             |
| 11    | Noding (noding/, hprtree/)               | 16    | Done                             |
| 12    | IsSimpleOp (operation/valid/)            | 1     | Done                             |
| 13    | RelateNG (operation/relateng/)           | 26    | Done                             |
| 14    | WKB I/O (io/)                            | 12    | Done                             |
| 15    | OverlayNG (operation/overlayng/)         | ~57   | Done                             |

The Go implementation supports switching between RelateOp (old) and RelateNG (new)
via `Geom_GeometryRelate_SetRelateImpl("ng")`. Default is RelateOp.

The Go implementation supports switching between SnapIfNeededOverlayOp (old) and
OverlayNGRobust (new) via `Geom_GeometryOverlay_SetOverlayImpl("ng")`. Default is old.

---

## XML Test Suite

The JTS XML test suite runner has been ported. This allows running JTS's XML-based
test cases against the Go port to verify correctness.

**Ported files:**

| Java File                              | Go File                                    | Status |
| -------------------------------------- | ------------------------------------------ | ------ |
| `geomop/GeometryOperation.java`        | `jtstest_geomop_geometry_operation.go`     | Ported |
| `geomop/GeometryMethodOperation.java`  | `jtstest_geomop_geometry_method_operation.go` | Ported |
| `testrunner/Result.java`               | `jtstest_testrunner_result.go`             | Ported |
| `testrunner/BooleanResult.java`        | `jtstest_testrunner_boolean_result.go`     | Ported |
| `testrunner/DoubleResult.java`         | `jtstest_testrunner_double_result.go`      | Ported |
| `testrunner/IntegerResult.java`        | `jtstest_testrunner_integer_result.go`     | Ported |
| `testrunner/GeometryResult.java`       | `jtstest_testrunner_geometry_result.go`    | Ported |
| `testrunner/ResultMatcher.java`        | `jtstest_testrunner_result_matcher.go`     | Ported |
| `testrunner/EqualityResultMatcher.java`| `jtstest_testrunner_equality_result_matcher.go` | Ported |
| `testrunner/BufferResultMatcher.java`  | `jtstest_testrunner_buffer_result_matcher.go` | Ported |
| `testrunner/TestParseException.java`   | `jtstest_testrunner_test_parse_exception.go` | Ported |
| `testrunner/Test.java`                 | `jtstest_testrunner_tst.go`                | Ported |
| `testrunner/TestCase.java`             | `jtstest_testrunner_test_case.go`          | Ported |
| `testrunner/TestRun.java`              | `jtstest_testrunner_test_run.go`           | Ported |
| `testrunner/TestReader.java`           | `jtstest_testrunner_test_reader.go`        | Ported |
| `geomop/TestCaseGeometryFunctions.java`| `jtstest_geomop_test_case_geometry_functions.go` | Ported |

**Test harness:** `xmltest/runner_test.go` (Go test that exercises the ported runner)

**Current results:**

| Metric  | Count | Percentage |
| ------- | ----- | ---------- |
| Total   | 7803  | 100%       |
| Passed  | 6718  | 86.1%      |
| Skipped | 1085  | 13.9%      |
| Panics  | 0     | 0%         |
| Errors  | 0     | 0%         |
| Failed  | 0     | 0%         |

**Skipped operations:** `buffer`, `bufferMitredJoin`, `convexHull`, `densify`,
`distance`, `getCentroid`, `getInteriorPoint`, `getLength`, `isValid`,
`isWithinDistance`, `minClearance`, `minClearanceLine`, `polygonize`,
`simplifyDP`, `simplifyTP`.

---

## Phase 16: Buffer Operation (Pending)

Computes the buffer (dilation/erosion) of a geometry for positive and negative
distances. The buffer operation produces a polygonal result representing all
points within a specified distance of the input geometry.

**Features:**
- **Positive/negative buffers**: Dilate or erode geometries
- **End cap styles**: Round, flat, or square line endings
- **Join styles**: Round, mitre, or bevel corners
- **Single-sided buffers**: Buffer on one side of lines only
- **Robust computation**: Falls back to snap-rounding for difficult cases

**Dependencies:** Requires noding infrastructure from Phase 15a-prereq
(IntersectionAdder, ScaledNoder, and snapround files).

### Phase 16a: Additional Noding Prerequisites

| Java File                         | Go File                           | Description                      | Status  |
| --------------------------------- | --------------------------------- | -------------------------------- | ------- |
| `noding/ScaledNoder.java`         | `noding_scaled_noder.go`          | Wraps noder with integer scaling | Pending |
| `noding/FastNodingValidator.java` | `noding_fast_noding_validator.go` | Validates noding (replace stub)  | Pending |

### Phase 16b: Buffer Core Classes

| Java File                                         | Go File                                           | Description                     | Status  |
| ------------------------------------------------- | ------------------------------------------------- | ------------------------------- | ------- |
| `operation/buffer/BufferParameters.java`          | `operation_buffer_buffer_parameters.go`           | Buffer configuration parameters | Pending |
| `operation/buffer/OffsetSegmentString.java`       | `operation_buffer_offset_segment_string.go`       | Stores offset curve points      | Pending |
| `operation/buffer/OffsetSegmentGenerator.java`    | `operation_buffer_offset_segment_generator.go`    | Generates offset segments       | Pending |
| `operation/buffer/OffsetCurveBuilder.java`        | `operation_buffer_offset_curve_builder.go`        | Computes raw offset curves      | Pending |
| `operation/buffer/BufferInputLineSimplifier.java` | `operation_buffer_buffer_input_line_simplifier.go`| Simplifies input lines          | Pending |
| `operation/buffer/BufferCurveSetBuilder.java`     | `operation_buffer_buffer_curve_set_builder.go`    | Creates offset curve set        | Pending |
| `operation/buffer/RightmostEdgeFinder.java`       | `operation_buffer_rightmost_edge_finder.go`       | Finds rightmost coordinate      | Pending |
| `operation/buffer/BufferSubgraph.java`            | `operation_buffer_buffer_subgraph.go`             | Connected subgraph handling     | Pending |
| `operation/buffer/SubgraphDepthLocater.java`      | `operation_buffer_subgraph_depth_locater.go`      | Locates subgraph depth          | Pending |
| `operation/buffer/BufferBuilder.java`             | `operation_buffer_buffer_builder.go`              | Builds buffer geometry          | Pending |
| `operation/buffer/BufferOp.java`                  | `operation_buffer_buffer_op.go`                   | Main entry point                | Pending |

### Phase 16c: Buffer Utilities (Optional)

| Java File                                    | Go File                                      | Description              | Status  |
| -------------------------------------------- | -------------------------------------------- | ------------------------ | ------- |
| `operation/buffer/OffsetCurve.java`          | `operation_buffer_offset_curve.go`           | Offset curve computation | Pending |
| `operation/buffer/OffsetCurveSection.java`   | `operation_buffer_offset_curve_section.go`   | Curve section handling   | Pending |
| `operation/buffer/SegmentMCIndex.java`       | `operation_buffer_segment_mc_index.go`       | Segment index for curves | Pending |
| `operation/buffer/VariableBuffer.java`       | `operation_buffer_variable_buffer.go`        | Variable-width buffer    | Pending |

### Phase 16d: Buffer Validation (Optional)

| Java File                                                    | Go File                                                      | Description               | Status  |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------- | ------- |
| `operation/buffer/validate/PointPairDistance.java`           | `operation_buffer_validate_point_pair_distance.go`           | Point pair distance       | Pending |
| `operation/buffer/validate/DistanceToPointFinder.java`       | `operation_buffer_validate_distance_to_point_finder.go`      | Point distance finder     | Pending |
| `operation/buffer/validate/BufferCurveMaximumDistanceFinder.java` | `operation_buffer_validate_buffer_curve_max_distance_finder.go` | Max distance finder   | Pending |
| `operation/buffer/validate/BufferDistanceValidator.java`     | `operation_buffer_validate_buffer_distance_validator.go`     | Validates distances       | Pending |
| `operation/buffer/validate/BufferResultValidator.java`       | `operation_buffer_validate_buffer_result_validator.go`       | Validates buffer result   | Pending |

### Phase 16e: Wire Up Geometry.buffer()

Update `Geom_Geometry` to expose buffer operation via `Buffer(distance)` method.

**Notes:**
- ~17 files total for core buffer (2 noding prereqs + 11 buffer core + 4 utilities)
- ~22 files if including validation
- BufferOp is the main entry point
- Buffer uses the existing PolygonBuilder from operation/overlay (Phase 8)
