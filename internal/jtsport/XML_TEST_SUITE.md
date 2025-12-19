# JTS XML Test Suite

This document describes the JTS XML test suite and how to port it to Go.

## Background

JTS has two separate test systems:

1. **Unit tests** - Standard JUnit tests with inline WKT strings in Java code.
   These have been ported to Go as `*_test.go` files with inline WKT.

2. **XML test suite** - A large collection of XML files containing test cases
   with geometry data and expected results. These are executed by a custom test
   runner (`CoreGeometryXMLTest.java`). **This test suite has not been ported.**

The XML test suite contains approximately 8,500+ test operations across 113
files, providing extensive coverage of spatial predicates, overlay operations,
and edge cases.

## XML Test File Locations

In the JTS repository (`../../locationtech/jts/` relative to this directory):

```
modules/tests/src/test/resources/testxml/
├── general/     # 49 files - Core operation tests
├── validate/    # 9 files  - Relate/spatial predicate validation
├── robust/      # 55 files - Edge cases and robustness tests
├── failure/     # 4 files  - Known failure cases (commented out in Java)
└── misc/        # 10 files - GEOS compatibility tests (commented out in Java)
```

The Java test runner (`CoreGeometryXMLTest.java`) only runs `general/` and
`validate/` by default. The other directories are commented out.

## XML Format

Each XML file contains a `<run>` element with multiple `<case>` elements. Each
case has input geometries and one or more test operations.

### Basic Structure

```xml
<run>
  <desc>Test description</desc>

  <case>
    <desc>Case description</desc>
    <a>POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))</a>
    <b>POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))</b>
    <test>
      <op name="intersection" arg1="A" arg2="B">
        POLYGON ((5 5, 10 5, 10 10, 5 10, 5 5))
      </op>
    </test>
    <test>
      <op name="intersects" arg1="A" arg2="B">true</op>
    </test>
  </case>

</run>
```

### Element Details

- `<a>` and `<b>`: Input geometries in WKT format.
- `<test>`: Contains a single operation to test.
- `<op>`: The operation with attributes:
  - `name`: Operation name (e.g., "intersection", "contains", "relate").
  - `arg1`: Which geometry to operate on ("A" or "B").
  - `arg2`: Second geometry argument for binary operations ("A", "B", or a WKT
    string).
  - `arg3`: Additional argument (e.g., distance for buffer, DE-9IM pattern for
    relate).
  - Element content: Expected result (WKT geometry, boolean, or number).

### Result Matchers

Some XML files specify a custom result matcher:

```xml
<run>
  <resultMatcher>org.locationtech.jtstest.testrunner.BufferResultMatcher</resultMatcher>
  ...
</run>
```

The `BufferResultMatcher` uses approximate geometry comparison rather than exact
equality. For non-buffer tests, exact geometry equality is used.

## Operations Tested

### Supported Operations (~95% of tests)

These operations are implemented in the Go port and can be tested immediately:

| XML Operation      | Count   | Go Method                | Notes                      |
| ---------------    | ------- | -----------              | -------                    |
| `relate`           | 561     | `Relate(g, pattern)`     | Checks DE-9IM pattern      |
| `intersects`       | 563     | `Intersects(g)`          |                            |
| `contains`         | 553     | `Contains(g)`            |                            |
| `covers`           | 533     | `Covers(g)`              |                            |
| `coveredBy`        | 507     | `CoveredBy(g)`           |                            |
| `within`           | 498     | `Within(g)`              |                            |
| `touches`          | 497     | `Touches(g)`             |                            |
| `overlaps`         | 497     | `Overlaps(g)`            |                            |
| `disjoint`         | 497     | `Disjoint(g)`            |                            |
| `crosses`          | 497     | `Crosses(g)`             |                            |
| `equalsTopo`       | 497     | `EqualsTopo(g)`          |                            |
| `isValid`          | 841     | `IsValid()`              |                            |
| `isSimple`         | 44      | `IsSimple()`             |                            |
| `intersection`     | 171     | `Intersection(g)`        | Includes `intersectionNG`  |
| `union`            | 138     | `Union(g)`               | Includes `unionNG`         |
| `difference`       | 162     | `Difference(g)`          | Includes `differenceNG`    |
| `symDifference`    | 120     | `SymDifference(g)`       | Includes `symdifferenceNG` |
| `distance`         | 32      | `Distance(g)`            |                            |
| `convexhull`       | 14      | `ConvexHull()`           |                            |
| `getCentroid`      | 38      | `GetCentroid()`          |                            |
| `getboundary`      | 12      | `GetBoundary()`          |                            |
| `equalsExact`      | 18      | `EqualsExact(g)`         |                            |
| `isWithinDistance` | 22      | `IsWithinDistance(g, d)` |                            |

Operations with `NG` suffix (e.g., `intersectionNG`) use the newer OverlayNG
algorithm. The Go port supports both old and new algorithms.

Operations with `SR` suffix (e.g., `intersectionSR`) use snap-rounding. These
should map to the same Go methods with appropriate precision model.

### Unsupported Operations (~5% of tests)

These operations are not yet implemented:

| XML Operation      | Count   | Blocker                     |
| ---------------    | ------- | ---------                   |
| `buffer`           | 29+     | Phase 16 - Buffer Operation |
| `bufferMitredJoin` | 5       | Phase 16                    |
| `simplifyDP`       | 18      | Simplify not ported         |
| `simplifyTP`       | 18      | Simplify not ported         |
| `densify`          | 14      | Densify not ported          |
| `polygonize`       | 6       | Polygonizer not ported      |
| `minClearance`     | 12      | MinimumClearance not ported |
| `minClearanceLine` | 12      | MinimumClearance not ported |
| `getInteriorPoint` | 24      | InteriorPoint not ported    |

The test runner should skip these operations with a clear log message.

## Implementation Plan

### Directory Structure

```
internal/jtsport/
├── jts/
│   └── ... (existing ported code)
├── xmltest/
│   ├── runner.go      # XML test runner
│   ├── parser.go      # XML parsing
│   ├── operations.go  # Operation dispatcher
│   └── matchers.go    # Result comparison
└── testdata/
    └── testxml/       # Copied from JTS
        ├── general/
        └── validate/
```

### XML Parser

Parse the JTS XML format into Go structs:

```go
type TestRun struct {
    Description   string
    ResultMatcher string // Optional, e.g., "BufferResultMatcher"
    Cases         []TestCase
}

type TestCase struct {
    Description string
    GeometryA   string // WKT
    GeometryB   string // WKT (optional)
    Tests       []Test
}

type Test struct {
    Operation    string            // e.g., "intersection", "contains"
    GeometryArg  string            // "A" or "B"
    Arguments    []string          // Additional args (other geometry, distance, pattern)
    Expected     string            // Expected result as string
}
```

Use Go's `encoding/xml` package. Handle whitespace in WKT content carefully.

### Operation Dispatcher

Map XML operation names to Go method calls:

```go
func dispatch(op string, geomA, geomB *jts.Geom_Geometry, args []string) (any, error) {
    switch strings.ToLower(op) {
    case "intersection", "intersectionng":
        return geomA.Intersection(geomB), nil
    case "union", "unionng":
        return geomA.Union(geomB), nil
    case "contains":
        return geomA.Contains(geomB), nil
    case "relate":
        pattern := args[0]
        return geomA.Relate(geomB, pattern), nil
    // ... etc
    default:
        return nil, fmt.Errorf("unsupported operation: %s", op)
    }
}
```

### Result Matchers

Compare actual results to expected results:

1. **Geometry results**: Parse expected WKT, compare using `EqualsExact` or
   `EqualsTopo`. Handle empty geometries and geometry type differences.

2. **Boolean results**: Parse "true"/"false" strings.

3. **Numeric results**: Parse float64, compare with tolerance.

4. **Buffer results**: Use approximate comparison (area difference within
   tolerance).

### Test Integration

Create a Go test that runs all XML files:

```go
func TestXMLSuite(t *testing.T) {
    files, _ := filepath.Glob("../testdata/testxml/*/*.xml")
    for _, file := range files {
        t.Run(filepath.Base(file), func(t *testing.T) {
            runXMLTestFile(t, file)
        })
    }
}
```

Use `t.Skip()` for unsupported operations rather than failing.

### Handling Edge Cases

1. **Empty geometries**: `POINT EMPTY`, `POLYGON EMPTY`, etc. are valid WKT.

2. **Geometry collections**: Results may be `GEOMETRYCOLLECTION` containing
   mixed types.

3. **Coordinate precision**: Some tests may require specific precision handling.

4. **Operation variants**: Map `*NG` and `*SR` suffixes to appropriate
   implementations.

5. **arg1/arg2 ordering**: Some operations are on geometry A with B as argument,
   others vice versa.

## File Inventory

Key XML files to start with (high value, well-supported operations):

### general/

- `TestOverlayEmpty.xml` - Empty geometry overlay tests (294 tests)
- `TestRelateAA.xml` - Polygon-polygon relate tests
- `TestRelateLA.xml` - Line-polygon relate tests
- `TestRelateLL.xml` - Line-line relate tests
- `TestRelatePL.xml` - Point-line relate tests
- `TestRelatePA.xml` - Point-polygon relate tests
- `TestNGOverlayA.xml` - OverlayNG tests
- `TestValid2.xml` - Validity tests (752 tests)

### validate/

- `TestRelateAA.xml` - 1,177 polygon-polygon relate tests
- `TestRelateLL.xml` - 1,584 line-line relate tests
- `TestRelatePL.xml` - 1,089 point-line relate tests
- `TestRelatePA.xml` - 451 point-polygon relate tests
- `TestRelateLA.xml` - 847 line-polygon relate tests

### robust/ (optional, for later)

Contains edge cases and regression tests from GEOS/PostGIS bugs. These are
valuable but more complex. Add after the basic runner works.

## References

- JTS test runner: `modules/tests/src/test/java/org/locationtech/jtstest/testrunner/`
- JTS test entry point: `modules/tests/src/test/java/org/locationtech/jtstest/CoreGeometryXMLTest.java`
- XML test files: `modules/tests/src/test/resources/testxml/`
