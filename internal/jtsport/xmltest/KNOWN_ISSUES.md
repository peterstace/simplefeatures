# Known Issues

This document tracks known issues discovered while running the JTS XML test
suite against the Go port.

## Summary

Current XML test suite results:

- **Total tests:** 7803
- **Passed:** 6718 (86.1%)
- **Skipped:** 1085 (13.9%) - unsupported operations
- **Panics:** 0 (0%)
- **Errors:** 0 (0%)
- **Failed:** 0 (0%)

## Skipped Operations

The following operations are explicitly skipped as unsupported:

- `buffer`, `bufferMitredJoin` - Phase 16 (Buffer Operation) not yet ported
- `convexHull` - algorithm/ConvexHull not ported
- `densify` - Densify not ported
- `distance`, `isWithinDistance` - operation/distance/DistanceOp not ported
- `getCentroid` - algorithm/Centroid not ported
- `getInteriorPoint` - InteriorPoint not ported
- `getLength` - returns 0.0 stub, skipped to avoid false passes
- `isValid` - IsValidOp not ported
- `minClearance`, `minClearanceLine` - MinimumClearance not ported
- `polygonize` - Polygonizer not ported
- `simplifyDP`, `simplifyTP` - Simplify not ported
