package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

var operationBufferValidate_BufferResultValidator_VERBOSE = false

// operationBufferValidate_BufferResultValidator_MAX_ENV_DIFF_FRAC is the
// maximum allowable fraction of buffer distance the actual distance can differ by.
// 1% sometimes causes an error - 1.2% should be safe.
const operationBufferValidate_BufferResultValidator_MAX_ENV_DIFF_FRAC = .012

// OperationBufferValidate_BufferResultValidator_IsValid checks whether the geometry buffer is valid.
func OperationBufferValidate_BufferResultValidator_IsValid(g *Geom_Geometry, distance float64, result *Geom_Geometry) bool {
	validator := OperationBufferValidate_NewBufferResultValidator(g, distance, result)
	if validator.IsValid() {
		return true
	}
	return false
}

// OperationBufferValidate_BufferResultValidator_IsValidMsg checks whether the geometry buffer is valid,
// and returns an error message if not.
//
// Returns an appropriate error message or empty string if the buffer is valid.
func OperationBufferValidate_BufferResultValidator_IsValidMsg(g *Geom_Geometry, distance float64, result *Geom_Geometry) string {
	validator := OperationBufferValidate_NewBufferResultValidator(g, distance, result)
	if !validator.IsValid() {
		return validator.GetErrorMessage()
	}
	return ""
}

// OperationBufferValidate_BufferResultValidator validates that the result of a buffer operation
// is geometrically correct, within a computed tolerance.
//
// This is a heuristic test, and may return false positive results
// (I.e. it may fail to detect an invalid result.)
// It should never return a false negative result, however
// (I.e. it should never report a valid result as invalid.)
//
// This test may be (much) more expensive than the original
// buffer computation.
type OperationBufferValidate_BufferResultValidator struct {
	input          *Geom_Geometry
	distance       float64
	result         *Geom_Geometry
	isValid        bool
	errorMsg       string
	errorLocation  *Geom_Coordinate
	errorIndicator *Geom_Geometry
}

// OperationBufferValidate_NewBufferResultValidator creates a new BufferResultValidator.
func OperationBufferValidate_NewBufferResultValidator(input *Geom_Geometry, distance float64, result *Geom_Geometry) *OperationBufferValidate_BufferResultValidator {
	return &OperationBufferValidate_BufferResultValidator{
		input:    input,
		distance: distance,
		result:   result,
		isValid:  true,
	}
}

// IsValid validates the buffer result.
func (v *OperationBufferValidate_BufferResultValidator) IsValid() bool {
	v.checkPolygonal()
	if !v.isValid {
		return v.isValid
	}
	v.checkExpectedEmpty()
	if !v.isValid {
		return v.isValid
	}
	v.checkEnvelope()
	if !v.isValid {
		return v.isValid
	}
	v.checkArea()
	if !v.isValid {
		return v.isValid
	}
	v.checkDistance()
	return v.isValid
}

// GetErrorMessage returns the error message.
func (v *OperationBufferValidate_BufferResultValidator) GetErrorMessage() string {
	return v.errorMsg
}

// GetErrorLocation returns the location of the error.
func (v *OperationBufferValidate_BufferResultValidator) GetErrorLocation() *Geom_Coordinate {
	return v.errorLocation
}

// GetErrorIndicator gets a geometry which indicates the location and nature of a validation failure.
//
// If the failure is due to the buffer curve being too far or too close
// to the input, the indicator is a line segment showing the location and size
// of the discrepancy.
//
// Returns a geometric error indicator or nil if no error was found.
func (v *OperationBufferValidate_BufferResultValidator) GetErrorIndicator() *Geom_Geometry {
	return v.errorIndicator
}

func (v *OperationBufferValidate_BufferResultValidator) report(checkName string) {
	if !operationBufferValidate_BufferResultValidator_VERBOSE {
		return
	}
	status := "passed"
	if !v.isValid {
		status = "FAILED"
	}
	Util_Debug_Println("Check " + checkName + ": " + status)
}

func (v *OperationBufferValidate_BufferResultValidator) checkPolygonal() {
	if !java.InstanceOf[*Geom_Polygon](v.result) &&
		!java.InstanceOf[*Geom_MultiPolygon](v.result) {
		v.isValid = false
	}
	v.errorMsg = "Result is not polygonal"
	v.errorIndicator = v.result
	v.report("Polygonal")
}

func (v *OperationBufferValidate_BufferResultValidator) checkExpectedEmpty() {
	// can't check areal features
	if v.input.GetDimension() >= 2 {
		return
	}
	// can't check positive distances
	if v.distance > 0.0 {
		return
	}

	// at this point can expect an empty result
	if !v.result.IsEmpty() {
		v.isValid = false
		v.errorMsg = "Result is non-empty"
		v.errorIndicator = v.result
	}
	v.report("ExpectedEmpty")
}

func (v *OperationBufferValidate_BufferResultValidator) checkEnvelope() {
	if v.distance < 0.0 {
		return
	}

	padding := v.distance * operationBufferValidate_BufferResultValidator_MAX_ENV_DIFF_FRAC
	if padding == 0.0 {
		padding = 0.001
	}

	expectedEnv := Geom_NewEnvelopeFromEnvelope(v.input.GetEnvelopeInternal())
	expectedEnv.ExpandBy(v.distance)

	bufEnv := Geom_NewEnvelopeFromEnvelope(v.result.GetEnvelopeInternal())
	bufEnv.ExpandBy(padding)

	if !bufEnv.ContainsEnvelope(expectedEnv) {
		v.isValid = false
		v.errorMsg = "Buffer envelope is incorrect"
		v.errorIndicator = v.input.GetFactory().ToGeometry(bufEnv)
	}
	v.report("Envelope")
}

func (v *OperationBufferValidate_BufferResultValidator) checkArea() {
	inputArea := v.input.GetArea()
	resultArea := v.result.GetArea()

	if v.distance > 0.0 && inputArea > resultArea {
		v.isValid = false
		v.errorMsg = "Area of positive buffer is smaller than input"
		v.errorIndicator = v.result
	}
	if v.distance < 0.0 && inputArea < resultArea {
		v.isValid = false
		v.errorMsg = "Area of negative buffer is larger than input"
		v.errorIndicator = v.result
	}
	v.report("Area")
}

func (v *OperationBufferValidate_BufferResultValidator) checkDistance() {
	distValid := OperationBufferValidate_NewBufferDistanceValidator(v.input, v.distance, v.result)
	if !distValid.IsValid() {
		v.isValid = false
		v.errorMsg = distValid.GetErrorMessage()
		v.errorLocation = distValid.GetErrorLocation()
		v.errorIndicator = distValid.GetErrorIndicator()
	}
	v.report("Distance")
}
