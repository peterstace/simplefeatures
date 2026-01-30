package jts

// OperationOverlayng_PrecisionReducer provides functions to reduce the
// precision of a geometry by rounding it to a given precision model.
//
// This class handles only polygonal and linear inputs.

// OperationOverlayng_PrecisionReducer_ReducePrecision reduces the precision of
// a geometry by rounding and snapping it to the supplied PrecisionModel. The
// input geometry must be polygonal or linear.
//
// The output is always a valid geometry. This implies that input components may
// be merged if they are closer than the grid precision. If merging is not
// desired, then the individual geometry components should be processed
// separately.
//
// The output is fully noded (i.e. coincident lines are merged and noded). This
// provides an effective way to node / snap-round a collection of LineStrings.
//
// Panics with an error if the reduction fails due to invalid input geometry.
func OperationOverlayng_PrecisionReducer_ReducePrecision(geom *Geom_Geometry, pm *Geom_PrecisionModel) *Geom_Geometry {
	ov := OperationOverlayng_NewOverlayNGUnary(geom, pm)
	// Ensure reducing an area only produces polygonal result.
	// (I.e. collapse lines are not output.)
	if geom.GetDimension() == 2 {
		ov.SetAreaResultOnly(true)
	}
	var reduced *Geom_Geometry
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Reduction failed, possible invalid input")
			}
		}()
		reduced = ov.GetResult()
	}()
	return reduced
}
