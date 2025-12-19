# Missing Tests Report

This document identifies Java test methods from JTS that have not been ported to
the Go implementation.

## Summary

| Package     | Files Compared   | Complete   | Missing Tests   |
| ---------   | ---------------- | ---------- | --------------- |
| algorithm   | 13               | 13         | 0               |
| geom        | 18               | 18         | 0               |
| index       | 6                | 6          | 0               |
| io          | 6                | 6          | 0               |
| noding      | 7                | 7          | 0               |
| operation   | 20               | 19         | 2               |
| math        | 1                | 1          | 0               |
| planargraph | 1                | 1          | 0               |
| shape       | 1                | 1          | 0               |
| util        | 1                | 1          | 0               |
| **Total**   | **74**           | **73**     | **2**           |

## Missing Tests by Package

### operation

**operation/union/CascadedPolygonUnionTest.java**
- `testDiscs1` - Uses buffer operation to create disc geometries (buffer not
  yet ported).
- `testDiscs2` - Uses buffer operation to create disc geometries (buffer not
  yet ported).

## Notes

### Tests Not Applicable to Go

Some tests are intentionally skipped because they test Java-specific
functionality or internal implementation details. These have stub test functions
with `t.Log` explaining why they're skipped:

- **Serialization tests** - Java serialization has no direct Go equivalent.
- **SpatialIndexTester** - Internal testing utility. HPRtree's testSpatialIndex
  was ported with the SpatialIndexTester logic inlined. STRtree has basic
  coverage through public API.
- **STRtreeDemo.TestTree** - Exposes internal tree methods for testing; Go
  tests validate behavior through public API instead.

### Tests Blocked by Unported Features

- **CascadedPolygonUnionTest.testDiscs1/2** - Require buffer operation (Phase
  16 in PORTING.md).

### Complete Test Files (73 of 74)

The following test files have all Java tests fully ported to Go:

- algorithm/AngleTest.java
- algorithm/AreaTest.java
- algorithm/CGAlgorithmsDDTest.java
- algorithm/DistanceTest.java
- algorithm/IntersectionTest.java
- algorithm/LengthTest.java
- algorithm/locate/IndexedPointInAreaLocatorTest.java
- algorithm/locate/SimplePointInAreaLocatorTest.java
- algorithm/PointLocationTest.java
- algorithm/PointLocatorTest.java
- algorithm/PolygonNodeTopologyTest.java
- algorithm/RayCrossingCounterTest.java
- algorithm/RobustLineIntersectorTest.java
- geom/CoordinateArraysTest.java
- geom/CoordinateListTest.java
- geom/CoordinateSequencesTest.java
- geom/CoordinateTest.java
- geom/EnvelopeTest.java
- geom/GeometryCollectionIteratorTest.java
- geom/GeometryFactoryTest.java
- geom/GeometryOverlayTest.java
- geom/impl/CoordinateArraySequenceTest.java
- geom/impl/PackedCoordinateSequenceDoubleTest.java
- geom/impl/PackedCoordinateSequenceFloatTest.java
- geom/impl/PackedCoordinateSequenceTest.java
- geom/IntersectionMatrixTest.java
- geom/LineSegmentTest.java
- geom/PrecisionModelTest.java
- geom/TriangleTest.java
- geom/util/GeometryExtracterTest.java
- geom/util/GeometryMapperTest.java
- index/hprtree/HPRtreeTest.java
- index/intervalrtree/SortedPackedIntervalRTreeTest.java
- index/strtree/EnvelopeDistanceTest.java
- index/strtree/IntervalTest.java
- index/strtree/SIRtreeTest.java
- index/strtree/STRtreeTest.java
- io/OrdinateFormatTest.java
- io/WKBTest.java
- io/WKBReaderTest.java
- io/WKBWriterTest.java
- io/WKTReaderTest.java
- io/WKTWriterTest.java
- math/DDTest.java
- noding/NodedSegmentStringTest.java
- noding/SegmentPointComparatorTest.java
- noding/snap/SnappingNoderTest.java
- noding/snapround/HotPixelTest.java
- noding/snapround/SegmentStringNodingTest.java
- noding/snapround/SnapRoundingNoderTest.java
- noding/snapround/SnapRoundingTest.java
- operation/linemerge/LineMergerTest.java
- operation/linemerge/LineSequencerTest.java
- operation/overlayng/CoverageUnionTest.java
- operation/overlayng/ElevationModelTest.java
- operation/overlayng/LineLimiterTest.java
- operation/overlayng/OverlayGraphTest.java
- operation/overlayng/OverlayNGRobustTest.java
- operation/overlayng/OverlayNGTest.java
- operation/overlayng/PrecisionReducerTest.java
- operation/overlayng/PrecisionUtilTest.java
- operation/overlayng/RingClipperTest.java
- operation/overlayng/UnaryUnionNGTest.java
- operation/relateng/AdjacentEdgeLocatorTest.java
- operation/relateng/LinearBoundaryTest.java
- operation/relateng/PolygonNodeConverterTest.java
- operation/relateng/RelateGeometryTest.java
- operation/relateng/RelateNGTest.java
- operation/relateng/RelatePointLocatorTest.java
- operation/union/OverlapUnionTest.java
- planargraph/DirectedEdgeTest.java
- shape/fractal/HilbertCodeTest.java
- util/IntArrayListTest.java
