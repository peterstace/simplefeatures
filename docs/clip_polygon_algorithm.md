# Polygon Clipping Algorithm

This document describes the algorithm used by `clipPolygonByRect` to clip a
polygon against an axis-aligned rectangle. It covers simple polygons (no
holes), concave polygons that split into multiple pieces, and polygons with
holes that cross the rectangle boundary.

## Overview

The algorithm has four phases:

1. **Sutherland-Hodgman clipping** — clip each ring independently against the
   rectangle's four edges.
2. **Arc extraction** — decompose each clipped ring into "interior arcs"
   (portions that pass through the rectangle's interior) and "boundary arcs"
   (portions that lie along the rectangle's edges).
3. **Topology resolution** — reconnect interior arcs using correct boundary
   paths to produce output rings.
4. **Assembly** — classify output rings, assign free holes, and build the
   output geometry.

## Phase 1: Sutherland-Hodgman Clipping

Each ring (exterior and holes) is clipped independently. The Sutherland-Hodgman
algorithm clips a closed polygon ring against a single half-plane. Four
sequential passes clip against the four rectangle edges:

1. Left:   x ≥ min.X
2. Right:  x ≤ max.X
3. Bottom: y ≥ min.Y
4. Top:    y ≤ max.Y

### Single-edge pass

The input is an open ring (a list of vertices without a repeated closing
vertex). The algorithm walks the edges of the ring. For each edge from vertex A
to vertex B, it applies one of four rules:

| A       | B       | Action                         |
| ------- | ------- | ------------------------------ |
| inside  | inside  | emit B                         |
| inside  | outside | emit intersection(A,B)         |
| outside | inside  | emit intersection(A,B), emit B |
| outside | outside | emit nothing                   |

"Inside" means the vertex satisfies the half-plane condition (e.g. x ≥ min.X
for the left edge).

The walk starts from the last vertex in the list (treating the ring as
implicitly closed), so the first edge processed is from the last vertex to the
first vertex.

### Intersection computation

For a vertical clipping edge x = k, the intersection of segment A→B is
computed as:

```
t = (k - A.x) / (B.x - A.x)
result = interpolateCoords(A, B, t)
result.x = k   // set exactly to avoid float imprecision
```

Horizontal edges (y = k) are analogous. The `interpolateCoords` function
handles Z and M dimensions using a numerically robust lerp. The explicit
assignment of the boundary coordinate ensures that later boundary detection
(which uses `==`) is reliable.

### After all four passes

Duplicate consecutive vertices (by XY) are removed, including wrap-around
(first == last). If fewer than 3 vertices remain, the ring was entirely outside
the rectangle and is discarded.

### Properties of the S-H output

The output ring is a mixture of two kinds of edges:

- **Interior edges** — edges where the path passes through the rectangle's
  interior (at least one endpoint is not on the rectangle boundary).
- **Boundary edges** — edges where both endpoints lie on the same rectangle
  edge.

The S-H algorithm preserves the winding direction of the input ring.

For convex polygons, the output is always a valid simple ring. For concave
polygons, the output ring may **self-touch** along the rectangle boundary. For
example, a C-shaped polygon clipped across its opening produces a ring where
two boundary arcs overlap on the same rectangle edge. The topology resolution
phase (Phase 3) handles this.

## Phase 2: Arc Extraction

Each clipped ring is decomposed into **interior arcs**. An interior arc is a
maximal sequence of consecutive interior edges. Each arc starts and ends at a
vertex on the rectangle boundary.

### Identifying boundary edges

An edge between vertices P and Q is a boundary edge if both P and Q lie on the
same rectangle edge:

- both P.x == min.X and Q.x == min.X (left edge), or
- both P.x == max.X and Q.x == max.X (right edge), or
- both P.y == min.Y and Q.y == min.Y (bottom edge), or
- both P.y == max.Y and Q.y == max.Y (top edge).

### Extracting arcs

Walk the ring starting from a boundary edge. Skip consecutive boundary edges.
When a non-boundary edge is encountered, begin an interior arc at the current
vertex. Continue collecting vertices until the next boundary edge is
encountered. The arc ends at the vertex where the boundary edge begins.

Each arc records:

- Its coordinate sequence
- Its **start parameter** and **end parameter** on the rectangle boundary (see
  parameterisation below)

### Special cases

- **No boundary edges at all**: the entire ring is interior (no contact with
  the rectangle boundary). This happens when a hole is entirely inside the
  rectangle and has no vertices on the rectangle boundary. Such a ring is
  classified as a "free hole" and is not decomposed into arcs. It will be
  assigned to an output polygon in Phase 4. A hole whose envelope is covered
  by the rectangle but that has vertices on the rectangle boundary is **not**
  free — it must be clipped to avoid producing invalid polygons with shared
  edges between the exterior and hole rings.

- **No interior edges at all**: the entire ring is boundary. For the exterior
  ring, this means the original polygon completely contains the rectangle. No
  arcs are extracted and in Phase 3 the output is the rectangle itself. For a
  hole ring, this means the hole covers the entire rectangle. No polygon area
  survives — the result is an empty polygon.

## Phase 3: Topology Resolution

### Rectangle boundary parameterisation

The rectangle boundary is parameterised counter-clockwise starting from the
bottom-left corner:

```
Bottom edge (left to right): parameter 0 to W
Right edge  (bottom to top): parameter W to W+H
Top edge    (right to left): parameter W+H to 2W+H
Left edge   (top to bottom): parameter 2W+H to 2W+2H (= perimeter)
```

Where W = max.X - min.X and H = max.Y - min.Y. The parameter wraps: the
bottom-left corner has parameter 0 and also parameter 2W+2H.

A point on the boundary maps to a unique parameter based on which edge it lies
on. Corner points are assigned to the edge they appear on first in the CCW
traversal (e.g. the bottom-right corner has parameter W, assigned to the start
of the right edge).

### Winding normalisation

Before topology resolution, the clipped exterior ring is normalised to CCW. If
it has negative signed area (CW), it is reversed. Hole rings are **not**
reversed during this step — each hole's winding is handled independently. After
topology resolution, if the exterior was reversed, all output rings are
reversed back to restore the original winding convention.

This means the topology resolution algorithm only needs to handle the CCW
exterior case.

### Preparing arcs

Interior arcs from the exterior ring are used as-is.

Interior arcs from hole rings may or may not need reversal, depending on the
hole's own winding direction. In the simplefeatures data model, interior rings
are not required to have any particular orientation relative to the exterior
ring — a hole may be CW or CCW independently.

The rule: after the exterior is normalised to CCW, a hole arc must be reversed
if and only if the hole ring is CCW (positive signed area). This is because a
CCW ring has its enclosed area to the left — for a hole, that enclosed area is
the empty space, so the polygon interior is to the right. Reversal puts the
polygon interior to the left, matching the CCW output convention. A CW hole
ring already has the polygon interior to the left and its arcs are used as-is.
This is determined by computing the signed area of each clipped hole ring
independently.

Reversed arcs have their start and end parameters **swapped** (so that the
"output start" is the arc's original end, and vice versa) and are marked with a
`reverse` flag so the walking algorithm traverses their coordinates backwards.

### The walking algorithm

All arcs (from the exterior and from holes) are collected. Each arc has an
"output start" parameter and an "output end" parameter on the rectangle
boundary.

The algorithm produces output rings by walking:

1. Pick any unused arc. Call it the "first arc" of this ring.
2. Traverse the arc's coordinates (reversed if flagged). The ring is now at the
   arc's output end parameter on the boundary.
3. Find the **next arc**: the arc whose output start parameter is the next one
   going CCW from the current position on the boundary. This is the arc with
   the smallest positive CCW distance from the current end parameter.
4. Build a **boundary path** from the current end parameter to the next arc's
   start parameter. This path follows the rectangle boundary CCW and includes
   any rectangle corners in between.
5. If the next arc is the first arc, the ring is complete. Otherwise, go to
   step 2 with the next arc.
6. Repeat from step 1 for any remaining unused arcs.

### Building boundary paths

Given a start parameter p₁ and end parameter p₂, the boundary path includes
any rectangle corners whose parameters lie strictly between p₁ and p₂ going
CCW.

For example, if p₁ = 3 (on the bottom edge) and p₂ = 10 (on the top edge) for
a rectangle with W=4, H=2 (corners at 0, 4, 6, 10):

- Corners at parameters 4 and 6 are between 3 and 10.
- The path is: corner at param 4 (bottom-right), corner at param 6
  (top-right).

The boundary path does not include the start and end points themselves (those
are already the last and first coordinates of the adjacent arcs).

For Z/M coordinates, corner values are linearly interpolated between the Z/M
values at the start and end of the boundary path, proportional to the boundary
distance.

### Why this produces correct output

The key invariant is: **along the rectangle boundary going CCW, arc endpoints
alternate between "ends" and "starts"**. This is because each arc enters the
interior from the boundary and returns to the boundary. Between consecutive
end/start pairs, the boundary segment is part of the output polygon's boundary.

For exterior arcs, the boundary between an arc's end and the next arc's start
represents a portion of the rectangle boundary that replaces the part of the
original polygon that extended outside the rectangle.

For hole arcs (reversed), the boundary between the reversed arc's end and the
next arc's start represents a portion of the rectangle boundary that was
"inside the polygon but outside the hole" — i.e., it's still part of the
polygon.

When a concave polygon self-touches after S-H clipping (e.g., a C-shape
clipped across its opening), the overlapping boundary arcs are replaced by
correct pairings that produce two separate output rings.

## Phase 4: Assembly

### Ring classification

All output rings from topology resolution are exterior rings (CCW after
normalisation). Boundary-crossing holes create concavities or split the
exterior into multiple pieces, but never produce new holes. This is a property
of rectangle clipping: a hole that crosses the boundary "opens up" onto the
boundary, becoming part of the exterior's boundary rather than enclosing a
separate hole.

Holes in the output come only from "free holes" — holes that are entirely
inside the rectangle and have no contact with the boundary.

### Free hole assignment

Each free hole is assigned to the output exterior ring that contains it. A
point from the hole (its first vertex) is tested against each output ring using
a ray-casting point-in-polygon test.

### Output construction

- If there are no output rings: return an empty polygon.
- If there is one output ring (with any free holes): return a Polygon.
- If there are multiple output rings: return a MultiPolygon.

## Fast Paths

Before running the full algorithm, several fast paths are checked:

1. **Empty polygon**: return empty polygon.
2. **Degenerate rectangle** (point or line envelope): return empty polygon. A
   polygon intersected with a lower-dimensional shape produces at most a
   lower-dimensional result, which is discarded.
3. **Disjoint envelopes**: return empty polygon.
4. **Polygon entirely inside rectangle**: return the polygon unchanged.

## Worked Example: C-shaped Polygon

Input polygon (CCW):

```
(0, 2.5) → (3, 2.5) → (3, 2.8) → (0.5, 2.8) →
(0.5, 3.2) → (3, 3.2) → (3, 3.5) → (0, 3.5)
```

Rectangle: (1, 2) to (5, 4).

### Phase 1: S-H clipping

Only the left edge (x ≥ 1) produces changes. The other three edges leave all
vertices unchanged.

After clipping:

```
(1, 2.5) → (3, 2.5) → (3, 2.8) → (1, 2.8) →
(1, 3.2) → (3, 3.2) → (3, 3.5) → (1, 3.5)
```

### Phase 2: Arc extraction

Edge classification (boundary = both endpoints on same rect edge):

| Edge              | Type     |
| ----------------- | -------- |
| (1,2.5) → (3,2.5) | interior |
| (3,2.5) → (3,2.8) | interior |
| (3,2.8) → (1,2.8) | interior |
| (1,2.8) → (1,3.2) | boundary |
| (1,3.2) → (3,3.2) | interior |
| (3,3.2) → (3,3.5) | interior |
| (3,3.5) → (1,3.5) | interior |
| (1,3.5) → (1,2.5) | boundary |

Interior arcs:

- **Arc 1**: (1,2.5) → (3,2.5) → (3,2.8) → (1,2.8). Start param = 11.5, end
  param = 11.2.
- **Arc 2**: (1,3.2) → (3,3.2) → (3,3.5) → (1,3.5). Start param = 10.8, end
  param = 10.5.

### Phase 3: Topology resolution

No hole arcs. Signed area is positive (CCW), so no reversal needed.

Sorted arc endpoints on the boundary (going CCW from param 0):

- 10.5 — Arc 2 end
- 10.8 — Arc 2 start
- 11.2 — Arc 1 end
- 11.5 — Arc 1 start

Walking:

**Ring A**: Start with Arc 2. Traverse (1,3.2) → (3,3.2) → (3,3.5) → (1,3.5).
End at param 10.5. Next start going CCW: param 10.8 (Arc 2's own start). Build
boundary path from 10.5 to 10.8: no corners between them (both on left edge).
Arc 2 is the first arc, so the ring is complete: (1,3.2), (3,3.2), (3,3.5),
(1,3.5).

**Ring B**: Start with Arc 1. Traverse (1,2.5) → (3,2.5) → (3,2.8) → (1,2.8).
End at param 11.2. Next start going CCW: param 11.5 (Arc 1's own start). Build
boundary path from 11.2 to 11.5: no corners. Ring complete: (1,2.5), (3,2.5),
(3,2.8), (1,2.8).

### Phase 4: Assembly

Two output rings → MultiPolygon:

```
MULTIPOLYGON(
  ((1 2.5, 3 2.5, 3 2.8, 1 2.8, 1 2.5)),
  ((1 3.2, 3 3.2, 3 3.5, 1 3.5, 1 3.2))
)
```

## Worked Example: Polygon Containing Rectangle (with hole)

Input polygon:

- Exterior (CCW): (-2, -2) → (6, -2) → (6, 6) → (-2, 6) — contains the
  entire rectangle.
- Hole (CW): (-1, 1) → (2, 1) → (2, 3) → (-1, 3) — crosses left edge of
  rectangle.

Rectangle: (0, 0) to (4, 4).

### Phase 1: S-H clipping

Exterior clips to the rectangle itself (all boundary): (0,0), (4,0), (4,4),
(0,4).

Hole clips to: (0,1), (2,1), (2,3), (0,3).

### Phase 2: Arc extraction

Exterior: no interior arcs (all boundary).

Hole edges:

| Edge          | Type     |
| ------------- | -------- |
| (0,1) → (2,1) | interior |
| (2,1) → (2,3) | interior |
| (2,3) → (0,3) | interior |
| (0,3) → (0,1) | boundary |

Hole interior arc: (0,1) → (2,1) → (2,3) → (0,3). Start param = 15 (on left
edge), end param = 13 (on left edge).

### Phase 3: Topology resolution

Exterior signed area is positive (CCW), no reversal. The hole arc is CW. Its
start/end are swapped for the CCW output: output start = 13 (original end),
output end = 15 (original start).

Only one arc. Walking:

Start with the hole arc. Since `fromHole = true`, traverse in reverse: (0,3) →
(2,3) → (2,1) → (0,1). End at output end param = 15. Next start going CCW:
param 13 (the same arc). Build boundary path from param 15 to param 13: this
wraps around the entire rectangle. Corners at params 0, 4, 8, 12 are all
between 15 and 13 (going CCW past 16/0). Path: (0,0), (4,0), (4,4), (0,4).
Ring complete.

Output ring: (0,3), (2,3), (2,1), (0,1), (0,0), (4,0), (4,4), (0,4).

### Phase 4: Assembly

Single output ring. The rectangle with a concavity where the hole was:

```
POLYGON((0 3, 2 3, 2 1, 0 1, 0 0, 4 0, 4 4, 0 4, 0 3))
```
