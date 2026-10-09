/*
 * HashVectors prints test vectors for the Go implementation of
 * java.util.HashMap, LinkedHashMap, HashSet and LinkedHashSet in
 * internal/jts/java (hashmap.go and hashset.go). Regenerate the vectors from
 * the repository root with:
 *
 *   java -XX:+UnlockExperimentalVMOptions -XX:hashCode=3 \
 *     --add-opens java.base/java.util=ALL-UNNAMED \
 *     internal/jts/tools/jdkvectors/HashVectors.java > internal/jts/java/testdata/hash.txt
 *
 * -XX:hashCode=3 makes identity hash codes come from a global counter, as in
 * the oracle. --add-opens lets the program read the collections' internal
 * state (table, threshold, bin lists and red-black trees) by reflection.
 *
 * The program runs sequences of random operations on one collection and
 * prints, after each operation, its result and a digest of the collection's
 * iteration order and of its internal state. Keys are Integers, Strings,
 * keys with controllable hash codes that are Comparable to themselves (CKey)
 * or not Comparable at all (NKey), keys with identity hash codes (OKey),
 * keys that are not equal to themselves (XKey), and mixtures of those with
 * Longs, Doubles, Booleans and null. Colliding hash codes force bins to be
 * treeified, split and untreeified, which changes iteration order.
 *
 * Output format, one sequence after another:
 *
 *   seq <n> <container> <constructor> <arg>...
 *   h0 <identity hash of an Object created just before the sequence>
 *   k <type> <param>...          one line per pool key, indexed from 0
 *   <op> <arg>... = <result> <size> <order digest> <state digest>
 *   ih <pool index>:<identity hash>...
 *   end
 *
 * The constructor line is followed by "= EX:<class>:<message>" instead of
 * the operations when the constructor throws. Keys are referred to by pool
 * index and values (Integers or null) by "null" or their decimal value.
 * Digests are the 32-bit FNV-1a hash of the strings built by order() and
 * state(), in hex. The ih line gives the identity hash code of every CKey,
 * NKey, OKey and XKey in the pool, read in pool order after the sequence.
 *
 * The arguments "verbose <n>" make the program also print the full order
 * and state strings for sequence n to stderr, for debugging.
 *
 * The program uses java.util.Random with fixed seeds, so its output is
 * deterministic. It avoids string concatenation with "+" and lambdas, whose
 * first use computes identity hash codes inside the JVM, and runs a warm-up
 * pass before the printed sequences for the same reason, so that the
 * identity hash counter advances only for the sequences' own keys.
 */

import java.io.BufferedWriter;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.io.PrintStream;
import java.io.Writer;
import java.lang.reflect.Field;
import java.util.AbstractMap;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Random;

public final class HashVectors {

    /**
     * A key with a chosen hash code that is Comparable to its own class. Its
     * hash code can change while it is in a collection, as a mutable JTS key
     * such as a Coordinate's can.
     */
    static class CKey implements Comparable<CKey> {
        final int idx;
        final int id;
        int hash;
        final int rank;

        CKey(int idx, int id, int hash, int rank) {
            this.idx = idx;
            this.id = id;
            this.hash = hash;
            this.rank = rank;
        }

        @Override
        public int hashCode() {
            return hash;
        }

        @Override
        public boolean equals(Object o) {
            return o != null && o.getClass() == getClass() && ((CKey) o).id == id;
        }

        @Override
        public int compareTo(CKey o) {
            return Integer.compare(rank, o.rank);
        }

        @Override
        public String toString() {
            return new StringBuilder().append('c').append(idx).toString();
        }
    }

    /** A subclass of CKey, which HashMap does not treat as comparable. */
    static final class CKeySub extends CKey {
        CKeySub(int idx, int id, int hash, int rank) {
            super(idx, id, hash, rank);
        }

        @Override
        public String toString() {
            return new StringBuilder().append('t').append(idx).toString();
        }
    }

    /** A key with a chosen hash code that is not Comparable. */
    static final class NKey {
        final int idx;
        final int id;
        int hash;

        NKey(int idx, int id, int hash) {
            this.idx = idx;
            this.id = id;
            this.hash = hash;
        }

        @Override
        public int hashCode() {
            return hash;
        }

        @Override
        public boolean equals(Object o) {
            return o instanceof NKey && ((NKey) o).id == id;
        }

        @Override
        public String toString() {
            return new StringBuilder().append('n').append(idx).toString();
        }
    }

    /**
     * A key with Object's hashCode and equals, so its hash code is its
     * identity hash code.
     */
    static final class OKey {
        final int idx;

        OKey(int idx) {
            this.idx = idx;
        }

        @Override
        public String toString() {
            return new StringBuilder().append('o').append(idx).toString();
        }
    }

    /**
     * A key whose equals is false even for itself, so that it is only found
     * by the identity tests that HashMap makes before calling equals.
     */
    static final class XKey {
        final int idx;
        int hash;

        XKey(int idx, int hash) {
            this.idx = idx;
            this.hash = hash;
        }

        @Override
        public int hashCode() {
            return hash;
        }

        @Override
        public boolean equals(Object o) {
            return false;
        }

        @Override
        public String toString() {
            return new StringBuilder().append('x').append(idx).toString();
        }
    }

    // Reflective access to the collections' internals.
    static Field fTable;
    static Field fThreshold;
    static Field fSize;
    static Field fNext;
    static Field fKey;
    static Field fParent;
    static Field fLeft;
    static Field fRight;
    static Field fPrev;
    static Field fRed;
    static Field fBefore;
    static Field fAfter;
    static Field fHead;
    static Field fTail;
    static Field fSetMap;
    static Class<?> treeNodeClass;

    static Field field(Class<?> c, String name) throws Exception {
        Field f = c.getDeclaredField(name);
        f.setAccessible(true);
        return f;
    }

    static void initReflection() throws Exception {
        fTable = field(HashMap.class, "table");
        fThreshold = field(HashMap.class, "threshold");
        fSize = field(HashMap.class, "size");
        Class<?> node = Class.forName("java.util.HashMap$Node");
        fNext = field(node, "next");
        fKey = field(node, "key");
        treeNodeClass = Class.forName("java.util.HashMap$TreeNode");
        fParent = field(treeNodeClass, "parent");
        fLeft = field(treeNodeClass, "left");
        fRight = field(treeNodeClass, "right");
        fPrev = field(treeNodeClass, "prev");
        fRed = field(treeNodeClass, "red");
        Class<?> entry = Class.forName("java.util.LinkedHashMap$Entry");
        fBefore = field(entry, "before");
        fAfter = field(entry, "after");
        fHead = field(LinkedHashMap.class, "head");
        fTail = field(LinkedHashMap.class, "tail");
        fSetMap = field(HashSet.class, "map");
    }

    public static void main(String[] args) throws Exception {
        initReflection();
        int verbose = -1;
        if (args.length == 2 && args[0].equals("verbose")) {
            verbose = Integer.parseInt(args[1]);
        }
        // Warm-up pass: same code paths, output discarded.
        run(new StringBuilder(), -2, null);
        Writer w = new BufferedWriter(new OutputStreamWriter(System.out, "UTF-8"));
        StringBuilder sb = new StringBuilder();
        stepCount = 0;
        treeSteps = 0;
        run(sb, verbose, w);
        w.write(sb.toString());
        w.flush();
        System.err.println(new StringBuilder().append(stepCount).append(" steps, ").append(treeSteps)
                .append(" with a treeified bin"));
    }

    static int stepCount;
    static int treeSteps;

    static final int FAMILY_INT = 0;
    static final int FAMILY_STR = 1;
    static final int FAMILY_CKEY = 2;
    static final int FAMILY_NKEY = 3;
    static final int FAMILY_MIXED = 4;
    static final int FAMILY_OKEY = 5;
    static final int FAMILY_SPLIT = 6;

    static final String[] CONTAINERS = {"HM", "LHM", "LHMA", "HS", "LHS"};

    static void run(StringBuilder out, int verbose, Writer w) throws IOException {
        Random r = new Random(20261009L);
        int n = 0;
        n = ctorErrors(out, n);
        for (int round = 0; round < 8; round++) {
            for (int family = 0; family <= FAMILY_SPLIT; family++) {
                if (round >= 3 && (family == FAMILY_INT || family == FAMILY_STR || family == FAMILY_OKEY)) {
                    continue;
                }
                for (int c = 0; c < CONTAINERS.length; c++) {
                    Seq s = new Seq(n, CONTAINERS[c], family, new Random(r.nextLong()), out, n == verbose);
                    s.run();
                    n++;
                    if (w != null && out.length() > 1 << 20) {
                        w.write(out.toString());
                        out.setLength(0);
                    }
                }
            }
        }
        if (w == null) {
            out.setLength(0);
        }
    }

    // ------------------------------------------------------------------
    // Constructor errors

    static int ctorErrors(StringBuilder out, int n) {
        int[][] intArgs = {{-1}, {-100}};
        for (int[] a : intArgs) {
            ctorError(out, n++, "HM", "new_int", a[0], 0f, false);
            ctorError(out, n++, "HS", "new_int", a[0], 0f, false);
            ctorError(out, n++, "LHM", "new_int", a[0], 0f, false);
            ctorError(out, n++, "LHS", "new_int", a[0], 0f, false);
        }
        float[] lfs = {0f, -1f, Float.NaN, -0f, Float.NEGATIVE_INFINITY};
        for (float lf : lfs) {
            ctorError(out, n++, "HM", "new_int_float", 16, lf, true);
            ctorError(out, n++, "HS", "new_int_float", 16, lf, true);
            ctorError(out, n++, "LHM", "new_int_float", 16, lf, true);
            ctorError(out, n++, "LHS", "new_int_float", 16, lf, true);
            ctorError(out, n++, "LHMA", "new_int_float_boolean", 16, lf, true);
        }
        ctorError(out, n++, "HM", "new_int_float", -1, Float.NaN, true);
        return n;
    }

    static void ctorError(StringBuilder out, int n, String container, String ctor, int cap, float lf,
            boolean withLf) {
        out.append("seq ").append(n).append(' ').append(container).append(' ').append(ctor).append(' ')
                .append(cap);
        if (withLf) {
            out.append(' ').append(floatHex(lf));
            if (container.equals("LHMA")) {
                out.append(" true");
            }
        }
        out.append('\n');
        String res;
        try {
            Object o;
            switch (container) {
                case "HM":
                    o = withLf ? new HashMap<Object, Object>(cap, lf) : new HashMap<Object, Object>(cap);
                    break;
                case "HS":
                    o = withLf ? new HashSet<Object>(cap, lf) : new HashSet<Object>(cap);
                    break;
                case "LHM":
                    o = withLf ? new LinkedHashMap<Object, Object>(cap, lf) : new LinkedHashMap<Object, Object>(cap);
                    break;
                case "LHMA":
                    o = new LinkedHashMap<Object, Object>(cap, lf, true);
                    break;
                default:
                    o = withLf ? new LinkedHashSet<Object>(cap, lf) : new LinkedHashSet<Object>(cap);
                    break;
            }
            res = o.getClass().getName();
        } catch (RuntimeException e) {
            res = exToken(e);
        }
        out.append("= ").append(res).append('\n');
        out.append("end\n");
    }

    static String floatHex(float f) {
        String h = Integer.toHexString(Float.floatToRawIntBits(f));
        StringBuilder sb = new StringBuilder();
        for (int i = h.length(); i < 8; i++) {
            sb.append('0');
        }
        return sb.append(h).toString();
    }

    static String exToken(RuntimeException e) {
        StringBuilder sb = new StringBuilder("EX:").append(e.getClass().getName());
        if (e.getMessage() != null) {
            sb.append(':').append(e.getMessage().replace(' ', '_'));
        }
        return sb.toString();
    }

    // ------------------------------------------------------------------
    // Tokens and digests

    static void tok(StringBuilder sb, Object o) {
        if (o == null) {
            sb.append("null");
        } else if (o instanceof Integer) {
            sb.append('i').append(((Integer) o).intValue());
        } else if (o instanceof Long) {
            sb.append('l').append(((Long) o).longValue());
        } else if (o instanceof Double) {
            sb.append('d').append(Long.toHexString(Double.doubleToRawLongBits((Double) o)));
        } else if (o instanceof Boolean) {
            sb.append('b').append(((Boolean) o).booleanValue() ? '1' : '0');
        } else if (o instanceof String) {
            sb.append('s').append((String) o);
        } else {
            sb.append(o.toString());
        }
    }

    static void vtok(StringBuilder sb, Object v) {
        if (v == null) {
            sb.append("null");
        } else {
            sb.append(((Integer) v).intValue());
        }
    }

    static String digest(CharSequence s) {
        int h = 0x811c9dc5;
        for (int i = 0; i < s.length(); i++) {
            h ^= s.charAt(i) & 0xff;
            h *= 0x01000193;
        }
        String x = Integer.toHexString(h);
        StringBuilder sb = new StringBuilder();
        for (int i = x.length(); i < 8; i++) {
            sb.append('0');
        }
        return sb.append(x).toString();
    }

    // ------------------------------------------------------------------
    // Sequences

    static final class Seq {
        final int n;
        String container;
        final int family;
        final Random r;
        final StringBuilder out;
        final boolean verbose;
        Object[] pool;
        HashMap<Object, Object> map;
        HashSet<Object> set;

        Seq(int n, String container, int family, Random r, StringBuilder out, boolean verbose) {
            this.n = n;
            this.container = container;
            this.family = family;
            this.r = r;
            this.out = out;
            this.verbose = verbose;
        }

        boolean isMap() {
            return map != null;
        }

        void run() {
            int h0 = System.identityHashCode(new Object());
            pool = makePool();
            StringBuilder hdr = new StringBuilder();
            hdr.append("seq ").append(n).append(' ').append(container).append(' ');
            construct(hdr);
            out.append(hdr).append('\n');
            out.append("h0 ").append(h0).append('\n');
            for (int i = 0; i < pool.length; i++) {
                out.append("k ");
                describeKey(out, pool[i]);
                out.append('\n');
            }
            // The constructor ran before the pool was printed, but its
            // outcome is recorded as the first step.
            step(new StringBuilder("init"), "ok");
            // Lookups before the first insertion, when the table may not
            // have been allocated: Java does not compute the key's hash code
            // then, except for remove.
            for (int i = 0; i < 2; i++) {
                StringBuilder op = new StringBuilder();
                int k = key();
                String res;
                if (isMap()) {
                    switch (r.nextInt(4)) {
                        case 0:
                            op.append("get ").append(k);
                            res = vstr(map.get(pool[k]));
                            break;
                        case 1:
                            op.append("containsKey ").append(k);
                            res = bstr(map.containsKey(pool[k]));
                            break;
                        case 2:
                            op.append("remove ").append(k);
                            res = vstr(map.remove(pool[k]));
                            break;
                        default:
                            op.append("replace2 ").append(k).append(" 1");
                            res = vstr(map.replace(pool[k], Integer.valueOf(1)));
                            break;
                    }
                } else if (r.nextBoolean()) {
                    op.append("contains ").append(k);
                    res = bstr(set.contains(pool[k]));
                } else {
                    op.append("remove ").append(k);
                    res = bstr(set.remove(pool[k]));
                }
                step(op, res);
            }
            int steps = 50 + r.nextInt(200);
            int phaseLen = Math.max(10, steps / 4);
            for (int i = 0; i < steps; i++) {
                int phase = (i / phaseLen) % 4;
                StringBuilder op = new StringBuilder();
                String res = isMap() ? mapOp(op, phase) : setOp(op, phase);
                step(op, res);
            }
            out.append("ih");
            for (int i = 0; i < pool.length; i++) {
                Object k = pool[i];
                if (k instanceof CKey || k instanceof NKey || k instanceof OKey || k instanceof XKey) {
                    out.append(' ').append(i).append(':').append(System.identityHashCode(k));
                }
            }
            out.append('\n');
            out.append("end\n");
        }

        void step(StringBuilder op, String res) {
            String order = order();
            String state = state();
            stepCount++;
            if (state.indexOf('T') >= 0) {
                treeSteps++;
            }
            out.append(op).append(" = ").append(res).append(' ').append(isMap() ? map.size() : set.size())
                    .append(' ').append(digest(order)).append(' ').append(digest(state)).append('\n');
            if (verbose) {
                PrintStream e = System.err;
                e.println(op);
                e.println(order);
                e.println(state);
            }
        }

        // ---- pools

        Object[] makePool() {
            switch (family) {
                case FAMILY_INT:
                    return intPool();
                case FAMILY_STR:
                    return strPool();
                case FAMILY_CKEY:
                    return ckeyPool();
                case FAMILY_NKEY:
                    return nkeyPool();
                case FAMILY_OKEY:
                    return okeyPool();
                case FAMILY_SPLIT:
                    return splitPool();
                default:
                    return mixedPool();
            }
        }

        Object[] intPool() {
            int size = 10 + r.nextInt(140);
            int mode = r.nextInt(5);
            boolean withNull = r.nextInt(4) == 0;
            Object[] p = new Object[size];
            int c = r.nextInt(4);
            for (int i = 0; i < size; i++) {
                int v;
                switch (mode) {
                    case 0:
                        v = r.nextInt(300);
                        break;
                    case 1:
                        v = r.nextInt();
                        break;
                    case 2: {
                        int j = r.nextInt(1 << 12);
                        v = (j << 16) | (j ^ c);
                        break;
                    }
                    case 3:
                        v = 64 * r.nextInt(64) + r.nextInt(2);
                        break;
                    default:
                        v = r.nextInt(3) == 0 ? -r.nextInt(1000) : 16 * r.nextInt(100);
                        break;
                }
                if (withNull && v == 0) {
                    v = 1;
                }
                p[i] = Integer.valueOf(v);
            }
            if (withNull) {
                p[r.nextInt(size)] = null;
            }
            return p;
        }

        String collidingString(int blocks, String prefix) {
            StringBuilder sb = new StringBuilder(prefix);
            for (int b = 0; b < blocks; b++) {
                sb.append(r.nextBoolean() ? "Aa" : "BB");
            }
            return sb.toString();
        }

        Object[] strPool() {
            int size = 10 + r.nextInt(120);
            boolean withNull = r.nextInt(4) == 0;
            String[] prefixes = {"", "x", "yz", "Q"};
            int blocks = 2 + r.nextInt(3);
            Object[] p = new Object[size];
            for (int i = 0; i < size; i++) {
                String s;
                if (r.nextInt(3) == 0) {
                    int len = r.nextInt(6);
                    StringBuilder sb = new StringBuilder();
                    for (int j = 0; j < len; j++) {
                        sb.append("abcdeAB0".charAt(r.nextInt(8)));
                    }
                    s = sb.toString();
                } else {
                    s = collidingString(blocks, prefixes[r.nextInt(prefixes.length)]);
                }
                if (withNull && s.isEmpty()) {
                    s = "e";
                }
                p[i] = new String(s);
            }
            if (withNull) {
                p[r.nextInt(size)] = null;
            }
            return p;
        }

        int[] hashSet(int count) {
            int[] hs = new int[count];
            for (int i = 0; i < count; i++) {
                switch (r.nextInt(3)) {
                    case 0:
                        hs[i] = r.nextInt(8) + 64 * r.nextInt(8);
                        break;
                    case 1:
                        hs[i] = r.nextInt();
                        break;
                    default:
                        hs[i] = r.nextInt(4) + 16 * r.nextInt(4);
                        break;
                }
            }
            return hs;
        }

        Object[] ckeyPool() {
            int size = 10 + r.nextInt(130);
            int[] hs = hashSet(1 + r.nextInt(6));
            int ranks = 1 + r.nextInt(size);
            Object[] p = new Object[size];
            for (int i = 0; i < size; i++) {
                int t = r.nextInt(20);
                if (t < 3 && i > 0) {
                    // An equal copy of an earlier key.
                    Object x = p[r.nextInt(i)];
                    if (x instanceof CKey) {
                        CKey o = (CKey) x;
                        // Some copies have a different rank, so compareTo is
                        // inconsistent with equals and a tree bin can come
                        // to hold equal keys.
                        int rank = r.nextInt(4) == 0 ? r.nextInt(ranks) : o.rank;
                        p[i] = o instanceof CKeySub ? new CKeySub(i, o.id, o.hash, rank)
                                : new CKey(i, o.id, o.hash, rank);
                        continue;
                    }
                }
                int hash = hs[r.nextInt(hs.length)];
                int rank = r.nextInt(ranks);
                if (t < 6) {
                    p[i] = new CKeySub(i, r.nextInt(size), hash, rank);
                } else {
                    p[i] = new CKey(i, i + 1000, hash, rank);
                }
            }
            if (r.nextInt(4) == 0) {
                p[r.nextInt(size)] = null;
            }
            return p;
        }

        Object[] nkeyPool() {
            int size = 10 + r.nextInt(130);
            int[] hs = hashSet(1 + r.nextInt(6));
            Object[] p = new Object[size];
            for (int i = 0; i < size; i++) {
                if (i > 0 && r.nextInt(8) == 0) {
                    Object x = p[r.nextInt(i)];
                    if (x instanceof NKey) {
                        NKey o = (NKey) x;
                        p[i] = new NKey(i, o.id, o.hash);
                        continue;
                    }
                }
                p[i] = new NKey(i, i + 1000, hs[r.nextInt(hs.length)]);
            }
            if (r.nextInt(4) == 0) {
                p[r.nextInt(size)] = null;
            }
            return p;
        }

        /**
         * Keys whose hashes fall in a few bins of a 64-bin table but differ
         * in the higher bits, so that tree bins are split when the table
         * grows.
         */
        Object[] splitPool() {
            int size = 60 + r.nextInt(200);
            int bins = 1 + r.nextInt(4);
            int spread = 2 + r.nextInt(30);
            int ranks = 1 + r.nextInt(40);
            Object[] p = new Object[size];
            for (int i = 0; i < size; i++) {
                int hash = r.nextInt(bins) + 64 * r.nextInt(spread);
                switch (r.nextInt(5)) {
                    case 0:
                        p[i] = Integer.valueOf(hash == 0 ? 64 * spread : hash);
                        break;
                    case 1:
                        p[i] = new NKey(i, i + 1000, hash);
                        break;
                    case 2:
                        p[i] = new XKey(i, hash);
                        break;
                    default:
                        p[i] = new CKey(i, i + 1000, hash, r.nextInt(ranks));
                        break;
                }
            }
            return p;
        }

        Object[] okeyPool() {
            int size = 10 + r.nextInt(130);
            Object[] p = new Object[size];
            for (int i = 0; i < size; i++) {
                p[i] = new OKey(i);
            }
            if (r.nextInt(4) == 0) {
                p[r.nextInt(size)] = null;
            }
            return p;
        }

        Object[] mixedPool() {
            int size = 10 + r.nextInt(130);
            int[] hs = {2112, 1231, 1237, 5, 69, 133};
            int[] use = new int[1 + r.nextInt(3)];
            for (int i = 0; i < use.length; i++) {
                use[i] = hs[r.nextInt(hs.length)];
            }
            int ranks = 1 + r.nextInt(8);
            Object[] p = new Object[size];
            for (int i = 0; i < size; i++) {
                int h = use[r.nextInt(use.length)];
                Object k;
                switch (r.nextInt(11)) {
                    case 0:
                        k = Integer.valueOf(h);
                        break;
                    case 1:
                        k = Long.valueOf(h & 0xffffffffL);
                        break;
                    case 2:
                        k = Double.valueOf(Double.longBitsToDouble(h & 0xffffffffL));
                        break;
                    case 3:
                        k = h == 1231 ? Boolean.TRUE : h == 1237 ? Boolean.FALSE : Integer.valueOf(h + 1);
                        break;
                    case 4:
                        k = h == 2112 ? new String(r.nextBoolean() ? "Aa" : "BB") : new String("q");
                        break;
                    case 5:
                    case 6:
                        k = new CKey(i, i + 1000, r.nextInt(5) == 0 ? 0 : h, r.nextInt(ranks));
                        break;
                    case 7:
                        k = new CKeySub(i, i + 1000, h, r.nextInt(ranks));
                        break;
                    case 8:
                        k = new OKey(i);
                        break;
                    case 9:
                        k = new XKey(i, h);
                        break;
                    default:
                        k = new NKey(i, i + 1000, r.nextInt(5) == 0 ? 0 : h);
                        break;
                }
                p[i] = k;
            }
            p[r.nextInt(size)] = null;
            return p;
        }

        void describeKey(StringBuilder sb, Object k) {
            if (k == null) {
                sb.append('z');
            } else if (k instanceof Integer) {
                sb.append("i ").append(((Integer) k).intValue());
            } else if (k instanceof Long) {
                sb.append("l ").append(((Long) k).longValue());
            } else if (k instanceof Double) {
                sb.append("d ").append(Long.toHexString(Double.doubleToRawLongBits((Double) k)));
            } else if (k instanceof Boolean) {
                sb.append("b ").append(((Boolean) k).booleanValue() ? 1 : 0);
            } else if (k instanceof String) {
                sb.append("s ").append((String) k);
            } else if (k instanceof OKey) {
                sb.append('o');
            } else if (k instanceof XKey) {
                sb.append("x ").append(((XKey) k).hash);
            } else if (k instanceof CKeySub) {
                CKey c = (CKey) k;
                sb.append("t ").append(c.id).append(' ').append(c.hash).append(' ').append(c.rank);
            } else if (k instanceof CKey) {
                CKey c = (CKey) k;
                sb.append("c ").append(c.id).append(' ').append(c.hash).append(' ').append(c.rank);
            } else {
                NKey c = (NKey) k;
                sb.append("n ").append(c.id).append(' ').append(c.hash);
            }
        }

        // ---- argument generation

        int key() {
            return r.nextInt(pool.length);
        }

        Integer value() {
            return r.nextInt(10) == 0 ? null : Integer.valueOf(r.nextInt(20));
        }

        void appendValue(StringBuilder op, Integer v) {
            op.append(' ');
            vtok(op, v);
        }

        int[] keys(int max) {
            int[] ks = new int[r.nextInt(max + 1)];
            for (int i = 0; i < ks.length; i++) {
                ks[i] = key();
            }
            return ks;
        }

        List<Object> keyList(StringBuilder op, int max) {
            int[] ks = keys(max);
            op.append(' ').append(ks.length);
            List<Object> l = new ArrayList<Object>();
            for (int k : ks) {
                op.append(' ').append(k);
                l.add(pool[k]);
            }
            return l;
        }

        List<Object> valueList(StringBuilder op, int max) {
            int n = r.nextInt(max + 1);
            op.append(' ').append(n);
            List<Object> l = new ArrayList<Object>();
            for (int i = 0; i < n; i++) {
                Integer v = value();
                appendValue(op, v);
                l.add(v);
            }
            return l;
        }

        /** Appends a map spec "H|L n k v ..." and returns the map it describes. */
        HashMap<Object, Object> sourceMap(StringBuilder op, int max) {
            boolean linked = r.nextBoolean();
            HashMap<Object, Object> m = linked ? new LinkedHashMap<Object, Object>() : new HashMap<Object, Object>();
            int[] ks = keys(max);
            op.append(' ').append(linked ? 'L' : 'H').append(' ').append(ks.length);
            for (int k : ks) {
                Integer v = value();
                op.append(' ').append(k);
                appendValue(op, v);
                m.put(pool[k], v);
            }
            return m;
        }

        int capacity() {
            if (family >= FAMILY_CKEY && r.nextBoolean()) {
                // Large enough for colliding keys to be treeified at once.
                int[] caps = {64, 65, 100, 128, 256};
                return caps[r.nextInt(caps.length)];
            }
            int[] caps = {0, 1, 2, 3, 5, 8, 12, 16, 17, 31, 64, 100, 128, 500, 1000};
            return caps[r.nextInt(caps.length)];
        }

        float loadFactor() {
            float[] lfs = {0.1f, 0.25f, 0.33f, 0.5f, 0.75f, 0.75f, 1f, 1.5f, 3f, 10f};
            return lfs[r.nextInt(lfs.length)];
        }

        void construct(StringBuilder hdr) {
            boolean setKind = container.equals("HS") || container.equals("LHS");
            boolean linked = container.startsWith("L");
            if (container.equals("LHMA")) {
                int cap = capacity();
                float lf = loadFactor();
                hdr.append("new_int_float_boolean ").append(cap).append(' ').append(floatHex(lf)).append(" true");
                map = new LinkedHashMap<Object, Object>(cap, lf, true);
                return;
            }
            switch (r.nextInt(5)) {
                case 0:
                    hdr.append("new");
                    if (setKind) {
                        set = linked ? new LinkedHashSet<Object>() : new HashSet<Object>();
                    } else {
                        map = linked ? new LinkedHashMap<Object, Object>() : new HashMap<Object, Object>();
                    }
                    return;
                case 1: {
                    int cap = capacity();
                    hdr.append("new_int ").append(cap);
                    if (setKind) {
                        set = linked ? new LinkedHashSet<Object>(cap) : new HashSet<Object>(cap);
                    } else {
                        map = linked ? new LinkedHashMap<Object, Object>(cap) : new HashMap<Object, Object>(cap);
                    }
                    return;
                }
                case 2:
                case 3: {
                    int cap = capacity();
                    float lf = loadFactor();
                    hdr.append("new_int_float ").append(cap).append(' ').append(floatHex(lf));
                    if (setKind) {
                        set = linked ? new LinkedHashSet<Object>(cap, lf) : new HashSet<Object>(cap, lf);
                    } else {
                        map = linked ? new LinkedHashMap<Object, Object>(cap, lf)
                                : new HashMap<Object, Object>(cap, lf);
                    }
                    return;
                }
                default:
                    if (setKind) {
                        hdr.append("new_coll");
                        List<Object> l = keyList(hdr, 2 * pool.length);
                        set = linked ? new LinkedHashSet<Object>(l) : new HashSet<Object>(l);
                    } else {
                        hdr.append("new_map");
                        HashMap<Object, Object> src = sourceMap(hdr, 2 * pool.length);
                        map = linked ? new LinkedHashMap<Object, Object>(src) : new HashMap<Object, Object>(src);
                    }
            }
        }

        /**
         * Appends "mutate k h" for a random key with a settable hash code
         * and sets it, or returns null if the pool has none.
         */
        String mutate(StringBuilder op) {
            int start = key();
            for (int j = 0; j < pool.length; j++) {
                int k = (start + j) % pool.length;
                Object o = pool[k];
                int h = r.nextInt(4) + 64 * r.nextInt(4);
                if (o instanceof CKey) {
                    ((CKey) o).hash = h;
                } else if (o instanceof NKey) {
                    ((NKey) o).hash = h;
                } else if (o instanceof XKey) {
                    ((XKey) o).hash = h;
                } else {
                    continue;
                }
                op.append("mutate ").append(k).append(' ').append(h);
                return "ok";
            }
            return null;
        }

        // ---- map operations

        String mapOp(StringBuilder op, int phase) {
            int t = r.nextInt(100);
            // Phase 0 grows the map, 1 churns, 2 shrinks, 3 churns.
            if (phase == 0 && t < 60 || phase != 2 && t < 25) {
                int k = key();
                Integer v = value();
                op.append("put ").append(k);
                appendValue(op, v);
                return vstr(map.put(pool[k], v));
            }
            if (phase == 2 && t < 50) {
                if (t < 10) {
                    op.append("iterRemove");
                    return iterRemove(op);
                }
                int k = key();
                op.append("remove ").append(k);
                return vstr(map.remove(pool[k]));
            }
            t = r.nextInt(46);
            switch (t) {
                case 0: {
                    int k = key();
                    Integer v = value();
                    op.append("putIfAbsent ").append(k);
                    appendValue(op, v);
                    return vstr(map.putIfAbsent(pool[k], v));
                }
                case 1:
                case 2: {
                    int k = key();
                    op.append("get ").append(k);
                    return vstr(map.get(pool[k]));
                }
                case 3: {
                    int k = key();
                    Integer v = value();
                    op.append("getOrDefault ").append(k);
                    appendValue(op, v);
                    return vstr(map.getOrDefault(pool[k], v));
                }
                case 4: {
                    int k = key();
                    op.append("containsKey ").append(k);
                    return bstr(map.containsKey(pool[k]));
                }
                case 5: {
                    Integer v = value();
                    op.append("containsValue");
                    appendValue(op, v);
                    return bstr(map.containsValue(v));
                }
                case 6:
                case 7: {
                    int k = key();
                    op.append("remove ").append(k);
                    return vstr(map.remove(pool[k]));
                }
                case 8: {
                    int k = key();
                    Integer v = value();
                    op.append("remove2 ").append(k);
                    appendValue(op, v);
                    return bstr(map.remove(pool[k], v));
                }
                case 9: {
                    int k = key();
                    Integer v = value();
                    op.append("replace2 ").append(k);
                    appendValue(op, v);
                    return vstr(map.replace(pool[k], v));
                }
                case 10: {
                    int k = key();
                    Integer o = value();
                    Integer v = value();
                    op.append("replace3 ").append(k);
                    appendValue(op, o);
                    appendValue(op, v);
                    return bstr(map.replace(pool[k], o, v));
                }
                case 11:
                    if (r.nextInt(4) != 0) {
                        return mapOp(op, phase);
                    }
                    op.append("clear");
                    map.clear();
                    return "ok";
                case 12: {
                    op.append("putAll");
                    HashMap<Object, Object> src = sourceMap(op, 20);
                    map.putAll(src);
                    return "ok";
                }
                case 13:
                case 14:
                    op.append("iterRemove");
                    return iterRemove(op);
                case 15: {
                    int k = key();
                    op.append("ksRemove ").append(k);
                    return bstr(map.keySet().remove(pool[k]));
                }
                case 16: {
                    Integer v = value();
                    op.append("vsRemove");
                    appendValue(op, v);
                    return bstr(map.values().remove(v));
                }
                case 17: {
                    int k = key();
                    Integer v = value();
                    op.append("esRemove ").append(k);
                    appendValue(op, v);
                    return bstr(map.entrySet().remove(new AbstractMap.SimpleEntry<Object, Object>(pool[k], v)));
                }
                case 18: {
                    int k = key();
                    Integer v = value();
                    op.append("esContains ").append(k);
                    appendValue(op, v);
                    return bstr(map.entrySet().contains(new AbstractMap.SimpleEntry<Object, Object>(pool[k], v)));
                }
                case 19: {
                    int k = key();
                    op.append("ksContains ").append(k);
                    return bstr(map.keySet().contains(pool[k]));
                }
                case 20: {
                    Integer v = value();
                    op.append("vsContains");
                    appendValue(op, v);
                    return bstr(map.values().contains(v));
                }
                case 21: {
                    op.append("ksRemoveAll");
                    List<Object> l = keyList(op, 2 * map.size() + 2);
                    return bstr(map.keySet().removeAll(l));
                }
                case 22: {
                    op.append("ksRetainAll");
                    List<Object> l = keyList(op, pool.length);
                    return bstr(map.keySet().retainAll(l));
                }
                case 23: {
                    op.append("vsRemoveAll");
                    List<Object> l = valueList(op, 4);
                    return bstr(map.values().removeAll(l));
                }
                case 24: {
                    op.append("vsRetainAll");
                    List<Object> l = valueList(op, 12);
                    return bstr(map.values().retainAll(l));
                }
                case 25: {
                    op.append("esRemoveAll");
                    List<Object> l = entryList(op);
                    return bstr(map.entrySet().removeAll(l));
                }
                case 26: {
                    int pos = r.nextInt(map.size() + 2);
                    Integer v = value();
                    op.append("esSetValue ").append(pos);
                    appendValue(op, v);
                    Iterator<Map.Entry<Object, Object>> it = map.entrySet().iterator();
                    for (int i = 0; i < pos && it.hasNext(); i++) {
                        it.next();
                    }
                    if (!it.hasNext()) {
                        return "none";
                    }
                    return vstr(it.next().setValue(v));
                }
                case 27:
                    op.append("clone");
                    map = cloneMap(map);
                    return map.getClass().getName();
                case 28: {
                    boolean linked = r.nextBoolean();
                    op.append("copy ").append(linked ? 'L' : 'H');
                    map = linked ? new LinkedHashMap<Object, Object>(map) : new HashMap<Object, Object>(map);
                    container = linked ? "LHM" : "HM";
                    return map.getClass().getName();
                }
                case 29: {
                    op.append("equalsCopy");
                    HashMap<Object, Object> c = new HashMap<Object, Object>(map);
                    boolean a = map.equals(c);
                    boolean b = c.equals(map);
                    boolean ks = map.keySet().equals(c.keySet());
                    boolean es = map.entrySet().equals(c.entrySet());
                    return new StringBuilder().append(bstr(a)).append('/').append(bstr(b)).append('/')
                            .append(bstr(ks)).append('/').append(bstr(es)).toString();
                }
                case 30: {
                    int k = key();
                    Integer v = value();
                    op.append("equalsMod ").append(k);
                    appendValue(op, v);
                    HashMap<Object, Object> c = new LinkedHashMap<Object, Object>(map);
                    c.put(pool[k], v);
                    return new StringBuilder().append(bstr(map.equals(c))).append('/').append(bstr(c.equals(map)))
                            .toString();
                }
                case 31:
                    op.append("hashCode");
                    return new StringBuilder().append(map.hashCode()).append('/').append(map.keySet().hashCode())
                            .append('/').append(map.entrySet().hashCode()).toString();
                case 32:
                    op.append("toString");
                    return new StringBuilder().append(digest(map.toString())).append('/')
                            .append(digest(map.keySet().toString())).append('/')
                            .append(digest(map.values().toString())).append('/')
                            .append(digest(map.entrySet().toString())).toString();
                case 33: {
                    int len = r.nextInt(map.size() + 3);
                    op.append("toArray ").append(len);
                    StringBuilder sb = new StringBuilder();
                    for (Object o : map.keySet().toArray()) {
                        tok(sb, o);
                        sb.append(',');
                    }
                    sb.append('|');
                    for (Object o : map.values().toArray(new Object[len])) {
                        if (o == null) {
                            sb.append('_');
                        } else {
                            vtok(sb, o);
                        }
                        sb.append(',');
                    }
                    sb.append('|');
                    for (Object o : map.entrySet().toArray()) {
                        Map.Entry<?, ?> e = (Map.Entry<?, ?>) o;
                        tok(sb, e.getKey());
                        sb.append('=');
                        vtok(sb, e.getValue());
                        sb.append(',');
                    }
                    return digest(sb);
                }
                case 34:
                    op.append("iterNSEE");
                    try {
                        Iterator<Object> it = map.values().iterator();
                        while (it.hasNext()) {
                            it.next();
                        }
                        it.next();
                        return "none";
                    } catch (RuntimeException e) {
                        return exToken(e);
                    }
                case 35: {
                    boolean first = r.nextBoolean();
                    op.append("iterISE ").append(bstr(first));
                    try {
                        Iterator<Object> it = map.keySet().iterator();
                        if (first && it.hasNext()) {
                            it.next();
                            it.remove();
                        }
                        it.remove();
                        return "none";
                    } catch (RuntimeException e) {
                        return exToken(e);
                    }
                }
                case 36: {
                    int k = key();
                    Integer v = value();
                    op.append("iterCME ").append(k);
                    appendValue(op, v);
                    Iterator<Map.Entry<Object, Object>> it = map.entrySet().iterator();
                    if (it.hasNext()) {
                        it.next();
                    }
                    String put = vstr(map.put(pool[k], v));
                    String res;
                    if (!it.hasNext()) {
                        res = "end";
                    } else {
                        try {
                            it.next();
                            res = "ok";
                        } catch (RuntimeException e) {
                            res = exToken(e);
                        }
                    }
                    return new StringBuilder(put).append('/').append(res).toString();
                }
                case 37: {
                    int k = key();
                    op.append("addView ").append(k);
                    try {
                        map.keySet().add(pool[k]);
                        return "none";
                    } catch (RuntimeException e) {
                        return exToken(e);
                    }
                }
                case 38: {
                    op.append("ksContainsAll");
                    List<Object> l = keyList(op, 6);
                    return bstr(map.keySet().containsAll(l));
                }
                case 39: {
                    String res = mutate(op);
                    return res != null ? res : mapOp(op, phase);
                }
                default: {
                    int k = key();
                    Integer v = value();
                    op.append("put ").append(k);
                    appendValue(op, v);
                    return vstr(map.put(pool[k], v));
                }
            }
        }

        List<Object> entryList(StringBuilder op) {
            int n = r.nextInt(2 * map.size() + 3);
            op.append(' ').append(n);
            List<Object> l = new ArrayList<Object>();
            for (int i = 0; i < n; i++) {
                int k = key();
                Integer v = value();
                op.append(' ').append(k);
                appendValue(op, v);
                l.add(new AbstractMap.SimpleEntry<Object, Object>(pool[k], v));
            }
            return l;
        }

        @SuppressWarnings("unchecked")
        static HashMap<Object, Object> cloneMap(HashMap<Object, Object> m) {
            return (HashMap<Object, Object>) m.clone();
        }

        /** Iterates over a view of the map or the set, removing some elements. */
        String iterRemove(StringBuilder op) {
            int mod = 1 + r.nextInt(4);
            int rem = r.nextInt(mod);
            Iterator<?> it;
            if (isMap()) {
                int view = r.nextInt(3);
                op.append(' ').append("KVE".charAt(view));
                it = view == 0 ? map.keySet().iterator() : view == 1 ? map.values().iterator()
                        : map.entrySet().iterator();
            } else {
                it = set.iterator();
            }
            op.append(' ').append(mod).append(' ').append(rem);
            int removed = 0;
            for (int i = 0; it.hasNext(); i++) {
                it.next();
                if (i % mod == rem) {
                    it.remove();
                    removed++;
                }
            }
            return Integer.toString(removed);
        }

        // ---- set operations

        String setOp(StringBuilder op, int phase) {
            int t = r.nextInt(100);
            if (phase == 0 && t < 60 || phase != 2 && t < 25) {
                int k = key();
                op.append("add ").append(k);
                return bstr(set.add(pool[k]));
            }
            if (phase == 2 && t < 50) {
                if (t < 10) {
                    op.append("iterRemove");
                    return iterRemove(op);
                }
                int k = key();
                op.append("remove ").append(k);
                return bstr(set.remove(pool[k]));
            }
            t = r.nextInt(23);
            switch (t) {
                case 0:
                case 1:
                case 2: {
                    int k = key();
                    op.append("contains ").append(k);
                    return bstr(set.contains(pool[k]));
                }
                case 3:
                case 4: {
                    int k = key();
                    op.append("remove ").append(k);
                    return bstr(set.remove(pool[k]));
                }
                case 5:
                    if (r.nextInt(4) != 0) {
                        return setOp(op, phase);
                    }
                    op.append("clear");
                    set.clear();
                    return "ok";
                case 6: {
                    op.append("addAll");
                    List<Object> l = keyList(op, 20);
                    return bstr(set.addAll(l));
                }
                case 7: {
                    op.append("removeAll");
                    List<Object> l = keyList(op, 2 * set.size() + 2);
                    return bstr(set.removeAll(l));
                }
                case 8: {
                    op.append("retainAll");
                    List<Object> l = keyList(op, pool.length);
                    return bstr(set.retainAll(l));
                }
                case 9: {
                    op.append("containsAll");
                    List<Object> l = keyList(op, 6);
                    return bstr(set.containsAll(l));
                }
                case 10:
                case 11:
                    op.append("iterRemove");
                    return iterRemove(op);
                case 12:
                    op.append("clone");
                    set = cloneSet(set);
                    return set.getClass().getName();
                case 13: {
                    boolean linked = r.nextBoolean();
                    op.append("copy ").append(linked ? 'L' : 'H');
                    set = linked ? new LinkedHashSet<Object>(set) : new HashSet<Object>(set);
                    container = linked ? "LHS" : "HS";
                    return set.getClass().getName();
                }
                case 14: {
                    op.append("equalsCopy");
                    HashSet<Object> c = new HashSet<Object>(set);
                    return new StringBuilder().append(bstr(set.equals(c))).append('/').append(bstr(c.equals(set)))
                            .toString();
                }
                case 15: {
                    int k = key();
                    op.append("equalsMod ").append(k);
                    HashSet<Object> c = new LinkedHashSet<Object>(set);
                    c.add(pool[k]);
                    return new StringBuilder().append(bstr(set.equals(c))).append('/').append(bstr(c.equals(set)))
                            .toString();
                }
                case 16:
                    op.append("hashCode");
                    return Integer.toString(set.hashCode());
                case 17:
                    op.append("toString");
                    return digest(set.toString());
                case 18: {
                    int len = r.nextInt(set.size() + 3);
                    op.append("toArray ").append(len);
                    StringBuilder sb = new StringBuilder();
                    for (Object o : set.toArray()) {
                        tok(sb, o);
                        sb.append(',');
                    }
                    sb.append('|');
                    for (Object o : set.toArray(new Object[len])) {
                        if (o == null) {
                            sb.append('_');
                        } else {
                            tok(sb, o);
                        }
                        sb.append(',');
                    }
                    return digest(sb);
                }
                case 19:
                    op.append("iterNSEE");
                    try {
                        Iterator<Object> it = set.iterator();
                        while (it.hasNext()) {
                            it.next();
                        }
                        it.next();
                        return "none";
                    } catch (RuntimeException e) {
                        return exToken(e);
                    }
                case 21: {
                    String res = mutate(op);
                    return res != null ? res : setOp(op, phase);
                }
                case 20: {
                    int k = key();
                    op.append("iterCME ").append(k);
                    Iterator<Object> it = set.iterator();
                    if (it.hasNext()) {
                        it.next();
                    }
                    String add = bstr(set.add(pool[k]));
                    String res;
                    if (!it.hasNext()) {
                        res = "end";
                    } else {
                        try {
                            it.next();
                            res = "ok";
                        } catch (RuntimeException e) {
                            res = exToken(e);
                        }
                    }
                    return new StringBuilder(add).append('/').append(res).toString();
                }
                default: {
                    int k = key();
                    op.append("add ").append(k);
                    return bstr(set.add(pool[k]));
                }
            }
        }

        @SuppressWarnings("unchecked")
        static HashSet<Object> cloneSet(HashSet<Object> s) {
            return (HashSet<Object>) s.clone();
        }

        static String vstr(Object v) {
            if (v == null) {
                return "null";
            }
            return Integer.toString((Integer) v);
        }

        static String bstr(boolean b) {
            return b ? "true" : "false";
        }

        // ---- order and state

        String order() {
            StringBuilder sb = new StringBuilder();
            if (isMap()) {
                for (Map.Entry<Object, Object> e : map.entrySet()) {
                    tok(sb, e.getKey());
                    sb.append('=');
                    vtok(sb, e.getValue());
                    sb.append(',');
                }
            } else {
                for (Object k : set) {
                    tok(sb, k);
                    sb.append(',');
                }
            }
            return sb.toString();
        }

        String state() {
            try {
                HashMap<?, ?> m = isMap() ? map : (HashMap<?, ?>) fSetMap.get(set);
                return mapState(m);
            } catch (IllegalAccessException e) {
                throw new RuntimeException(e);
            }
        }
    }

    /**
     * Describes the internal state of m: capacity, threshold and size, then
     * for each non-empty bin its index, the keys along its next links, and
     * for a tree bin the tree in preorder with node colours. "P!" marks
     * broken prev links in a tree bin and "L!" a broken LinkedHashMap list.
     */
    static String mapState(HashMap<?, ?> m) throws IllegalAccessException {
        StringBuilder sb = new StringBuilder();
        Object[] tab = (Object[]) fTable.get(m);
        sb.append('c').append(tab == null ? 0 : tab.length).append('t').append(fThreshold.getInt(m)).append('s')
                .append(fSize.getInt(m)).append(';');
        if (tab != null) {
            for (int i = 0; i < tab.length; i++) {
                Object first = tab[i];
                if (first == null) {
                    continue;
                }
                sb.append(i).append(':');
                boolean tree = treeNodeClass.isInstance(first);
                boolean prevOK = true;
                Object prev = null;
                for (Object e = first; e != null; e = fNext.get(e)) {
                    tok(sb, fKey.get(e));
                    sb.append(',');
                    if (tree && fPrev.get(e) != prev) {
                        prevOK = false;
                    }
                    prev = e;
                }
                if (tree) {
                    Object root = first;
                    while (fParent.get(root) != null) {
                        root = fParent.get(root);
                    }
                    sb.append('T');
                    tree(sb, root);
                    if (!prevOK) {
                        sb.append("P!");
                    }
                }
                sb.append(';');
            }
        }
        if (m instanceof LinkedHashMap) {
            Object before = null;
            boolean ok = true;
            Object e = fHead.get(m);
            for (; e != null; e = fAfter.get(e)) {
                if (fBefore.get(e) != before) {
                    ok = false;
                }
                before = e;
            }
            if (fTail.get(m) != before) {
                ok = false;
            }
            if (!ok) {
                sb.append("L!");
            }
        }
        return sb.toString();
    }

    static void tree(StringBuilder sb, Object n) throws IllegalAccessException {
        if (n == null) {
            sb.append('.');
            return;
        }
        sb.append('(');
        tok(sb, fKey.get(n));
        sb.append(fRed.getBoolean(n) ? 'r' : 'b');
        tree(sb, fLeft.get(n));
        tree(sb, fRight.get(n));
        sb.append(')');
    }
}
