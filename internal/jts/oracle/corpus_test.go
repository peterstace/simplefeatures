package oracle_test

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jts/java"
)

// TestCorpus runs the operation battery on geometries from simplefeatures'
// own tests and on generated geometries, and binary operations on pairs of
// them.
func TestCorpus(t *testing.T) {
	d := startDriver(t)
	wkts := append(scrapeWKT(t), generatedWKT()...)
	pairs := 2000
	if testing.Short() {
		wkts = wkts[:len(wkts)/10]
		pairs = 200
	}
	st := stats{}
	var hexes []string
	for _, w := range wkts {
		mustOK(t, d.call("pm", "FLOATING"))
		if hex, ok := toWKB(d, "x", w); ok {
			hexes = append(hexes, hex)
		}
	}
	t.Logf("%d geometries", len(hexes))
	for i, hex := range hexes {
		runCase(t, d, st, fmt.Sprintf("corpus %d", i), []string{hex})
	}
	rnd := rand.New(rand.NewSource(1))
	for i := 0; i < pairs && len(hexes) > 1; i++ {
		a, b := hexes[rnd.Intn(len(hexes))], hexes[rnd.Intn(len(hexes))]
		runCase(t, d, st, fmt.Sprintf("pair %d", i), []string{a, b})
	}
	t.Logf("\n%s", st)
}

// runCase loads the geometries (as a, b, ...) on both sides with a floating
// precision model and runs the battery.
func runCase(t *testing.T, d *driver, st stats, desc string, hexes []string) {
	t.Helper()
	p := newPort()
	mustOK(t, d.call("reset"))
	mustOK(t, d.call("pm", "FLOATING"))
	var names []string
	for i, hex := range hexes {
		name := string(rune('a' + i))
		mustOK(t, d.call("load", name, hex))
		if failure := p.load(name, hex); failure != "" {
			st.add("load", outcome(failure))
			if outcome(failure) == mismatched {
				t.Errorf("%s: port failed to load %s: %s", desc, name, failure)
			}
			return
		}
		names = append(names, name)
	}
	compareBattery(t, d, p, st, desc, names)
}

// compareBattery runs each battery operation on both sides and compares.
func compareBattery(t *testing.T, d *driver, p *port, st stats, desc string, names []string) {
	t.Helper()
	for _, op := range battery(names) {
		reply := d.call(append([]string{"op"}, op...)...)
		hseq, want, ok := strings.Cut(reply, " ")
		if !ok {
			t.Fatalf("driver reply: %s", reply)
		}
		seq, err := strconv.ParseInt(hseq, 10, 32)
		if err != nil {
			t.Fatalf("driver reply: %s", reply)
		}
		java.SetIdentityHashSeq(int32(seq) + 1)
		got := p.op(append([]string{"op"}, op...))
		switch {
		case strings.HasPrefix(got, "UNTRANSLATED"):
			st.add(op[0], untranslated)
		case got == want:
			st.add(op[0], matched)
		default:
			st.add(op[0], mismatched)
			t.Errorf("%s: op %v\n%s", desc, op, diff(want, got))
		}
	}
}

var wktLiteral = regexp.MustCompile(
	"[\"`]((?:POINT|LINESTRING|POLYGON|MULTIPOINT|MULTILINESTRING|MULTIPOLYGON|GEOMETRYCOLLECTION)" +
		"(?: ?(?:Z|M|ZM))? ?(?:EMPTY|\\([^\"`]*\\)))[\"`]")

// scrapeWKT collects the WKT literals in simplefeatures' geom tests.
func scrapeWKT(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("../../../geom/*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	var out []string
	for _, f := range files {
		buf, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range wktLiteral.FindAllStringSubmatch(string(buf), -1) {
			if !seen[m[1]] {
				seen[m[1]] = true
				out = append(out, m[1])
			}
		}
	}
	sort.Strings(out)
	return out
}

// generatedWKT returns deterministic random geometries: polygons and lines
// on integer grids (lots of coincident vertices and collinear edges) and with
// full-precision coordinates (lots of nearly-degenerate intersections).
func generatedWKT() []string {
	rnd := rand.New(rand.NewSource(42))
	var out []string
	coord := func(grid bool) string {
		if grid {
			return fmt.Sprintf("%d %d", rnd.Intn(10), rnd.Intn(10))
		}
		x, y := rnd.Float64()*10, rnd.Float64()*10
		return strconv.FormatFloat(x, 'g', -1, 64) + " " + strconv.FormatFloat(y, 'g', -1, 64)
	}
	ring := func(grid bool, n int) string {
		// A star-shaped ring around a random centre, so it is usually valid.
		cx, cy := 2+rnd.Float64()*6, 2+rnd.Float64()*6
		if grid {
			cx, cy = math.Round(cx), math.Round(cy)
		}
		pts := make([]string, 0, n+1)
		for i := 0; i < n; i++ {
			a := 2 * math.Pi * (float64(i) + rnd.Float64()*0.8) / float64(n)
			r := 0.5 + rnd.Float64()*3
			x, y := cx+r*math.Cos(a), cy+r*math.Sin(a)
			if grid {
				x, y = math.Round(x), math.Round(y)
			}
			pts = append(pts, strconv.FormatFloat(x, 'g', -1, 64)+" "+strconv.FormatFloat(y, 'g', -1, 64))
		}
		pts = append(pts, pts[0])
		return "(" + strings.Join(pts, ", ") + ")"
	}
	for i := 0; i < 150; i++ {
		grid := i%2 == 0
		n := 3 + rnd.Intn(12)
		out = append(out, "POLYGON("+ring(grid, n)+")")
		var pts []string
		for j := 0; j < 2+rnd.Intn(6); j++ {
			pts = append(pts, coord(grid))
		}
		out = append(out, "LINESTRING("+strings.Join(pts, ", ")+")")
		out = append(out, "MULTIPOINT(("+coord(grid)+"), ("+coord(grid)+"))")
		out = append(out, "MULTIPOLYGON(("+ring(grid, n)+"), ("+ring(grid, 3+rnd.Intn(5))+"))")
	}
	return out
}
