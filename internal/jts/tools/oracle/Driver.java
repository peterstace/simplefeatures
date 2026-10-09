import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import org.locationtech.jts.algorithm.ConvexHull;
import org.locationtech.jts.algorithm.distance.DiscreteHausdorffDistance;
import org.locationtech.jts.geom.Coordinate;
import org.locationtech.jts.geom.CoordinateSequence;
import org.locationtech.jts.geom.Geometry;
import org.locationtech.jts.geom.GeometryCollection;
import org.locationtech.jts.geom.GeometryFactory;
import org.locationtech.jts.geom.LineString;
import org.locationtech.jts.geom.Point;
import org.locationtech.jts.geom.Polygon;
import org.locationtech.jts.geom.PrecisionModel;
import org.locationtech.jts.geom.prep.PreparedGeometry;
import org.locationtech.jts.geom.prep.PreparedGeometryFactory;
import org.locationtech.jts.io.WKBReader;
import org.locationtech.jts.io.WKTReader;
import org.locationtech.jts.operation.buffer.BufferOp;
import org.locationtech.jts.operation.buffer.BufferParameters;
import org.locationtech.jts.operation.distance.DistanceOp;
import org.locationtech.jts.operation.overlayng.OverlayNGRobust;
import org.locationtech.jts.operation.relateng.RelateNG;
import org.locationtech.jts.operation.relateng.RelatePredicate;
import org.locationtech.jts.operation.valid.IsSimpleOp;
import org.locationtech.jts.operation.valid.IsValidOp;
import org.locationtech.jts.simplify.DouglasPeuckerSimplifier;
import org.locationtech.jts.simplify.TopologyPreservingSimplifier;

/**
 * The JVM side of the oracle: runs JTS operations requested on stdin and
 * prints canonical results on stdout, one line per request.
 *
 * <p>Requests are lines of space-separated fields:
 *
 * <ul>
 *   <li>{@code pm FLOATING | FLOATING_SINGLE | FIXED <scale bits>} sets the
 *       precision model for subsequent loads.
 *   <li>{@code load <name> <hex WKB>} parses and stores a geometry.
 *   <li>{@code wkt <name> <WKT with spaces as '_'>} parses and stores a
 *       geometry from WKT, replying with its hex WKB.
 *   <li>{@code op <op> <args...>} runs an operation. Geometry arguments are
 *       stored names; doubles are raw bits in hex; ints are decimal.
 *   <li>{@code reset} forgets stored geometries.
 * </ul>
 *
 * <p>Each reply to {@code op} starts with the identity hash code that the next
 * newly hashed object will receive minus one, so that the Go side can start its
 * identity hash counter at the same value. The JVM must run with
 * -XX:hashCode=3 (sequential identity hash codes) and
 * -XX:-InlineMathNatives (Math matches StrictMath).
 */
public final class Driver {
    private final Map<String, Geometry> geoms = new HashMap<>();
    private GeometryFactory factory = new GeometryFactory();

    public static void main(String[] args) throws Exception {
        PrintStream out = new PrintStream(System.out, false, StandardCharsets.UTF_8);
        // JTS diagnostics must not corrupt the protocol.
        System.setOut(System.err);
        warmUp(java.nio.file.Path.of(args[0]));
        Driver d = new Driver();
        BufferedReader in = new BufferedReader(new InputStreamReader(System.in, StandardCharsets.UTF_8));
        String line;
        while ((line = in.readLine()) != null) {
            String reply;
            try {
                reply = d.handle(line.split(" "));
            } catch (Throwable t) {
                reply = "ERR " + t;
            }
            out.println(reply);
            out.flush();
        }
    }

    /**
     * Loads and initializes every JTS class and runs each operation once, so
     * that class loading and static initialization (which consume identity
     * hash codes) happen before any measured operation.
     */
    private static void warmUp(java.nio.file.Path classesDir) throws Exception {
        try (java.util.stream.Stream<java.nio.file.Path> s = java.nio.file.Files.walk(classesDir)) {
            for (java.nio.file.Path p : (Iterable<java.nio.file.Path>) s::iterator) {
                String rel = classesDir.relativize(p).toString();
                if (!rel.endsWith(".class")) {
                    continue;
                }
                String name = rel.substring(0, rel.length() - ".class".length()).replace('/', '.');
                try {
                    Class.forName(name, true, Driver.class.getClassLoader());
                } catch (Throwable t) {
                    // Classes needing absent dependencies (e.g. java.awt) are irrelevant.
                }
            }
        }
        Driver d = new Driver();
        String[][] battery = {
            {"wkt", "a", "POLYGON_((0_0,_10_0,_10_10,_0_10,_0_0),_(1_1,_2_1,_2_2,_1_1))"},
            {"wkt", "b", "MULTIPOLYGON_(((5_5,_15_5,_15_15,_5_15,_5_5)))"},
            {"wkt", "c", "GEOMETRYCOLLECTION_(POINT_(1_2),_LINESTRING_(0_0,_3_4,_1_9))"},
        };
        for (String[] cmd : battery) {
            d.handle(cmd);
        }
        for (String op : new String[] {"relate", "overlay", "unaryUnion", "buffer", "isValid", "isSimple", "area",
                "length", "centroid", "interiorPoint", "convexHull", "boundary", "envelope", "normalize", "distance",
                "hausdorff", "simplifyDP", "simplifyTP", "dump", "relatePredicate", "prepared"}) {
            for (String[] args : new String[][] {{"a", "b"}, {"b", "c"}, {"c", "a"}}) {
                String[] cmd;
                switch (op) {
                    case "overlay": cmd = new String[] {"op", op, args[0], args[1], "1"}; break;
                    case "buffer": cmd = new String[] {"op", op, args[0], "3ff0000000000000", "8", "1", "1",
                            "4014000000000000"}; break;
                    case "simplifyDP": case "simplifyTP": cmd = new String[] {"op", op, args[0],
                            "3ff0000000000000"}; break;
                    case "relatePredicate": case "prepared": cmd = new String[] {"op", op, args[0], args[1],
                            "intersects"}; break;
                    default: cmd = new String[] {"op", op, args[0], args[1]};
                }
                d.handle(cmd);
            }
        }
    }

    private String handle(String[] f) throws Exception {
        switch (f[0]) {
            case "pm":
                factory = new GeometryFactory(precisionModel(f));
                return "OK";
            case "load":
                geoms.put(f[1], new WKBReader(factory).read(WKBReader.hexToBytes(f[2])));
                return "OK";
            case "wkt": {
                Geometry g = new WKTReader(factory).read(f[2].replace('_', ' '));
                geoms.put(f[1], g);
                return "OK " + org.locationtech.jts.io.WKBWriter.toHex(
                        new org.locationtech.jts.io.WKBWriter(dimension(g)).write(g));
            }
            case "reset":
                geoms.clear();
                return "OK";
            case "op": {
                int hseq = new Object().hashCode();
                try {
                    return hseq + " OK " + op(f);
                } catch (Throwable t) {
                    return hseq + " EXC " + t.getClass().getName();
                }
            }
            default:
                return "ERR unknown command";
        }
    }

    private static PrecisionModel precisionModel(String[] f) {
        switch (f[1]) {
            case "FLOATING":
                return new PrecisionModel();
            case "FLOATING_SINGLE":
                return new PrecisionModel(PrecisionModel.FLOATING_SINGLE);
            case "FIXED":
                return new PrecisionModel(dbl(f[2]));
            default:
                throw new IllegalArgumentException("precision model " + f[1]);
        }
    }

    private static int dimension(Geometry g) {
        boolean z = false;
        for (Coordinate c : g.getCoordinates()) {
            if (!Double.isNaN(c.getZ())) {
                z = true;
            }
        }
        return z ? 3 : 2;
    }

    private Geometry g(String name) {
        Geometry g = geoms.get(name);
        if (g == null) {
            throw new IllegalArgumentException("no geometry " + name);
        }
        return g;
    }

    private static double dbl(String hexBits) {
        return Double.longBitsToDouble(Long.parseUnsignedLong(hexBits, 16));
    }

    private static String bits(double d) {
        return Long.toHexString(Double.doubleToRawLongBits(d));
    }

    private String op(String[] f) {
        String op = f[1];
        switch (op) {
            case "relate":
                return RelateNG.relate(g(f[2]), g(f[3])).toString();
            case "relatePredicate":
                return Boolean.toString(RelateNG.relate(g(f[2]), g(f[3]), predicate(f[4])));
            case "prepared":
                return Boolean.toString(prepared(PreparedGeometryFactory.prepare(g(f[2])), g(f[3]), f[4]));
            case "overlay":
                return dump(OverlayNGRobust.overlay(g(f[2]), g(f[3]), Integer.parseInt(f[4])));
            case "unaryUnion":
                return dump(OverlayNGRobust.union(g(f[2])));
            case "buffer": {
                BufferParameters p = new BufferParameters(Integer.parseInt(f[4]), Integer.parseInt(f[5]),
                        Integer.parseInt(f[6]), dbl(f[7]));
                return dump(BufferOp.bufferOp(g(f[2]), dbl(f[3]), p));
            }
            case "isValid":
                return Boolean.toString(IsValidOp.isValid(g(f[2])));
            case "isSimple":
                return Boolean.toString(IsSimpleOp.isSimple(g(f[2])));
            case "area":
                return bits(g(f[2]).getArea());
            case "length":
                return bits(g(f[2]).getLength());
            case "centroid":
                return dump(g(f[2]).getCentroid());
            case "interiorPoint":
                return dump(g(f[2]).getInteriorPoint());
            case "convexHull":
                return dump(new ConvexHull(g(f[2])).getConvexHull());
            case "boundary":
                return dump(g(f[2]).getBoundary());
            case "envelope":
                return dump(g(f[2]).getEnvelope());
            case "normalize": {
                Geometry c = g(f[2]).copy();
                c.normalize();
                return dump(c);
            }
            case "distance":
                return bits(DistanceOp.distance(g(f[2]), g(f[3])));
            case "hausdorff":
                return bits(DiscreteHausdorffDistance.distance(g(f[2]), g(f[3])));
            case "simplifyDP":
                return dump(DouglasPeuckerSimplifier.simplify(g(f[2]), dbl(f[3])));
            case "simplifyTP":
                return dump(TopologyPreservingSimplifier.simplify(g(f[2]), dbl(f[3])));
            case "dump":
                return dump(g(f[2]));
            default:
                throw new IllegalArgumentException("unknown op " + op);
        }
    }

    private static org.locationtech.jts.operation.relateng.TopologyPredicate predicate(String name) {
        switch (name) {
            case "intersects": return RelatePredicate.intersects();
            case "contains": return RelatePredicate.contains();
            case "within": return RelatePredicate.within();
            case "covers": return RelatePredicate.covers();
            case "coveredBy": return RelatePredicate.coveredBy();
            case "crosses": return RelatePredicate.crosses();
            case "disjoint": return RelatePredicate.disjoint();
            case "equalsTopo": return RelatePredicate.equalsTopo();
            case "overlaps": return RelatePredicate.overlaps();
            case "touches": return RelatePredicate.touches();
            default: throw new IllegalArgumentException("predicate " + name);
        }
    }

    private static boolean prepared(PreparedGeometry p, Geometry g, String name) {
        switch (name) {
            case "intersects": return p.intersects(g);
            case "contains": return p.contains(g);
            case "containsProperly": return p.containsProperly(g);
            case "covers": return p.covers(g);
            case "coveredBy": return p.coveredBy(g);
            case "crosses": return p.crosses(g);
            case "disjoint": return p.disjoint(g);
            case "overlaps": return p.overlaps(g);
            case "touches": return p.touches(g);
            case "within": return p.within(g);
            default: throw new IllegalArgumentException("prepared predicate " + name);
        }
    }

    /**
     * A canonical text form of a geometry that exposes every detail the Go
     * port must reproduce: classes, structure, and the bits of every ordinate.
     */
    static String dump(Geometry g) {
        StringBuilder sb = new StringBuilder();
        dump(g, sb);
        return sb.toString();
    }

    private static void dump(Geometry g, StringBuilder sb) {
        if (g == null) {
            sb.append("null");
            return;
        }
        sb.append(g.getClass().getSimpleName()).append('{');
        if (g instanceof GeometryCollection) {
            for (int i = 0; i < g.getNumGeometries(); i++) {
                if (i > 0) {
                    sb.append(',');
                }
                dump(g.getGeometryN(i), sb);
            }
        } else if (g instanceof Polygon) {
            Polygon p = (Polygon) g;
            dump(p.getExteriorRing(), sb);
            for (int i = 0; i < p.getNumInteriorRing(); i++) {
                sb.append(',');
                dump(p.getInteriorRingN(i), sb);
            }
        } else if (g instanceof LineString) {
            dump(((LineString) g).getCoordinateSequence(), sb);
        } else if (g instanceof Point) {
            dump(((Point) g).getCoordinateSequence(), sb);
        }
        sb.append('}');
    }

    private static void dump(CoordinateSequence seq, StringBuilder sb) {
        sb.append(seq.getClass().getSimpleName()).append('/').append(seq.getDimension()).append('/')
                .append(seq.getMeasures()).append('[');
        for (int i = 0; i < seq.size(); i++) {
            if (i > 0) {
                sb.append(';');
            }
            Coordinate c = seq.getCoordinate(i);
            sb.append(c.getClass().getSimpleName()).append(':').append(bits(c.getX())).append(',')
                    .append(bits(c.getY())).append(',').append(bits(c.getZ())).append(',').append(bits(c.getM()));
        }
        sb.append(']');
    }
}
