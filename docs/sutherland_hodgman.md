# Sutherland-Hodgman Polygon Clipping Algorithm

## Motivation

ClipByRect computes the intersection of a geometry with an axis-aligned
rectangle. While this can be computed using a general-purpose overlay algorithm,
a dedicated rectangle clipper is significantly faster because it exploits the
simple structure of the clipping region.

The Sutherland-Hodgman algorithm is the classic approach. It clips a polygon
against a convex clipping region by decomposing the problem into a sequence of
simpler clips, one for each edge of the clipping region. For an axis-aligned
rectangle, this means four passes.

## Core Idea

A convex polygon can be described as the intersection of half-planes. An
axis-aligned rectangle is the intersection of four half-planes:

```
left:   x >= xmin
right:  x <= xmax
bottom: y >= ymin
top:    y <= ymax
```

Clipping a polygon against the rectangle is equivalent to clipping it against
each of these four half-planes in sequence. The output of each pass becomes the
input to the next.

```
input polygon
    |
    v
[clip to left edge]
    |
    v
[clip to right edge]
    |
    v
[clip to bottom edge]
    |
    v
[clip to top edge]
    |
    v
output polygon
```

The order of the four passes does not matter. The result is the same regardless
of which edge is processed first.

## Clipping Against a Single Half-Plane

Each pass walks the edges of the input polygon. An edge is a segment from
vertex A to the next vertex B (wrapping from the last vertex back to the
first).

For each edge A to B, there are exactly four cases depending on whether A and B
are inside or outside the half-plane:

### Case 1: Both Inside

```
        |
   A--->B      A is inside, B is inside.
        |
        |      Emit: B
```

Both vertices are on the retained side. The entire edge survives. Emit B. (A
was already emitted by the previous edge, or will be handled as the first
vertex.)

### Case 2: Inside to Outside

```
        |
   A--->+--B   A is inside, B is outside.
        |
        |      Emit: intersection point I
```

The edge crosses from the retained side to the clipped side. Emit the
intersection point where the segment crosses the clipping edge. B is discarded.

### Case 3: Outside to Inside

```
        |
   B<---+--A   A is outside, B is inside.
        |
        |      Emit: intersection point I, then B
```

The edge crosses from the clipped side to the retained side. Emit the
intersection point, then emit B. Two vertices are produced for this edge
because the output polygon needs to include both the point where it re-enters
the clipping boundary and the destination vertex.

### Case 4: Both Outside

```
        |
        |  A-->B   A is outside, B is outside.
        |
        |          Emit: nothing
```

The entire edge is on the clipped side. It contributes nothing to the output.

### Summary Table

| A       | B       | Emitted vertices |
| ------- | ------- | ---------------- |
| inside  | inside  | B                |
| inside  | outside | I                |
| outside | inside  | I, B             |
| outside | outside | (none)           |

Where I is the intersection of segment AB with the clipping edge.

## Inside Test

For each of the four edges of the rectangle, the inside test is a single
comparison:

| Clipping edge     | Condition for "inside" |
| ----------------- | ---------------------- |
| left (x = xmin)   | point.x >= xmin        |
| right (x = xmax)  | point.x <= xmax        |
| bottom (y = ymin) | point.y >= ymin        |
| top (y = ymax)    | point.y <= ymax        |

## Computing Intersections

Because the clipping edges are axis-aligned, intersection computation reduces
to simple linear interpolation.

### Intersection With a Vertical Edge (x = k)

Given segment from A to B and a vertical clipping edge at x = k:

```
t = (k - A.x) / (B.x - A.x)
I = (k, A.y + t * (B.y - A.y))
```

The parameter t is the fractional distance along the segment from A to B where
it crosses the edge. Since we only compute this when A and B are on opposite
sides of the edge, B.x - A.x is guaranteed to be non-zero.

### Intersection With a Horizontal Edge (y = k)

Given segment from A to B and a horizontal clipping edge at y = k:

```
t = (k - A.y) / (B.y - A.y)
I = (A.x + t * (B.x - A.x), k)
```

## Worked Example

Clip the triangle with vertices P0(0, 1), P1(4, 3), P2(2, -1) against a
rectangle with xmin=1, xmax=3, ymin=0, ymax=2.

### Pass 1: Clip to Left Edge (x >= 1)

Process each edge of the triangle:

**Edge P0(0,1) to P1(4,3):** P0 is outside (x=0 < 1), P1 is inside (x=4 >= 1).
Case 3: emit intersection, then P1.

```
t = (1 - 0) / (4 - 0) = 0.25
I_a = (1, 1 + 0.25 * (3 - 1)) = (1, 1.5)
```

Emit: (1, 1.5), (4, 3).

**Edge P1(4,3) to P2(2,-1):** Both inside (x=4 >= 1, x=2 >= 1). Case 1: emit
P2.

Emit: (2, -1).

**Edge P2(2,-1) to P0(0,1):** P2 is inside (x=2 >= 1), P0 is outside (x=0 <
1). Case 2: emit intersection.

```
t = (1 - 2) / (0 - 2) = 0.5
I_b = (1, -1 + 0.5 * (1 - (-1))) = (1, 0)
```

Emit: (1, 0).

**Result after pass 1:** (1, 1.5), (4, 3), (2, -1), (1, 0).

### Pass 2: Clip to Right Edge (x <= 3)

Input: (1, 1.5), (4, 3), (2, -1), (1, 0).

**Edge (1, 1.5) to (4, 3):** Inside to outside. Emit intersection.

```
t = (3 - 1) / (4 - 1) = 2/3
I = (3, 1.5 + 2/3 * (3 - 1.5)) = (3, 2.5)
```

Emit: (3, 2.5).

**Edge (4, 3) to (2, -1):** Outside to inside. Emit intersection, then (2, -1).

```
t = (3 - 4) / (2 - 4) = 0.5
I = (3, 3 + 0.5 * (-1 - 3)) = (3, 1)
```

Emit: (3, 1), (2, -1).

**Edge (2, -1) to (1, 0):** Both inside. Emit (1, 0).

Emit: (1, 0).

**Edge (1, 0) to (1, 1.5):** Both inside. Emit (1, 1.5).

Emit: (1, 1.5).

**Result after pass 2:** (3, 2.5), (3, 1), (2, -1), (1, 0), (1, 1.5).

### Pass 3: Clip to Bottom Edge (y >= 0)

Input: (3, 2.5), (3, 1), (2, -1), (1, 0), (1, 1.5).

**Edge (3, 2.5) to (3, 1):** Both inside. Emit (3, 1).

**Edge (3, 1) to (2, -1):** Inside to outside. Emit intersection.

```
t = (0 - 1) / (-1 - 1) = 0.5
I = (3 + 0.5 * (2 - 3), 0) = (2.5, 0)
```

Emit: (2.5, 0).

**Edge (2, -1) to (1, 0):** Outside to inside. Emit intersection, then (1, 0).

```
t = (0 - (-1)) / (0 - (-1)) = 1
I = (2 + 1 * (1 - 2), 0) = (1, 0)
```

Emit: (1, 0), (1, 0).

Note: The intersection coincides with (1, 0). This produces a duplicate vertex,
which is harmless and can be cleaned up afterward.

**Edge (1, 0) to (1, 1.5):** Both inside. Emit (1, 1.5).

Emit: (1, 1.5).

**Edge (1, 1.5) to (3, 2.5):** Both inside. Emit (3, 2.5).

Emit: (3, 2.5).

**Result after pass 3:** (3, 1), (2.5, 0), (1, 0), (1, 0), (1, 1.5), (3, 2.5).

### Pass 4: Clip to Top Edge (y <= 2)

Input: (3, 1), (2.5, 0), (1, 0), (1, 0), (1, 1.5), (3, 2.5).

**Edge (3, 1) to (2.5, 0):** Both inside. Emit (2.5, 0).

**Edge (2.5, 0) to (1, 0):** Both inside. Emit (1, 0).

**Edge (1, 0) to (1, 0):** Both inside. Emit (1, 0).

**Edge (1, 0) to (1, 1.5):** Both inside. Emit (1, 1.5).

**Edge (1, 1.5) to (3, 2.5):** Inside to outside. Emit intersection.

```
t = (2 - 1.5) / (2.5 - 1.5) = 0.5
I = (1 + 0.5 * (3 - 1), 2) = (2, 2)
```

Emit: (2, 2).

**Edge (3, 2.5) to (3, 1):** Outside to inside. Emit intersection, then (3, 1).

```
t = (2 - 2.5) / (1 - 2.5) = 1/3
I = (3 + 1/3 * (3 - 3), 2) = (3, 2)
```

Emit: (3, 2), (3, 1).

**Final result:** (2.5, 0), (1, 0), (1, 0), (1, 1.5), (2, 2), (3, 2), (3, 1).

After removing the duplicate vertex at (1, 0):

**(2.5, 0), (1, 0), (1, 1.5), (2, 2), (3, 2), (3, 1).**

This is the triangle clipped to the rectangle.

## Pseudocode

```
function clipPolygonToRect(vertices, xmin, ymin, xmax, ymax):
    output = vertices
    output = clipToEdge(output, LEFT,   xmin)
    output = clipToEdge(output, RIGHT,  xmax)
    output = clipToEdge(output, BOTTOM, ymin)
    output = clipToEdge(output, TOP,    ymax)
    return output

function clipToEdge(vertices, edge, value):
    if len(vertices) == 0:
        return []

    output = []
    A = vertices[last]

    for each B in vertices:
        aInside = isInside(A, edge, value)
        bInside = isInside(B, edge, value)

        if aInside and bInside:
            append B to output
        else if aInside and not bInside:
            append intersection(A, B, edge, value) to output
        else if not aInside and bInside:
            append intersection(A, B, edge, value) to output
            append B to output
        // else: both outside, emit nothing

        A = B

    return output

function isInside(point, edge, value):
    switch edge:
        LEFT:   return point.x >= value
        RIGHT:  return point.x <= value
        BOTTOM: return point.y >= value
        TOP:    return point.y <= value

function intersection(A, B, edge, value):
    switch edge:
        LEFT, RIGHT:
            t = (value - A.x) / (B.x - A.x)
            return (value, A.y + t * (B.y - A.y))
        BOTTOM, TOP:
            t = (value - A.y) / (B.y - A.y)
            return (A.x + t * (B.x - A.x), value)
```

## Complexity

- **Time:** O(N) per clipping edge, where N is the number of vertices. With 4
  edges, the total is O(4N) = O(N). In the worst case, each pass can at most
  double the vertex count (every edge crosses the clipping line), so the
  intermediate vertex lists remain bounded.

- **Space:** O(N) for the output vertex list. The algorithm can be implemented
  with two buffers that are swapped between passes.

## Properties

- **Preserves winding order.** If the input polygon has counter-clockwise
  winding, the output will too.

- **Works for concave polygons.** Unlike some clipping algorithms,
  Sutherland-Hodgman handles concave (non-convex) input polygons. However, when
  a concave polygon is clipped, the result may contain coincident (overlapping)
  edges along the clipping boundary. These are topologically valid but may need
  post-processing depending on the application.

- **May produce degenerate edges.** When a polygon vertex lies exactly on the
  clipping edge, the intersection point coincides with the vertex, producing
  zero-length edges or duplicate vertices. A post-processing step to remove
  duplicate consecutive vertices handles this.

## Extension to LineStrings

Sutherland-Hodgman is designed for closed polygonal rings. For open LineStrings,
a related but simpler approach works: clip each segment independently against
the rectangle (using Cohen-Sutherland or Liang-Barsky line clipping), then
merge consecutive surviving segments into output LineStrings. Segments that are
entirely outside produce gaps, potentially splitting one LineString into
multiple.

## Extension to Other Geometry Types

- **Point:** Test whether the point lies within the rectangle. Emit it or
  discard it.

- **MultiPoint:** Test each point individually.

- **LineString:** Clip each segment against the rectangle. Consecutive surviving
  segments form output LineStrings. A single input LineString may produce
  multiple output LineStrings (yielding a MultiLineString).

- **Polygon (no holes):** Clip the exterior ring using Sutherland-Hodgman. If
  the result is empty, the polygon is entirely outside the rectangle.

- **Polygon (with holes):** Clip the exterior ring and each hole ring
  independently. Discard hole rings that become empty. Hole rings that are
  entirely inside the rectangle survive unchanged. This is explained in
  detail below.

- **Multi-geometries:** Process each component independently and collect the
  results.

- **GeometryCollection:** Recurse into each element.

## Handling Polygons With Holes

Polygons with holes require care because holes can interact with the clipping
rectangle in ways that change the topology.

### Case 1: Hole Entirely Inside the Rectangle

The hole survives unchanged. No special handling needed.

```
+--rect-----------+
|                  |
|   +-exterior-+   |
|   |          |   |
|   |  +-hole+ |   |
|   |  |     | |   |
|   |  +-----+ |   |
|   |          |   |
|   +----------+   |
|                  |
+------------------+
```

### Case 2: Hole Entirely Outside the Rectangle

The hole is irrelevant to the clipped result. Discard it.

### Case 3: Hole Crosses the Rectangle Boundary

This is the complex case. When a hole's ring crosses the clipping boundary,
parts of the clipping boundary become part of the polygon's boundary.

```
+-rect-----+
|           |
|   +--exterior--+
|   |       |    |
|   |  +--hole---+--+
|   |  |    |       |
|   |  +----+-------+
|   |       |    |
|   +-------+----+
|           |
+-----------+
```

In this situation, clipping the exterior ring and hole ring independently and
then naively combining them will not produce a valid polygon. The clipped hole
shares edges along the rectangle boundary with the clipped exterior ring,
effectively splitting the polygon.

There are two strategies for handling this:

**Strategy A: Clip rings then resolve topology.** Clip each ring
independently, then use a topology-aware algorithm to merge the clipped rings
into a valid polygon or multipolygon. This is essentially a simplified overlay
operation restricted to the rectangle boundary.

**Strategy B: Fallback to general intersection.** Detect when holes cross the
rectangle boundary and fall back to a general-purpose intersection algorithm
for those cases. This avoids implementing the topology resolution but sacrifices
the performance advantage for these inputs.

### Topology Resolution for Strategy A

When holes cross the rectangle boundary, the clipped result may need to be a
MultiPolygon rather than a single Polygon. For example, a U-shaped polygon
clipped by a rectangle across the opening can produce two separate polygons.

The resolution algorithm:

1. Clip the exterior ring and all hole rings independently.
2. Identify clipped ring segments that lie along the rectangle boundary.
3. Pair up boundary segments: where the exterior ring enters the boundary, a
   hole ring (or the same ring) must exit, and vice versa.
4. Walk the combined ring structure, tracing the outline of each resulting
   polygon.
5. Determine which resulting rings are exteriors and which are holes based on
   winding order (or signed area).
6. Group holes with their enclosing exterior rings.

This is the most algorithmically complex part of implementing ClipByRect. It
resembles the edge-tracing phase of an overlay algorithm, but is constrained to
only four possible boundary edges, which simplifies the data structures
involved.

## Unit Test Cases

This section enumerates unit test cases for ClipByRect. The clipping rectangle
is denoted R. Each test specifies the input geometry, the rectangle, and the
expected output.

Each test case is defined with concrete coordinates in a single canonical
orientation. The test implementation systematically applies all 8 symmetry
transformations of the dihedral group D4 (4 rotations × 2 reflections) to both
the input geometry and the clipping rectangle, generating 8 sub-tests per case.
This ensures every case is tested against all edges and corners of R without
needing to list each orientation explicitly.

### Output Type Rules

Two rules govern the output geometry type:

1. **Dimension preservation.** The topological dimension of the output always
   matches the input. A LineString input never produces a Point output; it
   produces an empty LineString instead.

2. **Multi-type promotion only.** A singular type may be promoted to its
   multi-type when clipping produces multiple components (e.g. a LineString that
   is split into multiple segments becomes a MultiLineString). However, a
   multi-type input is never demoted: a MultiPoint input always produces a
   MultiPoint output, even if only one point survives.

The output type only ever stays the same or gets wider (singular to multi), never
narrower. For multi-type inputs, the output type is always the same as the input
type. For singular-type inputs, the output is the same singular type unless the
result has multiple components, in which case it is promoted to the corresponding
multi-type.

When a child element of a multi-type is clipped away entirely, it is omitted from
the result (not included as an empty geometry). For example, a MultiPoint with
three points where one is outside R produces a MultiPoint with two points. In
contrast, a singular type that is clipped away entirely becomes an empty geometry
of the same type: a Point outside R produces an empty Point.

GeometryCollections are processed recursively. Each non-GeometryCollection child
is clipped independently per the rules above. Children that clip to empty (or
were already empty) are omitted from the result. Nested GeometryCollections that
become empty after their children are omitted are themselves omitted. The
top-level GeometryCollection is never omitted: if all children are removed, the
result is an empty GeometryCollection.

### Rectangle Configurations

The clipping rectangle can take four forms:

- **Normal rectangle** (positive width and height): The standard case. All
  geometry-type tests below assume a normal rectangle.

- **Empty envelope**: Always returns an empty geometry of the input type.

- **Point envelope** (min = max): Degenerate. Lower-dimensional intersections
  are discarded. A Point at the same location survives, but a LineString passing
  through the point produces an empty LineString (the intersection is a point,
  which is lower-dimensional than a LineString).

- **Line envelope** (zero width or zero height): Degenerate. Lower-dimensional
  intersections are discarded. A LineString collinear with the line survives, but
  a LineString that merely crosses it produces an empty LineString (the
  intersection is a point). Similarly, a Polygon crossing the line produces an
  empty Polygon (the intersection is at most a line, which is lower-dimensional
  than a Polygon).

The Degenerate Rectangle Cases section below exercises the empty, point, and line
envelope configurations with representative geometries.

### Point

| #   | Case                     | Expected output |
| --- | ------------------------ | --------------- |
| PT1 | Empty Point              | Empty Point     |
| PT2 | Point strictly inside R  | Same Point      |
| PT3 | Point strictly outside R | Empty Point     |
| PT4 | Point on edge of R       | Same Point      |
| PT5 | Point on corner of R     | Same Point      |

### MultiPoint

| #   | Case                                    | Expected output                     |
| --- | --------------------------------------- | ----------------------------------- |
| MP1 | Empty MultiPoint                        | Empty MultiPoint                    |
| MP2 | All points inside R                     | Same MultiPoint                     |
| MP3 | All points outside R                    | Empty MultiPoint                    |
| MP4 | Some points inside, some outside        | MultiPoint with only inside points  |
| MP5 | Single point inside R                   | MultiPoint with 1 point             |
| MP6 | Points on edges and corners of R        | All retained                        |
| MP7 | Mix of inside, on-boundary, and outside | Inside and boundary points retained |
| MP8 | MultiPoint containing empty points      | Empty points excluded from output   |

### LineString

#### Spatial relationship to R

| #   | Case                                                | Expected output                                 |
| --- | --------------------------------------------------- | ----------------------------------------------- |
| LS1 | Empty LineString                                    | Empty LineString                                |
| LS2 | Entirely inside R                                   | Same LineString                                 |
| LS3 | Entirely outside R                                  | Empty LineString                                |
| LS4 | Entirely outside R on one side (e.g. all left of R) | Empty LineString                                |
| LS5 | Crosses R, entering and exiting once                | LineString clipped to entry and exit points     |
| LS6 | Crosses R, entering and exiting multiple times      | MultiLineString of surviving segments           |
| LS7 | One endpoint inside, one outside                    | LineString from inside endpoint to intersection |
| LS8 | One endpoint outside, one inside                    | LineString from intersection to inside endpoint |
| LS9 | Both endpoints outside, segment passes through R    | LineString clipped to two intersection points   |

#### Boundary interactions

| #    | Case                                                                        | Expected output                                |
| ---- | --------------------------------------------------------------------------- | ---------------------------------------------- |
| LS10 | Endpoint exactly on edge of R, other inside                                 | LineString preserved                           |
| LS11 | Endpoint exactly on corner of R, other inside                               | LineString preserved                           |
| LS12 | Both endpoints on boundary of R                                             | LineString preserved                           |
| LS13 | LineString lies entirely along one edge of R                                | LineString preserved (collinear with boundary) |
| LS14 | LineString lies entirely along two adjacent edges (L-shaped along boundary) | LineString preserved                           |
| LS15 | Segment touches corner of R but does not enter (V-shape touching corner)    | Empty LineString                               |
| LS16 | Segment touches edge of R tangentially (parallel approach, touches, leaves) | Empty LineString                               |

#### Direction and shape

| #    | Case                                                             | Expected output                                   |
| ---- | ---------------------------------------------------------------- | ------------------------------------------------- |
| LS17 | Diagonal line crossing two edges of R                            | LineString with 2 intersection points             |
| LS18 | Axis-aligned line crossing two opposite edges of R               | LineString clipped to opposite edge intersections |
| LS19 | Closed LineString (ring) inside R                                | Same closed LineString                            |
| LS20 | Closed LineString (ring) partially overlapping R                 | LineString of surviving arcs                      |
| LS21 | Multi-segment LineString with some segments inside, some outside | MultiLineString of surviving segments             |
| LS22 | Zigzag LineString entering and exiting R many times              | MultiLineString with multiple components          |

#### Corner and edge-coincident cases

| #    | Case                                                 | Expected output                              |
| ---- | ---------------------------------------------------- | -------------------------------------------- |
| LS23 | Vertex exactly on R boundary, adjacent edges inside  | LineString preserved through boundary vertex |
| LS24 | Vertex exactly on R boundary, adjacent edges outside | Empty LineString                             |
| LS25 | LineString passes through two opposite corners of R  | LineString clipped between corners           |

### MultiLineString

| #    | Case                                                    | Expected output                                  |
| ---- | ------------------------------------------------------- | ------------------------------------------------ |
| MLS1 | Empty MultiLineString                                   | Empty MultiLineString                            |
| MLS2 | All component LineStrings inside R                      | Same MultiLineString                             |
| MLS3 | All component LineStrings outside R                     | Empty MultiLineString                            |
| MLS4 | Some components inside, some outside                    | MultiLineString with surviving components        |
| MLS5 | Single component, clipped to a single segment           | MultiLineString with 1 component                 |
| MLS6 | One component crosses R, another is inside R            | MultiLineString with clipped and unclipped parts |
| MLS7 | Component that produces multiple fragments when clipped | Fragments included in output MultiLineString     |
| MLS8 | MultiLineString containing empty LineStrings            | Empty components excluded                        |

### Polygon (No Holes)

#### Spatial relationship to R

| #   | Case                                                       | Expected output                |
| --- | ---------------------------------------------------------- | ------------------------------ |
| PG1 | Empty Polygon                                              | Empty Polygon                  |
| PG2 | Entirely inside R                                          | Same Polygon                   |
| PG3 | Entirely outside R (no overlap)                            | Empty Polygon                  |
| PG4 | R entirely inside Polygon                                  | Polygon equal to R             |
| PG5 | Partially overlapping one edge of R                        | Polygon clipped to R boundary  |
| PG6 | Partially overlapping two adjacent edges (corner clip)     | Polygon clipped at corner      |
| PG7 | Partially overlapping two opposite edges (strip through R) | Polygon clipped on two sides   |
| PG8 | Partially overlapping three edges                          | Polygon clipped on three sides |

#### Boundary interactions

| #    | Case                                                  | Expected output                     |
| ---- | ----------------------------------------------------- | ----------------------------------- |
| PG9  | Polygon shares an entire edge with R                  | Polygon preserved along shared edge |
| PG10 | Polygon vertex exactly on R edge                      | Polygon with vertex on boundary     |
| PG11 | Polygon vertex exactly on R corner                    | Polygon with vertex on corner       |
| PG12 | Polygon edge collinear with R edge, polygon inside R  | Polygon preserved                   |
| PG13 | Polygon edge collinear with R edge, polygon outside R | Empty Polygon                       |
| PG14 | Polygon touches R at a single point (vertex-to-edge)  | Empty Polygon                       |
| PG15 | Polygon touches R at a single corner point            | Empty Polygon                       |

#### Shape variations

| #    | Case                                              | Expected output                             |
| ---- | ------------------------------------------------- | ------------------------------------------- |
| PG16 | Convex polygon clipped by R                       | Convex clipped polygon                      |
| PG17 | Concave polygon, concavity inside R               | Concave clipped polygon                     |
| PG18 | Concave polygon, concavity facing R boundary      | Clipped polygon reflecting concavity        |
| PG19 | U-shaped polygon clipped across the opening       | MultiPolygon (two separate pieces)          |
| PG20 | Very thin sliver polygon partially inside R       | Thin clipped polygon                        |
| PG21 | Triangle clipped to produce various vertex counts | Polygon with 3-7 vertices depending on clip |

#### Winding order

| #    | Case                            | Expected output                                      |
| ---- | ------------------------------- | ---------------------------------------------------- |
| PG22 | Counter-clockwise exterior ring | Output preserves winding order (reflections test CW) |

### Polygon (With Holes)

#### Hole entirely inside R

| #   | Case                                                  | Expected output                  |
| --- | ----------------------------------------------------- | -------------------------------- |
| PH1 | Exterior and hole both inside R                       | Same Polygon with hole           |
| PH2 | Exterior clipped, hole entirely inside clipped region | Polygon with hole preserved      |
| PH3 | Multiple holes, all inside R                          | Polygon with all holes preserved |

#### Hole entirely outside R

| #   | Case                                                         | Expected output                        |
| --- | ------------------------------------------------------------ | -------------------------------------- |
| PH4 | Hole entirely outside R (but inside exterior ring outside R) | Polygon without hole (hole discarded)  |
| PH5 | Hole in part of exterior ring that is clipped away           | Polygon without hole (hole irrelevant) |

#### Hole crossing R boundary

| #    | Case                                                                 | Expected output                            |
| ---- | -------------------------------------------------------------------- | ------------------------------------------ |
| PH6  | Hole crosses one edge of R                                           | Polygon with concavity (no hole in output) |
| PH7  | Hole crosses two adjacent edges of R (corner)                        | Polygon with concavity (no hole in output) |
| PH8  | Hole crosses two opposite edges of R (splits polygon)                | MultiPolygon                               |
| PH9  | Hole crosses all four edges of R, leaving corner fragments           | MultiPolygon                               |
| PH10 | Hole inside R shares edge with R boundary, exterior extends beyond R | Polygon with concavity (no hole in output) |

#### Multiple holes interacting with R

| #    | Case                                       | Expected output                         |
| ---- | ------------------------------------------ | --------------------------------------- |
| PH11 | One hole inside R, another outside R       | Polygon with only inside hole           |
| PH12 | One hole inside R, another splits polygon  | MultiPolygon, hole in correct component |
| PH13 | Multiple holes each crossing one edge of R | Polygon with multiple concavities       |

#### Topology-changing cases

| #    | Case                                          | Expected output |
| ---- | --------------------------------------------- | --------------- |
| PH14 | Hole splits polygon into three or more pieces | MultiPolygon    |
| PH15 | Hole covers entire clipped area               | Empty Polygon   |

### MultiPolygon

| #     | Case                                                                | Expected output                          |
| ----- | ------------------------------------------------------------------- | ---------------------------------------- |
| MPG1  | Empty MultiPolygon                                                  | Empty MultiPolygon                       |
| MPG2  | All component Polygons inside R                                     | Same MultiPolygon                        |
| MPG3  | All component Polygons outside R                                    | Empty MultiPolygon                       |
| MPG4  | Some components inside, some outside                                | MultiPolygon with surviving components   |
| MPG5  | One component partially clipped, another fully inside               | MultiPolygon combining both              |
| MPG6  | One component becomes multiple polygons when clipped (e.g. U-shape) | MultiPolygon with all resulting polygons |
| MPG7  | Multiple components, each partially clipped                         | MultiPolygon with all fragments          |
| MPG8  | Components with holes, some holes clipped                           | MultiPolygon preserving relevant holes   |
| MPG9  | Single component fully inside R                                     | MultiPolygon with 1 component            |
| MPG10 | MultiPolygon containing empty Polygons                              | Empty components excluded                |

### GeometryCollection

| #    | Case                                                | Expected output                                 |
| ---- | --------------------------------------------------- | ----------------------------------------------- |
| GC1  | Empty GeometryCollection                            | Empty GeometryCollection                        |
| GC2  | Contains only Points, all inside R                  | GeometryCollection with all Points              |
| GC3  | Contains only Points, all outside R                 | Empty GeometryCollection                        |
| GC4  | Contains mixed types, all inside R                  | GeometryCollection with all components          |
| GC5  | Contains mixed types, all outside R                 | Empty GeometryCollection                        |
| GC6  | Contains mixed types, some inside, some outside     | GeometryCollection with surviving components    |
| GC7  | Contains Point, LineString, and Polygon             | Each component clipped independently            |
| GC8  | Contains nested GeometryCollection                  | Recursive clipping into nested collection       |
| GC9  | Contains nested GeometryCollection with mixed types | Recursive clipping at all levels                |
| GC10 | All child geometries are empty                      | Empty GeometryCollection                        |
| GC11 | Contains MultiPoint, MultiLineString, MultiPolygon  | Each multi-type component clipped independently |
| GC12 | Deeply nested GeometryCollections (3+ levels)       | Recursive clipping at all levels                |
| GC13 | Nested GC whose children all clip to empty          | Nested GC omitted from parent                   |

### Degenerate Rectangle Cases

These cases test non-standard rectangles with representative geometries from
each type.

| #    | Case                 | Geometry                                | Expected output    |
| ---- | -------------------- | --------------------------------------- | ------------------ |
| DR1  | Empty envelope       | Point inside would-be area              | Empty Point        |
| DR2  | Empty envelope       | LineString                              | Empty LineString   |
| DR3  | Empty envelope       | Polygon                                 | Empty Polygon      |
| DR4  | Point envelope (0x0) | Point at same location                  | Same Point         |
| DR5  | Point envelope (0x0) | Point at different location             | Empty Point        |
| DR6  | Point envelope (0x0) | MultiPoint with some points at location | Clipped MultiPoint |
| DR7  | Point envelope (0x0) | MultiPoint with no points at location   | Empty MultiPoint   |
| DR8  | Point envelope (0x0) | LineString through that point           | Empty LineString   |
| DR9  | Point envelope (0x0) | Polygon containing that point           | Empty Polygon      |
| DR10 | Line envelope        | Point on the line                       | Same Point         |
| DR11 | Line envelope        | Point off the line                      | Empty Point        |
| DR12 | Line envelope        | MultiPoint with some points on the line | Clipped MultiPoint |
| DR13 | Line envelope        | MultiPoint with no points on the line   | Empty MultiPoint   |
| DR14 | Line envelope        | LineString crossing the line            | Empty LineString   |
| DR15 | Line envelope        | LineString collinear with line          | Clipped LineString |
| DR16 | Line envelope        | Polygon crossing the line               | Empty Polygon      |

### Numerical Edge Cases

| #   | Case                                                               | Notes                                      |
| --- | ------------------------------------------------------------------ | ------------------------------------------ |
| NE1 | Very large coordinates (near float64 max)                          | Test numerical stability                   |
| NE2 | Very small coordinates (near float64 epsilon)                      | Test precision                             |
| NE3 | Negative coordinates                                               | All quadrants should work                  |
| NE4 | Coordinates that produce intersection parameters t very close to 0 | Near-vertex intersection                   |
| NE5 | Coordinates that produce intersection parameters t very close to 1 | Near-vertex intersection                   |
| NE6 | Polygon vertex exactly at intersection point with R edge           | Degenerate intersection (duplicate vertex) |
| NE7 | Segment nearly parallel to R edge (small angle)                    | Intersection computation precision         |
| NE8 | Zero-length segments in input (duplicate consecutive vertices)     | Should not cause division by zero          |

### Coordinate Dimension Preservation

| #   | Case                  | Notes                                                       |
| --- | --------------------- | ----------------------------------------------------------- |
| CD1 | XY geometry clipped   | Output has XY coordinates                                   |
| CD2 | XYZ geometry clipped  | Output has XYZ coordinates, Z interpolated at intersections |
| CD3 | XYM geometry clipped  | Output has XYM coordinates, M interpolated at intersections |
| CD4 | XYZM geometry clipped | Output has XYZM coordinates, Z and M interpolated           |
