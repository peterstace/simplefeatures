import java.util.Set;

/** Go language facts used when choosing identifiers. */
final class Go {
    private Go() {}

    static final Set<String> KEYWORDS = Set.of(
            "break", "case", "chan", "const", "continue", "default", "defer", "else", "fallthrough", "for", "func",
            "go", "goto", "if", "import", "interface", "map", "package", "range", "return", "select", "struct",
            "switch", "type", "var");

    static final Set<String> PREDECLARED = Set.of(
            "any", "append", "bool", "byte", "cap", "clear", "close", "comparable", "complex", "complex128",
            "complex64", "copy", "delete", "error", "false", "float32", "float64", "imag", "int", "int16", "int32",
            "int64", "int8", "iota", "len", "make", "max", "min", "new", "nil", "panic", "print", "println", "real",
            "recover", "rune", "string", "true", "uint", "uint16", "uint32", "uint64", "uint8", "uintptr");

    /** Package names and identifiers that generated code relies on. */
    static final Set<String> RESERVED_LOCALS = Set.of("java", "junit", "unsafe", "math", "testing", "this", "t");
}
