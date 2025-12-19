package jts

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geom_PrecisionModelType represents the types of precision models which JTS supports.
type Geom_PrecisionModelType struct {
	name string
}

func (t *Geom_PrecisionModelType) String() string {
	return t.name
}

// Geom_PrecisionModel_Fixed indicates that coordinates have a fixed number of decimal places.
// The number of decimal places is determined by the log10 of the scale factor.
var Geom_PrecisionModel_Fixed = &Geom_PrecisionModelType{name: "FIXED"}

// Geom_PrecisionModel_Floating corresponds to the standard Java double-precision floating-point
// representation, which is based on the IEEE-754 standard.
var Geom_PrecisionModel_Floating = &Geom_PrecisionModelType{name: "FLOATING"}

// Geom_PrecisionModel_FloatingSingle corresponds to the standard Java single-precision
// floating-point representation, which is based on the IEEE-754 standard.
var Geom_PrecisionModel_FloatingSingle = &Geom_PrecisionModelType{name: "FLOATING SINGLE"}

// Geom_PrecisionModel_MaximumPreciseValue is the maximum precise value representable in a
// double. Since IEEE754 double-precision numbers allow 53 bits of mantissa,
// the value is equal to 2^53 - 1. This provides almost 16 decimal digits of
// precision.
const Geom_PrecisionModel_MaximumPreciseValue = 9007199254740992.0

// Geom_PrecisionModel_MostPrecise determines which of two PrecisionModels is the most precise
// (allows the greatest number of significant digits).
func Geom_PrecisionModel_MostPrecise(pm1, pm2 *Geom_PrecisionModel) *Geom_PrecisionModel {
	if pm1.CompareTo(pm2) >= 0 {
		return pm1
	}
	return pm2
}

// Geom_PrecisionModel specifies the precision model of the Coordinates in a Geometry.
// In other words, specifies the grid of allowable points for a Geometry.
// A precision model may be floating (Floating or FloatingSingle), in which case
// normal floating-point value semantics apply.
//
// For a Fixed precision model the MakePrecise method allows rounding a coordinate
// to a "precise" value; that is, one whose precision is known exactly.
//
// Coordinates are assumed to be precise in geometries. That is, the coordinates
// are assumed to be rounded to the precision model given for the geometry. All
// internal operations assume that coordinates are rounded to the precision
// model. Constructive methods (such as boolean operations) always round computed
// coordinates to the appropriate precision model.
//
// Three types of precision model are supported:
//   - Floating: represents full double precision floating point. This is the
//     default precision model used in JTS
//   - FloatingSingle: represents single precision floating point
//   - Fixed: represents a model with a fixed number of decimal places. A Fixed
//     Precision Model is specified by a scale factor. The scale factor specifies
//     the size of the grid which numbers are rounded to.
type Geom_PrecisionModel struct {
	modelType *Geom_PrecisionModelType
	scale     float64
	gridSize  float64
}

// Geom_NewPrecisionModel creates a PrecisionModel with a default precision of Floating.
func Geom_NewPrecisionModel() *Geom_PrecisionModel {
	return &Geom_PrecisionModel{
		modelType: Geom_PrecisionModel_Floating,
	}
}

// Geom_NewPrecisionModelWithType creates a PrecisionModel that specifies an explicit
// precision model type. If the model type is Fixed the scale factor will
// default to 1.
func Geom_NewPrecisionModelWithType(modelType *Geom_PrecisionModelType) *Geom_PrecisionModel {
	pm := &Geom_PrecisionModel{
		modelType: modelType,
	}
	if modelType == Geom_PrecisionModel_Fixed {
		pm.setScale(1.0)
	}
	return pm
}

// Geom_NewPrecisionModelWithScale creates a PrecisionModel that specifies Fixed
// precision. Fixed-precision coordinates are represented as precise internal
// coordinates, which are rounded to the grid defined by the scale factor.
// The provided scale may be negative, to specify an exact grid size. The scale
// is then computed as the reciprocal.
func Geom_NewPrecisionModelWithScale(scale float64) *Geom_PrecisionModel {
	pm := &Geom_PrecisionModel{
		modelType: Geom_PrecisionModel_Fixed,
	}
	pm.setScale(scale)
	return pm
}

// Geom_NewPrecisionModelWithScaleAndOffsets creates a PrecisionModel that specifies
// Fixed precision. Fixed-precision coordinates are represented as precise
// internal coordinates, which are rounded to the grid defined by the scale
// factor.
//
// Deprecated: offsets are no longer supported, since internal representation is
// rounded floating point.
func Geom_NewPrecisionModelWithScaleAndOffsets(scale, offsetX, offsetY float64) *Geom_PrecisionModel {
	pm := &Geom_PrecisionModel{
		modelType: Geom_PrecisionModel_Fixed,
	}
	pm.setScale(scale)
	return pm
}

// Geom_NewPrecisionModelFromPrecisionModel creates a new PrecisionModel from an
// existing one.
func Geom_NewPrecisionModelFromPrecisionModel(pm *Geom_PrecisionModel) *Geom_PrecisionModel {
	return &Geom_PrecisionModel{
		modelType: pm.modelType,
		scale:     pm.scale,
		gridSize:  pm.gridSize,
	}
}

// IsFloating tests whether the precision model supports floating point.
func (pm *Geom_PrecisionModel) IsFloating() bool {
	return pm.modelType == Geom_PrecisionModel_Floating ||
		pm.modelType == Geom_PrecisionModel_FloatingSingle
}

// GetMaximumSignificantDigits returns the maximum number of significant digits
// provided by this precision model.
//
// Intended for use by routines which need to print out decimal representations
// of precise values (such as WKTWriter).
//
// This method would be more correctly called GetMinimumDecimalPlaces, since it
// actually computes the number of decimal places that is required to correctly
// display the full precision of an ordinate value.
//
// Since it is difficult to compute the required number of decimal places for
// scale factors which are not powers of 10, the algorithm uses a very rough
// approximation in this case. This has the side effect that for scale factors
// which are powers of 10 the value returned is 1 greater than the true value.
func (pm *Geom_PrecisionModel) GetMaximumSignificantDigits() int {
	maxSigDigits := 16
	if pm.modelType == Geom_PrecisionModel_Floating {
		maxSigDigits = 16
	} else if pm.modelType == Geom_PrecisionModel_FloatingSingle {
		maxSigDigits = 6
	} else if pm.modelType == Geom_PrecisionModel_Fixed {
		maxSigDigits = 1 + int(math.Ceil(math.Log10(pm.GetScale())))
	}
	return maxSigDigits
}

// GetScale returns the scale factor used to specify a fixed precision model.
// The number of decimal places of precision is equal to the base-10 logarithm
// of the scale factor. Non-integral and negative scale factors are supported.
// Negative scale factors indicate that the places of precision is to the left
// of the decimal point.
func (pm *Geom_PrecisionModel) GetScale() float64 {
	return pm.scale
}

// GridSize computes the grid size for a fixed precision model. This is equal to
// the reciprocal of the scale factor. If the grid size has been set explicitly
// (via a negative scale factor) it will be returned.
func (pm *Geom_PrecisionModel) GridSize() float64 {
	if pm.IsFloating() {
		return math.NaN()
	}
	if pm.gridSize != 0 {
		return pm.gridSize
	}
	return 1.0 / pm.scale
}

// GetType gets the type of this precision model.
func (pm *Geom_PrecisionModel) GetType() *Geom_PrecisionModelType {
	return pm.modelType
}

// setScale sets the multiplying factor used to obtain a precise coordinate.
// This method is private because PrecisionModel is an immutable (value) type.
func (pm *Geom_PrecisionModel) setScale(scale float64) {
	// A negative scale indicates the grid size is being set.
	// The scale is set as well, as the reciprocal.
	if scale < 0 {
		pm.gridSize = math.Abs(scale)
		pm.scale = 1.0 / pm.gridSize
	} else {
		pm.scale = math.Abs(scale)
		// Leave gridSize as 0, to ensure it is computed using scale.
		pm.gridSize = 0.0
	}
}

// GetOffsetX returns the x-offset used to obtain a precise coordinate.
//
// Deprecated: Offsets are no longer used.
func (pm *Geom_PrecisionModel) GetOffsetX() float64 {
	return 0
}

// GetOffsetY returns the y-offset used to obtain a precise coordinate.
//
// Deprecated: Offsets are no longer used.
func (pm *Geom_PrecisionModel) GetOffsetY() float64 {
	return 0
}

// ToInternalCoordinate sets internal to the precise representation of external.
//
// Deprecated: use MakePreciseCoordinate instead.
func (pm *Geom_PrecisionModel) ToInternalCoordinate(external, internal *Geom_Coordinate) {
	if pm.IsFloating() {
		internal.X = external.X
		internal.Y = external.Y
	} else {
		internal.X = pm.MakePrecise(external.X)
		internal.Y = pm.MakePrecise(external.Y)
	}
	internal.SetZ(external.GetZ())
}

// ToInternal returns the precise representation of external.
//
// Deprecated: use MakePreciseCoordinate instead.
func (pm *Geom_PrecisionModel) ToInternal(external *Geom_Coordinate) *Geom_Coordinate {
	internal := Geom_NewCoordinateFromCoordinate(external)
	pm.MakePreciseCoordinate(internal)
	return internal
}

// ToExternalNewCoordinate returns the external representation of internal.
//
// Deprecated: no longer needed, since internal representation is same as
// external representation.
func (pm *Geom_PrecisionModel) ToExternalNewCoordinate(internal *Geom_Coordinate) *Geom_Coordinate {
	return Geom_NewCoordinateFromCoordinate(internal)
}

// ToExternal sets external to the external representation of internal.
//
// Deprecated: no longer needed, since internal representation is same as
// external representation.
func (pm *Geom_PrecisionModel) ToExternal(internal, external *Geom_Coordinate) {
	external.X = internal.X
	external.Y = internal.Y
}

// MakePrecise rounds a numeric value to the PrecisionModel grid.
// Asymmetric Arithmetic Rounding is used, to provide uniform rounding behaviour
// no matter where the number is on the number line.
//
// This method has no effect on NaN values.
func (pm *Geom_PrecisionModel) MakePrecise(val float64) float64 {
	// Don't change NaN values.
	if math.IsNaN(val) {
		return val
	}
	if pm.modelType == Geom_PrecisionModel_FloatingSingle {
		floatSingleVal := float32(val)
		return float64(floatSingleVal)
	}
	if pm.modelType == Geom_PrecisionModel_Fixed {
		if pm.gridSize > 0 {
			return java.Round(val/pm.gridSize) * pm.gridSize
		}
		return java.Round(val*pm.scale) / pm.scale
	}
	// modelType == Floating - no rounding necessary.
	return val
}

// MakePreciseCoordinate rounds a Coordinate to the PrecisionModel grid.
func (pm *Geom_PrecisionModel) MakePreciseCoordinate(coord *Geom_Coordinate) {
	// Optimization for full precision.
	if pm.modelType == Geom_PrecisionModel_Floating {
		return
	}
	coord.X = pm.MakePrecise(coord.X)
	coord.Y = pm.MakePrecise(coord.Y)
	// MD says it's OK that we're not makePrecise'ing the z [Jon Aquino].
}

// String returns a string representation of this PrecisionModel.
func (pm *Geom_PrecisionModel) String() string {
	description := "UNKNOWN"
	if pm.modelType == Geom_PrecisionModel_Floating {
		description = "Floating"
	} else if pm.modelType == Geom_PrecisionModel_FloatingSingle {
		description = "Floating-Single"
	} else if pm.modelType == Geom_PrecisionModel_Fixed {
		description = fmt.Sprintf("Fixed (Scale=%v)", pm.GetScale())
	}
	return description
}

// Equals tests if this PrecisionModel equals another.
func (pm *Geom_PrecisionModel) Equals(other *Geom_PrecisionModel) bool {
	return pm.modelType == other.modelType && pm.scale == other.scale
}

// HashCode computes a hash code for this PrecisionModel.
func (pm *Geom_PrecisionModel) HashCode() int {
	prime := 31
	result := 1
	typeHash := 0
	if pm.modelType != nil {
		// Use a simple hash based on the type name.
		for _, c := range pm.modelType.name {
			typeHash = 31*typeHash + int(c)
		}
	}
	result = prime*result + typeHash
	temp := math.Float64bits(pm.scale)
	result = prime*result + int(temp^(temp>>32))
	return result
}

// CompareTo compares this PrecisionModel object with the specified object for
// order. A PrecisionModel is greater than another if it provides greater
// precision. The comparison is based on the value returned by
// GetMaximumSignificantDigits. This comparison is not strictly accurate when
// comparing floating precision models to fixed models; however, it is correct
// when both models are either floating or fixed.
func (pm *Geom_PrecisionModel) CompareTo(other *Geom_PrecisionModel) int {
	sigDigits := pm.GetMaximumSignificantDigits()
	otherSigDigits := other.GetMaximumSignificantDigits()
	if sigDigits < otherSigDigits {
		return -1
	}
	if sigDigits > otherSigDigits {
		return 1
	}
	return 0
}
