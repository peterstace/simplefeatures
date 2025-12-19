# GeometryCollection Overlay Operations Implementation Plan

This document describes the plan for implementing full GeometryCollection support
in simplefeatures overlay operations (Union, Intersection, Difference,
SymmetricDifference) using the ported JTS Overlay NG algorithms.

## Background

### Simplefeatures Semantics

In simplefeatures, a GeometryCollection semantically represents the union of its
components. This means:

- `GC{A, B}` is equivalent to `Union(A, B)`
- Operations on a GC should treat it as a combined shape

### JTS Overlay NG Restrictions

JTS Overlay NG has strict limitations on GeometryCollection inputs:

1. **Homogeneity requirement**: All elements must have the same dimension
   (0=Points, 1=Lines, 2=Polygons).

2. **Validation location**: `EdgeNodingBuilder.addGeometryCollection()` checks
   dimension consistency and throws "Overlay input is mixed-dimension" if
   violated.

3. **Why**: The noding and topology labeling algorithms assume consistent
   dimensionality.

### What JTS CAN Handle

- **Unary union** (`OverlayNGRobust.union`) handles mixed-dimension GCs by
  processing each dimension separately and combining results.

- **Binary operations** on non-GC geometries handle mixed-dimension results
  correctly (e.g., `Intersection(Polygon, Line)` returns a LineString).

## Current State

| Function              | Implementation | GC Support |
|-----------------------|----------------|------------|
| `Union`               | JTS            | Yes        |
| `UnaryUnion`          | DCEL `setOp()` | No         |
| `UnionMany`           | `UnaryUnion`   | No         |
| `Intersection`        | DCEL `setOp()` | No         |
| `Difference`          | DCEL `setOp()` | No         |
| `SymmetricDifference` | DCEL `setOp()` | No         |

## Target State

| Function              | Implementation                         | GC Support |
|-----------------------|----------------------------------------|------------|
| `Union`               | JTS (no change)                        | Yes        |
| `UnaryUnion`          | JTS `unionUnary()`                     | Yes        |
| `UnionMany`           | `UnaryUnion` (no change)               | Yes        |
| `Intersection`        | JTS with GC decomposition              | Yes        |
| `Difference`          | JTS with GC decomposition              | Yes        |
| `SymmetricDifference` | Composition: `Union(Diff(A,B), Diff(B,A))` | Yes    |

## Mathematical Foundation

### GeometryCollection Semantics

A GeometryCollection semantically represents the **union** of its components:

```
GC{A₁, A₂, ..., Aₙ} ≡ A₁ ∪ A₂ ∪ ... ∪ Aₙ
```

This means any operation on a GC must respect this equivalence.

### Intersection with GeometryCollections

Using set theory, intersection distributes over union:

```
(A₁ ∪ A₂) ∩ (B₁ ∪ B₂) = (A₁ ∩ B₁) ∪ (A₁ ∩ B₂) ∪ (A₂ ∩ B₁) ∪ (A₂ ∩ B₂)
```

More generally:

```
(⋃ᵢ Aᵢ) ∩ (⋃ⱼ Bⱼ) = ⋃ᵢⱼ (Aᵢ ∩ Bⱼ)
```

This is a **cartesian product** of intersections: for every pair `(Aᵢ, Bⱼ)`,
compute their intersection, then union all results.

### Difference with GeometryCollections

Set difference does **not** distribute symmetrically. Instead:

**When A is a union:**
```
(A₁ ∪ A₂) - B = (A₁ - B) ∪ (A₂ - B)
```

**When B is a union:**
```
A - (B₁ ∪ B₂) = (A - B₁) - B₂
```

Combining these for `GC{A₁, A₂} - GC{B₁, B₂}`:

```
(A₁ ∪ A₂) - (B₁ ∪ B₂)
= ((A₁ - B₁) - B₂) ∪ ((A₂ - B₁) - B₂)
```

This means: for each element of A, subtract all elements of B **sequentially**,
then union the results.

### Symmetric Difference

Symmetric difference can be expressed in terms of union and difference:

```
A △ B = (A - B) ∪ (B - A)
```

This allows us to implement it using the other operations.

## Implementation Strategy

### Core Pattern

For GC-aware binary operations:

1. Handle empty inputs (return `Geometry{}`).
2. If either input is a GC, union its components via `unionUnary()`. Both
   inputs must be unioned independently before proceeding.
3. After union, either result may still be a GC if it contained mixed
   dimensions. Extract the parts and apply the operation element-wise.
4. Combine results via union.
5. Normalize result if needed.

### Why Union First?

When an input is a GC like `GC{Polygon1, Polygon2}` where the polygons overlap:

- Semantically, this represents `Union(Polygon1, Polygon2)`
- We must union first to get the correct combined shape before applying the
  operation
- Without this, overlapping regions would be processed incorrectly

### Mixed-Dimension GCs After Union

`unionUnary()` can return a GC when the input has mixed dimensions:

- Input: `GC{Polygon, Line}` (or even nested `GC{GC{Polygon}, Line}`)
- Output: `GC{Polygon, Line}` (elements may still be GCs in edge cases)

In this case, we use `extractParts()` to recursively extract all non-GC
elements before applying the operation.

## Implementation Details

### 1. UnaryUnion

Simple change to use JTS instead of DCEL:

```go
func UnaryUnion(g Geometry) (Geometry, error) {
    if g.IsEmpty() {
        return Geometry{}, nil
    }
    return unionUnary(g)
}
```

### 2. Helper Function

```go
// extractParts recursively extracts all non-GC elements from a geometry.
// If g is not a GeometryCollection, returns a slice containing just g.
// If g is a GeometryCollection, recursively extracts from all children.
func extractParts(g Geometry) []Geometry {
    if !g.IsGeometryCollection() {
        return []Geometry{g}
    }
    gc := g.MustAsGeometryCollection()
    var parts []Geometry
    for i := range gc.NumGeometries() {
        parts = append(parts, extractParts(gc.GeometryN(i))...)
    }
    return parts
}
```

### 3. Intersection

The algorithm computes the Cartesian product of intersections:

```
Intersection(GC{A₁, A₂}, GC{B₁, B₂})
= Union(A₁∩B₁, A₁∩B₂, A₂∩B₁, A₂∩B₂)
```

```go
func Intersection(a, b Geometry) (Geometry, error) {
    if a.IsEmpty() || b.IsEmpty() {
        return Geometry{}, nil
    }
    // Normalize GC inputs by unioning their parts. Both inputs must be
    // unioned independently before proceeding, to ensure overlapping
    // components within each GC are treated as combined shapes.
    if a.IsGeometryCollection() {
        var err error
        a, err = unionUnary(a)
        if err != nil {
            return Geometry{}, err
        }
    }
    if b.IsGeometryCollection() {
        var err error
        b, err = unionUnary(b)
        if err != nil {
            return Geometry{}, err
        }
    }

    // Extract non-GC parts from each input.
    partsA := extractParts(a)
    partsB := extractParts(b)

    // Compute Cartesian product of intersections.
    var results []Geometry
    for _, partA := range partsA {
        for _, partB := range partsB {
            result, err := intersectionBinary(partA, partB)
            if err != nil {
                return Geometry{}, err
            }
            if !result.IsEmpty() {
                results = append(results, result)
            }
        }
    }
    if len(results) == 0 {
        return Geometry{}, nil
    }
    return unionUnary(NewGeometryCollection(results).AsGeometry())
}

// intersectionBinary performs intersection using JTS for non-GC inputs.
func intersectionBinary(a, b Geometry) (Geometry, error) {
    var result Geometry
    err := catch(func() error {
        wktReader := jts.Io_NewWKTReader()
        jtsA, err := wktReader.Read(a.AsText())
        if err != nil {
            return wrap(err, "converting geometry A to JTS")
        }
        jtsB, err := wktReader.Read(b.AsText())
        if err != nil {
            return wrap(err, "converting geometry B to JTS")
        }
        jtsResult := jtsA.Intersection(jtsB)
        result, err = UnmarshalWKT(jtsResult.ToText())
        return wrap(err, "converting JTS intersection result to simplefeatures")
    })
    return result, err
}
```

### 4. Difference

The algorithm subtracts each part of B from each part of A sequentially:

```
Difference(GC{A₁, A₂}, GC{B₁, B₂})
= Union(((A₁ - B₁) - B₂), ((A₂ - B₁) - B₂))
```

```go
func Difference(a, b Geometry) (Geometry, error) {
    if a.IsEmpty() {
        return Geometry{}, nil
    }
    if b.IsEmpty() {
        return UnaryUnion(a)
    }
    // Normalize GC inputs by unioning their parts. Both inputs must be
    // unioned independently before proceeding, to ensure overlapping
    // components within each GC are treated as combined shapes.
    if a.IsGeometryCollection() {
        var err error
        a, err = unionUnary(a)
        if err != nil {
            return Geometry{}, err
        }
    }
    if b.IsGeometryCollection() {
        var err error
        b, err = unionUnary(b)
        if err != nil {
            return Geometry{}, err
        }
    }

    // Extract non-GC parts from each input.
    partsA := extractParts(a)
    partsB := extractParts(b)

    // For each part of A, subtract all parts of B sequentially.
    var results []Geometry
    for _, partA := range partsA {
        result := partA
        for _, partB := range partsB {
            var err error
            result, err = differenceBinary(result, partB)
            if err != nil {
                return Geometry{}, err
            }
            if result.IsEmpty() {
                break
            }
        }
        if !result.IsEmpty() {
            results = append(results, result)
        }
    }
    if len(results) == 0 {
        return Geometry{}, nil
    }
    return unionUnary(NewGeometryCollection(results).AsGeometry())
}

// differenceBinary performs difference using JTS for non-GC inputs.
func differenceBinary(a, b Geometry) (Geometry, error) {
    var result Geometry
    err := catch(func() error {
        wktReader := jts.Io_NewWKTReader()
        jtsA, err := wktReader.Read(a.AsText())
        if err != nil {
            return wrap(err, "converting geometry A to JTS")
        }
        jtsB, err := wktReader.Read(b.AsText())
        if err != nil {
            return wrap(err, "converting geometry B to JTS")
        }
        jtsResult := jtsA.Difference(jtsB)
        result, err = UnmarshalWKT(jtsResult.ToText())
        return wrap(err, "converting JTS difference result to simplefeatures")
    })
    return result, err
}
```

### 5. SymmetricDifference

Using the composition approach for simplicity and correctness:

```go
func SymmetricDifference(a, b Geometry) (Geometry, error) {
    if a.IsEmpty() && b.IsEmpty() {
        return Geometry{}, nil
    }
    if a.IsEmpty() {
        return UnaryUnion(b)
    }
    if b.IsEmpty() {
        return UnaryUnion(a)
    }
    // SymDiff(A, B) = Union(Diff(A, B), Diff(B, A)).
    diffAB, err := Difference(a, b)
    if err != nil {
        return Geometry{}, err
    }
    diffBA, err := Difference(b, a)
    if err != nil {
        return Geometry{}, err
    }
    return Union(diffAB, diffBA)
}
```

## Test Plan

### Test Categories

#### 1. Non-GC Baseline Tests

Verify existing behavior is preserved for non-GC inputs:

- Polygon × Polygon
- MultiPolygon × Polygon
- LineString × Polygon
- Point × Polygon
- Empty geometry handling

#### 2. Simple GC Tests (Homogeneous)

GCs containing elements of the same dimension:

- `GC{Polygon, Polygon}` × Polygon
- Polygon × `GC{Polygon, Polygon}`
- `GC{Polygon, Polygon}` × `GC{Polygon, Polygon}`
- `GC{LineString, LineString}` × Polygon
- `GC{Point, Point}` × Polygon

#### 3. Mixed-Dimension GC Tests

GCs containing elements of different dimensions:

- `GC{Polygon, LineString}` × Polygon
- `GC{Polygon, Point}` × Polygon
- `GC{Polygon, LineString, Point}` × Polygon
- Polygon × `GC{Polygon, LineString}`
- `GC{Polygon, LineString}` × `GC{Polygon, Point}`

#### 4. Nested GC Tests

GCs containing other GCs:

- `GC{GC{Polygon, Polygon}, Polygon}` × Polygon
- `GC{GC{Polygon}, GC{LineString}}` × Polygon
- `GC{GC{GC{Polygon}}}` × Polygon (deeply nested)
- Polygon × `GC{GC{Polygon, LineString}}`

#### 5. Overlapping Component Tests

GCs where components overlap each other:

- `GC{Polygon1, Polygon2}` where polygons overlap × Polygon3
- `GC{Polygon1, Polygon2}` where polygons overlap × `GC{Polygon3, Polygon4}`
  where those also overlap

#### 6. Empty Geometry Tests

- `GC{}` × Polygon
- `GC{POLYGON EMPTY}` × Polygon
- `GC{Polygon, POLYGON EMPTY}` × Polygon
- Polygon × `GC{}`

#### 7. Multi-Geometry in GC Tests

GCs containing Multi* types:

- `GC{MultiPolygon, Polygon}` × Polygon
- `GC{MultiPolygon, MultiLineString}` × Polygon

### Test Matrix

For each test category above, test all four operations:

| Input A | Input B | Union | Intersection | Difference | SymmetricDifference |
|---------|---------|-------|--------------|------------|---------------------|
| ...     | ...     | ✓     | ✓            | ✓          | ✓                   |

### Specific Edge Cases

1. **Result is empty**: Intersection of disjoint geometries.
2. **Result is single element**: Should unwrap from collection if normalization
   is applied.
3. **Result has mixed dimensions**: Intersection of polygon with a GC containing
   a line that crosses the polygon.
4. **A equals B**: All operations should handle this correctly.
5. **A contains B**: Difference should return A with B-shaped hole.
6. **B contains A**: Difference should return empty.

## Implementation Order

1. **UnaryUnion**: Simple change, enables the rest.
2. **Intersection**: Establishes the pattern for GC decomposition.
3. **Difference**: Similar pattern but with asymmetric GC handling.
4. **SymmetricDifference**: Composition of Union and Difference.
5. **Normalization**: Add if testing reveals it's needed.

