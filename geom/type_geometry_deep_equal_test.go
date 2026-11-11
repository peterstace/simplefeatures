package geom_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

// TestReflectDeepEqualWorks proves that reflect.DeepEqual works for comparing
// geom.Geometry values.
func TestReflectDeepEqualWorks(t *testing.T) {
	tests := []struct {
		name string
		wkt  string
	}{
		{"Point", "POINT(1 2)"},
		{"LineString", "LINESTRING(1 2,3 4)"},
		{"Polygon", "POLYGON((0 0,0 1,1 0,0 0))"},
		{"MultiPoint", "MULTIPOINT((1 2),(3 4))"},
		{"MultiLineString", "MULTILINESTRING((1 2,3 4),(5 6,7 8))"},
		{"MultiPolygon", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))"},
		{"GeometryCollection", "GEOMETRYCOLLECTION(POINT(1 2))"},
		{"EmptyPoint", "POINT EMPTY"},
		{"EmptyLineString", "LINESTRING EMPTY"},
		{"EmptyPolygon", "POLYGON EMPTY"},
		{"EmptyMultiPoint", "MULTIPOINT EMPTY"},
		{"EmptyMultiLineString", "MULTILINESTRING EMPTY"},
		{"EmptyMultiPolygon", "MULTIPOLYGON EMPTY"},
		{"EmptyGeometryCollection", "GEOMETRYCOLLECTION EMPTY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g1 := test.FromWKT(t, tt.wkt)
			g2 := test.FromWKT(t, tt.wkt)
			test.ExactEquals(t, g1, g2) // The geometries are semantically equal.
			test.DeepEqual(t, g1, g2)   // And reflect.DeepEqual works as expected.
		})
	}
}

// TestReflectDeepEqualWorksForZeroValue proves that reflect.DeepEqual handles
// the special zero value semantics of geom.Geometry. The zero value of
// Geometry is an empty GeometryCollection with a nil impl, and this is
// semantically equal to manually constructed empty GeometryCollections.
func TestReflectDeepEqualWorksForZeroValue(t *testing.T) {
	// Zero value has impl == nil.
	var g1 geom.Geometry

	// Constructed from WKT.
	g2 := test.FromWKT(t, "GEOMETRYCOLLECTION EMPTY")

	// Manually constructed with nil slice.
	g3 := geom.NewGeometryCollection(nil).AsGeometry()

	// Manually constructed with empty slice.
	g4 := geom.NewGeometryCollection([]geom.Geometry{}).AsGeometry()

	// All four geometries are semantically equal.
	test.ExactEquals(t, g1, g2)
	test.ExactEquals(t, g1, g3)
	test.ExactEquals(t, g1, g4)
	test.ExactEquals(t, g2, g3)
	test.ExactEquals(t, g2, g4)
	test.ExactEquals(t, g3, g4)

	// And reflect.DeepEqual works for all pairs.
	test.DeepEqual(t, g1, g2)
	test.DeepEqual(t, g1, g3)
	test.DeepEqual(t, g1, g4)
	test.DeepEqual(t, g2, g3)
	test.DeepEqual(t, g2, g4)
	test.DeepEqual(t, g3, g4)
}

// TestMethodsReturnCanonicalForm checks that every method returning a
// geometry gives a result in canonical form (see [canonicalFormViolations]),
// which is what makes reflect.DeepEqual agree with ExactEquals. The methods
// are found by reflection, so new methods are checked without changes to
// this test.
func TestMethodsReturnCanonicalForm(t *testing.T) {
	receivers := canonicalFormReceivers(t)
	args := canonicalFormArgs()
	for _, zero := range []any{
		geom.Geometry{},
		geom.Point{},
		geom.LineString{},
		geom.Polygon{},
		geom.MultiPoint{},
		geom.MultiLineString{},
		geom.MultiPolygon{},
		geom.GeometryCollection{},
		geom.Sequence{},
		geom.Envelope{},
	} {
		typ := reflect.TypeOf(zero)
		for i := 0; i < typ.NumMethod(); i++ {
			method := typ.Method(i)
			if !returnsGeometry(method.Type) {
				continue
			}
			t.Run(typ.Name()+"."+method.Name, func(t *testing.T) {
				argSets := argumentSets(t, method.Type, args)
				var calls int
				for _, recv := range receivers[typ] {
					for _, argSet := range argSets {
						in := append([]reflect.Value{recv.value}, argSet...)
						results, ok := callRecovering(method.Func, in)
						if !ok {
							continue
						}
						calls++
						for _, r := range results {
							if isGeometryType(r.Type()) {
								checkCanonicalForm(t, recv.label+formatArgs(argSet), r)
							}
						}
					}
				}
				if calls == 0 {
					t.Error("every call panicked")
				}
			})
		}
	}
}

// TestFunctionsReturnCanonicalForm checks that package level functions
// returning geometries give results in canonical form. New functions must be
// added to the list.
func TestFunctionsReturnCanonicalForm(t *testing.T) {
	poly := test.FromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	for _, fn := range []struct {
		name string
		fn   func(geom.Geometry) (any, error)
	}{
		{"UnmarshalWKT", func(g geom.Geometry) (any, error) {
			return geom.UnmarshalWKT(g.AsText())
		}},
		{"UnmarshalWKB", func(g geom.Geometry) (any, error) {
			return geom.UnmarshalWKB(g.AsBinary())
		}},
		{"UnmarshalGeoJSON", func(g geom.Geometry) (any, error) {
			buf, err := g.MarshalJSON()
			if err != nil {
				return nil, err
			}
			return geom.UnmarshalGeoJSON(buf)
		}},
		{"UnmarshalTWKB", func(g geom.Geometry) (any, error) {
			buf, err := geom.MarshalTWKB(g, 3)
			if err != nil {
				return nil, err
			}
			return geom.UnmarshalTWKB(buf)
		}},
		{"NewGeometryCollection", func(g geom.Geometry) (any, error) {
			return geom.NewGeometryCollection([]geom.Geometry{g, poly}), nil
		}},
		{"Union", func(g geom.Geometry) (any, error) { return geom.Union(g, poly) }},
		{"Intersection", func(g geom.Geometry) (any, error) { return geom.Intersection(g, poly) }},
		{"Difference", func(g geom.Geometry) (any, error) { return geom.Difference(g, poly) }},
		{"SymmetricDifference", func(g geom.Geometry) (any, error) { return geom.SymmetricDifference(g, poly) }},
		{"UnaryUnion", func(g geom.Geometry) (any, error) { return geom.UnaryUnion(g) }},
		{"UnionMany", func(g geom.Geometry) (any, error) { return geom.UnionMany([]geom.Geometry{g, poly}) }},
		{"Buffer", func(g geom.Geometry) (any, error) { return geom.Buffer(g, 0.5) }},
		{"RotatedMinimumAreaBoundingRectangle", func(g geom.Geometry) (any, error) {
			return geom.RotatedMinimumAreaBoundingRectangle(g), nil
		}},
		{"RotatedMinimumWidthBoundingRectangle", func(g geom.Geometry) (any, error) {
			return geom.RotatedMinimumWidthBoundingRectangle(g), nil
		}},
	} {
		t.Run(fn.name, func(t *testing.T) {
			var calls int
			for _, wkt := range canonicalFormInputs() {
				result, err := fn.fn(test.FromWKT(t, wkt))
				if err != nil {
					continue
				}
				calls++
				checkCanonicalForm(t, wkt, reflect.ValueOf(result))
			}
			if calls == 0 {
				t.Error("every call returned an error")
			}
		})
	}
}

// canonicalFormInputs returns WKTs covering each geometry type in each
// coordinates type: empty, with an empty member, non-empty, and nested in
// collections.
func canonicalFormInputs() []string {
	var wkts []string
	for _, dim := range []struct{ suffix, extra string }{
		{"", ""}, {" Z", " 7"}, {" M", " 8"}, {" ZM", " 7 8"},
	} {
		coords := func(xys ...float64) string {
			var parts []string
			for i := 0; i+1 < len(xys); i += 2 {
				parts = append(parts, fmt.Sprintf("%g %g%s", xys[i], xys[i+1], dim.extra))
			}
			return strings.Join(parts, ",")
		}
		pt := coords(1, 2)
		ls := coords(0, 0, 2, 3)
		cwShell := coords(0, 0, 0, 3, 3, 0, 0, 0)
		ccwShell := coords(0, 0, 3, 0, 0, 3, 0, 0)
		ccwHole := coords(0.5, 0.5, 1, 0.5, 0.5, 1, 0.5, 0.5)

		s := dim.suffix
		base := []string{
			"POINT" + s + " EMPTY",
			"POINT" + s + "(" + pt + ")",
			"LINESTRING" + s + " EMPTY",
			"LINESTRING" + s + "(" + ls + ")",
			"POLYGON" + s + " EMPTY",
			"POLYGON" + s + "((" + cwShell + "),(" + ccwHole + "))",
			"POLYGON" + s + "((" + ccwShell + "))",
			"MULTIPOINT" + s + " EMPTY",
			"MULTIPOINT" + s + "(EMPTY)",
			"MULTIPOINT" + s + "((" + pt + "),EMPTY)",
			"MULTILINESTRING" + s + " EMPTY",
			"MULTILINESTRING" + s + "(EMPTY)",
			"MULTILINESTRING" + s + "((" + ls + "),EMPTY)",
			"MULTIPOLYGON" + s + " EMPTY",
			"MULTIPOLYGON" + s + "(EMPTY)",
			"MULTIPOLYGON" + s + "(((" + cwShell + ")),EMPTY)",
			"MULTIPOLYGON" + s + "(((" + ccwShell + ")),EMPTY)",
		}
		wkts = append(wkts, base...)
		for _, b := range base {
			wkts = append(wkts, "GEOMETRYCOLLECTION"+s+"("+b+")")
		}
		wkts = append(wkts,
			"GEOMETRYCOLLECTION"+s+" EMPTY",
			"GEOMETRYCOLLECTION"+s+"(GEOMETRYCOLLECTION"+s+" EMPTY)",
			"GEOMETRYCOLLECTION"+s+"(GEOMETRYCOLLECTION"+s+"(POLYGON"+s+" EMPTY))",
		)
	}
	return wkts
}

type canonicalFormReceiver struct {
	label string
	value reflect.Value
}

// canonicalFormReceivers derives receivers for each type's methods from
// [canonicalFormInputs]. It reports any receiver that isn't in canonical
// form, since a method that returns its receiver unchanged would otherwise
// take the blame.
func canonicalFormReceivers(t *testing.T) map[reflect.Type][]canonicalFormReceiver {
	t.Helper()
	receivers := make(map[reflect.Type][]canonicalFormReceiver)
	for _, wkt := range canonicalFormInputs() {
		g := test.FromWKT(t, wkt)
		var concrete any
		switch g.Type() {
		case geom.TypePoint:
			concrete = g.MustAsPoint()
		case geom.TypeLineString:
			concrete = g.MustAsLineString()
		case geom.TypePolygon:
			concrete = g.MustAsPolygon()
		case geom.TypeMultiPoint:
			concrete = g.MustAsMultiPoint()
		case geom.TypeMultiLineString:
			concrete = g.MustAsMultiLineString()
		case geom.TypeMultiPolygon:
			concrete = g.MustAsMultiPolygon()
		case geom.TypeGeometryCollection:
			concrete = g.MustAsGeometryCollection()
		}
		for _, recv := range []any{g, concrete, g.DumpCoordinates(), g.Envelope()} {
			v := reflect.ValueOf(recv)
			if isGeometryType(v.Type()) {
				checkCanonicalForm(t, wkt, v)
			}
			receivers[v.Type()] = append(receivers[v.Type()], canonicalFormReceiver{wkt, v})
		}
	}
	return receivers
}

func canonicalFormArgs() map[reflect.Type][]reflect.Value {
	transformXY := func(xy geom.XY) geom.XY { return xy }
	transform := func(geom.CoordinatesType, []float64) error { return nil }
	return map[reflect.Type][]reflect.Value{
		reflect.TypeOf(0.0): {reflect.ValueOf(0.5)},
		reflect.TypeOf(0):   {reflect.ValueOf(0), reflect.ValueOf(1)},
		reflect.TypeOf(geom.DimXY): {
			reflect.ValueOf(geom.DimXY),
			reflect.ValueOf(geom.DimXYZ),
			reflect.ValueOf(geom.DimXYM),
			reflect.ValueOf(geom.DimXYZM),
		},
		reflect.TypeOf(transformXY): {reflect.ValueOf(transformXY)},
		reflect.TypeOf(transform):   {reflect.ValueOf(transform)},
	}
}

func returnsGeometry(fnType reflect.Type) bool {
	for i := 0; i < fnType.NumOut(); i++ {
		if isGeometryType(fnType.Out(i)) {
			return true
		}
	}
	return false
}

// isGeometryType reports whether typ is a geometry type, or a slice of a
// geometry type.
func isGeometryType(typ reflect.Type) bool {
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}
	switch typ {
	case reflect.TypeOf(geom.Geometry{}),
		reflect.TypeOf(geom.Point{}),
		reflect.TypeOf(geom.LineString{}),
		reflect.TypeOf(geom.Polygon{}),
		reflect.TypeOf(geom.MultiPoint{}),
		reflect.TypeOf(geom.MultiLineString{}),
		reflect.TypeOf(geom.MultiPolygon{}),
		reflect.TypeOf(geom.GeometryCollection{}),
		reflect.TypeOf(geom.Sequence{}):
		return true
	}
	return false
}

// argumentSets returns every combination of arguments from args for the
// parameters of the method type fnType, leaving out the receiver and any
// variadic parameter.
func argumentSets(t *testing.T, fnType reflect.Type, args map[reflect.Type][]reflect.Value) [][]reflect.Value {
	t.Helper()
	n := fnType.NumIn()
	if fnType.IsVariadic() {
		n--
	}
	sets := [][]reflect.Value{nil}
	for i := 1; i < n; i++ {
		options, ok := args[fnType.In(i)]
		if !ok {
			t.Fatalf("no arguments for parameter type %v, add some to canonicalFormArgs", fnType.In(i))
		}
		var next [][]reflect.Value
		for _, set := range sets {
			for _, opt := range options {
				next = append(next, append(set[:len(set):len(set)], opt))
			}
		}
		sets = next
	}
	return sets
}

// callRecovering calls fn, reporting false if it panics. Some methods panic
// for some inputs, such as PointN with an index that is out of range.
func callRecovering(fn reflect.Value, in []reflect.Value) (results []reflect.Value, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return fn.Call(in), true
}

func formatArgs(args []reflect.Value) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, a := range args {
		if a.Kind() == reflect.Func {
			parts[i] = "func"
		} else {
			parts[i] = fmt.Sprint(a.Interface())
		}
	}
	return " with args (" + strings.Join(parts, ", ") + ")"
}

// checkCanonicalForm fails the test if v isn't in canonical form. A slice
// is checked element by element, since the slice itself isn't a geometry.
func checkCanonicalForm(t *testing.T, label string, v reflect.Value) {
	t.Helper()
	var violations []string
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			violations = append(violations, canonicalFormViolations(v.Index(i), fmt.Sprintf("[%d]", i))...)
		}
	} else {
		violations = canonicalFormViolations(v, v.Type().Name())
	}
	for _, msg := range violations {
		t.Errorf("%s: %s", label, msg)
	}
}

// canonicalFormViolations reads v through its unexported fields and
// describes each way it differs from canonical form. reflect.DeepEqual agrees
// with ExactEquals on values in canonical form, where:
//
//   - Every slice is nil or non-empty.
//   - An empty Point has zero X, Y, Z and M values.
//   - A Geometry holding an empty XY GeometryCollection is the zero Geometry.
func canonicalFormViolations(v reflect.Value, path string) []string {
	var violations []string
	switch v.Kind() {
	case reflect.Slice:
		if !v.IsNil() && v.Len() == 0 {
			violations = append(violations, path+": empty slice that isn't nil")
		}
		for i := 0; i < v.Len(); i++ {
			violations = append(violations, canonicalFormViolations(v.Index(i), fmt.Sprintf("%s[%d]", path, i))...)
		}
	case reflect.Interface:
		if !v.IsNil() {
			violations = append(violations, canonicalFormViolations(v.Elem(), path+"("+v.Elem().Type().Name()+")")...)
		}
	case reflect.Struct:
		switch v.Type() {
		case reflect.TypeOf(geom.Point{}):
			coords := v.FieldByName("coords")
			if !v.FieldByName("full").Bool() {
				for _, name := range []string{"X", "Y", "Z", "M"} {
					if coords.FieldByName(name).Float() != 0 {
						violations = append(violations, path+": empty Point with non-zero "+name)
					}
				}
			}
		case reflect.TypeOf(geom.Geometry{}):
			impl := v.FieldByName("impl")
			if !impl.IsNil() && impl.Elem().Type() == reflect.TypeOf(geom.GeometryCollection{}) {
				gc := impl.Elem()
				if gc.FieldByName("geoms").Len() == 0 && geom.CoordinatesType(gc.FieldByName("ctype").Uint()) == geom.DimXY {
					violations = append(violations, path+": empty XY GeometryCollection that isn't the zero Geometry")
				}
			}
		}
		for i := 0; i < v.NumField(); i++ {
			violations = append(violations, canonicalFormViolations(v.Field(i), path+"."+v.Type().Field(i).Name)...)
		}
	}
	return violations
}
