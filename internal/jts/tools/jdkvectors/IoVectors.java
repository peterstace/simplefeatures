// IoVectors prints test vectors for the io area of the Go runtime package
// internal/jts/java: StreamTokenizer, the Reader and Writer classes, the byte
// streams, PrintStream, UTF-8 coding, java.util.Random and
// System.arraycopy.
//
// Regenerate the vectors (from the repository root) with:
//
//	java internal/jts/tools/jdkvectors/IoVectors.java internal/jts/java/testdata
//
// Strings are written in double quotes with every character outside printable
// ASCII, and every double quote and backslash, escaped as a backslash, "u"
// and four hex digits; null is written as null.
// Doubles and floats are written as the hex of their raw bits. An exception
// is written as !<class> followed by its message. In the tokenizer vectors,
// "=" stands for an nval unchanged since the previous token.

import java.io.BufferedReader;
import java.io.File;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.LineNumberReader;
import java.io.OutputStream;
import java.io.PrintStream;
import java.io.Reader;
import java.io.StreamTokenizer;
import java.io.StringReader;
import java.io.StringWriter;
import java.io.Writer;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Random;

public class IoVectors {

    public static void main(String[] args) throws Exception {
        Path dir = Path.of(args.length > 0 ? args[0] : ".");
        Files.createDirectories(dir);
        write(dir.resolve("io_tokenizer.txt"), tokenizerVectors());
        write(dir.resolve("io_readers.txt"), readerVectors());
        write(dir.resolve("io_random.txt"), randomVectors());
        write(dir.resolve("io_arraycopy.txt"), arraycopyVectors());
        write(dir.resolve("io_charset.txt"), charsetVectors());
        write(dir.resolve("io_writers.txt"), writerVectors());
        write(dir.resolve("io_misc.txt"), miscVectors());
    }

    static void write(Path p, StringBuilder sb) throws IOException {
        Files.writeString(p, sb.toString(), StandardCharsets.UTF_8);
    }

    // ------------------------------------------------------------------
    // Formatting

    static String q(String s) {
        if (s == null) {
            return "null";
        }
        StringBuilder sb = new StringBuilder("\"");
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            if (c >= 0x20 && c < 0x7f && c != '"' && c != '\\') {
                sb.append(c);
            } else {
                sb.append(String.format("\\u%04x", (int) c));
            }
        }
        return sb.append('"').toString();
    }

    static String q(char[] cs, int off, int len) {
        return q(new String(cs, off, len));
    }

    static String hex(byte[] b) {
        StringBuilder sb = new StringBuilder("x");
        for (byte v : b) {
            sb.append(String.format("%02x", v & 0xff));
        }
        return sb.toString();
    }

    static String dbits(double d) {
        return Long.toHexString(Double.doubleToRawLongBits(d));
    }

    static String fbits(float f) {
        return Integer.toHexString(Float.floatToRawIntBits(f));
    }

    static String ex(Throwable t) {
        return "!" + t.getClass().getName() + " " + q(t.getMessage());
    }

    interface Op {
        String run() throws Exception;
    }

    static String attempt(Op op) {
        try {
            return op.run();
        } catch (Exception e) {
            return ex(e);
        }
    }

    // ------------------------------------------------------------------
    // StreamTokenizer

    static final String[][] TOKENIZER_CONFIGS = {
        {"default", ""},
        {"wkt", "R W97,122 W65,90 W160,255 W48,57 W45,45 W43,43 W46,46 S0,32 C35"},
        {"eol", "E1"},
        {"slashstar", "A1"},
        {"slashslash", "B1"},
        {"slashboth", "A1 B1 E1"},
        {"lower", "L1 E1"},
        {"noslash", "o47 A1"},
        {"ordinary", "O48,57 o46 o45 Q40 C41"},
        {"reset", "R"},
        {"resetnum", "R P W97,122 S0,32 Q34 C35 E1"},
        {"words", "W0,255 S32,32"},
        {"clamp", "W-5,40 S300,400 O-1,1000 o-3 o999 C-1 Q999 W250,300"},
        {"quoteslash", "Q47 B1"},
    };

    static void applyConfig(StreamTokenizer st, String ops) {
        for (String op : ops.trim().split(" ")) {
            if (op.isEmpty()) {
                continue;
            }
            char k = op.charAt(0);
            String rest = op.substring(1);
            int a = 0, b = 0;
            if (rest.contains(",")) {
                String[] p = rest.split(",");
                a = Integer.parseInt(p[0]);
                b = Integer.parseInt(p[1]);
            } else if (!rest.isEmpty()) {
                a = Integer.parseInt(rest);
            }
            switch (k) {
                case 'R' -> st.resetSyntax();
                case 'W' -> st.wordChars(a, b);
                case 'S' -> st.whitespaceChars(a, b);
                case 'O' -> st.ordinaryChars(a, b);
                case 'o' -> st.ordinaryChar(a);
                case 'C' -> st.commentChar(a);
                case 'Q' -> st.quoteChar(a);
                case 'P' -> st.parseNumbers();
                case 'E' -> st.eolIsSignificant(a != 0);
                case 'A' -> st.slashStarComments(a != 0);
                case 'B' -> st.slashSlashComments(a != 0);
                case 'L' -> st.lowerCaseMode(a != 0);
                default -> throw new IllegalArgumentException(op);
            }
        }
    }

    static final String ALPHABET =
        "aaabbcxyzAQZ0000111222345678999....----++  \t\n\r\r\n\n//**##\"\"''\\\\(),,;=e"
            + "é ÿ\u0080Āα \u0000\u0001\u007f😀";

    static String randomText(Random r, int maxLen) {
        int n = r.nextInt(maxLen + 1);
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < n; i++) {
            if (r.nextInt(30) == 0) {
                sb.append((char) r.nextInt(0x10000));
            } else {
                sb.append(ALPHABET.charAt(r.nextInt(ALPHABET.length())));
            }
        }
        return sb.toString();
    }

    static String randomNumber(Random r) {
        StringBuilder sb = new StringBuilder();
        if (r.nextInt(3) == 0) {
            sb.append('-');
        }
        int kind = r.nextInt(10);
        int digits = kind == 0 ? 300 + r.nextInt(100) : kind == 1 ? 15 + r.nextInt(15) : 1 + r.nextInt(12);
        int dot = r.nextInt(digits + 2) - 1;
        for (int i = 0; i < digits; i++) {
            if (i == dot) {
                sb.append('.');
            }
            sb.append((char) ('0' + (i == 0 && r.nextBoolean() ? 0 : r.nextInt(10))));
        }
        if (dot == digits) {
            sb.append('.');
        }
        if (r.nextInt(8) == 0) {
            sb.append('e').append(r.nextBoolean() ? "-" : "").append(r.nextInt(400));
        }
        if (r.nextInt(15) == 0) {
            sb.append(r.nextBoolean() ? "." : "-").append(r.nextInt(1000));
        }
        return sb.toString();
    }

    static final String[] WS = {" ", " ", " ", "  ", "\t", "\n", "\r\n", "\r", "\n\n"};

    static String randomWkt(Random r, int depth) {
        String[] types = {"POINT", "LINESTRING", "POLYGON", "MULTIPOINT", "MULTILINESTRING", "MULTIPOLYGON",
            "GEOMETRYCOLLECTION", "LinearRing", "point"};
        String[] dims = {"", "", " Z", " M", " ZM", "Z", "M"};
        StringBuilder sb = new StringBuilder();
        sb.append(types[r.nextInt(types.length)]).append(dims[r.nextInt(dims.length)]);
        sb.append(WS[r.nextInt(WS.length)]);
        if (r.nextInt(8) == 0) {
            return sb.append(r.nextBoolean() ? "EMPTY" : "empty").toString();
        }
        sb.append('(');
        int n = 1 + r.nextInt(4);
        for (int i = 0; i < n; i++) {
            if (i > 0) {
                sb.append(',').append(r.nextBoolean() ? " " : WS[r.nextInt(WS.length)]);
            }
            if (depth > 0 && r.nextInt(3) == 0) {
                sb.append(randomWkt(r, depth - 1));
            } else if (r.nextInt(4) == 0) {
                sb.append('(');
                int m = 1 + r.nextInt(3);
                for (int j = 0; j < m; j++) {
                    if (j > 0) {
                        sb.append(", ");
                    }
                    sb.append(randomNumber(r)).append(' ').append(randomNumber(r));
                }
                sb.append(')');
            } else {
                int m = 2 + r.nextInt(3);
                for (int j = 0; j < m; j++) {
                    if (j > 0) {
                        sb.append(WS[r.nextInt(WS.length)]);
                    }
                    sb.append(r.nextInt(12) == 0 ? "NaN" : randomNumber(r));
                }
            }
            if (r.nextInt(20) == 0) {
                sb.append(" # comment ").append(randomNumber(r)).append('\n');
            }
        }
        sb.append(')');
        return sb.toString();
    }

    static StringBuilder tokenizerVectors() throws IOException {
        StringBuilder sb = new StringBuilder();
        for (String[] c : TOKENIZER_CONFIGS) {
            sb.append("CONFIG ").append(c[0]).append(' ').append(q(c[1])).append('\n');
        }
        Random r = new Random(20261009);
        List<String> inputs = new ArrayList<>();
        String[] fixed = {
            "", " ", "\n", "\r", "\r\n", "\n\r", "-", ".", "-.", ".-", "--1", "-x", "1.2.3", "1..2", "-0", "0.0",
            "007", "1.", ".5", "-.5", "123456789012345678901234567890", "0.1", "0.3", "1e10",
            "\"abc", "'a\\nb'", "\"\\0\\12\\123\\1234\\477\\8\\\"\"", "\"a\\", "'x\r'y", "\"\\a\\b\\f\\r\\t\\v\\q\"",
            "a/b", "a//b\nc", "a/*b*/c", "/* x\r\n y */z", "/*/ */", "/**/", "/", "//", "/*", "a#b\nc", "#\r\nx",
            "POINT (1 2)", "POINT(1 2)", "LINESTRING (0 0, 10 10.5, -3.25 4e5)", "POLYGON EMPTY",
            "MULTIPOINT ((1 2), (3 4))", "POINT ZM (1 2 3 4)", "point z(1 2 3)", "GEOMETRYCOLLECTION (POINT (1 2))",
            "abcéĀdef", "ÿ\u0080x", "a\u0000b", "x😀y", " 　a",
        };
        for (String s : fixed) {
            inputs.add(s);
        }
        for (int i = 0; i < 160; i++) {
            inputs.add(randomText(r, 40));
        }
        for (int i = 0; i < 70; i++) {
            inputs.add(randomWkt(r, 2));
        }
        for (int i = 0; i < 90; i++) {
            StringBuilder s = new StringBuilder();
            int n = 1 + r.nextInt(5);
            for (int j = 0; j < n; j++) {
                s.append(randomNumber(r)).append(WS[r.nextInt(WS.length)]);
            }
            inputs.add(s.toString());
        }
        int caseNo = 0;
        int configNo = 0;
        for (String[] c : TOKENIZER_CONFIGS) {
            boolean full = c[0].equals("default") || c[0].equals("wkt");
            configNo++;
            for (int i = 0; i < inputs.size(); i++) {
                if (!full && ((i + configNo) % 3 != 0 || inputs.get(i).length() > 100)) {
                    continue;
                }
                String mode = switch (caseNo % 7) {
                    case 3 -> "b";
                    case 5 -> "s";
                    default -> "r";
                };
                String input = inputs.get(i);
                if (c[0].equals("lower") && !input.chars().allMatch(ch -> ch < 0x100)) {
                    // Lowercasing beyond Latin-1 is String.toLowerCase's
                    // concern, not the tokenizer's.
                    continue;
                }
                if (mode.equals("s")) {
                    StringBuilder bytes = new StringBuilder();
                    for (char ch : input.toCharArray()) {
                        bytes.append((char) (ch & 0xff));
                    }
                    input = bytes.toString();
                }
                tokenizeCase(sb, c, mode, input, new Random(caseNo));
                caseNo++;
            }
        }
        return sb;
    }

    @SuppressWarnings("deprecation")
    static void tokenizeCase(StringBuilder sb, String[] config, String mode, String input, Random r)
            throws IOException {
        StreamTokenizer st;
        switch (mode) {
            case "b" -> st = new StreamTokenizer(new BufferedReader(new StringReader(input), 3));
            case "s" -> {
                byte[] b = new byte[input.length()];
                for (int i = 0; i < b.length; i++) {
                    b[i] = (byte) input.charAt(i);
                }
                st = new StreamTokenizer(new ByteArrayInputStream(b));
            }
            default -> st = new StreamTokenizer(new StringReader(input));
        }
        applyConfig(st, config[1]);
        sb.append("CASE ").append(config[0]).append(' ').append(mode).append(' ').append(q(input)).append('\n');
        if (r.nextInt(4) == 0) {
            st.pushBack();
            sb.append("S ").append(q(st.toString())).append(' ').append(st.lineno()).append('\n');
        }
        int eofs = 0;
        long prevNval = 0;
        for (int n = 0; n < 2000 && eofs < 2; n++) {
            if (n > 0 && r.nextInt(6) == 0) {
                st.pushBack();
                sb.append("P\n");
                if (r.nextInt(3) == 0) {
                    st.pushBack();
                    sb.append("P\n");
                }
            }
            int t = st.nextToken();
            if (t != st.ttype) {
                throw new AssertionError();
            }
            long nvalBits = Double.doubleToRawLongBits(st.nval);
            sb.append("T ").append(t).append(' ').append(st.lineno()).append(' ')
                .append(nvalBits == prevNval ? "=" : Long.toHexString(nvalBits)).append(' ').append(q(st.sval))
                .append('\n');
            prevNval = nvalBits;
            if (n % 16 == 0 || t == StreamTokenizer.TT_EOF && eofs == 0) {
                sb.append("S ").append(q(st.toString())).append(' ').append(st.lineno()).append('\n');
            }
            if (t == StreamTokenizer.TT_EOF) {
                eofs++;
            }
        }
        sb.append("END\n");
    }

    // ------------------------------------------------------------------
    // Readers

    static final String LINE_ALPHABET = "ab \n\n\r\r\r\né\t";

    static StringBuilder readerVectors() throws Exception {
        StringBuilder sb = new StringBuilder();
        Random r = new Random(42);
        String[] fixed = {"", "\n", "\r", "\r\n", "\n\r", "\r\r\n\n", "a", "a\n", "a\r", "a\r\n", "a\nb", "a\rb",
            "a\r\nb", "\n\na\r\r", "abc\r", "abc\r\ndef\rghi\njkl", "\r\n\r\n\r\n"};
        List<String> inputs = new ArrayList<>(List.of(fixed));
        for (int i = 0; i < 140; i++) {
            StringBuilder s = new StringBuilder();
            int n = r.nextInt(r.nextInt(8) == 0 ? 120 : 25);
            for (int j = 0; j < n; j++) {
                s.append(LINE_ALPHABET.charAt(r.nextInt(LINE_ALPHABET.length())));
            }
            inputs.add(s.toString());
        }
        String[] kinds = {"BR", "LNR", "SR", "ISR"};
        int[] sizes = {1, 2, 3, 4, 7, 16, 8192};
        int caseNo = 0;
        for (String input : inputs) {
            for (String kind : kinds) {
                int size = sizes[caseNo % sizes.length];
                readerCase(sb, kind, size, input, new Random(caseNo * 31L + 7), caseNo % 5 == 0);
                caseNo++;
            }
        }
        return sb;
    }

    static void readerCase(StringBuilder sb, String kind, int size, String input, Random r, boolean linesOnly)
            throws Exception {
        Reader rd;
        switch (kind) {
            case "BR" -> rd = new BufferedReader(new StringReader(input), size);
            case "LNR" -> rd = new LineNumberReader(new StringReader(input), size);
            case "ISR" -> rd = new InputStreamReader(new ByteArrayInputStream(input.getBytes(StandardCharsets.UTF_8)));
            default -> rd = new StringReader(input);
        }
        sb.append("RCASE ").append(kind).append(' ').append(size).append(' ').append(q(input)).append('\n');
        int nops = linesOnly ? 0 : 5 + r.nextInt(30);
        for (int i = 0; i < nops; i++) {
            int k = r.nextInt(12);
            String line;
            final Reader fr = rd;
            switch (k) {
                case 0, 1, 2 -> {
                    if (!(rd instanceof BufferedReader br)) {
                        line = "r = " + attempt(() -> "" + fr.read());
                    } else {
                        line = "L = " + attempt(() -> q(br.readLine()));
                    }
                }
                case 3, 4 -> line = "r = " + attempt(() -> "" + fr.read());
                case 5 -> {
                    int arrLen = r.nextInt(8);
                    int off = r.nextInt(arrLen + 2) - (r.nextInt(10) == 0 ? 2 : 0);
                    int len = r.nextInt(arrLen + 2) - (r.nextInt(10) == 0 ? 2 : 0);
                    char[] buf = new char[arrLen];
                    line = "a " + arrLen + " " + off + " " + len + " = " + attempt(() -> {
                        int n = fr.read(buf, off, len);
                        return n + " " + q(new String(buf));
                    });
                }
                case 6 -> {
                    int limit = r.nextInt(12) - (r.nextInt(15) == 0 ? 3 : 0);
                    line = "m " + limit + " = " + attempt(() -> {
                        fr.mark(limit);
                        return "ok";
                    });
                }
                case 7 -> line = "R = " + attempt(() -> {
                    fr.reset();
                    return "ok";
                });
                case 8 -> {
                    long n = r.nextInt(12) - (r.nextInt(10) == 0 ? 3 : 0);
                    line = "s " + n + " = " + attempt(() -> "" + fr.skip(n));
                }
                case 9 -> line = "y = " + attempt(() -> "" + fr.ready());
                case 10 -> {
                    if (rd instanceof LineNumberReader lnr && r.nextInt(4) == 0) {
                        int ln = r.nextInt(5);
                        lnr.setLineNumber(ln);
                        line = "N " + ln;
                    } else {
                        line = "M = " + fr.markSupported();
                    }
                }
                default -> {
                    if (r.nextInt(8) == 0) {
                        line = "C = " + attempt(() -> {
                            fr.close();
                            return "ok";
                        });
                    } else {
                        line = "r = " + attempt(() -> "" + fr.read());
                    }
                }
            }
            sb.append(line);
            if (rd instanceof LineNumberReader lnr) {
                sb.append(" n=").append(lnr.getLineNumber());
            }
            sb.append('\n');
        }
        // Drain the rest.
        final Reader fr = rd;
        for (int i = 0; i < 300; i++) {
            String line;
            String res;
            if (rd instanceof BufferedReader br) {
                res = attempt(() -> q(br.readLine()));
                line = "L = " + res;
            } else {
                res = attempt(() -> "" + fr.read());
                line = "r = " + res;
            }
            sb.append(line);
            if (rd instanceof LineNumberReader lnr) {
                sb.append(" n=").append(lnr.getLineNumber());
            }
            sb.append('\n');
            if (res.equals("null") || res.equals("-1") || res.startsWith("!")) {
                break;
            }
        }
        sb.append("END\n");
    }

    // ------------------------------------------------------------------
    // Random

    static StringBuilder randomVectors() {
        StringBuilder sb = new StringBuilder();
        Random meta = new Random(7);
        List<Long> seeds = new ArrayList<>(List.of(0L, 1L, -1L, 42L, 123456789L, Long.MIN_VALUE, Long.MAX_VALUE,
            0x5DEECE66DL, 1L << 48, (1L << 48) - 1));
        for (int i = 0; i < 150; i++) {
            seeds.add(meta.nextLong());
        }
        int[] bounds = {1, 2, 3, 5, 7, 10, 13, 100, 1000, 1 << 30, (1 << 30) + 1, Integer.MAX_VALUE, 1500000000,
            1 << 16, 1 << 20, 6, 0x40000001, 0x7ffffffe, 0, -1, Integer.MIN_VALUE};
        for (long seed : seeds) {
            Random r = new Random(seed);
            sb.append("SEED ").append(seed).append('\n');
            int n = 40 + meta.nextInt(40);
            for (int i = 0; i < n; i++) {
                int k = meta.nextInt(10);
                switch (k) {
                    case 0 -> sb.append("i ").append(r.nextInt()).append('\n');
                    case 1 -> {
                        int b = meta.nextInt(3) == 0 ? 1 + meta.nextInt(Integer.MAX_VALUE)
                            : meta.nextInt(4) == 0 ? 1 << meta.nextInt(31) : bounds[meta.nextInt(bounds.length)];
                        String res;
                        try {
                            res = "" + r.nextInt(b);
                        } catch (IllegalArgumentException e) {
                            res = ex(e);
                        }
                        sb.append("b ").append(b).append(' ').append(res).append('\n');
                    }
                    case 2 -> sb.append("l ").append(r.nextLong()).append('\n');
                    case 3 -> sb.append("z ").append(r.nextBoolean()).append('\n');
                    case 4 -> sb.append("f ").append(fbits(r.nextFloat())).append('\n');
                    case 5, 6 -> sb.append("d ").append(dbits(r.nextDouble())).append('\n');
                    case 7 -> sb.append("g ").append(dbits(r.nextGaussian())).append('\n');
                    case 8 -> {
                        byte[] b = new byte[meta.nextInt(11)];
                        r.nextBytes(b);
                        sb.append("y ").append(b.length).append(' ').append(hex(b)).append('\n');
                    }
                    default -> {
                        if (meta.nextInt(4) == 0) {
                            long s = meta.nextLong();
                            r.setSeed(s);
                            sb.append("s ").append(s).append('\n');
                        } else {
                            sb.append("g ").append(dbits(r.nextGaussian())).append('\n');
                        }
                    }
                }
            }
        }
        return sb;
    }

    // ------------------------------------------------------------------
    // System.arraycopy

    static Object makeArray(String type, int n) {
        Object a = switch (type) {
            case "int" -> new int[n];
            case "long" -> new long[n];
            case "short" -> new short[n];
            case "byte" -> new byte[n];
            case "char" -> new char[n];
            case "boolean" -> new boolean[n];
            case "float" -> new float[n];
            case "double" -> new double[n];
            case "String" -> new String[n];
            case "Object" -> new Object[n];
            case "StringObj" -> new Object[n];
            case "mixed" -> new Object[n];
            default -> throw new IllegalArgumentException(type);
        };
        for (int i = 0; i < n; i++) {
            int v = i + 1;
            switch (type) {
                case "int" -> ((int[]) a)[i] = v;
                case "long" -> ((long[]) a)[i] = v;
                case "short" -> ((short[]) a)[i] = (short) v;
                case "byte" -> ((byte[]) a)[i] = (byte) v;
                case "char" -> ((char[]) a)[i] = (char) ('a' + i);
                case "boolean" -> ((boolean[]) a)[i] = i % 3 == 0;
                case "float" -> ((float[]) a)[i] = v + 0.5f;
                case "double" -> ((double[]) a)[i] = v + 0.25;
                case "String", "StringObj" -> ((Object[]) a)[i] = "s" + v;
                case "Object" -> ((Object[]) a)[i] = i % 4 == 3 ? null : "o" + v;
                case "mixed" -> ((Object[]) a)[i] = i % 3 == 2 ? (Object) Integer.valueOf(v) : "m" + v;
                default -> throw new IllegalArgumentException(type);
            }
        }
        return a;
    }

    static String arrayText(Object a) {
        StringBuilder sb = new StringBuilder("[");
        int n = java.lang.reflect.Array.getLength(a);
        for (int i = 0; i < n; i++) {
            if (i > 0) {
                sb.append(',');
            }
            Object e = java.lang.reflect.Array.get(a, i);
            if (e instanceof Double d) {
                sb.append(dbits(d));
            } else if (e instanceof Float f) {
                sb.append(fbits(f));
            } else if (e instanceof Character c) {
                sb.append((int) c);
            } else if (e instanceof String s) {
                sb.append(q(s));
            } else {
                sb.append(e);
            }
        }
        return sb.append(']').toString();
    }

    static StringBuilder arraycopyVectors() {
        StringBuilder sb = new StringBuilder();
        Random r = new Random(99);
        String[] prim = {"int", "long", "short", "byte", "char", "boolean", "float", "double"};
        String[] obj = {"String", "Object", "StringObj", "mixed"};
        for (int i = 0; i < 1200; i++) {
            String st, dt;
            int kind = r.nextInt(10);
            if (kind < 4) {
                st = prim[r.nextInt(prim.length)];
                dt = r.nextInt(6) == 0 ? (r.nextBoolean() ? prim[r.nextInt(prim.length)] : obj[r.nextInt(obj.length)]) : st;
            } else if (kind < 8) {
                st = obj[r.nextInt(obj.length)];
                dt = r.nextInt(3) == 0 ? obj[r.nextInt(obj.length)] : st;
                if (r.nextInt(10) == 0) {
                    dt = prim[r.nextInt(prim.length)];
                }
            } else {
                st = r.nextBoolean() ? prim[r.nextInt(prim.length)] : obj[r.nextInt(obj.length)];
                dt = "same";
            }
            int sl = r.nextInt(9);
            int dl = dt.equals("same") ? sl : r.nextInt(9);
            int sp = r.nextInt(sl + 3) - (r.nextInt(8) == 0 ? 3 : 0);
            int dp = r.nextInt(dl + 3) - (r.nextInt(8) == 0 ? 3 : 0);
            int len = r.nextInt(Math.max(sl, dl) + 2) - (r.nextInt(10) == 0 ? 2 : 0);
            if (r.nextInt(3) == 0) {
                // In bounds, to reach the copying itself.
                sp = r.nextInt(sl + 1);
                dp = r.nextInt(dl + 1);
                len = r.nextInt(Math.min(sl - sp, dl - dp) + 1);
            }
            Object src = makeArray(st, sl);
            Object dst = dt.equals("same") ? src : makeArray(dt, dl);
            sb.append("AC ").append(st).append(' ').append(sl).append(' ').append(sp).append(' ').append(dt)
                .append(' ').append(dl).append(' ').append(dp).append(' ').append(len).append(" = ");
            try {
                System.arraycopy(src, sp, dst, dp, len);
                sb.append("ok");
            } catch (RuntimeException e) {
                sb.append(ex(e));
            }
            sb.append(' ').append(arrayText(dst)).append('\n');
        }
        // Object arrays whose elements may not fit the destination.
        String[][] pairs = {{"mixed", "String"}, {"Object", "String"}, {"String", "Object"}, {"StringObj", "String"},
            {"mixed", "StringObj"}};
        for (String[] p : pairs) {
            for (int n = 0; n <= 6; n++) {
                for (int sp = 0; sp <= 1; sp++) {
                    Object src = makeArray(p[0], n + sp);
                    Object dst = makeArray(p[1], n + 1);
                    int fsp = sp;
                    int fn = n;
                    sb.append("AC ").append(p[0]).append(' ').append(n + sp).append(' ').append(sp).append(' ')
                        .append(p[1]).append(' ').append(n + 1).append(' ').append(1).append(' ').append(n)
                        .append(" = ").append(attempt(() -> {
                            System.arraycopy(src, fsp, dst, 1, fn);
                            return "ok";
                        })).append(' ').append(arrayText(dst)).append('\n');
                }
            }
        }
        // Null and non-array arguments.
        sb.append("NULLSRC = ").append(attempt(() -> {
            System.arraycopy(null, 0, new int[1], 0, 0);
            return "ok";
        })).append('\n');
        sb.append("NULLDST = ").append(attempt(() -> {
            System.arraycopy(new int[1], 0, null, 0, 0);
            return "ok";
        })).append('\n');
        sb.append("NOTARRAYSRC = ").append(attempt(() -> {
            System.arraycopy("x", 0, new int[1], 0, 0);
            return "ok";
        })).append('\n');
        sb.append("NOTARRAYDST = ").append(attempt(() -> {
            System.arraycopy(new int[1], 0, "x", 0, 0);
            return "ok";
        })).append('\n');
        return sb;
    }

    // ------------------------------------------------------------------
    // UTF-8 decoding and encoding

    static StringBuilder charsetVectors() throws Exception {
        StringBuilder sb = new StringBuilder();
        Random r = new Random(5);
        List<byte[]> inputs = new ArrayList<>();
        int[][] fixed = {
            {}, {0x41}, {0xc3, 0xa9}, {0xc3}, {0xc0, 0x80}, {0xc1, 0xbf}, {0xe2, 0x82, 0xac}, {0xe2, 0x82},
            {0xe2}, {0xe0, 0x80, 0x80}, {0xe0, 0xa0, 0x80}, {0xed, 0xa0, 0x80}, {0xed, 0xbf, 0xbf},
            {0xf0, 0x9f, 0x98, 0x80}, {0xf0, 0x9f, 0x98}, {0xf0, 0x9f}, {0xf0}, {0xf0, 0x80, 0x80, 0x80},
            {0xf4, 0x90, 0x80, 0x80}, {0xf4, 0x8f, 0xbf, 0xbf}, {0xf5, 0x80, 0x80, 0x80}, {0xf8, 0x88, 0x80, 0x80, 0x80},
            {0xff}, {0xfe, 0xff}, {0x80}, {0xbf, 0x41}, {0xe2, 0x41}, {0xe2, 0x82, 0x41}, {0xf0, 0x9f, 0x41},
            {0xf0, 0x9f, 0x98, 0x41}, {0xf0, 0x41}, {0xf0, 0x90, 0x41}, {0xf4, 0x41}, {0xf7, 0xbf, 0xbf, 0xbf},
        };
        for (int[] f : fixed) {
            byte[] b = new byte[f.length];
            for (int i = 0; i < f.length; i++) {
                b[i] = (byte) f[i];
            }
            inputs.add(b);
        }
        int[] interesting = {0x00, 0x41, 0x7f, 0x80, 0x8f, 0x90, 0x9f, 0xa0, 0xbf, 0xc0, 0xc1, 0xc2, 0xdf, 0xe0, 0xe1,
            0xec, 0xed, 0xee, 0xef, 0xf0, 0xf1, 0xf3, 0xf4, 0xf5, 0xf7, 0xf8, 0xfe, 0xff};
        for (int i = 0; i < 1500; i++) {
            byte[] b = new byte[r.nextInt(12)];
            for (int j = 0; j < b.length; j++) {
                b[j] = (byte) (r.nextInt(3) == 0 ? r.nextInt(256) : interesting[r.nextInt(interesting.length)]);
            }
            inputs.add(b);
        }
        for (byte[] b : inputs) {
            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            bos.write(b);
            sb.append("DEC ").append(hex(b)).append(" = ").append(q(bos.toString())).append('\n');
            InputStreamReader isr = new InputStreamReader(new ByteArrayInputStream(b));
            StringBuilder s = new StringBuilder();
            int c;
            while ((c = isr.read()) >= 0) {
                s.append((char) c);
            }
            sb.append("ISR ").append(hex(b)).append(" = ").append(q(s.toString())).append('\n');
        }
        // Encoding through PrintStream, in one or two writes.
        String units = "aéࠀ￿𐀀􏿿😀";
        for (int i = 0; i < 600; i++) {
            String s1 = randomUnits(r, units);
            String s2 = randomUnits(r, units);
            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            PrintStream ps = new PrintStream(bos);
            ps.print(s1);
            String mid = hex(bos.toByteArray());
            ps.print(s2);
            String end = hex(bos.toByteArray());
            ps.close();
            sb.append("ENC ").append(q(s1)).append(' ').append(q(s2)).append(" = ").append(mid).append(' ')
                .append(end).append(' ').append(hex(bos.toByteArray())).append('\n');
        }
        return sb;
    }

    static String randomUnits(Random r, String units) {
        StringBuilder sb = new StringBuilder();
        int n = r.nextInt(5);
        for (int i = 0; i < n; i++) {
            sb.append(units.charAt(r.nextInt(units.length())));
        }
        return sb.toString();
    }

    // ------------------------------------------------------------------
    // Writers, output streams and PrintStream

    /** A Writer implementing only the abstract methods, as a subclass would. */
    static class LogWriter extends Writer {
        final StringBuilder log = new StringBuilder();

        @Override
        public void write(char[] cbuf, int off, int len) {
            log.append("w(").append(new String(cbuf, off, len)).append(')');
        }

        @Override
        public void flush() {
            log.append("f");
        }

        @Override
        public void close() {
            log.append("c");
        }
    }

    /** An OutputStream implementing only write(int). */
    static class LogOutputStream extends OutputStream {
        final StringBuilder log = new StringBuilder();

        @Override
        public void write(int b) {
            log.append(b).append(';');
        }
    }

    /** An InputStream implementing only read(). */
    static class CountingInputStream extends InputStream {
        int n;
        final int limit;

        CountingInputStream(int limit) {
            this.limit = limit;
        }

        @Override
        public int read() throws IOException {
            if (n == 7 && limit < 0) {
                n++;
                throw new IOException("boom");
            }
            return n < Math.abs(limit) ? (n++ * 37) & 0xff : -1;
        }
    }

    /** A Reader implementing only the abstract methods. */
    static class ArrayReader extends Reader {
        final String s;
        int pos;

        ArrayReader(String s) {
            this.s = s;
        }

        @Override
        public int read(char[] cbuf, int off, int len) {
            if (pos >= s.length()) {
                return -1;
            }
            int n = Math.min(Math.min(len, 2), s.length() - pos);
            s.getChars(pos, pos + n, cbuf, off);
            pos += n;
            return n;
        }

        @Override
        public void close() {
            pos = s.length();
        }
    }

    static StringBuilder writerVectors() throws Exception {
        StringBuilder sb = new StringBuilder();
        Random r = new Random(11);
        String[] strs = {"", "a", "hello", "é😀", null};
        for (int i = 0; i < 500; i++) {
            boolean sub = i % 2 == 1;
            StringWriter sw = new StringWriter();
            LogWriter lw = new LogWriter();
            Writer w = sub ? lw : sw;
            sb.append("WCASE ").append(sub ? "log" : "string").append('\n');
            int nops = 1 + r.nextInt(6);
            for (int j = 0; j < nops; j++) {
                String s = strs[r.nextInt(strs.length)];
                int k = r.nextInt(9);
                int a = r.nextInt(7) - 1;
                int b = r.nextInt(7) - 1;
                String op = switch (k) {
                    case 0 -> "wi " + (a * 40000 + 65) + " = " + attempt(() -> {
                        w.write(a * 40000 + 65);
                        return "ok";
                    });
                    case 1 -> "ws " + q(s) + " = " + attempt(() -> {
                        w.write(s);
                        return "ok";
                    });
                    case 2 -> "wsr " + q(s) + " " + a + " " + b + " = " + attempt(() -> {
                        w.write(s, a, b);
                        return "ok";
                    });
                    case 3 -> "wc " + q(s) + " " + a + " " + b + " = " + attempt(() -> {
                        char[] cs = s == null ? null : s.toCharArray();
                        w.write(cs, a, b);
                        return "ok";
                    });
                    case 4 -> "ac " + q(s) + " = " + attempt(() -> "" + (w.append(s) == w));
                    case 5 -> "acr " + q(s) + " " + a + " " + b + " = " + attempt(() -> "" + (w.append(s, a, b) == w));
                    case 6 -> "ach " + (a + 66) + " = " + attempt(() -> "" + (w.append((char) (a + 66)) == w));
                    case 7 -> "wca " + q(s) + " = " + attempt(() -> {
                        w.write(s == null ? null : s.toCharArray());
                        return "ok";
                    });
                    default -> "fl = " + attempt(() -> {
                        w.flush();
                        return "ok";
                    });
                };
                sb.append(op).append('\n');
            }
            sb.append("RESULT ").append(q(sub ? lw.log.toString() : sw.toString())).append('\n');
        }
        // OutputStream base class and ByteArrayOutputStream.
        for (int i = 0; i < 200; i++) {
            boolean sub = i % 2 == 1;
            LogOutputStream los = new LogOutputStream();
            ByteArrayOutputStream bos = new ByteArrayOutputStream(r.nextInt(3));
            OutputStream os = sub ? los : bos;
            sb.append("OCASE ").append(sub ? "log" : "bytes").append('\n');
            int nops = 1 + r.nextInt(6);
            for (int j = 0; j < nops; j++) {
                byte[] data = new byte[r.nextInt(5)];
                for (int t = 0; t < data.length; t++) {
                    data[t] = (byte) (r.nextInt(256));
                }
                int a = r.nextInt(6) - 1;
                int b = r.nextInt(6) - 1;
                int k = r.nextInt(4);
                String op = switch (k) {
                    case 0 -> "wi " + (a * 100 + 7) + " = " + attempt(() -> {
                        os.write(a * 100 + 7);
                        return "ok";
                    });
                    case 1 -> "wa " + hex(data) + " = " + attempt(() -> {
                        os.write(data);
                        return "ok";
                    });
                    case 2 -> "war " + hex(data) + " " + a + " " + b + " = " + attempt(() -> {
                        os.write(data, a, b);
                        return "ok";
                    });
                    default -> {
                        if (!sub && r.nextBoolean()) {
                            bos.reset();
                            yield "reset";
                        }
                        yield "fl = " + attempt(() -> {
                            os.flush();
                            return "ok";
                        });
                    }
                };
                sb.append(op).append('\n');
            }
            if (sub) {
                sb.append("RESULT ").append(q(los.log.toString())).append('\n');
            } else {
                sb.append("RESULT ").append(hex(bos.toByteArray())).append(' ').append(bos.size()).append(' ')
                    .append(q(bos.toString())).append(' ').append(q(bos.toString(1))).append('\n');
            }
        }
        // InputStream base class and ByteArrayInputStream.
        for (int i = 0; i < 200; i++) {
            int variant = i % 3;
            byte[] data = new byte[r.nextInt(12)];
            for (int t = 0; t < data.length; t++) {
                data[t] = (byte) r.nextInt(256);
            }
            int boff = r.nextInt(4), blen = r.nextInt(14) - 1;
            int limit = r.nextInt(15) * (r.nextInt(4) == 0 ? -1 : 1);
            InputStream is = switch (variant) {
                case 0 -> new CountingInputStream(limit);
                case 1 -> new ByteArrayInputStream(data);
                default -> new ByteArrayInputStream(data, Math.min(boff, data.length), blen);
            };
            sb.append("ICASE ").append(variant).append(' ').append(limit).append(' ').append(hex(data)).append(' ')
                .append(Math.min(boff, data.length)).append(' ').append(blen).append('\n');
            int nops = 1 + r.nextInt(8);
            for (int j = 0; j < nops; j++) {
                int k = r.nextInt(8);
                int arrLen = r.nextInt(6);
                int a = r.nextInt(arrLen + 2) - 1;
                int b = r.nextInt(arrLen + 2) - 1;
                String op = switch (k) {
                    case 0, 1 -> "r = " + attempt(() -> "" + is.read());
                    case 2 -> "ra " + arrLen + " = " + attempt(() -> {
                        byte[] buf = new byte[arrLen];
                        int n = is.read(buf);
                        return n + " " + hex(buf);
                    });
                    case 3 -> "rar " + arrLen + " " + a + " " + b + " = " + attempt(() -> {
                        byte[] buf = new byte[arrLen];
                        int n = is.read(buf, a, b);
                        return n + " " + hex(buf);
                    });
                    case 4 -> "s " + (a * 3) + " = " + attempt(() -> "" + is.skip(a * 3));
                    case 5 -> "av = " + attempt(() -> "" + is.available());
                    case 6 -> "m = " + attempt(() -> {
                        is.mark(5);
                        return "ok " + is.markSupported();
                    });
                    default -> "R = " + attempt(() -> {
                        is.reset();
                        return "ok";
                    });
                };
                sb.append(op).append('\n');
            }
            sb.append("END\n");
        }
        // Reader base class.
        for (int i = 0; i < 100; i++) {
            String s = randomUnits(r, "abc\néxyz");
            s = s + randomUnits(r, "abc\néxyz");
            ArrayReader ar = new ArrayReader(s);
            sb.append("BCASE ").append(q(s)).append('\n');
            int nops = 1 + r.nextInt(8);
            for (int j = 0; j < nops; j++) {
                int k = r.nextInt(7);
                long n = r.nextInt(6) - 1;
                String op = switch (k) {
                    case 0, 1 -> "r = " + attempt(() -> "" + ar.read());
                    case 2 -> "s " + n + " = " + attempt(() -> "" + ar.skip(n));
                    case 3 -> "m = " + attempt(() -> {
                        ar.mark(3);
                        return "ok";
                    });
                    case 4 -> "R = " + attempt(() -> {
                        ar.reset();
                        return "ok";
                    });
                    case 5 -> "y = " + attempt(() -> "" + ar.ready() + " " + ar.markSupported());
                    default -> "ra = " + attempt(() -> {
                        char[] buf = new char[3];
                        int got = ar.read(buf);
                        return got + " " + q(new String(buf));
                    });
                };
                sb.append(op).append('\n');
            }
            sb.append("END\n");
        }
        // PrintStream output.
        Object[] objs = {null, "str", 42, -7L, 'c', true, 1.5, 0.1f, (short) -3, (byte) 9};
        for (int i = 0; i < 300; i++) {
            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            PrintStream ps = new PrintStream(bos, i % 2 == 0);
            StringBuilder ops = new StringBuilder();
            int nops = 1 + r.nextInt(6);
            for (int j = 0; j < nops; j++) {
                int k = r.nextInt(16);
                int oi = r.nextInt(objs.length);
                Object o = objs[oi];
                String s = strs[r.nextInt(strs.length)];
                double d = r.nextInt(4) == 0 ? r.nextDouble() * 1e10 : r.nextInt(100) / 8.0;
                switch (k) {
                    case 0 -> {
                        ps.print(s);
                        ops.append(" ps ").append(q(s));
                    }
                    case 1 -> {
                        ps.println(s);
                        ops.append(" pls ").append(q(s));
                    }
                    case 2 -> {
                        ps.print(o);
                        ops.append(" po ").append(oi);
                    }
                    case 3 -> {
                        ps.println(o);
                        ops.append(" plo ").append(oi);
                    }
                    case 4 -> {
                        ps.println();
                        ops.append(" pl");
                    }
                    case 5 -> {
                        ps.print(d);
                        ops.append(" pd ").append(dbits(d));
                    }
                    case 6 -> {
                        ps.println((int) d);
                        ops.append(" pli ").append((int) d);
                    }
                    case 7 -> {
                        ps.print((long) d * 1000003);
                        ops.append(" pj ").append((long) d * 1000003);
                    }
                    case 8 -> {
                        ps.print((float) d);
                        ops.append(" pf ").append(fbits((float) d));
                    }
                    case 9 -> {
                        ps.print(d > 5);
                        ops.append(" pz ").append(d > 5);
                    }
                    case 10 -> {
                        ps.print((char) (0x41 + (int) d));
                        ops.append(" pc ").append(0x41 + (int) d);
                    }
                    case 11 -> {
                        ps.print(s == null ? new char[0] : s.toCharArray());
                        ops.append(" pca ").append(q(s == null ? "" : s));
                    }
                    case 12 -> {
                        ps.append(s);
                        ops.append(" ap ").append(q(s));
                    }
                    case 13 -> {
                        ps.write((int) d);
                        ops.append(" wi ").append((int) d);
                    }
                    case 14 -> {
                        ps.println(d);
                        ops.append(" pld ").append(dbits(d));
                    }
                    default -> {
                        if (r.nextInt(3) == 0) {
                            ps.close();
                            ops.append(" close");
                        } else {
                            ps.flush();
                            ops.append(" flush");
                        }
                    }
                }
            }
            sb.append("PCASE ").append(i % 2 == 0).append(ops).append(" = ").append(hex(bos.toByteArray()))
                .append(' ').append(ps.checkError()).append('\n');
        }
        return sb;
    }

    // ------------------------------------------------------------------
    // File paths, system properties, Date, and the reader use of
    // WKBHexFileReader

    static String fileText(File f) {
        return q(f.getPath()) + " " + q(f.getName()) + " " + q(f.getParent()) + " " + f.isAbsolute() + " "
            + f.hashCode() + " " + q(f.toString());
    }

    static StringBuilder miscVectors() throws Exception {
        StringBuilder sb = new StringBuilder();
        String[] paths = {"", "/", "//", "a", "a/", "a//b", "/a/b/", "///a///b///", "a/b/c.txt", ".", "..", "/..",
            "a/./b", "/a", "a\u00e9/b"};
        for (String p : paths) {
            sb.append("FILE ").append(q(p)).append(" = ").append(fileText(new File(p))).append('\n');
        }
        String[] parents = {null, "", "/", "a", "a/", "/x//"};
        String[] children = {"", "b", "/b", "b/", "//b//c"};
        for (String p : parents) {
            for (String c : children) {
                sb.append("FILE2 ").append(q(p)).append(' ').append(q(c)).append(" = ")
                    .append(fileText(new File(p, c))).append('\n');
                File pf = p == null ? null : new File(p);
                sb.append("FILE3 ").append(q(p)).append(' ').append(q(c)).append(" = ")
                    .append(fileText(new File(pf, c))).append('\n');
            }
        }
        sb.append("CMP ").append(new File("a").compareTo(new File("b"))).append(' ')
            .append(new File("a/").equals(new File("a"))).append(' ').append(new File("b").compareTo(new File("ab")))
            .append('\n');
        for (String key : new String[] {"line.separator", "file.separator", "path.separator", "jts.overlay",
            "jts.relate", "jts.debug", null, ""}) {
            sb.append("PROP ").append(q(key)).append(" = ")
                .append(attempt(() -> q(System.getProperty(key)))).append(' ')
                .append(attempt(() -> q(System.getProperty(key, "def")))).append('\n');
        }
        sb.append("LSEP ").append(q(System.lineSeparator())).append('\n');
        for (long t : new long[] {0, 1, -1, 1234567890123L, Long.MIN_VALUE, Long.MAX_VALUE, 1L << 32, -(1L << 33)}) {
            Date d = new Date(t);
            sb.append("DATE ").append(t).append(" = ").append(d.hashCode()).append(' ').append(d.getTime())
                .append(' ').append(d.compareTo(new Date(0))).append(' ').append(d.equals(new Date(t))).append(' ')
                .append(d.before(new Date(1))).append(' ').append(d.after(new Date(1))).append('\n');
        }
        // WKBHexFileReader's isAtEndOfFile: mark, tokenize, reset, then read
        // lines.
        String[] files = {"", "\n", "  \n\n", "0101\n", "0101\n\n0202", "  0101  \r\n0202\n   ", "\n\n\nAB",
            "x".repeat(1200) + "\n", "/comment\nAB", "'q'\n", "\n".repeat(1100) + "AB"};
        for (String f : files) {
            BufferedReader br = new BufferedReader(new StringReader(f));
            sb.append("HEXFILE ").append(q(f)).append(" =");
            for (int i = 0; i < 10; i++) {
                String r = attempt(() -> {
                    br.mark(1000);
                    StreamTokenizer st = new StreamTokenizer(br);
                    int type = st.nextToken();
                    if (type == StreamTokenizer.TT_EOF) {
                        return "EOF";
                    }
                    br.reset();
                    return q(br.readLine());
                });
                sb.append(' ').append(r);
                if (!r.startsWith("\"")) {
                    break;
                }
            }
            sb.append('\n');
        }
        // WKTFileReader's isAtEndOfFile.
        for (String f : files) {
            BufferedReader br = new BufferedReader(new StringReader(f), 7);
            sb.append("WKTFILE ").append(q(f)).append(" =");
            for (int i = 0; i < 10; i++) {
                String r = attempt(() -> {
                    int ch;
                    do {
                        br.mark(1);
                        ch = br.read();
                        if (ch < 0) {
                            return "EOF";
                        }
                    } while (Character.isWhitespace(ch));
                    br.reset();
                    StreamTokenizer st = new StreamTokenizer(br);
                    return st.nextToken() + ":" + q(st.sval);
                });
                sb.append(' ').append(r);
                if (r.equals("EOF") || r.startsWith("!")) {
                    break;
                }
            }
            sb.append('\n');
        }
        return sb;
    }
}
