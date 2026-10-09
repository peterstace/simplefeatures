import com.sun.tools.javac.code.Symbol.ClassSymbol;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Stream;

/**
 * The identifiers that the hand-written runtime packages declare. The
 * translator only emits references to runtime identifiers that exist, so the
 * generated package always compiles; everything else becomes a recorded stub.
 */
final class Runtime {
    private final Set<String> types = new HashSet<>();
    private final Set<String> funcs = new HashSet<>();
    private final Map<String, Set<String>> methods = new HashMap<>();
    private final Map<String, List<String>> embeds = new HashMap<>();

    private static final Pattern TYPE = Pattern.compile("^type (\\w+)(\\[[^\\]]*\\])? (struct|interface)?.*");
    private static final Pattern FUNC = Pattern.compile("^func (\\w+)[\\[(].*");
    private static final Pattern METHOD = Pattern.compile("^func \\(\\w+ \\*?(\\w+)(\\[[^\\]]*\\])?\\) (\\w+)[\\[(].*");
    private static final Pattern IFACE_METHOD = Pattern.compile("^\\t(\\w+)\\(.*");
    private static final Pattern EMBED = Pattern.compile("^\\t\\*?([A-Z]\\w*)(\\s*//.*)?$");
    private static final Pattern FIELD = Pattern.compile("^\\t(F_\\w+)\\s+\\S.*");
    private static final Pattern VAR = Pattern.compile("^(?:var|const) (\\w+)\\b.*");

    /** Returns the qualified Go name for a non-JTS class, or null if it has no runtime package. */
    static String qualifiedName(ClassSymbol c) {
        String pkg = c.packge().fullname.toString();
        String flat = c.flatname.toString();
        String local = flat.substring(pkg.length() + 1).replace('$', '_');
        if (pkg.startsWith("java.") || pkg.startsWith("javax.")) {
            return "java." + local;
        }
        if (pkg.equals("junit.framework")) {
            return "junit." + local;
        }
        if (pkg.equals("org.junit")) {
            return "junit.J4_" + local;
        }
        return null;
    }

    static Runtime scan(Path... dirs) throws IOException {
        Runtime rt = new Runtime();
        for (Path dir : dirs) {
            if (!Files.isDirectory(dir)) {
                continue;
            }
            String pkg = dir.getFileName().toString();
            try (Stream<Path> files = Files.list(dir)) {
                for (Path f : files.filter(p -> p.toString().endsWith(".go") && !p.toString().endsWith("_test.go"))
                        .toList()) {
                    List<String> lines = Files.readAllLines(f);
                    // Files under development are behind a build tag.
                    if (lines.stream().anyMatch(l -> l.startsWith("//go:build"))) {
                        continue;
                    }
                    rt.scanFile(pkg, lines);
                }
            }
        }
        return rt;
    }

    private void scanFile(String pkg, List<String> lines) {
        String block = null;
        boolean blockIsIface = false;
        for (String line : lines) {
            if (block != null) {
                if (line.equals("}")) {
                    block = null;
                    continue;
                }
                Matcher m;
                if (blockIsIface && (m = IFACE_METHOD.matcher(line)).matches()) {
                    methods.computeIfAbsent(block, k -> new HashSet<>()).add(m.group(1));
                } else if ((m = EMBED.matcher(line)).matches()) {
                    embeds.computeIfAbsent(block, k -> new ArrayList<>()).add(pkg + "." + m.group(1));
                } else if (!blockIsIface && (m = FIELD.matcher(line)).matches()) {
                    methods.computeIfAbsent(block, k -> new HashSet<>()).add(m.group(1));
                }
                continue;
            }
            Matcher m = TYPE.matcher(line);
            if (m.matches()) {
                String name = pkg + "." + m.group(1);
                types.add(name);
                if (line.endsWith("{")) {
                    block = name;
                    blockIsIface = "interface".equals(m.group(3));
                }
                continue;
            }
            m = METHOD.matcher(line);
            if (m.matches()) {
                methods.computeIfAbsent(pkg + "." + m.group(1), k -> new HashSet<>()).add(m.group(3));
                continue;
            }
            m = FUNC.matcher(line);
            if (m.matches()) {
                funcs.add(pkg + "." + m.group(1));
                continue;
            }
            m = VAR.matcher(line);
            if (m.matches()) {
                funcs.add(pkg + "." + m.group(1));
            }
        }
    }

    boolean hasType(String qualified) {
        return types.contains(qualified);
    }

    boolean hasFunc(String qualified) {
        return funcs.contains(qualified);
    }

    boolean hasMethod(String qualifiedType, String method) {
        Set<String> seen = new HashSet<>();
        return hasMethod(qualifiedType, method, seen);
    }

    private boolean hasMethod(String t, String method, Set<String> seen) {
        if (!seen.add(t)) {
            return false;
        }
        if (methods.getOrDefault(t, Set.of()).contains(method)) {
            return true;
        }
        for (String e : embeds.getOrDefault(t, List.of())) {
            if (hasMethod(e, method, seen)) {
                return true;
            }
        }
        return false;
    }
}
