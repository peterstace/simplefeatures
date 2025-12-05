# Bug Report: Difference Operation Incorrectly Returns Empty

## Problem Summary

The `geom.Difference(A, B)` operation was incorrectly returning an empty
geometry when it should have returned a non-empty result. This occurred with
specific polygon inputs where polygon B contained an "almost-spike" - two
consecutive vertices that were extremely close together (~3e-8 degrees apart).

### Reproducing Inputs

**Polygon A** (minimized):
```
POLYGON((-87.62444917095299 41.39729331748811,
         -87.62446704222914 41.397301659153754,
         -87.62371745215871 41.397807336601254,
         -87.62444917095299 41.39729331748811))
```

**Polygon B** (minimized, contains near-coincident points):
```
POLYGON((-87.6242938761274 41.39744938185984,
         -87.62418667793823 41.3977256149851,
         -87.62418671102395 41.3977256212699,    <- P1
         -87.62407665709884 41.39770472967701,   <- P2 (very close to P1)
         -87.62391637672572 41.39738765504086,
         -87.6242938761274 41.39744938185984))
```

The vertices P1 `(-87.62418671102395, 41.3977256212699)` and P2
`(-87.62407665709884, 41.39770472967701)` are approximately 3e-8 degrees apart
in one coordinate.

### Expected vs Actual Result

- **Expected**: A MULTIPOLYGON representing the parts of A not covered by B
- **Actual**: Empty geometry

## Investigation Process

### Step 1: Delta Debugging to Minimize Inputs

The original inputs were complex polygons with many vertices. A delta debugging
approach was used to systematically remove vertices while preserving the bug,
resulting in the minimized test case above.

### Step 2: Tracing the DCEL Construction

The set operations in simplefeatures use a Doubly Connected Edge List (DCEL)
data structure. Debug prints were added to trace:

1. **Re-noding**: How geometries are split at intersection points
2. **Ghost creation**: How auxiliary edges are added to ensure a valid planar
   subdivision
3. **Face cycle detection**: How faces are identified from edge cycles
4. **Interaction point detection**: How vertices that need special handling are
   identified

### Step 3: Identifying the Degenerate Structure

The debug output revealed that vertex P2 had only **1 incident edge** instead
of the expected 2 or more. This violated the DCEL invariant that every vertex
must have at least 2 incident edges (one incoming, one outgoing).

Key observation from debug output:
```
Vertex (-87.62418671102395, 41.3977256212699): 1 incident edges
```

This degenerate structure caused the face extraction to fail with "no rings to
extract".

### Step 4: Tracing the Re-noding Step

Further investigation of the re-noding phase revealed the root cause:

1. The point×line re-noding step checks if any vertex from the geometry is
   close to any line segment
2. Point P1 was being detected as "close to" line segment P2→V3
3. This happened because P1 ≈ P2, so P1 was indeed very close to a line
   starting at P2
4. P1 was added as a "cut point" on line P2→V3
5. This created an edge P1→V3, but the edge P2→P1 was essentially zero-length
6. The resulting DCEL structure was degenerate

## Root Cause

The `appendCutsForPointXLine` function in `dcel_re_noding.go` was adding cut
points without checking if the point was essentially the same as an endpoint.

Original code:
```go
appendCutsForPointXLine := func(ln line, cuts []XY) []XY {
    ptIndex.tree.RangeSearch(ln.box(), func(i int) error {
        xy := ptIndex.points[i]
        if !ln.hasEndpoint(xy) && distBetweenXYAndLine(xy, ln) < ulp*0x200 {
            cuts = append(cuts, xy)
        }
        return nil
    })
    return cuts
}
```

The check `!ln.hasEndpoint(xy)` only checked for exact equality. When P1 and P2
were very close but not exactly equal, P1 passed this check and was added as a
cut to line P2→V3.

## The Fix

### Initial Attempt (Too Aggressive)

The first fix added a relative distance check that skipped cuts within 0.1% of
the line length from an endpoint. This was applied to both:

1. `appendCutsForPointXLine` (point×line re-noding)
2. `appendNewNode` (line×line re-noding)

However, this broke existing tests (test 77 and 79) because it was too
aggressive for line×line intersections, where mathematically computed
intersection points should not be filtered out.

### Final Fix (Targeted)

The fix was refined to only apply to point×line re-noding:

```go
appendCutsForPointXLine := func(ln line, cuts []XY) []XY {
    ptIndex.tree.RangeSearch(ln.box(), func(i int) error {
        xy := ptIndex.points[i]
        // Skip if xy is an endpoint.
        if ln.hasEndpoint(xy) {
            return nil
        }
        // Skip if xy is very close to an endpoint relative to the line
        // length. This prevents degenerate cuts when near-coincident
        // points exist (e.g. a polygon with an almost-spike where two
        // consecutive vertices are extremely close together).
        lineLenSq := ln.a.distanceSquaredTo(ln.b)
        distToASq := xy.distanceSquaredTo(ln.a)
        distToBSq := xy.distanceSquaredTo(ln.b)
        // Skip if distance to endpoint is less than 1e-3 (0.1%) of line length.
        minEndpointDistSq := lineLenSq * 1e-6 // (1e-3)^2
        if distToASq < minEndpointDistSq || distToBSq < minEndpointDistSq {
            return nil
        }
        // Also skip based on ULP threshold.
        if distToASq < thresholdSq || distToBSq < thresholdSq {
            return nil
        }
        if distBetweenXYAndLine(xy, ln) < threshold {
            cuts = append(cuts, xy)
        }
        return nil
    })
    return cuts
}
```

### Key Insight

The fix distinguishes between two types of re-noding:

1. **Point×line re-noding**: Checks if existing vertices lie on line segments.
   This can be affected by near-coincident points and needs the relative
   distance filter.

2. **Line×line re-noding**: Computes mathematical intersections between line
   segments. These are precise and should not be filtered by relative distance.

## Files Changed

### `geom/dcel_re_noding.go`

Added relative distance check to `appendCutsForPointXLine` to skip adding cut
points that are within 0.1% of the line length from an endpoint.

### `geom/alg_set_op_test.go`

Added regression test case:
```go
{
    // Regression test for bug where Difference incorrectly returned empty
    // due to near-coincident points in polygon B causing degenerate
    // re-noding.
    input1:  "POLYGON((-87.62444917095299 41.39729331748811,...))",
    input2:  "POLYGON((-87.6242938761274 41.39744938185984,...))",
    fwdDiff: "MULTIPOLYGON(...)",
}
```

## Verification

All tests pass after the fix:
- The new regression test (test 80) passes
- Previously passing tests (including 77 and 79) continue to pass
- Full `go test ./geom` passes

## Lessons Learned

1. **Near-coincident points are a common source of numerical geometry bugs**.
   When two points are very close but not exactly equal, they can cause
   unexpected behavior in algorithms that check for exact equality.

2. **Relative thresholds are often more robust than absolute thresholds**.
   Using a percentage of the line length (0.1%) works across different scales
   of geometry, while absolute thresholds may be too small for some inputs and
   too large for others.

3. **Different phases of an algorithm may need different tolerance handling**.
   The point×line re-noding needed filtering that line×line re-noding did not,
   because they have different numerical characteristics.

4. **Delta debugging is invaluable for minimizing complex test cases**.
   The original inputs had dozens of vertices; the minimized case has only 3-5
   vertices per polygon, making the bug much easier to understand and debug.
