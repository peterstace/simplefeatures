package java

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestMathVectors replays the vectors that
// internal/jts/tools/jdkvectors/MathVectors.java generated from the JVM
// against both the Math_ and StrictMath_ implementation of each method,
// requiring identical bits. NaN results only need to be NaN, as Java does not
// specify NaN payloads.
func TestMathVectors(t *testing.T) {
	cases := mathVectorCases()
	f, err := os.Open("testdata/math.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	seen := map[string]int{}
	var failures int
	var c mathVectorCase
	var section string
	sc := bufio.NewScanner(f)
	for lineNo := 1; sc.Scan(); lineNo++ {
		line := sc.Text()
		if line == "" || line[0] == '#' {
			continue
		}
		if line[0] == '[' {
			section = strings.Trim(line, "[]")
			var ok bool
			if c, ok = cases[section]; !ok {
				t.Fatalf("line %d: unknown function %q", lineNo, section)
			}
			continue
		}
		fields := strings.Fields(line)
		if section == "" || len(fields) != len(c.args)+1 {
			t.Fatalf("line %d: want %d fields in section %q: %q", lineNo, len(c.args)+1, section, line)
		}
		seen[section]++
		args := make([]uint64, len(c.args))
		for i, kind := range []byte(c.args) {
			args[i] = parseMathVectorValue(t, lineNo, kind, fields[i])
		}
		wantStr := fields[len(fields)-1]
		wantExc := wantStr == "ArithmeticException"
		var want uint64
		if !wantExc {
			want = parseMathVectorValue(t, lineNo, c.ret, wantStr)
			wantStr = formatMathVectorValue(c.ret, want)
		}
		for _, impl := range c.impls {
			got, gotExc := callMathImpl(impl.call, args)
			if wantExc && gotExc == "ArithmeticException" ||
				!wantExc && gotExc == "" && sameMathVectorValue(c.ret, got, want) {
				continue
			}
			failures++
			if failures <= 100 {
				gotStr := gotExc
				if gotStr == "" {
					gotStr = formatMathVectorValue(c.ret, got)
				}
				t.Errorf("line %d: %s(%s) = %s, want %s", lineNo, impl.name,
					formatMathVectorArgs(c.args, args), gotStr, wantStr)
			}
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if failures > 100 {
		t.Errorf("%d failures in total", failures)
	}
	for name := range cases {
		if seen[name] == 0 {
			t.Errorf("no vectors for %s", name)
		}
	}
}

func TestMathRandom(t *testing.T) {
	distinct := map[float64]bool{}
	for i := 0; i < 1000; i++ {
		for _, v := range []float64{Math_random(), StrictMath_random()} {
			if !(v >= 0 && v < 1) {
				t.Fatalf("random value %v outside [0, 1)", v)
			}
			if scaled := v * 0x1p53; math.Trunc(scaled) != scaled {
				t.Fatalf("random value %v is not a multiple of 2^-53", v)
			}
			distinct[v] = true
		}
	}
	if len(distinct) < 1990 {
		t.Fatalf("only %d distinct values in 2000 calls", len(distinct))
	}
}

// mathImpl is one Go function under test, adapted to take and return values
// encoded as in parseMathVectorValue.
type mathImpl struct {
	name string
	call func([]uint64) uint64
}

// mathVectorCase describes one section of the vector file: the
// kinds of the arguments and result (d double, f float, i int, l long), and
// the Go functions that must produce the expected result.
type mathVectorCase struct {
	args  string
	ret   byte
	impls []mathImpl
}

func mathVectorCases() map[string]mathVectorCase {
	return map[string]mathVectorCase{
		"sin":           d1Case(Math_sin, StrictMath_sin),
		"cos":           d1Case(Math_cos, StrictMath_cos),
		"tan":           d1Case(Math_tan, StrictMath_tan),
		"asin":          d1Case(Math_asin, StrictMath_asin),
		"acos":          d1Case(Math_acos, StrictMath_acos),
		"atan":          d1Case(Math_atan, StrictMath_atan),
		"atan2":         d2Case(Math_atan2, StrictMath_atan2),
		"exp":           d1Case(Math_exp, StrictMath_exp),
		"expm1":         d1Case(Math_expm1, StrictMath_expm1),
		"log":           d1Case(Math_log, StrictMath_log),
		"log10":         d1Case(Math_log10, StrictMath_log10),
		"log1p":         d1Case(Math_log1p, StrictMath_log1p),
		"pow":           d2Case(Math_pow, StrictMath_pow),
		"hypot":         d2Case(Math_hypot, StrictMath_hypot),
		"IEEEremainder": d2Case(Math_IEEEremainder, StrictMath_IEEEremainder),
		"cbrt":          d1Case(Math_cbrt, StrictMath_cbrt),
		"sqrt":          d1Case(Math_sqrt, StrictMath_sqrt),
		"sinh":          d1Case(Math_sinh, StrictMath_sinh),
		"cosh":          d1Case(Math_cosh, StrictMath_cosh),
		"tanh":          d1Case(Math_tanh, StrictMath_tanh),

		"floor":        d1Case(Math_floor, StrictMath_floor),
		"ceil":         d1Case(Math_ceil, StrictMath_ceil),
		"rint":         d1Case(Math_rint, StrictMath_rint),
		"round_double": dToLCase(Math_round_double, StrictMath_round_double),
		"round_float":  fToICase(Math_round_float, StrictMath_round_float),

		"toDegrees":                   d1Case(Math_toDegrees, StrictMath_toDegrees),
		"toRadians":                   d1Case(Math_toRadians, StrictMath_toRadians),
		"abs_double":                  d1Case(Math_abs_double, StrictMath_abs_double),
		"signum_double":               d1Case(Math_signum_double, StrictMath_signum_double),
		"ulp_double":                  d1Case(Math_ulp_double, StrictMath_ulp_double),
		"nextUp_double":               d1Case(Math_nextUp_double, StrictMath_nextUp_double),
		"nextDown_double":             d1Case(Math_nextDown_double, StrictMath_nextDown_double),
		"getExponent_double":          dToICase(Math_getExponent_double, StrictMath_getExponent_double),
		"max_double_double":           d2Case(Math_max_double_double, StrictMath_max_double_double),
		"min_double_double":           d2Case(Math_min_double_double, StrictMath_min_double_double),
		"copySign_double_double":      d2Case(StrictMath_copySign_double_double),
		"copySign_double_double@Math": d2Case(Math_copySign_double_double),
		"nextAfter_double_double":     d2Case(Math_nextAfter_double_double, StrictMath_nextAfter_double_double),
		"scalb_double_int":            diCase(Math_scalb_double_int, StrictMath_scalb_double_int),
		"abs_float":                   f1Case(Math_abs_float, StrictMath_abs_float),
		"signum_float":                f1Case(Math_signum_float, StrictMath_signum_float),
		"ulp_float":                   f1Case(Math_ulp_float, StrictMath_ulp_float),
		"nextUp_float":                f1Case(Math_nextUp_float, StrictMath_nextUp_float),
		"nextDown_float":              f1Case(Math_nextDown_float, StrictMath_nextDown_float),
		"getExponent_float":           fToICase(Math_getExponent_float, StrictMath_getExponent_float),
		"max_float_float":             f2Case(Math_max_float_float, StrictMath_max_float_float),
		"min_float_float":             f2Case(Math_min_float_float, StrictMath_min_float_float),
		"copySign_float_float":        f2Case(StrictMath_copySign_float_float),
		"copySign_float_float@Math":   f2Case(Math_copySign_float_float),
		"nextAfter_float_double":      fdCase(Math_nextAfter_float_double, StrictMath_nextAfter_float_double),
		"scalb_float_int":             fiCase(Math_scalb_float_int, StrictMath_scalb_float_int),

		"addExact_int_int":      i2Case(Math_addExact_int_int, StrictMath_addExact_int_int),
		"subtractExact_int_int": i2Case(Math_subtractExact_int_int, StrictMath_subtractExact_int_int),
		"multiplyExact_int_int": i2Case(Math_multiplyExact_int_int, StrictMath_multiplyExact_int_int),
		"floorDiv_int_int":      i2Case(Math_floorDiv_int_int, StrictMath_floorDiv_int_int),
		"floorMod_int_int":      i2Case(Math_floorMod_int_int, StrictMath_floorMod_int_int),
		"max_int_int":           i2Case(Math_max_int_int, StrictMath_max_int_int),
		"min_int_int":           i2Case(Math_min_int_int, StrictMath_min_int_int),
		"abs_int":               i1Case(Math_abs_int, StrictMath_abs_int),
		"incrementExact_int":    i1Case(Math_incrementExact_int),
		"decrementExact_int":    i1Case(Math_decrementExact_int),
		"negateExact_int":       i1Case(Math_negateExact_int),

		"addExact_long_long":      l2Case(Math_addExact_long_long, StrictMath_addExact_long_long),
		"subtractExact_long_long": l2Case(Math_subtractExact_long_long, StrictMath_subtractExact_long_long),
		"multiplyExact_long_long": l2Case(Math_multiplyExact_long_long, StrictMath_multiplyExact_long_long),
		"floorDiv_long_long":      l2Case(Math_floorDiv_long_long, StrictMath_floorDiv_long_long),
		"floorMod_long_long":      l2Case(Math_floorMod_long_long, StrictMath_floorMod_long_long),
		"max_long_long":           l2Case(Math_max_long_long, StrictMath_max_long_long),
		"min_long_long":           l2Case(Math_min_long_long, StrictMath_min_long_long),
		"abs_long":                l1Case(Math_abs_long, StrictMath_abs_long),
		"incrementExact_long":     l1Case(Math_incrementExact_long),
		"decrementExact_long":     l1Case(Math_decrementExact_long),
		"negateExact_long":        l1Case(Math_negateExact_long),
		"toIntExact":              lToICase(Math_toIntExact, StrictMath_toIntExact),
	}
}

func d1Case(fns ...func(float64) float64) mathVectorCase {
	c := mathVectorCase{args: "d", ret: 'd'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return math.Float64bits(fn(math.Float64frombits(a[0])))
		}})
	}
	return c
}

func d2Case(fns ...func(float64, float64) float64) mathVectorCase {
	c := mathVectorCase{args: "dd", ret: 'd'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return math.Float64bits(fn(math.Float64frombits(a[0]), math.Float64frombits(a[1])))
		}})
	}
	return c
}

func diCase(fns ...func(float64, int32) float64) mathVectorCase {
	c := mathVectorCase{args: "di", ret: 'd'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return math.Float64bits(fn(math.Float64frombits(a[0]), int32(a[1])))
		}})
	}
	return c
}

func dToLCase(fns ...func(float64) int64) mathVectorCase {
	c := mathVectorCase{args: "d", ret: 'l'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(fn(math.Float64frombits(a[0])))
		}})
	}
	return c
}

func dToICase(fns ...func(float64) int32) mathVectorCase {
	c := mathVectorCase{args: "d", ret: 'i'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(int64(fn(math.Float64frombits(a[0]))))
		}})
	}
	return c
}

func f1Case(fns ...func(float32) float32) mathVectorCase {
	c := mathVectorCase{args: "f", ret: 'f'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(math.Float32bits(fn(math.Float32frombits(uint32(a[0])))))
		}})
	}
	return c
}

func f2Case(fns ...func(float32, float32) float32) mathVectorCase {
	c := mathVectorCase{args: "ff", ret: 'f'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(math.Float32bits(fn(math.Float32frombits(uint32(a[0])), math.Float32frombits(uint32(a[1])))))
		}})
	}
	return c
}

func fdCase(fns ...func(float32, float64) float32) mathVectorCase {
	c := mathVectorCase{args: "fd", ret: 'f'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(math.Float32bits(fn(math.Float32frombits(uint32(a[0])), math.Float64frombits(a[1]))))
		}})
	}
	return c
}

func fiCase(fns ...func(float32, int32) float32) mathVectorCase {
	c := mathVectorCase{args: "fi", ret: 'f'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(math.Float32bits(fn(math.Float32frombits(uint32(a[0])), int32(a[1]))))
		}})
	}
	return c
}

func fToICase(fns ...func(float32) int32) mathVectorCase {
	c := mathVectorCase{args: "f", ret: 'i'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(int64(fn(math.Float32frombits(uint32(a[0])))))
		}})
	}
	return c
}

func i1Case(fns ...func(int32) int32) mathVectorCase {
	c := mathVectorCase{args: "i", ret: 'i'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(int64(fn(int32(a[0]))))
		}})
	}
	return c
}

func i2Case(fns ...func(int32, int32) int32) mathVectorCase {
	c := mathVectorCase{args: "ii", ret: 'i'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(int64(fn(int32(a[0]), int32(a[1]))))
		}})
	}
	return c
}

func l1Case(fns ...func(int64) int64) mathVectorCase {
	c := mathVectorCase{args: "l", ret: 'l'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(fn(int64(a[0])))
		}})
	}
	return c
}

func l2Case(fns ...func(int64, int64) int64) mathVectorCase {
	c := mathVectorCase{args: "ll", ret: 'l'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(fn(int64(a[0]), int64(a[1])))
		}})
	}
	return c
}

func lToICase(fns ...func(int64) int32) mathVectorCase {
	c := mathVectorCase{args: "l", ret: 'i'}
	for _, fn := range fns {
		fn := fn
		c.impls = append(c.impls, mathImpl{goFuncName(fn), func(a []uint64) uint64 {
			return uint64(int64(fn(int64(a[0]))))
		}})
	}
	return c
}

func goFuncName(fn any) string {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return name[strings.LastIndexByte(name, '.')+1:]
}

// callMathImpl calls fn, reporting the simple class name of a Java exception
// it throws.
func callMathImpl(fn func([]uint64) uint64, args []uint64) (got uint64, exc string) {
	defer func() {
		if r := recover(); r != nil {
			ex := Throwable_ofPanic(r)
			name := className(ex)
			exc = name[strings.LastIndexByte(name, '.')+1:]
		}
	}()
	return fn(args), ""
}

// parseMathVectorValue decodes a value from the vector file: doubles and
// floats as hex bits, ints and longs as sign-extended decimal.
func parseMathVectorValue(t *testing.T, lineNo int, kind byte, s string) uint64 {
	t.Helper()
	var v uint64
	var err error
	switch kind {
	case 'd':
		v, err = strconv.ParseUint(s, 16, 64)
	case 'f':
		v, err = strconv.ParseUint(s, 16, 32)
	case 'i':
		var x int64
		x, err = strconv.ParseInt(s, 10, 32)
		v = uint64(x)
	case 'l':
		var x int64
		x, err = strconv.ParseInt(s, 10, 64)
		v = uint64(x)
	default:
		t.Fatalf("line %d: bad kind %q", lineNo, kind)
	}
	if err != nil {
		t.Fatalf("line %d: %v", lineNo, err)
	}
	return v
}

func sameMathVectorValue(kind byte, got, want uint64) bool {
	switch kind {
	case 'd':
		g, w := math.Float64frombits(got), math.Float64frombits(want)
		if w != w {
			return g != g
		}
	case 'f':
		g, w := math.Float32frombits(uint32(got)), math.Float32frombits(uint32(want))
		if w != w {
			return g != g
		}
	}
	return got == want
}

func formatMathVectorValue(kind byte, v uint64) string {
	switch kind {
	case 'd':
		return fmt.Sprintf("%016x (%v)", v, math.Float64frombits(v))
	case 'f':
		return fmt.Sprintf("%08x (%v)", v, math.Float32frombits(uint32(v)))
	}
	return strconv.FormatInt(int64(v), 10)
}

func formatMathVectorArgs(kinds string, args []uint64) string {
	strs := make([]string, len(args))
	for i, kind := range []byte(kinds) {
		strs[i] = formatMathVectorValue(kind, args[i])
	}
	return strings.Join(strs, ", ")
}
