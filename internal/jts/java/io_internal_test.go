package java

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"
)

// The vectors in testdata/io_*.txt come from the JVM; see
// internal/jts/tools/jdkvectors/IoVectors.java for their format.

type vectorLine struct {
	no     int
	fields []string
}

func readVectors(t *testing.T, name string) []vectorLine {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []vectorLine
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	no := 0
	for sc.Scan() {
		no++
		out = append(out, vectorLine{no, splitVectorFields(sc.Text())})
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// splitVectorFields splits a line at spaces, keeping each quoted string
// (which never contains a raw '"') as one field.
func splitVectorFields(line string) []string {
	var out []string
	for len(line) > 0 {
		var f string
		if line[0] == '"' {
			end := strings.IndexByte(line[1:], '"') + 2
			f, line = line[:end], line[end:]
		} else if i := strings.IndexByte(line, ' '); i >= 0 {
			f, line = line[:i], line[i:]
		} else {
			f, line = line, ""
		}
		out = append(out, f)
		line = strings.TrimPrefix(line, " ")
	}
	return out
}

// unq decodes a quoted vector string, or null.
func unq(f string) *String {
	if f == "null" {
		return nil
	}
	if len(f) < 2 || f[0] != '"' || f[len(f)-1] != '"' {
		panic("not a quoted string: " + f)
	}
	f = f[1 : len(f)-1]
	var u []uint16
	for i := 0; i < len(f); i++ {
		if f[i] == '\\' {
			v, err := strconv.ParseUint(f[i+2:i+6], 16, 16)
			if err != nil {
				panic(err)
			}
			u = append(u, uint16(v))
			i += 5
		} else {
			u = append(u, uint16(f[i]))
		}
	}
	return FromUnits(u)
}

// quote encodes s as the vectors do.
func quote(s *String) string {
	if s == nil {
		return "null"
	}
	var sb strings.Builder
	sb.WriteByte('"')
	for _, c := range s.units {
		if c >= 0x20 && c < 0x7f && c != '"' && c != '\\' {
			sb.WriteByte(byte(c))
		} else {
			fmt.Fprintf(&sb, "\\u%04x", c)
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

func atoi(t *testing.T, s string) int64 {
	t.Helper()
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func hexBits(t *testing.T, s string) uint64 {
	t.Helper()
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func hexBytes(t *testing.T, s string) []int8 {
	t.Helper()
	if !strings.HasPrefix(s, "x") {
		t.Fatalf("not hex bytes: %q", s)
	}
	s = s[1:]
	out := make([]int8, len(s)/2)
	for i := range out {
		v, err := strconv.ParseUint(s[2*i:2*i+2], 16, 8)
		if err != nil {
			t.Fatal(err)
		}
		out[i] = int8(v)
	}
	return out
}

func fmtBytes(b []int8) string {
	var sb strings.Builder
	sb.WriteByte('x')
	for _, v := range b {
		fmt.Fprintf(&sb, "%02x", uint8(v))
	}
	return sb.String()
}

// attemptJava runs f and returns its result, or the Java exception it threw
// formatted as the vectors format it.
func attemptJava(f func() string) (res string) {
	defer func() {
		if r := recover(); r != nil {
			if !Throwable_isInstance(r) {
				panic(r)
			}
			name := className(r)
			if name == "java.lang.RuntimeException" {
				// arraycopy's stand-in for ArrayStoreException.
				msg := Throwable_fromAny(r).GetMessage().Go()
				if p := "java.lang.ArrayStoreException: "; strings.HasPrefix(msg, p) {
					res = "!java.lang.ArrayStoreException " + quote(FromGo(strings.TrimPrefix(msg, p)))
					return
				}
			}
			res = "!" + name + " " + quote(Throwable_fromAny(r).GetMessage())
		}
	}()
	return f()
}

// npeMessage matches a NullPointerException in a result.
var npeMessage = regexp.MustCompile(`!java\.lang\.NullPointerException (null|"[^"]*")`)

// sameResult compares a result with a vector. NullPointerException messages
// are not compared: the JVM's describe the failing bytecode.
func sameResult(got, want string) bool {
	const npe = "!java.lang.NullPointerException"
	return npeMessage.ReplaceAllString(got, npe) == npeMessage.ReplaceAllString(want, npe)
}

// ---------------------------------------------------------------------------
// StreamTokenizer

func applyTokenizerConfig(t *testing.T, st *StreamTokenizer, ops string) {
	t.Helper()
	for _, op := range strings.Fields(ops) {
		k, rest := op[0], op[1:]
		var a, b int32
		if p := strings.Split(rest, ","); len(p) == 2 {
			a, b = int32(atoi(t, p[0])), int32(atoi(t, p[1]))
		} else if rest != "" {
			a = int32(atoi(t, rest))
		}
		switch k {
		case 'R':
			st.ResetSyntax()
		case 'W':
			st.WordChars_int_int(a, b)
		case 'S':
			st.WhitespaceChars_int_int(a, b)
		case 'O':
			st.OrdinaryChars_int_int(a, b)
		case 'o':
			st.OrdinaryChar_int(a)
		case 'C':
			st.CommentChar_int(a)
		case 'Q':
			st.QuoteChar_int(a)
		case 'P':
			st.ParseNumbers()
		case 'E':
			st.EolIsSignificant_boolean(a != 0)
		case 'A':
			st.SlashStarComments_boolean(a != 0)
		case 'B':
			st.SlashSlashComments_boolean(a != 0)
		case 'L':
			st.LowerCaseMode_boolean(a != 0)
		default:
			t.Fatalf("bad op %q", op)
		}
	}
}

func TestStreamTokenizerVectors(t *testing.T) {
	configs := map[string]string{}
	var st *StreamTokenizer
	var caseLine int
	var prevNval uint64
	started := false
	cases, tokens := 0, 0
	for _, l := range readVectors(t, "io_tokenizer.txt") {
		f := l.fields
		switch f[0] {
		case "CONFIG":
			configs[f[1]] = unq(f[2]).Go()
		case "CASE":
			input := unq(f[3])
			switch f[2] {
			case "r":
				st = StreamTokenizer_new_Reader(&StringReader_new(input).Reader)
			case "b":
				st = StreamTokenizer_new_Reader(&BufferedReader_new_Reader_int(&StringReader_new(input).Reader, 3).Reader)
			case "s":
				b := make([]int8, len(input.units))
				for i, c := range input.units {
					b[i] = int8(c)
				}
				st = StreamTokenizer_new_InputStream(&ByteArrayInputStream_new_byteArr(b).InputStream)
			}
			applyTokenizerConfig(t, st, configs[f[1]])
			caseLine = l.no
			prevNval = 0
			started = false
			cases++
		case "P":
			st.PushBack()
		case "T":
			tokens++
			if !started {
				started = true
			}
			ttype := st.NextToken()
			nvalBits := math.Float64bits(st.F_nval)
			want := prevNval
			if f[3] != "=" {
				want = hexBits(t, f[3])
			}
			prevNval = want
			if ttype != int32(atoi(t, f[1])) || st.F_ttype != ttype || st.Lineno() != int32(atoi(t, f[2])) ||
				!sameNval(nvalBits, want) || quote(st.F_sval) != f[4] {
				t.Fatalf("line %d (case at %d): got T %d %d %x %s, want %v", l.no, caseLine,
					ttype, st.Lineno(), nvalBits, quote(st.F_sval), f[1:])
			}
		case "S":
			if !started {
				st.PushBack()
			}
			if got := quote(st.ToString()); got != f[1] || st.Lineno() != int32(atoi(t, f[2])) {
				t.Fatalf("line %d (case at %d): got S %s %d, want %v", l.no, caseLine, got, st.Lineno(), f[1:])
			}
		case "END":
		default:
			t.Fatalf("line %d: unknown %q", l.no, f[0])
		}
	}
	t.Logf("%d cases, %d tokens", cases, tokens)
}

// sameNval reports whether two nval bit patterns match, treating any NaN as
// equal to any other NaN. StreamTokenizer gets NaN from inf/inf, and the CPU
// sets its sign: positive on arm64, negative on amd64.
func sameNval(got, want uint64) bool {
	if math.IsNaN(math.Float64frombits(got)) && math.IsNaN(math.Float64frombits(want)) {
		return true
	}
	return got == want
}

// TestStreamTokenizerWKTReaderUse checks the calls JTS's WKTReader makes.
func TestStreamTokenizerWKTReaderUse(t *testing.T) {
	st := StreamTokenizer_new_Reader(&StringReader_new(Lit("POINT (1.5 -2e3)")).Reader)
	st.ResetSyntax()
	st.WordChars_int_int('a', 'z')
	st.WordChars_int_int('A', 'Z')
	st.WordChars_int_int(128+32, 255)
	st.WordChars_int_int('0', '9')
	st.WordChars_int_int('-', '-')
	st.WordChars_int_int('+', '+')
	st.WordChars_int_int('.', '.')
	st.WhitespaceChars_int_int(0, ' ')
	st.CommentChar_int('#')
	var got []string
	for st.NextToken() != StreamTokenizer_TT_EOF {
		got = append(got, st.ToString().Go())
	}
	want := []string{
		"Token[POINT], line 1", "Token['('], line 1", "Token[1.5], line 1", "Token[-2e3], line 1",
		"Token[')'], line 1",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Readers

func TestReaderVectors(t *testing.T) {
	var rd *Reader
	var br *BufferedReader
	var lnr *LineNumberReader
	var caseLine int
	cases := 0
	for _, l := range readVectors(t, "io_readers.txt") {
		f := l.fields
		if f[0] == "RCASE" {
			input := unq(f[3])
			size := int32(atoi(t, f[2]))
			br, lnr = nil, nil
			switch f[1] {
			case "BR":
				br = BufferedReader_new_Reader_int(&StringReader_new(input).Reader, size)
				rd = &br.Reader
			case "LNR":
				lnr = LineNumberReader_new_Reader_int(&StringReader_new(input).Reader, size)
				br = &lnr.BufferedReader
				rd = &lnr.Reader
			case "SR":
				rd = &StringReader_new(input).Reader
			case "ISR":
				b := []byte(input.Go())
				if !utf8.Valid(b) {
					t.Fatal("bad input")
				}
				rd = &InputStreamReader_new_InputStream(&ByteArrayInputStream_new_byteArr(bytesToJava(b)).InputStream).Reader
			}
			caseLine = l.no
			cases++
			continue
		}
		if f[0] == "END" {
			continue
		}
		// Split off the line number check.
		wantLine := ""
		if last := f[len(f)-1]; strings.HasPrefix(last, "n=") {
			wantLine = last
			f = f[:len(f)-1]
		}
		var got string
		eq := 0
		for i, x := range f {
			if x == "=" {
				eq = i
				break
			}
		}
		args := f[1:maxInt(eq, 1)]
		want := strings.Join(f[eq+1:], " ")
		switch f[0] {
		case "L":
			got = attemptJava(func() string { return quote(br.ReadLine()) })
		case "r":
			got = attemptJava(func() string { return strconv.Itoa(int(rd.Read())) })
		case "a":
			arrLen, off, n := atoi(t, args[0]), int32(atoi(t, args[1])), int32(atoi(t, args[2]))
			buf := make([]uint16, arrLen)
			got = attemptJava(func() string {
				k := rd.Read_charArr_int_int(buf, off, n)
				return strconv.Itoa(int(k)) + " " + quote(FromUnits(buf))
			})
		case "m":
			got = attemptJava(func() string { rd.Mark_int(int32(atoi(t, args[0]))); return "ok" })
		case "R":
			got = attemptJava(func() string { rd.Reset(); return "ok" })
		case "s":
			got = attemptJava(func() string { return strconv.FormatInt(rd.Skip_long(atoi(t, args[0])), 10) })
		case "y":
			got = attemptJava(func() string { return strconv.FormatBool(rd.Ready()) })
		case "M":
			got = strconv.FormatBool(rd.MarkSupported())
		case "N":
			lnr.SetLineNumber_int(int32(atoi(t, f[1])))
		case "C":
			got = attemptJava(func() string { rd.Close(); return "ok" })
		default:
			t.Fatalf("line %d: unknown op %q", l.no, f[0])
		}
		if f[0] != "N" && !sameResult(got, want) {
			t.Fatalf("line %d (case at %d): %v: got %s, want %s", l.no, caseLine, f[:eq], got, want)
		}
		if wantLine != "" {
			if gotLine := "n=" + strconv.Itoa(int(lnr.GetLineNumber())); gotLine != wantLine {
				t.Fatalf("line %d (case at %d): got %s, want %s", l.no, caseLine, gotLine, wantLine)
			}
		}
	}
	t.Logf("%d cases", cases)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// Random

func TestRandomVectors(t *testing.T) {
	var r *Random
	n := 0
	for _, l := range readVectors(t, "io_random.txt") {
		f := l.fields
		var got, want string
		switch f[0] {
		case "SEED":
			r = Random_new_long(atoi(t, f[1]))
			continue
		case "i":
			got, want = strconv.Itoa(int(r.NextInt())), f[1]
		case "b":
			bound := int32(atoi(t, f[1]))
			got = attemptJava(func() string { return strconv.Itoa(int(r.NextInt_int(bound))) })
			want = strings.Join(f[2:], " ")
		case "l":
			got, want = strconv.FormatInt(r.NextLong(), 10), f[1]
		case "z":
			got, want = strconv.FormatBool(r.NextBoolean()), f[1]
		case "f":
			got, want = strconv.FormatUint(uint64(math.Float32bits(r.NextFloat())), 16), f[1]
		case "d":
			got, want = strconv.FormatUint(math.Float64bits(r.NextDouble()), 16), f[1]
		case "g":
			got, want = strconv.FormatUint(math.Float64bits(r.NextGaussian()), 16), f[1]
		case "y":
			b := make([]int8, atoi(t, f[1]))
			r.NextBytes_byteArr(b)
			got, want = fmtBytes(b), f[2]
		case "s":
			r.SetSeed_long(atoi(t, f[1]))
			continue
		default:
			t.Fatalf("line %d: unknown op %q", l.no, f[0])
		}
		n++
		if !sameResult(got, want) {
			t.Fatalf("line %d: %s: got %s, want %s", l.no, f[0], got, want)
		}
	}
	t.Logf("%d values", n)
}

// ---------------------------------------------------------------------------
// System.arraycopy

func makeTestArray(typ string, n int) any {
	switch typ {
	case "int":
		a := make([]int32, n)
		for i := range a {
			a[i] = int32(i + 1)
		}
		return a
	case "long":
		a := make([]int64, n)
		for i := range a {
			a[i] = int64(i + 1)
		}
		return a
	case "short":
		a := make([]int16, n)
		for i := range a {
			a[i] = int16(i + 1)
		}
		return a
	case "byte":
		a := make([]int8, n)
		for i := range a {
			a[i] = int8(i + 1)
		}
		return a
	case "char":
		a := make([]uint16, n)
		for i := range a {
			a[i] = uint16('a' + i)
		}
		return a
	case "boolean":
		a := make([]bool, n)
		for i := range a {
			a[i] = i%3 == 0
		}
		return a
	case "float":
		a := make([]float32, n)
		for i := range a {
			a[i] = float32(i+1) + 0.5
		}
		return a
	case "double":
		a := make([]float64, n)
		for i := range a {
			a[i] = float64(i+1) + 0.25
		}
		return a
	case "String":
		a := make([]*String, n)
		for i := range a {
			a[i] = FromGo("s" + strconv.Itoa(i+1))
		}
		return a
	case "StringObj":
		a := make([]any, n)
		for i := range a {
			a[i] = FromGo("s" + strconv.Itoa(i+1))
		}
		return a
	case "Object":
		a := make([]any, n)
		for i := range a {
			if i%4 != 3 {
				a[i] = FromGo("o" + strconv.Itoa(i+1))
			}
		}
		return a
	case "mixed":
		a := make([]any, n)
		for i := range a {
			if i%3 == 2 {
				a[i] = int32(i + 1)
			} else {
				a[i] = FromGo("m" + strconv.Itoa(i+1))
			}
		}
		return a
	}
	panic(typ)
}

func arrayText(a any) string {
	var parts []string
	elem := func(e any) string {
		switch x := e.(type) {
		case nil:
			return "null"
		case *String:
			if x == nil {
				return "null"
			}
			return quote(x)
		case float64:
			return strconv.FormatUint(math.Float64bits(x), 16)
		case float32:
			return strconv.FormatUint(uint64(math.Float32bits(x)), 16)
		case uint16:
			return strconv.Itoa(int(x))
		}
		return fmt.Sprint(e)
	}
	switch x := a.(type) {
	case []int32:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []int64:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []int16:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []int8:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []uint16:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []bool:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []float32:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []float64:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []*String:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	case []any:
		for _, e := range x {
			parts = append(parts, elem(e))
		}
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func TestArraycopyVectors(t *testing.T) {
	n := 0
	for _, l := range readVectors(t, "io_arraycopy.txt") {
		f := l.fields
		n++
		switch f[0] {
		case "AC":
			sl, sp := atoi(t, f[2]), int32(atoi(t, f[3]))
			dl, dp, length := atoi(t, f[5]), int32(atoi(t, f[6])), int32(atoi(t, f[7]))
			src := makeTestArray(f[1], int(sl))
			dst := src
			if f[4] != "same" {
				dst = makeTestArray(f[4], int(dl))
			}
			got := attemptJava(func() string { System_arraycopy(src, sp, dst, dp, length); return "ok" })
			got += " " + arrayText(dst)
			if want := strings.Join(f[9:], " "); !sameResult(got, want) {
				t.Fatalf("line %d: %v: got %s, want %s", l.no, f[:8], got, want)
			}
		default:
			var src, dst any = []int32{0}, []int32{0}
			switch f[0] {
			case "NULLSRC":
				src = nil
			case "NULLDST":
				dst = nil
			case "NOTARRAYSRC":
				src = Lit("x")
			case "NOTARRAYDST":
				dst = Lit("x")
			}
			got := attemptJava(func() string { System_arraycopy(src, 0, dst, 0, 0); return "ok" })
			if want := strings.Join(f[2:], " "); !sameResult(got, want) {
				t.Fatalf("line %d: %s: got %s, want %s", l.no, f[0], got, want)
			}
		}
	}
	t.Logf("%d cases", n)
}

// ---------------------------------------------------------------------------
// UTF-8 coding

func TestCharsetVectors(t *testing.T) {
	n := 0
	for _, l := range readVectors(t, "io_charset.txt") {
		f := l.fields
		n++
		switch f[0] {
		case "DEC":
			bos := ByteArrayOutputStream_new()
			bos.Write_byteArr(hexBytes(t, f[1]))
			if got := quote(bos.ToString()); got != f[3] {
				t.Fatalf("line %d: got %s, want %s", l.no, got, f[3])
			}
		case "ISR":
			isr := InputStreamReader_new_InputStream(&ByteArrayInputStream_new_byteArr(hexBytes(t, f[1])).InputStream)
			var u []uint16
			for c := isr.Read(); c >= 0; c = isr.Read() {
				u = append(u, uint16(c))
			}
			if got := quote(FromUnits(u)); got != f[3] {
				t.Fatalf("line %d: got %s, want %s", l.no, got, f[3])
			}
		case "ENC":
			bos := ByteArrayOutputStream_new()
			ps := PrintStream_new_OutputStream(&bos.OutputStream)
			ps.Print_String(unq(f[1]))
			mid := fmtBytes(bos.ToByteArray())
			ps.Print_String(unq(f[2]))
			end := fmtBytes(bos.ToByteArray())
			ps.Close()
			got := mid + " " + end + " " + fmtBytes(bos.ToByteArray())
			if want := strings.Join(f[4:], " "); got != want {
				t.Fatalf("line %d: got %s, want %s", l.no, got, want)
			}
		default:
			t.Fatalf("line %d: unknown %q", l.no, f[0])
		}
	}
	t.Logf("%d cases", n)
}

// ---------------------------------------------------------------------------
// Writers and streams

// logWriter is a Writer subclass implementing only the abstract methods, as
// a generated subclass would.
type logWriter struct {
	Writer
	log []uint16
}

func (w *logWriter) Write_charArr_int_int(cbuf []uint16, off, n int32) {
	// As log.append("w(").append(new String(cbuf, off, n)) in the Java
	// LogWriter, which appends "w(" before the String constructor throws.
	w.log = append(w.log, 'w', '(')
	requireArray(cbuf)
	if off|n < 0 || n > int32(len(cbuf))-off {
		panic(StringIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"Range [%d, %d + %d) out of bounds for length %d", off, off, n, len(cbuf)))))
	}
	w.log = append(append(w.log, cbuf[off:off+n]...), ')')
}
func (w *logWriter) Flush() { w.log = append(w.log, 'f') }
func (w *logWriter) Close() { w.log = append(w.log, 'c') }

// logOutputStream is an OutputStream subclass implementing only write(int).
type logOutputStream struct {
	OutputStream
	log strings.Builder
}

func (s *logOutputStream) Write_int(b int32) { fmt.Fprintf(&s.log, "%d;", b) }

// countingInputStream is an InputStream subclass implementing only read().
type countingInputStream struct {
	InputStream
	n, limit int32
}

func (s *countingInputStream) Read() int32 {
	if s.n == 7 && s.limit < 0 {
		s.n++
		panic(IOException_new_String(Lit("boom")))
	}
	lim := s.limit
	if lim < 0 {
		lim = -lim
	}
	if s.n < lim {
		v := (s.n * 37) & 0xff
		s.n++
		return v
	}
	return -1
}

// arrayReader is a Reader subclass implementing only the abstract methods.
type arrayReader struct {
	Reader
	s   *String
	pos int32
}

func (r *arrayReader) Read_charArr_int_int(cbuf []uint16, off, n int32) int32 {
	if r.pos >= r.s.Length() {
		return -1
	}
	k := lesserInt32(lesserInt32(n, 2), r.s.Length()-r.pos)
	copy(cbuf[off:off+k], r.s.units[r.pos:r.pos+k])
	r.pos += k
	return k
}

func (r *arrayReader) Close() { r.pos = r.s.Length() }

func printTestObject(i int64) any {
	return []any{nil, Lit("str"), int32(42), int64(-7), uint16('c'), true, 1.5, float32(0.1), int16(-3), int8(9)}[i]
}

func TestWriterVectors(t *testing.T) {
	var (
		mode string
		sw   *StringWriter
		lw   *logWriter
		w    *Writer
		los  *logOutputStream
		bos  *ByteArrayOutputStream
		out  *OutputStream
		is   *InputStream
		ar   *arrayReader
	)
	n := 0
	for _, l := range readVectors(t, "io_writers.txt") {
		f := l.fields
		eq := len(f)
		for i, x := range f {
			if x == "=" {
				eq = i
				break
			}
		}
		want := ""
		if eq < len(f) {
			want = strings.Join(f[eq+1:], " ")
		}
		var got string
		switch f[0] {
		case "WCASE":
			mode = "w"
			sw = StringWriter_new()
			lw = &logWriter{}
			lw.JSetSelf(lw)
			lw.Writer_init()
			w = &sw.Writer
			if f[1] == "log" {
				w = &lw.Writer
			}
			continue
		case "OCASE":
			mode = "o"
			los = &logOutputStream{}
			los.JSetSelf(los)
			los.OutputStream_init()
			bos = ByteArrayOutputStream_new()
			out = &bos.OutputStream
			if f[1] == "log" {
				out = &los.OutputStream
			}
			continue
		case "ICASE":
			mode = "i"
			data := hexBytes(t, f[3])
			switch f[1] {
			case "0":
				c := &countingInputStream{limit: int32(atoi(t, f[2]))}
				c.JSetSelf(c)
				c.InputStream_init()
				is = &c.InputStream
			case "1":
				is = &ByteArrayInputStream_new_byteArr(data).InputStream
			default:
				is = &ByteArrayInputStream_new_byteArr_int_int(data, int32(atoi(t, f[4])), int32(atoi(t, f[5]))).InputStream
			}
			continue
		case "BCASE":
			mode = "b"
			ar = &arrayReader{s: unq(f[1])}
			ar.JSetSelf(ar)
			ar.Reader_init()
			continue
		case "END":
			continue
		case "RESULT":
			want = strings.Join(f[1:], " ")
			switch {
			case mode == "w" && w == &lw.Writer:
				got = quote(FromUnits(lw.log))
			case mode == "w":
				got = quote(sw.ToString())
			case out == &los.OutputStream:
				got = quote(FromGo(los.log.String()))
			default:
				got = fmtBytes(bos.ToByteArray()) + " " + strconv.Itoa(int(bos.Size())) + " " + quote(bos.ToString()) +
					" " + quote(bos.ToString_int(1))
			}
		case "PCASE":
			got, want = runPrintCase(t, f)
		default:
			switch mode {
			case "w":
				got = runWriterOp(t, f, w)
			case "o":
				if f[0] == "reset" {
					bos.Reset()
					continue
				}
				got = runOutputStreamOp(t, f, out)
			case "i":
				got = runInputStreamOp(t, f, is)
			case "b":
				got = runBaseReaderOp(t, f, &ar.Reader)
			}
		}
		n++
		if !sameResult(got, want) {
			t.Fatalf("line %d: %v: got %s, want %s", l.no, f[:eq], got, want)
		}
	}
	t.Logf("%d checks", n)
}

func runWriterOp(t *testing.T, f []string, w *Writer) string {
	t.Helper()
	arg := func(i int) int32 { return int32(atoi(t, f[i])) }
	chars := func() []uint16 {
		if s := unq(f[1]); s != nil {
			return append(make([]uint16, 0, len(s.units)), s.units...)
		}
		return nil
	}
	switch f[0] {
	case "wi":
		return attemptJava(func() string { w.Write_int(arg(1)); return "ok" })
	case "ws":
		return attemptJava(func() string { w.Write_String(unq(f[1])); return "ok" })
	case "wsr":
		return attemptJava(func() string { w.Write_String_int_int(unq(f[1]), arg(2), arg(3)); return "ok" })
	case "wc":
		return attemptJava(func() string { w.Write_charArr_int_int(chars(), arg(2), arg(3)); return "ok" })
	case "wca":
		return attemptJava(func() string { w.Write_charArr(chars()); return "ok" })
	case "ac":
		return attemptJava(func() string {
			return strconv.FormatBool(Same(w.Append_CharSequence(Ref(unq(f[1]))), Ref(w)))
		})
	case "acr":
		return attemptJava(func() string {
			return strconv.FormatBool(Same(w.Append_CharSequence_int_int(Ref(unq(f[1])), arg(2), arg(3)), Ref(w)))
		})
	case "ach":
		return attemptJava(func() string { return strconv.FormatBool(Same(w.Append_char(uint16(arg(1))), Ref(w))) })
	case "fl":
		return attemptJava(func() string { w.Flush(); return "ok" })
	}
	t.Fatalf("unknown writer op %v", f)
	return ""
}

func runOutputStreamOp(t *testing.T, f []string, out *OutputStream) string {
	t.Helper()
	arg := func(i int) int32 { return int32(atoi(t, f[i])) }
	switch f[0] {
	case "wi":
		return attemptJava(func() string { out.Write_int(arg(1)); return "ok" })
	case "wa":
		return attemptJava(func() string { out.Write_byteArr(hexBytes(t, f[1])); return "ok" })
	case "war":
		return attemptJava(func() string { out.Write_byteArr_int_int(hexBytes(t, f[1]), arg(2), arg(3)); return "ok" })
	case "fl":
		return attemptJava(func() string { out.Flush(); return "ok" })
	}
	t.Fatalf("unknown output stream op %v", f)
	return ""
}

func runInputStreamOp(t *testing.T, f []string, is *InputStream) string {
	t.Helper()
	arg := func(i int) int32 { return int32(atoi(t, f[i])) }
	switch f[0] {
	case "r":
		return attemptJava(func() string { return strconv.Itoa(int(is.Read())) })
	case "ra":
		return attemptJava(func() string {
			buf := make([]int8, arg(1))
			k := is.Read_byteArr(buf)
			return strconv.Itoa(int(k)) + " " + fmtBytes(buf)
		})
	case "rar":
		return attemptJava(func() string {
			buf := make([]int8, arg(1))
			k := is.Read_byteArr_int_int(buf, arg(2), arg(3))
			return strconv.Itoa(int(k)) + " " + fmtBytes(buf)
		})
	case "s":
		return attemptJava(func() string { return strconv.FormatInt(is.Skip_long(int64(arg(1))), 10) })
	case "av":
		return attemptJava(func() string { return strconv.Itoa(int(is.Available())) })
	case "m":
		return attemptJava(func() string { is.Mark_int(5); return "ok " + strconv.FormatBool(is.MarkSupported()) })
	case "R":
		return attemptJava(func() string { is.Reset(); return "ok" })
	}
	t.Fatalf("unknown input stream op %v", f)
	return ""
}

func runBaseReaderOp(t *testing.T, f []string, rd *Reader) string {
	t.Helper()
	switch f[0] {
	case "r":
		return attemptJava(func() string { return strconv.Itoa(int(rd.Read())) })
	case "s":
		return attemptJava(func() string { return strconv.FormatInt(rd.Skip_long(atoi(t, f[1])), 10) })
	case "m":
		return attemptJava(func() string { rd.Mark_int(3); return "ok" })
	case "R":
		return attemptJava(func() string { rd.Reset(); return "ok" })
	case "y":
		return attemptJava(func() string {
			return strconv.FormatBool(rd.Ready()) + " " + strconv.FormatBool(rd.MarkSupported())
		})
	case "ra":
		return attemptJava(func() string {
			buf := make([]uint16, 3)
			k := rd.Read_charArr(buf)
			return strconv.Itoa(int(k)) + " " + quote(FromUnits(buf))
		})
	}
	t.Fatalf("unknown reader op %v", f)
	return ""
}

func runPrintCase(t *testing.T, f []string) (string, string) {
	t.Helper()
	bos := ByteArrayOutputStream_new()
	ps := PrintStream_new_OutputStream_boolean(&bos.OutputStream, f[1] == "true")
	i := 2
	for f[i] != "=" {
		op := f[i]
		i++
		next := func() string { i++; return f[i-1] }
		switch op {
		case "ps":
			ps.Print_String(unq(next()))
		case "pls":
			ps.Println_String(unq(next()))
		case "po":
			ps.Print_Object(printTestObject(atoi(t, next())))
		case "plo":
			ps.Println_Object(printTestObject(atoi(t, next())))
		case "pl":
			ps.Println()
		case "pd":
			ps.Print_double(math.Float64frombits(hexBits(t, next())))
		case "pld":
			ps.Println_double(math.Float64frombits(hexBits(t, next())))
		case "pli":
			ps.Println_int(int32(atoi(t, next())))
		case "pj":
			ps.Print_long(atoi(t, next()))
		case "pf":
			ps.Print_float(math.Float32frombits(uint32(hexBits(t, next()))))
		case "pz":
			ps.Print_boolean(next() == "true")
		case "pc":
			ps.Print_char(uint16(atoi(t, next())))
		case "pca":
			s := unq(next())
			ps.Print_charArr(append(make([]uint16, 0, len(s.units)), s.units...))
		case "ap":
			ps.Append_CharSequence(Ref(unq(next())))
		case "wi":
			ps.Write_int(int32(atoi(t, next())))
		case "close":
			ps.Close()
		case "flush":
			ps.Flush()
		default:
			t.Fatalf("unknown print op %q", op)
		}
	}
	got := fmtBytes(bos.ToByteArray()) + " " + strconv.FormatBool(ps.CheckError())
	return got, strings.Join(f[i+1:], " ")
}

// ---------------------------------------------------------------------------
// File paths, properties, Date, and the file readers' use of readers

func fileText(f *File) string {
	return quote(f.GetPath()) + " " + quote(f.GetName()) + " " + quote(f.GetParent()) + " " +
		strconv.FormatBool(f.IsAbsolute()) + " " + strconv.Itoa(int(f.HashCode())) + " " + quote(f.ToString())
}

func TestMiscVectors(t *testing.T) {
	n := 0
	for _, l := range readVectors(t, "io_misc.txt") {
		f := l.fields
		var got, want string
		switch f[0] {
		case "FILE":
			got, want = fileText(File_new_String(unq(f[1]))), strings.Join(f[3:], " ")
		case "FILE2":
			got, want = fileText(File_new_String_String(unq(f[1]), unq(f[2]))), strings.Join(f[4:], " ")
		case "FILE3":
			var parent *File
			if p := unq(f[1]); p != nil {
				parent = File_new_String(p)
			}
			got, want = fileText(File_new_File_String(parent, unq(f[2]))), strings.Join(f[4:], " ")
		case "CMP":
			got = strconv.Itoa(int(File_new_String(Lit("a")).CompareTo_Object(File_new_String(Lit("b"))))) + " " +
				strconv.FormatBool(File_new_String(Lit("a/")).Equals_Object(File_new_String(Lit("a")))) + " " +
				strconv.Itoa(int(File_new_String(Lit("b")).CompareTo_Object(File_new_String(Lit("ab")))))
			want = strings.Join(f[1:], " ")
		case "PROP":
			key := unq(f[1])
			got = attemptJava(func() string { return quote(System_getProperty_String(key)) }) + " " +
				attemptJava(func() string { return quote(System_getProperty_String_String(key, Lit("def"))) })
			want = strings.Join(f[3:], " ")
		case "LSEP":
			got, want = quote(System_lineSeparator()), f[1]
		case "DATE":
			ms := atoi(t, f[1])
			d := Date_new_long(ms)
			got = fmt.Sprintf("%d %d %d %t %t %t", d.HashCode(), d.GetTime(), d.CompareTo_Object(Date_new_long(0)),
				d.Equals_Object(Date_new_long(ms)), d.Before_Date(Date_new_long(1)), d.After_Date(Date_new_long(1)))
			want = strings.Join(f[3:], " ")
		case "HEXFILE":
			br := BufferedReader_new_Reader(&StringReader_new(unq(f[1])).Reader)
			var res []string
			for i := 0; i < 10; i++ {
				r := attemptJava(func() string {
					br.Mark_int(1000)
					st := StreamTokenizer_new_Reader(&br.Reader)
					if st.NextToken() == StreamTokenizer_TT_EOF {
						return "EOF"
					}
					br.Reset()
					return quote(br.ReadLine())
				})
				res = append(res, r)
				if !strings.HasPrefix(r, "\"") {
					break
				}
			}
			got, want = strings.Join(res, " "), strings.Join(f[3:], " ")
		case "WKTFILE":
			br := BufferedReader_new_Reader_int(&StringReader_new(unq(f[1])).Reader, 7)
			var res []string
			for i := 0; i < 10; i++ {
				r := attemptJava(func() string {
					for {
						br.Mark_int(1)
						ch := br.Read()
						if ch < 0 {
							return "EOF"
						}
						if !Character_isWhitespace_int(ch) {
							break
						}
					}
					br.Reset()
					st := StreamTokenizer_new_Reader(&br.Reader)
					tt := st.NextToken()
					return strconv.Itoa(int(tt)) + ":" + quote(st.F_sval)
				})
				res = append(res, r)
				if r == "EOF" || strings.HasPrefix(r, "!") {
					break
				}
			}
			got, want = strings.Join(res, " "), strings.Join(f[3:], " ")
		default:
			t.Fatalf("line %d: unknown %q", l.no, f[0])
		}
		n++
		if !sameResult(got, want) {
			t.Fatalf("line %d: %s: got %s, want %s", l.no, f[0], got, want)
		}
	}
	t.Logf("%d checks", n)
}

// ---------------------------------------------------------------------------
// Files, System.out, threads and other environment-dependent behaviour.
// The expected messages were checked against JDK 21.

func TestFileStreams(t *testing.T) {
	dir := t.TempDir()
	path := FromGo(dir + "//data.txt")
	ps := PrintStream_new_String(path)
	ps.Println_String(Lit("POINT (1 2)"))
	ps.Print_String(Lit("é\r\nlast"))
	ps.Close()
	if ps.CheckError() {
		t.Fatal("unexpected error")
	}

	f := File_new_String(FromGo(dir + "/data.txt"))
	if !f.Exists() || !f.IsFile() || f.IsDirectory() || f.Length() != 20 {
		t.Fatalf("exists=%t isFile=%t isDir=%t length=%d", f.Exists(), f.IsFile(), f.IsDirectory(), f.Length())
	}
	br := BufferedReader_new_Reader(&FileReader_new_File(f).Reader)
	var lines []string
	for l := br.ReadLine(); l != nil; l = br.ReadLine() {
		lines = append(lines, l.Go())
	}
	br.Close()
	if got := strings.Join(lines, "|"); got != "POINT (1 2)|é|last" {
		t.Fatalf("got %q", got)
	}

	in := FileInputStream_new_File(f)
	if in.Available() != 20 || in.Read() != 'P' || in.Skip_long(6) != 6 || in.Read() != '1' {
		t.Fatal("unexpected FileInputStream state")
	}
	buf := make([]int8, 100)
	if n := in.Read_byteArr(buf); n != 12 {
		t.Fatalf("read %d bytes", n)
	}
	if in.Read() != -1 || in.Read_byteArr(buf) != -1 {
		t.Fatal("expected end of file")
	}
	in.Close()
	if got := attemptJava(func() string { return strconv.Itoa(int(in.Read())) }); got != `!java.io.IOException "Stream Closed"` {
		t.Fatalf("read after close: %s", got)
	}

	out := FileOutputStream_new_File_boolean(f, true)
	out.Write_byteArr(bytesToJava([]byte("!")))
	out.Close()
	if got := attemptJava(func() string { out.Write_int(1); return "ok" }); got != `!java.io.IOException "Stream Closed"` {
		t.Fatalf("write after close: %s", got)
	}
	out.Write_byteArr([]int8{})
	if f.Length() != 21 {
		t.Fatalf("length after append %d", f.Length())
	}

	d := File_new_File_String(File_new_String(FromGo(dir)), Lit("d"))
	if !d.Mkdir() || d.Mkdir() || !d.IsDirectory() {
		t.Fatal("mkdir")
	}
	names := File_new_String(FromGo(dir)).List()
	if len(names) != 2 || names[0].Go() != "d" || names[1].Go() != "data.txt" {
		t.Fatalf("list %v", names)
	}
	files := File_new_String(FromGo(dir)).ListFiles()
	if files[1].GetPath().Go() != dir+"/data.txt" || !files[0].Delete() || d.Exists() {
		t.Fatal("listFiles/delete")
	}
	if File_new_String(FromGo(dir+"/data.txt")).List() != nil {
		t.Fatal("list of a file")
	}

	for _, c := range []struct {
		open func()
		want string
	}{
		{
			func() { FileInputStream_new_String(Lit("/nonexistent/x")) },
			`!java.io.FileNotFoundException "/nonexistent/x (No such file or directory)"`,
		},
		{
			func() { FileInputStream_new_String(FromGo(dir)) },
			`!java.io.FileNotFoundException "` + dir + ` (Is a directory)"`,
		},
		{
			func() { FileOutputStream_new_String(FromGo(dir)) },
			`!java.io.FileNotFoundException "` + dir + ` (Is a directory)"`,
		},
		{
			func() { FileReader_new_String(Lit("/nonexistent")) },
			`!java.io.FileNotFoundException "/nonexistent (No such file or directory)"`,
		},
		{
			func() { FileInputStream_new_String(FromUnits([]uint16{'a', 0, 'b'})) },
			`!java.io.FileNotFoundException "Invalid file path"`,
		},
		{func() { FileInputStream_new_String(nil) }, `!java.lang.NullPointerException null`},
	} {
		if got := attemptJava(func() string { c.open(); return "ok" }); got != c.want {
			t.Errorf("got %s, want %s", got, c.want)
		}
	}
	if !FileNotFoundException_isInstance(FileNotFoundException_new()) ||
		!IOException_isInstance(FileNotFoundException_new_String(Lit("x"))) {
		t.Fatal("FileNotFoundException hierarchy")
	}
}

func TestFileAbsolutePath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if got := File_new_String(Lit("x")).GetAbsolutePath().Go(); got != wd+"/x" {
		t.Fatalf("got %q", got)
	}
	if got := File_new_String(Lit("/a//b/")).GetAbsolutePath().Go(); got != "/a/b" {
		t.Fatalf("got %q", got)
	}
	if got := System_getProperty_String(Lit("user.dir")).Go(); got != wd {
		t.Fatalf("user.dir %q", got)
	}
}

type lockedBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

func TestSystemOutRedirect(t *testing.T) {
	var out, errOut lockedBuffer
	prevOut := SetStdoutWriter(&out)
	prevErr := SetStderrWriter(&errOut)
	defer SetStdoutWriter(prevOut)
	defer SetStderrWriter(prevErr)

	System_out.Println_String(Lit("hello"))
	System_out.Print_int(-3)
	System_out.Print_char(' ')
	System_out.Println_double(0.1)
	System_out.Println_Object(nil)
	System_out.Print_String(FromUnits([]uint16{0xd83d}))
	System_out.Println_String(FromUnits([]uint16{0xde00, '!'}))
	System_err.Println_String(Lit("err"))
	if got := out.String(); got != "hello\n-3 0.1\nnull\n\U0001F600!\n" {
		t.Fatalf("got %q", got)
	}
	if got := errOut.String(); got != "err\n" {
		t.Fatalf("got %q", got)
	}
}

type testThread struct {
	Thread
	ran *bool
}

func (th *testThread) Run() { *th.ran = true }

func TestThread(t *testing.T) {
	ran := false
	th := Thread_new_Runnable(RunnableFunc(func() { ran = true }))
	if !strings.HasPrefix(th.GetName().Go(), "Thread-") {
		t.Fatalf("name %q", th.GetName().Go())
	}
	th.Start()
	th.Join()
	if !ran {
		t.Fatal("did not run")
	}
	if got := attemptJava(func() string { th.Start(); return "ok" }); got != `!java.lang.IllegalThreadStateException null` {
		t.Fatalf("second start: %s", got)
	}
	if !strings.HasSuffix(th.ToString().Go(), ","+th.GetName().Go()+",5,]") {
		t.Fatalf("toString %q", th.ToString().Go())
	}

	// A subclass overriding run, as a generated class would.
	sub := &testThread{ran: new(bool)}
	sub.JSetSelf(sub)
	sub.Thread_init()
	sub.Thread.Start()
	sub.Join()
	if !*sub.ran {
		t.Fatal("subclass run not called")
	}

	var errOut lockedBuffer
	prev := SetStderrWriter(&errOut)
	defer SetStderrWriter(prev)
	bad := Thread_new_Runnable_String(RunnableFunc(func() {
		panic(IllegalStateException_new_String(Lit("bang")))
	}), Lit("worker"))
	bad.Start()
	bad.Join()
	if got := errOut.String(); got != "Exception in thread \"worker\" java.lang.IllegalStateException: bang\n" {
		t.Fatalf("got %q", got)
	}
	Thread_sleep_long(1)
}

func TestPrintStackTrace(t *testing.T) {
	bos := ByteArrayOutputStream_new()
	ps := PrintStream_new_OutputStream(&bos.OutputStream)
	ex := RuntimeException_new_String_Throwable(Lit("m"), &IllegalStateException_new_String(Lit("c")).Throwable)
	ex.PrintStackTrace_PrintStream(ps)
	if got := bos.ToString().Go(); got != "java.lang.RuntimeException: m\nCaused by: java.lang.IllegalStateException: c\n" {
		t.Fatalf("got %q", got)
	}
}

func TestMiscRuntime(t *testing.T) {
	if a, b := Random_new().NextLong(), Random_new().NextLong(); a == b {
		t.Fatal("unseeded Randoms repeat")
	}
	rt := Runtime_getRuntime()
	if rt != Runtime_getRuntime() || rt.TotalMemory() <= 0 || rt.FreeMemory() < 0 || rt.MaxMemory() <= 0 ||
		rt.AvailableProcessors() < 1 {
		t.Fatal("Runtime")
	}
	rt.Gc()
	System_gc()
	if a, b := System_nanoTime(), System_nanoTime(); b < a {
		t.Fatal("nanoTime went backwards")
	}
	if d := Date_new(); d.GetTime() < 1700000000000 {
		t.Fatalf("Date %d", d.GetTime())
	}
	ref := SoftReference_new_Object(Lit("x"))
	if ref.Get() != any(Lit("x")) {
		t.Fatal("SoftReference")
	}
	if !Supplier_isInstance(SupplierFunc(func() any { return nil })) || Supplier_isInstance(RunnableFunc(func() {})) {
		t.Fatal("functional interface markers")
	}
	if FunctionFunc(func(a any) any { return a }).Apply_Object(int32(3)) != any(int32(3)) {
		t.Fatal("FunctionFunc")
	}
	var got any
	ConsumerFunc(func(a any) { got = a }).Accept_Object(int32(4))
	if got != any(int32(4)) {
		t.Fatal("ConsumerFunc")
	}
	if h1, h2 := System_identityHashCode(ref), System_identityHashCode(ref); h1 != h2 || System_identityHashCode(nil) != 0 {
		t.Fatal("identityHashCode")
	}
	if !CloneNotSupportedException_isInstance(CloneNotSupportedException_new()) ||
		!Exception_isInstance(CloneNotSupportedException_new_String(Lit("x"))) {
		t.Fatal("CloneNotSupportedException")
	}
	func() {
		defer func() {
			if _, ok := recover().(UnsupportedError); !ok {
				t.Fatal("System.exit should be unsupported")
			}
		}()
		System_exit(0)
	}()
}
