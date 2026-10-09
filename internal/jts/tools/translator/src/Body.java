import com.sun.tools.javac.code.Kinds.Kind;
import com.sun.tools.javac.code.Symbol;
import com.sun.tools.javac.code.Symbol.ClassSymbol;
import com.sun.tools.javac.code.Symbol.MethodSymbol;
import com.sun.tools.javac.code.Symbol.OperatorSymbol;
import com.sun.tools.javac.code.Symbol.VarSymbol;
import com.sun.tools.javac.code.Type;
import com.sun.tools.javac.code.TypeTag;
import com.sun.tools.javac.tree.JCTree;
import com.sun.tools.javac.tree.JCTree.*;
import com.sun.tools.javac.tree.TreeInfo;
import com.sun.tools.javac.tree.TreeScanner;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.StringJoiner;

/** Translates Java statements and expressions of one Go function body. */
final class Body {
    final Translator tr;
    final Model model;
    final Model.ClassInfo cls;
    final StringBuilder out = new StringBuilder();
    final Map<VarSymbol, String> locals;
    private final Set<String> usedNames = new HashSet<>();
    Type retType;
    private int closureDepth;
    /** Whether the innermost closure is a catch clause, which returns (int, bool). */
    private boolean inCatch;
    private int tmpCounter;

    private record Breakable(boolean loop, int closureDepth) {}

    private final Deque<Breakable> breakables = new ArrayDeque<>();
    private final Deque<Set<Integer>> tryFlows = new ArrayDeque<>();

    Body(Translator tr, Model.ClassInfo cls, Type retType) {
        this(tr, cls, retType, new HashMap<>(), new HashSet<>());
    }

    private Body(Translator tr, Model.ClassInfo cls, Type retType, Map<VarSymbol, String> locals,
            Set<VarSymbol> read) {
        this.tr = tr;
        this.model = tr.model;
        this.cls = cls;
        this.retType = retType;
        this.locals = locals;
        this.read = read;
    }

    /** A body for a lambda inside this one, sharing this body's locals. */
    private Body child(Type ret) {
        return new Body(tr, cls, ret, locals, read);
    }

    void w(String s) {
        out.append(s).append('\n');
    }

    /** Local variables that are read somewhere in the function. */
    private final Set<VarSymbol> read;

    /** Records which locals the tree reads, so unread ones get "_ = x". */
    void scanReads(JCTree tree) {
        if (tree == null) {
            return;
        }
        new TreeScanner() {
            @Override
            public void visitIdent(JCIdent id) {
                // Constant reads are emitted as the constant's value.
                if (id.sym instanceof VarSymbol v && v.owner.kind == Kind.MTH
                        && (id.type == null || id.type.constValue() == null)) {
                    read.add(v);
                }
            }

            @Override
            public void visitAssert(JCAssert a) {
                // Assertions are not translated.
            }

            @Override
            public void visitAssign(JCAssign a) {
                if (!(TreeInfo.skipParens(a.lhs) instanceof JCIdent)) {
                    scan(a.lhs);
                }
                scan(a.rhs);
            }
        }.scan(tree);
    }

    /** Marks a declared local as used if nothing reads it. */
    void markUsed(String goName, VarSymbol v) {
        if (!read.contains(v)) {
            w("_ = " + goName);
        }
    }

    String declareLocal(VarSymbol v) {
        String base = Model.goIdent(v.name.toString());
        String n = base;
        // Java forbids shadowing locals within a method, but sibling scopes
        // may reuse a name; Go block scoping handles that, so reuse is fine.
        usedNames.add(n);
        locals.put(v, n);
        return n;
    }

    private String tmp() {
        return "jt" + (tmpCounter++);
    }

    /** Declares jret when a try block may return a value. */
    void prologue(JCBlock body) {
        if (retType.hasTag(TypeTag.VOID) || body == null) {
            return;
        }
        boolean[] hasTry = {false};
        new TreeScanner() {
            @Override
            public void visitTry(JCTry t) {
                hasTry[0] = true;
            }

            @Override
            public void visitClassDef(JCClassDecl t) {}
        }.scan(body);
        if (hasTry[0]) {
            w("var jret " + tr.goType(retType));
            w("_ = jret");
        }
    }

    // ---------------------------------------------------------------------
    // Statements

    void stmts(List<JCStatement> ss) {
        for (JCStatement s : ss) {
            stmt(s);
        }
    }

    void stmt(JCStatement s) {
        switch (s) {
            case JCBlock b -> {
                if ((b.flags & com.sun.tools.javac.code.Flags.STATIC) != 0) {
                    throw new Unsupported("static block in body");
                }
                w("{");
                stmts(b.stats);
                w("}");
            }
            case JCVariableDecl v -> localVar(v, false);
            case JCExpressionStatement es -> exprStmt(es.expr);
            case JCIf i -> {
                w("if " + cond(i.cond) + " {");
                inner(i.thenpart);
                if (i.elsepart != null) {
                    w("} else {");
                    inner(i.elsepart);
                }
                w("}");
            }
            case JCWhileLoop wl -> {
                w("for " + cond(wl.cond) + " {");
                loopBody(wl.body);
                w("}");
            }
            case JCDoWhileLoop dw -> {
                w("for jfirst := true; jfirst || " + cond(dw.cond) + "; jfirst = false {");
                loopBody(dw.body);
                w("}");
            }
            case JCForLoop f -> forLoop(f);
            case JCEnhancedForLoop ef -> enhancedFor(ef);
            case JCSwitch sw -> switchStmt(sw);
            case JCReturn r -> returnStmt(r);
            case JCBreak b -> {
                if (b.label != null) {
                    throw new Unsupported("labeled break");
                }
                Breakable target = breakables.peek();
                if (target.closureDepth < closureDepth) {
                    tryFlows.peek().add(2);
                    w(flowReturn(2));
                } else {
                    w("break");
                }
            }
            case JCContinue c -> {
                if (c.label != null) {
                    throw new Unsupported("labeled continue");
                }
                Breakable target = innermostLoop();
                if (target.closureDepth < closureDepth) {
                    tryFlows.peek().add(3);
                    w(flowReturn(3));
                } else {
                    w("continue");
                }
            }
            case JCThrow t -> w("panic(" + exprTo(t.expr, model.syms.objectType) + ")");
            case JCTry t -> tryStmt(t);
            case JCSynchronized sy -> synchronizedStmt(sy);
            case JCClassDecl cd -> {
                // Local classes are emitted separately.
            }
            case JCAssert a -> {
                // Assertions are disabled by default in Java.
            }
            case JCSkip sk -> {}
            case JCLabeledStatement l -> throw new Unsupported("labeled statement");
            default -> throw new Unsupported("statement " + s.getClass().getSimpleName());
        }
    }

    private String flowReturn(int code) {
        return "return " + code + (inCatch ? ", true" : "");
    }

    private Breakable innermostLoop() {
        for (Breakable b : breakables) {
            if (b.loop) {
                return b;
            }
        }
        throw new Unsupported("continue outside loop");
    }

    private void loopBody(JCStatement body) {
        breakables.push(new Breakable(true, closureDepth));
        inner(body);
        breakables.pop();
    }

    /** Translates the body of a compound statement, which Go already braces. */
    private void inner(JCStatement s) {
        if (s instanceof JCBlock b && (b.flags & com.sun.tools.javac.code.Flags.STATIC) == 0) {
            stmts(b.stats);
        } else {
            stmt(s);
        }
    }

    void localVar(JCVariableDecl v, boolean assignOnly) {
        String n = assignOnly ? locals.get(v.sym) : declareLocal(v.sym);
        if (assignOnly) {
            if (v.init != null) {
                w(n + " = " + exprTo(v.init, v.sym.type));
            }
            return;
        }
        String t = tr.goType(v.sym.type);
        if (v.init != null) {
            w("var " + n + " " + t + " = " + exprTo(v.init, v.sym.type));
        } else {
            w("var " + n + " " + t);
        }
        markUsed(n, v.sym);
    }

    private void forLoop(JCForLoop f) {
        w("{");
        for (JCStatement s : f.init) {
            stmt(s);
        }
        String post;
        if (f.step.isEmpty()) {
            post = "";
        } else if (f.step.size() == 1) {
            post = simpleStmt(f.step.head.expr);
        } else {
            StringJoiner sj = new StringJoiner("; ", "func() { ", " }()");
            for (JCExpressionStatement es : f.step) {
                sj.add(simpleStmt(es.expr));
            }
            post = sj.toString();
        }
        String c = f.cond == null ? "" : cond(f.cond);
        w("for ; " + c + "; " + post + " {");
        loopBody(f.body);
        w("}");
        w("}");
    }

    private void enhancedFor(JCEnhancedForLoop ef) {
        Type it = model.types.erasure(ef.expr.type);
        VarSymbol v = ef.var.sym;
        if (it.hasTag(TypeTag.ARRAY)) {
            Type elem = ((Type.ArrayType) it).elemtype;
            w("for _, jv := range " + expr(ef.expr) + " {");
            String n = declareLocal(v);
            w("var " + n + " " + tr.goType(v.type) + " = " + tr.conv("jv", elem, v.type));
            markUsed(n, v);
        } else {
            String recv = expr(ef.expr);
            Translator.Repr r = tr.repr(it);
            String iter;
            switch (r.kind()) {
                case RTCLASS, RTIFACE -> {
                    tr.requireMethod(r.go().replace("*", ""), "Iterator");
                    iter = recv + ".Iterator()";
                }
                case JCLASS, JIFACE -> iter = recv + ".Iterator()";
                default -> throw new Unsupported("enhanced for over " + it);
            }
            tr.requireMethod("java.Iterator", "HasNext");
            tr.requireMethod("java.Iterator", "Next");
            String t = tmp();
            w("for " + t + " := " + iter + "; " + t + ".HasNext(); {");
            String n = declareLocal(v);
            w("var " + n + " " + tr.goType(v.type) + " = "
                    + tr.conv(t + ".Next()", model.syms.objectType, v.type));
            markUsed(n, v);
        }
        loopBody(ef.body);
        w("}");
    }

    private void switchStmt(JCSwitch sw) {
        Type selT = sw.selector.type;
        boolean isEnum = (selT.tsym.flags() & com.sun.tools.javac.code.Flags.ENUM) != 0;
        if (!selT.isPrimitive() && !isEnum) {
            throw new Unsupported("switch on " + selT);
        }
        w("{");
        Set<VarSymbol> hoisted = new HashSet<>();
        for (JCCase c : sw.cases) {
            for (JCStatement s : c.stats) {
                if (s instanceof JCVariableDecl v) {
                    hoisted.add(v.sym);
                    String n = declareLocal(v.sym);
                    w("var " + n + " " + tr.goType(v.sym.type));
                    markUsed(n, v.sym);
                }
            }
        }
        if (isEnum) {
            tr.requireMethod("java.Enum", "Ordinal");
            w("switch " + expr(sw.selector) + ".Ordinal() {");
        } else {
            w("switch " + expr(sw.selector) + " {");
        }
        breakables.push(new Breakable(false, closureDepth));
        for (int i = 0; i < sw.cases.size(); i++) {
            JCCase c = sw.cases.get(i);
            List<String> labels = new ArrayList<>();
            boolean isDefault = false;
            for (JCCaseLabel l : c.labels) {
                if (l instanceof JCDefaultCaseLabel) {
                    isDefault = true;
                } else if (l instanceof JCConstantCaseLabel cl && isEnum) {
                    labels.add("int32(" + enumOrdinal(selT, (VarSymbol) TreeInfo.symbol(cl.expr)) + ")");
                } else if (l instanceof JCConstantCaseLabel cl) {
                    labels.add(tr.constant(selT, cl.expr.type.constValue()));
                } else {
                    throw new Unsupported("case label " + l);
                }
            }
            if (isDefault && !labels.isEmpty()) {
                throw new Unsupported("default mixed with case labels");
            }
            w(isDefault ? "default:" : "case " + String.join(", ", labels) + ":");
            List<JCStatement> stats = new ArrayList<>(c.stats);
            boolean endsWithBreak = !stats.isEmpty() && stats.get(stats.size() - 1) instanceof JCBreak b
                    && b.label == null && closureDepth == breakables.peek().closureDepth;
            if (endsWithBreak) {
                stats.remove(stats.size() - 1);
            }
            for (JCStatement s : stats) {
                if (s instanceof JCVariableDecl v && hoisted.contains(v.sym)) {
                    localVar(v, true);
                } else {
                    stmt(s);
                }
            }
            boolean last = i == sw.cases.size() - 1;
            if (!last && !endsWithBreak && !completesAbruptly(stats)) {
                w("fallthrough");
            }
        }
        breakables.pop();
        w("}");
        w("}");
    }

    private int enumOrdinal(Type enumType, VarSymbol constant) {
        int i = 0;
        for (Symbol s : enumType.tsym.members().getSymbols(com.sun.tools.javac.code.Scope.LookupKind.NON_RECURSIVE)) {
            if (s instanceof VarSymbol v && (v.flags() & com.sun.tools.javac.code.Flags.ENUM) != 0) {
                i++;
            }
        }
        // Scope iteration is in reverse declaration order.
        int n = i;
        i = 0;
        for (Symbol s : enumType.tsym.members().getSymbols(com.sun.tools.javac.code.Scope.LookupKind.NON_RECURSIVE)) {
            if (s instanceof VarSymbol v && (v.flags() & com.sun.tools.javac.code.Flags.ENUM) != 0) {
                if (v == constant) {
                    return n - 1 - i;
                }
                i++;
            }
        }
        throw new Unsupported("enum constant not found");
    }

    private boolean completesAbruptly(List<JCStatement> stats) {
        if (stats.isEmpty()) {
            return false;
        }
        JCStatement s = stats.get(stats.size() - 1);
        return s instanceof JCReturn || s instanceof JCThrow || s instanceof JCContinue
                || s instanceof JCBreak;
    }

    private void returnStmt(JCReturn r) {
        if (r.expr == null) {
            if (closureDepth > 0) {
                tryFlows.peek().add(1);
                w(flowReturn(1));
            } else {
                w("return");
            }
            return;
        }
        String v = exprTo(r.expr, retType);
        if (closureDepth > 0) {
            tryFlows.peek().add(1);
            w("jret = " + v);
            w(flowReturn(1));
        } else {
            w("return " + v);
        }
    }

    private void tryStmt(JCTry t) {
        if (t.resources != null && !t.resources.isEmpty()) {
            throw new Unsupported("try-with-resources");
        }
        tryStmt(t.body, t.catchers, t.finalizer, null);
    }

    /** A synchronized block: lock, then try { body } finally { unlock }. */
    private void synchronizedStmt(JCSynchronized s) {
        String m = monitor(s.lock);
        w(m + ".Lock()");
        tryStmt(s.body, com.sun.tools.javac.util.List.nil(), null, m + ".Unlock()");
    }

    /** A Go expression for the mutex that is the monitor of Java object e. */
    private String monitor(JCExpression e) {
        Translator.Repr r = tr.repr(e.type);
        if (r.kind() == Translator.Kind.JCLASS && r.info().root.hasMonitor) {
            return "(&" + expr(e) + ".jmon_)";
        }
        return tr.requireFunc("java.MonitorOf") + "(" + exprTo(e, model.syms.objectType) + ")";
    }

    private void tryStmt(JCBlock tryBody, List<JCCatch> catchers, JCBlock finalizer, String rawFinally) {
        tr.requireFunc("java.Try");
        Set<Integer> flows = new HashSet<>();
        tryFlows.push(flows);
        closureDepth++;
        boolean savedInCatch = inCatch;
        inCatch = false;
        StringBuilder saved = new StringBuilder(out);
        out.setLength(0);
        w("func() int {");
        stmt(tryBody);
        w("return 0");
        w("}");
        String body = out.toString().trim();
        out.setLength(0);
        String catcher = "nil";
        if (!catchers.isEmpty()) {
            inCatch = true;
            w("func(jex any) (int, bool) {");
            for (JCCatch c : catchers) {
                List<Type> alts = new ArrayList<>();
                if (c.param.vartype instanceof JCTypeUnion u) {
                    for (JCExpression a : u.alternatives) {
                        alts.add(a.type);
                    }
                } else {
                    alts.add(c.param.sym.type);
                }
                StringJoiner checks = new StringJoiner(" || ");
                for (Type a : alts) {
                    checks.add(instanceOf("jex", model.syms.objectType, a));
                }
                w("if " + checks + " {");
                String n = declareLocal(c.param.sym);
                Type pt = alts.size() == 1 ? c.param.sym.type : model.syms.throwableType;
                w("var " + n + " " + tr.goType(pt) + " = " + tr.conv("jex", model.syms.objectType, pt));
                markUsed(n, c.param.sym);
                stmt(c.body);
                w("return 0, true");
                w("}");
            }
            w("return 0, false");
            w("}");
            catcher = out.toString().trim();
            out.setLength(0);
            inCatch = false;
        }
        String fin = "nil";
        if (finalizer != null) {
            w("func() int {");
            stmt(finalizer);
            w("return 0");
            w("}");
            fin = out.toString().trim();
            out.setLength(0);
        } else if (rawFinally != null) {
            fin = "func() int { " + rawFinally + "; return 0 }";
        }
        closureDepth--;
        inCatch = savedInCatch;
        tryFlows.pop();
        out.append(saved);
        if (flows.isEmpty()) {
            w("java.Try(" + body + ", " + catcher + ", " + fin + ")");
            return;
        }
        w("if jf := java.Try(" + body + ", " + catcher + ", " + fin + "); jf != 0 {");
        if (flows.contains(1)) {
            if (closureDepth > 0) {
                tryFlows.peek().add(1);
                w("if jf == 1 { " + flowReturn(1) + " }");
            } else if (retType.hasTag(TypeTag.VOID)) {
                w("if jf == 1 { return }");
            } else {
                w("if jf == 1 { return jret }");
            }
        }
        if (flows.contains(2)) {
            if (breakables.peek().closureDepth < closureDepth) {
                tryFlows.peek().add(2);
                w("if jf == 2 { " + flowReturn(2) + " }");
            } else {
                w("if jf == 2 { break }");
            }
        }
        if (flows.contains(3)) {
            if (innermostLoop().closureDepth < closureDepth) {
                tryFlows.peek().add(3);
                w("if jf == 3 { " + flowReturn(3) + " }");
            } else {
                w("if jf == 3 { continue }");
            }
        }
        w("}");
    }

    // ---------------------------------------------------------------------
    // Expression statements

    void exprStmt(JCExpression e) {
        w(simpleStmt(e));
    }

    /** A Go simple statement for a Java expression statement. */
    String simpleStmt(JCExpression e) {
        e = TreeInfo.skipParens(e);
        switch (e) {
            case JCAssign a: {
                String vt = volatileTarget(a.lhs);
                if (vt != null) {
                    return vt + ".Store(" + exprTo(a.rhs, a.lhs.type) + ")";
                }
                return lvalue(a.lhs) + " = " + exprTo(a.rhs, a.lhs.type);
            }
            case JCAssignOp a:
                return compoundStmt(a);
            case JCUnary u:
                switch (u.getTag()) {
                    case PREINC, POSTINC:
                        checkIncType(u.arg.type);
                        return lvalue(u.arg) + "++";
                    case PREDEC, POSTDEC:
                        checkIncType(u.arg.type);
                        return lvalue(u.arg) + "--";
                    default:
                        throw new Unsupported("unary statement");
                }
            case JCMethodInvocation mi: {
                MethodSymbol m = (MethodSymbol) TreeInfo.symbol(mi.meth);
                if (m.isConstructor()) {
                    return ctorCall(mi, m);
                }
                return callRaw(mi, m);
            }
            case JCNewClass nc:
                return "_ = " + newClass(nc);
            default:
                throw new Unsupported("expression statement " + e.getClass().getSimpleName());
        }
    }

    private void checkIncType(Type t) {
        if (!t.isPrimitive()) {
            throw new Unsupported("increment of boxed value");
        }
    }

    private String compoundStmt(JCAssignOp a) {
        if (isSimpleLvalue(a.lhs)) {
            String lv = lvalue(a.lhs);
            return lv + " = " + compoundValue(a, lv);
        }
        String lv = lvalue(a.lhs);
        String p = tmp();
        return "func() { " + p + " := &" + lv + "; *" + p + " = " + compoundValue(a, "(*" + p + ")") + " }()";
    }

    /** The value assigned by a compound assignment whose target reads as lv. */
    private String compoundValue(JCAssignOp a, String lv) {
        OperatorSymbol op = a.operator;
        Type res = op.type.getReturnType();
        JCTree.Tag tag = a.getTag().noAssignOp();
        if (tag == JCTree.Tag.PLUS && tr.isString(res)) {
            return "java.Concat(" + strOf(lv, a.lhs.type) + ", " + strOf(expr(a.rhs), a.rhs.type) + ")";
        }
        List<Type> pt = op.type.getParameterTypes();
        String l = tr.conv(lv, a.lhs.type, pt.get(0));
        String r = tr.conv(expr(a.rhs), a.rhs.type, pt.get(1));
        return tr.conv(binop(tag, l, r, pt.get(0), pt.get(1), res), res, a.lhs.type);
    }

    private boolean isSimpleLvalue(JCExpression e) {
        e = TreeInfo.skipParens(e);
        return switch (e) {
            case JCIdent id -> true;
            case JCFieldAccess fa -> isSimpleLvalue(fa.selected) || fa.sym != null && fa.sym.isStatic();
            case JCArrayAccess aa -> isSimpleLvalue(aa.indexed)
                    && (TreeInfo.skipParens(aa.index) instanceof JCIdent || aa.index.type.constValue() != null);
            default -> false;
        };
    }

    private void requireNonVolatile(VarSymbol v) {
        if (tr.volatileType(v) != null) {
            throw new Unsupported("non-assignment write to volatile field " + v.name);
        }
    }

    /** The Go location of a volatile field assigned by e, or null. */
    private String volatileTarget(JCExpression e) {
        e = TreeInfo.skipParens(e);
        if (e instanceof JCIdent id && id.sym instanceof VarSymbol v && v.owner.kind != Kind.MTH
                && !v.isStatic() && tr.volatileType(v) != null) {
            ImplicitThis it = implicitThis((ClassSymbol) v.owner);
            return fieldLoc(it.expr, it.info, v);
        }
        if (e instanceof JCFieldAccess fa && fa.sym instanceof VarSymbol v && !v.isStatic()
                && tr.volatileType(v) != null) {
            return fieldLoc(expr(fa.selected), classInfoOf(fa.selected.type), v);
        }
        return null;
    }

    String lvalue(JCExpression e) {
        e = TreeInfo.skipParens(e);
        switch (e) {
            case JCIdent id -> {
                if (id.sym instanceof VarSymbol v) {
                    if (v.owner.kind == Kind.MTH) {
                        String n = locals.get(v);
                        if (n == null) {
                            throw new Unsupported("assignment to captured variable");
                        }
                        return n;
                    }
                    if (v.isStatic()) {
                        return staticField(v);
                    }
                    ImplicitThis it = implicitThis((ClassSymbol) v.owner);
                    requireNonVolatile(v);
                    return fieldLoc(it.expr, it.info, v);
                }
                throw new Unsupported("lvalue ident");
            }
            case JCFieldAccess fa -> {
                VarSymbol v = (VarSymbol) fa.sym;
                if (v.isStatic()) {
                    return staticField(v);
                }
                requireNonVolatile(v);
                return fieldLoc(expr(fa.selected), classInfoOf(fa.selected.type), v);
            }
            case JCArrayAccess aa -> {
                return expr(aa.indexed) + "[" + exprTo(aa.index, model.syms.intType) + "]";
            }
            default -> throw new Unsupported("lvalue " + e.getClass().getSimpleName());
        }
    }

    // ---------------------------------------------------------------------
    // Expressions

    String cond(JCExpression e) {
        return exprTo(e, model.syms.booleanType);
    }

    String exprTo(JCExpression e, Type t) {
        return tr.conv(expr(e), e.type, t);
    }

    String expr(JCExpression e) {
        if (e.type != null) {
            Object cv = e.type.constValue();
            if (cv != null) {
                return tr.constant(e.type, cv);
            }
        }
        switch (e) {
            case JCParens p:
                return "(" + expr(p.expr) + ")";
            case JCLiteral l:
                if (l.typetag == TypeTag.BOT) {
                    return "nil";
                }
                return tr.constant(l.type, l.value);
            case JCIdent id:
                return ident(id);
            case JCFieldAccess fa:
                return fieldAccess(fa);
            case JCMethodInvocation mi: {
                MethodSymbol m = (MethodSymbol) TreeInfo.symbol(mi.meth);
                if (m.isConstructor()) {
                    throw new Unsupported("constructor call as expression");
                }
                String raw = callRaw(mi, m);
                if (mi.type.hasTag(TypeTag.VOID)) {
                    return raw;
                }
                return tr.conv(raw, tr.goSignature(m).getReturnType(), mi.type);
            }
            case JCNewClass nc:
                return newClass(nc);
            case JCNewArray na:
                return newArray(na);
            case JCAssign a:
                tr.requireFunc("java.Assign");
                return "java.Assign(&" + lvalue(a.lhs) + ", " + exprTo(a.rhs, a.lhs.type) + ")";
            case JCAssignOp a: {
                String lv = lvalue(a.lhs);
                String p = tmp();
                return "func() " + tr.goType(a.lhs.type) + " { " + p + " := &" + lv + "; *" + p + " = "
                        + compoundValue(a, "(*" + p + ")") + "; return *" + p + " }()";
            }
            case JCUnary u:
                return unary(u);
            case JCBinary b:
                return binary(b);
            case JCConditional c: {
                String t = tr.goType(c.type);
                return "func() " + t + " { if " + cond(c.cond) + " { return " + exprTo(c.truepart, c.type)
                        + " }; return " + exprTo(c.falsepart, c.type) + " }()";
            }
            case JCTypeCast tc: {
                // java.lang.reflect.Array.newInstance only serves to make an
                // array of a class's type, which Go slices do not record.
                if (TreeInfo.skipParens(tc.expr) instanceof JCMethodInvocation ni
                        && TreeInfo.symbol(ni.meth) instanceof MethodSymbol nm
                        && nm.owner.getQualifiedName().contentEquals("java.lang.reflect.Array")
                        && nm.name.contentEquals("newInstance") && ni.args.size() == 2
                        && tc.type.hasTag(TypeTag.ARRAY)) {
                    String gt = tr.goType(tc.type);
                    return "func() " + gt + " { _ = " + expr(ni.args.get(0)) + "; return make(" + gt + ", "
                            + exprTo(ni.args.get(1), model.syms.intType) + ") }()";
                }
                // A JDK method's Object[] result is a Go any holding an
                // array of its real element type, so convert it directly.
                if (TreeInfo.skipParens(tc.expr) instanceof JCMethodInvocation mi
                        && TreeInfo.symbol(mi.meth) instanceof MethodSymbol m && !m.isConstructor()
                        && !mi.type.hasTag(TypeTag.VOID)) {
                    Type goRet = tr.goSignature(m).getReturnType();
                    if (!model.types.isSameType(model.types.erasure(goRet), model.types.erasure(mi.type))
                            && goRet.tsym == model.syms.objectType.tsym) {
                        return tr.conv(callRaw(mi, m), goRet, tc.type);
                    }
                }
                return tr.conv(expr(tc.expr), tc.expr.type, tc.type);
            }
            case JCInstanceOf io:
                return instanceOf(expr(io.expr), io.expr.type, io.pattern.type);
            case JCArrayAccess aa:
                return expr(aa.indexed) + "[" + exprTo(aa.index, model.syms.intType) + "]";
            case JCLambda l:
                return lambda(l);
            case JCMemberReference r:
                return memberRef(r);
            default:
                throw new Unsupported("expression " + e.getClass().getSimpleName());
        }
    }

    private String ident(JCIdent id) {
        Symbol s = id.sym;
        if (s.name == model.names._this) {
            return "this";
        }
        if (s instanceof VarSymbol v) {
            if (v.owner.kind == Kind.MTH) {
                String n = locals.get(v);
                if (n != null) {
                    return n;
                }
                if (cls.captured.contains(v)) {
                    return "this.cap_" + v.name;
                }
                throw new Unsupported("unknown local " + v.name);
            }
            if (v.isStatic()) {
                return staticField(v);
            }
            ImplicitThis it = implicitThis((ClassSymbol) v.owner);
            return fieldRef(it.expr, it.info, v);
        }
        throw new Unsupported("identifier " + s.kind);
    }

    private String fieldAccess(JCFieldAccess fa) {
        if (fa.name == model.names._this) {
            return outerRef((ClassSymbol) fa.selected.type.tsym);
        }
        if (fa.name == model.names._class) {
            Type ct = model.types.erasure(fa.selected.type);
            String name;
            if (ct.isPrimitive()) {
                name = ct.toString();
            } else if (ct.hasTag(TypeTag.ARRAY)) {
                throw new Unsupported("array class literal");
            } else {
                name = ((ClassSymbol) ct.tsym).flatname.toString();
            }
            return tr.requireFunc("java.ClassOf") + "(" + Translator.goQuote(name) + ")";
        }
        if (fa.name == model.names.length && fa.selected.type.hasTag(TypeTag.ARRAY)) {
            return "int32(len(" + expr(fa.selected) + "))";
        }
        if (fa.sym instanceof VarSymbol v) {
            if (v.isStatic()) {
                return staticField(v);
            }
            return fieldRef(expr(fa.selected), classInfoOf(fa.selected.type), v);
        }
        throw new Unsupported("field access " + fa.sym);
    }

    private Model.ClassInfo classInfoOf(Type t) {
        Symbol s = model.types.erasure(t).tsym;
        if (!model.isJts(s)) {
            return null;
        }
        return model.info(s);
    }

    String staticField(VarSymbol v) {
        if (model.isJts(v.owner)) {
            tr.requireVisible(model.info(v.owner));
            return model.staticFieldName(v);
        }
        String q = Runtime.qualifiedName((ClassSymbol) v.owner);
        if (q == null) {
            throw new Unsupported("static field of " + v.owner);
        }
        return tr.requireFunc(q + "_" + v.name);
    }

    /** A read of field v. */
    private String fieldRef(String recv, Model.ClassInfo recvInfo, VarSymbol v) {
        String f = fieldLoc(recv, recvInfo, v);
        return tr.volatileType(v) != null ? f + ".Load()" : f;
    }

    /** The location of field v, for assignment. */
    private String fieldLoc(String recv, Model.ClassInfo recvInfo, VarSymbol v) {
        if (!model.isJts(v.owner)) {
            // Public fields of runtime classes are exported Go fields named F_<name>.
            String q = Runtime.qualifiedName((ClassSymbol) v.owner);
            if (q == null) {
                throw new Unsupported("field of " + v.owner);
            }
            tr.requireMethod(q, "F_" + v.name);
            return recv + ".F_" + v.name;
        }
        Model.ClassInfo owner = model.info(v.owner);
        String f = model.fieldName(v);
        if (owner != recvInfo && shadowed(recvInfo, owner, v)) {
            return tr.upcast(recv, owner) + "." + f;
        }
        return recv + "." + f;
    }

    private boolean shadowed(Model.ClassInfo from, Model.ClassInfo owner, VarSymbol v) {
        for (Model.ClassInfo c = from; c != null && c != owner; c = c.superInfo) {
            for (Symbol s : c.sym.members().getSymbolsByName(v.name)) {
                if (s.kind == Kind.VAR && !s.isStatic()) {
                    return true;
                }
            }
        }
        return false;
    }

    private record ImplicitThis(String expr, Model.ClassInfo info) {}

    private ImplicitThis implicitThis(ClassSymbol owner) {
        String e = "this";
        Model.ClassInfo c = cls;
        while (!c.sym.isSubClass(owner, model.types)) {
            if (!c.hasOuter) {
                throw new Unsupported("no enclosing instance for " + owner);
            }
            e = e + ".outer_";
            c = model.info(c.sym.owner.enclClass());
        }
        return new ImplicitThis(e, c);
    }

    private String outerRef(ClassSymbol target) {
        String e = "this";
        Model.ClassInfo c = cls;
        while (c.sym != target) {
            if (!c.hasOuter) {
                throw new Unsupported("no enclosing instance " + target);
            }
            e = e + ".outer_";
            c = model.info(c.sym.owner.enclClass());
        }
        return e;
    }

    // ---------------------------------------------------------------------
    // Calls

    List<String> args(MethodSymbol m, List<JCExpression> argTrees, Type varargsElement) {
        List<Type> params = tr.goSignature(m).getParameterTypes();
        List<String> out = new ArrayList<>();
        if (varargsElement != null) {
            int fixed = params.size() - 1;
            for (int i = 0; i < fixed; i++) {
                out.add(exprTo(argTrees.get(i), params.get(i)));
            }
            Type arr = params.get(fixed);
            Type elem = arr.hasTag(TypeTag.ARRAY) ? ((Type.ArrayType) arr).elemtype : model.syms.objectType;
            String arrGo = arr.hasTag(TypeTag.ARRAY) ? tr.goType(arr) : "[]any";
            StringJoiner sj = new StringJoiner(", ", arrGo + "{", "}");
            for (int i = fixed; i < argTrees.size(); i++) {
                sj.add(exprTo(argTrees.get(i), elem));
            }
            out.add(sj.toString());
            return out;
        }
        for (int i = 0; i < argTrees.size(); i++) {
            out.add(exprTo(argTrees.get(i), params.get(i)));
        }
        return out;
    }

    private String callRaw(JCMethodInvocation mi, MethodSymbol m) {
        List<String> a = args(m, mi.args, mi.varargsElement);
        if (m.isStatic()) {
            return staticFunc(m) + "(" + String.join(", ", a) + ")";
        }
        JCExpression meth = mi.meth;
        if (meth instanceof JCIdent) {
            ImplicitThis it = implicitThis((ClassSymbol) m.owner);
            return invoke(it.expr, it.info.sym.type, m, a);
        }
        JCFieldAccess fa = (JCFieldAccess) meth;
        if (fa.selected instanceof JCIdent sid && sid.name == model.names._super) {
            return superCall(m, a);
        }
        if (fa.selected instanceof JCFieldAccess sfa && sfa.name == model.names._super) {
            throw new Unsupported("interface super call");
        }
        return invoke(expr(fa.selected), fa.selected.type, m, a);
    }

    String staticFunc(MethodSymbol m) {
        if (model.isJts(m.owner)) {
            tr.requireVisible(model.info(m.owner));
            return model.staticMethodName(m);
        }
        if (Runtime.qualifiedName((ClassSymbol) m.owner) == null) {
            throw new Unsupported("static method of " + m.owner);
        }
        return tr.requireFunc(model.staticMethodName(m));
    }

    String invoke(String recv, Type recvType, MethodSymbol m, List<String> a) {
        String n = model.instanceMethodName(m);
        String args = String.join(", ", a);
        Translator.Repr r = tr.repr(recvType);
        switch (r.kind()) {
            case JCLASS: {
                Model.ClassInfo ci = r.info();
                if (tr.isPrivate(m) && model.isJts(m.owner)) {
                    return recv + "." + n + "(" + args + ")";
                }
                if (ci.members.contains(n)) {
                    Model.ClassInfo intro = model.introducer(ci, n);
                    if (intro.dispatchers.containsKey(n) && !model.overriddenStrictlyBelow(ci, n)
                            && model.hasImplAtOrAbove(ci, n, intro)) {
                        return recv + "." + n + "_impl(" + args + ")";
                    }
                    return recv + "." + n + "(" + args + ")";
                }
                if (model.isObjectMethod(m)) {
                    return objectImpl(n, tr.toAny(recv, r), a);
                }
                ClassSymbol rs = ci.root.rtSuper;
                if (rs == null) {
                    throw new Unsupported("method " + m + " not found on " + ci.goName);
                }
                tr.requireMethod(Runtime.qualifiedName(rs), n);
                return recv + "." + n + "(" + args + ")";
            }
            case JIFACE:
                if (model.isObjectMethod(m)) {
                    return objectDynamic(n, recv, a);
                }
                return recv + "." + n + "(" + args + ")";
            case ANY: {
                String owner = m.owner.name.toString();
                String f = tr.requireFunc("java." + owner + "_" + n);
                return f + "(" + recv + (a.isEmpty() ? "" : ", " + args) + ")";
            }
            case STRING:
                tr.requireMethod("java.String", n);
                return recv + "." + n + "(" + args + ")";
            case RTCLASS:
            case RTIFACE:
                if (model.isObjectMethod(m) && (r.kind() == Translator.Kind.RTIFACE
                        || !tr.rt.hasMethod(r.go().replace("*", ""), n))) {
                    return objectDynamic(n, tr.toAny(recv, r), a);
                }
                MethodSymbol root = model.root(m);
                if ((root.flags() & com.sun.tools.javac.code.Flags.DEFAULT) != 0
                        && !tr.rt.hasMethod(r.go().replace("*", ""), n)) {
                    // A JDK default method that the runtime's Go type lacks is
                    // a function taking the receiver.
                    String f = tr.requireFunc("java." + root.owner.name + "_default_" + n);
                    return f + "(" + tr.toAny(recv, r) + (a.isEmpty() ? "" : ", " + args) + ")";
                }
                tr.requireMethod(r.go().replace("*", ""), n);
                return recv + "." + n + "(" + args + ")";
            case ARRAY:
                if (m.name.contentEquals("clone")) {
                    tr.requireFunc("java.CloneSlice");
                    return "java.CloneSlice(" + recv + ")";
                }
                return objectImpl(n, tr.toAny(recv, r), a);
            default:
                throw new Unsupported("call on " + recvType);
        }
    }

    private String objectImpl(String n, String recvAny, List<String> a) {
        String f = tr.requireFunc("java.ObjectImpl_" + n);
        return f + "(" + recvAny + (a.isEmpty() ? "" : ", " + String.join(", ", a)) + ")";
    }

    private String objectDynamic(String n, String recvAny, List<String> a) {
        String f = tr.requireFunc("java.Object_" + n);
        return f + "(" + recvAny + (a.isEmpty() ? "" : ", " + String.join(", ", a)) + ")";
    }

    private String superCall(MethodSymbol m, List<String> a) {
        String n = model.instanceMethodName(m);
        String args = String.join(", ", a);
        Model.ClassInfo s = cls.superInfo;
        if (s != null) {
            if (s.members.contains(n)) {
                Model.ClassInfo intro = model.introducer(s, n);
                String name = intro.dispatchers.containsKey(n) && model.hasImplAtOrAbove(s, n, intro)
                        ? n + "_impl" : n;
                return "this." + s.goName + "." + name + "(" + args + ")";
            }
            if (model.isObjectMethod(m)) {
                return objectImpl(n, tr.toAny("this", cls.sym.type), a);
            }
            ClassSymbol rs = s.root.rtSuper;
            if (rs == null) {
                throw new Unsupported("super method " + m);
            }
            tr.requireMethod(Runtime.qualifiedName(rs), n);
            return "this." + s.goName + "." + n + "(" + args + ")";
        }
        if (cls.rtSuper != null) {
            if (model.isObjectMethod(m) && !rtSuperOverrides(cls.rtSuper, m)) {
                return objectImpl(n, tr.toAny("this", cls.sym.type), a);
            }
            String q = Runtime.qualifiedName(cls.rtSuper);
            tr.requireMethod(q, n);
            return "this." + q.substring(q.indexOf('.') + 1) + "." + n + "(" + args + ")";
        }
        return objectImpl(n, tr.toAny("this", cls.sym.type), a);
    }

    private boolean rtSuperOverrides(ClassSymbol rs, MethodSymbol m) {
        for (Type t : model.types.closure(rs.type)) {
            if (t.tsym == model.syms.objectType.tsym) {
                continue;
            }
            for (Symbol s : t.tsym.members().getSymbolsByName(m.name)) {
                if (s instanceof MethodSymbol ms && ms.overrides(m, rs, model.types, false)) {
                    return true;
                }
            }
        }
        return false;
    }

    /** Translates this(...) or super(...) at the start of a constructor. */
    String ctorCall(JCMethodInvocation mi, MethodSymbol ctor) {
        List<String> a = args(ctor, mi.args, mi.varargsElement);
        String args = String.join(", ", a);
        com.sun.tools.javac.util.Name callee = TreeInfo.name(mi.meth);
        if (callee == model.names._this) {
            return "this." + model.ctorInitName(ctor) + "(" + args + ")";
        }
        ClassSymbol owner = (ClassSymbol) ctor.owner;
        if (owner == model.syms.objectType.tsym || owner == model.syms.enumSym) {
            return "";
        }
        if (model.isJts(owner)) {
            Model.ClassInfo si = model.info(owner);
            tr.requireVisible(si);
            String outer = "";
            if (si.hasOuter) {
                throw new Unsupported("super constructor of inner class");
            }
            return "this." + si.goName + "." + model.ctorInitName(ctor) + "(" + outer + args + ")";
        }
        String q = Runtime.qualifiedName(owner);
        if (q == null) {
            throw new Unsupported("super constructor of " + owner);
        }
        String init = model.ctorInitName(ctor);
        tr.requireMethod(q, init);
        return "this." + q.substring(q.indexOf('.') + 1) + "." + init + "(" + args + ")";
    }

    private String newClass(JCNewClass nc) {
        MethodSymbol ctor = (MethodSymbol) nc.constructor;
        ClassSymbol owner = (ClassSymbol) ctor.owner;
        List<String> a = new ArrayList<>();
        if (model.isJts(owner)) {
            Model.ClassInfo ci = model.info(owner);
            tr.requireVisible(ci);
            if (ci.isAbstract()) {
                throw new Unsupported("instantiating abstract class");
            }
            if (ci.hasOuter) {
                if (nc.encl != null) {
                    a.add(expr(nc.encl));
                } else {
                    a.add(implicitThis(ci.sym.owner.enclClass()).expr);
                }
            }
            for (VarSymbol v : ci.captured) {
                a.add(capturedValue(v));
            }
            a.addAll(args(ctor, nc.args, nc.varargsElement));
            return model.ctorNewName(ctor) + "(" + String.join(", ", a) + ")";
        }
        if (Runtime.qualifiedName(owner) == null) {
            throw new Unsupported("new " + owner);
        }
        a.addAll(args(ctor, nc.args, nc.varargsElement));
        return tr.requireFunc(model.ctorNewName(ctor)) + "(" + String.join(", ", a) + ")";
    }

    private String capturedValue(VarSymbol v) {
        String n = locals.get(v);
        if (n != null) {
            return n;
        }
        if (cls.captured.contains(v)) {
            return "this.cap_" + v.name;
        }
        throw new Unsupported("captured variable not in scope: " + v.name);
    }

    private String newArray(JCNewArray na) {
        Type t = na.type;
        String gt = tr.goType(t);
        if (na.elems != null) {
            Type elem = ((Type.ArrayType) t).elemtype;
            StringJoiner sj = new StringJoiner(", ", gt + "{", "}");
            for (JCExpression el : na.elems) {
                sj.add(exprTo(el, elem));
            }
            return sj.toString();
        }
        if (na.dims.size() == 1) {
            return "make(" + gt + ", " + exprTo(na.dims.head, model.syms.intType) + ")";
        }
        Type inner = t;
        for (int i = 0; i < na.dims.size(); i++) {
            inner = ((Type.ArrayType) inner).elemtype;
        }
        String f = tr.requireFunc("java.NewArray" + na.dims.size());
        StringJoiner sj = new StringJoiner(", ", f + "[" + tr.goType(inner) + "](", ")");
        for (JCExpression d : na.dims) {
            sj.add(exprTo(d, model.syms.intType));
        }
        return sj.toString();
    }

    // ---------------------------------------------------------------------
    // Lambdas and method references

    /** A function literal implementing the functional interface type fi. */
    private record FuncShape(String adapter, MethodSymbol desc, Type goSig, List<String> params, String ret) {}

    private FuncShape funcShape(Type fi) {
        MethodSymbol desc = (MethodSymbol) model.types.findDescriptorSymbol(fi.tsym);
        Type goSig = tr.goSignature(desc);
        List<String> params = new ArrayList<>();
        int i = 0;
        for (Type p : goSig.getParameterTypes()) {
            params.add("jl" + i++ + " " + tr.goType(p));
        }
        Type r = goSig.getReturnType();
        String ret = r.hasTag(TypeTag.VOID) ? "" : " " + tr.goType(r);
        return new FuncShape(tr.lambdaAdapter(fi), desc, goSig, params, ret);
    }

    private String lambda(JCLambda l) {
        FuncShape f = funcShape(l.type);
        Body c = child(f.goSig.getReturnType());
        c.scanReads(l.body);
        List<Type> goParams = f.goSig.getParameterTypes();
        for (int i = 0; i < l.params.size(); i++) {
            VarSymbol p = l.params.get(i).sym;
            String n = c.declareLocal(p);
            c.w("var " + n + " " + tr.goType(p.type) + " = " + tr.conv("jl" + i, goParams.get(i), p.type));
            c.markUsed(n, p);
        }
        if (l.body instanceof JCBlock blk) {
            c.stmts(blk.stats);
            if (!f.ret.isEmpty()) {
                c.w("panic(\"unreachable\")");
            }
        } else if (f.ret.isEmpty()) {
            c.exprStmt((JCExpression) l.body);
        } else {
            c.w("return " + c.exprTo((JCExpression) l.body, f.goSig.getReturnType()));
        }
        return f.adapter + "(func(" + String.join(", ", f.params) + ")" + f.ret + " {\n" + c.out + "})";
    }

    private String memberRef(JCMemberReference r) {
        if (r.getMode() != com.sun.source.tree.MemberReferenceTree.ReferenceMode.INVOKE) {
            throw new Unsupported("constructor reference");
        }
        FuncShape f = funcShape(r.type);
        MethodSymbol target = (MethodSymbol) r.sym;
        List<Type> goParams = f.goSig.getParameterTypes();
        List<Type> targetParams = tr.goSignature(target).getParameterTypes();
        String bound = null;
        String call;
        switch (r.kind) {
            case STATIC: {
                List<String> a = new ArrayList<>();
                for (int i = 0; i < targetParams.size(); i++) {
                    a.add(tr.conv("jl" + i, goParams.get(i), targetParams.get(i)));
                }
                call = staticFunc(target) + "(" + String.join(", ", a) + ")";
                break;
            }
            case UNBOUND: {
                List<String> a = new ArrayList<>();
                for (int i = 0; i < targetParams.size(); i++) {
                    a.add(tr.conv("jl" + (i + 1), goParams.get(i + 1), targetParams.get(i)));
                }
                String recv = tr.conv("jl0", goParams.get(0), r.expr.type);
                call = invoke(recv, r.expr.type, target, a);
                break;
            }
            case BOUND: {
                bound = expr(r.expr);
                List<String> a = new ArrayList<>();
                for (int i = 0; i < targetParams.size(); i++) {
                    a.add(tr.conv("jl" + i, goParams.get(i), targetParams.get(i)));
                }
                call = invoke("jrecv", r.expr.type, target, a);
                break;
            }
            default:
                throw new Unsupported("method reference kind " + r.kind);
        }
        String body;
        if (f.ret.isEmpty()) {
            body = call;
        } else {
            body = "return " + tr.conv(call, tr.goSignature(target).getReturnType(), f.goSig.getReturnType());
        }
        String fn = f.adapter + "(func(" + String.join(", ", f.params) + ")" + f.ret + " { " + body + " })";
        if (bound == null) {
            return fn;
        }
        String t = tr.goType(r.expr.type);
        return "func() " + tr.goType(r.type) + " { jrecv := " + bound + "; return " + fn + " }()";
    }

    // ---------------------------------------------------------------------
    // Operators

    private String unary(JCUnary u) {
        switch (u.getTag()) {
            case PREINC, PREDEC, POSTINC, POSTDEC: {
                checkIncType(u.arg.type);
                String f = switch (u.getTag()) {
                    case PREINC -> "java.PreInc";
                    case PREDEC -> "java.PreDec";
                    case POSTINC -> "java.PostInc";
                    default -> "java.PostDec";
                };
                tr.requireFunc(f);
                return f + "(&" + lvalue(u.arg) + ")";
            }
            case NOT:
                return "(!" + cond(u.arg) + ")";
            default:
                break;
        }
        OperatorSymbol op = u.operator;
        Type pt = op.type.getParameterTypes().head;
        String x = exprTo(u.arg, pt);
        return switch (u.getTag()) {
            case NEG -> "(-(" + x + "))";
            case COMPL -> "(^(" + x + "))";
            case POS -> x;
            default -> throw new Unsupported("unary " + u.getTag());
        };
    }

    private String binary(JCBinary b) {
        JCTree.Tag tag = b.getTag();
        if (tag == JCTree.Tag.AND) {
            return "(" + cond(b.lhs) + " && " + cond(b.rhs) + ")";
        }
        if (tag == JCTree.Tag.OR) {
            return "(" + cond(b.lhs) + " || " + cond(b.rhs) + ")";
        }
        OperatorSymbol op = b.operator;
        List<Type> pt = op.type.getParameterTypes();
        Type res = op.type.getReturnType();
        if (tag == JCTree.Tag.PLUS && tr.isString(res)) {
            return concat(b);
        }
        if ((tag == JCTree.Tag.EQ || tag == JCTree.Tag.NE) && !pt.get(0).isPrimitive()) {
            return refEq(b, tag == JCTree.Tag.EQ);
        }
        String l = exprTo(b.lhs, pt.get(0));
        String r = exprTo(b.rhs, pt.get(1));
        return binop(tag, l, r, pt.get(0), pt.get(1), res);
    }

    private String binop(JCTree.Tag tag, String l, String r, Type lt, Type rt, Type res) {
        boolean fp = tr.isFloating(res);
        switch (tag) {
            case PLUS:
                return "(" + l + " + " + r + ")";
            case MINUS:
                return "(" + l + " - " + r + ")";
            case MUL:
                if (fp) {
                    return tr.goType(res) + "(" + l + " * " + r + ")";
                }
                return "(" + l + " * " + r + ")";
            case DIV:
                return "(" + l + " / " + r + ")";
            case MOD:
                if (res.hasTag(TypeTag.DOUBLE)) {
                    return "math.Mod(" + l + ", " + r + ")";
                }
                if (res.hasTag(TypeTag.FLOAT)) {
                    return "float32(math.Mod(float64(" + l + "), float64(" + r + ")))";
                }
                return "(" + l + " % " + r + ")";
            case SL:
                return "(" + l + " << (" + r + " & " + shiftMask(res) + "))";
            case SR:
                return "(" + l + " >> (" + r + " & " + shiftMask(res) + "))";
            case USR:
                if (res.hasTag(TypeTag.LONG)) {
                    return "int64(uint64(" + l + ") >> (" + r + " & 63))";
                }
                return "int32(uint32(" + l + ") >> (" + r + " & 31))";
            case BITAND:
                if (res.hasTag(TypeTag.BOOLEAN)) {
                    return tr.requireFunc("java.BoolAnd") + "(" + l + ", " + r + ")";
                }
                return "(" + l + " & " + r + ")";
            case BITOR:
                if (res.hasTag(TypeTag.BOOLEAN)) {
                    return tr.requireFunc("java.BoolOr") + "(" + l + ", " + r + ")";
                }
                return "(" + l + " | " + r + ")";
            case BITXOR:
                if (res.hasTag(TypeTag.BOOLEAN)) {
                    return "(" + l + " != " + r + ")";
                }
                return "(" + l + " ^ " + r + ")";
            case EQ:
                return "(" + l + " == " + r + ")";
            case NE:
                return "(" + l + " != " + r + ")";
            case LT:
                return "(" + l + " < " + r + ")";
            case GT:
                return "(" + l + " > " + r + ")";
            case LE:
                return "(" + l + " <= " + r + ")";
            case GE:
                return "(" + l + " >= " + r + ")";
            default:
                throw new Unsupported("binary " + tag);
        }
    }

    private static String shiftMask(Type res) {
        return res.hasTag(TypeTag.LONG) ? "63" : "31";
    }

    private String refEq(JCBinary b, boolean eq) {
        String op = eq ? " == " : " != ";
        if (b.lhs.type.hasTag(TypeTag.BOT)) {
            return "(" + expr(b.rhs) + op + "nil)";
        }
        if (b.rhs.type.hasTag(TypeTag.BOT)) {
            return "(" + expr(b.lhs) + op + "nil)";
        }
        Translator.Repr l = tr.repr(b.lhs.type);
        Translator.Repr r = tr.repr(b.rhs.type);
        if (l.kind() == Translator.Kind.JCLASS && r.kind() == Translator.Kind.JCLASS && l.info().root == r.info().root) {
            return "(unsafe.Pointer(" + expr(b.lhs) + ")" + op + "unsafe.Pointer(" + expr(b.rhs) + "))";
        }
        if (l.go().equals(r.go()) && l.kind() != Translator.Kind.ARRAY && l.kind() != Translator.Kind.ANY) {
            return "(" + expr(b.lhs) + op + expr(b.rhs) + ")";
        }
        // Go's == panics on interfaces holding slices (Java arrays).
        String same = tr.requireFunc("java.Same") + "(" + tr.toAny(expr(b.lhs), l) + ", " + tr.toAny(expr(b.rhs), r) + ")";
        return eq ? same : "(!" + same + ")";
    }

    private String concat(JCBinary b) {
        List<JCExpression> parts = new ArrayList<>();
        flattenConcat(b, parts);
        StringJoiner sj = new StringJoiner(", ", tr.requireFunc("java.Concat") + "(", ")");
        for (JCExpression p : parts) {
            sj.add(strOf(expr(p), p.type));
        }
        return sj.toString();
    }

    private void flattenConcat(JCExpression e, List<JCExpression> parts) {
        if (e instanceof JCBinary b && b.getTag() == JCTree.Tag.PLUS && tr.isString(b.type)
                && b.type.constValue() == null) {
            flattenConcat(b.lhs, parts);
            flattenConcat(b.rhs, parts);
        } else {
            parts.add(e);
        }
    }

    String strOf(String e, Type t) {
        return switch (t.getTag()) {
            case CHAR -> tr.requireFunc("java.StrOfChar") + "(" + e + ")";
            case BYTE, SHORT, INT -> tr.requireFunc("java.StrOfInt") + "(int32(" + e + "))";
            case LONG -> tr.requireFunc("java.StrOfLong") + "(" + e + ")";
            case FLOAT -> tr.requireFunc("java.StrOfFloat") + "(" + e + ")";
            case DOUBLE -> tr.requireFunc("java.StrOfDouble") + "(" + e + ")";
            case BOOLEAN -> tr.requireFunc("java.StrOfBool") + "(" + e + ")";
            case BOT -> tr.literal("null");
            default -> tr.isString(t) ? e : tr.requireFunc("java.StrOfObject") + "(" + tr.toAny(e, t) + ")";
        };
    }

    String instanceOf(String e, Type from, Type target) {
        Translator.Repr f = tr.repr(from);
        Translator.Repr t = tr.repr(target);
        switch (t.kind()) {
            case JCLASS:
                if (f.kind() == Translator.Kind.JCLASS && f.info().root == t.info().root && t.info().polymorphic()) {
                    return t.info().goName + "_isKind(" + tr.upcast(e, t.info().root) + ")";
                }
                return t.info().goName + "_isInstance(" + tr.toAny(e, f) + ")";
            case JIFACE:
            case RTIFACE:
            case STRING:
            case ARRAY:
                tr.requireFunc("java.InstanceOf");
                return "java.InstanceOf[" + t.go() + "](" + tr.toAny(e, f) + ")";
            case RTCLASS: {
                String q = t.go().substring(1);
                return tr.requireFunc(q + "_isInstance") + "(" + tr.toAny(e, f) + ")";
            }
            case ANY: {
                String q = "java." + model.types.erasure(target).tsym.name + "_isInstance";
                return tr.requireFunc(q) + "(" + tr.toAny(e, f) + ")";
            }
            default:
                throw new Unsupported("instanceof " + target);
        }
    }
}
