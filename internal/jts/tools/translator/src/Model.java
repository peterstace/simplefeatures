import com.sun.tools.javac.code.Flags;
import com.sun.tools.javac.code.Kinds.Kind;
import com.sun.tools.javac.code.Scope.LookupKind;
import com.sun.tools.javac.code.Symbol;
import com.sun.tools.javac.code.Symbol.ClassSymbol;
import com.sun.tools.javac.code.Symbol.MethodSymbol;
import com.sun.tools.javac.code.Symbol.VarSymbol;
import com.sun.tools.javac.code.Symtab;
import com.sun.tools.javac.code.Type;
import com.sun.tools.javac.code.TypeTag;
import com.sun.tools.javac.code.Types;
import com.sun.tools.javac.tree.JCTree;
import com.sun.tools.javac.tree.JCTree.JCClassDecl;
import com.sun.tools.javac.tree.JCTree.JCCompilationUnit;
import com.sun.tools.javac.tree.TreeScanner;
import com.sun.tools.javac.util.Context;
import com.sun.tools.javac.util.Names;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;

/**
 * Whole-program model of the JTS sources: classes, their Go names, class
 * hierarchies, kind IDs, method names, and virtual dispatch.
 */
final class Model {
    final Types types;
    final Symtab syms;
    final Names names;

    final Map<ClassSymbol, ClassInfo> classes = new LinkedHashMap<>();
    final Map<ClassSymbol, JCCompilationUnit> unitOf = new HashMap<>();
    private final Map<String, ClassSymbol> goNameOwners = new HashMap<>();
    private final Map<MethodSymbol, MethodSymbol> rootMemo = new HashMap<>();

    /**
     * Whether the runtime declares a Go method on a JDK interface type (given
     * by its Java class symbol). Set by the translator once the runtime is
     * scanned.
     */
    java.util.function.BiPredicate<ClassSymbol, String> runtimeInterfaceHas = (c, n) -> false;

    /** Information about one class or interface declared in the JTS sources. */
    static final class ClassInfo {
        ClassSymbol sym;
        JCClassDecl decl;
        String goName;
        boolean emitted;
        boolean test;
        /** Test class emitted into a non-test file because it extends a main class. */
        boolean inMain;
        boolean isInterface;
        ClassInfo superInfo;
        ClassSymbol rtSuper;
        final List<ClassInfo> subs = new ArrayList<>();
        ClassInfo root;
        int kindLo = -1;
        int kindHi = -1;
        ClassInfo enclosingInfo;
        boolean hasOuter;
        final LinkedHashSet<VarSymbol> captured = new LinkedHashSet<>();

        /** Concrete non-private instance methods declared here, by Go name. */
        final Map<String, MethodSymbol> declared = new TreeMap<>();
        /** Virtual members (Go names) of this class. */
        final Set<String> members = new LinkedHashSet<>();
        /** Members first appearing at this class. */
        final Set<String> introduced = new LinkedHashSet<>();
        /** Go names that have a dispatcher at this class. */
        final Map<String, MethodSymbol> dispatchers = new TreeMap<>();
        /** Members whose implementation enters at this class as a forwarder. */
        final Map<String, MethodSymbol> forwarders = new TreeMap<>();
        /** Whether objects of this hierarchy root have a Java monitor (jmon_). */
        boolean hasMonitor;
        /** Interfaces whose marker methods are declared on this class. */
        final Set<ClassSymbol> markers = new LinkedHashSet<>();

        boolean polymorphic() {
            return root.kindLo >= 0;
        }

        boolean isAbstract() {
            return (sym.flags() & Flags.ABSTRACT) != 0;
        }

        boolean isAncestorOrSelfOf(ClassInfo other) {
            for (ClassInfo c = other; c != null; c = c.superInfo) {
                if (c == this) {
                    return true;
                }
            }
            return false;
        }
    }

    Model(Context ctx) {
        types = Types.instance(ctx);
        syms = Symtab.instance(ctx);
        names = Names.instance(ctx);
    }

    void addUnit(JCCompilationUnit unit, boolean emitted, boolean test) {
        new TreeScanner() {
            ClassInfo current;

            @Override
            public void visitClassDef(JCClassDecl tree) {
                ClassInfo info = new ClassInfo();
                info.sym = tree.sym;
                info.decl = tree;
                info.emitted = emitted;
                info.test = test;
                info.isInterface = (tree.sym.flags() & Flags.INTERFACE) != 0;
                info.enclosingInfo = current;
                classes.put(tree.sym, info);
                unitOf.put(tree.sym, unit);
                ClassInfo saved = current;
                current = info;
                super.visitClassDef(tree);
                current = saved;
            }
        }.scan(unit);
    }

    boolean isJts(Symbol s) {
        return s instanceof ClassSymbol cs && classes.containsKey(cs);
    }

    ClassInfo info(Symbol s) {
        return classes.get((ClassSymbol) s);
    }

    void build() {
        for (ClassInfo c : classes.values()) {
            c.goName = goClassName(c.sym);
            ClassSymbol prev = goNameOwners.put(c.goName, c.sym);
            if (prev != null) {
                throw new IllegalStateException("Go name collision: " + c.goName);
            }
        }
        for (ClassInfo c : classes.values()) {
            if (c.isInterface) {
                continue;
            }
            Type sup = c.sym.getSuperclass();
            if (sup != null && sup.hasTag(TypeTag.CLASS)) {
                ClassSymbol ss = (ClassSymbol) sup.tsym;
                if (isJts(ss)) {
                    c.superInfo = classes.get(ss);
                    if (c.emitted) {
                        c.superInfo.subs.add(c);
                    }
                } else if (ss != syms.objectType.tsym) {
                    c.rtSuper = ss;
                }
            }
            c.hasOuter = c.sym.hasOuterInstance();
        }
        for (ClassInfo c : classes.values()) {
            c.subs.sort(Comparator.comparing(x -> x.goName));
        }
        List<ClassInfo> roots = new ArrayList<>();
        for (ClassInfo c : classes.values()) {
            if (!c.isInterface && c.superInfo == null) {
                roots.add(c);
            }
        }
        roots.sort(Comparator.comparing(x -> x.goName));
        int[] next = {1};
        for (ClassInfo r : roots) {
            if (r.subs.isEmpty()) {
                setRoot(r, r);
                continue;
            }
            number(r, r, next);
        }
        for (ClassInfo r : roots) {
            computeMembers(r);
        }
        computeMarkers();
        computeCaptures();
        computeMonitors();
        for (ClassInfo c : classes.values()) {
            if (c.test && c.superInfo != null && !c.superInfo.test) {
                markInMain(c);
            }
        }
    }

    private void markInMain(ClassInfo c) {
        c.inMain = true;
        for (ClassInfo s : c.subs) {
            markInMain(s);
        }
    }

    /** Whether code for c is in a _test.go file. */
    static boolean inTestFile(ClassInfo c) {
        return c.test && !c.inMain;
    }

    private void setRoot(ClassInfo c, ClassInfo root) {
        c.root = root;
    }

    private void number(ClassInfo c, ClassInfo root, int[] next) {
        c.root = root;
        c.kindLo = next[0]++;
        for (ClassInfo s : c.subs) {
            number(s, root, next);
        }
        c.kindHi = next[0] - 1;
    }

    /** Concrete leaf kinds in the subtree of c. */
    List<ClassInfo> concreteSubtree(ClassInfo c) {
        List<ClassInfo> out = new ArrayList<>();
        collectConcrete(c, out);
        return out;
    }

    private void collectConcrete(ClassInfo c, List<ClassInfo> out) {
        if (!c.isAbstract()) {
            out.add(c);
        }
        for (ClassInfo s : c.subs) {
            collectConcrete(s, out);
        }
    }

    // ---------------------------------------------------------------------
    // Naming

    String goClassName(ClassSymbol sym) {
        String pkg = sym.packge().fullname.toString();
        String flat = sym.flatname.toString();
        String local = pkg.isEmpty() ? flat : flat.substring(pkg.length() + 1);
        return pkgPrefix(pkg) + "_" + local.replace('$', '_');
    }

    static String pkgPrefix(String pkg) {
        String rest;
        if (pkg.equals("org.locationtech.jts")) {
            rest = "jts";
        } else if (pkg.startsWith("org.locationtech.jts.")) {
            rest = pkg.substring("org.locationtech.jts.".length());
        } else {
            rest = pkg;
        }
        StringBuilder sb = new StringBuilder();
        for (String part : rest.split("\\.")) {
            sb.append(Character.toUpperCase(part.charAt(0))).append(part.substring(1));
        }
        // go test treats functions named TestX, BenchmarkX, ExampleX, and
        // FuzzX (X not lowercase) specially.
        for (String p : new String[] {"Test", "Benchmark", "Example", "Fuzz"}) {
            if (sb.length() > p.length() && sb.indexOf(p) == 0) {
                sb.setCharAt(p.length(), Character.toLowerCase(sb.charAt(p.length())));
            }
        }
        return sb.toString();
    }

    static String capitalize(String s) {
        return Character.toUpperCase(s.charAt(0)) + s.substring(1);
    }

    static String lowerFirst(String s) {
        return Character.toLowerCase(s.charAt(0)) + s.substring(1);
    }

    String sigName(Type t) {
        t = types.erasure(t);
        return switch (t.getTag()) {
            case BOOLEAN -> "boolean";
            case BYTE -> "byte";
            case SHORT -> "short";
            case CHAR -> "char";
            case INT -> "int";
            case LONG -> "long";
            case FLOAT -> "float";
            case DOUBLE -> "double";
            case ARRAY -> sigName(((Type.ArrayType) t).elemtype) + "Arr";
            case CLASS -> t.tsym.name.toString();
            default -> throw new IllegalStateException("sigName " + t);
        };
    }

    String sigSuffix(MethodSymbol m) {
        StringBuilder sb = new StringBuilder();
        for (Type p : types.erasure(m.type).getParameterTypes()) {
            sb.append('_').append(sigName(p));
        }
        return sb.toString();
    }

    /** The declaration that m ultimately overrides (m itself if none). */
    MethodSymbol root(MethodSymbol m) {
        MethodSymbol memo = rootMemo.get(m);
        if (memo != null) {
            return memo;
        }
        MethodSymbol result = m;
        if (!m.isStatic() && !m.isConstructor() && (m.flags() & Flags.PRIVATE) == 0) {
            ClassSymbol owner = (ClassSymbol) m.owner;
            for (Type st : types.closure(owner.type)) {
                if (st.tsym == owner) {
                    continue;
                }
                for (Symbol s : st.tsym.members().getSymbolsByName(m.name, LookupKind.NON_RECURSIVE)) {
                    if (s instanceof MethodSymbol ms && !ms.isStatic() && !ms.isConstructor()
                            && (ms.flags() & Flags.PRIVATE) == 0 && m.overrides(ms, owner, types, false)) {
                        MethodSymbol r = root(ms);
                        if (result == m || isJts(result.owner) && !isJts(r.owner)) {
                            result = r;
                        }
                    }
                }
            }
            if (m.owner.isInterface() || owner != syms.objectType.tsym) {
                // Interface methods that redeclare Object methods root at Object.
                for (Symbol s : syms.objectType.tsym.members().getSymbolsByName(m.name, LookupKind.NON_RECURSIVE)) {
                    if (s instanceof MethodSymbol ms && (ms.flags() & Flags.PUBLIC) != 0
                            && types.isSameTypes(types.erasure(ms.type).getParameterTypes(),
                                    types.erasure(m.type).getParameterTypes())) {
                        result = ms;
                    }
                }
            }
        }
        rootMemo.put(m, result);
        return result;
    }

    boolean isObjectMethod(MethodSymbol m) {
        return root(m).owner == syms.objectType.tsym;
    }

    /** Go name of an instance method (as called on any receiver). */
    String instanceMethodName(MethodSymbol m) {
        if ((m.flags() & Flags.PRIVATE) != 0 && isJts(m.owner)) {
            String n = lowerFirst(m.name.toString()) + sigSuffix(m);
            if (fieldNamed((ClassSymbol) m.owner, n)) {
                n += "_m";
            }
            return goIdent(n);
        }
        MethodSymbol r = root(m);
        return capitalize(r.name.toString()) + sigSuffix(r);
    }

    private boolean fieldNamed(ClassSymbol c, String name) {
        for (Symbol s : c.members().getSymbolsByName(names.fromString(name), LookupKind.NON_RECURSIVE)) {
            if (s.kind == Kind.VAR) {
                return true;
            }
        }
        return false;
    }

    /** Owner prefix for static members and constructors. */
    String ownerPrefix(ClassSymbol c) {
        if (isJts(c)) {
            return classes.get(c).goName;
        }
        return Runtime.qualifiedName(c);
    }

    String staticMethodName(MethodSymbol m) {
        ClassSymbol owner = (ClassSymbol) m.owner;
        int count = 0;
        for (Symbol s : owner.members().getSymbolsByName(m.name, LookupKind.NON_RECURSIVE)) {
            if (s instanceof MethodSymbol ms && ms.isStatic()) {
                count++;
            }
        }
        String n = ownerPrefix(owner) + "_" + m.name;
        return count > 1 ? n + sigSuffix(m) : n;
    }

    private boolean ctorOverloaded(ClassSymbol owner) {
        int count = 0;
        for (Symbol s : owner.members().getSymbolsByName(names.init, LookupKind.NON_RECURSIVE)) {
            if (s instanceof MethodSymbol) {
                count++;
            }
        }
        return count > 1;
    }

    String ctorNewName(MethodSymbol ctor) {
        ClassSymbol owner = (ClassSymbol) ctor.owner;
        String n = ownerPrefix(owner) + "_new";
        return ctorOverloaded(owner) ? n + sigSuffix(ctor) : n;
    }

    String ctorInitName(MethodSymbol ctor) {
        ClassSymbol owner = (ClassSymbol) ctor.owner;
        String base = isJts(owner) ? classes.get(owner).goName : owner.name.toString();
        String n = base + "_init";
        return ctorOverloaded(owner) ? n + sigSuffix(ctor) : n;
    }

    String staticFieldName(VarSymbol v) {
        return ownerPrefix((ClassSymbol) v.owner) + "_" + v.name;
    }

    String fieldName(VarSymbol v) {
        String n = v.name.toString();
        return Go.KEYWORDS.contains(n) ? n + "_" : n;
    }

    static String goIdent(String n) {
        return Go.KEYWORDS.contains(n) || Go.PREDECLARED.contains(n) || Go.RESERVED_LOCALS.contains(n) ? n + "_" : n;
    }

    // ---------------------------------------------------------------------
    // Virtual dispatch

    boolean isVirtualCandidate(MethodSymbol m) {
        long f = m.flags();
        return !m.isStatic() && !m.isConstructor() && (f & Flags.PRIVATE) == 0;
    }

    private void computeMembers(ClassInfo c) {
        if (c.superInfo != null) {
            c.members.addAll(c.superInfo.members);
        }
        Map<String, MethodSymbol> newMembers = new LinkedHashMap<>();
        for (Symbol s : c.sym.members().getSymbols(LookupKind.NON_RECURSIVE)) {
            if (s instanceof MethodSymbol m && isVirtualCandidate(m)) {
                String n = instanceMethodName(m);
                if ((m.flags() & Flags.ABSTRACT) == 0) {
                    c.declared.put(n, m);
                }
                newMembers.putIfAbsent(n, m);
            }
        }
        Set<String> declaredInHierarchy = new java.util.HashSet<>();
        collectDeclared(c.root, declaredInHierarchy);
        for (Type it : types.closure(c.sym.type)) {
            if (it.tsym.isInterface()) {
                boolean jts = isJts(it.tsym);
                for (Symbol s : it.tsym.members().getSymbols(LookupKind.NON_RECURSIVE)) {
                    if (!(s instanceof MethodSymbol m) || !isVirtualCandidate(m) || isObjectMethod(m)) {
                        continue;
                    }
                    String n = instanceMethodName(m);
                    // A JDK interface method is only a member when JTS code
                    // implements it, or when it is a default method that the
                    // runtime's Go interface requires and no runtime
                    // superclass supplies; otherwise a runtime superclass
                    // implements it.
                    boolean isDefault = (m.flags() & Flags.DEFAULT) != 0;
                    if (jts || !isDefault && declaredInHierarchy.contains(n)
                            || isDefault && c.root.rtSuper == null
                                    && runtimeInterfaceHas.test((ClassSymbol) it.tsym, n)) {
                        newMembers.putIfAbsent(n, m);
                    }
                }
            }
        }
        if (c.superInfo == null) {
            // Methods inherited from outside JTS become members of the root
            // when some class in the hierarchy overrides them.
            Set<String> overriddenInHierarchy = new LinkedHashSet<>();
            collectOverriddenExternal(c, overriddenInHierarchy, newMembers);
        }
        for (Map.Entry<String, MethodSymbol> e : newMembers.entrySet()) {
            if (c.members.add(e.getKey())) {
                c.introduced.add(e.getKey());
            }
        }
        for (ClassInfo s : c.subs) {
            computeMembers(s);
        }
        for (String n : c.introduced) {
            MethodSymbol m = newMembers.get(n);
            boolean overriddenBelow = false;
            for (ClassInfo s : c.subs) {
                if (declaresInSubtree(s, n)) {
                    overriddenBelow = true;
                    break;
                }
            }
            boolean hasImplHere = c.declared.containsKey(n);
            if (!hasImplHere) {
                MethodSymbol fwd = inheritedImpl(c, m);
                if (fwd != null) {
                    c.forwarders.put(n, fwd);
                    hasImplHere = true;
                }
            }
            if (overriddenBelow || !hasImplHere) {
                c.dispatchers.put(n, m);
            }
        }
    }

    private void collectOverriddenExternal(ClassInfo c, Set<String> seen, Map<String, MethodSymbol> out) {
        for (Symbol s : c.sym.members().getSymbols(LookupKind.NON_RECURSIVE)) {
            if (s instanceof MethodSymbol m && isVirtualCandidate(m)) {
                MethodSymbol r = root(m);
                if (!isJts(r.owner) && !r.owner.isInterface() && r != m) {
                    out.putIfAbsent(instanceMethodName(m), r);
                }
            }
        }
        for (ClassInfo s : c.subs) {
            collectOverriddenExternal(s, seen, out);
        }
    }

    private void collectDeclared(ClassInfo c, Set<String> out) {
        out.addAll(c.declared.isEmpty() ? declaredNames(c) : c.declared.keySet());
        for (ClassInfo s : c.subs) {
            collectDeclared(s, out);
        }
    }

    private Set<String> declaredNames(ClassInfo c) {
        Set<String> out = new java.util.HashSet<>();
        for (Symbol s : c.sym.members().getSymbols(LookupKind.NON_RECURSIVE)) {
            if (s instanceof MethodSymbol m && isVirtualCandidate(m) && (m.flags() & Flags.ABSTRACT) == 0) {
                out.add(instanceMethodName(m));
            }
        }
        return out;
    }

    private boolean declaresInSubtree(ClassInfo c, String n) {
        if (c.declared.containsKey(n)) {
            return true;
        }
        for (ClassInfo s : c.subs) {
            if (declaresInSubtree(s, n)) {
                return true;
            }
        }
        return false;
    }

    /**
     * Implementation of member m at class c that does not come from a JTS
     * class: an interface default method or a method of a runtime superclass.
     */
    private MethodSymbol inheritedImpl(ClassInfo c, MethodSymbol m) {
        if (!isJts(m.owner) && !m.owner.isInterface()) {
            return m;
        }
        for (Type it : types.closure(c.sym.type)) {
            if (it.tsym.isInterface()) {
                for (Symbol s : it.tsym.members().getSymbolsByName(m.name, LookupKind.NON_RECURSIVE)) {
                    if (s instanceof MethodSymbol dm && (dm.flags() & Flags.DEFAULT) != 0
                            && instanceMethodName(dm).equals(instanceMethodName(m))) {
                        return dm;
                    }
                }
            }
        }
        return null;
    }

    /** The class at or above c where member n is introduced. */
    ClassInfo introducer(ClassInfo c, String n) {
        ClassInfo found = null;
        for (ClassInfo x = c; x != null; x = x.superInfo) {
            if (x.introduced.contains(n)) {
                found = x;
            }
        }
        return found;
    }

    boolean overriddenStrictlyBelow(ClassInfo c, String n) {
        for (ClassInfo s : c.subs) {
            if (declaresInSubtree(s, n)) {
                return true;
            }
        }
        return false;
    }

    /** Whether some class at or above c (up to the introducer) has an _impl for n. */
    boolean hasImplAtOrAbove(ClassInfo c, String n, ClassInfo introducer) {
        for (ClassInfo x = c; x != null; x = x.superInfo) {
            if (x.declared.containsKey(n) || x.forwarders.containsKey(n)) {
                return true;
            }
            if (x == introducer) {
                break;
            }
        }
        return false;
    }

    /** Class whose N_impl applies to concrete class k, searching up to stop. */
    ClassInfo implClass(ClassInfo k, String n, ClassInfo stop) {
        for (ClassInfo x = k; x != null; x = x.superInfo) {
            if (x.declared.containsKey(n) || x.forwarders.containsKey(n)) {
                return x;
            }
            if (x == stop) {
                break;
            }
        }
        return null;
    }

    // ---------------------------------------------------------------------
    // Monitors

    /**
     * Gives a monitor to each hierarchy that has synchronized instance methods
     * or whose objects are locked by synchronized blocks.
     */
    private void computeMonitors() {
        for (ClassInfo c : classes.values()) {
            if (c.isInterface) {
                continue;
            }
            for (Symbol s : c.sym.members().getSymbols(LookupKind.NON_RECURSIVE)) {
                if (s instanceof MethodSymbol m && !m.isStatic() && (m.flags() & Flags.SYNCHRONIZED) != 0) {
                    c.root.hasMonitor = true;
                }
            }
            new TreeScanner() {
                @Override
                public void visitSynchronized(JCTree.JCSynchronized tree) {
                    Symbol s = types.erasure(tree.lock.type).tsym;
                    if (isJts(s) && !classes.get(s).isInterface) {
                        classes.get(s).root.hasMonitor = true;
                    }
                    super.visitSynchronized(tree);
                }
            }.scan(c.decl);
        }
    }

    // ---------------------------------------------------------------------
    // Interface markers

    private void computeMarkers() {
        for (ClassInfo c : classes.values()) {
            Set<ClassSymbol> inherited = new LinkedHashSet<>();
            if (c.superInfo != null) {
                for (Type it : types.closure(c.superInfo.sym.type)) {
                    if (it.tsym.isInterface()) {
                        inherited.add((ClassSymbol) it.tsym);
                    }
                }
            }
            for (Type it : types.closure(c.sym.type)) {
                if (it.tsym.isInterface() && it.tsym != c.sym && isJts(it.tsym) && !inherited.contains(it.tsym)) {
                    c.markers.add((ClassSymbol) it.tsym);
                }
            }
        }
    }

    // ---------------------------------------------------------------------
    // Captured variables of local and anonymous classes

    private void computeCaptures() {
        for (ClassInfo c : classes.values()) {
            if (c.sym.isDirectlyOrIndirectlyLocal()) {
                collectCaptures(c);
            }
        }
    }

    private void collectCaptures(ClassInfo c) {
        Set<VarSymbol> declaredInside = new java.util.HashSet<>();
        List<VarSymbol> referenced = new ArrayList<>();
        new TreeScanner() {
            @Override
            public void visitVarDef(JCTree.JCVariableDecl tree) {
                declaredInside.add(tree.sym);
                super.visitVarDef(tree);
            }

            @Override
            public void visitIdent(JCTree.JCIdent tree) {
                if (tree.sym instanceof VarSymbol v && v.owner.kind == Kind.MTH) {
                    referenced.add(v);
                }
            }
        }.scan(c.decl);
        for (VarSymbol v : referenced) {
            if (!declaredInside.contains(v) && v.getConstValue() == null) {
                c.captured.add(v);
            }
        }
    }
}
