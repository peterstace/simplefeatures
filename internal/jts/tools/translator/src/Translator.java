import com.sun.tools.javac.code.Flags;
import com.sun.tools.javac.code.Symbol;
import com.sun.tools.javac.code.Symbol.ClassSymbol;
import com.sun.tools.javac.code.Symbol.MethodSymbol;
import com.sun.tools.javac.code.Type;
import com.sun.tools.javac.code.TypeTag;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;

/** Shared state and type/conversion logic for translating JTS to Go. */
final class Translator {
    final Model model;
    final Runtime rt;
    final Map<String, Integer> literals = new LinkedHashMap<>();
    /** Unsupported reason -> occurrences (method descriptions). */
    final Map<String, List<String>> unsupported = new TreeMap<>();
    /** Runtime identifier -> [uses, supported]. */
    final Map<String, int[]> runtimeRefs = new TreeMap<>();

    private static final Set<String> ANY_TYPES = Set.of(
            "java.lang.Object", "java.lang.Comparable", "java.lang.Cloneable", "java.io.Serializable",
            "java.lang.Number", "java.lang.Integer", "java.lang.Double", "java.lang.Long", "java.lang.Short",
            "java.lang.Byte", "java.lang.Character", "java.lang.Boolean", "java.lang.Float",
            "java.lang.CharSequence");

    /** In lenient mode missing runtime members are recorded but not fatal. */
    boolean lenient;

    /** JTS functional interfaces that lambdas implement, needing adapter types. */
    final java.util.Set<Model.ClassInfo> lambdaInterfaces = new java.util.TreeSet<>(
            java.util.Comparator.comparing(c -> c.goName));

    /** Whether the code being emitted goes into a _test.go file. */
    boolean emittingTest;

    Translator(Model model, Runtime rt) {
        this.model = model;
        this.rt = rt;
    }

    // ---------------------------------------------------------------------
    // Representations

    enum Kind { PRIM, ANY, JCLASS, JIFACE, STRING, RTCLASS, RTIFACE, ARRAY, VOID }

    record Repr(Kind kind, String go, Model.ClassInfo info, Type elem) {}

    Repr repr(Type t) {
        t = model.types.erasure(t);
        switch (t.getTag()) {
            case BOOLEAN: return new Repr(Kind.PRIM, "bool", null, null);
            case BYTE: return new Repr(Kind.PRIM, "int8", null, null);
            case SHORT: return new Repr(Kind.PRIM, "int16", null, null);
            case CHAR: return new Repr(Kind.PRIM, "uint16", null, null);
            case INT: return new Repr(Kind.PRIM, "int32", null, null);
            case LONG: return new Repr(Kind.PRIM, "int64", null, null);
            case FLOAT: return new Repr(Kind.PRIM, "float32", null, null);
            case DOUBLE: return new Repr(Kind.PRIM, "float64", null, null);
            case VOID: return new Repr(Kind.VOID, "", null, null);
            case BOT: return new Repr(Kind.ANY, "any", null, null);
            case ARRAY: {
                Type elem = ((Type.ArrayType) t).elemtype;
                return new Repr(Kind.ARRAY, "[]" + repr(elem).go, null, elem);
            }
            case CLASS: {
                ClassSymbol cs = (ClassSymbol) t.tsym;
                String full = cs.getQualifiedName().toString();
                if (model.isJts(cs)) {
                    Model.ClassInfo ci = model.info(cs);
                    if (!visible(ci)) {
                        note("invisible class " + ci.goName);
                        return new Repr(Kind.ANY, "any", null, null);
                    }
                    if (ci.isInterface) {
                        return new Repr(Kind.JIFACE, ci.goName, ci, null);
                    }
                    return new Repr(Kind.JCLASS, "*" + ci.goName, ci, null);
                }
                if (full.equals("java.lang.String")) {
                    return new Repr(Kind.STRING, "*java.String", null, null);
                }
                if (ANY_TYPES.contains(full)) {
                    return new Repr(Kind.ANY, "any", null, null);
                }
                String q = Runtime.qualifiedName(cs);
                if (q != null && rt.hasType(q)) {
                    if (cs.isInterface()) {
                        return new Repr(Kind.RTIFACE, q, null, null);
                    }
                    return new Repr(Kind.RTCLASS, "*" + q, null, null);
                }
                note("runtime type " + (q != null ? q : full) + (cs.isInterface() ? " interface" : " class"));
                return new Repr(Kind.ANY, "any", null, null);
            }
            default:
                throw new Unsupported("type " + t);
        }
    }

    /** Whether generated code being emitted may refer to class c. */
    boolean visible(Model.ClassInfo c) {
        return c.emitted && (emittingTest || !Model.inTestFile(c));
    }

    /** Fails if class c may not be referenced from the code being emitted. */
    void requireVisible(Model.ClassInfo c) {
        if (!visible(c)) {
            throw new Unsupported((c.emitted ? "test class " : "excluded class ") + c.goName);
        }
    }

    /** The Go type converting a function literal to functional interface fi. */
    String lambdaAdapter(Type fi) {
        Symbol s = model.types.erasure(fi).tsym;
        if (model.isJts(s)) {
            Model.ClassInfo ci = model.info(s);
            requireVisible(ci);
            lambdaInterfaces.add(ci);
            return ci.goName + "_jfunc";
        }
        String q = Runtime.qualifiedName((ClassSymbol) s);
        if (q == null || !rt.hasType(q + "Func")) {
            throw new Unsupported("runtime " + (q == null ? s : q + "Func"));
        }
        return q + "Func";
    }

    /** The Go type of a Java volatile instance field, or null if v is not volatile. */
    String volatileType(com.sun.tools.javac.code.Symbol.VarSymbol v) {
        if ((v.flags() & com.sun.tools.javac.code.Flags.VOLATILE) == 0 || v.isStatic()) {
            return null;
        }
        Repr r = repr(v.type);
        if (r.kind == Kind.PRIM && r.go.equals("bool")) {
            return "java.VolatileBool";
        }
        if (r.kind == Kind.JCLASS) {
            return "java.VolatilePtr[" + r.go.substring(1) + "]";
        }
        throw new Unsupported("volatile field of type " + v.type);
    }

    String goType(Type t) {
        return repr(t).go;
    }

    boolean isFloating(Type t) {
        return t.hasTag(TypeTag.FLOAT) || t.hasTag(TypeTag.DOUBLE);
    }

    boolean isIntegral(Type t) {
        return switch (t.getTag()) {
            case BYTE, SHORT, CHAR, INT, LONG -> true;
            default -> false;
        };
    }

    boolean isString(Type t) {
        return t.tsym != null && t.tsym.getQualifiedName().contentEquals("java.lang.String");
    }

    // ---------------------------------------------------------------------
    // Conversions

    /** Converts Go expression e of Java type from into Java type to. */
    String conv(String e, Type from, Type to) {
        if (to.hasTag(TypeTag.VOID)) {
            return e;
        }
        if (from.hasTag(TypeTag.BOT)) {
            if (to.isPrimitive()) {
                throw new Unsupported("null to primitive");
            }
            return "nil";
        }
        if (from.isPrimitive() && to.isPrimitive()) {
            return primConv(e, from, to);
        }
        if (from.isPrimitive()) {
            Type target = model.types.unboxedType(to);
            Type prim = target.hasTag(TypeTag.NONE) ? from : target;
            return "any(" + typedPrim(primConv(e, from, prim), prim) + ")";
        }
        if (to.isPrimitive()) {
            Type prim = model.types.unboxedType(from);
            if (prim.hasTag(TypeTag.NONE)) {
                // e.g. a cast from Object to int goes through the box type.
                prim = to;
            }
            return primConv("(" + e + ").(" + goType(prim) + ")", prim, to);
        }
        return refConv(e, model.types.erasure(from), model.types.erasure(to));
    }

    private String typedPrim(String e, Type prim) {
        String g = goType(prim);
        return e.startsWith(g + "(") ? e : g + "(" + e + ")";
    }

    String primConv(String e, Type from, Type to) {
        TypeTag f = from.getTag();
        TypeTag t = to.getTag();
        if (f == t) {
            return e;
        }
        if (f == TypeTag.BOOLEAN || t == TypeTag.BOOLEAN) {
            throw new Unsupported("boolean conversion");
        }
        if (isFloating(from) && isIntegral(to)) {
            String wide = f == TypeTag.FLOAT ? "float64(" + e + ")" : e;
            if (t == TypeTag.LONG) {
                return "java.F2L(" + wide + ")";
            }
            String i = "java.F2I(" + wide + ")";
            return t == TypeTag.INT ? i : goType(to) + "(" + i + ")";
        }
        return goType(to) + "(" + e + ")";
    }

    String refConv(String e, Type from, Type to) {
        Repr f = repr(from);
        Repr t = repr(to);
        if (f.go.equals(t.go)) {
            return e;
        }
        switch (t.kind) {
            case ANY:
                return toAny(e, f);
            case JCLASS: {
                if (f.kind == Kind.JCLASS) {
                    if (t.info.isAncestorOrSelfOf(f.info)) {
                        return upcast(e, t.info);
                    }
                    if (f.info.isAncestorOrSelfOf(t.info)) {
                        return t.info.goName + "_downcast(" + upcast(e, t.info.root) + ")";
                    }
                    throw new Unsupported("unrelated class conversion");
                }
                return t.info.goName + "_fromAny(" + toAny(e, f) + ")";
            }
            case JIFACE:
            case RTIFACE: {
                if (f.kind == Kind.JIFACE && t.kind == Kind.JIFACE && isSubInterface(f.info.sym, t.info.sym)) {
                    return e;
                }
                return "java.AnyTo[" + t.go + "](" + toAny(e, f) + ")";
            }
            case STRING:
            case RTCLASS: {
                if (t.kind == Kind.RTCLASS && (f.kind == Kind.ANY || f.kind == Kind.JIFACE || f.kind == Kind.RTIFACE)) {
                    return requireFunc(t.go.substring(1) + "_fromAny") + "(" + e + ")";
                }
                if (f.kind == Kind.JCLASS && t.kind == Kind.RTCLASS) {
                    return "(" + t.go + ")(unsafe.Pointer(" + e + "))";
                }
                if (f.kind == Kind.RTCLASS && t.kind == Kind.RTCLASS) {
                    if (model.types.isSubtype(from, to)) {
                        return "(" + t.go + ")(unsafe.Pointer(" + e + "))";
                    }
                }
                return "java.AnyTo[" + t.go + "](" + toAny(e, f) + ")";
            }
            case ARRAY: {
                if (f.kind == Kind.ARRAY) {
                    Repr fe = repr(f.elem);
                    Repr te = repr(t.elem);
                    if (fe.kind == Kind.JCLASS && te.kind == Kind.JCLASS) {
                        return "java.SliceCast[" + te.go + "](" + e + ")";
                    }
                    return "java.MapSlice(" + e + ", func(x " + fe.go + ") " + te.go + " { return "
                            + conv("x", f.elem, t.elem) + " })";
                }
                if (t.go.equals("[]any")) {
                    // A Java Object[] may hold an array of any element type.
                    return "java.ObjArr(" + toAny(e, f) + ")";
                }
                return "java.AnyTo[" + t.go + "](" + toAny(e, f) + ")";
            }
            default:
                throw new Unsupported("conversion to " + to);
        }
    }

    boolean isSubInterface(ClassSymbol sub, ClassSymbol sup) {
        return sub.isSubClass(sup, model.types);
    }

    String toAny(String e, Repr f) {
        return switch (f.kind) {
            case ANY, JIFACE, RTIFACE -> e;
            case JCLASS -> f.info.polymorphic()
                    ? f.info.root.goName + "_toAny(" + upcast(e, f.info.root) + ")"
                    : "java.Ref(" + e + ")";
            case STRING, RTCLASS -> "java.Ref(" + e + ")";
            case ARRAY -> "java.ArrRef(" + e + ")";
            default -> throw new Unsupported("toAny " + f.kind);
        };
    }

    String toAny(String e, Type t) {
        return toAny(e, repr(t));
    }

    String upcast(String e, Model.ClassInfo target) {
        return "(*" + target.goName + ")(unsafe.Pointer(" + e + "))";
    }

    // ---------------------------------------------------------------------
    // Constants and literals

    String constant(Type t, Object v) {
        switch (t.getTag()) {
            case BOOLEAN:
                return ((Integer) v) != 0 ? "true" : "false";
            case CHAR:
                return "uint16(" + ((Number) v).intValue() + ")";
            case BYTE:
                return "int8(" + ((Number) v).intValue() + ")";
            case SHORT:
                return "int16(" + ((Number) v).intValue() + ")";
            case INT:
                return "int32(" + ((Number) v).intValue() + ")";
            case LONG:
                return "int64(" + ((Number) v).longValue() + ")";
            case DOUBLE:
                return doubleLit(((Number) v).doubleValue());
            case FLOAT:
                return floatLit(((Number) v).floatValue());
            default:
                if (isString(t)) {
                    return literal((String) v);
                }
                throw new Unsupported("constant of type " + t);
        }
    }

    static String doubleLit(double d) {
        if (Double.isNaN(d)) {
            // Go's math.NaN() has different bits from Java's Double.NaN.
            return "math.Float64frombits(0x" + Long.toHexString(Double.doubleToRawLongBits(d)) + ")";
        }
        if (Double.isInfinite(d)) {
            return d > 0 ? "math.Inf(1)" : "math.Inf(-1)";
        }
        if (d == 0 && Double.doubleToRawLongBits(d) != 0) {
            return "math.Copysign(0, -1)";
        }
        return "float64(" + Double.toString(d) + ")";
    }

    static String floatLit(float f) {
        if (Float.isNaN(f)) {
            return "math.Float32frombits(0x" + Integer.toHexString(Float.floatToRawIntBits(f)) + ")";
        }
        if (Float.isInfinite(f) || f == 0 && Float.floatToRawIntBits(f) != 0) {
            return "float32(" + doubleLit(f) + ")";
        }
        return "float32(" + Float.toString(f) + ")";
    }

    /** Whether a constant of type t can be a Go constant declaration. */
    static boolean goConstable(Type t, Object v) {
        return switch (t.getTag()) {
            case DOUBLE -> {
                double d = ((Number) v).doubleValue();
                yield !Double.isNaN(d) && !Double.isInfinite(d) && !(d == 0 && Double.doubleToRawLongBits(d) != 0);
            }
            case FLOAT -> {
                float f = ((Number) v).floatValue();
                yield !Float.isNaN(f) && !Float.isInfinite(f) && !(f == 0 && Float.floatToRawIntBits(f) != 0);
            }
            case BOOLEAN, CHAR, BYTE, SHORT, INT, LONG -> true;
            default -> false;
        };
    }

    String literal(String s) {
        Integer id = literals.get(s);
        if (id == null) {
            id = literals.size();
            literals.put(s, id);
        }
        return "jlit_" + id;
    }

    static String goQuote(String s) {
        StringBuilder sb = new StringBuilder("\"");
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"' -> sb.append("\\\"");
                case '\\' -> sb.append("\\\\");
                case '\n' -> sb.append("\\n");
                case '\t' -> sb.append("\\t");
                case '\r' -> sb.append("\\r");
                default -> {
                    if (c < 0x20 || c > 0x7e) {
                        sb.append(String.format("\\u%04x", (int) c));
                    } else {
                        sb.append(c);
                    }
                }
            }
        }
        return sb.append('"').toString();
    }

    // ---------------------------------------------------------------------
    // Runtime references

    /** Records a reference to a runtime identifier and fails if it is missing. */
    String requireFunc(String qualified) {
        boolean ok = rt.hasFunc(qualified);
        record(qualified, ok);
        if (!ok && !lenient) {
            throw new Unsupported("runtime " + qualified);
        }
        return qualified;
    }

    void requireMethod(String qualifiedType, String method) {
        boolean ok = rt.hasMethod(qualifiedType, method);
        record(qualifiedType + "." + method, ok);
        if (!ok && !lenient) {
            throw new Unsupported("runtime " + qualifiedType + "." + method);
        }
    }

    private void record(String ref, boolean ok) {
        int[] c = runtimeRefs.computeIfAbsent(ref, k -> new int[2]);
        c[0]++;
        c[1] = ok ? 1 : 0;
    }

    private void note(String reason) {
        int[] c = runtimeRefs.computeIfAbsent(reason, k -> new int[2]);
        c[0]++;
    }

    void reportUnsupported(String reason, String where) {
        unsupported.computeIfAbsent(reason, k -> new java.util.ArrayList<>()).add(where);
    }

    // ---------------------------------------------------------------------
    // Method helpers

    /** Erased type of the root declaration of m (the Go signature). */
    Type rootErasure(MethodSymbol m) {
        if (m.isStatic() || m.isConstructor()) {
            return model.types.erasure(m.type);
        }
        return model.types.erasure(model.root(m).type);
    }

    /**
     * The Go signature of a call to m. It is the erased root declaration,
     * except that for JDK methods Object[] becomes Object: the runtime takes
     * Java arrays of any element type as a Go any, so that arrays passed to
     * methods like Arrays.sort are not copied.
     */
    Type goSignature(MethodSymbol m) {
        Type t = rootErasure(m);
        MethodSymbol root = m.isStatic() || m.isConstructor() ? m : model.root(m);
        if (model.isJts(root.owner)) {
            return t;
        }
        com.sun.tools.javac.util.ListBuffer<Type> params = new com.sun.tools.javac.util.ListBuffer<>();
        for (Type p : t.getParameterTypes()) {
            params.append(objectArrayToObject(p));
        }
        return new Type.MethodType(params.toList(), objectArrayToObject(t.getReturnType()), t.getThrownTypes(),
                model.syms.methodClass);
    }

    private Type objectArrayToObject(Type t) {
        if (t.hasTag(TypeTag.ARRAY) && ((Type.ArrayType) t).elemtype.tsym == model.syms.objectType.tsym) {
            return model.syms.objectType;
        }
        return t;
    }

    boolean isPrivate(Symbol s) {
        return (s.flags() & Flags.PRIVATE) != 0;
    }
}
