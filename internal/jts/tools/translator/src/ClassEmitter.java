import com.sun.tools.javac.code.Flags;
import com.sun.tools.javac.code.Scope.LookupKind;
import com.sun.tools.javac.code.Symbol;
import com.sun.tools.javac.code.Symbol.ClassSymbol;
import com.sun.tools.javac.code.Symbol.MethodSymbol;
import com.sun.tools.javac.code.Symbol.VarSymbol;
import com.sun.tools.javac.code.Type;
import com.sun.tools.javac.code.TypeTag;
import com.sun.tools.javac.tree.JCTree;
import com.sun.tools.javac.tree.JCTree.*;
import com.sun.tools.javac.tree.TreeInfo;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.StringJoiner;

/** Emits the Go declarations for one Java class or interface. */
final class ClassEmitter {
    private final Translator tr;
    private final Model model;
    private final StringBuilder out = new StringBuilder();

    ClassEmitter(Translator tr) {
        this.tr = tr;
        this.model = tr.model;
    }

    String text() {
        return out.toString();
    }

    private void w(String s) {
        out.append(s).append('\n');
    }

    void emit(Model.ClassInfo c) {
        w("// " + c.sym.flatname);
        if (c.isInterface) {
            emitInterface(c);
            emitStatics(c);
            emitStaticMethods(c);
            emitDefaultMethods(c);
            return;
        }
        emitStruct(c);
        emitIdentity(c);
        emitStatics(c);
        if ((c.sym.flags() & Flags.ENUM) != 0) {
            emitEnumMethods(c);
        }
        emitCtors(c);
        emitStaticMethods(c);
        emitInstanceMethods(c);
        emitDispatchers(c);
        emitForwarders(c);
        emitMarkers(c);
        if (c.test) {
            emitTests(c);
        }
    }

    private String where(Model.ClassInfo c, String member) {
        return c.sym.flatname + "." + member;
    }

    private String stub(RuntimeException u, Model.ClassInfo c, String member) {
        String reason = u instanceof Unsupported ? u.getMessage() : "translator bug: " + u;
        if (!(u instanceof Unsupported)) {
            StackTraceElement[] st = u.getStackTrace();
            reason += st.length > 0 ? " at " + st[0] : "";
        }
        tr.reportUnsupported(reason, where(c, member));
        return "panic(java.Unsupported(" + Translator.goQuote(reason) + "))";
    }

    // ---------------------------------------------------------------------
    // Struct and identity

    private void emitStruct(Model.ClassInfo c) {
        w("type " + c.goName + " struct {");
        if (c.superInfo != null) {
            w(c.superInfo.goName);
        } else if (c.rtSuper != null) {
            String q = Runtime.qualifiedName(c.rtSuper);
            if (q == null || !tr.rt.hasType(q)) {
                tr.reportUnsupported("runtime superclass " + c.rtSuper, where(c, "<class>"));
            } else {
                w(q);
            }
        }
        if (c.superInfo == null && c.polymorphic()) {
            w("jkind_ int32");
        }
        if (needsIdentityHash(c)) {
            w("jhash_ int32");
        }
        if (c.superInfo == null && c.hasMonitor) {
            w("jmon_ sync.Mutex");
        }
        if (c.hasOuter) {
            w("outer_ *" + model.info(c.sym.owner.enclClass()).goName);
        }
        for (VarSymbol v : c.captured) {
            w("cap_" + v.name + " " + tr.goType(v.type));
        }
        for (JCTree t : c.decl.defs) {
            if (t instanceof JCVariableDecl v && !v.sym.isStatic()) {
                String vt = tr.volatileType(v.sym);
                w(model.fieldName(v.sym) + " " + (vt != null ? vt : tr.goType(v.sym.type)));
            }
        }
        w("}");
    }

    private boolean needsIdentityHash(Model.ClassInfo c) {
        if (c.superInfo != null || c.rtSuper != null) {
            return false;
        }
        for (Symbol s : c.sym.members().getSymbolsByName(model.names.hashCode, LookupKind.NON_RECURSIVE)) {
            if (s instanceof MethodSymbol m && m.params.isEmpty()) {
                return false;
            }
        }
        return true;
    }

    private void emitIdentity(Model.ClassInfo c) {
        String g = c.goName;
        w("func (this *" + g + ") JClass_" + g + "() *" + g + " { return this }");
        w("func (this *" + g + ") JClassName() string { return " + Translator.goQuote(c.sym.flatname.toString())
                + " }");
        if (needsIdentityHash(c)) {
            w("func (this *" + g + ") JIdentityHash() *int32 { return &this.jhash_ }");
        }
        if (c.superInfo == null && c.hasMonitor) {
            w("func (this *" + g + ") JMonitor() *sync.Mutex { return &this.jmon_ }");
        }
        if (implementsComparableSelf(c)) {
            w("func (this *" + g + ") JComparableSelf() string { return "
                    + Translator.goQuote(c.sym.flatname.toString()) + " }");
        }
        if (!c.isAbstract()) {
            boolean hashField = needsIdentityHash(c.root) && c.root.rtSuper == null && c.root.superInfo == null;
            // A clone has a new identity: a fresh identity hash and monitor.
            w("func (this *" + g + ") JShallowCopy() any { c := new(" + g + "); *c = *this"
                    + (hashField ? "; c.jhash_ = 0" : "") + (c.root.hasMonitor ? "; c.jmon_ = sync.Mutex{}" : "")
                    + "; return c }");
        }
        String iface = "interface{ JClass_" + g + "() *" + g + " }";
        if (c.polymorphic()) {
            w("func " + g + "_fromAny(o any) *" + g + " { if o == nil { return nil }; return o.(" + iface
                    + ").JClass_" + g + "() }");
        } else {
            w("func " + g + "_fromAny(o any) *" + g + " { if o == nil { return nil }; return o.(*" + g + ") }");
        }
        w("func " + g + "_isInstance(o any) bool { _, ok := o.(" + iface + "); return ok }");
        if (!c.polymorphic()) {
            return;
        }
        String root = c.root.goName;
        w("func " + g + "_isKind(p *" + root + ") bool { return p != nil && p.jkind_ >= " + c.kindLo
                + " && p.jkind_ <= " + c.kindHi + " }");
        w("func " + g + "_downcast(p *" + root + ") *" + g + " {");
        w("if p == nil { return nil }");
        w("if p.jkind_ < " + c.kindLo + " || p.jkind_ > " + c.kindHi + " { panic(java.ClassCastError("
                + root + "_toAny(p), " + Translator.goQuote(c.sym.flatname.toString()) + ")) }");
        w("return (*" + g + ")(unsafe.Pointer(p))");
        w("}");
        if (c.superInfo == null) {
            w("func (this *" + g + ") JToAny() any { return " + g + "_toAny(this) }");
            w("func " + g + "_toAny(p *" + g + ") any {");
            w("if p == nil { return nil }");
            w("switch p.jkind_ {");
            for (Model.ClassInfo k : model.concreteSubtree(c)) {
                w("case " + k.kindLo + ": return (*" + k.goName + ")(unsafe.Pointer(p))");
            }
            w("}");
            w("panic(\"jts: bad kind\")");
            w("}");
        }
    }

    /**
     * Whether c directly implements Comparable&lt;c&gt;, which java.util.HashMap
     * checks (comparableClassFor) when ordering keys in treeified bins.
     */
    private boolean implementsComparableSelf(Model.ClassInfo c) {
        for (Type it : model.types.interfaces(c.sym.type)) {
            if (it.tsym == model.syms.comparableType.tsym && it.getTypeArguments().size() == 1
                    && it.getTypeArguments().head.tsym == c.sym) {
                return true;
            }
        }
        return false;
    }

    // ---------------------------------------------------------------------
    // Statics

    private void emitStatics(Model.ClassInfo c) {
        List<String> enumConsts = new ArrayList<>();
        for (JCTree t : c.decl.defs) {
            if (t instanceof JCVariableDecl v && v.sym.isStatic()) {
                String name = model.staticFieldName(v.sym);
                String type = tr.goType(v.sym.type);
                Object cv = v.sym.getConstValue();
                if ((v.sym.flags() & Flags.ENUM) != 0) {
                    try {
                        Body b = new Body(tr, c, model.syms.voidType);
                        String init = b.exprTo(v.init, v.sym.type);
                        tr.requireFunc("java.EnumConst");
                        w("var " + name + " " + type + " = java.EnumConst(" + init + ", "
                                + Translator.goQuote(v.name.toString()) + ", " + enumConsts.size() + ")");
                    } catch (Unsupported u) {
                        tr.reportUnsupported(u.getMessage(), where(c, v.name.toString()));
                        w("var " + name + " " + type + " // unsupported enum constant: " + u.getMessage());
                    }
                    enumConsts.add(name);
                    continue;
                }
                if (cv != null && Translator.goConstable(v.sym.type, cv)) {
                    w("const " + name + " " + type + " = " + tr.constant(v.sym.type, cv));
                } else if (cv != null) {
                    w("var " + name + " " + type + " = " + tr.constant(v.sym.type, cv));
                } else if (v.init == null) {
                    w("var " + name + " " + type);
                } else {
                    try {
                        Body b = new Body(tr, c, model.syms.voidType);
                        String init = b.exprTo(v.init, v.sym.type);
                        if (b.out.length() > 0) {
                            throw new Unsupported("static initializer needs statements");
                        }
                        if (canFail(v.init)) {
                            tr.requireFunc("java.StaticInit");
                            w("var " + name + " " + type + " = java.StaticInit(func() " + type + " { return " + init
                                    + " })");
                        } else {
                            w("var " + name + " " + type + " = " + init);
                        }
                    } catch (Unsupported u) {
                        tr.reportUnsupported(u.getMessage(), where(c, v.name.toString()));
                        w("var " + name + " " + type + " // unsupported initializer: " + u.getMessage());
                    }
                }
            }
            if (t instanceof JCBlock blk && (blk.flags & Flags.STATIC) != 0) {
                Body b = new Body(tr, c, model.syms.voidType);
                b.scanReads(blk);
                try {
                    b.stmts(blk.stats);
                    tr.requireFunc("java.StaticInit");
                    w("func init() {");
                    w("java.StaticInit(func() int {");
                    out.append(b.out);
                    w("return 0");
                    w("})");
                    w("}");
                } catch (Unsupported u) {
                    tr.reportUnsupported(u.getMessage(), where(c, "<clinit>"));
                    w("// unsupported static initializer: " + u.getMessage());
                }
            }
        }
    }

    private void emitEnumMethods(Model.ClassInfo c) {
        List<String> consts = new ArrayList<>();
        for (JCTree t : c.decl.defs) {
            if (t instanceof JCVariableDecl v && (v.sym.flags() & Flags.ENUM) != 0) {
                consts.add(model.staticFieldName(v.sym));
            }
        }
        String g = c.goName;
        w("func " + g + "_values() []*" + g + " { return []*" + g + "{" + String.join(", ", consts) + "} }");
        w("func " + g + "_valueOf(name *java.String) *" + g + " {");
        w("for _, v := range " + g + "_values() { if v.Name().Equals_Object(name) { return v } }");
        w("panic(java.IllegalArgumentException_new_String(java.Concat(java.Lit(\"No enum constant \"), name)))");
        w("}");
    }

    /** Whether evaluating e calls code (and so can fail). */
    private static boolean canFail(JCExpression e) {
        boolean[] calls = {false};
        new com.sun.tools.javac.tree.TreeScanner() {
            @Override
            public void visitApply(JCMethodInvocation t) {
                calls[0] = true;
            }

            @Override
            public void visitNewClass(JCNewClass t) {
                calls[0] = true;
            }
        }.scan(e);
        return calls[0];
    }

    private void emitStaticMethods(Model.ClassInfo c) {
        for (JCTree t : c.decl.defs) {
            if (t instanceof JCMethodDecl md && md.sym.isStatic()) {
                if (c.test && md.name.contentEquals("main")) {
                    continue;
                }
                Type sig = model.types.erasure(md.sym.type);
                emitFunc("func " + model.staticMethodName(md.sym), c, md, sig, null);
            }
        }
    }

    // ---------------------------------------------------------------------
    // Constructors

    private void emitCtors(Model.ClassInfo c) {
        for (JCTree t : c.decl.defs) {
            if (!(t instanceof JCMethodDecl md) || !md.sym.isConstructor()) {
                continue;
            }
            MethodSymbol ctor = md.sym;
            Type sig = model.types.erasure(ctor.type);
            List<String> pnames = new ArrayList<>();
            List<String> params = new ArrayList<>();
            for (int i = 0; i < md.params.size(); i++) {
                String n = Model.goIdent(md.params.get(i).name.toString());
                pnames.add(n);
                params.add(n + " " + tr.goType(sig.getParameterTypes().get(i)));
            }
            String init = model.ctorInitName(ctor);
            if (!c.isAbstract()) {
                List<String> newParams = new ArrayList<>();
                if (c.hasOuter) {
                    newParams.add("outer_ *" + model.info(c.sym.owner.enclClass()).goName);
                }
                for (VarSymbol v : c.captured) {
                    newParams.add("cap_" + v.name + " " + tr.goType(v.type));
                }
                newParams.addAll(params);
                w("func " + model.ctorNewName(ctor) + "(" + String.join(", ", newParams) + ") *" + c.goName + " {");
                w("this := &" + c.goName + "{}");
                if (c.polymorphic()) {
                    w("this.jkind_ = " + c.kindLo);
                }
                if (c.root.rtSuper != null && tr.rt.hasMethod(Runtime.qualifiedName(c.root.rtSuper), "JSetSelf")) {
                    w("this.JSetSelf(this)");
                }
                if (c.hasOuter) {
                    w("this.outer_ = outer_");
                }
                for (VarSymbol v : c.captured) {
                    w("this.cap_" + v.name + " = cap_" + v.name);
                }
                w("this." + init + "(" + String.join(", ", pnames) + ")");
                w("return this");
                w("}");
            }
            String body;
            try {
                Body b = new Body(tr, c, model.syms.voidType);
                b.scanReads(md.body);
                for (JCTree t2 : c.decl.defs) {
                    if (t2 instanceof JCVariableDecl fv && !fv.sym.isStatic()) {
                        b.scanReads(fv.init);
                    }
                    if (t2 instanceof JCBlock blk2) {
                        b.scanReads(blk2);
                    }
                }
                for (JCVariableDecl p : md.params) {
                    b.declareLocal(p.sym);
                }
                List<JCStatement> stats = md.body.stats;
                JCStatement first = stats.isEmpty() ? null : stats.get(0);
                boolean callsThis = false;
                if (first instanceof JCExpressionStatement es && es.expr instanceof JCMethodInvocation mi
                        && TreeInfo.isSelfCall(mi)) {
                    callsThis = TreeInfo.name(mi.meth) == model.names._this;
                    String call = b.ctorCall(mi, (MethodSymbol) TreeInfo.symbol(mi.meth));
                    if (!call.isEmpty()) {
                        b.w(call);
                    }
                    stats = stats.subList(1, stats.size());
                }
                if (!callsThis) {
                    fieldInits(c, b);
                }
                b.stmts(stats);
                body = b.out.toString();
            } catch (RuntimeException u) {
                body = stub(u, c, "<init>" + model.sigSuffix(ctor));
            }
            w("func (this *" + c.goName + ") " + init + "(" + String.join(", ", params) + ") {");
            out.append(body);
            w("}");
        }
    }

    private void fieldInits(Model.ClassInfo c, Body b) {
        for (JCTree t : c.decl.defs) {
            if (t instanceof JCVariableDecl v && !v.sym.isStatic() && v.init != null) {
                String value = b.exprTo(v.init, v.sym.type);
                if (tr.volatileType(v.sym) != null) {
                    b.w("this." + model.fieldName(v.sym) + ".Store(" + value + ")");
                } else {
                    b.w("this." + model.fieldName(v.sym) + " = " + value);
                }
            }
            if (t instanceof JCBlock blk && (blk.flags & Flags.STATIC) == 0) {
                b.stmt(blk);
            }
        }
    }

    // ---------------------------------------------------------------------
    // Methods

    private void emitInstanceMethods(Model.ClassInfo c) {
        for (JCTree t : c.decl.defs) {
            if (!(t instanceof JCMethodDecl md) || md.sym.isStatic() || md.sym.isConstructor()
                    || (md.sym.flags() & Flags.ABSTRACT) != 0) {
                continue;
            }
            String n = model.instanceMethodName(md.sym);
            if (!tr.isPrivate(md.sym)) {
                Model.ClassInfo intro = model.introducer(c, n);
                if (intro != null && intro.dispatchers.containsKey(n)) {
                    n = n + "_impl";
                }
            }
            emitFunc("func (this *" + c.goName + ") " + n, c, md, tr.goSignature(md.sym), null);
        }
    }

    /** Emits a function whose Go signature is sig, translating md's body. */
    private void emitFunc(String header, Model.ClassInfo c, JCMethodDecl md, Type sig, String receiverIface) {
        Type ret = sig.getReturnType();
        Type own = model.types.erasure(md.sym.type);
        List<String> params = new ArrayList<>();
        List<String> converts = new ArrayList<>();
        Body b = new Body(tr, c, ret);
        String body;
        try {
            for (int i = 0; i < md.params.size(); i++) {
                VarSymbol p = md.params.get(i).sym;
                Type goT = sig.getParameterTypes().get(i);
                Type ownT = own.getParameterTypes().get(i);
                String n = b.declareLocal(p);
                if (tr.goType(goT).equals(tr.goType(ownT))) {
                    params.add(n + " " + tr.goType(goT));
                } else {
                    params.add("jp_" + n + " " + tr.goType(goT));
                    converts.add("var " + n + " " + tr.goType(ownT) + " = " + tr.conv("jp_" + n, goT, ownT));
                    converts.add("_ = " + n);
                }
            }
            for (String s : converts) {
                b.w(s);
            }
            b.scanReads(md.body);
            if ((md.sym.flags() & Flags.SYNCHRONIZED) != 0) {
                if (md.sym.isStatic()) {
                    throw new Unsupported("static synchronized method");
                }
                b.w("this.jmon_.Lock()");
                b.w("defer this.jmon_.Unlock()");
            }
            b.prologue(md.body);
            b.stmts(md.body.stats);
            if (!ret.hasTag(TypeTag.VOID) && !endsInJump(md.body.stats)) {
                b.w("panic(\"unreachable\")");
            }
            body = b.out.toString();
        } catch (RuntimeException u) {
            params.clear();
            for (int i = 0; i < md.params.size(); i++) {
                params.add(Model.goIdent(md.params.get(i).name.toString()) + " "
                        + tr.goType(sig.getParameterTypes().get(i)));
            }
            body = stub(u, c, md.name + model.sigSuffix(md.sym)) + "\n";
        }
        String r = ret.hasTag(TypeTag.VOID) ? "" : " " + tr.goType(ret);
        String recvParam = receiverIface != null ? "this " + receiverIface + (params.isEmpty() ? "" : ", ") : "";
        w(header + "(" + recvParam + String.join(", ", params) + ")" + r + " {");
        out.append(body);
        w("}");
    }

    private static boolean endsInJump(List<JCStatement> stats) {
        if (stats.isEmpty()) {
            return false;
        }
        JCStatement last = stats.get(stats.size() - 1);
        return last instanceof JCReturn || last instanceof JCThrow;
    }

    // ---------------------------------------------------------------------
    // Dispatch

    private record Sig(List<String> params, List<String> names, String ret, boolean isVoid) {}

    private Sig sigOf(MethodSymbol m) {
        Type t = tr.goSignature(m);
        List<String> params = new ArrayList<>();
        List<String> names = new ArrayList<>();
        int i = 0;
        for (Type p : t.getParameterTypes()) {
            names.add("a" + i);
            params.add("a" + i + " " + tr.goType(p));
            i++;
        }
        boolean isVoid = t.getReturnType().hasTag(TypeTag.VOID);
        return new Sig(params, names, isVoid ? "" : " " + tr.goType(t.getReturnType()), isVoid);
    }

    private void emitDispatchers(Model.ClassInfo c) {
        for (Map.Entry<String, MethodSymbol> e : c.dispatchers.entrySet()) {
            String n = e.getKey();
            Sig s = sigOf(e.getValue());
            String args = String.join(", ", s.names);
            String ret = s.isVoid ? "" : "return ";
            w("func (this *" + c.goName + ") " + n + "(" + String.join(", ", s.params) + ")" + s.ret + " {");
            if (c.polymorphic()) {
                Map<Model.ClassInfo, List<Integer>> groups = new LinkedHashMap<>();
                for (Model.ClassInfo k : model.concreteSubtree(c)) {
                    Model.ClassInfo impl = model.implClass(k, n, c);
                    if (impl != null && impl != c) {
                        groups.computeIfAbsent(impl, x -> new ArrayList<>()).add(k.kindLo);
                    }
                }
                if (!groups.isEmpty()) {
                    w("switch this.jkind_ {");
                    for (Map.Entry<Model.ClassInfo, List<Integer>> g : groups.entrySet()) {
                        StringJoiner kinds = new StringJoiner(", ");
                        for (int k : g.getValue()) {
                            kinds.add(Integer.toString(k));
                        }
                        w("case " + kinds + ":");
                        w(ret + "(*" + g.getKey().goName + ")(unsafe.Pointer(this))." + n + "_impl(" + args + ")");
                        if (s.isVoid) {
                            w("return");
                        }
                    }
                    w("}");
                }
            }
            if (c.declared.containsKey(n) || c.forwarders.containsKey(n)) {
                w(ret + "this." + n + "_impl(" + args + ")");
            } else {
                w("panic(java.AbstractMethodError(" + Translator.goQuote(where(c, n)) + "))");
            }
            w("}");
        }
    }

    private void emitForwarders(Model.ClassInfo c) {
        for (Map.Entry<String, MethodSymbol> e : c.forwarders.entrySet()) {
            String n = e.getKey();
            MethodSymbol fm = e.getValue();
            Sig s = sigOf(fm);
            String name = c.dispatchers.containsKey(n) ? n + "_impl" : n;
            String args = String.join(", ", s.names);
            String ret = s.isVoid ? "" : "return ";
            String body;
            try {
                if (model.isJts(fm.owner)) {
                    Model.ClassInfo iface = model.info(fm.owner);
                    String self = tr.conv("this", c.sym.type, fm.owner.type);
                    body = ret + iface.goName + "_default_" + n + "(" + self + (args.isEmpty() ? "" : ", " + args) + ")";
                } else if (fm.owner == model.syms.objectType.tsym) {
                    String f = tr.requireFunc("java.ObjectImpl_" + n);
                    body = ret + f + "(" + tr.toAny("this", c.sym.type) + (args.isEmpty() ? "" : ", " + args) + ")";
                } else if (fm.owner.isInterface()) {
                    String f = tr.requireFunc("java." + fm.owner.name + "_default_" + n);
                    body = ret + f + "(" + tr.toAny("this", c.sym.type) + (args.isEmpty() ? "" : ", " + args) + ")";
                } else {
                    String q = Runtime.qualifiedName((ClassSymbol) fm.owner);
                    tr.requireMethod(q, n);
                    String field = Runtime.qualifiedName(c.root.rtSuper);
                    body = ret + "this." + field.substring(field.indexOf('.') + 1) + "." + n + "(" + args + ")";
                }
            } catch (Unsupported u) {
                body = stub(u, c, n);
            }
            w("func (this *" + c.goName + ") " + name + "(" + String.join(", ", s.params) + ")" + s.ret + " {");
            w(body);
            w("}");
        }
    }

    private void emitMarkers(Model.ClassInfo c) {
        for (ClassSymbol i : c.markers) {
            w("func (this *" + c.goName + ") JIface_" + model.info(i).goName + "() {}");
        }
    }

    // ---------------------------------------------------------------------
    // Interfaces

    private void emitInterface(Model.ClassInfo c) {
        w("type " + c.goName + " interface {");
        w("JIface_" + c.goName + "()");
        for (Type st : model.types.interfaces(c.sym.type)) {
            if (model.isJts(st.tsym)) {
                w(model.info(st.tsym).goName);
            }
        }
        java.util.Set<String> seen = new java.util.HashSet<>();
        for (Type st : model.types.closure(c.sym.type)) {
            boolean jts = model.isJts(st.tsym);
            if (jts && st.tsym != c.sym) {
                continue;
            }
            for (Symbol s : st.tsym.members().getSymbols(LookupKind.NON_RECURSIVE)) {
                if (!(s instanceof MethodSymbol m) || m.isStatic() || model.isObjectMethod(m)) {
                    continue;
                }
                if (!jts && (m.flags() & Flags.DEFAULT) != 0) {
                    continue;
                }
                String n = model.instanceMethodName(m);
                if (seen.add(n)) {
                    Sig sig = sigOf(m);
                    w(n + "(" + String.join(", ", sig.params) + ")" + sig.ret);
                }
            }
        }
        w("}");
    }

    private void emitDefaultMethods(Model.ClassInfo c) {
        for (JCTree t : c.decl.defs) {
            if (t instanceof JCMethodDecl md && (md.sym.flags() & Flags.DEFAULT) != 0) {
                String n = model.instanceMethodName(md.sym);
                emitFunc("func " + c.goName + "_default_" + n, c, md, tr.goSignature(md.sym), c.goName);
            }
        }
    }

    // ---------------------------------------------------------------------
    // JUnit tests

    private void emitTests(Model.ClassInfo c) {
        if (c.isAbstract() || c.sym.isAnonymous() || c.sym.isDirectlyOrIndirectlyLocal()) {
            return;
        }
        ClassSymbol testCase = findClass("junit.framework.TestCase");
        if (testCase == null || !c.sym.isSubClass(testCase, model.types)) {
            emitJUnit4Tests(c);
            return;
        }
        MethodSymbol stringCtor = null;
        MethodSymbol noArgCtor = null;
        for (Symbol s : c.sym.members().getSymbolsByName(model.names.init, LookupKind.NON_RECURSIVE)) {
            MethodSymbol m = (MethodSymbol) s;
            if (m.params.isEmpty()) {
                noArgCtor = m;
            } else if (m.params.size() == 1 && tr.isString(m.params.head.type)) {
                stringCtor = m;
            }
        }
        List<MethodSymbol> tests = new ArrayList<>();
        java.util.Set<String> names = new java.util.HashSet<>();
        for (ClassSymbol k = c.sym; k != null && model.isJts(k); k = (ClassSymbol) k.getSuperclass().tsym) {
            for (Symbol s : k.members().getSymbols(LookupKind.NON_RECURSIVE)) {
                if (s instanceof MethodSymbol m && m.name.toString().startsWith("test") && m.params.isEmpty()
                        && (m.flags() & Flags.PUBLIC) != 0 && !m.isStatic()
                        && m.getReturnType().hasTag(TypeTag.VOID) && names.add(m.name.toString())) {
                    tests.add(m);
                }
            }
        }
        tests.sort(java.util.Comparator.comparing(m -> m.name.toString()));
        for (MethodSymbol m : tests) {
            String fn = "Test" + c.goName + "_" + m.name;
            try {
                Body b = new Body(tr, c, model.syms.voidType);
                String create;
                if (stringCtor != null) {
                    create = model.ctorNewName(stringCtor) + "(" + tr.literal(m.name.toString()) + ")";
                } else if (noArgCtor != null) {
                    create = model.ctorNewName(noArgCtor) + "()";
                } else {
                    throw new Unsupported("test class without usable constructor");
                }
                String setUp = b.invoke("tc", c.sym.type, findMethod(c.sym, "setUp"), List.of());
                String tearDown = b.invoke("tc", c.sym.type, findMethod(c.sym, "tearDown"), List.of());
                String call = b.invoke("tc", c.sym.type, m, List.of());
                tr.requireFunc("junit.Run");
                w("func " + fn + "(t *testing.T) {");
                w("junit.Run(t, func() {");
                w("tc := " + create);
                w(setUp);
                w("defer " + tearDown);
                w(call);
                w("})");
                w("}");
            } catch (Unsupported u) {
                tr.reportUnsupported(u.getMessage(), where(c, m.name + " (test)"));
                w("func " + fn + "(t *testing.T) { t.Skip(" + Translator.goQuote("unsupported: " + u.getMessage())
                        + ") }");
            }
        }
    }

    /** Emits Go tests for a JUnit 4 class: its @Test methods. */
    private void emitJUnit4Tests(Model.ClassInfo c) {
        MethodSymbol ctor = null;
        for (Symbol s : c.sym.members().getSymbolsByName(model.names.init, LookupKind.NON_RECURSIVE)) {
            if (s instanceof MethodSymbol m && m.params.isEmpty()) {
                ctor = m;
            }
        }
        for (Symbol s : c.sym.members().getSymbols(LookupKind.NON_RECURSIVE)) {
            if (!(s instanceof MethodSymbol m) || m.isStatic() || !hasAnnotation(m, "org.junit.Test")) {
                continue;
            }
            String fn = "Test" + c.goName + "_" + m.name;
            try {
                if (ctor == null) {
                    throw new Unsupported("JUnit 4 test class without a no-argument constructor");
                }
                Body b = new Body(tr, c, model.syms.voidType);
                String call = b.invoke("tc", c.sym.type, m, List.of());
                tr.requireFunc("junit.Run");
                w("func " + fn + "(t *testing.T) {");
                w("junit.Run(t, func() {");
                w("tc := " + model.ctorNewName(ctor) + "()");
                w(call);
                w("})");
                w("}");
            } catch (Unsupported u) {
                tr.reportUnsupported(u.getMessage(), where(c, m.name + " (test)"));
                w("func " + fn + "(t *testing.T) { t.Skip(" + Translator.goQuote("unsupported: " + u.getMessage())
                        + ") }");
            }
        }
    }

    private static boolean hasAnnotation(Symbol s, String name) {
        for (var a : s.getAnnotationMirrors()) {
            if (a.getAnnotationType().toString().equals(name)) {
                return true;
            }
        }
        return false;
    }

    private ClassSymbol findClass(String name) {
        for (Symbol s : model.syms.getAllClasses()) {
            if (s instanceof ClassSymbol cs && cs.getQualifiedName().contentEquals(name)) {
                return cs;
            }
        }
        return null;
    }

    private MethodSymbol findMethod(ClassSymbol c, String name) {
        for (Type t : model.types.closure(c.type)) {
            for (Symbol s : t.tsym.members().getSymbolsByName(model.names.fromString(name),
                    LookupKind.NON_RECURSIVE)) {
                if (s instanceof MethodSymbol m && m.params.isEmpty()) {
                    return m;
                }
            }
        }
        throw new Unsupported("no method " + name);
    }
}
