# Ghost Edge Construction Algorithm - Design Document

## Background

The DCEL (Doubly Connected Edge List) construction process in SimpleFeatures requires all geometry components to be connected. When performing binary operations (union, intersection, etc.) on geometries with disjoint components, "ghost edges" are inserted to connect these components into a single connected structure.

The current implementation uses a near-minimum spanning tree (MST) approach with a greedy nearest-neighbor algorithm to connect disjoint components.

## Problem Statement

The current ghost edge construction approach has two key issues:

1. **Unwanted control points**: Ghost edges frequently cross over input geometry, requiring additional control points to be inserted at crossing locations. These control points are not part of the original input and bloat the output.

2. **Numeric stability**: The additional control points and their associated crossing calculations can introduce numeric stability issues.

## Current Approach

The existing algorithm (`createGhosts` in `geom/dcel_ghosts.go`):

1. Extracts representative points from each geometry component (typically start points)
2. Builds a near-MST using:
   - R-tree for nearest neighbor searches
   - Disjoint set to prevent cycles
   - Greedy connection: each point connects to its nearest unconnected point
3. Returns a MultiLineString of ghost edges

**Limitations:**
- No consideration of where ghost edges are placed relative to input geometry
- Arbitrary point selection (start points) may not minimize edge crossings
- Greedy nearest-neighbor often creates edges that intersect input geometry

## Proposed Solution

### High-Level Strategy

Instead of minimizing total edge length (MST approach), minimize ghost edge crossings with input geometry. The key insight: by casting rays in a consistent direction (rightward) and carefully selecting connection points, we can nearly always avoid crossing existing geometry.

### Algorithm

The new algorithm operates in three phases:

#### Phase 1: Initial Renoding

Renode the two input geometries with each other. This ensures that when the geometries are overlaid, they only interact at control points (no edge-edge crossings without vertices at the crossing locations).

```
a, b, ghosts = reNodeGeometries(a, b, emptyMultiLineString)
```

This is already done in the current implementation (`geom/dcel.go:5`), so no change needed here.

#### Phase 2: Ghost Edge Construction

**Step 1: Identify Connected Components**

Use a Union-Find (disjoint set) data structure to identify which control points are already connected:

1. Initialize Union-Find with all control points from both input geometries
2. For each edge in the input geometries, union the edge's endpoints
3. The resulting sets represent disjoint components that need connecting

**Step 2: Select Representative Points**

For each component, identify the rightmost point (using y-coordinate as tiebreaker for points with the same x-coordinate).

**Step 3: Process Components Right-to-Left**

Sort components by their representative point's x-coordinate in descending order (with y-coordinate as secondary sort key for tiebreaking).

Process each component in order:

1. Cast a horizontal ray from the representative point in the positive x direction
2. Find the first intersection with any other component (if any)
3. Handle the intersection:

**Case A: Ray hits a vertex**
- Create a ghost edge directly from the origin point to the hit vertex
- Union the two components in the Union-Find structure

**Case B: Ray hits an edge**
- The edge has two endpoints; check each for obstructions
- An endpoint is **obstructed** if there exists any control point OR any edge from any input geometry that lies on the line segment between the ray origin and that endpoint
- If at least one endpoint is unobstructed:
  - Create a ghost edge to an unobstructed endpoint
  - Union the two components
- If both endpoints are obstructed:
  - Split the hit edge at the ray intersection point (creating a new control point)
  - Create a ghost edge to this new point
  - Union the two components

**Case C: No intersection**
- Create a ghost edge to a conceptual vertical line positioned to the right of all geometry
  - Calculate the maximum x-coordinate across all input geometry
  - Position the vertical line at some x-coordinate beyond this maximum (e.g., `max_x + 2`)
  - The ghost edge connects from the origin point horizontally to this vertical line

**Step 4: Construct MultiLineString**

Collect all created ghost edges into a MultiLineString.

#### Phase 3: Final Renoding

Renode the input geometries again, this time including the newly created ghost edges:

```
a, b, ghosts = reNodeGeometries(a, b, ghosts)
```

This handles the case where we had to split an edge in Phase 2, ensuring that all geometries (including ghosts) only interact at control points.

**Note:** While this technically can introduce additional control points (counter to our goal), this should be rare because:
- We only split edges when both endpoints are obstructed
- The rightward ray-casting approach naturally avoids crossing existing geometry in most cases

## Design Rationale

### Why Rightward Ray Casting?

- **Consistency**: Processing components in a deterministic order (right-to-left) ensures reproducible results
- **Minimal crossings**: By always going rightward, earlier-processed (rightmost) components create a "scaffolding" that later components can connect to without crossing
- **Simplicity**: Horizontal rays simplify intersection calculations

### Why Check Both Endpoints for Obstructions?

When a ray hits an edge, connecting to an existing endpoint (if unobstructed) avoids creating new control points. Only when both endpoints are obstructed do we need to introduce a new point.

### Why the Vertical Line Fallback?

Some components may be the absolute rightmost element. The vertical line provides a consistent connection point that:
- Maintains the invariant that all components are connected
- Doesn't introduce crossings (it's beyond all geometry)
- Can be consistently positioned deterministically

### Why Three Phases?

1. **Phase 1** ensures input geometries only interact at vertices
2. **Phase 2** creates ghost edges that rarely cross input geometry
3. **Phase 3** handles the rare case where we had to split an input edge, maintaining the vertex-only interaction invariant

## Implementation Considerations

### Data Structures

- **Union-Find**: Standard disjoint set with path compression for component tracking
- **Ray-edge intersection**: Robust geometric predicates for determining intersections
- **Obstruction checking**: Efficient spatial queries (possibly R-tree) to find obstructions

### Edge Cases

- **Single component**: If Union-Find produces only one set, no ghost edges needed
- **Collinear points**: When casting rays, handle cases where multiple vertices/edges lie on the same horizontal line
- **Vertical edges**: Edges parallel to the y-axis need special handling during ray intersection
- **Duplicate points**: The Union-Find naturally handles points at the same location
- **Empty geometries**: Should be filtered out before processing

### Numeric Stability

- Use robust geometric predicates for all intersection calculations
- When splitting edges, ensure the split point is computed consistently
- Consider tolerance for determining when points are "on" a line segment

### Performance

- Expected O(n log n) for sorting components
- Ray casting: O(n) per component with spatial indexing (R-tree)
- Union-Find: nearly O(1) amortized per operation
- Overall: O(n² log n) worst case, but typically much better with spatial indexing

## Testing Strategy

### Unit Tests

- Component identification (Union-Find correctness)
- Ray-vertex intersection cases
- Ray-edge intersection with unobstructed endpoints
- Ray-edge intersection with obstructed endpoints
- No intersection case (vertical line fallback)
- Obstruction detection

### Integration Tests

- Compare output control point counts: new vs. old algorithm
- Verify DCEL construction still works correctly
- Test with various geometry types: points, lines, polygons, multi-geometries
- Test with complex real-world geometries

### Regression Tests

- Existing binary operation tests should continue to pass
- Output geometries should be topologically equivalent (even if control points differ)

## Future Considerations

- Could we avoid the vertical line by using a different connection strategy for the rightmost component?
- Alternative directions (leftward, or alternating) might have different properties
- Could we use a more sophisticated obstruction avoidance (e.g., routing around obstructions)?

## References

- Current implementation: `geom/dcel_ghosts.go`
- DCEL construction: `geom/dcel.go`
- Union-Find: Standard algorithm, may need new implementation in this codebase
