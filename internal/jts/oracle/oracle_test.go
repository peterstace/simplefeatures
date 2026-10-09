package oracle

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const testXMLDir = cacheDir + "/jts-1.20.0/modules/tests/src/test/resources/testxml"

type xmlRun struct {
	PM *struct {
		Type  string `xml:"type,attr"`
		Scale string `xml:"scale,attr"`
	} `xml:"precisionModel"`
	Cases []struct {
		Desc string   `xml:"desc"`
		A    *xmlGeom `xml:"a"`
		B    *xmlGeom `xml:"b"`
	} `xml:"case"`
}

type xmlGeom struct {
	File string `xml:"file,attr"`
	Text string `xml:",chardata"`
}

// stats counts outcomes per operation.
type stats map[string]*[3]int

const (
	matched = iota
	mismatched
	untranslated
)

func (s stats) add(op string, outcome int) {
	c, ok := s[op]
	if !ok {
		c = new([3]int)
		s[op] = c
	}
	c[outcome]++
}

func (s stats) String() string {
	var ops []string
	for op := range s {
		ops = append(ops, op)
	}
	sort.Strings(ops)
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-16s %8s %8s %12s\n", "op", "match", "MISMATCH", "untranslated")
	for _, op := range ops {
		c := s[op]
		fmt.Fprintf(&sb, "%-16s %8d %8d %12d\n", op, c[matched], c[mismatched], c[untranslated])
	}
	return sb.String()
}

func TestXML(t *testing.T) {
	d := startDriver(t)
	files, err := filepath.Glob(filepath.Join(testXMLDir, "*", "*.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("JTS XML tests not found")
	}
	st := stats{}
	for _, file := range files {
		rel, _ := filepath.Rel(testXMLDir, file)
		t.Run(rel, func(t *testing.T) {
			runXMLFile(t, d, file, st)
		})
	}
	t.Logf("\n%s", st)
}

func runXMLFile(t *testing.T, d *driver, file string, st stats) {
	buf, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var run xmlRun
	if err := xml.Unmarshal(buf, &run); err != nil {
		t.Skipf("unparseable XML: %v", err)
	}
	pm := []string{"pm", "FLOATING"}
	if run.PM != nil {
		switch {
		case run.PM.Scale != "":
			scale, err := strconv.ParseFloat(run.PM.Scale, 64)
			if err != nil {
				t.Fatal(err)
			}
			pm = []string{"pm", "FIXED", bits(scale)}
		case run.PM.Type == "FLOATING_SINGLE":
			pm = []string{"pm", "FLOATING_SINGLE"}
		}
	}
	for i, c := range run.Cases {
		p := newPort()
		mustOK(t, d.call("reset"))
		mustOK(t, d.call(pm...))
		if failure := capture(func() string { p.setPrecisionModel(pm); return "" }); failure != "" {
			st.add("pm", outcome(failure))
			if outcome(failure) == mismatched {
				t.Errorf("case %d: port failed to set precision model: %s", i, failure)
			}
			continue
		}
		names := []string{}
		for _, g := range []struct {
			name string
			x    *xmlGeom
		}{{"a", c.A}, {"b", c.B}} {
			if g.x == nil || g.x.File != "" {
				continue
			}
			hex, ok := toWKB(d, g.name, g.x.Text)
			if !ok {
				continue
			}
			mustOK(t, d.call("load", g.name, hex))
			if failure := p.load(g.name, hex); failure != "" {
				st.add("load", outcome(failure))
				if outcome(failure) == mismatched {
					t.Errorf("case %d (%s): port failed to load %s: %s", i, strings.TrimSpace(c.Desc), g.name, failure)
				}
				continue
			}
			names = append(names, g.name)
		}
		compareBattery(t, d, p, st, fmt.Sprintf("case %d (%s)", i, strings.TrimSpace(c.Desc)), names)
	}
}

func outcome(reply string) int {
	if strings.HasPrefix(reply, "UNTRANSLATED") {
		return untranslated
	}
	return mismatched
}

// toWKB has the driver parse a test geometry (WKT, or hex WKB as some test
// files use) and returns it as hex WKB.
func toWKB(d *driver, name, text string) (string, bool) {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return "", false
	}
	if strings.Trim(text, "0123456789abcdefABCDEF") == "" {
		return text, true
	}
	reply := d.call("wkt", name, strings.ReplaceAll(text, " ", "_"))
	if !strings.HasPrefix(reply, "OK ") {
		return "", false
	}
	return strings.TrimPrefix(reply, "OK "), true
}

func mustOK(t *testing.T, reply string) {
	t.Helper()
	if reply != "OK" {
		t.Fatalf("driver: %s", reply)
	}
}

// battery is the list of operations run on every test case.
func battery(names []string) [][]string {
	var ops [][]string
	f := func(xs ...string) { ops = append(ops, xs) }
	for _, a := range names {
		f("dump", a)
		f("isValid", a)
		f("isSimple", a)
		f("area", a)
		f("length", a)
		f("centroid", a)
		f("interiorPoint", a)
		f("convexHull", a)
		f("boundary", a)
		f("envelope", a)
		f("normalize", a)
		f("unaryUnion", a)
		for _, dist := range []float64{0, 1, 10, -1} {
			f("buffer", a, bits(dist), "8", "1", "1", bits(5))
		}
		f("buffer", a, bits(2.5), "3", "2", "2", bits(2))
		f("simplifyDP", a, bits(1))
		f("simplifyTP", a, bits(1))
	}
	if len(names) == 2 {
		a, b := names[0], names[1]
		f("relate", a, b)
		for _, p := range []string{"intersects", "contains", "within", "covers", "coveredBy", "crosses",
			"disjoint", "equalsTopo", "overlaps", "touches"} {
			f("relatePredicate", a, b, p)
		}
		for _, p := range []string{"intersects", "contains", "containsProperly", "covers", "coveredBy",
			"crosses", "disjoint", "overlaps", "touches", "within"} {
			f("prepared", a, b, p)
		}
		for opCode := 1; opCode <= 4; opCode++ {
			f("overlay", a, b, strconv.Itoa(opCode))
		}
		f("distance", a, b)
		f("hausdorff", a, b)
	}
	return ops
}

// diff describes how got differs from want, showing the region around the
// first difference.
func diff(want, got string) string {
	i := 0
	for i < len(want) && i < len(got) && want[i] == got[i] {
		i++
	}
	lo := int(math.Max(0, float64(i-80)))
	clip := func(s string) string {
		hi := i + 160
		if hi > len(s) {
			hi = len(s)
		}
		if lo > len(s) {
			return ""
		}
		return s[lo:hi]
	}
	return fmt.Sprintf("  first difference at byte %d\n  want: ...%s\n  got:  ...%s", i, clip(want), clip(got))
}
