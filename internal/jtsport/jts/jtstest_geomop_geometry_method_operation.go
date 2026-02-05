package jts

import (
	"reflect"
	"strconv"
	"strings"
)

var _ JtstestGeomop_GeometryOperation = (*JtstestGeomop_GeometryMethodOperation)(nil)

// JtstestGeomop_GeometryMethodOperation_IsBooleanFunction returns true if the
// named function returns a boolean.
func JtstestGeomop_GeometryMethodOperation_IsBooleanFunction(name string) bool {
	return jtstestGeomop_GeometryMethodOperation_getGeometryReturnType(name) == reflect.TypeOf(false)
}

// JtstestGeomop_GeometryMethodOperation_IsIntegerFunction returns true if the
// named function returns an integer.
func JtstestGeomop_GeometryMethodOperation_IsIntegerFunction(name string) bool {
	return jtstestGeomop_GeometryMethodOperation_getGeometryReturnType(name) == reflect.TypeOf(0)
}

// JtstestGeomop_GeometryMethodOperation_IsDoubleFunction returns true if the
// named function returns a double.
func JtstestGeomop_GeometryMethodOperation_IsDoubleFunction(name string) bool {
	return jtstestGeomop_GeometryMethodOperation_getGeometryReturnType(name) == reflect.TypeOf(0.0)
}

// JtstestGeomop_GeometryMethodOperation_IsGeometryFunction returns true if the
// named function returns a Geometry.
func JtstestGeomop_GeometryMethodOperation_IsGeometryFunction(name string) bool {
	rt := jtstestGeomop_GeometryMethodOperation_getGeometryReturnType(name)
	if rt == nil {
		return false
	}
	return jtstestGeomop_GeometryMethodOperation_isGeometryType(rt)
}

func jtstestGeomop_GeometryMethodOperation_getGeometryReturnType(functionName string) reflect.Type {
	geomType := reflect.TypeOf((*Geom_Geometry)(nil))
	for i := 0; i < geomType.NumMethod(); i++ {
		method := geomType.Method(i)
		if !strings.EqualFold(method.Name, functionName) {
			continue
		}
		methodType := method.Type
		if methodType.NumOut() == 0 {
			continue
		}
		returnClass := methodType.Out(0)
		// Filter out only acceptable classes. (For instance, don't accept the
		// relate()=>IntersectionMatrix method.)
		if returnClass.Kind() == reflect.Bool ||
			jtstestGeomop_GeometryMethodOperation_isGeometryType(returnClass) ||
			returnClass.Kind() == reflect.Float64 ||
			returnClass.Kind() == reflect.Int {
			return returnClass
		}
	}
	return nil
}

// jtstestGeomop_GeometryMethodOperation_isGeometryType checks if a type is a
// geometry type (either *Geom_Geometry or a type that embeds *Geom_Geometry).
func jtstestGeomop_GeometryMethodOperation_isGeometryType(t reflect.Type) bool {
	geomType := reflect.TypeOf((*Geom_Geometry)(nil))
	if t == geomType {
		return true
	}
	// Check if it's a pointer to a struct that embeds *Geom_Geometry.
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		elem := t.Elem()
		for j := 0; j < elem.NumField(); j++ {
			if elem.Field(j).Type == geomType {
				return true
			}
		}
	}
	return false
}

// JtstestGeomop_GeometryMethodOperation invokes a named operation on a set of
// arguments, the first of which is a Geometry. This class provides operations
// which are the methods defined on the Geometry class. Other GeometryOperation
// classes can delegate to instances of this class to run standard Geometry
// methods.
type JtstestGeomop_GeometryMethodOperation struct {
	geometryMethods []reflect.Method
	convArg         [1]any
}

func JtstestGeomop_NewGeometryMethodOperation() *JtstestGeomop_GeometryMethodOperation {
	geomType := reflect.TypeOf((*Geom_Geometry)(nil))
	methods := make([]reflect.Method, geomType.NumMethod())
	for i := 0; i < geomType.NumMethod(); i++ {
		methods[i] = geomType.Method(i)
	}
	return &JtstestGeomop_GeometryMethodOperation{
		geometryMethods: methods,
	}
}

func (op *JtstestGeomop_GeometryMethodOperation) IsJtstestGeomop_GeometryOperation() {}

func (op *JtstestGeomop_GeometryMethodOperation) GetReturnType(opName string) string {
	rt := jtstestGeomop_GeometryMethodOperation_getGeometryReturnType(opName)
	if rt == nil {
		return ""
	}
	if rt.Kind() == reflect.Bool {
		return "boolean"
	}
	if rt.Kind() == reflect.Int {
		return "int"
	}
	if rt.Kind() == reflect.Float64 {
		return "double"
	}
	if jtstestGeomop_GeometryMethodOperation_isGeometryType(rt) {
		return "geometry"
	}
	return ""
}

func (op *JtstestGeomop_GeometryMethodOperation) Invoke(
	opName string,
	geometry *Geom_Geometry,
	args []any,
) (JtstestTestrunner_Result, error) {
	// Check for TestCaseGeometryFunctions operations first.
	if result := op.invokeTestCaseGeometryFunction(opName, geometry, args); result != nil {
		return result, nil
	}

	actualArgs := make([]any, len(args))
	geomMethod := op.getGeometryMethod(opName, args, actualArgs)
	if geomMethod == nil {
		return nil, JtstestTestrunner_NewJTSTestReflectionException(opName, args)
	}
	return op.invokeMethod(geomMethod, geometry, actualArgs)
}

func (op *JtstestGeomop_GeometryMethodOperation) getGeometryMethod(
	opName string,
	args []any,
	actualArgs []any,
) *reflect.Method {
	// Normalize operation name to handle Go-specific method naming.
	normalizedName := jtstestGeomop_GeometryMethodOperation_normalizeOpName(opName, len(args))

	// Could index methods by name for efficiency...
	for i := range op.geometryMethods {
		if !strings.EqualFold(op.geometryMethods[i].Name, normalizedName) {
			continue
		}
		if op.convertArgs(op.geometryMethods[i].Type, args, actualArgs) {
			return &op.geometryMethods[i]
		}
	}
	return nil
}

// jtstestGeomop_GeometryMethodOperation_normalizeOpName maps test operation
// names to Go method names. This handles:
// - NG suffixes (unionNG -> Union) that use the same method in Go
// - Java's union() (0-arg) -> Go's UnionSelf()
func jtstestGeomop_GeometryMethodOperation_normalizeOpName(opName string, argCount int) string {
	opLower := strings.ToLower(opName)

	// Handle zero-arg union -> UnionSelf.
	if opLower == "union" && argCount == 0 {
		return "UnionSelf"
	}

	// Strip NG/SR suffixes - they use the same underlying methods in Go.
	// The overlay implementation is controlled by a global setting.
	opLower = strings.TrimSuffix(opLower, "ng")
	opLower = strings.TrimSuffix(opLower, "sr")

	return opLower
}

func jtstestGeomop_GeometryMethodOperation_nonNullItemCount(obj []any) int {
	count := 0
	for i := 0; i < len(obj); i++ {
		if obj[i] != nil {
			count++
		}
	}
	return count
}

func (op *JtstestGeomop_GeometryMethodOperation) convertArgs(
	methodType reflect.Type,
	args []any,
	actualArgs []any,
) bool {
	// methodType includes receiver as first param, so NumIn()-1 is the actual param count.
	paramCount := methodType.NumIn() - 1
	if paramCount != jtstestGeomop_GeometryMethodOperation_nonNullItemCount(args) {
		return false
	}
	for i := 0; i < len(args); i++ {
		// +1 to skip receiver.
		paramType := methodType.In(i + 1)
		isCompatible := op.convertArg(paramType, args[i])
		if !isCompatible {
			return false
		}
		actualArgs[i] = op.convArg[0]
	}
	return true
}

func (op *JtstestGeomop_GeometryMethodOperation) convertArg(
	destClass reflect.Type,
	srcValue any,
) bool {
	op.convArg[0] = nil
	if srcStr, ok := srcValue.(string); ok {
		return op.convertArgFromString(destClass, srcStr)
	}
	srcType := reflect.TypeOf(srcValue)
	if srcType.AssignableTo(destClass) {
		op.convArg[0] = srcValue
		return true
	}
	return false
}

func (op *JtstestGeomop_GeometryMethodOperation) convertArgFromString(
	destClass reflect.Type,
	srcStr string,
) bool {
	op.convArg[0] = nil
	if destClass.Kind() == reflect.Bool {
		if srcStr == "true" {
			op.convArg[0] = true
			return true
		} else if srcStr == "false" {
			op.convArg[0] = false
			return true
		}
		return false
	}
	if destClass.Kind() == reflect.Int {
		// Try as an int.
		if i, err := strconv.Atoi(srcStr); err == nil {
			op.convArg[0] = i
			return true
		}
		return false
	}
	if destClass.Kind() == reflect.Float64 {
		// Try as a double.
		if f, err := strconv.ParseFloat(srcStr, 64); err == nil {
			op.convArg[0] = f
			return true
		}
		return false
	}
	if destClass.Kind() == reflect.String {
		op.convArg[0] = srcStr
		return true
	}
	return false
}

func (op *JtstestGeomop_GeometryMethodOperation) invokeMethod(
	method *reflect.Method,
	geometry *Geom_Geometry,
	args []any,
) (JtstestTestrunner_Result, error) {
	// Build args for Call: receiver + args.
	callArgs := make([]reflect.Value, len(args)+1)
	callArgs[0] = reflect.ValueOf(geometry)
	for i, arg := range args {
		callArgs[i+1] = reflect.ValueOf(arg)
	}

	results := method.Func.Call(callArgs)
	if len(results) == 0 {
		return nil, JtstestTestrunner_NewJTSTestReflectionExceptionWithMessage(
			"Unsupported result type: void")
	}

	result := results[0]
	returnType := result.Type()

	if returnType.Kind() == reflect.Bool {
		return JtstestTestrunner_NewBooleanResult(result.Bool()), nil
	}
	if jtstestGeomop_GeometryMethodOperation_isGeometryType(returnType) {
		geom := jtstestGeomop_GeometryMethodOperation_extractGeometry(result)
		return JtstestTestrunner_NewGeometryResult(geom), nil
	}
	if returnType.Kind() == reflect.Float64 {
		return JtstestTestrunner_NewDoubleResult(result.Float()), nil
	}
	if returnType.Kind() == reflect.Int {
		return JtstestTestrunner_NewIntegerResult(int(result.Int())), nil
	}
	return nil, JtstestTestrunner_NewJTSTestReflectionExceptionWithMessage(
		"Unsupported result type: " + returnType.String())
}

// jtstestGeomop_GeometryMethodOperation_extractGeometry extracts the
// *Geom_Geometry from a value that may be *Geom_Geometry or a geometry subtype
// (like *Geom_Point) that embeds *Geom_Geometry.
func jtstestGeomop_GeometryMethodOperation_extractGeometry(v reflect.Value) *Geom_Geometry {
	if v.IsNil() {
		return nil
	}
	// Direct *Geom_Geometry.
	if g, ok := v.Interface().(*Geom_Geometry); ok {
		return g
	}
	// For subtypes like *Geom_Point, access the embedded Geom_Geometry field.
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		elem := v.Elem()
		geomType := reflect.TypeOf((*Geom_Geometry)(nil))
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			if field.Type() == geomType {
				return field.Interface().(*Geom_Geometry)
			}
		}
	}
	return nil
}

// invokeTestCaseGeometryFunction handles operations from TestCaseGeometryFunctions.
// These are operations like intersectionNG, unionNG, etc. that should call the
// OverlayNG operations directly rather than going through Geometry methods.
// Returns nil if the operation is not a TestCaseGeometryFunctions operation.
func (op *JtstestGeomop_GeometryMethodOperation) invokeTestCaseGeometryFunction(
	opName string,
	geometry *Geom_Geometry,
	args []any,
) JtstestTestrunner_Result {
	opLower := strings.ToLower(opName)

	// Handle zero-argument operations.
	if len(args) == 0 {
		switch opLower {
		case "minclearance":
			return JtstestTestrunner_NewDoubleResult(
				JtstestGeomop_TestCaseGeometryFunctions_MinClearance(geometry))
		case "minclearanceline":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_MinClearanceLine(geometry))
		case "polygonize":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_Polygonize(geometry))
		}
	}

	// Handle single-argument operations (geometry + distance/parameter).
	if len(args) == 1 {
		// Check for operations that take a distance, not a geometry.
		switch opLower {
		case "buffermitredjoin":
			distance, ok := op.parseScale(args[0])
			if !ok {
				return nil
			}
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_BufferMitredJoin(geometry, distance))
		case "densify":
			distance, ok := op.parseScale(args[0])
			if !ok {
				return nil
			}
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_Densify(geometry, distance))
		case "simplifydp":
			distance, ok := op.parseScale(args[0])
			if !ok {
				return nil
			}
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_SimplifyDP(geometry, distance))
		case "simplifytp":
			distance, ok := op.parseScale(args[0])
			if !ok {
				return nil
			}
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_SimplifyTP(geometry, distance))
		}

		// Handle two-geometry NG operations.
		geom1, ok := args[0].(*Geom_Geometry)
		if !ok {
			return nil
		}
		switch opLower {
		case "intersectionng":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_IntersectionNG(geometry, geom1))
		case "unionng":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_UnionNG(geometry, geom1))
		case "differenceng":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_DifferenceNG(geometry, geom1))
		case "symdifferenceng":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_SymDifferenceNG(geometry, geom1))
		}
	}

	// Handle two-geometry SR operations (geometry + scale).
	if len(args) == 2 {
		geom1, ok := args[0].(*Geom_Geometry)
		if !ok {
			return nil
		}
		scale, ok := op.parseScale(args[1])
		if !ok {
			return nil
		}
		switch opLower {
		case "intersectionsr":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_IntersectionSR(geometry, geom1, scale))
		case "unionsr":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_UnionSR(geometry, geom1, scale))
		case "differencesr":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_DifferenceSR(geometry, geom1, scale))
		case "symdifferencesr":
			return JtstestTestrunner_NewGeometryResult(
				JtstestGeomop_TestCaseGeometryFunctions_SymDifferenceSR(geometry, geom1, scale))
		}
	}

	return nil
}

// parseScale parses a scale value from a string or numeric type.
func (op *JtstestGeomop_GeometryMethodOperation) parseScale(arg any) (float64, bool) {
	switch v := arg.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	}
	return 0, false
}
