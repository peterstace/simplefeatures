# JTS Port Design

## Purpose

Package `internal/jts` is a Go translation of JTS core 1.20.0 (the
`modules/core` Maven module, Java package `org.locationtech.jts`). It excludes
`org.locationtech.jts.awt` (a Java2D bridge) and the XML-based readers in
`io.gml2` and `io.kml`. The `geom` package exposes JTS operations to
simplefeatures users on top of it.

Goals, in priority order:

1. **Correctness.** For every input, the Go code produces the same result as
   JTS 1.20.0 running on the JVM, bit for bit.
2. **Speed.**

Readable or idiomatic Go is not a goal for the translated code. Neither is
following JTS master between releases: a new JTS release is ported by running
the translator against it.

## Components

| Component  | Location                         | Language | Written by          |
| ---------- | -------------------------------- | -------- | ------------------- |
| Translator | `tools/translator`               | Java     | hand                |
| Port       | `*.go` in this directory         | Go       | translator only     |
| Runtime    | `java/`                          | Go       | hand                |
| JUnit shim | `junit/`                         | Go       | hand                |
| Oracle     | `tools/oracle` (JVM), `oracle/`  | Java, Go | hand                |

The translator parses and type-checks the JTS sources with javac's compiler
API and writes the Go package. Generated files are never edited by hand: a
wrong result is fixed in the translator or the runtime, and the package is
regenerated with `tools/translate.sh`.

The runtime implements the JDK classes and language semantics that the
generated code uses. It is small compared to the port and is where human review
effort goes.

The oracle runs the same operations through JTS on the JVM and through the
port, and requires identical output.

`tools/setup.sh` puts JTS 1.20.0 sources, compiled JTS classes, and the JUnit
jars needed to type-check the JTS tests in `tools/.cache` (not checked in).

All Go code builds with Go 1.18, the version `go.mod` declares.

## Translation Model

All generated code is in one Go package, because JTS packages import each other
cyclically. Each Java source file produces one Go file, named after the Java
package and file (`geom/LineString.java` becomes `Geom_LineString.go`). Test
classes produce `_test.go` files, except test classes that extend a main class:
those go in non-test files, because the main class's dispatch code refers to
them.

### Names

| Java                                        | Go                                                |
| ------------------------------------------- | ------------------------------------------------- |
| class `geom.impl.CoordinateArraySequence`   | type `GeomImpl_CoordinateArraySequence`           |
| nested class `geom.Geometry.Foo`            | type `Geom_Geometry_Foo`                          |
| static method `algorithm.Orientation.index` | func `Algorithm_Orientation_index`                |
| static field `geom.Geometry.TYPENAME_POINT` | var or const `Geom_Geometry_TYPENAME_POINT`       |
| constructor `new LineString(seq, factory)`  | func `Geom_LineString_new`                        |
| instance method `getArea()`                 | method `GetArea`                                  |
| instance method `equals(Geometry)`          | method `Equals_Geometry`                          |
| private instance method `init(...)`         | method `init_...` (lowercase first letter)        |
| instance field `points`                     | field `points`                                    |

The type prefix comes from the Java package below `org.locationtech.jts`. The
JTS test-support package `test.jts` gets the prefix `Testjts`, because
`go test` treats functions named `TestXxx` as tests.

Static members keep Java's lowercase first letter after the type prefix, which
keeps them apart from nested types. The identifier is still exported because it
starts with the type prefix.

Instance method names carry the erased parameter types of the method they
ultimately override (the root declaration), so overloads never collide and an
override always has the same Go name as the method it overrides. A zero-argument
method has no suffix. Static methods and constructors get the suffix only when
their class overloads the name.

JDK methods follow the same rules (`list.Get_int(i)`, `map.Put_Object_Object(k,
v)`), with overload sets taken from the Java 8 API that JTS compiles against.
The runtime's names are therefore fixed by the JDK's signatures.

Locals and parameters that collide with Go keywords or predeclared identifiers
get a trailing underscore.

### Types

| Java                                  | Go                                     |
| ------------------------------------- | -------------------------------------- |
| `boolean`                             | `bool`                                 |
| `byte`, `short`, `char`               | `int8`, `int16`, `uint16`              |
| `int`, `long`                         | `int32`, `int64`                       |
| `float`, `double`                     | `float32`, `float64`                   |
| class `C`                             | `*C`                                   |
| JTS interface `I`                     | Go interface `I`                       |
| `Object`, type variables, boxed types | `any`                                  |
| `String`                              | `*java.String`                         |
| `T[]`                                 | `[]T`                                  |

Generics are erased, as javac erases them. Where javac would insert a cast after
reading through an erased type, the translator inserts the equivalent Go
conversion.

A value of class type stored in an `any` or a Go interface is always the pointer
to the object's concrete class. Identity comparison and type tests on `any` rely
on this.

A JDK method parameter or result of type `Object[]` is an `any` holding a slice
of any element type. A `Coordinate[]` passed to `Arrays.sort(Object[])` is
therefore the caller's slice, sorted in place, rather than a converted copy.

A class whose type is not available in the code being generated (an excluded
class, a test class referred to from main code, or a JDK class the runtime does
not have) is represented as `any`; any use of its members becomes a stub.

### Inheritance and Dispatch

A class embeds its superclass struct by value as its first field. Every part of
an object therefore starts at the same address, so an upcast or downcast is a
pointer reinterpretation through `unsafe.Pointer` that preserves nil. A
generated test checks that the embedded superclass is at offset zero for every
class.

The root of each class hierarchy has a `jkind_` field holding the object's
concrete class ID. IDs are numbered in preorder, so the classes in any subtree
have contiguous IDs and `instanceof` against a class is a range check.

A method that a subclass overrides gets a dispatcher at the class that
introduces it. The dispatcher switches on `jkind_` and calls the `_impl` method
of the class whose implementation applies. A call site calls the `_impl` method
directly when no subclass of the receiver's static type overrides the method.

Every Go method has the signature of the erased root declaration. Where Java's
types differ, for example a covariant return type or a generic interface
implemented with a concrete parameter type, the translator converts at method
entry, at return, and at the call site.

Each Java interface maps to a Go interface that includes a marker method, so
Go's structural typing cannot make `instanceof` true for a class that does not
declare the interface. A default method body becomes a function, and each class
that inherits the default gets a method that calls it.

A method of a JDK interface is part of a class's dispatch only if a JTS class
implements it, or if it is a default method that the runtime's Go interface
declares (such as `Iterator.remove`), in which case the class gets a method that
calls the runtime function `java.<Interface>_default_<method>`. Otherwise a
runtime superclass supplies the method. A call to any other JDK default method
(such as `Collection.stream`) calls that runtime function directly.

Each concrete class has a `JShallowCopy` method, which implements
`Object.clone()`, and each class that directly implements `Comparable` of itself
has a `JComparableSelf` method, which `HashMap` needs to order keys in
treeified bins.

Anonymous, local, and inner classes become ordinary structs with fields for the
enclosing instance and the captured locals.

### Classes That Extend the Runtime

Some JTS classes extend JDK classes: exceptions extend `RuntimeException`,
`CoordinateList` extends `ArrayList`, enums extend `Enum`, and tests extend
JUnit's `TestCase`. The runtime base class is embedded like any other
superclass. Code that holds such an object through the runtime type (a `catch
(RuntimeException e)` that rethrows `e`, say) only has a pointer to the embedded
part. To get from there back to the whole object, these runtime classes record
it: the generated constructor calls `JSetSelf`, and `java.Ref` returns the
recorded object when it converts the part to an `any`.

Public instance fields of runtime classes are exported Go fields named `F_` plus
the Java name, such as `StreamTokenizer.F_sval`.

An enum constant is created by the enum's constructor and then given its name
and ordinal with `java.EnumConst`. The translator generates `values()` and
`valueOf()`, and a `switch` on an enum switches on the ordinal.

### Expressions

- Every floating-point multiplication is wrapped in an explicit conversion
  (`float64(a*b)`). The Go spec allows fusing `a*b+c` into one fused
  multiply-add, which rounds once where Java rounds twice; an explicit
  conversion forbids fusing.
- Constant expressions are emitted as the value javac folded, so Go's exact
  constant arithmetic never applies.
- `int` and `long` arithmetic wraps. Shift counts are masked as Java masks them,
  and `>>>` shifts the unsigned equivalent.
- Floating-point to integer casts call runtime functions that saturate and map
  NaN to zero, as Java does.
- Compound assignments include Java's implicit narrowing conversion.
- String concatenation uses Java's string conversions, including
  `Double.toString`.
- Increments, decrements, and assignments used as expressions call runtime
  helpers that take a pointer to the target.
- `?:` becomes an immediately called function literal, so only one branch is
  evaluated.
- A class literal (`Foo.class`) becomes a `*java.Class` identified by the
  class's binary name.

### Statements

- `throw` panics with the exception object. `try` blocks become runtime calls
  that take closures; `return`, `break`, and `continue` inside them are carried
  out of the closure as a status value. A Go runtime panic (nil dereference,
  index out of range, failed type assertion, integer division by zero) is
  converted to the corresponding Java exception before `catch` clauses are
  matched.
- A `switch` case that does not end in a jump falls through, using
  `fallthrough`.
- `assert` statements are dropped, because Java runs with assertions disabled by
  default.

### Concurrency

JTS makes some objects safe to share between threads, such as prepared
geometries and spatial indexes, which build their indexes lazily.

- A class hierarchy whose objects are locked (by a `synchronized` instance
  method or a `synchronized` block on such an object) has a `jmon_` mutex in its
  root struct. A `synchronized` method locks it for the whole method, and a
  `synchronized` block becomes a lock plus a `try`/`finally` that unlocks. Other
  objects used as locks get a monitor from the runtime: `ArrayList` has one, and
  any other object gets one from a table keyed by identity.
- Java monitors are reentrant and Go mutexes are not. No JTS method locks a
  monitor it already holds.
- A `volatile` field is a `java.VolatileBool` or `java.VolatilePtr`, which read
  and write atomically. The volatile fields guard lazily built indexes.

### Static Fields

Static fields become package-level variables, and compile-time constants become
typed constants. Go initializes package-level variables at program start in
dependency order, whereas Java initializes each class on first use. The two
orders give the same result only when static initializers have no side effects
on other classes.

Every static initializer that calls a method or constructor runs inside
`java.StaticInit`. If it panics, the field keeps its zero value and the panic
is recorded; a generated
test fails while any are recorded. This keeps one failing initializer from
stopping the whole program, which Java's lazy initialization would also avoid.

### Unsupported Constructs

When the translator cannot translate a method body, it emits a body that panics
with the reason and records the method in `tools/translator/REPORT.md`. The same
happens when a body refers to a runtime member that does not exist yet; the
translator learns what exists by scanning the runtime's Go source. The package
always compiles, and the report is the list of remaining work.

`tools/translator/runtime-api.txt` lists every runtime member the JTS sources
need, including those behind other unsupported constructs, with whether the
runtime has it yet.

## Runtime Semantics

Each item below can change results, so the runtime reproduces JDK 21 exactly.

- **`Math`**: `StrictMath` (fdlibm) semantics. HotSpot's `Math.sin` and
  `Math.cos` intrinsics on aarch64 differ from fdlibm in about 4% of inputs, so
  the oracle runs the JVM with `-XX:-InlineMathNatives`, which makes `Math`
  match `StrictMath`. Go's `math` package uses different algorithms and is not
  used for these functions.
- **Number formatting and parsing**: `Double.toString`, `Double.parseDouble`,
  `Double.compare`, `Double.equals`, `Double.hashCode`, `String.format`, and the
  `DecimalFormat` patterns that JTS uses.
- **Hashed collections**: `HashMap`, `HashSet`, and `LinkedHashMap` reproduce
  Java's table sizing, hash spreading, resize order, and treeified bins, so
  iteration order matches.
- **Other collections**: `TreeMap` and `TreeSet`; `PriorityQueue` with Java's
  sift algorithm, which decides the order of equal elements; `ArrayList`,
  `ArrayDeque`, `LinkedList`, `Stack`, and `EnumSet`.
- **Sorting**: `Collections.sort` and `Arrays.sort` on objects use a port of
  Java's TimSort. Any stable sort would give the same order for a consistent
  comparator, but some JTS comparators are inconsistent (`ConvexHull`'s radial
  comparator, for example), and then only TimSort's exact comparison sequence
  reproduces Java's result. `Arrays.sort(double[])` uses Java's total order,
  with `-0.0` before `0.0` and NaN last.
- **Identity hash codes**: a counter. The oracle runs the JVM with
  `-XX:hashCode=3`, which makes HotSpot assign identity hash codes from a global
  counter, and syncs the Go counter with it before each operation.
- **`StreamTokenizer`** and `java.util.Random`: ported algorithms.
- **Exceptions**: the class hierarchy, and messages where JTS inspects them.

Each runtime area has a differential test: a Java program under
`tools/jdkvectors` records JDK results in `java/testdata`, and a Go test replays
them.

## Oracle

The oracle compiles JTS 1.20.0 from the release tag with `javac`, since JTS core
has no dependencies. A long-running driver process (`tools/oracle/Driver.java`)
runs with:

```
-Duser.language=en -Duser.country=US
-XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives
-XX:+UnlockExperimentalVMOptions -XX:hashCode=3
```

At startup the driver loads and initializes every JTS class and runs each
operation once. Class loading and static initialization consume identity hash
codes, and the warm-up keeps them from shifting the counter during measured
operations.

The Go harness (`oracle/`) loads each input on both sides from the same WKB with
the same precision model, then sends each operation to the driver and to the
port. Results must be identical: geometries are compared by class, structure,
coordinate sequence class, and the bits of every ordinate; numbers by their
bits; and exceptions by class. Any difference that cannot be removed is listed,
with its reason, under Known Deviations.

Cases come from:

- the XML test files in JTS's `modules/tests` (`TestXML`)
- the WKT literals in simplefeatures' `geom` tests, and deterministic random
  geometries: star-shaped polygons, lines, and multipoints, both on an integer
  grid (coincident vertices, collinear edges) and with full-precision
  coordinates (near-degenerate intersections) (`TestCorpus`)

Each geometry gets the unary operations, and pairs of geometries get the binary
operations: relate and its predicates, prepared predicates, the four overlay
operations, distance, and Hausdorff distance.

The translated JTS JUnit tests (`go test` in this directory) complement the
oracle. They cover the tests that JTS's own build runs: files named
`*Test.java`, apart from `*PerfTest.java`, `*StressTest.java`, and the
`test.jts.perf` package.

The oracle tests always run: they fetch and build JTS with `tools/setup.sh` if
needed, and fail if there is no JDK. CI runs them in an image with Temurin JDK
21.0.9 (`.ci/unit.Dockerfile`).

## Workflow

Work is split by Java construct or runtime class, not by file. A mismatch found
by the oracle or by a translated JUnit test is fixed at its root cause in the
translator or runtime, which fixes every file with the same construct.

Review targets the translator's rules, the runtime, and Known Deviations. It
does not cover individual generated files.

## Known Deviations

The oracle has found no difference in results. The following differences from
Java are known; none is reachable from the operations the oracle exercises.

Language semantics:

- Calling a method on a null receiver panics only when the method reads the
  receiver's fields, whereas Java always throws `NullPointerException`.
- Indexing a null array throws `ArrayIndexOutOfBoundsException` rather than
  `NullPointerException`, and `a.length` on a null array is 0.
- Java evaluates operands left to right. Go fixes the order of calls but not of
  plain variable reads relative to calls, so an expression that reads a
  variable which a later operand's call modifies could differ. JTS has no such
  expressions that the oracle has found.
- Arrays do not record their element class: storing an object of the wrong
  class through a covariant array reference does not throw
  `ArrayStoreException`, and `Array.newInstance` makes an array of the static
  type.
- `==` on boxed values compares values; Java compares identity, which only
  agrees for small cached values.
- Static initialization is eager (see Static Fields).

Runtime:

- `Math.random` and `new Random()` are seeded nondeterministically, as in Java,
  so their results differ from run to run.
- Boxed values have no identity, so the identity hash of a boxed value (which
  JTS never requests) does not follow Java's.
- `File.list` returns names in sorted order; Java returns file system order.
- Thread IDs, `Date.toString` time zones, and `Runtime` memory figures come from
  Go's environment.
- Classes are identified by name only: `Class.forName` never throws
  `ClassNotFoundException`.
- Not implemented: Java serialization, reflection (`Class.getMethod`,
  `Class.isAssignableFrom`), `Formatter` date conversions,
  `DecimalFormat` currency patterns, and parts of `java.util.regex`. The JTS
  members that need them (`Debug`, the deprecated
  `GeometryExtracter.toGeometryType`, and test helpers) are stubs.
