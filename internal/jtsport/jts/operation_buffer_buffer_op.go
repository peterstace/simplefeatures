package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationBuffer_BufferOp_CAP_ROUND specifies a round line buffer end cap style.
// Deprecated: use OperationBuffer_BufferParameters
const OperationBuffer_BufferOp_CAP_ROUND = OperationBuffer_BufferParameters_CAP_ROUND

// OperationBuffer_BufferOp_CAP_BUTT specifies a butt (or flat) line buffer end cap style.
// Deprecated: use OperationBuffer_BufferParameters
const OperationBuffer_BufferOp_CAP_BUTT = OperationBuffer_BufferParameters_CAP_FLAT

// OperationBuffer_BufferOp_CAP_FLAT specifies a butt (or flat) line buffer end cap style.
// Deprecated: use OperationBuffer_BufferParameters
const OperationBuffer_BufferOp_CAP_FLAT = OperationBuffer_BufferParameters_CAP_FLAT

// OperationBuffer_BufferOp_CAP_SQUARE specifies a square line buffer end cap style.
// Deprecated: use OperationBuffer_BufferParameters
const OperationBuffer_BufferOp_CAP_SQUARE = OperationBuffer_BufferParameters_CAP_SQUARE

// operationBuffer_bufferOp_MAX_PRECISION_DIGITS is a number of digits of precision which leaves some computational "headroom"
// for floating point operations.
//
// This value should be less than the decimal precision of double-precision values (16).
const operationBuffer_bufferOp_MAX_PRECISION_DIGITS = 12

// OperationBuffer_BufferOp computes the buffer of a geometry, for both positive and negative buffer distances.
//
// In GIS, the positive (or negative) buffer of a geometry is defined as
// the Minkowski sum (or difference) of the geometry
// with a circle of radius equal to the absolute value of the buffer distance.
// In the CAD/CAM world buffers are known as offset curves.
// In morphological analysis the operation of positive and negative buffering
// is referred to as erosion and dilation.
//
// The buffer operation always returns a polygonal result.
// The negative or zero-distance buffer of lines and points is always an empty Polygon.
//
// Since true buffer curves may contain circular arcs,
// computed buffer polygons are only approximations to the true geometry.
// The user can control the accuracy of the approximation by specifying
// the number of linear segments used to approximate arcs.
//
// Buffer results are always valid geometry.
// Given this, computing a zero-width buffer of an invalid polygonal geometry is
// an effective way to "validify" the geometry.
// Note however that in the case of self-intersecting "bow-tie" geometries,
// only the largest enclosed area will be retained.
type OperationBuffer_BufferOp struct {
	argGeom       *Geom_Geometry
	distance      float64
	bufParams     *OperationBuffer_BufferParameters
	resultGeom    *Geom_Geometry
	saveException error

	isInvertOrientation bool
}

// operationBuffer_bufferOp_precisionScaleFactor computes a scale factor to limit the precision of
// a given combination of Geometry and buffer distance.
// The scale factor is determined by
// the number of digits of precision in the (geometry + buffer distance),
// limited by the supplied maxPrecisionDigits value.
//
// The scale factor is based on the absolute magnitude of the (geometry + buffer distance).
// since this determines the number of digits of precision which must be handled.
func operationBuffer_bufferOp_precisionScaleFactor(g *Geom_Geometry, distance float64, maxPrecisionDigits int) float64 {
	env := g.GetEnvelopeInternal()
	envMax := Math_MathUtil_Max4(
		math.Abs(env.GetMaxX()),
		math.Abs(env.GetMaxY()),
		math.Abs(env.GetMinX()),
		math.Abs(env.GetMinY()))

	expandByDistance := 0.0
	if distance > 0.0 {
		expandByDistance = distance
	}
	bufEnvMax := envMax + 2*expandByDistance

	// the smallest power of 10 greater than the buffer envelope
	bufEnvPrecisionDigits := int(math.Log10(bufEnvMax) + 1.0)
	minUnitLog10 := maxPrecisionDigits - bufEnvPrecisionDigits

	scaleFactor := math.Pow(10.0, float64(minUnitLog10))
	return scaleFactor
}

// OperationBuffer_BufferOp_BufferOp computes the buffer of a geometry for a given buffer distance.
func OperationBuffer_BufferOp_BufferOp(g *Geom_Geometry, distance float64) *Geom_Geometry {
	gBuf := OperationBuffer_NewBufferOp(g)
	geomBuf := gBuf.GetResultGeometry(distance)
	return geomBuf
}

// OperationBuffer_BufferOp_BufferOpWithParams computes the buffer for a geometry for a given buffer distance
// and accuracy of approximation.
func OperationBuffer_BufferOp_BufferOpWithParams(g *Geom_Geometry, distance float64, params *OperationBuffer_BufferParameters) *Geom_Geometry {
	bufOp := OperationBuffer_NewBufferOpWithParams(g, params)
	geomBuf := bufOp.GetResultGeometry(distance)
	return geomBuf
}

// OperationBuffer_BufferOp_BufferOpWithQuadrantSegments computes the buffer for a geometry for a given buffer distance
// and accuracy of approximation.
func OperationBuffer_BufferOp_BufferOpWithQuadrantSegments(g *Geom_Geometry, distance float64, quadrantSegments int) *Geom_Geometry {
	bufOp := OperationBuffer_NewBufferOp(g)
	bufOp.SetQuadrantSegments(quadrantSegments)
	geomBuf := bufOp.GetResultGeometry(distance)
	return geomBuf
}

// OperationBuffer_BufferOp_BufferOpWithQuadrantSegmentsAndEndCapStyle computes the buffer for a geometry for a given buffer distance
// and accuracy of approximation.
func OperationBuffer_BufferOp_BufferOpWithQuadrantSegmentsAndEndCapStyle(g *Geom_Geometry, distance float64, quadrantSegments int, endCapStyle int) *Geom_Geometry {
	bufOp := OperationBuffer_NewBufferOp(g)
	bufOp.SetQuadrantSegments(quadrantSegments)
	bufOp.SetEndCapStyle(endCapStyle)
	geomBuf := bufOp.GetResultGeometry(distance)
	return geomBuf
}

// OperationBuffer_BufferOp_BufferByZero buffers a geometry with distance zero.
// The result can be computed using the maximum-signed-area orientation,
// or by combining both orientations.
//
// This can be used to fix an invalid polygonal geometry to be valid
// (i.e. with no self-intersections).
// For some uses (e.g. fixing the result of a simplification)
// a better result is produced by using only the max-area orientation.
// Other uses (e.g. fixing geometry) require both orientations to be used.
//
// This function is for INTERNAL use only.
func OperationBuffer_BufferOp_BufferByZero(geom *Geom_Geometry, isBothOrientations bool) *Geom_Geometry {
	// compute buffer using maximum signed-area orientation
	buf0 := geom.Buffer(0)
	if !isBothOrientations {
		return buf0
	}

	// compute buffer using minimum signed-area orientation
	op := OperationBuffer_NewBufferOp(geom)
	op.isInvertOrientation = true
	buf0Inv := op.GetResultGeometry(0)

	// the buffer results should be non-adjacent, so combining is safe
	return operationBuffer_bufferOp_combine(buf0, buf0Inv)
}

// operationBuffer_bufferOp_combine combines the elements of two polygonal geometries together.
// The input geometries must be non-adjacent, to avoid creating an invalid result.
func operationBuffer_bufferOp_combine(poly0, poly1 *Geom_Geometry) *Geom_Geometry {
	// short-circuit - handles case where geometry is valid
	if poly1.IsEmpty() {
		return poly0
	}
	if poly0.IsEmpty() {
		return poly1
	}

	polys := make([]*Geom_Polygon, 0)
	operationBuffer_bufferOp_extractPolygons(poly0, &polys)
	operationBuffer_bufferOp_extractPolygons(poly1, &polys)
	if len(polys) == 1 {
		return polys[0].Geom_Geometry
	}
	return poly0.GetFactory().CreateMultiPolygonFromPolygons(polys).Geom_Geometry
}

func operationBuffer_bufferOp_extractPolygons(poly0 *Geom_Geometry, polys *[]*Geom_Polygon) {
	for i := 0; i < poly0.GetNumGeometries(); i++ {
		p := poly0.GetGeometryN(i)
		*polys = append(*polys, java.Cast[*Geom_Polygon](p))
	}
}

// OperationBuffer_NewBufferOp initializes a buffer computation for the given geometry.
func OperationBuffer_NewBufferOp(g *Geom_Geometry) *OperationBuffer_BufferOp {
	return &OperationBuffer_BufferOp{
		argGeom:   g,
		bufParams: OperationBuffer_NewBufferParameters(),
	}
}

// OperationBuffer_NewBufferOpWithParams initializes a buffer computation for the given geometry
// with the given set of parameters.
func OperationBuffer_NewBufferOpWithParams(g *Geom_Geometry, bufParams *OperationBuffer_BufferParameters) *OperationBuffer_BufferOp {
	return &OperationBuffer_BufferOp{
		argGeom:   g,
		bufParams: bufParams,
	}
}

// SetEndCapStyle specifies the end cap style of the generated buffer.
// The styles supported are CAP_ROUND, CAP_FLAT, and CAP_SQUARE.
// The default is CAP_ROUND.
func (bo *OperationBuffer_BufferOp) SetEndCapStyle(endCapStyle int) {
	bo.bufParams.SetEndCapStyle(endCapStyle)
}

// SetQuadrantSegments sets the number of line segments in a quarter-circle
// used to approximate angle fillets for round end caps and joins.
func (bo *OperationBuffer_BufferOp) SetQuadrantSegments(quadrantSegments int) {
	bo.bufParams.SetQuadrantSegments(quadrantSegments)
}

// GetResultGeometry returns the buffer computed for a geometry for a given buffer distance.
func (bo *OperationBuffer_BufferOp) GetResultGeometry(distance float64) *Geom_Geometry {
	bo.distance = distance
	bo.computeGeometry()
	return bo.resultGeom
}

func (bo *OperationBuffer_BufferOp) computeGeometry() {
	bo.bufferOriginalPrecision()
	if bo.resultGeom != nil {
		return
	}

	argPM := bo.argGeom.GetFactory().GetPrecisionModel()
	if argPM.GetType() == Geom_PrecisionModel_Fixed {
		bo.bufferFixedPrecision(argPM)
	} else {
		bo.bufferReducedPrecision()
	}
}

func (bo *OperationBuffer_BufferOp) bufferReducedPrecision() {
	// try and compute with decreasing precision
	for precDigits := operationBuffer_bufferOp_MAX_PRECISION_DIGITS; precDigits >= 0; precDigits-- {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// update the saved exception to reflect the new input geometry
					if err, ok := r.(error); ok {
						bo.saveException = err
					}
					// don't propagate the exception - it will be detected by fact that resultGeometry is null
				}
			}()
			bo.bufferReducedPrecisionDigits(precDigits)
		}()
		if bo.resultGeom != nil {
			return
		}
	}

	// tried everything - have to bail
	panic(bo.saveException)
}

func (bo *OperationBuffer_BufferOp) bufferReducedPrecisionDigits(precisionDigits int) {
	sizeBasedScaleFactor := operationBuffer_bufferOp_precisionScaleFactor(bo.argGeom, bo.distance, precisionDigits)

	fixedPM := Geom_NewPrecisionModelWithScale(sizeBasedScaleFactor)
	bo.bufferFixedPrecision(fixedPM)
}

func (bo *OperationBuffer_BufferOp) bufferOriginalPrecision() {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				bo.saveException = err
			}
			// don't propagate the exception - it will be detected by fact that resultGeometry is null
		}
	}()
	// use fast noding by default
	bufBuilder := bo.createBufferBuilder()
	bo.resultGeom = bufBuilder.Buffer(bo.argGeom, bo.distance)
}

func (bo *OperationBuffer_BufferOp) createBufferBuilder() *operationBuffer_BufferBuilder {
	bufBuilder := operationBuffer_newBufferBuilder(bo.bufParams)
	bufBuilder.SetInvertOrientation(bo.isInvertOrientation)
	return bufBuilder
}

func (bo *OperationBuffer_BufferOp) bufferFixedPrecision(fixedPM *Geom_PrecisionModel) {
	// Snap-Rounding provides both robustness and a fixed output precision.
	//
	// SnapRoundingNoder does not require rounded input,
	// so could be used by itself.
	// But using ScaledNoder may be faster, since it avoids
	// rounding within SnapRoundingNoder.
	// (Note this only works for buffering, because
	// ScaledNoder may invalidate topology.)
	snapNoder := NodingSnapround_NewSnapRoundingNoder(Geom_NewPrecisionModelWithScale(1.0))
	noder := Noding_NewScaledNoder(snapNoder, fixedPM.GetScale())

	bufBuilder := bo.createBufferBuilder()
	bufBuilder.SetWorkingPrecisionModel(fixedPM)
	bufBuilder.SetNoder(noder)
	// this may throw an exception, if robustness errors are encountered
	bo.resultGeom = bufBuilder.Buffer(bo.argGeom, bo.distance)
}
