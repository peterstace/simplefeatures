package jts

import "math"

// OperationOverlayng_PrecisionUtil provides functions for computing precision
// model scale factors that ensure robust geometry operations.

// OperationOverlayng_PrecisionUtil_MAX_ROBUST_DP_DIGITS is a number of digits
// of precision which leaves some computational "headroom" to ensure robust
// evaluation of certain double-precision floating point geometric operations.
// This value should be less than the maximum decimal precision of
// double-precision values (16).
const OperationOverlayng_PrecisionUtil_MAX_ROBUST_DP_DIGITS = 14

// OperationOverlayng_PrecisionUtil_RobustPM determines a precision model to use
// for robust overlay operations. WARNING: this is very slow.
func OperationOverlayng_PrecisionUtil_RobustPM(a, b *Geom_Geometry) *Geom_PrecisionModel {
	scale := OperationOverlayng_PrecisionUtil_RobustScale(a, b)
	return Geom_NewPrecisionModelWithScale(scale)
}

// OperationOverlayng_PrecisionUtil_RobustPMSingle determines a precision model
// for a single geometry. WARNING: this is very slow.
func OperationOverlayng_PrecisionUtil_RobustPMSingle(a *Geom_Geometry) *Geom_PrecisionModel {
	scale := OperationOverlayng_PrecisionUtil_RobustScaleSingle(a)
	return Geom_NewPrecisionModelWithScale(scale)
}

// OperationOverlayng_PrecisionUtil_SafeScale computes a safe scale factor for
// a numeric value.
func OperationOverlayng_PrecisionUtil_SafeScale(value float64) float64 {
	return operationOverlayng_PrecisionUtil_precisionScale(value, OperationOverlayng_PrecisionUtil_MAX_ROBUST_DP_DIGITS)
}

// OperationOverlayng_PrecisionUtil_SafeScaleGeom computes a safe scale factor
// for a geometry.
func OperationOverlayng_PrecisionUtil_SafeScaleGeom(geom *Geom_Geometry) float64 {
	return OperationOverlayng_PrecisionUtil_SafeScale(operationOverlayng_PrecisionUtil_maxBoundMagnitude(geom.GetEnvelopeInternal()))
}

// OperationOverlayng_PrecisionUtil_SafeScaleGeoms computes a safe scale factor
// for two geometries.
func OperationOverlayng_PrecisionUtil_SafeScaleGeoms(a, b *Geom_Geometry) float64 {
	maxBnd := operationOverlayng_PrecisionUtil_maxBoundMagnitude(a.GetEnvelopeInternal())
	if b != nil {
		maxBndB := operationOverlayng_PrecisionUtil_maxBoundMagnitude(b.GetEnvelopeInternal())
		maxBnd = math.Max(maxBnd, maxBndB)
	}
	return OperationOverlayng_PrecisionUtil_SafeScale(maxBnd)
}

func operationOverlayng_PrecisionUtil_maxBoundMagnitude(env *Geom_Envelope) float64 {
	return Math_MathUtil_Max4(
		math.Abs(env.GetMaxX()),
		math.Abs(env.GetMaxY()),
		math.Abs(env.GetMinX()),
		math.Abs(env.GetMinY()),
	)
}

// precisionScale computes the scale factor which will produce a given number
// of digits of precision when used to round the given number.
func operationOverlayng_PrecisionUtil_precisionScale(value float64, precisionDigits int) float64 {
	// The smallest power of 10 greater than the value.
	// Use log/log(10) instead of log10 to match Java's floating-point behavior.
	magnitude := int(math.Log(value)/math.Log(10) + 1.0)
	precDigits := precisionDigits - magnitude

	scaleFactor := math.Pow(10.0, float64(precDigits))
	return scaleFactor
}

// OperationOverlayng_PrecisionUtil_InherentScale computes the inherent scale of
// a number. The inherent scale is the scale factor for rounding which preserves
// all digits of precision (significant digits) present in the numeric value.
// WARNING: this is very slow.
func OperationOverlayng_PrecisionUtil_InherentScale(value float64) float64 {
	numDec := operationOverlayng_PrecisionUtil_numberOfDecimals(value)
	scaleFactor := math.Pow(10.0, float64(numDec))
	return scaleFactor
}

// OperationOverlayng_PrecisionUtil_InherentScaleGeom computes the inherent
// scale of a geometry. WARNING: this is very slow.
func OperationOverlayng_PrecisionUtil_InherentScaleGeom(geom *Geom_Geometry) float64 {
	scaleFilter := newInherentScaleFilter()
	geom.ApplyCoordinateFilter(scaleFilter)
	return scaleFilter.GetScale()
}

// OperationOverlayng_PrecisionUtil_InherentScaleGeoms computes the inherent
// scale of two geometries. WARNING: this is very slow.
func OperationOverlayng_PrecisionUtil_InherentScaleGeoms(a, b *Geom_Geometry) float64 {
	scale := OperationOverlayng_PrecisionUtil_InherentScaleGeom(a)
	if b != nil {
		scaleB := OperationOverlayng_PrecisionUtil_InherentScaleGeom(b)
		scale = math.Max(scale, scaleB)
	}
	return scale
}

// numberOfDecimals determines the number of decimal places represented in a
// double-precision number.
func operationOverlayng_PrecisionUtil_numberOfDecimals(value float64) int {
	// Ensure that scientific notation is NOT used.
	s := Io_OrdinateFormat_Default.Format(value)
	if len(s) >= 2 && s[len(s)-2:] == ".0" {
		return 0
	}
	decIndex := -1
	for i, c := range s {
		if c == '.' {
			decIndex = i
			break
		}
	}
	if decIndex <= 0 {
		return 0
	}
	return len(s) - decIndex - 1
}

type inherentScaleFilter struct {
	scale float64
}

var _ Geom_CoordinateFilter = (*inherentScaleFilter)(nil)

func (f *inherentScaleFilter) IsGeom_CoordinateFilter() {}

func newInherentScaleFilter() *inherentScaleFilter {
	return &inherentScaleFilter{scale: 0}
}

func (f *inherentScaleFilter) GetScale() float64 {
	return f.scale
}

func (f *inherentScaleFilter) Filter(coord *Geom_Coordinate) {
	f.updateScaleMax(coord.GetX())
	f.updateScaleMax(coord.GetY())
}

func (f *inherentScaleFilter) updateScaleMax(value float64) {
	scaleVal := OperationOverlayng_PrecisionUtil_InherentScale(value)
	if scaleVal > f.scale {
		f.scale = scaleVal
	}
}

// OperationOverlayng_PrecisionUtil_RobustScale determines a scale factor which
// maximizes the digits of precision and is safe to use for overlay operations.
// WARNING: this is very slow.
func OperationOverlayng_PrecisionUtil_RobustScale(a, b *Geom_Geometry) float64 {
	inherentScale := OperationOverlayng_PrecisionUtil_InherentScaleGeoms(a, b)
	safeScale := OperationOverlayng_PrecisionUtil_SafeScaleGeoms(a, b)
	return operationOverlayng_PrecisionUtil_robustScale(inherentScale, safeScale)
}

// OperationOverlayng_PrecisionUtil_RobustScaleSingle determines a scale factor
// for a single geometry. WARNING: this is very slow.
func OperationOverlayng_PrecisionUtil_RobustScaleSingle(a *Geom_Geometry) float64 {
	inherentScale := OperationOverlayng_PrecisionUtil_InherentScaleGeom(a)
	safeScale := OperationOverlayng_PrecisionUtil_SafeScaleGeom(a)
	return operationOverlayng_PrecisionUtil_robustScale(inherentScale, safeScale)
}

func operationOverlayng_PrecisionUtil_robustScale(inherentScale, safeScale float64) float64 {
	// Use safe scale if lower, since it is important to preserve some precision
	// for robustness.
	if inherentScale <= safeScale {
		return inherentScale
	}
	return safeScale
}
