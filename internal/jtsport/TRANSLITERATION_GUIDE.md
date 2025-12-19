# Java to Go Transliteration Guide

## Overview

Strict requirements:

1. **1-1 Mapping**: Each Java class maps to exactly one Go struct. Java
   interfaces map to Go interfaces. **Each Java file maps to exactly one Go
   file, and each Go file maps to exactly one Java file.** Never combine
   multiple Java files into a single Go file, even if they seem related (e.g.,
   multiple test classes). Each Java method maps to a Go method (for Java
   instance methods) or a Go function (for Java static methods).

2. **Preserve Element Order**: The order of elements (fields, constants,
   methods, inner classes) in the Go file must match the order in the Java
   source file. This enables side-by-side manual review. Only deviate from Java
   ordering when Go language constraints make it physically impossible (e.g.,
   forward references that Go cannot resolve). See "Element Ordering" section
   below for details.

3. **Behavioral Equivalence**: The Go code must behave identically to the Java
   code. This includes tests. Never create tests when there are no corresponding
   Java tests to port.

4. **No Shortcuts**: Do not replace Java helper methods with Go standard library
   functions, even when equivalent functionality exists. Transliterate the Java
   method directly to maintain structural correspondence and verifiability. For
   example, if Java has a private `stringOfChar(char, int)` method, implement it
   in Go rather than using `strings.Repeat()`.

## Package Organization

All code goes into a single Go package (`package jts`) in one flat directory.
Namespacing is achieved through systematic prefixing.

## Element Ordering

**The order of elements in Go files must match the Java source.** This is
critical for enabling side-by-side manual review and ensuring nothing is
accidentally omitted or duplicated.

### What Must Match

- Static fields/constants appear in the same order
- Instance fields appear in the same order (within the struct definition)
- Methods appear in the same order
- Inner classes/types appear in the same order
- Section comments (like `// Output` or `// Predicates`) should be preserved

### Example

If Java has:
```java
public class Foo {
    public static final double PI = 3.14;
    private static Bar createBar() { ... }
    public static Foo valueOf(String s) { ... }
    private static final double EPSILON = 0.001;
    private double value;
    public Foo() { ... }
    public double getValue() { ... }
    public void setValue(double v) { ... }
}
```

Go must follow the same order:
```go
var Foo_PI = 3.14
func foo_createBar() *Bar { ... }
func Foo_ValueOfString(s string) *Foo { ... }
const foo_epsilon = 0.001
type Foo struct { value float64 }
func NewFoo() *Foo { ... }
func (f *Foo) GetValue() float64 { ... }
func (f *Foo) SetValue(v float64) { ... }
```

### Acceptable Exceptions

Only deviate from Java ordering when Go makes it physically impossible:

1. **Forward references in var/const initialization**: If a Go var/const
   initialization references another var/const that hasn't been declared yet,
   reorder only the minimum necessary declarations to resolve the reference.

When reordering is necessary, add a transliteration note:

```go
// TRANSLITERATION NOTE: Moved above foo_epsilon due to Go initialization order
// requirements - foo_epsilon references this value.
var foo_baseValue = 1.0
```

### Naming Conventions

| Java                                           | Go                                             |
| ------                                         | -----                                          |
| File: `geom/Polygon.java`                      | `geom_polygon.go`                              |
| File: `geom/impl/CoordinateArraySequence.java` | `geom_impl_coordinate_array_sequence.go`       |
| Type: `geom.Polygon`                           | `type Geom_Polygon struct`                     |
| Type: `geom.impl.CoordinateArraySequence`      | `type GeomImpl_CoordinateArraySequence struct` |
| Constructor: `new Polygon(...)`                | `Geom_NewPolygon(...) *Geom_Polygon`           |
| Instance method: `polygon.getArea()`           | `func (p *Geom_Polygon) GetArea() float64`     |
| Static method: `algorithm.Area.ofRing(...)`    | `Algorithm_Area_OfRing(...)`                   |
| Static field: `geom.Geometry.TYPENAME_POINT`   | `const Geom_Geometry_TYPENAME_POINT`           |
| Public symbol                                  | `Geom_Polygon`                                 |
| Private symbol                                 | `geom_internalHelper` (lowercase first letter) |

### Complete Example

**Java: geom/Polygon.java**
```java
package org.locationtech.jts.geom;

public class Polygon extends Geometry {
    private LinearRing shell;

    public Polygon(LinearRing shell, GeometryFactory factory) {
        this.shell = shell;
    }

    public double getArea() {
        return algorithm.Area.ofRing(shell);
    }

    private boolean isValidRing() {
        return shell.getNumPoints() >= 4;
    }
}
```

**Go: geom_polygon.go**
```go
package jts

type Geom_Polygon struct {
    *Geom_Geometry
    child any
    shell *Geom_LinearRing
}

func (p *Geom_Polygon) GetChild() any { return p.child }

func Geom_NewPolygon(shell *Geom_LinearRing, factory *Geom_GeometryFactory) *Geom_Polygon {
    geom := &Geom_Geometry{}
    poly := &Geom_Polygon{Geom_Geometry: geom, shell: shell}
    geom.child = poly
    return poly
}

func (p *Geom_Polygon) GetArea() float64 {
    return Algorithm_Area_OfRing(p.shell)
}

func (p *Geom_Polygon) isValidRing() bool {
    return p.shell.GetNumPoints() >= 4
}
```

## Pointers vs Values

**Always use pointers for structs that map to Java classes.** This maintains
Java's reference semantics (mutability, identity, nil/null).

- Constructors return `*ClassName`
- All method receivers use `*ClassName`
- Struct fields storing references use `*ClassName`

## Constructors

Go doesn't support overloading, so multiple constructors become multiple
functions with descriptive names. Use constructor chaining where Java does.

```go
// Default constructor
func NewAccount() *Account {
    return NewAccountWithBalance(0.0)
}

// Constructor with balance (chains to full constructor)
func NewAccountWithBalance(balance float64) *Account {
    return NewAccountWithBalanceAndCurrency(balance, "USD")
}

// Full constructor
func NewAccountWithBalanceAndCurrency(balance float64, currency string) *Account {
    return &Account{balance: balance, currency: currency}
}
```

Naming patterns: `NewClassName()`, `NewClassNameWithX(x)`, `NewClassNameFromY(y)`.

## Method Overloading

Go doesn't support overloading. Use distinct names based on what differs:

| Overload Type   | Java                                   | Go                         |
| --------------- | ------                                 | -----                      |
| By type         | `clamp(double)`, `clamp(int)`          | `ClampFloat64`, `ClampInt` |
| By arity        | `max(a,b,c)`, `max(a,b,c,d)`           | `Max3`, `Max4`             |
| By semantics    | `intersects(p1,p2,q)` (envelope+point) | `IntersectsPointEnvelope`  |

Apply naming symmetrically to all overloads.

## Static Members

Static fields become package-level `const` or `var`. Static methods become
package-level functions. Use `JavaPackage_ClassName_MemberName` naming.

```go
package jts

const Math_MathUtils_PI = 3.14159
var util_Counter_globalCount = 0

func Math_MathUtils_CircleArea(radius float64) float64 {
    return Math_MathUtils_PI * radius * radius
}
```

For complex static field initialization, use IIFEs (not `init()` or helper
functions):

```go
var Util_Counter_default = func() *Util_Counter {
    return Util_NewCounter()
}()
```

## Polymorphism

Java's class-based polymorphism is implemented using a **child-chain dispatch
pattern**. Each level stores a `child` pointer to its immediate child, forming
a linked list from base to leaf.

### Rules Summary

**Structure:**

1. Every type has a `child java.Polymorphic` field (nil for leaf types).
2. Every type implements `GetChild() java.Polymorphic { return x.child }`.
3. Child types embed **pointers** to parent types (`*Parent`, not `Parent`).
4. Use `java.GetSelf(obj)` to get the leaf type.
5. Use `java.InstanceOf[T](obj)` for runtime type checks.

**Methods:**

6. **Dispatchers** (`MethodName()`) are defined once at the level where the
   method is introduced. Never redefined at lower levels.
7. **Implementations** (`MethodName_BODY()`) provide actual behavior. Define at
   any level that provides or overrides behavior.
8. To **override**: define only `MethodName_BODY()` (don't redefine dispatcher).
9. To **introduce** a new method: define both dispatcher and `_BODY()`.
10. **Abstract methods**: dispatcher panics, no `_BODY()` exists.
11. **super calls**: `c.Parent.MethodName_BODY(args)`.

### Complete Example

**Java:**
```java
public abstract class PaymentMethod {
    public abstract boolean process(double amount);
    public double getTransactionFee(double amount) { return 0.0; }
}

public abstract class CardPayment extends PaymentMethod {
    protected String cardNumber;
    @Override public boolean process(double amount) {
        return verifySecurityCode();
    }
    public boolean verifySecurityCode() { return cardNumber.length() == 16; }
}

public class CreditCard extends CardPayment {
    private double creditLimit, currentBalance;
    @Override public boolean process(double amount) {
        if (!super.process(amount)) return false;
        if (currentBalance + amount > creditLimit) return false;
        currentBalance += amount;
        return true;
    }
    @Override public double getTransactionFee(double amount) {
        return amount * 0.025;
    }
}
```

**Go:**
```go
import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// BASE CLASS
type PaymentMethod struct {
    child any
}

func (p *PaymentMethod) GetChild() any { return p.child }

// Abstract method dispatcher (panics if not overridden)
func (p *PaymentMethod) Process(amount float64) bool {
    if impl, ok := java.GetSelf(p).(interface{ Process_BODY(float64) bool }); ok {
        return impl.Process_BODY(amount)
    }
    panic("abstract method called")
}

// Concrete method with default implementation
func (p *PaymentMethod) GetTransactionFee(amount float64) float64 {
    if impl, ok := java.GetSelf(p).(interface{ GetTransactionFee_BODY(float64) float64 }); ok {
        return impl.GetTransactionFee_BODY(amount)
    }
    return p.GetTransactionFee_BODY(amount)
}

func (p *PaymentMethod) GetTransactionFee_BODY(amount float64) float64 {
    return 0.0
}

// INTERMEDIATE CLASS
type CardPayment struct {
    *PaymentMethod
    child      any
    CardNumber string
}

func (c *CardPayment) GetChild() any { return c.child }

// Override: define only _BODY, dispatcher inherited from parent
func (c *CardPayment) Process_BODY(amount float64) bool {
    return c.VerifySecurityCode()
}

// New method at this level: needs dispatcher + implementation
func (c *CardPayment) VerifySecurityCode() bool {
    if impl, ok := java.GetSelf(c).(interface{ VerifySecurityCode_BODY() bool }); ok {
        return impl.VerifySecurityCode_BODY()
    }
    return c.VerifySecurityCode_BODY()
}

func (c *CardPayment) VerifySecurityCode_BODY() bool {
    return len(c.CardNumber) == 16
}

// LEAF CLASS
type CreditCard struct {
    *CardPayment
    child          any
    CreditLimit    float64
    CurrentBalance float64
}

func (c *CreditCard) GetChild() any { return c.child }

// Constructor wires up the child chain
func NewCreditCard(cardNumber string, creditLimit, currentBalance float64) *CreditCard {
    pm := &PaymentMethod{}
    cp := &CardPayment{PaymentMethod: pm, CardNumber: cardNumber}
    cc := &CreditCard{CardPayment: cp, CreditLimit: creditLimit, CurrentBalance: currentBalance}
    pm.child = cp
    cp.child = cc
    return cc
}

// Override with super call
func (c *CreditCard) Process_BODY(amount float64) bool {
    if !c.CardPayment.Process_BODY(amount) { // super.process()
        return false
    }
    if c.CurrentBalance+amount > c.CreditLimit {
        return false
    }
    c.CurrentBalance += amount
    return true
}

func (c *CreditCard) GetTransactionFee_BODY(amount float64) float64 {
    return amount * 0.025
}
```

### Runtime Type Checking

**IMPORTANT:** When translating Java `instanceof` checks, ALWAYS use
`java.InstanceOf[T](obj)` rather than Go's direct type assertion
`obj.GetSelf().(T)`. Direct type assertions only match the exact leaf type,
while `InstanceOf` correctly handles inheritance hierarchies.

```go
// CORRECT: Java instanceof → java.InstanceOf
// Java: if (obj instanceof GeometryCollection)
if java.InstanceOf[*Geom_GeometryCollection](obj) { ... }

// WRONG: Direct type assertion fails for subtypes
// This only matches *Geom_GeometryCollection exactly, NOT MultiPolygon
if _, ok := obj.GetSelf().(*Geom_GeometryCollection); ok { ... } // DON'T DO THIS

// InstanceOf returns true for parent types too:
java.InstanceOf[*PaymentMethod](creditCard) // true
java.InstanceOf[*CardPayment](creditCard)   // true
java.InstanceOf[*CreditCard](creditCard)    // true
```

For example, in Java `MultiPolygon extends GeometryCollection`, so
`geom instanceof GeometryCollection` returns true for MultiPolygon. The Go
equivalent must use `java.InstanceOf[*Geom_GeometryCollection](geom)`
to get the same behavior.

### Java Casts

**IMPORTANT:** When translating Java casts, ALWAYS use `java.Cast[T](obj)`.
Never use `java.GetLeaf(obj).(*T)` or `obj.GetSelf().(*T)` as a cast
replacement—these only match the exact leaf type and fail for subtypes.

```go
// Java:
addLineString((LineString) g);

// CORRECT: Use java.Cast
addLineString(java.Cast[*Geom_LineString](g))

// WRONG: GetLeaf + type assertion fails if g is a LinearRing
addLineString(java.GetLeaf(g).(*Geom_LineString)) // DON'T DO THIS
```

`java.Cast[T]` walks the type hierarchy to find `T`, matching Java's cast
semantics where `(LineString) linearRing` succeeds because `LinearRing extends
LineString`. It panics if the cast fails (like Java's ClassCastException).

When combining `instanceof` with a cast:

```go
// Java:
// if (geom instanceof Polygon) {
//     Polygon poly = (Polygon) geom;
//     // use poly
// }

// Go:
if java.InstanceOf[*Geom_Polygon](geom) {
    poly := java.Cast[*Geom_Polygon](geom)
    // use poly
}
```

## Floating-Point Precision (strictfp)

Go may use FMA optimizations that break algorithms depending on specific
rounding. For `strictfp` Java code, wrap every floating-point operation in
`float64()` to force rounding:

```go
// Java: c = ((a*b - C) + d*e) + f*g
c = float64(float64(float64(a*b)-C)+float64(d*e)) + float64(f*g)
```

## Math.round()

Use `java.Round()` instead of `math.Round()` when porting
`Math.round()` calls. They differ on negative half-way values.

## Math.abs() for integers

Use `java.AbsInt()` when porting `Math.abs()` calls on integer values.
Go's `math.Abs()` only works on `float64`, so this provides the integer version.

## Map Iteration Order

**Go's map iteration order is randomized on each iteration**, while Java's
`HashMap` iteration order is consistent (though unspecified) within a single JVM
run, and Java's `TreeMap` provides sorted iteration by key.

When transliterating Java code that iterates over a map and the iteration order
affects output (e.g., building a result list), use `java.SortedKeysString()` (for
string keys) or `java.SortedKeysInt()` (for int keys) to ensure deterministic
behavior:

```java
// Java - HashMap iteration (consistent within JVM run)
for (Entry<String, Point> entry : map.entrySet()) {
    resultList.add(entry.getValue());
}
```

```go
// Go - sort keys for consistent iteration
for _, key := range java.SortedKeysString(m) {
    resultList = append(resultList, m[key])
}
```

This applies to both `HashMap` and `TreeMap` translations. The sorting ensures
the Go code produces consistent output across runs, matching the behavioral
consistency of Java.

**When to use `java.SortedKeysString()` / `java.SortedKeysInt()`:**
- When iteration order affects output (building result collections)
- When iteration order affects algorithm correctness
- When translating Java `TreeMap` (which explicitly guarantees sorted order)

**When NOT needed:**
- Read-only lookups (`map[key]`)
- Iteration where order doesn't matter (e.g., summing all values)
- Maps only used for membership checks

## JUnit Assertions

The `internal/jtsport/junit` package provides JUnit-style assertion helpers for
ported tests. These enable 1-1 line mapping between Java JUnit tests and Go
tests.

| JUnit                              | Go                                      |
| ---------------------------------- | --------------------------------------- |
| `assertEquals(expected, actual)`   | `junit.AssertEquals(t, expected, actual)` |
| `assertTrue(condition)`            | `junit.AssertTrue(t, condition)`        |
| `assertFalse(condition)`           | `junit.AssertFalse(t, condition)`       |
| `assertNull(value)`                | `junit.AssertNull(t, value)`            |
| `assertNotNull(value)`             | `junit.AssertNotNull(t, value)`         |
| `fail(message)`                    | `junit.Fail(t, message)`                |

**Example:**

```java
// Java
assertEquals(3, iar.size());
assertTrue(result.isValid());
```

```go
// Go
junit.AssertEquals(t, 3, iar.Size())
junit.AssertTrue(t, result.IsValid())
```

## Marker Interfaces

Java marker interfaces (empty interfaces for categorization) become Go
interfaces with an exported marker method:

```go
type Puntal interface {
    IsPuntal()
}

func (p *Point) IsPuntal()      {}
func (m *MultiPoint) IsPuntal() {}
```

## Java Interfaces

Java interfaces with methods map to native Go interfaces. Each interface has an
exported marker method for type identification.

### Structure

For a Java interface `Foo`:

```go
type Foo interface {
    IsFoo()           // Marker method
    GetData() any     // Interface methods
    Size() int
}
```

### First-Level Implementations

Types that directly implement the Java interface provide:

1. The marker method
2. A compile-time implementation check
3. All interface methods (including default implementations)

```go
var _ Foo = (*FooImpl)(nil)  // Compile-time check

type FooImpl struct {
    // fields
}

func (f *FooImpl) IsFoo() {}  // Marker method

func (f *FooImpl) GetData() any { return f.data }
func (f *FooImpl) Size() int    { return len(f.items) }
```

### Default Method Implementations

When Java interfaces have default method implementations, copy the default body
into each Go implementation that doesn't override it.

**Java:**
```java
public interface SegmentString {
    int size();
    Coordinate getCoordinate(int i);

    // Default implementation
    default Coordinate prevInRing(int index) {
        int prevIndex = index - 1;
        if (prevIndex < 0) prevIndex = size() - 2;
        return getCoordinate(prevIndex);
    }
}
```

**Go:**
```go
type Noding_SegmentString interface {
    IsNoding_SegmentString()
    Size() int
    GetCoordinate(i int) *Geom_Coordinate
    PrevInRing(index int) *Geom_Coordinate
}

// Each implementation includes the default behavior:
func (ss *Noding_BasicSegmentString) PrevInRing(index int) *Geom_Coordinate {
    prevIndex := index - 1
    if prevIndex < 0 {
        prevIndex = ss.Size() - 2
    }
    return ss.GetCoordinate(prevIndex)
}
```

### Extending Classes That Implement Interfaces

When a class extends another class that implements an interface, the child
class inherits the interface implementation through struct embedding. The child
class does NOT need its own marker method or polymorphic infrastructure for the
interface.

```go
// Parent implements the interface
var _ Noding_SegmentString = (*Noding_BasicSegmentString)(nil)

type Noding_BasicSegmentString struct {
    pts  []*Geom_Coordinate
    data any
}

func (ss *Noding_BasicSegmentString) IsNoding_SegmentString() {}

// Child extends parent and inherits interface implementation
type RelateSegmentString struct {
    *Noding_BasicSegmentString  // Embedding provides interface methods
    isA bool
    // additional fields
}
// No marker method needed - inherited from BasicSegmentString
```

## Dead Code

Include all methods from the Java source, even if they are unused (dead code).
This maintains strict 1-1 correspondence and makes side-by-side verification
easier. Do not omit private helper methods just because they are not called, and
do not omit debug/print methods just because they only produce output.

## Copyright Headers

Do not copy copyright headers from Java files. Go files start directly with
`package`.

## Transliteration Notes

Whenever the Go code differs from the Java source in a way that breaks strict
1-1 structural correspondence, add a comment explaining the difference:

```go
// TRANSLITERATION NOTE: <explanation of the difference and why it's necessary>
```

Examples of when to add a transliteration note:

- Extra methods required by Go (e.g., `Error()` for the error interface)
- Inlined logic that replaces Java's polymorphic dispatch
- Different control flow due to Go's lack of exceptions
- Any structural deviation from the Java source

The comment should be placed immediately before the divergent code.

## Handling Stubs

When porting a file that depends on unported classes, create stubs in a
dedicated `stubs.go` file:

```go
package jts

// =============================================================================
// STUBS: Stub types for classes not yet ported. Will be replaced when ported.
// =============================================================================

// STUB: Geom_GeometryFactory - will be ported in Phase 2f.
type Geom_GeometryFactory struct{}

// STUB: Method stubbed - Geom_GeometryFactory not yet ported.
func (gf *Geom_GeometryFactory) GetSRID() int {
    panic("Geom_GeometryFactory not yet ported")
}
```

When porting a stubbed class: delete from `stubs.go`, implement in its own file.

## File Manifest

The `MANIFEST.csv` file tracks all ported and pending files. It provides a
file-level inventory for validating completeness and tracking manual reviews.

### Format

CSV with three columns:

```
java_file,go_file,status
algorithm/Area.java,algorithm_area.go,ported
algorithm/AreaTest.java,algorithm_area_test.go,ported
geom/Polygon.java,geom_polygon.go,reviewed
operation/buffer/BufferOp.java,operation_buffer_buffer_op.go,pending
```

### Columns

| Column      | Description                                                    |
| ----------- | -----------------------------------------------------          |
| `java_file` | Relative path from JTS source root (e.g., `geom/Polygon.java`) |
| `go_file`   | Go filename in the `jts` package (e.g., `geom_polygon.go`)     |
| `status`    | One of: `pending`, `ported`, `reviewed`                        |

### Status Values

- **pending**: Not yet ported.
- **ported**: Ported but not manually reviewed.
- **reviewed**: Ported and manually verified for correctness.

### Rules

- One entry per file (test files are separate entries).
- Ordered alphabetically by `java_file`.
- Only tracks ported or pending files (no "not needed" entries).
