/*
 * TextVectors writes test vectors for the Go implementation of the JDK's
 * text and number handling in internal/jts/java (the text_*.go files):
 * Double/Float toString and parsing, DecimalFormat, String.format, integer
 * parsing and formatting, Character properties, String case conversion and
 * regular expressions, StringBuilder, BigDecimal and Locale. Regenerate the
 * vectors from the repository root with:
 *
 *   java internal/jts/tools/jdkvectors/TextVectors.java internal/jts/java/testdata
 *
 * It writes text_tostring.txt, text_parse.txt, text_decimalformat.txt,
 * text_format.txt and text_misc.txt. An optional second argument multiplies
 * the number of digest batches, for a more thorough local check: write the
 * vectors to a temporary directory and run the Go test with
 * TEXT_VECTORS_DIR set to it. The committed files use the default of 1.
 *
 * Each line is a tab-separated record whose first field names the check.
 * Strings are escaped: '\' as "\\" and every UTF-16 unit outside printable
 * ASCII (and tab) as "\\uXXXX". Doubles and floats are their raw IEEE 754
 * bits in hex. A thrown exception is written as "!" followed by its class
 * name, ':' and its escaped message.
 *
 * Besides explicit vectors, the large random checks are written as digests:
 * the inputs come from a SplitMix64 generator that the Go test reimplements,
 * and each "digest" record holds the FNV-1a 64-bit hash of the UTF-8 outputs
 * of one batch, each followed by '\n'. On a mismatch in a dd, fd, pdd or dfd
 * record, run this program with the arguments "dump <check> <batch>" (for
 * dfd, the check is "dfd:<setup>") to print the batch's inputs and outputs.
 *
 * The output is deterministic.
 */

import java.io.IOException;
import java.io.PrintWriter;
import java.math.BigDecimal;
import java.math.MathContext;
import java.math.RoundingMode;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.text.NumberFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Locale;
import java.util.function.Function;

public class TextVectors {
    public static void main(String[] args) throws IOException {
        Locale.setDefault(Locale.US);
        if (args.length == 3 && args[0].equals("dump")) {
            dump(args[1], Integer.parseInt(args[2]));
            return;
        }
        Path dir = Path.of(args[0]);
        if (args.length > 1) {
            scale = Integer.parseInt(args[1]);
        }
        try (PrintWriter w = writer(dir.resolve("text_tostring.txt"))) {
            toStringVectors(w);
        }
        try (PrintWriter w = writer(dir.resolve("text_parse.txt"))) {
            parseVectors(w);
        }
        try (PrintWriter w = writer(dir.resolve("text_decimalformat.txt"))) {
            decimalFormatVectors(w);
        }
        try (PrintWriter w = writer(dir.resolve("text_format.txt"))) {
            formatVectors(w);
        }
        try (PrintWriter w = writer(dir.resolve("text_misc.txt"))) {
            miscVectors(w);
        }
    }

    static int scale = 1;

    static PrintWriter writer(Path p) throws IOException {
        return new PrintWriter(Files.newBufferedWriter(p, StandardCharsets.UTF_8));
    }

    // ------------------------------------------------------------------
    // Shared helpers (mirrored by the Go test)

    /** SplitMix64. */
    static final class Rng {
        long state;

        Rng(long seed) {
            state = seed;
        }

        long next() {
            long z = (state += 0x9e3779b97f4a7c15L);
            z = (z ^ (z >>> 30)) * 0xbf58476d1ce4e5b9L;
            z = (z ^ (z >>> 27)) * 0x94d049bb133111ebL;
            return z ^ (z >>> 31);
        }

        /** Uniform in [0, n) for n > 0 (n small). */
        int below(int n) {
            return (int) Long.remainderUnsigned(next(), n);
        }
    }

    static final long FNV_OFFSET = 0xcbf29ce484222325L;
    static final long FNV_PRIME = 0x100000001b3L;

    static long fnv(long h, String s) {
        for (byte b : s.getBytes(StandardCharsets.UTF_8)) {
            h ^= (b & 0xff);
            h *= FNV_PRIME;
        }
        h ^= '\n';
        h *= FNV_PRIME;
        return h;
    }

    static final double[] POW10 = {1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13,
        1e14, 1e15, 1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22};

    /** A double from a mix of distributions. */
    static double genDouble(Rng r) {
        switch (r.below(8)) {
            case 0:
            case 1:
                return Double.longBitsToDouble(r.next());
            case 2: {
                // A short decimal k / 10^j.
                long k = r.next() >>> (64 - 1 - r.below(40));
                return k / POW10[r.below(23)];
            }
            case 3: {
                // A short decimal k * 10^j.
                long k = r.next() >>> (64 - 1 - r.below(30));
                return k * POW10[r.below(23)];
            }
            case 4:
                // Subnormal.
                return Double.longBitsToDouble(r.next() >>> 12);
            case 5: {
                // Near a power of two.
                int e = r.below(2046) + 1;
                long bits = ((long) e << 52) + (r.below(5) - 2);
                return Double.longBitsToDouble(bits & 0x7fffffffffffffffL);
            }
            case 6: {
                // An ordinate-like value.
                long k = r.next() >>> (64 - 1 - r.below(50));
                double v = k / POW10[r.below(16)];
                return (r.next() & 1) == 0 ? v : -v;
            }
            default: {
                // Small integers and halves.
                long k = r.below(2000000) - 1000000;
                return k / (double) (1 << r.below(12));
            }
        }
    }

    static float genFloat(Rng r) {
        switch (r.below(4)) {
            case 0:
            case 1:
                return Float.intBitsToFloat((int) r.next());
            case 2:
                return (float) genDouble(r);
            default: {
                int k = r.below(2000000) - 1000000;
                return k / (float) (1 << r.below(20));
            }
        }
    }

    static final String DIGITS = "0123456789";

    /** A decimal string, mostly valid, sometimes long or near a tie. */
    static String genDecimalString(Rng r) {
        StringBuilder sb = new StringBuilder();
        if (r.below(4) == 0) {
            sb.append(r.below(2) == 0 ? '-' : '+');
        }
        int mode = r.below(6);
        int nd = mode == 0 ? 1 + r.below(60) : mode == 1 ? 17 + r.below(30) : 1 + r.below(20);
        int point = r.below(nd + 2) - 1;
        for (int i = 0; i < nd; i++) {
            if (i == point) {
                sb.append('.');
            }
            sb.append(DIGITS.charAt(r.below(10)));
        }
        if (mode == 2) {
            // A double's exact halfway point is long; approximate one.
            for (int i = 0; i < 20; i++) {
                sb.append(i == 0 ? '5' : '0');
            }
            sb.append(DIGITS.charAt(r.below(10)));
        }
        if (r.below(3) > 0) {
            sb.append(r.below(2) == 0 ? 'e' : 'E');
            int e = r.below(5) == 0 ? r.below(700) - 350 : r.below(80) - 40;
            sb.append(e);
        }
        if (r.below(10) == 0) {
            sb.append("fFdD".charAt(r.below(4)));
        }
        return sb.toString();
    }

    static String esc(String s) {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            if (c == '\\') {
                sb.append("\\\\");
            } else if (c >= 0x20 && c < 0x7f) {
                sb.append(c);
            } else {
                sb.append(String.format("\\u%04x", (int) c));
            }
        }
        return sb.toString();
    }

    static String dbits(double d) {
        return String.format("%016x", Double.doubleToRawLongBits(d));
    }

    static String fbits(float f) {
        return String.format("%08x", Float.floatToRawIntBits(f));
    }

    static String exc(Throwable t) {
        String m = t.getMessage();
        return "!" + t.getClass().getName() + ":" + (m == null ? "null" : esc(m));
    }

    interface Thrower<T> {
        T get() throws Exception;
    }

    static <T> String attempt(Thrower<T> f, Function<T, String> render) {
        try {
            return render.apply(f.get());
        } catch (Exception e) {
            return exc(e);
        }
    }

    static void line(PrintWriter w, String... fields) {
        w.println(String.join("\t", fields));
    }

    // ------------------------------------------------------------------
    // Double.toString and Float.toString

    static final int TOSTRING_BATCHES = 400;
    static final int BATCH = 1000;

    static double[] edgeDoubles() {
        List<Double> l = new ArrayList<>(Arrays.asList(0.0, -0.0, Double.MIN_VALUE, Double.MIN_NORMAL,
                Double.MAX_VALUE, Double.NaN, Double.POSITIVE_INFINITY, Double.NEGATIVE_INFINITY, 1.0, -1.0, 0.1,
                0.2, 0.3, 1.0 / 3, 2.0 / 3, 1e7, 1e7 - 1, 9999999.999999998, 1e-3, 0.0009999999999999998, 1e-4,
                1e21, 1e22, 1e23, 2e23, 5e-324, 1e-323, 1.5e-323, 2.0e-323, 9007199254740992.0, 9007199254740993.0,
                123456789012345678.0, 1.7976931348623157e308, 2.2250738585072014E-308, 2.225073858507201E-308,
                4.35, 2.675, 1.005, 0.5, 0.25, 100.0, 1000000.0, 12345678.9, 0.001, 0.01, 2e-3, 1.0E-5,
                Math.PI, Math.E, 3.0e10, 1.234e-7, 5.0E-324 * 3));
        for (int e = -325; e <= 309; e++) {
            l.add(Double.parseDouble("1e" + e));
            l.add(Double.parseDouble("5e" + e));
            l.add(Double.parseDouble("9.999999999999999e" + e));
        }
        for (int e = -1074; e <= 1023; e += 7) {
            l.add(Math.scalb(1.0, e));
        }
        double[] out = new double[l.size()];
        for (int i = 0; i < out.length; i++) {
            out[i] = l.get(i);
        }
        return out;
    }

    static void toStringVectors(PrintWriter w) {
        for (double d : edgeDoubles()) {
            line(w, "d", dbits(d), esc(Double.toString(d)));
        }
        Rng r = new Rng(1);
        for (int i = 0; i < 4000; i++) {
            double d = genDouble(r);
            line(w, "d", dbits(d), esc(Double.toString(d)));
        }
        float[] edgeFloats = {0f, -0f, Float.MIN_VALUE, Float.MIN_NORMAL, Float.MAX_VALUE, Float.NaN,
            Float.POSITIVE_INFINITY, Float.NEGATIVE_INFINITY, 1f, 0.1f, 0.3f, 1e7f, 1e-3f, 1e10f, 1.0E-45f,
            3.4028235E38f, 1.17549435E-38f, 16777216f, 16777217f, 0.33333334f, 2e-45f, 7e-45f, 8e-45f};
        for (float f : edgeFloats) {
            line(w, "f", fbits(f), esc(Float.toString(f)));
        }
        for (int i = 0; i < 2000; i++) {
            float f = genFloat(r);
            line(w, "f", fbits(f), esc(Float.toString(f)));
        }
        for (int b = 0; b < TOSTRING_BATCHES * scale; b++) {
            line(w, "dd", Integer.toString(b), Long.toHexString(digestToString(b, false)));
        }
        for (int b = 0; b < TOSTRING_BATCHES / 4 * scale; b++) {
            line(w, "fd", Integer.toString(b), Long.toHexString(digestToString(b, true)));
        }
    }

    static long digestToString(int batch, boolean isFloat) {
        Rng r = new Rng((isFloat ? 0x2000000L : 0x1000000L) + batch);
        long h = FNV_OFFSET;
        for (int i = 0; i < BATCH; i++) {
            h = fnv(h, isFloat ? Float.toString(genFloat(r)) : Double.toString(genDouble(r)));
        }
        return h;
    }

    // ------------------------------------------------------------------
    // Parsing

    static final String[] PARSE_CASES = {
        "0", "-0", "+0", "0.0", "-0.0", "00000.000e5", "1", "-1", "1.", ".1", "-.1", "1e5", "1E5", "1e+5", "1e-5",
        "1.5e300", "1e309", "-1e309", "1e-400", "-1e-400", "NaN", "-NaN", "+NaN", "Infinity", "-Infinity",
        "+Infinity", "Infinityx", "nan", "infinity", "inf", "  1.5  ", "\t1.5\n", "1.5f", "1.5F", "1.5d", "1.5D",
        "1.5ff", "1.5fd", "1e5f", "1.e5", ".e5", ".", "", " ", "-", "+", "e5", "1e", "1e+", "1e-", "1.2.3",
        "1..2", "0x1p0", "0x1.8p1", "-0x1.8p1", "0X1P0", "0x.8p0", "0x1p", "0x1", "0xp1", "0x1.p1", "0x.p1",
        "0x1p1f", "0x1p1d", "0x1p-1074", "0x1p-1075", "0x1.0000000000001p-1075", "0x1p1024", "0x1.fffffffffffffp1023",
        "0x1.fffffffffffff8p1023", "0x1.fffffffffffff7ffp1023", "0x1p-149", "0x1p-150", "0x1.000001p-150",
        "0x1.fffffep127", "0x1.ffffffp127", "0x1.fffffefp127", "0x123456789abcdefp-20", "0x0p0", "0x0.0p99999999999",
        "0x1p2147483647", "0x1p2147483648", "0x1p-2147483648", "-0x1p99999999999", "0x1p+5", "0x1p-0",
        "1e2147483647", "1e2147483648", "1e-2147483648", "1e99999999999999", "0.0000000000000000000000001e25",
        "4.9e-324", "2.4703282292062327e-324", "2.4703282292062328e-324", "2.47032822920623272e-324",
        "1.7976931348623157e308", "1.7976931348623158e308", "1.7976931348623159e308", "1.797693134862315807e308",
        "2.2250738585072011e-308", "2.2250738585072012e-308", "9007199254740993", "9007199254740993.0000000001",
        "0.1", "0.2", "0.3", "123456789012345678901234567890", "1e23", "8.5e-323", "1.4e-45", "7.0e-46",
        "7.006492321624085e-46", "7.006492321624086e-46", "3.4028235e38", "3.4028236e38", "3.40282356779733661637539395458142568448e38",
        "1.00000017881393432617187499", "1.000000178813934326171875", "1.00000017881393432617187501",
        "0x", "0x1.8", "1_000", "1,5", "１２", "1e５", "0.0e", "--1", "+-1", "1-", "1e1.5", " 1", "1 ",
    };

    static void parseVectors(PrintWriter w) {
        for (String s : PARSE_CASES) {
            parseLine(w, s);
        }
        // Long digit strings with a sticky tail.
        StringBuilder sb = new StringBuilder("2.4703282292062327208828439643411068618252990130716238221279284125033775363510437593264991818081799618989828234772285886546332835517796989819938739800539093906315035659515570226392290858392449105184435931802849936536152500319370457678249219365623669863658480757001585769269903706311928279558551332927834338409351978015531246597263579574622766465272827220056374006485499977096599470454020828166226237857393450736339007967761930577506740176324673600968951340535537458516661134223766678604162159680461914467291840300530057530849048765391711386591646239524912623653881879636239373280423891018672348497668235089863388587925628302755995657524455507255189313690836254779186948667994968324049705821028513185451396213837722826145437693412532098591327667236328125e-324");
        parseLine(w, sb.toString());
        parseLine(w, sb.toString().replace("125e-324", "125000000000001e-324"));
        parseLine(w, sb.toString().replace("125e-324", "124999999999999e-324"));
        Rng r = new Rng(7);
        for (int i = 0; i < 3000; i++) {
            parseLine(w, genDecimalString(r));
        }
        for (int b = 0; b < 100 * scale; b++) {
            line(w, "pdd", Integer.toString(b), Long.toHexString(digestParse(b)));
        }
        for (int b = 0; b < 20 * scale; b++) {
            line(w, "phd", Integer.toString(b), Long.toHexString(digestParseHex(b)));
        }
    }

    static final String HEX_DIGITS = "0123456789abcdefABCDEF";

    /** A hexadecimal floating-point string. */
    static String genHexString(Rng r) {
        StringBuilder sb = new StringBuilder();
        if (r.below(3) == 0) {
            sb.append('-');
        }
        sb.append(r.below(2) == 0 ? "0x" : "0X");
        int nd = 1 + r.below(r.below(4) == 0 ? 40 : 16);
        int point = r.below(nd + 2) - 1;
        for (int i = 0; i < nd; i++) {
            if (i == point) {
                sb.append('.');
            }
            sb.append(HEX_DIGITS.charAt(r.below(HEX_DIGITS.length())));
        }
        sb.append(r.below(2) == 0 ? 'p' : 'P');
        int e = r.below(3) == 0 ? r.below(2400) - 1200 : r.below(300) - 150;
        sb.append(e);
        return sb.toString();
    }

    static long digestParseHex(int batch) {
        Rng r = new Rng(0x6000000L + batch);
        long h = FNV_OFFSET;
        for (int i = 0; i < BATCH; i++) {
            String s = genHexString(r);
            h = fnv(h, attempt(() -> Double.parseDouble(s), TextVectors::dbits));
            h = fnv(h, attempt(() -> Float.parseFloat(s), TextVectors::fbits));
        }
        return h;
    }

    static void parseLine(PrintWriter w, String s) {
        line(w, "pd", esc(s), attempt(() -> Double.parseDouble(s), TextVectors::dbits));
        line(w, "pf", esc(s), attempt(() -> Float.parseFloat(s), TextVectors::fbits));
    }

    static long digestParse(int batch) {
        Rng r = new Rng(0x3000000L + batch);
        long h = FNV_OFFSET;
        for (int i = 0; i < BATCH; i++) {
            String s = genDecimalString(r);
            h = fnv(h, attempt(() -> Double.parseDouble(s), TextVectors::dbits));
            h = fnv(h, attempt(() -> Float.parseFloat(s), TextVectors::fbits));
        }
        return h;
    }

    // ------------------------------------------------------------------
    // DecimalFormat

    /** The formats under test, named as the Go test names them. */
    static final String[] DF_SETUPS = {
        "ordinate:325", "ordinate:0", "ordinate:1", "ordinate:2", "ordinate:3", "ordinate:4", "ordinate:5",
        "ordinate:6", "ordinate:8", "ordinate:10", "ordinate:15", "ordinate:16", "ordinate:17", "ordinate:18",
        "ordinate:20", "us", "root", "us-obj", "pat:#,##0.###", "pat:0.00", "pat:#.#", "pat:0.###E0",
        "pat:##0.##E0", "pat:00.0E00", "pat:0.0##%", "pat:#,##0.00;(#,##0.00)", "pat:'#'0.0'x'", "pat:0",
        "pat:#,###", "pat:.##", "pat:#.", "pat:0.0‰", "pat:", "pat:#,##,##0.0#", "de", "tr", "fr",
        "int:us", "pct:us", "pat:0.##;-0.##", "mode:HALF_UP", "mode:DOWN", "mode:CEILING", "mode:FLOOR",
        "mode:UP", "mode:HALF_DOWN", "maxint:2", "minint:5", "minfrac:4",
    };

    static NumberFormat dfSetup(String setup) {
        String[] p = setup.split(":", 2);
        switch (p[0]) {
            case "ordinate": {
                DecimalFormat f = (DecimalFormat) NumberFormat.getInstance(Locale.US);
                f.applyPattern("0");
                f.setMaximumFractionDigits(Integer.parseInt(p[1]));
                return f;
            }
            case "us":
            case "us-obj":
                return NumberFormat.getInstance(Locale.US);
            case "root":
                return NumberFormat.getInstance(Locale.ROOT);
            case "de":
                return NumberFormat.getInstance(Locale.GERMANY);
            case "tr":
                return NumberFormat.getInstance(Locale.forLanguageTag("tr"));
            case "fr":
                return NumberFormat.getInstance(Locale.FRANCE);
            case "int":
                return NumberFormat.getIntegerInstance(Locale.US);
            case "pct":
                return NumberFormat.getPercentInstance(Locale.US);
            case "pat":
                return new DecimalFormat(p[1], DecimalFormatSymbols.getInstance(Locale.US));
            case "mode": {
                NumberFormat f = NumberFormat.getInstance(Locale.US);
                f.setRoundingMode(RoundingMode.valueOf(p[1]));
                f.setMaximumFractionDigits(2);
                return f;
            }
            case "maxint": {
                NumberFormat f = NumberFormat.getInstance(Locale.US);
                f.setMaximumIntegerDigits(Integer.parseInt(p[1]));
                return f;
            }
            case "minint": {
                NumberFormat f = NumberFormat.getInstance(Locale.US);
                f.setMinimumIntegerDigits(Integer.parseInt(p[1]));
                return f;
            }
            case "minfrac": {
                NumberFormat f = NumberFormat.getInstance(Locale.US);
                f.setMinimumFractionDigits(Integer.parseInt(p[1]));
                return f;
            }
        }
        throw new IllegalArgumentException(setup);
    }

    static String dfFormat(String setup, NumberFormat f, double d) {
        if (setup.equals("us-obj")) {
            return f.format((Object) d);
        }
        return f.format(d);
    }

    static double[] dfValues() {
        List<Double> l = new ArrayList<>(Arrays.asList(0.0, -0.0, Double.MIN_VALUE, Double.MIN_NORMAL,
                Double.MAX_VALUE, Double.NaN, Double.POSITIVE_INFINITY, Double.NEGATIVE_INFINITY, 0.1, 0.2, 0.3,
                1.0 / 3, 2.0 / 3, 2.675, 1.005, 100.0, Math.PI, 1e23, 9007199254740993.0));
        for (int e = -12; e <= 24; e += 3) {
            l.add(Double.parseDouble("1e" + e));
            l.add(Double.parseDouble("9.999999999999999e" + e));
        }
        double[] extra = {0.0005, 0.0015, 0.0025, 0.5, 1.5, 2.5, -2.5, 0.125, 0.375, 1.0005, 1.0015, 1.2345,
            1.2355, 9.9995, 999.9995, 1234567.8905, 0.045, 0.055, 1.115, 1.125, 2147483647.0, 2147483647.4,
            2147483647.5, 2147483648.0, 1e-6, 123456789.123, 12345678901234567890.0, 1.123456789012345,
            0.0123456789012345, 0.0000000000123456789012345, -1.123, 1234567890.0, 0.9995, 0.99949, 9.5,
            10.5, 1e15 + 0.3, 4.35, 0.035, 0.0035, 1.45, 1.55, 3.0005};
        for (double d : extra) {
            l.add(d);
            l.add(-d);
        }
        double[] out = new double[l.size()];
        for (int i = 0; i < out.length; i++) {
            out[i] = l.get(i);
        }
        return out;
    }

    static void decimalFormatVectors(PrintWriter w) {
        double[] values = dfValues();
        Rng r = new Rng(11);
        for (String setup : DF_SETUPS) {
            NumberFormat f = dfSetup(setup);
            for (double d : values) {
                line(w, "df", esc(setup), dbits(d), esc(dfFormat(setup, f, d)));
            }
            for (int i = 0; i < 40; i++) {
                double d = genDouble(r);
                line(w, "df", esc(setup), dbits(d), esc(dfFormat(setup, f, d)));
            }
            long[] longs = {0, 1, -1, 12, 1234, 1234567, Long.MAX_VALUE, Long.MIN_VALUE, -987654321098L};
            for (long v : longs) {
                line(w, "dfl", esc(setup), Long.toString(v), esc(f.format(v)));
            }
        }
        String[] badPatterns = {"0.0.0", "#0#", "0#", "##0,", ";0", "0;", "0E", "#E0", "0%%", "'0", "0.0E0E0",
            "#,##0.0#0", "0‰%"};
        for (String p : badPatterns) {
            line(w, "dfbad", esc(p), attempt(() -> new DecimalFormat(p, DecimalFormatSymbols.getInstance(Locale.US)),
                    f -> "ok"));
        }
        for (int b = 0; b < 40 * scale; b++) {
            line(w, "dfd", "ordinate:325", Integer.toString(b), Long.toHexString(digestDecimalFormat("ordinate:325", b)));
            line(w, "dfd", "us", Integer.toString(b), Long.toHexString(digestDecimalFormat("us", b)));
            line(w, "dfd", "us-obj", Integer.toString(b), Long.toHexString(digestDecimalFormat("us-obj", b)));
            line(w, "dfd", "ordinate:3", Integer.toString(b), Long.toHexString(digestDecimalFormat("ordinate:3", b)));
        }
    }

    static long digestDecimalFormat(String setup, int batch) {
        Rng r = new Rng(0x4000000L + batch);
        NumberFormat f = dfSetup(setup);
        long h = FNV_OFFSET;
        for (int i = 0; i < BATCH; i++) {
            h = fnv(h, dfFormat(setup, f, genDouble(r)));
        }
        return h;
    }

    // ------------------------------------------------------------------
    // String.format

    static String encodeArg(Object o) {
        if (o == null) {
            return "n";
        } else if (o instanceof Integer) {
            return "i:" + o;
        } else if (o instanceof Long) {
            return "l:" + o;
        } else if (o instanceof Short) {
            return "sh:" + o;
        } else if (o instanceof Byte) {
            return "by:" + o;
        } else if (o instanceof Double) {
            return "d:" + dbits((Double) o);
        } else if (o instanceof Float) {
            return "f:" + fbits((Float) o);
        } else if (o instanceof String) {
            return "s:" + esc((String) o);
        } else if (o instanceof Character) {
            return "c:" + Integer.toHexString((Character) o);
        } else if (o instanceof Boolean) {
            return "b:" + o;
        } else if (o instanceof BigDecimal) {
            return "bd:" + o;
        }
        throw new IllegalArgumentException(o.getClass().toString());
    }

    static void fmtLine(PrintWriter w, Locale l, String format, Object... args) {
        List<String> fields = new ArrayList<>();
        fields.add("fmt");
        fields.add(l == null ? "-" : l.toLanguageTag());
        fields.add(esc(format));
        fields.add(attempt(() -> l == null ? String.format(format, args) : String.format(l, format, args),
                TextVectors::esc));
        for (Object a : args) {
            fields.add(encodeArg(a));
        }
        w.println(String.join("\t", fields));
    }

    static void formatVectors(PrintWriter w) {
        // The conversions JTS uses.
        fmtLine(w, null, "Wrong number of control points for element %d - expected %d or %d, found %d", 3, 4, 7, 9);
        fmtLine(w, null, "%s%d%s: %s %s %s", "A", 0, "a", "LINESTRING (0 0, 1 1)", "true", "x");
        fmtLine(w, null, "POLYGON ((%s, %s, %s, %s))", "0 0", "1 0", "1 1", "0 0");
        fmtLine(w, Locale.ROOT, "%8.4f", 1.23456789);
        fmtLine(w, null, "%8.4f", -12345.678915);
        fmtLine(w, null, "%1$d -> %2$d = %4$f /= %3$f", 0, 1, 20d, 19.999999999d);
        fmtLine(w, null, "\n# pts = %d, Scale = %f\n", 1000, 1e-6);
        fmtLine(w, null, "%d: Pts - input %d, output %d - %s", 1, 2, 3, "ok");
        fmtLine(w, null, "\nTest %d: %s   (# errs: %d = %d%%)\n", 5, "desc", 1, 20);
        fmtLine(w, null, "OverlayNG  scale: %f", 1e10);
        fmtLine(w, null, "%s - Hash code: %d   original:  %d\n", "LINESTRING", -12345, 99);
        fmtLine(w, null, "Loading iIndex with %d points\n", 100000);
        // General conversions and flags.
        Object[] general = {null, "abc", "", 'x', 42, -42L, 1.5, true, false, new BigDecimal("1.50")};
        String[] generalFormats = {"%s", "%S", "%10s", "%-10s|", "%.2s", "%10.2s|", "%b", "%B", "%5b", "%h", "%H",
            "%c", "%C", "%-5c|", "%%", "%5%", "%-5%|", "%n", "%#s", "%+s", "%0s", "%,s", "%(s", "% s", "%-s", "%.2c",
            "%#b", "%#h", "%-b"};
        for (String f : generalFormats) {
            for (Object a : general) {
                fmtLine(w, null, f, a);
            }
        }
        // Integers.
        Object[] ints = {0, 1, -1, 42, -42, Integer.MAX_VALUE, Integer.MIN_VALUE, 0L, Long.MAX_VALUE, Long.MIN_VALUE,
            -1234567890123L, (short) -5, (short) 300, (byte) -1, (byte) 100, 1.5, "s", 'c', null};
        String[] intFormats = {"%d", "%5d", "%-5d|", "%05d", "%+d", "% d", "%,d", "%(d", "%+,010d", "%(,d", "%o", "%x",
            "%X", "%#o", "%#x", "%#X", "%08x", "%-8x|", "%#010x", "%(o", "%+x", "% x", "%,x", "%.2d", "%#d", "%-05d",
            "%+ d", "%0d", "%-d", "%c", "%,(15d", "%010d"};
        for (String f : intFormats) {
            for (Object a : ints) {
                fmtLine(w, null, f, a);
            }
        }
        // Floating point.
        Rng r = new Rng(21);
        List<Object> floats = new ArrayList<>(Arrays.asList(0.0, -0.0, 1.0, -1.0, 0.5, 1.5, 2.5, 0.125, 0.15, 0.25,
            0.35, 1.005, 2.675, 1e-5, 1e-4, 1e-3, 0.0001234, 123456.789, 1e7, 1.0e16, 1e21, 1e100, 1e-100,
            Double.MIN_VALUE, Double.MAX_VALUE, Double.NaN, Double.POSITIVE_INFINITY, Double.NEGATIVE_INFINITY,
            9.999995, 9.9999949, 99.99995, 0.000099999995, 1.0f, 0.1f, 3.4e38f, Float.NaN, 1.4e-45f, 4.9e-324,
            123456789L, new BigDecimal("123.456"), new BigDecimal("-0.0005"), new BigDecimal("1E+5"),
            new BigDecimal("0.00"), new BigDecimal("9.9995"), new BigDecimal("12345678901234567890.123456789"), 5, null));
        for (int i = 0; i < 40; i++) {
            floats.add(genDouble(r));
        }
        String[] floatFormats = {"%f", "%.0f", "%.1f", "%.2f", "%.3f", "%.10f", "%.20f", "%10.3f", "%-10.3f|",
            "%010.3f", "%+f", "% f", "%(f", "%,f", "%,.2f", "%#.0f", "%e", "%.0e", "%.3e", "%.10e", "%E", "%12.3e",
            "%+e", "%(e", "%#.0e", "%g", "%.0g", "%.1g", "%.3g", "%.10g", "%G", "%12.4g", "%#g", "%a", "%.0a",
            "%.1a", "%.3a", "%.13a", "%A", "%20a", "%020a", "%+a", "%,e", "%#g", "%,g", "%(a", "%-020f", "%(,.2f",
            "%08.2f", "%-+12.4e|"};
        for (String f : floatFormats) {
            for (Object a : floats) {
                fmtLine(w, null, f, a);
            }
        }
        // Localized output.
        Locale[] locales = {Locale.US, Locale.ROOT, Locale.GERMANY, Locale.FRANCE, Locale.forLanguageTag("tr"),
            Locale.forLanguageTag("ar-EG"), Locale.forLanguageTag("hi-IN"), Locale.forLanguageTag("en-IN"),
            Locale.forLanguageTag("de-CH")};
        for (Locale l : locales) {
            for (String f : new String[]{"%,d", "%,.3f", "%.2e", "%010.2f", "%S", "%,g", "%d", "%X"}) {
                for (Object a : new Object[]{1234567, -9876543.21, 0.000123, "istanbul", 1234567.0}) {
                    fmtLine(w, l, f, a);
                }
            }
        }
        // Argument indexes and malformed specifiers.
        fmtLine(w, null, "%2$s %1$s %<s %s %s", "a", "b");
        fmtLine(w, null, "%s %s %s", "a", "b");
        fmtLine(w, null, "%3$s", "a", "b");
        fmtLine(w, null, "%<s", "a");
        fmtLine(w, null, "%0$s", "a");
        fmtLine(w, null, "%99999999999$s", "a");
        fmtLine(w, null, "%99999999999s", "a");
        fmtLine(w, null, "%.99999999999s", "a");
        fmtLine(w, null, "%", "a");
        fmtLine(w, null, "abc%");
        fmtLine(w, null, "%q", "a");
        fmtLine(w, null, "%-", "a");
        fmtLine(w, null, "%5", "a");
        fmtLine(w, null, "%.f", 1.0);
        fmtLine(w, null, "%--5s", "a");
        fmtLine(w, null, "%t", "a");
        fmtLine(w, null, "%ty", (Object) null);
        fmtLine(w, null, "%5n");
        fmtLine(w, null, "%-n");
        fmtLine(w, null, "%.2n");
        fmtLine(w, null, "%.2%");
        fmtLine(w, null, "%+%");
        fmtLine(w, null, "%c", 0x1F600);
        fmtLine(w, null, "%c", -1);
        fmtLine(w, null, "%c", 0x110000);
        fmtLine(w, null, "no specifiers");
        fmtLine(w, null, "100%% done %s", "now");
        fmtLine(w, null, "%s", "ßİi");
        fmtLine(w, null, "%S", "ßİiǆ");
        fmtLine(w, Locale.forLanguageTag("tr"), "%S", "istanbul");
        for (int b = 0; b < 40 * scale; b++) {
            line(w, "fmtd", Integer.toString(b), Long.toHexString(digestFormat(b)));
        }
    }

    static final String[] DIGEST_CONVERSIONS = {"f", "e", "g", "a"};

    /** A random floating-point format specifier and argument. */
    static String genFloatFormat(Rng r) {
        int prec = r.below(22) - 1;
        String conv = DIGEST_CONVERSIONS[r.below(4)];
        return "%" + (prec >= 0 ? "." + prec : "") + conv;
    }

    static long digestFormat(int batch) {
        Rng r = new Rng(0x5000000L + batch);
        long h = FNV_OFFSET;
        for (int i = 0; i < BATCH; i++) {
            String f = genFloatFormat(r);
            double d = genDouble(r);
            h = fnv(h, String.format(Locale.ROOT, f, d));
            float x = genFloat(r);
            h = fnv(h, String.format(Locale.ROOT, f, x));
        }
        return h;
    }

    // ------------------------------------------------------------------
    // Everything else

    static void miscVectors(PrintWriter w) {
        integerVectors(w);
        characterVectors(w);
        caseVectors(w);
        regexVectors(w);
        builderVectors(w);
        bigDecimalVectors(w);
        localeVectors(w);
        hexStringVectors(w);
        stringVectors(w);
    }

    static void stringVectors(PrintWriter w) {
        String[] strs = {"", "a", "abcabc", "a\ud801\udc00b\ud801\udc00", "\ud801", "\udc00\ud801x", "hello world"};
        int[] idx = {-1, 0, 1, 2, 3, 4, 6, 10};
        int[] chars = {'a', 'b', 'x', 0x10400, 0xd801, 0xdc00, -1};
        for (String s : strs) {
            for (int i : idx) {
                line(w, "scp", esc(s), Integer.toString(i), attempt(() -> s.codePointAt(i), String::valueOf),
                        attempt(() -> s.codePointBefore(i), String::valueOf),
                        attempt(() -> s.offsetByCodePoints(0, i), String::valueOf),
                        attempt(() -> s.offsetByCodePoints(i, -1), String::valueOf),
                        attempt(() -> s.codePointCount(0, i), String::valueOf),
                        attempt(() -> s.charAt(i), c -> Integer.toString(c)),
                        attempt(() -> s.substring(i), TextVectors::esc),
                        attempt(() -> s.substring(1, i), TextVectors::esc),
                        attempt(() -> s.startsWith("b", i), String::valueOf),
                        attempt(() -> s.lastIndexOf("bc", i), String::valueOf),
                        attempt(() -> s.indexOf("bc", i), String::valueOf));
                for (int c : chars) {
                    line(w, "sidx", esc(s), Integer.toString(i), Integer.toString(c),
                            Integer.toString(s.indexOf(c, i)), Integer.toString(s.lastIndexOf(c, i)),
                            Integer.toString(s.indexOf(c)), Integer.toString(s.lastIndexOf(c)));
                }
            }
            for (String t : strs) {
                line(w, "s2", esc(s), esc(t), esc(s.concat(t)), b(s.contentEquals(t)), b(s.contains(t)),
                        b(s.contentEquals(new StringBuilder(t))), b(s.contains(new StringBuilder(t))),
                        Integer.toString(s.compareTo(t)), b(s.endsWith(t)), b(s.startsWith(t)),
                        Integer.toString(s.lastIndexOf(t)), Integer.toString(s.indexOf(t)),
                        b(s.regionMatches(1, t, 0, 2)), b(s.regionMatches(true, 0, t, 1, 1)),
                        esc(String.join("|", s, t, s)), esc(s.replace('a', 'Z')), esc(s.trim()));
            }
        }
        line(w, "sjoin", esc(String.join(", ", new ArrayList<>(Arrays.asList("a", null, "c")))));
        char[] cs = {'a', 'b', 'c'};
        for (int[] r : new int[][]{{0, 3}, {1, 1}, {2, 2}, {-1, 1}, {1, 3}, {3, 0}}) {
            line(w, "svo", Integer.toString(r[0]), Integer.toString(r[1]),
                    attempt(() -> String.valueOf(cs, r[0], r[1]), TextVectors::esc),
                    attempt(() -> new String(new int[]{0x61, 0x10400, 0x62}, r[0], r[1]), TextVectors::esc));
        }
    }

    static void integerVectors(PrintWriter w) {
        String[] inputs = {"0", "-0", "+0", "1", "-1", "+1", "2147483647", "2147483648", "-2147483648",
            "-2147483649", "9223372036854775807", "9223372036854775808", "-9223372036854775808",
            "-9223372036854775809", "", "-", "+", "1a", "a", "ff", "FF", "-ff", "7fffffff", "80000000", "zz",
            "ZZ", "11111111111111111111111111111111", "0x10", "#10", "010", "0", "-0x10", "-#10", "-010", "0x",
            "#", "08", "0x-1", "--1", " 1", "1 ", "١٢٣", "１２３", "4294967295", "4294967296", "18446744073709551615",
            "18446744073709551616", "1234567890123456789012", "0000000000000000000001", "-0000000000000000000001",
            "123456789012", "12345678901234567890", "99999999999999999999", "1844674407370955161a",
            "+18446744073709551615", "-1", "32767", "32768", "-32768", "-32769", "127", "128", "-128", "-129"};
        int[] radixes = {10, 16, 2, 8, 36, 1, 37};
        for (String s : inputs) {
            for (int radix : radixes) {
                line(w, "pi", esc(s), Integer.toString(radix), attempt(() -> Integer.parseInt(s, radix), String::valueOf));
                line(w, "pl", esc(s), Integer.toString(radix), attempt(() -> Long.parseLong(s, radix), String::valueOf));
                line(w, "pui", esc(s), Integer.toString(radix),
                        attempt(() -> Integer.parseUnsignedInt(s, radix), String::valueOf));
                line(w, "pul", esc(s), Integer.toString(radix),
                        attempt(() -> Long.parseUnsignedLong(s, radix), String::valueOf));
                line(w, "ps", esc(s), Integer.toString(radix), attempt(() -> Short.parseShort(s, radix), String::valueOf));
                line(w, "pb", esc(s), Integer.toString(radix), attempt(() -> Byte.parseByte(s, radix), String::valueOf));
            }
            line(w, "di", esc(s), attempt(() -> Integer.decode(s), String::valueOf));
            line(w, "dl", esc(s), attempt(() -> Long.decode(s), String::valueOf));
            line(w, "ds", esc(s), attempt(() -> Short.decode(s), String::valueOf));
            line(w, "db", esc(s), attempt(() -> Byte.decode(s), String::valueOf));
        }
        line(w, "pi", "null", "10", attempt(() -> Integer.parseInt(null, 10), String::valueOf));
        Rng r = new Rng(31);
        for (int i = 0; i < 120; i++) {
            long v = r.next() >> r.below(64);
            int iv = (int) v;
            for (int radix : radixes) {
                line(w, "tsi", Integer.toString(iv), Integer.toString(radix), esc(Integer.toString(iv, radix)),
                        esc(Integer.toUnsignedString(iv, radix)));
                line(w, "tsl", Long.toString(v), Integer.toString(radix), esc(Long.toString(v, radix)),
                        esc(Long.toUnsignedString(v, radix)));
            }
            line(w, "bits", Long.toString(v), Integer.toBinaryString(iv), Integer.toHexString(iv),
                    Integer.toOctalString(iv), Long.toBinaryString(v), Long.toHexString(v), Long.toOctalString(v),
                    Integer.toString(Integer.bitCount(iv)), Integer.toString(Integer.reverse(iv)),
                    Integer.toString(Integer.reverseBytes(iv)), Integer.toString(Integer.highestOneBit(iv)),
                    Integer.toString(Integer.lowestOneBit(iv)), Integer.toString(Integer.numberOfLeadingZeros(iv)),
                    Integer.toString(Integer.numberOfTrailingZeros(iv)), Integer.toString(Integer.signum(iv)),
                    Integer.toString(Integer.rotateLeft(iv, (int) v)), Integer.toString(Integer.rotateRight(iv, (int) v)),
                    Long.toString(Long.reverse(v)), Long.toString(Long.reverseBytes(v)), Long.toString(Long.highestOneBit(v)),
                    Integer.toString(Long.signum(v)), Long.toString(Long.rotateLeft(v, iv)), Integer.toString(Long.hashCode(v)),
                    Integer.toString(Integer.compareUnsigned(iv, (int) (v >> 7))),
                    Integer.toString(Long.compareUnsigned(v, v >> 7)));
        }
    }

    static void characterVectors(PrintWriter w) {
        // One digest per block of 4096 code points.
        for (int block = 0; block <= Character.MAX_CODE_POINT >> 12; block++) {
            long h = FNV_OFFSET;
            for (int cp = block << 12; cp < (block + 1) << 12; cp++) {
                h = fnv(h, charProps(cp));
            }
            line(w, "chr", Integer.toString(block), Long.toHexString(h));
        }
        // Explicit vectors for the Latin-1 range, for readable failures.
        for (int cp = 0; cp < 0x250; cp++) {
            line(w, "chx", Integer.toHexString(cp), charProps(cp));
        }
    }

    static String charProps(int cp) {
        return Character.getType(cp) + " " + Integer.toHexString(Character.toUpperCase(cp)) + " "
                + Integer.toHexString(Character.toLowerCase(cp)) + " " + Integer.toHexString(Character.toTitleCase(cp))
                + " " + Character.digit(cp, 36) + " " + Character.digit(cp, 10) + " " + b(Character.isLetter(cp))
                + b(Character.isDigit(cp)) + b(Character.isLetterOrDigit(cp)) + b(Character.isWhitespace(cp))
                + b(Character.isSpaceChar(cp)) + b(Character.isUpperCase(cp)) + b(Character.isLowerCase(cp))
                + b(Character.isTitleCase(cp)) + b(Character.isAlphabetic(cp)) + b(Character.isIdeographic(cp))
                + b(Character.isDefined(cp)) + b(Character.isISOControl(cp)) + b(Character.isJavaIdentifierStart(cp))
                + b(Character.isJavaIdentifierPart(cp)) + b(Character.isIdentifierIgnorable(cp))
                + (cp <= 0xffff ? " " + Integer.toHexString(Character.toUpperCase((char) cp)) + " "
                        + Integer.toHexString(Character.toLowerCase((char) cp)) : "");
    }

    static String b(boolean v) {
        return v ? "1" : "0";
    }

    static void caseVectors(PrintWriter w) {
        String[] strs = {"abc", "ABC", "point (10 20)", "POINT", "istanbul", "ISTANBUL", "İ", "i̇",
            "İ", "ı", "Į̇", "ß", "straße", "ﬃ", "ŉ", "ǰ", "ΐ",
            "ẞ", "Σ", "ΟΣ", "ΟΣ ", "ΟΣΟ", "AΣ.", "ΣΟ",
            "µ", "ÿ", "ǅ", "ა", "Ა", "𐐀", "𐐨", "\ud800", "\udc00x",
            "J́", "Ì", "Í", "Ĩ", "Ì", "i̇́", "և", "ᾀ", "ᾳ",
            "ͅ", "Ɪ", "Ɐ", "ɐ", "ɥɦ", "ʞ", "Hello, World! 123", ""};
        Locale[] locales = {Locale.ROOT, Locale.US, Locale.forLanguageTag("tr"), Locale.forLanguageTag("az"),
            Locale.forLanguageTag("lt"), Locale.GERMANY, Locale.forLanguageTag("el")};
        for (String s : strs) {
            for (Locale l : locales) {
                line(w, "case", esc(s), l.toLanguageTag(), esc(s.toUpperCase(l)), esc(s.toLowerCase(l)));
            }
            for (String t : strs) {
                line(w, "cmpci", esc(s), esc(t), Integer.toString(Integer.signum(s.compareToIgnoreCase(t))),
                        b(s.equalsIgnoreCase(t)), b(s.regionMatches(true, 0, t, 0, Math.min(s.length(), t.length()))));
            }
        }
        Rng r = new Rng(41);
        for (int i = 0; i < 300; i++) {
            StringBuilder sb = new StringBuilder();
            int n = r.below(8);
            for (int j = 0; j < n; j++) {
                int cp = r.below(3) == 0 ? r.below(0x3000) : r.below(128);
                if (cp >= 0xd800 && cp <= 0xdfff) {
                    cp = 'x';
                }
                sb.appendCodePoint(cp);
            }
            String s = sb.toString();
            for (Locale l : locales) {
                line(w, "case", esc(s), l.toLanguageTag(), esc(s.toUpperCase(l)), esc(s.toLowerCase(l)));
            }
        }
    }

    static void regexVectors(PrintWriter w) {
        String[][] cases = {
            {"\\.", "1.2.3", ""}, {"\\s", "a b\tc\u000bd", "_"}, {"\\s+", "  a   b  ", " "}, {"\\s*,\\s*", "a , b,c ,d", ","},
            {"a*", "baaac", "X"}, {"a*?", "baaac", "X"}, {"a*+", "baaac", "X"}, {"x*", "axb", "-"}, {"", "abc", "-"},
            {"(a)(b)?", "ab a", "[$2$1]"}, {"(?<n>\\d+)", "a12b345", "<${n}>"}, {"\\d{2,3}", "1 12 123 1234", "#"},
            {"[a-c&&[b-d]]", "abcd", "X"}, {"[^abc]", "abcdef", "X"}, {"[\\w&&[^\\d]]+", "ab12cd", "X"}, {"^a|b$", "abab", "X"},
            {"(?=b)", "abab", "|"}, {"(?<=a)b", "abab", "X"}, {"(?<!a)b", "bab", "X"}, {"(a|ab)(c|bcd)(d*)", "abcd", "[$1,$2,$3]"},
            {"\\bfoo\\b", "foo foobar barfoo foo", "X"}, {"\\Bo", "foo boo", "X"}, {"(\\w)\\1", "aabbcd", "X"}, {".", "a\nb\r c", "X"},
            {"$", "ab\n", "X"}, {"\\p{Upper}+", "abCDeF", "_"}, {"\\p{L}+", "ab1é中2", "_"}, {"\\P{L}", "ab1é", "_"},
            {"\\p{IsAlphabetic}", "a1", "_"}, {"[\\p{Lu}\\d]", "aB3", "_"}, {"\\Qa.b\\E", "a.bab", "X"}, {"\\x41\\u0042\\0103", "ABC", "X"},
            {"(a)|b", "ab", "[$1]"}, {"a{0}", "aa", "X"}, {"a{2,}", "aaaaa", "X"}, {"(ab)+", "ababab", "$1"}, {"\\\\", "a\\b", "/"},
            {"[]a]", "]a", "X"}, {"[a-]", "a-b", "X"}, {"a|", "abc", "X"}, {"(a*)*", "aab", "[$1]"}, {"(a|)*", "aab", "[$1]"},
            {"\\$", "a$b", "\\$"}, {"b", "abc", "$0$0"}, {"b", "abc", "\\\\"}, {"b", "abc", "$"}, {"b", "abc", "$2"},
            {"(b)", "abc", "$11"}, {"b", "abc", "${x}"}, {"b", "abc", "x\\"}, {"*a", "abc", "X"}, {"(a", "abc", "X"},
            {"a)", "abc", "X"}, {"[a", "abc", "X"}, {"a{2", "abc", "X"}, {"a{,2}", "abc", "X"}, {"\\k<x>", "abc", "X"},
            {"\\9", "abc", "X"}, {"[z-a]", "abc", "X"}, {"\\i", "abc", "X"}, {"a{3,2}", "abc", "X"}, {"\\", "abc", "X"},
            {"\\x4", "abc", "X"}, {"\\u12", "abc", "X"}, {"\\p{Foo}", "abc", "X"}, {"{1}", "abc", "X"}, {"a**", "abc", "X"},
            {"(?<1a>x)", "abc", "X"},
        };
        for (String[] c : cases) {
            String re = c[0], s = c[1], rep = c[2];
            line(w, "rxall", esc(re), esc(s), esc(rep), attempt(() -> s.replaceAll(re, rep), TextVectors::esc));
            line(w, "rxfirst", esc(re), esc(s), esc(rep), attempt(() -> s.replaceFirst(re, rep), TextVectors::esc));
            line(w, "rxmatch", esc(re), esc(s), attempt(() -> s.matches(re), String::valueOf));
            for (int limit : new int[]{0, -1, 1, 2, 3}) {
                line(w, "rxsplit", esc(re), esc(s), Integer.toString(limit),
                        attempt(() -> s.split(re, limit), a -> {
                            StringBuilder sb = new StringBuilder(Integer.toString(a.length));
                            for (String x : a) {
                                sb.append('\t').append(esc(x));
                            }
                            return sb.toString();
                        }));
            }
        }
        String[][] splits = {{",", "a,b,,c,,"}, {",", ",a,b"}, {",", ""}, {",", ","}, {",", ",,"}, {"x", "abc"},
            {"", "abc"}, {"\\|", "a|b|"}, {"(?=,)", "a,b,c"}, {"\\s+", " a b "}, {"o", "foo:boo"}};
        for (String[] c : splits) {
            for (int limit : new int[]{0, -1, 1, 2, 5}) {
                line(w, "rxsplit", esc(c[0]), esc(c[1]), Integer.toString(limit),
                        attempt(() -> c[1].split(c[0], limit), a -> {
                            StringBuilder sb = new StringBuilder(Integer.toString(a.length));
                            for (String x : a) {
                                sb.append('\t').append(esc(x));
                            }
                            return sb.toString();
                        }));
            }
        }
        String[][] replaces = {{"abcabc", "b", "XY"}, {"abc", "", "-"}, {"", "", "-"}, {"aaa", "aa", "b"}, {"abc", "d", "e"}};
        for (String[] c : replaces) {
            line(w, "replace", esc(c[0]), esc(c[1]), esc(c[2]), esc(c[0].replace(c[1], c[2])));
        }
    }

    /** Applies a scripted sequence of operations to a StringBuilder. */
    static void builderVectors(PrintWriter w) {
        String[] scripts = {
            "append:abc;insert:1:XY;reverse;deleteCharAt:0;setCharAt:1:Q;length;capacity",
            "new:hello;append:12345678901234567890;capacity;setLength:3;capacity;trimToSize;capacity",
            "newcap:2;append:abc;capacity;append:d;capacity;ensureCapacity:100;capacity",
            "append:abcdef;delete:1:3;replace:0:2:WXYZ;indexOf:Y;lastIndexOf:Z;substring:2;substring:1:3",
            "append:a𐀀b;reverse;codePointAt:1;length",
            "append:abc;charAt:5", "append:abc;setCharAt:3:x", "append:abc;deleteCharAt:3", "append:abc;delete:2:1",
            "append:abc;delete:4:5", "append:abc;delete:1:99", "append:abc;insert:4:x", "append:abc;insert:-1:x",
            "append:abc;replace:4:5:x", "append:abc;replace:1:99:x", "append:abc;setLength:-1", "append:abc;substring:4",
            "append:abc;substring:2:1", "append:abc;setLength:5;length;charAt:4", "newcap:-1", "append:abc;capacity",
            "appendnull;appendnullsb;appendnullcs;length", "appendd:1.0E-5;appendf:3.4E38;appendd:100;appendi:-7;appendl:-9;appendb",
            "append:x;appendcp:128512;appendcp:-1", "append:abc;appendcsr:hello:1:3;appendcsr:hello:3:1",
            "append:abc;insertcsr:1:hello:1:3;insertcsr:9:hello:1:3", "append:abcabc;indexOffrom:b:2;lastIndexOffrom:b:2;indexOffrom::10;lastIndexOffrom::10",
            "append:ab\ud800;reverse;length", "append:\udc00\ud800;reverse",
        };
        for (String s : scripts) {
            line(w, "sb", esc(s), esc(runScript(new StringBuilder(), s)));
        }
    }

    static String runScript(StringBuilder initial, String script) {
        StringBuilder sb = initial;
        StringBuilder out = new StringBuilder();
        for (String op : script.split(";")) {
            String[] a = op.split(":", -1);
            try {
                switch (a[0]) {
                    case "new" -> sb = new StringBuilder(a[1]);
                    case "newcap" -> sb = new StringBuilder(Integer.parseInt(a[1]));
                    case "append" -> sb.append(a[1]);
                    case "appendnull" -> sb.append((String) null);
                    case "appendnullsb" -> sb.append((StringBuffer) null);
                    case "appendnullcs" -> sb.append((CharSequence) null);
                    case "appendd" -> sb.append(Double.parseDouble(a[1]));
                    case "appendf" -> sb.append(Float.parseFloat(a[1]));
                    case "appendi" -> sb.append(Integer.parseInt(a[1]));
                    case "appendl" -> sb.append(Long.parseLong(a[1]));
                    case "appendb" -> sb.append(true);
                    case "appendcp" -> sb.appendCodePoint(Integer.parseInt(a[1]));
                    case "appendcsr" -> sb.append(a[1], Integer.parseInt(a[2]), Integer.parseInt(a[3]));
                    case "insertcsr" -> sb.insert(Integer.parseInt(a[1]), a[2], Integer.parseInt(a[3]), Integer.parseInt(a[4]));
                    case "insert" -> sb.insert(Integer.parseInt(a[1]), a[2]);
                    case "reverse" -> sb.reverse();
                    case "deleteCharAt" -> sb.deleteCharAt(Integer.parseInt(a[1]));
                    case "setCharAt" -> sb.setCharAt(Integer.parseInt(a[1]), a[2].charAt(0));
                    case "delete" -> sb.delete(Integer.parseInt(a[1]), Integer.parseInt(a[2]));
                    case "replace" -> sb.replace(Integer.parseInt(a[1]), Integer.parseInt(a[2]), a[3]);
                    case "setLength" -> sb.setLength(Integer.parseInt(a[1]));
                    case "trimToSize" -> sb.trimToSize();
                    case "ensureCapacity" -> sb.ensureCapacity(Integer.parseInt(a[1]));
                    case "length" -> out.append(" len=").append(sb.length());
                    case "capacity" -> out.append(" cap=").append(sb.capacity());
                    case "charAt" -> out.append(" charAt=").append((int) sb.charAt(Integer.parseInt(a[1])));
                    case "codePointAt" -> out.append(" cp=").append(sb.codePointAt(Integer.parseInt(a[1])));
                    case "indexOf" -> out.append(" indexOf=").append(sb.indexOf(a[1]));
                    case "lastIndexOf" -> out.append(" lastIndexOf=").append(sb.lastIndexOf(a[1]));
                    case "indexOffrom" -> out.append(" indexOf=").append(sb.indexOf(a[1], Integer.parseInt(a[2])));
                    case "lastIndexOffrom" -> out.append(" lastIndexOf=").append(sb.lastIndexOf(a[1], Integer.parseInt(a[2])));
                    case "substring" -> out.append(" sub=").append(a.length == 2 ? sb.substring(Integer.parseInt(a[1]))
                            : sb.substring(Integer.parseInt(a[1]), Integer.parseInt(a[2])));
                    default -> throw new IllegalStateException(op);
                }
            } catch (RuntimeException e) {
                out.append(" ").append(exc(e));
            }
            out.append(" [").append(sb).append("]");
        }
        return out.toString();
    }

    static void bigDecimalVectors(PrintWriter w) {
        Rng r = new Rng(51);
        List<Double> ds = new ArrayList<>(Arrays.asList(0.0, -0.0, 1.0, 0.1, 9.0, 10.0, 1e-7, 1e-6, 123.456,
            1e20, 1e-320, Double.MAX_VALUE, Double.MIN_VALUE, -2.5, 0.5));
        for (int i = 0; i < 60; i++) {
            ds.add(genDouble(r));
        }
        for (double d : ds) {
            if (Double.isNaN(d) || Double.isInfinite(d)) {
                continue;
            }
            BigDecimal x = new BigDecimal(d);
            BigDecimal v = BigDecimal.valueOf(d);
            line(w, "bd", dbits(d), x.toString(), x.toPlainString(), x.toEngineeringString(), v.toString(),
                    Integer.toString(x.scale()), Integer.toString(x.precision()), dbits(x.doubleValue()),
                    fbits(x.floatValue()), Long.toString(x.longValue()), Integer.toString(x.intValue()),
                    Integer.toString(x.hashCode()), v.stripTrailingZeros().toString(),
                    attempt(() -> x.setScale(20, RoundingMode.HALF_UP), BigDecimal::toString),
                    attempt(() -> x.setScale(2, BigDecimal.ROUND_HALF_EVEN), BigDecimal::toString),
                    attempt(() -> v.setScale(0), BigDecimal::toString),
                    x.round(new MathContext(5)).toString(), v.movePointLeft(3).toString(), v.movePointRight(3).toString());
        }
        // The operations in DDExpressionPerf.
        BigDecimal a = new BigDecimal(9.0).setScale(20);
        BigDecimal factor = new BigDecimal(10.0).setScale(20);
        BigDecimal aMul = factor.multiply(a);
        BigDecimal aDiv = a.divide(factor, BigDecimal.ROUND_HALF_UP);
        BigDecimal det = a.multiply(a).subtract(aMul.multiply(aDiv));
        line(w, "bdperf", a.toString(), factor.toString(), aMul.toString(), aDiv.toString(), det.toString());
        String[] strs = {"0", "1", "-1", "1.50", "1E+3", "1e-3", "-1.23E-10", "123456789012345678901234567890.5",
            "+.5", "5.", ".", "", "-", "1e", "1e+", "1ee2", "1.2.3", "abc", "1e99999999999", "١.٥", "0.00", "-0E+2",
            "12345678901234567890", "1234567890123456789x", "123456789012345678x9", "1e0000000000002", "1e00000000000000000002"};
        String[] ops = {"0.3", "-7", "1E+1", "0.0001", "3", "0"};
        for (String s : strs) {
            line(w, "bds", esc(s), attempt(() -> new BigDecimal(s), x -> x.toString() + "\t" + x.scale() + "\t" + x.toPlainString()));
            BigDecimal x;
            try {
                x = new BigDecimal(s);
            } catch (NumberFormatException e) {
                continue;
            }
            for (String o : ops) {
                BigDecimal y = new BigDecimal(o);
                line(w, "bdop", esc(s), o, x.add(y).toString(), x.subtract(y).toString(), x.multiply(y).toString(),
                        attempt(() -> x.divide(y, BigDecimal.ROUND_HALF_UP), BigDecimal::toString),
                        attempt(() -> x.divide(y, 5, BigDecimal.ROUND_HALF_EVEN), BigDecimal::toString),
                        attempt(() -> x.divide(y, 3, RoundingMode.DOWN), BigDecimal::toString),
                        attempt(() -> x.divide(y), BigDecimal::toString), Integer.toString(x.compareTo(y)),
                        b(x.equals(y)), x.max(y).toString(), x.min(y).toString());
            }
        }
        for (int b = 0; b < 20 * scale; b++) {
            line(w, "bdd", Integer.toString(b), Long.toHexString(digestBigDecimal(b)));
        }
        for (int mode = 0; mode <= 8; mode++) {
            int m = mode;
            for (String s : new String[]{"2.5", "-2.5", "1.15", "-1.15", "1.25", "0.1", "-0.1", "7"}) {
                line(w, "bdround", s, Integer.toString(mode),
                        attempt(() -> new BigDecimal(s).setScale(0, m), BigDecimal::toString),
                        attempt(() -> new BigDecimal(s).setScale(1, m), BigDecimal::toString));
            }
        }
    }

    /** A BigDecimal with a random unscaled value and scale. */
    static BigDecimal genBigDecimal(Rng r) {
        StringBuilder sb = new StringBuilder();
        if (r.below(2) == 0) {
            sb.append('-');
        }
        int nd = 1 + r.below(r.below(4) == 0 ? 60 : 20);
        for (int i = 0; i < nd; i++) {
            sb.append(DIGITS.charAt(r.below(10)));
        }
        int scale = r.below(5) == 0 ? r.below(800) - 400 : r.below(60) - 30;
        return new BigDecimal(new java.math.BigInteger(sb.toString()), scale);
    }

    static long digestBigDecimal(int batch) {
        Rng r = new Rng(0x7000000L + batch);
        long h = FNV_OFFSET;
        for (int i = 0; i < BATCH; i++) {
            BigDecimal x = genBigDecimal(r);
            h = fnv(h, dbits(x.doubleValue()));
            h = fnv(h, fbits(x.floatValue()));
            h = fnv(h, x.toString());
        }
        return h;
    }

    static void localeVectors(PrintWriter w) {
        String[] tags = {"tr", "en-US", "en", "und", "de-CH", "zh-Hant-TW", "sr-Latn-RS", "iw", "he", "in", "ji",
            "es-419", "de-DE-1996", "EN-us", "x-private", "en-US-x-foo", "ar-EG", "ja-JP-u-ca-japanese", "",
            "abcdefghi", "en_US", "en-USA", "en-123", "zh-yue-HK"};
        for (String t : tags) {
            Locale l = Locale.forLanguageTag(t);
            line(w, "loc", esc(t), esc(l.getLanguage()), esc(l.getScript()), esc(l.getCountry()), esc(l.getVariant()),
                    esc(l.toString()), esc(l.toLanguageTag()), Integer.toString(l.hashCode()));
        }
        String[][] ctors = {{"en", "US", ""}, {"EN", "us", ""}, {"iw", "", ""}, {"", "US", ""}, {"en", "", "POSIX"},
            {"", "", "x"}, {"de", "DE", "1996"}};
        for (String[] c : ctors) {
            @SuppressWarnings("deprecation")
            Locale l = new Locale(c[0], c[1], c[2]);
            line(w, "locnew", esc(c[0]), esc(c[1]), esc(c[2]), esc(l.getLanguage()), esc(l.getCountry()),
                    esc(l.getVariant()), esc(l.toString()), esc(l.toLanguageTag()), Integer.toString(l.hashCode()));
        }
    }

    static void hexStringVectors(PrintWriter w) {
        Rng r = new Rng(61);
        for (double d : dfValues()) {
            line(w, "hex", dbits(d), Double.toHexString(d));
        }
        for (int i = 0; i < 300; i++) {
            double d = genDouble(r);
            float f = genFloat(r);
            line(w, "hex", dbits(d), Double.toHexString(d));
            line(w, "hexf", fbits(f), Float.toHexString(f));
        }
    }

    // ------------------------------------------------------------------

    static void dump(String check, int batch) {
        switch (check) {
            case "dd": {
                Rng r = new Rng(0x1000000L + batch);
                for (int i = 0; i < BATCH; i++) {
                    double d = genDouble(r);
                    System.out.println(dbits(d) + "\t" + Double.toString(d));
                }
                break;
            }
            case "fd": {
                Rng r = new Rng(0x2000000L + batch);
                for (int i = 0; i < BATCH; i++) {
                    float f = genFloat(r);
                    System.out.println(fbits(f) + "\t" + Float.toString(f));
                }
                break;
            }
            case "pdd": {
                Rng r = new Rng(0x3000000L + batch);
                for (int i = 0; i < BATCH; i++) {
                    String s = genDecimalString(r);
                    System.out.println(esc(s) + "\t" + attempt(() -> Double.parseDouble(s), TextVectors::dbits) + "\t"
                            + attempt(() -> Float.parseFloat(s), TextVectors::fbits));
                }
                break;
            }
            default: {
                // dfd:<setup>
                String setup = check.substring(check.indexOf(':') + 1);
                Rng r = new Rng(0x4000000L + batch);
                NumberFormat f = dfSetup(setup);
                for (int i = 0; i < BATCH; i++) {
                    double d = genDouble(r);
                    System.out.println(dbits(d) + "\t" + dfFormat(setup, f, d));
                }
            }
        }
    }
}
