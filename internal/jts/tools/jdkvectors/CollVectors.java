/*
 * CollVectors prints test vectors for the Go implementation of the
 * java.util collection classes in internal/jts/java (arraylist.go,
 * linkedlist.go, vector.go, arraydeque.go, priorityqueue.go, treemap.go,
 * enumset.go, collections_util.go, arrays.go and timsort.go). Regenerate the
 * vectors from the repository root with:
 *
 *   java internal/jts/tools/jdkvectors/CollVectors.java internal/jts/java/testdata
 *
 * which writes coll_*.txt in that directory. The Go test coll_test.go replays
 * them.
 *
 * Each file holds scenarios. A scenario starts with a line
 *
 *   new <kind> <arg>...
 *
 * that creates a collection, followed by operation lines
 *
 *   <op> <arg>... => <result> | <state> | <compare log>
 *
 * where <state> renders the collection after the operation (its iteration
 * order), or is "=" if that is unchanged from the previous line, and
 * <compare log> is "<count>:<hash>" of the calls made to
 * Item.compareTo and to the test comparators during the operation, which
 * pins down the exact sequence of comparisons the algorithms make.
 *
 * Values are rendered as: null, i<int> (Integer), s<chars> (String),
 * k<key>.<id> (Item, compared by key only), e<name> (enum Ord),
 * <k=v> (Map.Entry), [a,b] (Collection), {<k=v>,...} (Map), A[a,b] (Object[]),
 * and "!<SimpleName>" for an exception. Collection arguments are written as
 * <type>[a,b] with type AL (ArrayList), LL (LinkedList), TS (TreeSet),
 * TSmod (TreeSet with the mod comparator) and EM (EnumSet); map arguments
 * as TM[k=v,...] (TreeMap), TMmod[...] and UM[...] (an unmodifiable TreeMap,
 * which is not a SortedMap). Comparators are nat (null), rev
 * (Collections.reverseOrder()), mod (by key mod 3, with many ties), chaos
 * (an inconsistent comparator), lognat (natural order through a comparator)
 * and revmod (Collections.reverseOrder(mod)).
 *
 * coll_static.txt holds operation lines without scenarios (state "null") for
 * Arrays, Collections and TimSort. Primitive arrays are written as D[...]
 * and F[...] (raw bits in hex), I, J, S, C, B (decimal) and Z (0 or 1).
 * The large inputs of the sortgen lines are given by a seed for the
 * generator in CollStatic.lcgNext rather than listed.
 */

import java.util.*;

public class CollVectors {
    static Random rnd;
    static StringBuilder out;
    static int cmpCount;
    static int cmpHash;

    static void logCmp(int a, int b) {
        cmpCount++;
        cmpHash = cmpHash * 1000003 + (a * 1009 + b);
    }

    enum Ord { X, Y, Z, M, A, B }

    static final class Item implements Comparable<Item> {
        final int key, id;

        Item(int key, int id) {
            this.key = key;
            this.id = id;
        }

        public int compareTo(Item o) {
            logCmp(key, o.key);
            return Integer.compare(key, o.key);
        }

        public boolean equals(Object o) {
            return o instanceof Item it && it.key == key && it.id == id;
        }

        public int hashCode() {
            return key * 31 + id;
        }

        public String toString() {
            return "k" + key + "." + id;
        }
    }

    static int keyOf(Object o) {
        if (o == null) throw new NullPointerException();
        if (o instanceof Integer i) return i;
        if (o instanceof Item it) return it.key;
        if (o instanceof String s) return s.hashCode();
        if (o instanceof Ord e) return e.ordinal();
        throw new ClassCastException();
    }

    static final class ModCmp implements Comparator<Object> {
        public int compare(Object a, Object b) {
            int x = keyOf(a), y = keyOf(b);
            logCmp(x, y);
            return Integer.compare(Math.floorMod(x, 3), Math.floorMod(y, 3));
        }
    }

    static final class ChaosCmp implements Comparator<Object> {
        public int compare(Object a, Object b) {
            int x = keyOf(a), y = keyOf(b);
            logCmp(x, y);
            int h = x * 0x9E3779B1 + y * 0x85EBCA6B;
            h ^= h >>> 15;
            h *= 0x2C1B3C6D;
            h ^= h >>> 12;
            return (h & 3) - 1;
        }
    }

    static final class LogNatCmp implements Comparator<Object> {
        @SuppressWarnings({"unchecked", "rawtypes"})
        public int compare(Object a, Object b) {
            logCmp(keyOf(a), keyOf(b));
            return ((Comparable) a).compareTo(b);
        }
    }

    static final ModCmp MOD = new ModCmp();

    static Comparator<Object> cmp(String name) {
        switch (name) {
            case "nat": return null;
            case "rev": return Collections.reverseOrder();
            case "mod": return MOD;
            case "chaos": return new ChaosCmp();
            case "lognat": return new LogNatCmp();
            case "revmod": return Collections.reverseOrder(MOD);
        }
        throw new IllegalArgumentException(name);
    }

    // ---------------------------------------------------------------------
    // Rendering

    static String r(Object o) {
        if (o == null) return "null";
        if (o instanceof Integer i) return "i" + i;
        if (o instanceof String s) return "s" + s;
        if (o instanceof Item it) return it.toString();
        if (o instanceof Boolean b) return b.toString();
        if (o instanceof Ord e) return "e" + e.name();
        if (o instanceof Map.Entry<?, ?> e) return "<" + r(e.getKey()) + "=" + r(e.getValue()) + ">";
        if (o instanceof Map<?, ?> m) {
            StringJoiner j = new StringJoiner(",", "{", "}");
            for (Map.Entry<?, ?> e : m.entrySet()) j.add(r(e));
            return j.toString();
        }
        if (o instanceof Collection<?> c) {
            StringJoiner j = new StringJoiner(",", "[", "]");
            for (Object e : c) j.add(r(e));
            return j.toString();
        }
        if (o instanceof Object[] a) {
            StringJoiner j = new StringJoiner(",", "A[", "]");
            for (Object e : a) j.add(r(e));
            return j.toString();
        }
        throw new IllegalArgumentException(o.getClass().toString());
    }

    static String ex(Throwable t) {
        return "!" + t.getClass().getSimpleName();
    }

    // ---------------------------------------------------------------------
    // Value generation

    static int elemType; // 0 Integer, 1 Item, 2 String, 3 Ord
    static boolean nulls;

    static Object v() {
        if (nulls && rnd.nextInt(60) == 0) return null;
        switch (elemType) {
            case 0: return rnd.nextInt(22) - 2;
            case 1: return new Item(rnd.nextInt(8), rnd.nextInt(6));
            case 2: {
                String s = "" + (char) ('a' + rnd.nextInt(5));
                if (rnd.nextBoolean()) s += (char) ('a' + rnd.nextInt(3));
                return s;
            }
            default: return Ord.values()[rnd.nextInt(Ord.values().length)];
        }
    }

    static Object mapValue() {
        if (rnd.nextInt(30) == 0) return null;
        return rnd.nextInt(100);
    }

    static List<Object> vals(int max) {
        int n = rnd.nextInt(max + 1);
        List<Object> l = new ArrayList<>();
        for (int i = 0; i < n; i++) l.add(v());
        return l;
    }

    static String lit(List<Object> l) {
        StringJoiner j = new StringJoiner(",", "[", "]");
        for (Object o : l) j.add(r(o));
        return j.toString();
    }

    // collArg returns a random collection argument and appends its literal.
    static Collection<Object> collArg(List<String> args, boolean sortedOK) {
        int savedCount = cmpCount, savedHash = cmpHash;
        try {
            return collArg0(args, sortedOK);
        } finally {
            cmpCount = savedCount;
            cmpHash = savedHash;
        }
    }

    static Collection<Object> collArg0(List<String> args, boolean sortedOK) {
        List<Object> l = vals(6);
        int t = rnd.nextInt(sortedOK ? 5 : 2);
        if (elemType == 3 && t >= 2) t = 4;
        Collection<Object> c;
        String tag;
        try {
            switch (t) {
                case 0: c = new ArrayList<>(l); tag = "AL"; break;
                case 1: c = new LinkedList<>(l); tag = "LL"; break;
                case 2: c = new TreeSet<>(l); tag = "TS"; break;
                case 3: c = new TreeSet<>(MOD); c.addAll(l); tag = "TSmod"; break;
                default: {
                    if (elemType == 3) {
                        EnumSet<Ord> es = EnumSet.noneOf(Ord.class);
                        for (Object o : l) if (o != null) es.add((Ord) o);
                        @SuppressWarnings("unchecked") Collection<Object> cc = (Collection<Object>) (Collection<?>) es;
                        c = cc;
                        tag = "EM";
                    } else {
                        c = new ArrayList<>(l);
                        tag = "AL";
                    }
                }
            }
        } catch (RuntimeException e) {
            // A TreeSet cannot hold null; fall back to a list.
            c = new ArrayList<>(l);
            tag = "AL";
        }
        // The literal lists the elements in insertion order, so that a sorted
        // set built from it has the same tree.
        args.add(tag + lit(l));
        return c;
    }

    // mapArg returns a random map argument and appends its literal, which
    // lists the puts that built it in order.
    static Map<Object, Object> mapArg(List<String> args) {
        int savedCount = cmpCount, savedHash = cmpHash;
        int n = rnd.nextInt(6);
        int t = rnd.nextInt(3);
        TreeMap<Object, Object> m = t == 1 ? new TreeMap<>(MOD) : new TreeMap<>();
        StringJoiner j = new StringJoiner(",", "[", "]");
        for (int i = 0; i < n; i++) {
            Object k = v();
            if (k == null) continue;
            Object val = mapValue();
            j.add(r(k) + "=" + r(val));
            m.put(k, val);
        }
        String tag = t == 0 ? "TM" : t == 1 ? "TMmod" : "UM";
        args.add(tag + j);
        cmpCount = savedCount;
        cmpHash = savedHash;
        return t == 2 ? Collections.unmodifiableMap(m) : m;
    }

    // ---------------------------------------------------------------------
    // Scenario driver

    interface Op {
        Object run(List<String> args) throws Throwable;
    }

    static Object state;
    // lastState is the rendered state of the previous operation line; an
    // unchanged state is written as "=".
    static String lastState;
    static String prefix = "";
    static boolean identityHash;
    static final Object OK = new Object();

    static void emit(String name, Op op) {
        List<String> args = new ArrayList<>();
        cmpCount = 0;
        cmpHash = 0;
        String res;
        try {
            Object o = op.run(args);
            res = o == OK ? "ok" : o instanceof String s ? s : r(o);
        } catch (Throwable t) {
            res = ex(t);
        }
        int cc = cmpCount, ch = cmpHash;
        StringBuilder sb = new StringBuilder(prefix).append(name);
        for (String a : args) sb.append(' ').append(a);
        String st = r(state);
        sb.append(" => ").append(res).append(" | ").append(st.equals(lastState) ? "=" : st).append(" | ").append(cc).append(':').append(ch);
        lastState = st;
        out.append(sb).append('\n');
    }

    static StringBuilder newLine() {
        lastState = null;
        return out;
    }

    static String bool(boolean b) {
        return b ? "true" : "false";
    }

    static String num(int n) {
        return Integer.toString(n);
    }

    static String toStr(Object o) {
        return o.toString().replace(' ', '_');
    }

    // ---------------------------------------------------------------------
    // Iterator programs. Commands are separated by '/'.

    @SuppressWarnings({"unchecked", "rawtypes"})
    static String iterProg(Iterator it, String prog, Object owner) {
        StringJoiner j = new StringJoiner("/");
        Object last = null;
        for (String c : prog.split("/")) {
            try {
                switch (c.charAt(0)) {
                    case 'h': j.add(bool(it.hasNext())); break;
                    case 'n': last = it.next(); j.add(r(last)); break;
                    case 'r': it.remove(); j.add("ok"); break;
                    case 'p': last = ((ListIterator) it).previous(); j.add(r(last)); break;
                    case 'H': j.add(bool(((ListIterator) it).hasPrevious())); break;
                    case 'i': j.add(num(((ListIterator) it).nextIndex())); break;
                    case 'I': j.add(num(((ListIterator) it).previousIndex())); break;
                    case 'a': ((ListIterator) it).add(parse(c.substring(1))); j.add("ok"); break;
                    case 's': ((ListIterator) it).set(parse(c.substring(1))); j.add("ok"); break;
                    case 'v': j.add(r(((Map.Entry) last).setValue(parse(c.substring(1))))); break;
                    case 'x': j.add(external(owner, c.substring(1))); break;
                    default: throw new IllegalStateException(c);
                }
            } catch (IllegalStateException | NoSuchElementException | ConcurrentModificationException
                     | UnsupportedOperationException | IndexOutOfBoundsException | NullPointerException
                     | ClassCastException e) {
                j.add(ex(e));
            }
        }
        return j.toString();
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static String external(Object owner, String c) {
        if (owner instanceof Map m) {
            if (c.charAt(0) == 'p') {
                String[] kv = c.substring(1).split("=");
                return r(m.put(parse(kv[0]), parse(kv[1])));
            }
            return r(m.remove(parse(c.substring(1))));
        }
        Collection coll = (Collection) owner;
        switch (c.charAt(0)) {
            case 'a': return bool(coll.add(parse(c.substring(1))));
            case 'r': return bool(coll.remove(parse(c.substring(1))));
            case 'c': coll.clear(); return "ok";
        }
        throw new IllegalStateException(c);
    }

    static Object parse(String s) {
        if (s.equals("null")) return null;
        switch (s.charAt(0)) {
            case 'i': return Integer.parseInt(s.substring(1));
            case 's': return s.substring(1);
            case 'e': return Ord.valueOf(s.substring(1));
            case 'k': {
                String[] p = s.substring(1).split("\\.");
                return new Item(Integer.parseInt(p[0]), Integer.parseInt(p[1]));
            }
        }
        throw new IllegalArgumentException(s);
    }

    // randomIterProg makes a program for an iterator over about n elements.
    static String randomIterProg(int n, boolean listIter, boolean entries, boolean external) {
        StringJoiner j = new StringJoiner("/");
        int len = n + 2 + rnd.nextInt(4);
        for (int i = 0; i < len; i++) {
            int x = rnd.nextInt(100);
            if (external && x < 3) {
                if (entries) j.add(rnd.nextBoolean() ? "xp" + r(v()) + "=" + r(mapValue()) : "xm" + r(v()));
                else j.add(rnd.nextBoolean() ? "xa" + r(v()) : "xr" + r(v()));
            } else if (x < 6) j.add("h");
            else if (x < 30 && !listIter) j.add("r");
            else if (x < 22) j.add("r");
            else if (listIter && x < 34) j.add("p");
            else if (listIter && x < 37) j.add("H");
            else if (listIter && x < 40) j.add(rnd.nextBoolean() ? "i" : "I");
            else if (listIter && x < 46) j.add("a" + r(v()));
            else if (listIter && x < 52) j.add("s" + r(v()));
            else if (entries && x < 40) j.add("v" + r(mapValue()));
            else j.add("n");
        }
        return j.toString();
    }

    // ---------------------------------------------------------------------
    // Collection operations shared by all collections.

    @SuppressWarnings({"unchecked", "rawtypes"})
    static boolean collectionOp(Collection c, int x) {
        switch (x) {
            case 0: emit("add", a -> { Object e = v(); a.add(r(e)); return bool(c.add(e)); }); return true;
            case 1: emit("remove", a -> { Object e = v(); a.add(r(e)); return bool(c.remove(e)); }); return true;
            case 2: emit("contains", a -> { Object e = v(); a.add(r(e)); return bool(c.contains(e)); }); return true;
            case 3: emit("size", a -> num(c.size())); return true;
            case 4: emit("isEmpty", a -> bool(c.isEmpty())); return true;
            case 5: emit("addAll", a -> bool(c.addAll(collArg(a, true)))); return true;
            case 6: emit("removeAll", a -> bool(c.removeAll(collArg(a, true)))); return true;
            case 7: emit("retainAll", a -> bool(c.retainAll(collArg(a, true)))); return true;
            case 8: emit("containsAll", a -> bool(c.containsAll(collArg(a, true)))); return true;
            case 9: emit("toArray", a -> r(c.toArray())); return true;
            case 10: emit("toArrayT", a -> {
                int n = rnd.nextInt(8);
                a.add(num(n));
                Object[] arr = new Object[n];
                Arrays.fill(arr, -1);
                Object[] res = c.toArray(arr);
                return (res == arr ? "same:" : "new:") + r(res);
            }); return true;
            case 11: emit("iter", a -> {
                String p = randomIterProg(c.size(), false, false, rnd.nextInt(8) == 0);
                a.add(p);
                return iterProg(c.iterator(), p, c);
            }); return true;
            case 12: emit("toString", a -> toStr(c)); return true;
            case 13:
                if (identityHash) emit("toString", a -> toStr(c));
                else emit("hashCode", a -> num(c.hashCode()));
                return true;
            case 14: if (rnd.nextInt(6) == 0) emit("clear", a -> { c.clear(); return OK; }); return true;
            case 15: emit("equals", a -> bool(c.equals(collArg(a, true)))); return true;
            case 16: emit("cme", a -> {
                // Modifying a collection while iterating over it.
                Object e = v();
                a.add(r(e));
                Iterator it = c.iterator();
                StringJoiner j = new StringJoiner("/");
                try { j.add(r(it.next())); } catch (RuntimeException t) { j.add(ex(t)); }
                try { j.add(bool(c.add(e))); } catch (RuntimeException t) { j.add(ex(t)); }
                try { j.add(bool(it.hasNext())); j.add(r(it.next())); } catch (RuntimeException t) { j.add(ex(t)); }
                return j.toString();
            }); return true;
        }
        return false;
    }

    static final int COLLECTION_OPS = 17;

    @SuppressWarnings({"unchecked", "rawtypes"})
    static boolean listOp(List l, int x) {
        switch (x) {
            case 0: emit("addAt", a -> { int i = idx(l, 1); Object e = v(); a.add(num(i)); a.add(r(e)); l.add(i, e); return OK; }); return true;
            case 1: emit("get", a -> { int i = idx(l, 0); a.add(num(i)); return r(l.get(i)); }); return true;
            case 2: emit("set", a -> { int i = idx(l, 0); Object e = v(); a.add(num(i)); a.add(r(e)); return r(l.set(i, e)); }); return true;
            case 3: emit("removeAt", a -> { int i = idx(l, 0); a.add(num(i)); return r(l.remove(i)); }); return true;
            case 4: emit("indexOf", a -> { Object e = v(); a.add(r(e)); return num(l.indexOf(e)); }); return true;
            case 5: emit("lastIndexOf", a -> { Object e = v(); a.add(r(e)); return num(l.lastIndexOf(e)); }); return true;
            case 6: emit("addAllAt", a -> { int i = idx(l, 1); a.add(num(i)); return bool(l.addAll(i, collArg(a, true))); }); return true;
            case 7: emit("sort", a -> {
                String[] names = {"nat", "rev", "mod", "lognat", "revmod", "chaos"};
                String n = names[rnd.nextInt(names.length)];
                a.add(n);
                l.sort(cmp(n));
                return OK;
            }); return true;
            case 8: emit("lit", a -> {
                int i = rnd.nextInt(4) == 0 ? -1 : idx(l, 1);
                String p = randomIterProg(l.size(), true, false, rnd.nextInt(8) == 0);
                a.add(num(i));
                a.add(p);
                return iterProg(i < 0 ? l.listIterator() : l.listIterator(i), p, l);
            }); return true;
            case 9: emit("sub", a -> {
                int f = idx(l, 1);
                int t = idx(l, 1);
                if (rnd.nextInt(8) != 0 && f > t) { int tmp = f; f = t; t = tmp; }
                String p = randomSubProg();
                a.add(num(f));
                a.add(num(t));
                a.add(p);
                return subProg(l, f, t, p);
            }); return true;
            case 10: if (rnd.nextInt(3) == 0) emit("clone", a -> {
                Object c;
                if (l instanceof ArrayList al) c = al.clone();
                else if (l instanceof LinkedList ll) c = ll.clone();
                else if (l instanceof Vector vv) c = vv.clone();
                else return "noclone";
                return c.getClass().getSimpleName() + ":" + r(c);
            }); return true;
            case 11: if (l instanceof ArrayList al) emit("ensureCME", a -> {
                // ensureCapacity and trimToSize count as modifications only
                // in some states of the backing array.
                int n = rnd.nextInt(30) - 1;
                a.add(num(n));
                Iterator it = al.iterator();
                if (n < 0) al.trimToSize(); else al.ensureCapacity(n);
                try {
                    it.next();
                    return "ok";
                } catch (RuntimeException e) {
                    return ex(e);
                }
            }); return true;
        }
        return false;
    }

    static final int LIST_OPS = 12;

    static int idx(List<?> l, int extra) {
        int n = l.size() + extra;
        int i = rnd.nextInt(n + 2) - 1;
        return i;
    }

    static String subLit() {
        List<Object> l = vals(4);
        return "AL" + lit(l);
    }

    static String randomSubProg() {
        StringJoiner j = new StringJoiner("/");
        int len = 1 + rnd.nextInt(8);
        for (int i = 0; i < len; i++) {
            int x = rnd.nextInt(27);
            switch (x) {
                case 17: j.add("U" + subLit()); break;
                case 18: j.add("V" + (rnd.nextInt(5) - 1) + ";" + subLit()); break;
                case 19: j.add("D" + subLit()); break;
                case 20: j.add("K" + subLit()); break;
                case 21: j.add("E" + subLit()); break;
                case 22: j.add("C" + r(v())); break;
                case 23: j.add("Y" + subLit()); break;
                case 24: j.add("T"); break;
                case 25: j.add("L" + (rnd.nextInt(5) - 1) + ":" + randomIterProg(4, true, false, false).replace('/', '~')); break;
                case 26: j.add("Q" + rnd.nextInt(6)); break;
                case 0: j.add("a" + r(v())); break;
                case 1: j.add("A" + (rnd.nextInt(5) - 1) + "," + r(v())); break;
                case 2: j.add("r" + (rnd.nextInt(5) - 1)); break;
                case 3: j.add("R" + r(v())); break;
                case 4: j.add("g" + (rnd.nextInt(5) - 1)); break;
                case 5: j.add("s" + (rnd.nextInt(5) - 1) + "," + r(v())); break;
                case 6: if (rnd.nextInt(4) == 0) j.add("c"); else j.add("z"); break;
                case 7: j.add("z"); break;
                case 8: j.add("t"); break;
                case 9: j.add("o"); break;
                case 10: j.add("i" + r(v())); break;
                case 11: j.add("l" + r(v())); break;
                case 12: j.add("h"); break;
                case 13: j.add("n" + rnd.nextInt(3) + "," + rnd.nextInt(4)); break;
                case 14: if (rnd.nextInt(3) == 0) j.add("xa" + r(v())); else j.add("t"); break;
                case 15: if (rnd.nextInt(3) == 0) j.add("xr" + rnd.nextInt(3)); else j.add("z"); break;
                default: j.add("w"); break;
            }
        }
        return j.toString();
    }

    static List<Object> subArg(String lit) {
        List<Object> out = new ArrayList<>();
        String body = lit.substring(3, lit.length() - 1);
        if (!body.isEmpty()) for (String e : body.split(",")) out.add(parse(e));
        return out;
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static String subProg(List l, int f, int t, String prog) {
        List s = l.subList(f, t);
        StringJoiner j = new StringJoiner("/");
        for (String c : prog.split("/")) {
            try {
                String rest = c.substring(1);
                switch (c.charAt(0)) {
                    case 'a': j.add(bool(s.add(parse(rest)))); break;
                    case 'A': { String[] p = rest.split(","); s.add(Integer.parseInt(p[0]), parse(p[1])); j.add("ok"); break; }
                    case 'r': j.add(r(s.remove(Integer.parseInt(rest)))); break;
                    case 'R': j.add(bool(s.remove(parse(rest)))); break;
                    case 'g': j.add(r(s.get(Integer.parseInt(rest)))); break;
                    case 's': { String[] p = rest.split(","); j.add(r(s.set(Integer.parseInt(p[0]), parse(p[1])))); break; }
                    case 'c': s.clear(); j.add("ok"); break;
                    case 'z': j.add(num(s.size())); break;
                    case 't': j.add(r(s)); break;
                    case 'o': s.sort(null); j.add("ok"); break;
                    case 'i': j.add(num(s.indexOf(parse(rest)))); break;
                    case 'l': j.add(num(s.lastIndexOf(parse(rest)))); break;
                    case 'h': j.add(num(s.hashCode())); break;
                    case 'n': { String[] p = rest.split(","); s = s.subList(Integer.parseInt(p[0]), Integer.parseInt(p[1])); j.add("ok"); break; }
                    case 'x':
                        if (rest.charAt(0) == 'a') j.add(bool(l.add(parse(rest.substring(1)))));
                        else j.add(r(l.remove(Integer.parseInt(rest.substring(1)))));
                        break;
                    case 'U': j.add(bool(s.addAll(subArg(rest)))); break;
                    case 'V': { String[] p = rest.split(";"); j.add(bool(s.addAll(Integer.parseInt(p[0]), subArg(p[1])))); break; }
                    case 'D': j.add(bool(s.removeAll(subArg(rest)))); break;
                    case 'K': j.add(bool(s.retainAll(subArg(rest)))); break;
                    case 'E': j.add(bool(s.equals(subArg(rest)))); break;
                    case 'C': j.add(bool(s.contains(parse(rest)))); break;
                    case 'Y': j.add(bool(s.containsAll(subArg(rest)))); break;
                    case 'T': j.add(toStr(s)); break;
                    case 'L': {
                        String[] p = rest.split(":", 2);
                        int i = Integer.parseInt(p[0]);
                        j.add(iterProg(s.listIterator(i), p[1].replace('~', '/'), s).replace('/', '~'));
                        break;
                    }
                    case 'Q': {
                        Object[] arr = new Object[Integer.parseInt(rest)];
                        Arrays.fill(arr, -1);
                        Object[] res = s.toArray(arr);
                        j.add((res == arr ? "same:" : "new:") + r(res));
                        break;
                    }
                    case 'w': {
                        StringJoiner w = new StringJoiner(",");
                        Iterator it = s.iterator();
                        for (int k = 0; it.hasNext(); k++) {
                            w.add(r(it.next()));
                            if (k % 2 == 1) it.remove();
                        }
                        j.add(w.toString());
                        break;
                    }
                    default: throw new IllegalStateException(c);
                }
            } catch (RuntimeException e) {
                j.add(ex(e));
            }
        }
        return j.toString();
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static boolean dequeOp(Deque d, int x) {
        switch (x) {
            case 0: emit("addFirst", a -> { Object e = v(); a.add(r(e)); d.addFirst(e); return OK; }); return true;
            case 1: emit("addLast", a -> { Object e = v(); a.add(r(e)); d.addLast(e); return OK; }); return true;
            case 2: emit("offerFirst", a -> { Object e = v(); a.add(r(e)); return bool(d.offerFirst(e)); }); return true;
            case 3: emit("offerLast", a -> { Object e = v(); a.add(r(e)); return bool(d.offerLast(e)); }); return true;
            case 4: emit("removeFirst", a -> r(d.removeFirst())); return true;
            case 5: emit("removeLast", a -> r(d.removeLast())); return true;
            case 6: emit("pollFirst", a -> r(d.pollFirst())); return true;
            case 7: emit("pollLast", a -> r(d.pollLast())); return true;
            case 8: emit("getFirst", a -> r(d.getFirst())); return true;
            case 9: emit("getLast", a -> r(d.getLast())); return true;
            case 10: emit("peekFirst", a -> r(d.peekFirst())); return true;
            case 11: emit("peekLast", a -> r(d.peekLast())); return true;
            case 12: emit("push", a -> { Object e = v(); a.add(r(e)); d.push(e); return OK; }); return true;
            case 13: emit("pop", a -> r(d.pop())); return true;
            case 14: emit("peek", a -> r(d.peek())); return true;
            case 15: emit("poll", a -> r(d.poll())); return true;
            case 16: emit("element", a -> r(d.element())); return true;
            case 17: emit("offer", a -> { Object e = v(); a.add(r(e)); return bool(d.offer(e)); }); return true;
            case 18: emit("removeHead", a -> r(d.remove())); return true;
            case 19: emit("removeFirstOcc", a -> { Object e = v(); a.add(r(e)); return bool(d.removeFirstOccurrence(e)); }); return true;
            case 20: emit("removeLastOcc", a -> { Object e = v(); a.add(r(e)); return bool(d.removeLastOccurrence(e)); }); return true;
            case 21: emit("desc", a -> {
                String p = randomIterProg(d.size(), false, false, rnd.nextInt(8) == 0);
                a.add(p);
                return iterProg(d.descendingIterator(), p, d);
            }); return true;
            case 22: if (d instanceof ArrayDeque ad) emit("clone", a -> "ArrayDeque:" + r(ad.clone())); return true;
        }
        return false;
    }

    static final int DEQUE_OPS = 23;

    @SuppressWarnings({"unchecked", "rawtypes"})
    static boolean vectorOp(Vector vec, int x) {
        switch (x) {
            case 0: emit("addElement", a -> { Object e = v(); a.add(r(e)); vec.addElement(e); return OK; }); return true;
            case 1: emit("elementAt", a -> { int i = idx(vec, 0); a.add(num(i)); return r(vec.elementAt(i)); }); return true;
            case 2: emit("firstElement", a -> r(vec.firstElement())); return true;
            case 3: emit("lastElement", a -> r(vec.lastElement())); return true;
            case 4: emit("insertElementAt", a -> { Object e = v(); int i = idx(vec, 1); a.add(r(e)); a.add(num(i)); vec.insertElementAt(e, i); return OK; }); return true;
            case 5: emit("removeElementAt", a -> { int i = idx(vec, 0); a.add(num(i)); vec.removeElementAt(i); return OK; }); return true;
            case 6: emit("setElementAt", a -> { Object e = v(); int i = idx(vec, 0); a.add(r(e)); a.add(num(i)); vec.setElementAt(e, i); return OK; }); return true;
            case 7: emit("removeElement", a -> { Object e = v(); a.add(r(e)); return bool(vec.removeElement(e)); }); return true;
            case 8: emit("setSize", a -> { int n = rnd.nextInt(vec.size() + 4) - 1; a.add(num(n)); vec.setSize(n); return OK; }); return true;
            case 9: emit("indexOfFrom", a -> { Object e = v(); int i = idx(vec, 1); a.add(r(e)); a.add(num(i)); return num(vec.indexOf(e, i)); }); return true;
            case 10: emit("lastIndexOfFrom", a -> { Object e = v(); int i = idx(vec, 1); a.add(r(e)); a.add(num(i)); return num(vec.lastIndexOf(e, i)); }); return true;
            case 11: emit("capacity", a -> num(vec.capacity())); return true;
            case 12: emit("ensureCapacity", a -> { int n = rnd.nextInt(40) - 2; a.add(num(n)); vec.ensureCapacity(n); return num(vec.capacity()); }); return true;
            case 13: emit("trimToSize", a -> { vec.trimToSize(); return num(vec.capacity()); }); return true;
            case 14: if (rnd.nextInt(5) == 0) emit("removeAllElements", a -> { vec.removeAllElements(); return OK; }); else emit("copyInto", a -> {
                int n = rnd.nextInt(vec.size() + 3);
                a.add(num(n));
                Object[] arr = new Object[n];
                vec.copyInto(arr);
                return r(arr);
            }); return true;
        }
        if (vec instanceof Stack st) {
            switch (x) {
                case 15: emit("stackPush", a -> { Object e = v(); a.add(r(e)); return r(st.push(e)); }); return true;
                case 16: emit("stackPop", a -> r(st.pop())); return true;
                case 17: emit("stackPeek", a -> r(st.peek())); return true;
                case 18: emit("empty", a -> bool(st.empty())); return true;
                case 19: emit("search", a -> { Object e = v(); a.add(r(e)); return num(st.search(e)); }); return true;
            }
        }
        return false;
    }

    static final int VECTOR_OPS = 20;

    @SuppressWarnings({"unchecked", "rawtypes"})
    static boolean pqOp(PriorityQueue q, int x) {
        switch (x) {
            case 0: emit("offer", a -> { Object e = v(); a.add(r(e)); return bool(q.offer(e)); }); return true;
            case 1: emit("poll", a -> r(q.poll())); return true;
            case 2: emit("peek", a -> r(q.peek())); return true;
            case 3: emit("element", a -> r(q.element())); return true;
            case 4: emit("removeHead", a -> r(q.remove())); return true;
            case 5: emit("poll", a -> r(q.poll())); return true;
            case 6: emit("add", a -> { Object e = v(); a.add(r(e)); return bool(q.add(e)); }); return true;
            case 7: emit("drain", a -> {
                // Removing most elements while iterating makes the iterator
                // return, and remove, elements that removeAt moved before it.
                StringJoiner j = new StringJoiner("/");
                for (int i = 0; i < q.size() + 4; i++) {
                    j.add("n");
                    if (rnd.nextInt(10) < 7) j.add("r");
                }
                String p = j.toString();
                a.add(p);
                return iterProg(q.iterator(), p, q);
            }); return true;
        }
        return false;
    }

    static final int PQ_OPS = 8;

    @SuppressWarnings({"unchecked", "rawtypes"})
    static boolean navSetOp(NavigableSet s, int x) {
        switch (x) {
            case 0: emit("first", a -> r(s.first())); return true;
            case 1: emit("last", a -> r(s.last())); return true;
            case 2: emit("lower", a -> { Object e = v(); a.add(r(e)); return r(s.lower(e)); }); return true;
            case 3: emit("floor", a -> { Object e = v(); a.add(r(e)); return r(s.floor(e)); }); return true;
            case 4: emit("ceiling", a -> { Object e = v(); a.add(r(e)); return r(s.ceiling(e)); }); return true;
            case 5: emit("higher", a -> { Object e = v(); a.add(r(e)); return r(s.higher(e)); }); return true;
            case 6: emit("pollFirst", a -> r(s.pollFirst())); return true;
            case 7: emit("pollLast", a -> r(s.pollLast())); return true;
            case 8: emit("desc", a -> {
                String p = randomIterProg(s.size(), false, false, false);
                a.add(p);
                return iterProg(s.descendingIterator(), p, s);
            }); return true;
            case 9: if (s instanceof TreeSet ts) emit("clone", a -> {
                TreeSet c = (TreeSet) ts.clone();
                return "TreeSet:" + r(c);
            }); return true;
        }
        return false;
    }

    static final int NAVSET_OPS = 10;

    // ---------------------------------------------------------------------
    // Scenarios

    static String elemName() {
        return new String[]{"int", "item", "str", "ord"}[elemType];
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void lists() {
        String[] kinds = {"ArrayList", "LinkedList", "Vector", "Stack"};
        for (int sc = 0; sc < 124; sc++) {
            String kind = kinds[sc % kinds.length];
            elemType = rnd.nextInt(3);
            nulls = rnd.nextInt(3) == 0;
            List l;
            String ctor = "new";
            List<String> cargs = new ArrayList<>();
            int ct = rnd.nextInt(3);
            List<Object> init = vals(8);
            if (ct == 1 && !kind.equals("Stack")) {
                ctor = "collection";
                cargs.add("AL" + lit(init));
            } else if (ct == 2 && (kind.equals("ArrayList") || kind.equals("Vector"))) {
                ctor = "cap";
                cargs.add(num(rnd.nextInt(6) - 1));
            }
            switch (kind) {
                case "ArrayList":
                    if (ctor.equals("cap")) {
                        try {
                            l = new ArrayList(Integer.parseInt(cargs.get(0)));
                        } catch (RuntimeException e) {
                            newLine().append("new ArrayList ").append(elemName()).append(" capfail ").append(cargs.get(0)).append(" => ").append(ex(e)).append('\n');
                            continue;
                        }
                    } else {
                        l = ctor.equals("new") ? new ArrayList() : new ArrayList(init);
                    }
                    break;
                case "LinkedList": l = ctor.equals("new") ? new LinkedList() : new LinkedList(init); break;
                case "Vector": {
                    if (ctor.equals("new") && rnd.nextBoolean()) {
                        int cap = rnd.nextInt(5), inc = rnd.nextInt(4) - 1;
                        ctor = "capinc";
                        cargs.add(num(cap));
                        cargs.add(num(inc));
                        l = new Vector(cap, inc);
                    } else if (ctor.equals("cap")) {
                        try {
                            l = new Vector(Integer.parseInt(cargs.get(0)));
                        } catch (RuntimeException e) {
                            newLine().append("new Vector ").append(elemName()).append(" capfail ").append(cargs.get(0)).append(" => ").append(ex(e)).append('\n');
                            continue;
                        }
                    } else {
                        l = ctor.equals("new") ? new Vector() : new Vector(init);
                    }
                    break;
                }
                default: l = new Stack(); break;
            }
            state = l;
            newLine().append("new ").append(kind).append(' ').append(elemName()).append(' ').append(ctor);
            for (String a : cargs) out.append(' ').append(a);
            out.append('\n');
            int nops = 30 + rnd.nextInt(40);
            for (int i = 0; i < nops; i++) {
                int group = rnd.nextInt(10);
                if (group < 4) collectionOp(l, rnd.nextInt(COLLECTION_OPS));
                else if (group < 8 || !(l instanceof Deque || l instanceof Vector)) listOp(l, rnd.nextInt(LIST_OPS));
                else if (l instanceof Deque d) dequeOp(d, rnd.nextInt(DEQUE_OPS));
                else vectorOp((Vector) l, rnd.nextInt(VECTOR_OPS));
            }
        }
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void deques() {
        for (int sc = 0; sc < 60; sc++) {
            elemType = rnd.nextInt(3);
            nulls = rnd.nextInt(3) == 0;
            ArrayDeque d;
            List<String> cargs = new ArrayList<>();
            String ctor;
            switch (rnd.nextInt(3)) {
                case 0: ctor = "new"; d = new ArrayDeque(); break;
                case 1: {
                    int n = rnd.nextInt(20) - 1;
                    ctor = "cap";
                    cargs.add(num(n));
                    d = new ArrayDeque(n);
                    break;
                }
                default: {
                    List<Object> init = vals(10);
                    init.removeIf(Objects::isNull);
                    ctor = "collection";
                    cargs.add("AL" + lit(init));
                    d = new ArrayDeque(init);
                }
            }
            state = d;
            identityHash = true;
            newLine().append("new ArrayDeque ").append(elemName()).append(' ').append(ctor);
            for (String a : cargs) out.append(' ').append(a);
            out.append('\n');
            int nops = 40 + rnd.nextInt(80);
            for (int i = 0; i < nops; i++) {
                if (rnd.nextInt(10) < 3) collectionOp(d, rnd.nextInt(COLLECTION_OPS));
                else dequeOp(d, rnd.nextInt(DEQUE_OPS));
            }
            identityHash = false;
        }
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void priorityQueues() {
        String[] cmps = {"nat", "nat", "mod", "rev", "lognat", "chaos", "revmod"};
        for (int sc = 0; sc < 90; sc++) {
            elemType = rnd.nextInt(3);
            nulls = rnd.nextInt(4) == 0;
            String cn = cmps[rnd.nextInt(cmps.length)];
            PriorityQueue q;
            List<String> cargs = new ArrayList<>();
            String ctor;
            cmpCount = 0;
            cmpHash = 0;
            try {
                switch (rnd.nextInt(6)) {
                    case 0: ctor = "new"; q = cn.equals("nat") ? new PriorityQueue() : new PriorityQueue(cmp(cn)); break;
                    case 1: {
                        int cap = rnd.nextInt(4);
                        ctor = "capcmp";
                        cargs.add(num(cap));
                        q = new PriorityQueue(cap, cmp(cn));
                        break;
                    }
                    case 2: {
                        List<Object> init = vals(12);
                        ctor = "collection";
                        cargs.add("AL" + lit(init));
                        cn = "nat";
                        q = new PriorityQueue(init);
                        break;
                    }
                    case 3: {
                        List<Object> init = vals(12);
                        init.removeIf(Objects::isNull);
                        TreeSet ts = cn.equals("mod") ? new TreeSet(MOD) : new TreeSet();
                        ts.addAll(init);
                        ctor = "sortedset";
                        cargs.add((cn.equals("mod") ? "TSmod" : "TS") + lit(new ArrayList(ts)));
                        if (!cn.equals("mod")) cn = "nat";
                        cmpCount = 0;
                        cmpHash = 0;
                        q = new PriorityQueue(ts);
                        break;
                    }
                    default: {
                        List<Object> init = vals(12);
                        init.removeIf(Objects::isNull);
                        PriorityQueue src = cn.equals("nat") ? new PriorityQueue() : new PriorityQueue(cmp(cn));
                        src.addAll(init);
                        ctor = "pq";
                        cargs.add("AL" + lit(init));
                        cmpCount = 0;
                        cmpHash = 0;
                        q = new PriorityQueue(src);
                        break;
                    }
                }
            } catch (RuntimeException e) {
                newLine().append("new PriorityQueue ").append(elemName()).append(" fail ").append(ex(e)).append('\n');
                continue;
            }
            state = q;
            newLine().append("new PriorityQueue ").append(elemName()).append(' ').append(cn).append(' ').append(ctor);
            for (String a : cargs) out.append(' ').append(a);
            out.append(" => ").append(r(q)).append(" | ").append(cmpCount).append(':').append(cmpHash);
            out.append('\n');
            int nops = 40 + rnd.nextInt(80);
            identityHash = true;
            for (int i = 0; i < nops; i++) {
                if (rnd.nextInt(10) < 4) collectionOp(q, rnd.nextInt(COLLECTION_OPS));
                else pqOp(q, rnd.nextInt(PQ_OPS));
            }
            identityHash = false;
        }
        // Removing index 3 of this heap moves the last element before the
        // iterator's cursor, so the iterator returns it at the end, and
        // removing it then removes it by identity.
        elemType = 0;
        nulls = false;
        List<Object> fixed = List.of(1, 5, 2, 6, 7, 3, 4);
        PriorityQueue<Object> fq = new PriorityQueue<>(fixed);
        state = fq;
        identityHash = true;
        newLine().append("new PriorityQueue int nat collection AL").append(lit(fixed)).append(" => ").append(r(fq)).append(" | 0:0\n");
        emit("drain", a -> { String p = "n/n/n/n/r/n/n/n/r/h/n"; a.add(p); return iterProg(fq.iterator(), p, fq); });
        for (int sc = 0; sc < 12; sc++) {
            List<Object> init = new ArrayList<>();
            for (int i = 0; i < 20 + rnd.nextInt(30); i++) init.add(rnd.nextInt(100));
            PriorityQueue<Object> q = new PriorityQueue<>(init);
            state = q;
            newLine().append("new PriorityQueue int nat collection AL").append(lit(init)).append(" => ").append(r(q)).append(" | 0:0\n");
            for (int k = 0; k < 3; k++) pqOp(q, 7);
        }
        identityHash = false;
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void treeSets() {
        String[] cmps = {"nat", "nat", "mod", "rev", "lognat", "chaos", "revmod"};
        for (int sc = 0; sc < 80; sc++) {
            elemType = rnd.nextInt(3);
            nulls = rnd.nextInt(4) == 0;
            String cn = cmps[rnd.nextInt(cmps.length)];
            TreeSet s;
            List<String> cargs = new ArrayList<>();
            String ctor;
            cmpCount = 0;
            cmpHash = 0;
            try {
                switch (rnd.nextInt(4)) {
                    case 0: ctor = "new"; s = cn.equals("nat") ? new TreeSet() : new TreeSet(cmp(cn)); break;
                    case 1: {
                        List<Object> init = vals(12);
                        ctor = "collection";
                        cargs.add("AL" + lit(init));
                        cn = "nat";
                        s = new TreeSet(init);
                        break;
                    }
                    case 2: {
                        List<Object> init = vals(12);
                        init.removeIf(Objects::isNull);
                        TreeSet src = cn.equals("mod") ? new TreeSet(MOD) : new TreeSet();
                        src.addAll(init);
                        ctor = "sortedset";
                        cargs.add((cn.equals("mod") ? "TSmod" : "TS") + lit(new ArrayList(src)));
                        if (!cn.equals("mod")) cn = "nat";
                        cmpCount = 0;
                        cmpHash = 0;
                        s = new TreeSet((SortedSet) src);
                        break;
                    }
                    default: {
                        List<Object> init = vals(12);
                        init.removeIf(Objects::isNull);
                        TreeSet src = cn.equals("mod") ? new TreeSet(MOD) : new TreeSet();
                        src.addAll(init);
                        ctor = "sortedcollection";
                        cargs.add((cn.equals("mod") ? "TSmod" : "TS") + lit(new ArrayList(src)));
                        cn = "nat";
                        cmpCount = 0;
                        cmpHash = 0;
                        s = new TreeSet((Collection) src);
                        break;
                    }
                }
            } catch (RuntimeException e) {
                newLine().append("new TreeSet ").append(elemName()).append(" fail ").append(ex(e)).append('\n');
                continue;
            }
            state = s;
            newLine().append("new TreeSet ").append(elemName()).append(' ').append(cn).append(' ').append(ctor);
            for (String a : cargs) out.append(' ').append(a);
            out.append(" => ").append(r(s)).append(" | ").append(cmpCount).append(':').append(cmpHash);
            out.append('\n');
            int nops = 40 + rnd.nextInt(80);
            for (int i = 0; i < nops; i++) {
                if (rnd.nextInt(10) < 5) collectionOp(s, rnd.nextInt(COLLECTION_OPS));
                else navSetOp(s, rnd.nextInt(NAVSET_OPS));
            }
        }
    }

    static List<Map.Entry<Object, Object>> held = new ArrayList<>();

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void treeMaps() {
        String[] cmps = {"nat", "nat", "mod", "rev", "lognat", "chaos", "revmod"};
        for (int sc = 0; sc < 90; sc++) {
            elemType = rnd.nextInt(3);
            nulls = rnd.nextInt(4) == 0;
            String cn = cmps[rnd.nextInt(cmps.length)];
            TreeMap m;
            List<String> cargs = new ArrayList<>();
            String ctor;
            cmpCount = 0;
            cmpHash = 0;
            held.clear();
            try {
                switch (rnd.nextInt(4)) {
                    case 0: ctor = "new"; m = cn.equals("nat") ? new TreeMap() : new TreeMap(cmp(cn)); break;
                    case 1: {
                        ctor = "map";
                        Map src = mapArg(cargs);
                        cn = "nat";
                        cmpCount = 0;
                        cmpHash = 0;
                        m = new TreeMap(src);
                        break;
                    }
                    case 2: {
                        ctor = "sortedmap";
                        Map src = mapArg(cargs);
                        if (!(src instanceof SortedMap)) {
                            ctor = "map";
                            cn = "nat";
                            cmpCount = 0;
                            cmpHash = 0;
                            m = new TreeMap(src);
                        } else {
                            cn = ((SortedMap) src).comparator() == null ? "nat" : "mod";
                            cmpCount = 0;
                            cmpHash = 0;
                            m = new TreeMap((SortedMap) src);
                        }
                        break;
                    }
                    default: {
                        ctor = "new";
                        m = cn.equals("nat") ? new TreeMap() : new TreeMap(cmp(cn));
                        int n = rnd.nextInt(30);
                        for (int i = 0; i < n; i++) {
                            Object k = v();
                            if (k == null) continue;
                            Object val = mapValue();
                            cargs.add(r(k) + "=" + r(val));
                            m.put(k, val);
                        }
                        ctor = "puts";
                    }
                }
            } catch (RuntimeException e) {
                newLine().append("new TreeMap ").append(elemName()).append(" fail ").append(ex(e)).append('\n');
                continue;
            }
            state = m;
            newLine().append("new TreeMap ").append(elemName()).append(' ').append(cn).append(' ').append(ctor);
            for (String a : cargs) out.append(' ').append(a);
            out.append(" => ").append(r(m)).append(" | ").append(cmpCount).append(':').append(cmpHash);
            out.append('\n');
            int nops = 40 + rnd.nextInt(100);
            for (int i = 0; i < nops; i++) treeMapOp(m, rnd.nextInt(49));
        }
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void treeMapOp(TreeMap m, int x) {
        switch (x) {
            case 0: case 1: case 2: case 3:
                emit("put", a -> { Object k = v(), val = mapValue(); a.add(r(k)); a.add(r(val)); return r(m.put(k, val)); }); return;
            case 4: emit("get", a -> { Object k = v(); a.add(r(k)); return r(m.get(k)); }); return;
            case 5: case 6: emit("mremove", a -> { Object k = v(); a.add(r(k)); return r(m.remove(k)); }); return;
            case 7: emit("containsKey", a -> { Object k = v(); a.add(r(k)); return bool(m.containsKey(k)); }); return;
            case 8: emit("containsValue", a -> { Object val = mapValue(); a.add(r(val)); return bool(m.containsValue(val)); }); return;
            case 9: emit("firstKey", a -> r(m.firstKey())); return;
            case 10: emit("lastKey", a -> r(m.lastKey())); return;
            case 11: emit("firstEntry", a -> r(m.firstEntry())); return;
            case 12: emit("lastEntry", a -> r(m.lastEntry())); return;
            case 13: emit("pollFirstEntry", a -> r(m.pollFirstEntry())); return;
            case 14: emit("pollLastEntry", a -> r(m.pollLastEntry())); return;
            case 15: emit("lowerKey", a -> { Object k = v(); a.add(r(k)); return r(m.lowerKey(k)); }); return;
            case 16: emit("floorKey", a -> { Object k = v(); a.add(r(k)); return r(m.floorKey(k)); }); return;
            case 17: emit("ceilingKey", a -> { Object k = v(); a.add(r(k)); return r(m.ceilingKey(k)); }); return;
            case 18: emit("higherKey", a -> { Object k = v(); a.add(r(k)); return r(m.higherKey(k)); }); return;
            case 19: emit("lowerEntry", a -> { Object k = v(); a.add(r(k)); return r(m.lowerEntry(k)); }); return;
            case 20: emit("floorEntry", a -> { Object k = v(); a.add(r(k)); return r(m.floorEntry(k)); }); return;
            case 21: emit("ceilingEntry", a -> { Object k = v(); a.add(r(k)); return r(m.ceilingEntry(k)); }); return;
            case 22: emit("higherEntry", a -> { Object k = v(); a.add(r(k)); return r(m.higherEntry(k)); }); return;
            case 23: emit("putIfAbsent", a -> { Object k = v(), val = mapValue(); a.add(r(k)); a.add(r(val)); return r(m.putIfAbsent(k, val)); }); return;
            case 24: emit("getOrDefault", a -> { Object k = v(), val = mapValue(); a.add(r(k)); a.add(r(val)); return r(m.getOrDefault(k, val)); }); return;
            case 25: emit("replace3", a -> { Object k = v(), v1 = mapValue(), v2 = mapValue(); a.add(r(k)); a.add(r(v1)); a.add(r(v2)); return bool(m.replace(k, v1, v2)); }); return;
            case 26: emit("replace", a -> { Object k = v(), val = mapValue(); a.add(r(k)); a.add(r(val)); return r(m.replace(k, val)); }); return;
            case 27: emit("remove2", a -> { Object k = v(), val = mapValue(); a.add(r(k)); a.add(r(val)); return bool(m.remove(k, val)); }); return;
            case 28: emit("msize", a -> num(m.size())); return;
            case 29: if (rnd.nextInt(8) == 0) emit("mclear", a -> { m.clear(); return OK; }); return;
            case 30: emit("putAll", a -> { m.putAll(mapArg(a)); return OK; }); return;
            case 31: emit("keys", a -> {
                String p = randomIterProg(m.size(), false, false, rnd.nextInt(8) == 0);
                a.add(p);
                return iterProg(m.keySet().iterator(), p, m);
            }); return;
            case 32: emit("values", a -> {
                String p = randomIterProg(m.size(), false, false, rnd.nextInt(8) == 0);
                a.add(p);
                return iterProg(m.values().iterator(), p, m);
            }); return;
            case 33: case 34: emit("entries", a -> {
                String p = randomIterProg(m.size(), false, true, rnd.nextInt(8) == 0);
                a.add(p);
                return iterProg(m.entrySet().iterator(), p, m);
            }); return;
            case 35: emit("desckeys", a -> {
                String p = randomIterProg(m.size(), false, false, false);
                a.add(p);
                return iterProg(m.navigableKeySet().descendingIterator(), p, m);
            }); return;
            case 36: {
                int y = rnd.nextInt(COLLECTION_OPS + NAVSET_OPS);
                NavigableSet ks = m.navigableKeySet();
                prefix = "ks ";
                if (y < COLLECTION_OPS) collectionOp(ks, y);
                else if (y - COLLECTION_OPS == 9) emit("size", a -> num(ks.size()));
                else navSetOp(ks, y - COLLECTION_OPS);
                prefix = "";
                return;
            }
            case 37: {
                int y = rnd.nextInt(COLLECTION_OPS);
                if (y == 0 || y == 5) y = 1;
                Collection vs = m.values();
                prefix = "vs ";
                identityHash = true;
                collectionOp(vs, y);
                identityHash = false;
                prefix = "";
                return;
            }
            case 38: emit("escontains", a -> { Object k = v(), val = mapValue(); a.add(r(k)); a.add(r(val)); return bool(m.entrySet().contains(new AbstractMap.SimpleEntry(k, val))); }); return;
            case 39: emit("esremove", a -> { Object k = v(), val = mapValue(); a.add(r(k)); a.add(r(val)); return bool(m.entrySet().remove(new AbstractMap.SimpleEntry(k, val))); }); return;
            case 40: emit("hold", a -> { held.clear(); held.addAll(m.entrySet()); return num(held.size()); }); return;
            case 41: emit("held", a -> { StringJoiner j = new StringJoiner(","); for (Map.Entry e : held) j.add(r(e)); return j.toString(); }); return;
            case 42: emit("mequals", a -> bool(m.equals(mapArg(a)))); return;
            case 43: emit("mhash", a -> num(m.hashCode())); return;
            case 44: emit("mtostring", a -> toStr(m)); return;
            case 45: emit("mclone", a -> { TreeMap c = (TreeMap) m.clone(); return "TreeMap:" + r(c); }); return;
            case 46: emit("mcme", a -> {
                Object k = v(), val = mapValue();
                a.add(r(k));
                a.add(r(val));
                Iterator it = m.entrySet().iterator();
                StringJoiner j = new StringJoiner("/");
                try { j.add(r(it.next())); } catch (RuntimeException t) { j.add(ex(t)); }
                try { j.add(r(m.put(k, val))); } catch (RuntimeException t) { j.add(ex(t)); }
                try { j.add(bool(it.hasNext())); j.add(r(it.next())); } catch (RuntimeException t) { j.add(ex(t)); }
                return j.toString();
            }); return;
            case 47: emit("esops", a -> esOps(m, a)); return;
            case 48: emit("exportedEntry", a -> {
                boolean first = rnd.nextBoolean();
                Object k = v();
                if (!first) a.add(r(k));
                Map.Entry e = first ? m.firstEntry() : m.ceilingEntry(k);
                if (e == null) return "null";
                String res = toStr(e) + ";" + e.hashCode() + ";" + e.equals(new AbstractMap.SimpleEntry(e.getKey(), e.getValue()));
                try { e.setValue(1); } catch (RuntimeException t) { res += ";" + ex(t); }
                return res;
            }); return;
        }
    }

    // esOps exercises a map's entry set and its entries.
    @SuppressWarnings({"unchecked", "rawtypes"})
    static String esOps(Map m, List<String> a) {
        Object k = v(), val = mapValue();
        a.add(r(k));
        a.add(r(val));
        Map other = mapArg(a);
        Set es = m.entrySet();
        StringJoiner j = new StringJoiner("/");
        j.add(r(es.toArray()));
        j.add(r(es.toArray(new Object[1])));
        try { j.add(bool(es.contains(new AbstractMap.SimpleEntry(k, val)))); } catch (RuntimeException t) { j.add(ex(t)); }
        try { j.add(bool(es.containsAll(other.entrySet()))); } catch (RuntimeException t) { j.add(ex(t)); }
        j.add(bool(es.equals(other.entrySet())));
        j.add(num(es.hashCode()));
        j.add(toStr(es));
        Iterator it = es.iterator();
        if (it.hasNext()) {
            Map.Entry e = (Map.Entry) it.next();
            j.add(toStr(e));
            j.add(num(e.hashCode()));
            j.add(bool(e.equals(new AbstractMap.SimpleEntry(e.getKey(), e.getValue()))));
            j.add(bool(e.equals(new AbstractMap.SimpleEntry(e.getKey(), val))));
        }
        j.add(r(m.keySet()) + ";" + toStr(m.keySet()) + ";" + num(m.keySet().hashCode()));
        try { j.add(bool(m.keySet().contains(k))); } catch (RuntimeException t) { j.add(ex(t)); }
        j.add(r(m.values()) + ";" + bool(m.values().contains(val)) + ";" + bool(m.containsValue(val)));
        return j.toString();
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void enumSets() {
        elemType = 3;
        for (int sc = 0; sc < 40; sc++) {
            nulls = rnd.nextInt(3) == 0;
            EnumSet s;
            int n = rnd.nextInt(6) + 1;
            List<Object> init = new ArrayList<>();
            for (int i = 0; i < n; i++) init.add(Ord.values()[rnd.nextInt(6)]);
            switch (n) {
                case 1: s = EnumSet.of((Ord) init.get(0)); break;
                case 2: s = EnumSet.of((Ord) init.get(0), (Ord) init.get(1)); break;
                case 3: s = EnumSet.of((Ord) init.get(0), (Ord) init.get(1), (Ord) init.get(2)); break;
                case 4: s = EnumSet.of((Ord) init.get(0), (Ord) init.get(1), (Ord) init.get(2), (Ord) init.get(3)); break;
                case 5: s = EnumSet.of((Ord) init.get(0), (Ord) init.get(1), (Ord) init.get(2), (Ord) init.get(3), (Ord) init.get(4)); break;
                default: s = EnumSet.of((Ord) init.get(0), init.subList(1, n).toArray(new Ord[0])); break;
            }
            state = s;
            newLine().append("new EnumSet ord of AL").append(lit(init)).append(" => ").append(r(s)).append('\n');
            int nops = 20 + rnd.nextInt(30);
            for (int i = 0; i < nops; i++) {
                int x = rnd.nextInt(COLLECTION_OPS + 3);
                if (x < COLLECTION_OPS) {
                    identityHash = true;
                    collectionOp(s, x);
                    identityHash = false;
                } else if (x == COLLECTION_OPS) {
                    emit("clone", a -> "RegularEnumSet:" + r(s.clone()));
                } else if (x == COLLECTION_OPS + 1) {
                    if (rnd.nextBoolean()) emit("containsInt", a -> bool(((Collection) s).contains(5)));
                    else emit("copyOfSet", a -> r(EnumSet.copyOf(s)));
                } else {
                    emit("copyOf", a -> r(EnumSet.copyOf((Collection) collArg(a, true))));
                }
            }
        }
    }

    public static void main(String[] args) throws Exception {
        String dir = args.length > 0 ? args[0] : ".";
        write(dir + "/coll_lists.txt", 1, CollVectors::lists);
        write(dir + "/coll_deques.txt", 2, CollVectors::deques);
        write(dir + "/coll_pq.txt", 3, CollVectors::priorityQueues);
        write(dir + "/coll_treeset.txt", 4, CollVectors::treeSets);
        write(dir + "/coll_treemap.txt", 5, CollVectors::treeMaps);
        write(dir + "/coll_enumset.txt", 6, CollVectors::enumSets);
        write(dir + "/coll_static.txt", 7, CollStatic::run);
    }

    static void write(String path, long seed, Runnable gen) throws Exception {
        rnd = new Random(seed);
        out = new StringBuilder();
        gen.run();
        java.nio.file.Files.writeString(java.nio.file.Path.of(path), out);
    }
}

// CollStatic generates the vectors for Arrays, Collections and TimSort.
class CollStatic {
    static Random rnd() {
        return CollVectors.rnd;
    }

    static void emit(String name, CollVectors.Op op) {
        CollVectors.state = null;
        CollVectors.emit(name, op);
    }

    static String lit(List<Object> l) {
        return CollVectors.lit(l);
    }

    static String r(Object o) {
        return CollVectors.r(o);
    }

    static String arrLit(Object[] a) {
        return "A" + lit(Arrays.asList(a));
    }

    static Object[] objArray(int max) {
        return CollVectors.vals(max).toArray();
    }

    static final String[] SORT_CMPS = {"nat", "rev", "mod", "lognat", "revmod", "chaos"};

    static Object sameType(Object x, Object[] pool) {
        for (int i = 0; i < 50; i++) {
            Object y = pool[rnd().nextInt(pool.length)];
            if (x != null && y != null && x.getClass() == y.getClass()) return y;
        }
        return x;
    }

    static String boxLit(Object o) {
        if (o == null) return "null";
        if (o instanceof Long l) return "L" + l;
        if (o instanceof Double d) return "D" + dbits(d);
        if (o instanceof Float f) return "F" + fbits(f);
        if (o instanceof Short s) return "S" + s;
        if (o instanceof Byte b) return "B" + b;
        if (o instanceof Character c) return "C" + (int) c;
        if (o instanceof Boolean b) return "Z" + (b ? 1 : 0);
        return r(o);
    }

    static String randCmp() {
        return SORT_CMPS[rnd().nextInt(SORT_CMPS.length)];
    }

    // lcg is the generator for large inputs, which are given by seed rather
    // than listed: x = x * 6364136223846793005 + 1442695040888963407, and
    // each value is (x >>> 33) % domain.
    static long lcg;

    static int lcgNext(int domain) {
        lcg = lcg * 6364136223846793005L + 1442695040888963407L;
        return (int) ((lcg >>> 33) % domain);
    }

    static Object[] genArray(long seed, int n, int domain, int type) {
        lcg = seed;
        Object[] a = new Object[n];
        for (int i = 0; i < n; i++) {
            int x = lcgNext(domain);
            a[i] = type == 0 ? (Object) x : new CollVectors.Item(x, lcgNext(1000));
        }
        return a;
    }

    static String hashArray(Object[] a) {
        int h = 0;
        for (Object o : a) h = h * 31 + r(o).hashCode();
        return a.length + "#" + h;
    }

    static String dbits(double d) {
        return String.format("%016x", Double.doubleToRawLongBits(d));
    }

    static String fbits(float f) {
        return String.format("%08x", Float.floatToRawIntBits(f));
    }

    static final long[] SPECIAL_D = {
        0x0000000000000000L, 0x8000000000000000L, 0x3ff0000000000000L, 0xbff0000000000000L,
        0x7ff8000000000000L, 0x7ff0000000000001L, 0xfff8000000000000L, 0x7ffc000000000123L,
        0x7ff0000000000000L, 0xfff0000000000000L, 0x0000000000000001L, 0x8000000000000001L,
        0x7fefffffffffffffL, 0x01a56e1fc2f8f359L, 0x4000000000000000L, 0xc008000000000000L,
    };

    static final int[] SPECIAL_F = {
        0x00000000, 0x80000000, 0x3f800000, 0xbf800000, 0x7fc00000, 0x7f800001, 0xffc00000,
        0x7fc00123, 0x7f800000, 0xff800000, 0x00000001, 0x80000001, 0x7f7fffff, 0x40000000,
    };

    static double[] dArray(int n) {
        double[] a = new double[n];
        for (int i = 0; i < n; i++) {
            if (rnd().nextInt(3) == 0) a[i] = Double.longBitsToDouble(SPECIAL_D[rnd().nextInt(SPECIAL_D.length)]);
            else a[i] = (rnd().nextInt(21) - 10) / 4.0;
        }
        return a;
    }

    // noMinValue replaces Double.MIN_VALUE, whose formatting is
    // Double.toString's concern and has its own vectors.
    static double[] noMinValue(double[] a) {
        for (int i = 0; i < a.length; i++) if (Math.abs(a[i]) == Double.MIN_VALUE) a[i] = 0.5;
        return a;
    }

    static float[] fArray(int n) {
        float[] a = new float[n];
        for (int i = 0; i < n; i++) {
            if (rnd().nextInt(3) == 0) a[i] = Float.intBitsToFloat(SPECIAL_F[rnd().nextInt(SPECIAL_F.length)]);
            else a[i] = (rnd().nextInt(21) - 10) / 4.0f;
        }
        return a;
    }

    static String dLit(double[] a) {
        StringJoiner j = new StringJoiner(",", "D[", "]");
        for (double d : a) j.add(dbits(d));
        return j.toString();
    }

    static String fLit(float[] a) {
        StringJoiner j = new StringJoiner(",", "F[", "]");
        for (float f : a) j.add(fbits(f));
        return j.toString();
    }

    static String iLit(int[] a) {
        StringJoiner j = new StringJoiner(",", "I[", "]");
        for (int x : a) j.add(Integer.toString(x));
        return j.toString();
    }

    static String lLit(long[] a) {
        StringJoiner j = new StringJoiner(",", "J[", "]");
        for (long x : a) j.add(Long.toString(x));
        return j.toString();
    }

    static String sLit(short[] a) {
        StringJoiner j = new StringJoiner(",", "S[", "]");
        for (short x : a) j.add(Short.toString(x));
        return j.toString();
    }

    static String cLit(char[] a) {
        StringJoiner j = new StringJoiner(",", "C[", "]");
        for (char x : a) j.add(Integer.toString(x));
        return j.toString();
    }

    static String bLit(byte[] a) {
        StringJoiner j = new StringJoiner(",", "B[", "]");
        for (byte x : a) j.add(Byte.toString(x));
        return j.toString();
    }

    static int[] iArray(int n) {
        int[] a = new int[n];
        for (int i = 0; i < n; i++) a[i] = rnd().nextInt(4) == 0 ? rnd().nextInt() : rnd().nextInt(20) - 10;
        return a;
    }

    static void run() {
        CollVectors.nulls = false;
        objectSorts();
        primitiveSorts();
        searches();
        copies();
        arraysMisc();
        asLists();
        collections();
    }

    static void objectSorts() {
        for (int i = 0; i < 250; i++) {
            CollVectors.elemType = rnd().nextInt(2) == 0 ? 0 : 1;
            CollVectors.nulls = rnd().nextInt(10) == 0;
            int max = rnd().nextInt(4) == 0 ? 80 : 40;
            Object[] a = objArray(max);
            String cn = randCmp();
            int kind = rnd().nextInt(3);
            if (kind == 0) {
                emit("sortarr", args -> {
                    args.add(cn);
                    args.add(arrLit(a));
                    try {
                        Arrays.sort(a, CollVectors.cmp(cn));
                    } catch (RuntimeException e) {
                        return CollVectors.ex(e) + ":" + arrLit(a);
                    }
                    return arrLit(a);
                });
            } else if (kind == 1) {
                int from = rnd().nextInt(a.length + 2) - 1;
                int to = rnd().nextInt(a.length + 2) - 1;
                if (rnd().nextInt(6) != 0 && from > to) { int t = from; from = to; to = t; }
                int f = from, t = to;
                emit("sortrange", args -> {
                    args.add(cn);
                    args.add(Integer.toString(f));
                    args.add(Integer.toString(t));
                    args.add(arrLit(a));
                    try {
                        Arrays.sort(a, f, t, CollVectors.cmp(cn));
                    } catch (RuntimeException e) {
                        return CollVectors.ex(e) + ":" + arrLit(a);
                    }
                    return arrLit(a);
                });
            } else {
                List<Object> l = new ArrayList<>(Arrays.asList(a));
                emit("sortlist", args -> {
                    args.add(cn);
                    args.add("AL" + lit(l));
                    try {
                        Collections.sort(l, CollVectors.cmp(cn));
                    } catch (RuntimeException e) {
                        return CollVectors.ex(e) + ":" + lit(l);
                    }
                    return lit(l);
                });
            }
        }
        CollVectors.nulls = false;
        for (int i = 0; i < 80; i++) {
            long seed = rnd().nextLong();
            int n = 32 + rnd().nextInt(i < 60 ? 600 : 6000);
            int domain = 1 + rnd().nextInt(rnd().nextBoolean() ? 20 : 100000);
            int type = rnd().nextInt(2);
            String cn = randCmp();
            emit("sortgen", args -> {
                args.add(Long.toString(seed));
                args.add(Integer.toString(n));
                args.add(Integer.toString(domain));
                args.add(Integer.toString(type));
                args.add(cn);
                Object[] a = genArray(seed, n, domain, type);
                // Presorted and reversed runs exercise galloping.
                if (domain % 3 == 0) Arrays.sort(a, 0, n / 2, (x, y) -> Integer.compare(CollVectors.keyOf(x), CollVectors.keyOf(y)));
                if (domain % 5 == 0) Arrays.sort(a, n / 3, n, (x, y) -> Integer.compare(CollVectors.keyOf(y), CollVectors.keyOf(x)));
                CollVectors.cmpCount = 0;
                CollVectors.cmpHash = 0;
                String ex = "";
                try {
                    Arrays.sort(a, CollVectors.cmp(cn));
                } catch (RuntimeException e) {
                    ex = CollVectors.ex(e) + ":";
                }
                return ex + hashArray(a);
            });
        }
    }

    static void primitiveSorts() {
        for (int i = 0; i < 220; i++) {
            int n = rnd().nextInt(i < 200 ? 40 : 300);
            int kind = rnd().nextInt(7);
            boolean range = rnd().nextInt(4) == 0;
            int from = range ? rnd().nextInt(n + 1) : 0;
            int to = range ? from + rnd().nextInt(n - from + 1) : n;
            String rn = range ? "r" : "";
            switch (kind) {
                case 0: case 1: {
                    double[] a = dArray(n);
                    emit("sortd" + rn, args -> {
                        if (range) { args.add("" + from); args.add("" + to); }
                        args.add(dLit(a));
                        if (range) Arrays.sort(a, from, to); else Arrays.sort(a);
                        return dLit(a);
                    });
                    break;
                }
                case 2: {
                    float[] a = fArray(n);
                    emit("sortf" + rn, args -> {
                        if (range) { args.add("" + from); args.add("" + to); }
                        args.add(fLit(a));
                        if (range) Arrays.sort(a, from, to); else Arrays.sort(a);
                        return fLit(a);
                    });
                    break;
                }
                case 3: {
                    int[] a = iArray(n);
                    emit("sorti" + rn, args -> {
                        if (range) { args.add("" + from); args.add("" + to); }
                        args.add(iLit(a));
                        if (range) Arrays.sort(a, from, to); else Arrays.sort(a);
                        return iLit(a);
                    });
                    break;
                }
                case 4: {
                    long[] a = new long[n];
                    for (int k = 0; k < n; k++) a[k] = rnd().nextInt(3) == 0 ? rnd().nextLong() : rnd().nextInt(20) - 10;
                    emit("sortl" + rn, args -> {
                        if (range) { args.add("" + from); args.add("" + to); }
                        args.add(lLit(a));
                        if (range) Arrays.sort(a, from, to); else Arrays.sort(a);
                        return lLit(a);
                    });
                    break;
                }
                case 5: {
                    short[] a = new short[n];
                    char[] c = new char[n];
                    for (int k = 0; k < n; k++) { a[k] = (short) rnd().nextInt(); c[k] = (char) rnd().nextInt(); }
                    emit("sorts" + rn, args -> {
                        if (range) { args.add("" + from); args.add("" + to); }
                        args.add(sLit(a));
                        if (range) Arrays.sort(a, from, to); else Arrays.sort(a);
                        return sLit(a);
                    });
                    emit("sortc" + rn, args -> {
                        if (range) { args.add("" + from); args.add("" + to); }
                        args.add(cLit(c));
                        if (range) Arrays.sort(c, from, to); else Arrays.sort(c);
                        return cLit(c);
                    });
                    break;
                }
                default: {
                    byte[] a = new byte[n];
                    rnd().nextBytes(a);
                    emit("sortb" + rn, args -> {
                        if (range) { args.add("" + from); args.add("" + to); }
                        args.add(bLit(a));
                        if (range) Arrays.sort(a, from, to); else Arrays.sort(a);
                        return bLit(a);
                    });
                }
            }
        }
        // Range errors.
        emit("sortir", args -> { args.add("3"); args.add("1"); args.add("I[1,2,3,4]"); Arrays.sort(new int[]{1, 2, 3, 4}, 3, 1); return "ok"; });
        emit("sortir", args -> { args.add("-1"); args.add("2"); args.add("I[1,2,3,4]"); Arrays.sort(new int[]{1, 2, 3, 4}, -1, 2); return "ok"; });
        emit("sortir", args -> { args.add("0"); args.add("5"); args.add("I[1,2,3,4]"); Arrays.sort(new int[]{1, 2, 3, 4}, 0, 5); return "ok"; });
    }

    static void searches() {
        for (int i = 0; i < 200; i++) {
            int n = rnd().nextInt(30);
            int kind = rnd().nextInt(4);
            switch (kind) {
                case 0: {
                    int[] a = new int[n];
                    for (int k = 0; k < n; k++) a[k] = rnd().nextInt(30);
                    Arrays.sort(a);
                    int key = rnd().nextInt(32) - 1;
                    if (rnd().nextBoolean()) {
                        emit("bsi", args -> { args.add("" + key); args.add(iLit(a)); return "" + Arrays.binarySearch(a, key); });
                    } else {
                        int from = rnd().nextInt(n + 2) - 1, to = rnd().nextInt(n + 2) - 1;
                        emit("bsir", args -> { args.add("" + from); args.add("" + to); args.add("" + key); args.add(iLit(a)); return "" + Arrays.binarySearch(a, from, to, key); });
                    }
                    break;
                }
                case 1: {
                    double[] a = dArray(n);
                    Arrays.sort(a);
                    double key = rnd().nextBoolean() && n > 0 ? a[rnd().nextInt(n)] : dArray(1)[0];
                    emit("bsd", args -> { args.add(dbits(key)); args.add(dLit(a)); return "" + Arrays.binarySearch(a, key); });
                    break;
                }
                case 2: {
                    float[] a = fArray(n);
                    Arrays.sort(a);
                    float key = rnd().nextBoolean() && n > 0 ? a[rnd().nextInt(n)] : fArray(1)[0];
                    emit("bsf", args -> { args.add(fbits(key)); args.add(fLit(a)); return "" + Arrays.binarySearch(a, key); });
                    break;
                }
                default: {
                    CollVectors.elemType = rnd().nextInt(2);
                    Object[] a = objArray(n);
                    String cn = rnd().nextBoolean() ? "nat" : "mod";
                    try { Arrays.sort(a, CollVectors.cmp(cn)); } catch (RuntimeException e) { /* searched unsorted */ }
                    Object key = CollVectors.v();
                    if (rnd().nextBoolean()) {
                        emit("bso", args -> { args.add(cn); args.add(r(key)); args.add(arrLit(a)); return "" + Arrays.binarySearch(a, key, CollVectors.cmp(cn)); });
                    } else {
                        List<Object> l = new ArrayList<>(Arrays.asList(a));
                        emit("bsl", args -> { args.add(cn); args.add(r(key)); args.add("AL" + lit(l)); return "" + Collections.binarySearch(l, key, CollVectors.cmp(cn)); });
                    }
                }
            }
        }
    }

    static void copies() {
        for (int i = 0; i < 100; i++) {
            int n = rnd().nextInt(10);
            int newLen = rnd().nextInt(14) - 1;
            int from = rnd().nextInt(n + 3) - 1;
            int to = rnd().nextInt(n + 6) - 1;
            switch (rnd().nextInt(6)) {
                case 0: { int[] a = iArray(n); emit("copyofi", args -> { args.add("" + newLen); args.add(iLit(a)); return iLit(Arrays.copyOf(a, newLen)); }); break; }
                case 1: { double[] a = dArray(n); emit("copyofd", args -> { args.add("" + newLen); args.add(dLit(a)); return dLit(Arrays.copyOf(a, newLen)); }); break; }
                case 2: { float[] a = fArray(n); emit("copyoff", args -> { args.add("" + newLen); args.add(fLit(a)); return fLit(Arrays.copyOf(a, newLen)); }); break; }
                case 3: { CollVectors.elemType = 0; Object[] a = objArray(n); emit("copyofo", args -> { args.add("" + newLen); args.add(arrLit(a)); return arrLit(Arrays.copyOf(a, newLen)); }); break; }
                case 4: { int[] a = iArray(n); emit("copyofrangei", args -> { args.add("" + from); args.add("" + to); args.add(iLit(a)); return iLit(Arrays.copyOfRange(a, from, to)); }); break; }
                default: { CollVectors.elemType = 2; Object[] a = objArray(n); emit("copyofrangeo", args -> { args.add("" + from); args.add("" + to); args.add(arrLit(a)); return arrLit(Arrays.copyOfRange(a, from, to)); }); break; }
            }
        }
    }

    static void arraysMisc() {
        for (int i = 0; i < 120; i++) {
            int n = rnd().nextInt(8);
            CollVectors.elemType = rnd().nextInt(3);
            CollVectors.nulls = rnd().nextInt(3) == 0;
            switch (rnd().nextInt(9)) {
                case 0: {
                    Object[] a = objArray(n);
                    Object[] b = rnd().nextBoolean() ? a.clone() : objArray(n);
                    if (rnd().nextInt(4) == 0 && b.length > 0) b[rnd().nextInt(b.length)] = CollVectors.v();
                    emit("eqo", args -> { args.add(arrLit(a)); args.add(arrLit(b)); return CollVectors.bool(Arrays.equals(a, b)); });
                    break;
                }
                case 1: {
                    double[] a = dArray(n);
                    double[] b = rnd().nextBoolean() ? a.clone() : dArray(n);
                    emit("eqd", args -> { args.add(dLit(a)); args.add(dLit(b)); return CollVectors.bool(Arrays.equals(a, b)); });
                    break;
                }
                case 2: { double[] a = dArray(n); emit("hashd", args -> { args.add(dLit(a)); return "" + Arrays.hashCode(a); }); break; }
                case 3: { int[] a = iArray(n); emit("hashi", args -> { args.add(iLit(a)); return "" + Arrays.hashCode(a); }); break; }
                case 4: { Object[] a = objArray(n); emit("hasho", args -> { args.add(arrLit(a)); return "" + Arrays.hashCode(a); }); break; }
                case 5: { Object[] a = objArray(n); emit("tostro", args -> { args.add(arrLit(a)); return CollVectors.toStr(Arrays.toString(a)); }); break; }
                case 6: { double[] a = noMinValue(dArray(n)); emit("tostrd", args -> { args.add(dLit(a)); return CollVectors.toStr(Arrays.toString(a)); }); break; }
                case 7: {
                    double[] a = dArray(n);
                    double val = dArray(1)[0];
                    int from = rnd().nextInt(n + 2) - 1, to = rnd().nextInt(n + 2) - 1;
                    emit("filldr", args -> { args.add(dbits(val)); args.add("" + from); args.add("" + to); args.add(dLit(a)); Arrays.fill(a, from, to, val); return dLit(a); });
                    break;
                }
                default: {
                    Object[] a = objArray(n);
                    Object val = CollVectors.v();
                    emit("fillo", args -> { args.add(r(val)); args.add(arrLit(a)); Arrays.fill(a, val); return arrLit(a); });
                }
            }
        }
        for (int i = 0; i < 60; i++) primMisc(rnd().nextInt(8));
        emit("deep", args -> CollVectors.toStr(Arrays.deepToString(new Object[]{1, new Object[]{"a", null, new double[]{1.5, -0.0}}, new int[]{3, 4}, null, new Object[0]})));
        Object[] self = new Object[2];
        self[0] = self;
        self[1] = "x";
        emit("deepself", args -> CollVectors.toStr(Arrays.deepToString(self)));
    }

    static String zLit(boolean[] a) {
        StringJoiner j = new StringJoiner(",", "Z[", "]");
        for (boolean x : a) j.add(x ? "1" : "0");
        return j.toString();
    }

    // primMisc covers the primitive overloads of toString, hashCode,
    // equals, copyOf, fill and binarySearch not covered elsewhere.
    static void primMisc(int n) {
        long[] j = new long[n];
        short[] sh = new short[n];
        char[] c = new char[n];
        byte[] b = new byte[n];
        boolean[] z = new boolean[n];
        float[] f = fArray(n);
        // Formatting Float.MIN_VALUE is Float.toString's concern, which has
        // its own vectors.
        for (int k = 0; k < n; k++) if (Math.abs(f[k]) == Float.MIN_VALUE) f[k] = 0.5f;
        for (int k = 0; k < n; k++) {
            j[k] = rnd().nextBoolean() ? rnd().nextLong() : rnd().nextInt(10);
            sh[k] = (short) rnd().nextInt();
            c[k] = (char) (rnd().nextBoolean() ? 'a' + rnd().nextInt(26) : rnd().nextInt(0x10000));
            b[k] = (byte) rnd().nextInt();
            z[k] = rnd().nextBoolean();
        }
        int newLen = rnd().nextInt(n + 3) - 1;
        emit("primj", args -> {
            args.add(lLit(j));
            args.add("" + newLen);
            long key = n > 0 && rnd().nextBoolean() ? j[rnd().nextInt(n)] : rnd().nextInt(10);
            args.add("" + key);
            long[] sorted = j.clone();
            Arrays.sort(sorted);
            return CollVectors.toStr(Arrays.toString(j)) + ";" + Arrays.hashCode(j) + ";" + Arrays.equals(j, j.clone())
                + ";" + Arrays.binarySearch(sorted, key) + ";" + lLit(Arrays.copyOf(j, Math.max(newLen, 0)));
        });
        emit("prims", args -> {
            args.add(sLit(sh));
            args.add("" + newLen);
            return CollVectors.toStr(Arrays.toString(sh)) + ";" + Arrays.hashCode(sh) + ";" + sLit(Arrays.copyOf(sh, Math.max(newLen, 0)));
        });
        emit("primc", args -> {
            args.add(cLit(c));
            args.add("" + newLen);
            char[] sorted = c.clone();
            Arrays.sort(sorted);
            char key = n > 0 ? c[0] : 'a';
            return Arrays.hashCode(c) + ";" + cLit(Arrays.copyOf(c, Math.max(newLen, 0))) + ";" + Arrays.binarySearch(sorted, key);
        });
        emit("primb", args -> {
            args.add(bLit(b));
            args.add("" + newLen);
            return CollVectors.toStr(Arrays.toString(b)) + ";" + Arrays.hashCode(b) + ";" + bLit(Arrays.copyOf(b, Math.max(newLen, 0)));
        });
        emit("primz", args -> {
            args.add(zLit(z));
            args.add("" + newLen);
            boolean[] filled = z.clone();
            Arrays.fill(filled, true);
            return CollVectors.toStr(Arrays.toString(z)) + ";" + Arrays.hashCode(z) + ";" + zLit(Arrays.copyOf(z, Math.max(newLen, 0))) + ";" + zLit(filled) + ";" + Arrays.equals(z, filled);
        });
        emit("primf", args -> {
            args.add(fLit(f));
            float[] g = f.clone();
            if (n > 0 && rnd().nextBoolean()) g[rnd().nextInt(n)] = Float.intBitsToFloat(SPECIAL_F[rnd().nextInt(SPECIAL_F.length)]);
            args.add(fLit(g));
            return CollVectors.toStr(Arrays.toString(f)) + ";" + Arrays.hashCode(f) + ";" + Arrays.equals(f, g);
        });
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void asLists() {
        for (int i = 0; i < 60; i++) {
            CollVectors.elemType = rnd().nextInt(3);
            CollVectors.nulls = rnd().nextInt(3) == 0;
            boolean typed = CollVectors.elemType == 2 && rnd().nextBoolean();
            Object[] a = objArray(8);
            if (typed) {
                String[] s = new String[a.length];
                for (int k = 0; k < a.length; k++) s[k] = (String) a[k];
                a = s;
            }
            Object[] arr = a;
            List l = Arrays.asList(arr);
            CollVectors.state = l;
            CollVectors.identityHash = false;
            CollVectors.newLine().append("new AsList ").append(typed ? "strarr" : "objarr").append(' ').append(arrLit(arr)).append('\n');
            int nops = 10 + rnd().nextInt(30);
            for (int k = 0; k < nops; k++) {
                int g = rnd().nextInt(10);
                if (g < 4) CollVectors.collectionOp(l, rnd().nextInt(CollVectors.COLLECTION_OPS));
                else if (g < 9) CollVectors.listOp(l, rnd().nextInt(CollVectors.LIST_OPS));
                else CollVectors.emit("array", args -> arrLit(arr));
            }
        }
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void collections() {
        for (int i = 0; i < 300; i++) {
            CollVectors.elemType = rnd().nextInt(3);
            CollVectors.nulls = rnd().nextInt(5) == 0;
            List<Object> base = CollVectors.vals(rnd().nextInt(5) == 0 ? 30 : 8);
            switch (rnd().nextInt(9)) {
                case 0: emit("reverse", args -> { List<Object> l = new ArrayList<>(base); args.add("AL" + lit(l)); Collections.reverse(l); return lit(l); }); break;
                case 1: emit("reverseLL", args -> { List<Object> l = new LinkedList<>(base); args.add("LL" + lit(l)); Collections.reverse(l); return lit(l); }); break;
                case 2: emit("min", args -> { String cn = randCmp(); args.add(cn); args.add("AL" + lit(base)); return r(Collections.min(base, CollVectors.cmp(cn))); }); break;
                case 3: emit("max", args -> { String cn = randCmp(); args.add(cn); args.add("AL" + lit(base)); return r(Collections.max(base, CollVectors.cmp(cn))); }); break;
                case 4: emit("swap", args -> { List<Object> l = new ArrayList<>(base); int x = rnd().nextInt(l.size() + 2) - 1, y = rnd().nextInt(l.size() + 1); args.add("" + x); args.add("" + y); args.add("AL" + lit(l)); Collections.swap(l, x, y); return lit(l); }); break;
                case 5: emit("addAllArr", args -> { List<Object> l = new ArrayList<>(base); Object[] extra = CollVectors.vals(4).toArray(); args.add("AL" + lit(l)); args.add(arrLit(extra)); boolean b = Collections.addAll(l, extra); return b + ":" + lit(l); }); break;
                case 6: emit("revcmp", args -> {
                    String cn = randCmp();
                    args.add(cn);
                    Comparator c = Collections.reverseOrder(CollVectors.cmp(cn));
                    Comparator c2 = Collections.reverseOrder(c);
                    Object x = CollVectors.v(), y = CollVectors.v();
                    args.add(r(x));
                    args.add(r(y));
                    return c.compare(x, y) + "," + c2.compare(x, y);
                }); break;
                case 7: emit("fill", args -> { List<Object> l = new LinkedList<>(base); Object x = CollVectors.v(); args.add(r(x)); args.add("LL" + lit(l)); Collections.fill(l, x); return lit(l); }); break;
                default: emit("minnat", args -> { args.add("AL" + lit(base)); return r(Collections.min((Collection) base)) + "," + r(Collections.max((Collection) base)); }); break;
            }
        }
        Object[] boxed = {
            1L, -5L, Long.MIN_VALUE, 1.5, -0.0, 0.0, Double.NaN, Double.POSITIVE_INFINITY, 2.5f, -0.0f, Float.NaN,
            (short) 3, (short) -7, (byte) 4, (byte) -128, 'a', 'z', true, false, 7, -3, "ab", null,
        };
        for (int i = 0; i < 150; i++) {
            Object x = boxed[rnd().nextInt(boxed.length)];
            Object y = rnd().nextInt(3) == 0 ? boxed[rnd().nextInt(boxed.length)] : sameType(x, boxed);
            emit("cmpboxed", args -> {
                args.add(boxLit(x));
                args.add(boxLit(y));
                @SuppressWarnings({"unchecked", "rawtypes"}) int c = ((Comparable) x).compareTo(y);
                return Integer.toString(c);
            });
        }
        emit("setEqualsCCE", args -> CollVectors.bool(new TreeSet<Object>(List.of(1, 2)).equals(new TreeSet<Object>(List.of("a", "b")))));
        emit("setEqualsNPE", args -> CollVectors.bool(new TreeSet<Object>(List.of(1)).equals(Collections.singleton(null))));
        emit("mapEqualsCCE", args -> {
            TreeMap<Object, Object> a = new TreeMap<>(), b = new TreeMap<>();
            a.put(1, 1);
            b.put("a", 1);
            return CollVectors.bool(a.equals(b));
        });
        // Special lists, each with random operations.
        String[] kinds = {"singletonList", "emptyList", "nCopies", "unmodifiableList", "unmodifiableLL", "singleton", "emptySet", "unmodifiableSet", "unmodifiableCollection"};
        for (int i = 0; i < 126; i++) {
            String kind = kinds[i % kinds.length];
            CollVectors.elemType = rnd().nextInt(3);
            CollVectors.nulls = rnd().nextInt(4) == 0;
            Collection c;
            List<String> cargs = new ArrayList<>();
            switch (kind) {
                case "singletonList": { Object x = CollVectors.v(); cargs.add(r(x)); c = Collections.singletonList(x); break; }
                case "emptyList": c = Collections.emptyList(); break;
                case "nCopies": { int n = rnd().nextInt(6); Object x = CollVectors.v(); cargs.add("" + n); cargs.add(r(x)); c = Collections.nCopies(n, x); break; }
                case "unmodifiableList": { List<Object> l = CollVectors.vals(6); cargs.add("AL" + lit(l)); c = Collections.unmodifiableList(new ArrayList<>(l)); break; }
                case "unmodifiableLL": { List<Object> l = CollVectors.vals(6); cargs.add("LL" + lit(l)); c = Collections.unmodifiableList(new LinkedList<>(l)); break; }
                case "singleton": { Object x = CollVectors.v(); cargs.add(r(x)); c = Collections.singleton(x); break; }
                case "emptySet": c = Collections.emptySet(); break;
                case "unmodifiableSet": { List<Object> l = CollVectors.vals(6); l.removeIf(Objects::isNull); cargs.add("TS" + lit(l)); c = Collections.unmodifiableSet(new TreeSet<>(l)); break; }
                default: { List<Object> l = CollVectors.vals(6); cargs.add("AL" + lit(l)); c = Collections.unmodifiableCollection(new ArrayList<>(l)); break; }
            }
            CollVectors.state = c;
            CollVectors.identityHash = kind.equals("unmodifiableCollection");
            CollVectors.newLine().append("new ").append(kind).append(' ').append(new String[]{"int", "item", "str"}[CollVectors.elemType]);
            for (String a : cargs) CollVectors.out.append(' ').append(a);
            CollVectors.out.append('\n');
            int nops = 10 + rnd().nextInt(20);
            for (int k = 0; k < nops; k++) {
                if (c instanceof List l && rnd().nextBoolean()) {
                    int x = rnd().nextInt(CollVectors.LIST_OPS);
                    if (x == 10) x = 1;
                    CollVectors.listOp(l, x);
                } else {
                    CollVectors.collectionOp(c, rnd().nextInt(CollVectors.COLLECTION_OPS));
                }
            }
            CollVectors.identityHash = false;
        }
        // The unmodifiable map and the empty map.
        for (int i = 0; i < 20; i++) {
            CollVectors.elemType = rnd().nextInt(3);
            CollVectors.nulls = false;
            List<String> cargs = new ArrayList<>();
            Map m;
            String kind;
            if (i % 2 == 0) {
                kind = "unmodifiableMap";
                Map src = CollVectors.mapArg(cargs);
                m = Collections.unmodifiableMap(src);
            } else {
                kind = "emptyMap";
                m = Collections.emptyMap();
            }
            CollVectors.state = m;
            CollVectors.newLine().append("new ").append(kind).append(' ').append(new String[]{"int", "item", "str"}[CollVectors.elemType]);
            for (String a : cargs) CollVectors.out.append(' ').append(a);
            CollVectors.out.append('\n');
            int nops = 10 + rnd().nextInt(20);
            for (int k = 0; k < nops; k++) {
                int x = rnd().nextInt(15);
                Map mm = m;
                switch (x) {
                    case 0: CollVectors.emit("put", a -> { Object key = CollVectors.v(); a.add(r(key)); a.add("i1"); return r(mm.put(key, 1)); }); break;
                    case 1: CollVectors.emit("get", a -> { Object key = CollVectors.v(); a.add(r(key)); return r(mm.get(key)); }); break;
                    case 2: CollVectors.emit("mremove", a -> { Object key = CollVectors.v(); a.add(r(key)); return r(mm.remove(key)); }); break;
                    case 3: CollVectors.emit("containsKey", a -> { Object key = CollVectors.v(); a.add(r(key)); return CollVectors.bool(mm.containsKey(key)); }); break;
                    case 4: CollVectors.emit("msize", a -> CollVectors.num(mm.size())); break;
                    case 5: CollVectors.emit("getOrDefault", a -> { Object key = CollVectors.v(); a.add(r(key)); a.add("i7"); return r(mm.getOrDefault(key, 7)); }); break;
                    case 6: CollVectors.emit("putIfAbsent", a -> { Object key = CollVectors.v(); a.add(r(key)); a.add("i2"); return r(mm.putIfAbsent(key, 2)); }); break;
                    case 7: CollVectors.emit("mtostring", a -> CollVectors.toStr(mm)); break;
                    case 8: CollVectors.emit("mhash", a -> CollVectors.num(mm.hashCode())); break;
                    case 9: CollVectors.emit("mequals", a -> CollVectors.bool(mm.equals(CollVectors.mapArg(a)))); break;
                    case 10: CollVectors.emit("entries", a -> { String p = CollVectors.randomIterProg(mm.size(), false, true, false); a.add(p); return CollVectors.iterProg(mm.entrySet().iterator(), p, mm); }); break;
                    case 11: CollVectors.emit("mclear", a -> { mm.clear(); return CollVectors.OK; }); break;
                    case 13: CollVectors.emit("esops", a -> CollVectors.esOps(mm, a)); break;
                    case 14: CollVectors.emit("misEmpty", a -> CollVectors.bool(mm.isEmpty())); break;
                    default: CollVectors.emit("putAll", a -> { mm.putAll(CollVectors.mapArg(a)); return CollVectors.OK; }); break;
                }
            }
        }
    }
}
