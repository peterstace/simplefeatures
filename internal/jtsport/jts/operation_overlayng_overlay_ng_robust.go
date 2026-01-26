package jts

import "math"

// OperationOverlayng_OverlayNGRobust performs an overlay operation using
// OverlayNG, providing full robustness by using a series of increasingly
// robust (but slower) noding strategies.
//
// The noding strategies used are:
//  1. A simple, fast noder using FLOATING precision.
//  2. A SnappingNoder using an automatically-determined snap tolerance
//  3. First snapping each geometry to itself, and then overlaying them using a
//     SnappingNoder.
//  4. The above two strategies are repeated with increasing snap tolerance, up
//     to a limit.
//  5. Finally a SnapRoundingNoder is used with an automatically-determined scale
//     factor intended to preserve input precision while still preventing
//     robustness problems.
//
// If all of the above attempts fail to compute a valid overlay, the original
// TopologyException is thrown. In practice this is extremely unlikely to occur.
//
// This algorithm relies on each overlay operation execution throwing a
// TopologyException if it is unable to compute the overlay correctly. Generally
// this occurs because the noding phase does not produce a valid noding. This
// requires the use of a ValidatingNoder in order to check the results of using
// a floating noder.

const (
	operationOverlayng_OverlayNGRobust_NUM_SNAP_TRIES = 5
	// A factor for a snapping tolerance distance which should allow noding to
	// be computed robustly.
	operationOverlayng_OverlayNGRobust_SNAP_TOL_FACTOR = 1e12
)

// OperationOverlayng_OverlayNGRobust_Union computes the unary union of a
// geometry using robust computation.
func OperationOverlayng_OverlayNGRobust_Union(geom *Geom_Geometry) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromGeometry(geom)
	op.SetUnionFunction(operationOverlayng_OverlayNGRobust_overlayUnion)
	return op.Union()
}

// OperationOverlayng_OverlayNGRobust_UnionCollection computes the unary union
// of a collection of geometries using robust computation.
func OperationOverlayng_OverlayNGRobust_UnionCollection(geoms []*Geom_Geometry) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromCollection(geoms)
	op.SetUnionFunction(operationOverlayng_OverlayNGRobust_overlayUnion)
	return op.Union()
}

// OperationOverlayng_OverlayNGRobust_UnionCollectionWithFactory computes the
// unary union of a collection of geometries using robust computation.
func OperationOverlayng_OverlayNGRobust_UnionCollectionWithFactory(geoms []*Geom_Geometry, geomFact *Geom_GeometryFactory) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromCollectionWithFactory(geoms, geomFact)
	op.SetUnionFunction(operationOverlayng_OverlayNGRobust_overlayUnion)
	return op.Union()
}

// overlayUnion is the union strategy used by OverlayNGRobust.
var operationOverlayng_OverlayNGRobust_overlayUnion = &overlayNGRobustUnionStrategy{}

type overlayNGRobustUnionStrategy struct{}

func (s *overlayNGRobustUnionStrategy) Union(g0, g1 *Geom_Geometry) *Geom_Geometry {
	return OperationOverlayng_OverlayNGRobust_Overlay(g0, g1, OperationOverlayng_OverlayNG_UNION)
}

func (s *overlayNGRobustUnionStrategy) IsFloatingPrecision() bool {
	return true
}

// OperationOverlayng_OverlayNGRobust_Overlay overlays two geometries, using
// heuristics to ensure computation completes correctly. In practice the
// heuristics are observed to be fully correct.
func OperationOverlayng_OverlayNGRobust_Overlay(geom0, geom1 *Geom_Geometry, opCode int) *Geom_Geometry {
	var result *Geom_Geometry
	var exOriginal any

	// First try overlay with a FLOAT noder, which is fast and causes least
	// change to geometry coordinates. By default the noder is validated, which
	// is required in order to detect certain invalid noding situations which
	// otherwise cause incorrect overlay output.
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Capture original exception, so it can be rethrown if the
				// remaining strategies all fail.
				exOriginal = r
			}
		}()
		result = OperationOverlayng_OverlayNG_OverlayDefault(geom0, geom1, opCode)
	}()

	if result != nil {
		return result
	}

	// On failure retry using snapping noding with a "safe" tolerance.
	// if this throws an exception just let it go, since it is something that
	// is not a TopologyException.
	result = operationOverlayng_OverlayNGRobust_overlaySnapTries(geom0, geom1, opCode)
	if result != nil {
		return result
	}

	// On failure retry using snap-rounding with a heuristic scale factor (grid size).
	result = operationOverlayng_OverlayNGRobust_overlaySR(geom0, geom1, opCode)
	if result != nil {
		return result
	}

	// Just can't get overlay to work, so throw original error.
	panic(exOriginal)
}

// overlaySnapTries attempts overlay using snapping with repeated tries with
// increasing snap tolerances.
func operationOverlayng_OverlayNGRobust_overlaySnapTries(geom0, geom1 *Geom_Geometry, opCode int) *Geom_Geometry {
	snapTol := operationOverlayng_OverlayNGRobust_snapToleranceFor2(geom0, geom1)

	for i := 0; i < operationOverlayng_OverlayNGRobust_NUM_SNAP_TRIES; i++ {
		result := operationOverlayng_OverlayNGRobust_overlaySnapping(geom0, geom1, opCode, snapTol)
		if result != nil {
			return result
		}

		// Now try snapping each input individually, and then doing the overlay.
		result = operationOverlayng_OverlayNGRobust_overlaySnapBoth(geom0, geom1, opCode, snapTol)
		if result != nil {
			return result
		}

		// Increase the snap tolerance and try again.
		snapTol = snapTol * 10
	}
	// Failed to compute overlay.
	return nil
}

// overlaySnapping attempts overlay using a SnappingNoder.
func operationOverlayng_OverlayNGRobust_overlaySnapping(geom0, geom1 *Geom_Geometry, opCode int, snapTol float64) *Geom_Geometry {
	var result *Geom_Geometry
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Ignore exception, return nil result to indicate failure.
			}
		}()
		result = operationOverlayng_OverlayNGRobust_overlaySnapTol(geom0, geom1, opCode, snapTol)
	}()
	return result
}

// overlaySnapBoth attempts overlay with first snapping each geometry individually.
func operationOverlayng_OverlayNGRobust_overlaySnapBoth(geom0, geom1 *Geom_Geometry, opCode int, snapTol float64) *Geom_Geometry {
	var result *Geom_Geometry
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Ignore exception, return nil result to indicate failure.
			}
		}()
		snap0 := operationOverlayng_OverlayNGRobust_snapSelf(geom0, snapTol)
		snap1 := operationOverlayng_OverlayNGRobust_snapSelf(geom1, snapTol)
		result = operationOverlayng_OverlayNGRobust_overlaySnapTol(snap0, snap1, opCode, snapTol)
	}()
	return result
}

// snapSelf self-snaps a geometry by running a union operation with it as the
// only input. This helps to remove narrow spike/gore artifacts to simplify the
// geometry, which improves robustness. Collapsed artifacts are removed from the
// result to allow using it in further overlay operations.
func operationOverlayng_OverlayNGRobust_snapSelf(geom *Geom_Geometry, snapTol float64) *Geom_Geometry {
	ov := OperationOverlayng_NewOverlayNGWithPM(geom, nil, nil, OperationOverlayng_OverlayNG_UNION)
	snapNoder := NodingSnap_NewSnappingNoder(snapTol)
	ov.SetNoder(snapNoder)
	// Ensure the result is not mixed-dimension, since it will be used in
	// further overlay computation. It may however be lower dimension, if it
	// collapses completely due to snapping.
	ov.SetStrictMode(true)
	return ov.GetResult()
}

func operationOverlayng_OverlayNGRobust_overlaySnapTol(geom0, geom1 *Geom_Geometry, opCode int, snapTol float64) *Geom_Geometry {
	snapNoder := NodingSnap_NewSnappingNoder(snapTol)
	return OperationOverlayng_OverlayNG_OverlayWithNoderOnly(geom0, geom1, opCode, snapNoder)
}

// snapToleranceFor2 computes a heuristic snap tolerance distance for overlaying
// a pair of geometries using a SnappingNoder.
func operationOverlayng_OverlayNGRobust_snapToleranceFor2(geom0, geom1 *Geom_Geometry) float64 {
	tol0 := operationOverlayng_OverlayNGRobust_snapTolerance(geom0)
	tol1 := operationOverlayng_OverlayNGRobust_snapTolerance(geom1)
	return math.Max(tol0, tol1)
}

func operationOverlayng_OverlayNGRobust_snapTolerance(geom *Geom_Geometry) float64 {
	magnitude := operationOverlayng_OverlayNGRobust_ordinateMagnitude(geom)
	return magnitude / operationOverlayng_OverlayNGRobust_SNAP_TOL_FACTOR
}

// ordinateMagnitude computes the largest magnitude of the ordinates of a
// geometry, based on the geometry envelope.
func operationOverlayng_OverlayNGRobust_ordinateMagnitude(geom *Geom_Geometry) float64 {
	if geom == nil || geom.IsEmpty() {
		return 0
	}
	env := geom.GetEnvelopeInternal()
	magMax := math.Max(
		math.Abs(env.GetMaxX()), math.Abs(env.GetMaxY()))
	magMin := math.Max(
		math.Abs(env.GetMinX()), math.Abs(env.GetMinY()))
	return math.Max(magMax, magMin)
}

// overlaySR attempts Overlay using Snap-Rounding with an
// automatically-determined scale factor.
func operationOverlayng_OverlayNGRobust_overlaySR(geom0, geom1 *Geom_Geometry, opCode int) *Geom_Geometry {
	var result *Geom_Geometry
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Ignore exception, return nil result to indicate failure.
			}
		}()
		scaleSafe := OperationOverlayng_PrecisionUtil_SafeScaleGeoms(geom0, geom1)
		pmSafe := Geom_NewPrecisionModelWithScale(scaleSafe)
		result = OperationOverlayng_OverlayNG_Overlay(geom0, geom1, opCode, pmSafe)
	}()
	return result
}
