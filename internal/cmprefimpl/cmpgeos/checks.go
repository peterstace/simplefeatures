package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/geos"
)

func unaryChecks(h *Handle, g geom.Geometry, log *log.Logger) error {
	if valid, err := checkIsValid(h, g, log); err != nil {
		return err
	} else if !valid {
		return nil
	}

	log.Println("checking AsBinary")
	if err := checkAsBinary(h, g, log); err != nil {
		return err
	}
	log.Println("checking FromBinary")
	if err := checkFromBinary(h, g, log); err != nil {
		return err
	}
	log.Println("checking AsText")
	if err := checkAsText(h, g, log); err != nil {
		return err
	}
	log.Println("checking FromText")
	if err := checkFromText(h, g, log); err != nil {
		return err
	}
	log.Println("checking IsEmpty")
	if err := checkIsEmpty(h, g, log); err != nil {
		return err
	}
	log.Println("checking Dimension")
	if err := checkDimension(h, g, log); err != nil {
		return err
	}
	log.Println("checking Envelope")
	if err := checkEnvelope(h, g, log); err != nil {
		return err
	}
	log.Println("checking IsSimple")
	if err := checkIsSimple(h, g, log); err != nil {
		return err
	}
	log.Println("checking Boundary")
	if err := checkBoundary(h, g, log); err != nil {
		return err
	}
	log.Println("checking ConvexHull")
	if err := checkConvexHull(h, g, log); err != nil {
		return err
	}
	log.Println("checking IsRing")
	if err := checkIsRing(h, g, log); err != nil {
		return err
	}
	log.Println("checking Length")
	if err := checkLength(h, g, log); err != nil {
		return err
	}
	log.Println("checking Area")
	if err := checkArea(h, g, log); err != nil {
		return err
	}
	log.Println("checking Centroid")
	if err := checkCentroid(h, g, log); err != nil {
		return err
	}
	log.Println("checking PointOnSurface")
	if err := checkPointOnSurface(h, g, log); err != nil {
		return err
	}
	log.Println("checking Simplify")
	if err := checkSimplify(h, g, log); err != nil {
		return err
	}

	return nil

	// TODO: Reverse isn't checked yet. There is some significant behaviour
	// differences between libgeos and PostGIS.
}

var errMismatch = errors.New("mismatch")

func checkIsValid(h *Handle, g geom.Geometry, log *log.Logger) (bool, error) {
	wkb := g.AsBinary()
	var validAsPerSimpleFeatures bool
	if _, err := geom.UnmarshalWKB(wkb); err == nil {
		validAsPerSimpleFeatures = true
	}
	log.Printf("Valid as per simplefeatures: %v", validAsPerSimpleFeatures)

	validAsPerLibgeos, err := h.isValid(g)
	if err != nil {
		// The geometry is _so_ invalid that libgeos can't even tell if it's
		// invalid or not.
		validAsPerLibgeos = false
	}
	log.Printf("Valid as per libgeos: %v", validAsPerLibgeos)

	// libgeos allows empty rings in Polygons, however simplefeatures doesn't
	// (it follows the PostGIS behaviour of disallowing empty rings).
	ignoreMismatch := hasEmptyRing(g)

	if !ignoreMismatch && validAsPerLibgeos != validAsPerSimpleFeatures {
		return false, errMismatch
	}
	return validAsPerSimpleFeatures, nil
}

func checkAsText(h *Handle, g geom.Geometry, log *log.Logger) error {
	// Skip any geometries that have a non-empty Point within a MultiPoint.
	// Libgeos erroneously produces WKT with missing parenthesis around each
	// non-empty point.
	if containsNonEmptyPointInMultiPoint(g) {
		return nil
	}

	// Skip any geometries that are collections or contain collections that
	// only contain empty geometries. Libgeos will render WKT for these
	// collections as being EMPTY, however this isn't correct behaviour.
	if containsCollectionWithOnlyEmptyElements(g) {
		return nil
	}

	// Skip geometries that GEOS is known to produce incorrect WKT for due to numerical rounding issues.
	if map[string]bool{
		"POLYGON((0.9292893218813453 1.0707106781186548,1 1.1414213562373097,1.1414213562373097 1,0.07071067811865475 -0.07071067811865475,0 -0.1414213562373095,-0.1414213562373095 -0.000000000000000013877787807814457,0.9292893218813453 1.0707106781186548))": true,
	}[g.AsText()] {
		return nil
	}

	want, err := h.asText(g)
	if err != nil {
		return err
	}

	// Account for easy-to-adjust for acceptable spacing differences between
	// libgeos and simplefeatures.
	want = strings.ReplaceAll(want, " (", "(")
	want = strings.ReplaceAll(want, ", ", ",")

	got := g.AsText()

	if err := wktsEqual(got, want); err != nil {
		log.Printf("WKTs not equal: %v", err)
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}

func wktsEqual(wktA, wktB string) error {
	toksA := tokenizeWKT(wktA)
	toksB := tokenizeWKT(wktB)
	if len(toksA) != len(toksB) {
		return fmt.Errorf(
			"token lengths differ: %d vs %d",
			len(toksA), len(toksB),
		)
	}
	for i, tokA := range toksA {
		tokB := toksB[i]
		fA, errA := strconv.ParseFloat(tokA, 64)
		fB, errB := strconv.ParseFloat(tokA, 64)
		var eq bool
		if errA == nil && errB == nil {
			// If this check gives false negatives (e.g. libgeos and
			// simplefeatures may use slightly different precision), then we
			// can always check a relative difference here instead of a strict
			// ==.
			eq = fA == fB
		} else {
			eq = tokA == tokB
		}
		if !eq {
			return fmt.Errorf(
				"tokens at position %d differ: %s vs %s",
				i, tokA, tokB,
			)
		}
	}
	return nil
}

func checkFromText(h *Handle, g geom.Geometry, log *log.Logger) error {
	// libgeos is unable to parse MultiPoints if the *first* Point is empty. It
	// gives the following error: ParseException: Unexpected token: WORD EMPTY.
	// Skip the check in that case.
	if g.IsMultiPoint() &&
		g.MustAsMultiPoint().NumPoints() > 0 &&
		g.MustAsMultiPoint().PointN(0).IsEmpty() {
		return nil
	}

	wkt := g.AsText()
	want, err := h.fromText(wkt)
	if err != nil {
		return err
	}

	got, err := geom.UnmarshalWKT(wkt)
	if err != nil {
		return err
	}

	if !geom.ExactEquals(got, want) {
		log.Printf("want: %v", want.AsText())
		log.Printf("got:  %v", got.AsText())
		return errMismatch
	}
	return nil
}

func checkAsBinary(h *Handle, g geom.Geometry, log *log.Logger) error {
	var wantDefined bool
	want, err := h.asBinary(g)
	if err == nil {
		wantDefined = true
	}
	hasPointEmpty := hasEmptyPoint(g)
	if !wantDefined && !hasPointEmpty {
		return errors.New("AsBinary wasn't defined by libgeos and the test is " +
			"NOT for a geometry containing a POINT EMPTY, which is unexpected",
		)
	}
	if !wantDefined {
		// Skip the test, since we don't have a WKB from libgeos to compare to.
		// This is only for the POINT EMPTY case. Simplefeatures _does_ produce
		// a WKB for POINT EMPTY although this is strictly an extension to the
		// spec.
		return nil
	}

	// GEOS uses a slightly different NaN representation (both are equally valid).
	want = bytes.ReplaceAll(want,
		[]byte{0x00, 0, 0, 0, 0, 0, 0xf8, 0x7f},
		[]byte{0x01, 0, 0, 0, 0, 0, 0xf8, 0x7f},
	)

	got := g.AsBinary()
	if !bytes.Equal(want, got) {
		log.Printf("want:\n%s", hex.Dump(want))
		log.Printf("got:\n%s", hex.Dump(got))
		return errMismatch
	}
	return nil
}

func checkFromBinary(h *Handle, g geom.Geometry, log *log.Logger) error {
	if containsMultiPolygonWithEmptyPolygon(g) {
		// libgeos omits the empty Polygon, but simplefeatures doesn't.
		return nil
	}

	wkb := g.AsBinary()

	// Skip any MultiPoints that contain empty Points. Libgeos seems has
	// trouble handling these.
	if g.IsMultiPoint() {
		mp := g.MustAsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			if mp.PointN(i).IsEmpty() {
				return nil
			}
		}
	}

	want, err := h.fromBinary(wkb)
	if err != nil {
		return err
	}

	got, err := geom.UnmarshalWKB(wkb)
	if err != nil {
		return err
	}

	if !geom.ExactEquals(want, got) {
		return errMismatch
	}
	return nil
}

func checkIsEmpty(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.isEmpty(g)
	if err != nil {
		return err
	}
	got := g.IsEmpty()

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got: %v", got)
		return errMismatch
	}
	return nil
}

func checkDimension(h *Handle, g geom.Geometry, log *log.Logger) error {
	var want int
	if !containsOnlyGeometryCollections(g) {
		// Libgeos gives -1 dimension for GeometryCollection trees that only
		// contain other GeometryCollections (all the way to the leaf nodes).
		// This is weird behaviour, and the dimension should actually be zero.
		// So we don't get 'want' from libgeos in that case (and allow want to
		// default to 0).
		var err error
		want, err = h.dimension(g)
		if err != nil {
			return err
		}
	}
	got := g.Dimension()

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got: %v", got)
		return errMismatch
	}
	return nil
}

func checkEnvelope(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.envelope(g)
	if err != nil {
		return err
	}
	got := g.Envelope()

	if want.Min() != got.Min() || want.Max() != got.Max() {
		log.Printf("want: %v", want.AsGeometry().AsText())
		log.Printf("got:  %v", got.AsGeometry().AsText())
		return errMismatch
	}
	return nil
}

func checkIsSimple(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, wantDefined, err := h.isSimple(g)
	if err != nil {
		if err == errLibgeosCrash {
			// Skip any tests that would cause libgeos to crash.
			return nil
		}
		return err
	}
	got, gotDefined := g.IsSimple()

	if wantDefined != gotDefined {
		log.Printf("want defined: %v", wantDefined)
		log.Printf("got defined: %v", gotDefined)
		return errMismatch
	}
	if !gotDefined {
		return nil
	}

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}

func checkBoundary(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, wantDefined, err := h.boundary(g)
	if err != nil {
		return err
	}

	if !wantDefined && !g.IsGeometryCollection() {
		return errors.New("boundary not defined by libgeos, but " +
			"input is not a geometry collection (this is unexpected)")
	}
	if !wantDefined {
		return nil
	}

	got := g.Boundary()

	// There are some slight differences in the behaviour for empty inputs, so
	// we don't check these cases (so long as the output is also empty).
	if got.IsEmpty() && want.IsEmpty() {
		return nil
	}

	if !geom.ExactEquals(want, got, geom.IgnoreOrder) {
		log.Printf("want: %v", want.AsText())
		log.Printf("got:  %v", got.AsText())
		return errMismatch
	}
	return nil
}

func checkConvexHull(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.convexHull(g)
	if err != nil {
		return err
	}
	got := g.ConvexHull()

	// libgeos and PostGIS have slightly different behaviour when the result is
	// empty (different geometry types). Simplefeatures matches PostGIS
	// behaviour right now.
	if got.IsEmpty() && want.IsEmpty() {
		return nil
	}

	if !geom.ExactEquals(want, got, geom.IgnoreOrder) {
		log.Printf("want: %v", want.AsText())
		log.Printf("got:  %v", got.AsText())
		return errMismatch
	}
	return nil
}

func checkIsRing(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.isRing(g)
	if err != nil {
		return err
	}
	got := g.IsLineString() && g.MustAsLineString().IsRing()

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}

func checkLength(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.length(g)
	if err != nil {
		return err
	}
	got := g.Length()

	// libgeos and PostGIS disagree on the definition of length for areal
	// geometries.  PostGIS always gives zero, while libgeos gives the length
	// of the boundary. Simplefeatures follows the PostGIS behaviour.
	if isArealGeometry(g) {
		return nil
	}

	if math.Abs(want-got) > 1e-6 {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}

func isArealGeometry(g geom.Geometry) bool {
	switch {
	case g.IsPolygon() || g.IsMultiPolygon():
		return true
	case g.IsGeometryCollection():
		gc := g.MustAsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if isArealGeometry(gc.GeometryN(i)) {
				return true
			}
		}
	}
	return false
}

func checkArea(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.area(g)
	if err != nil {
		return err
	}
	got := g.Area()

	if math.Abs(want-got) > 1e-6 {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}

func checkCentroid(h *Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.centroid(g)
	if err != nil {
		return err
	}
	got := g.Centroid().AsGeometry()

	if !geom.ExactEquals(want, got, geom.ToleranceXY(1e-9)) {
		log.Printf("want: %v", want.AsText())
		log.Printf("got:  %v", got.AsText())
		return errMismatch
	}
	return nil
}

func checkPointOnSurface(h *Handle, g geom.Geometry, log *log.Logger) error {
	// It's too difficult to perform a direct comparison against GEOS's
	// PointOnSurface, due to numeric stability related issue. This is because
	// there are floating point comparisons to find the "best" point. However,
	// sometimes there are many points that are equally best. Floating point
	// issues mean that it's hard to get the implementations to line up
	// precisely in all cases (and there is no objectively best way to do it).
	// Instead, we check invariants on the result.

	pt := g.PointOnSurface().AsGeometry()

	if pt.IsEmpty() != g.IsEmpty() {
		log.Printf("The geometry's empty status doesn't match the point's empty status")
		log.Printf("g empty:  %v", g.IsEmpty())
		log.Printf("pt empty: %v", pt.IsEmpty())
		return errMismatch
	}

	if !g.IsEmpty() && !g.IsGeometryCollection() {
		intersects, err := geos.Intersects(pt, g)
		if err != nil {
			return err
		}
		if !intersects {
			log.Printf("the pt doesn't intersect with the input")
			return errMismatch
		}
	}

	if g.Dimension() == 2 && !g.IsEmpty() && !g.IsGeometryCollection() {
		contains, err := geos.Contains(g, pt)
		if err != nil {
			return err
		}
		if !contains {
			log.Printf("the input doesn't contain the pt")
			return errMismatch
		}
	}

	return nil
}

func checkSimplify(h *Handle, g geom.Geometry, log *log.Logger) error {
	for _, threshold := range []float64{0.125, 0.25, 0.5, 1, 2, 4, 8, 16} {
		// If we get an error from GEOS, then we may or may not get an error from
		// simplefeatures.
		want, err := h.simplify(g, threshold)
		wantIsValid := err == nil

		// Even if GEOS couldn't simplify, we still want to attempt to simplify
		// with simplefeatures to ensure it doesn't crash (even if it may give an
		// error).
		got, err := geos.Simplify(g, threshold)
		gotIsValid := err == nil

		if wantIsValid && !gotIsValid {
			return fmt.Errorf("GEOS could simplify but simplefeatures could not: %w", err)
		}

		if gotIsValid && wantIsValid && !geom.ExactEquals(got, want) {
			log.Printf("Simplify results not equal for threshold=%v", threshold)
			log.Printf("want: %v", want.AsText())
			log.Printf("got:  %v", got.AsText())
			return errMismatch
		}
	}
	return nil
}

func binaryChecks(h *Handle, g1, g2 geom.Geometry, log *log.Logger) error {
	for _, g := range []geom.Geometry{g1, g2} {
		if valid, err := checkIsValid(h, g, log); err != nil {
			return err
		} else if !valid {
			return nil
		}
	}

	log.Println("checking Intersects")
	if err := checkIntersects(h, g1, g2, log); err != nil {
		return err
	}

	log.Println("checking ExactEquals")
	if err := checkExactEquals(h, g1, g2, log); err != nil {
		return err
	}

	log.Println("checking Distance")
	if err := checkDistance(h, g1, g2, log); err != nil {
		return err
	}

	log.Println("checking DCEL operations")
	if err := checkDCELOperations(h, g1, g2, log); err != nil {
		return err
	}

	return nil
}

func checkIntersects(h *Handle, g1, g2 geom.Geometry, log *log.Logger) error {
	skipList := map[string]bool{
		// postgres=# SELECT ST_Intersects(
		//   ST_GeomFromText('LINESTRING(1 0,0.5000000000000001 0.5,0 1)'),
		//   ST_GeomFromText('LINESTRING(0.5 0.5,1.5 1.5)')
		// );
		//  st_intersects
		// ---------------
		//  f # WRONG!!
		// (1 row)
		"LINESTRING(1 0,0.5000000000000001 0.5,0 1)": true,

		// Simplefeatures sometimes gives an incorrect result for this due to
		// numerical precision issues. Would be solved by
		// https://github.com/peterstace/simplefeatures/issues/274
		"LINESTRING(0.5 0,0.5000000000000001 0.5)":                              true,
		"MULTILINESTRING((0 0,2 2.000000000000001),(1 0,-1 2.000000000000001))": true,
	}
	if skipList[g1.AsText()] || skipList[g2.AsText()] {
		// Skipping test because GEOS gives the incorrect result for *some*
		// intersection operations involving this input.
		return nil
	}

	want, err := h.intersects(g1, g2)
	if err != nil {
		if err == errLibgeosCrash {
			return nil
		}
		return err
	}
	got := geom.Intersects(g1, g2)

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}

func checkExactEquals(h *Handle, g1, g2 geom.Geometry, log *log.Logger) error {
	want, err := h.exactEquals(g1, g2)
	if err != nil {
		return err
	}
	got := geom.ExactEquals(g1, g2)

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}

func checkDistance(h *Handle, g1, g2 geom.Geometry, log *log.Logger) error {
	want, err := h.distance(g1, g2)
	if err != nil {
		if err == errLibgeosCrash {
			// Skip any tests that would cause libgeos to crash.
			return nil
		}
		return err
	}
	got, ok := geom.Distance(g1, g2)
	if !ok {
		// GEOS gives 0 when distance is not defined.
		got = 0
	}

	if math.Abs(want-got) > 1e-12 {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}

var skipIntersection = map[string]bool{
	"LINESTRING(0 1,0.3333333333 0.6666666667,1 0)": true,
	"LINESTRING(1 0,0.5000000000000001 0.5,0 1)":    true,
	"MULTILINESTRING((0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2),(0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))": true,
	"MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))":                                                                 true,
	"MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0),(0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2))": true,
	"POLYGON((1 0,0.9807852804032305 -0.19509032201612808,0.923879532511287 -0.3826834323650894,0.8314696123025456 -0.5555702330196017,0.7071067811865481 -0.7071067811865469,0.5555702330196031 -0.8314696123025447,0.38268343236509084 -0.9238795325112863,0.19509032201612964 -0.9807852804032302,0.0000000000000016155445744325867 -1,-0.19509032201612647 -0.9807852804032308,-0.38268343236508784 -0.9238795325112875,-0.5555702330196005 -0.8314696123025463,-0.7071067811865459 -0.7071067811865491,-0.8314696123025438 -0.5555702330196043,-0.9238795325112857 -0.38268343236509234,-0.9807852804032299 -0.19509032201613122,-1 -0.0000000000000032310891488651735,-0.9807852804032311 0.19509032201612486,-0.9238795325112882 0.38268343236508634,-0.8314696123025475 0.555570233019599,-0.7071067811865505 0.7071067811865446,-0.5555702330196058 0.8314696123025428,-0.3826834323650936 0.9238795325112852,-0.19509032201613213 0.9807852804032297,-0.000000000000003736410698672604 1,0.1950903220161248 0.9807852804032311,0.38268343236508673 0.9238795325112881,0.5555702330195996 0.8314696123025469,0.7071067811865455 0.7071067811865496,0.8314696123025438 0.5555702330196044,0.9238795325112859 0.38268343236509206,0.98078528040323 0.19509032201613047,1 0))": true,
	"MULTILINESTRING((0 0,2 2.000000000000001),(1 0,-1 2.000000000000001))":                         true,
	"MULTILINESTRING((0 0,0.5 0.5,1 1,2 2.000000000000001),(1 0,0.5 0.5,0 1,-1 2.000000000000001))": true,
	"POLYGON((1.5 1,1.353553390593274 0.6464466094067265,1.0000000000000009 0.5,0.646446609406727 0.6464466094067254,0.5 0.9999999999999983,0.6464466094067247 1.3535533905932722,0.9999999999999977 1.5,1.3535533905932717 1.3535533905932757,1.5 1))": true,
	"POLYGON((1 0,-0.9 -0.2,-1 -0.0000000000000032310891488651735,-0.9 0.2,1 0))": true,
	"LINESTRING(0.5 0,0.5000000000000001 0.5)":                                    true,
	"LINESTRING(0.5 1,0.5000000000000001 0.5)":                                    true,
	"POLYGON((1 0,0.9807852804032304 -0.19509032201612825,0.9238795325112867 -0.3826834323650898,0.8314696123025452 -0.5555702330196022,0.7071067811865476 -0.7071067811865475,0.5555702330196023 -0.8314696123025452,0.38268343236508984 -0.9238795325112867,0.19509032201612833 -0.9807852804032304,0.00000000000000006123233995736766 -1,-0.1950903220161282 -0.9807852804032304,-0.3826834323650897 -0.9238795325112867,-0.555570233019602 -0.8314696123025455,-0.7071067811865475 -0.7071067811865476,-0.8314696123025453 -0.5555702330196022,-0.9238795325112867 -0.3826834323650899,-0.9807852804032304 -0.1950903220161286,-1 -0.00000000000000012246467991473532,-0.9807852804032304 0.19509032201612836,-0.9238795325112868 0.38268343236508967,-0.8314696123025455 0.555570233019602,-0.7071067811865477 0.7071067811865475,-0.5555702330196022 0.8314696123025452,-0.38268343236509034 0.9238795325112865,-0.19509032201612866 0.9807852804032303,-0.00000000000000018369701987210297 1,0.1950903220161283 0.9807852804032304,0.38268343236509 0.9238795325112866,0.5555702330196018 0.8314696123025455,0.7071067811865474 0.7071067811865477,0.8314696123025452 0.5555702330196022,0.9238795325112865 0.3826834323650904,0.9807852804032303 0.19509032201612872,1 0))": true,
	"POLYGON((1.5 1,1.3535533905932737 0.6464466094067263,1 0.5,0.6464466094067263 0.6464466094067263,0.5 0.9999999999999999,0.6464466094067262 1.3535533905932737,0.9999999999999999 1.5,1.3535533905932737 1.353553390593274,1.5 1))": true,

	// Cause simplefeatures DCEL operations to fail with "no rings" error. See
	// https://github.com/peterstace/simplefeatures/pull/497 for details.
	"POLYGON((-83.58253051 32.73168239,-83.59843118 32.74617142,-83.70048117 32.63984372,-83.58253051 32.73168239))": true,
	"POLYGON((-83.70047745 32.63984661,-83.68891846 32.5989632,-83.58253417 32.73167955,-83.70047745 32.63984661))":  true,
}

var skipDifference = map[string]bool{
	"LINESTRING(0 1,0.3333333333 0.6666666667,0.5 0.5,1 0)": true,
	"LINESTRING(0 1,0.3333333333 0.6666666667,1 0)":         true,
	"LINESTRING(1 0,0.5000000000000001 0.5,0 1)":            true,
	"MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,0.5 0.5),(0.5 0.5,1 0))": true,
	"MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))":                                               true,

	"MULTIPOLYGON(((1 0,0 1,0.5 1.5,1 1,1.5 1.5,2 1,1 0)),((1.5 1.5,1 2,0.5 1.5,0.333333333333333 1.66666666666667,0 2,1 3,2 2,1.5 1.5)),((3.5 1.5,4 1,3 0,2 1,2.5 1.5,3 1,3.5 1.5)),((3.5 1.5,3 2,2.5 1.5,2 2,3 3,4 2,3.5 1.5)))": true,
	"POLYGON((1 0,-0.9 -0.2,-1 -0.0000000000000032310891488651735,-0.9 0.2,1 0))": true,
	"POLYGON((1 0,0.9807852804032305 -0.19509032201612808,0.923879532511287 -0.3826834323650894,0.8314696123025456 -0.5555702330196017,0.7071067811865481 -0.7071067811865469,0.5555702330196031 -0.8314696123025447,0.38268343236509084 -0.9238795325112863,0.19509032201612964 -0.9807852804032302,0.0000000000000016155445744325867 -1,-0.19509032201612647 -0.9807852804032308,-0.38268343236508784 -0.9238795325112875,-0.5555702330196005 -0.8314696123025463,-0.7071067811865459 -0.7071067811865491,-0.8314696123025438 -0.5555702330196043,-0.9238795325112857 -0.38268343236509234,-0.9807852804032299 -0.19509032201613122,-1 -0.0000000000000032310891488651735,-0.9807852804032311 0.19509032201612486,-0.9238795325112882 0.38268343236508634,-0.8314696123025475 0.555570233019599,-0.7071067811865505 0.7071067811865446,-0.5555702330196058 0.8314696123025428,-0.3826834323650936 0.9238795325112852,-0.19509032201613213 0.9807852804032297,-0.000000000000003736410698672604 1,0.1950903220161248 0.9807852804032311,0.38268343236508673 0.9238795325112881,0.5555702330195996 0.8314696123025469,0.7071067811865455 0.7071067811865496,0.8314696123025438 0.5555702330196044,0.9238795325112859 0.38268343236509206,0.98078528040323 0.19509032201613047,1 0))": true,
	"MULTILINESTRING((0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2),(0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))": true,
	"MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0),(0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2))": true,
	"MULTILINESTRING((0 0,2 2.000000000000001),(1 0,-1 2.000000000000001))":                                                                            true,
	"MULTILINESTRING((0 0,0.5 0.5,1 1,2 2.000000000000001),(1 0,0.5 0.5,0 1,-1 2.000000000000001))":                                                    true,
	"MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(0 1,0.3333333333 0.6666666667,0.5 0.5),(0.5 0.5,1 0))":                                               true,
	"LINESTRING(0.5 0,0.5000000000000001 0.5)": true,
	"LINESTRING(0.5 1,0.5000000000000001 0.5)": true,

	// The following are not topological differences, but instead bugs in GEOS
	// v3.7.1. I believe that it may be fixed in GEOS v3.8.0 but I haven't
	// confirmed that.
	"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.0717462120245884 0.06467514421272098,0.9353248557872769 -0.07174621202458649,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))":                                                                                                                                                                                                                      true,
	"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.1 0,1 -0.1,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))":                                                                                                                                                                                                                                                                                        true,
	"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.1 0,1.0980785280403231 -0.019509032201612826,1.0923879532511287 -0.03826834323650898,1.0831469612302544 -0.05555702330196022,1.0707106781186548 -0.07071067811865475,1.0555570233019602 -0.08314696123025453,1.038268343236509 -0.09238795325112868,1.019509032201613 -0.09807852804032305,1 -0.1,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))": true,
	"POLYGON((1 0,0.9807852804032304 -0.19509032201612825,0.9238795325112867 -0.3826834323650898,0.8314696123025452 -0.5555702330196022,0.7071067811865476 -0.7071067811865475,0.5555702330196023 -0.8314696123025452,0.38268343236508984 -0.9238795325112867,0.19509032201612833 -0.9807852804032304,0.00000000000000006123233995736766 -1,-0.1950903220161282 -0.9807852804032304,-0.3826834323650897 -0.9238795325112867,-0.555570233019602 -0.8314696123025455,-0.7071067811865475 -0.7071067811865476,-0.8314696123025453 -0.5555702330196022,-0.9238795325112867 -0.3826834323650899,-0.9807852804032304 -0.1950903220161286,-1 -0.00000000000000012246467991473532,-0.9807852804032304 0.19509032201612836,-0.9238795325112868 0.38268343236508967,-0.8314696123025455 0.555570233019602,-0.7071067811865477 0.7071067811865475,-0.5555702330196022 0.8314696123025452,-0.38268343236509034 0.9238795325112865,-0.19509032201612866 0.9807852804032303,-0.00000000000000018369701987210297 1,0.1950903220161283 0.9807852804032304,0.38268343236509 0.9238795325112866,0.5555702330196018 0.8314696123025455,0.7071067811865474 0.7071067811865477,0.8314696123025452 0.5555702330196022,0.9238795325112865 0.3826834323650904,0.9807852804032303 0.19509032201612872,1 0))":                                                                                                                                                                                                                                                                               true,

	// Cause simplefeatures DCEL operations to fail with "no rings" error. See
	// https://github.com/peterstace/simplefeatures/pull/497 for details.
	"POLYGON((-83.58253051 32.73168239,-83.59843118 32.74617142,-83.70048117 32.63984372,-83.58253051 32.73168239))": true,
	"POLYGON((-83.70047745 32.63984661,-83.68891846 32.5989632,-83.58253417 32.73167955,-83.70047745 32.63984661))":  true,

	// Causes simplefeatures DCEL operations to fail with "polygon ring not simple" error.
	"POLYGON((-1 0,-0.9 0.2,-0.80952380952381 0.19047619047619,0 1,0 0.105263157894737,1 0,-0.9 -0.2,-1 0))":                                                                                                                                                   true,
	"POLYGON((0 -0.1414213562373095,1.1414213562373097 1,0.9292893218813453 1.0707106781186548,0 -0.1414213562373095))":                                                                                                                                        true,
	"POLYGON((0.9292893218813453 1.0707106781186548,1.0707106781186548 0.9292893218813453,0.07071067811865475 -0.07071067811865475,-0.07071067811865475 0.07071067811865475,0.9292893218813453 1.0707106781186548))":                                           true,
	"POLYGON((0.9292893218813453 1.0707106781186548,1 1.1414213562373097,1.1414213562373097 1,0.07071067811865475 -0.07071067811865475,0 -0.1414213562373095,-0.1414213562373095 -0.000000000000000013877787807814457,0.9292893218813453 1.0707106781186548))": true,
	"POLYGON((0.9292893218813453 1.0707106781186548,1.1414213562373097 1,0.07071067811865475 -0.07071067811865475,0 -0.1414213562373095,0.9292893218813453 1.0707106781186548))":                                                                               true,
}

var skipSymDiff = map[string]bool{
	"LINESTRING(0 1,0.3333333333 0.6666666667,0.5 0.5,1 0)": true,
	"LINESTRING(0 1,0.3333333333 0.6666666667,1 0)":         true,
	"LINESTRING(1 0,0.5000000000000001 0.5,0 1)":            true,
	"MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,0.5 0.5),(0.5 0.5,1 0))": true,
	"MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))":                                               true,

	// Cause simplefeatures DCEL operations to fail with "no rings" error. See
	// https://github.com/peterstace/simplefeatures/pull/497 for details.
	"POLYGON((-83.58253051 32.73168239,-83.59843118 32.74617142,-83.70048117 32.63984372,-83.58253051 32.73168239))": true,
	"POLYGON((-83.70047745 32.63984661,-83.68891846 32.5989632,-83.58253417 32.73167955,-83.70047745 32.63984661))":  true,

	// Causes simplefeatures "multipolygon has nested child polygons" error.
	"POLYGON((1 0,0.9807852804032304 -0.19509032201612825,0.9238795325112867 -0.3826834323650898,0.8314696123025452 -0.5555702330196022,0.7071067811865476 -0.7071067811865475,0.5555702330196023 -0.8314696123025452,0.38268343236508984 -0.9238795325112867,0.19509032201612833 -0.9807852804032304,0.00000000000000006123233995736766 -1,-0.1950903220161282 -0.9807852804032304,-0.3826834323650897 -0.9238795325112867,-0.555570233019602 -0.8314696123025455,-0.7071067811865475 -0.7071067811865476,-0.8314696123025453 -0.5555702330196022,-0.9238795325112867 -0.3826834323650899,-0.9807852804032304 -0.1950903220161286,-1 -0.00000000000000012246467991473532,-0.9807852804032304 0.19509032201612836,-0.9238795325112868 0.38268343236508967,-0.8314696123025455 0.555570233019602,-0.7071067811865477 0.7071067811865475,-0.5555702330196022 0.8314696123025452,-0.38268343236509034 0.9238795325112865,-0.19509032201612866 0.9807852804032303,-0.00000000000000018369701987210297 1,0.1950903220161283 0.9807852804032304,0.38268343236509 0.9238795325112866,0.5555702330196018 0.8314696123025455,0.7071067811865474 0.7071067811865477,0.8314696123025452 0.5555702330196022,0.9238795325112865 0.3826834323650904,0.9807852804032303 0.19509032201612872,1 0))": true,

	// Causes simplefeatures DCEL operations to fail with "polygon ring not simple" error.
	"POLYGON((-1 0,-0.9 0.2,-0.80952380952381 0.19047619047619,0 1,0 0.105263157894737,1 0,-0.9 -0.2,-1 0))":            true,
	"POLYGON((0 -0.1414213562373095,1.1414213562373097 1,0.9292893218813453 1.0707106781186548,0 -0.1414213562373095))": true,
	"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.0717462120245884 0.06467514421272098,0.9353248557872769 -0.07174621202458649,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))": true,
	"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.1 0,1 -0.1,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))":                                                                   true,
	"POLYGON((0.9292893218813453 1.0707106781186548,1.0707106781186548 0.9292893218813453,0.07071067811865475 -0.07071067811865475,-0.07071067811865475 0.07071067811865475,0.9292893218813453 1.0707106781186548))": true,
	"POLYGON((0.9292893218813453 1.0707106781186548,0.9444429766980399 1.0831469612302547,0.9617316567634911 1.0923879532511287,0.9804909677983873 1.0980785280403231,1 1.1,1.019509032201613 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302544,1.0707106781186548 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.038268343236509,1.0980785280403231 1.019509032201613,1.1 1,1.0980785280403231 0.9804909677983873,1.0923879532511287 0.9617316567634911,1.0831469612302547 0.9444429766980399,1.0707106781186548 0.9292893218813453,0.07071067811865475 -0.07071067811865475,0.05555702330196012 -0.0831469612302546,0.03826834323650888 -0.09238795325112872,0.019509032201612746 -0.09807852804032308,-0.00000000000000006049014748177263 -0.1,-0.019509032201612864 -0.09807852804032305,-0.03826834323650899 -0.09238795325112868,-0.0555570233019602 -0.08314696123025456,-0.07071067811865475 -0.07071067811865477,-0.08314696123025449 -0.05555702330196029,-0.09238795325112865 -0.03826834323650903,-0.09807852804032302 -0.019509032201612948,-0.1 -0.00000000000000010106430996148606,-0.09807852804032308 0.019509032201612663,-0.09238795325112874 0.038268343236508844,-0.08314696123025465 0.055557023301960044,-0.07071067811865475 0.07071067811865475,0.9292893218813453 1.0707106781186548))": true,
}

var skipUnion = map[string]bool{
	// Cause simplefeatures DCEL operations to fail with "no rings" error. See
	// https://github.com/peterstace/simplefeatures/pull/497 for details.
	"POLYGON((-83.58253051 32.73168239,-83.59843118 32.74617142,-83.70048117 32.63984372,-83.58253051 32.73168239))": true,
	"POLYGON((-83.70047745 32.63984661,-83.68891846 32.5989632,-83.58253417 32.73167955,-83.70047745 32.63984661))":  true,
}

func checkDCELOperations(h *Handle, g1, g2 geom.Geometry, log *log.Logger) error {
	// TODO: simplefeatures doesn't support GeometryCollections yet
	if g1.IsGeometryCollection() || g2.IsGeometryCollection() {
		return nil
	}

	for _, op := range []struct {
		name     string
		sfFunc   func(g1, g2 geom.Geometry) (geom.Geometry, error)
		geosFunc func(g1, g2 geom.Geometry) (geom.Geometry, error)
		skip     map[string]bool
	}{
		{
			"Union",
			func(g1, g2 geom.Geometry) (geom.Geometry, error) { return geom.Union(g1, g2) },
			func(g1, g2 geom.Geometry) (geom.Geometry, error) { return h.union(g1, g2) },
			skipUnion,
		},
		{
			"Intersection",
			func(g1, g2 geom.Geometry) (geom.Geometry, error) { return geom.Intersection(g1, g2) },
			func(g1, g2 geom.Geometry) (geom.Geometry, error) { return h.intersection(g1, g2) },
			skipIntersection,
		},
		{
			"Difference",
			func(g1, g2 geom.Geometry) (geom.Geometry, error) { return geom.Difference(g1, g2) },
			func(g1, g2 geom.Geometry) (geom.Geometry, error) { return h.difference(g1, g2) },
			skipDifference,
		},
		{
			"SymmetricDifference",
			func(g1, g2 geom.Geometry) (geom.Geometry, error) { return geom.SymmetricDifference(g1, g2) },
			func(g1, g2 geom.Geometry) (geom.Geometry, error) { return h.symmetricDifference(g1, g2) },
			skipSymDiff,
		},
	} {
		log.Println("checking", op.name)
		err := checkDCELOp(op.sfFunc, op.geosFunc, g1, g2, op.skip, log)
		if err != nil {
			return err
		}
	}

	log.Println("checking Relate")
	if err := checkRelate(h, g1, g2, log); err != nil {
		return err
	}

	return nil
}

func checkDCELOp(
	op func(g1, g2 geom.Geometry) (geom.Geometry, error),
	refImpl func(g1, g2 geom.Geometry) (geom.Geometry, error),
	g1, g2 geom.Geometry,
	skip map[string]bool,
	log *log.Logger,
) error {
	// Empty points will cause the reference impl to crash.
	if hasEmptyPoint(g1) || hasEmptyPoint(g2) {
		return nil
	}

	// Some geometries give results that are not topologically equivalent to
	// those from GEOS. These have been checked manually, and decided that the
	// difference is acceptable (they typically have to do with different
	// handling of numerically degenerate cases). Note that we bail out of this
	// test _after_ we calculate got. That way we're at least checking that it
	// doesn't crash or give an error.
	if skip[g1.AsText()] || skip[g2.AsText()] {
		return nil
	}

	got, err := op(g1, g2)
	if err != nil {
		return err
	}

	want, err := refImpl(g1, g2)
	if err != nil {
		if err == errInvalidAccordingToGEOS {
			// Because GEOS has given us back an invalid geometry (even according
			// to its own validation routines) we can't trust it for this test
			// case.
			return nil
		}
		return err
	}

	if !mantissaTerminatesQuickly(got) || !mantissaTerminatesQuickly(want) {
		// We're not going to be able to compare got and want because of
		// numeric precision issues.
		log.Printf("mantissa doesn't terminate quickly, using area heuristic")
		if err := checkEqualityHeuristic(want, got, log); err != nil {
			return err
		}
		return nil
	}

	if want.IsGeometryCollection() || got.IsGeometryCollection() {
		// We can't use Equals from GEOS on GeometryCollections, so we can't
		// use proper Equals for this case.
		log.Printf("want or got is a geometry collection, using area heuristic")
		if err := checkEqualityHeuristic(want, got, log); err != nil {
			return err
		}
		return nil
	}

	eq, err := geos.Equals(want, got)
	if err != nil {
		return err
	}
	if !eq {
		log.Printf("want: %v", want.AsText())
		log.Printf("got:  %v", got.AsText())
		return errMismatch
	}
	return nil
}

// checkEqualityHeuristic checks some necessary but not sufficient properties
// of two geometries if they are to be equal.
//
// TODO: we could come up with some smarter heuristics. E.g. distance sampled
// by many random points.
func checkEqualityHeuristic(want, got geom.Geometry, log *log.Logger) error {
	wantArea := want.Area()
	gotArea := got.Area()
	if math.Abs(wantArea-gotArea) > 1e-3 {
		log.Printf("wantWKT: %v\n", want.AsText())
		log.Printf("gotWKT:  %v\n", got.AsText())
		log.Printf("wantArea: %v\n", wantArea)
		log.Printf("gotArea:  %v\n", gotArea)
		return errMismatch
	}
	return nil
}

func checkRelate(h *Handle, g1, g2 geom.Geometry, log *log.Logger) error {
	got, err := geom.Relate(g1, g2)
	if err != nil {
		return err
	}
	want, err := h.relate(g1, g2)
	if err != nil {
		if err == errLibgeosCrash {
			// Skip any tests that would cause libgeos to crash.
			return nil
		}
		return err
	}

	// Skip any linear and non-simple geometries. This is because GEOS has
	// inconsistent behaviour with the generated relate matrix, making it hard
	// to match the exact behaviour.
	if linearAndNonSimple(g1) || linearAndNonSimple(g2) {
		return nil
	}

	if !mantissaTerminatesQuickly(g1) || !mantissaTerminatesQuickly(g2) {
		// Numerical precision issues cause a large number of geometries to
		// differ compared to GEOS. There aren't really any heuristics that we
		// can fall back to, so we just have to skip these sorts of geometries.
		return nil
	}

	// There is a bug in GEOS that triggers when linear elements have no
	// boundary (e.g. due to the mod-2 rule).  The result of the bug is that
	// the EB (or BE) is reported as 0 rather than F.
	if linearAndEmptyBoundary(g1) || linearAndEmptyBoundary(g2) {
		return nil
	}

	if got != want {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}

	return nil
}

func checkRelateMatch(h *Handle, log *log.Logger) error {
	for i := 0; i < 1_000_000; i++ {
		mat := rand9("F012")
		pat := rand9("F012T*")
		want, err := h.relateMatch(mat, pat)
		if err != nil {
			log.Printf("could not calculate want: %v", err)
			return err
		}
		got, err := geom.RelateMatches(mat, pat)
		if err != nil {
			log.Printf("could not calculate got: %v", err)
			return err
		}
		if got != want {
			log.Printf("mat:  %v", mat)
			log.Printf("pat:  %v", pat)
			log.Printf("want: %v", want)
			log.Printf("got:  %v", got)
			return errMismatch
		}
	}
	return nil
}

func rand9(alphabet string) string {
	var buf [9]byte
	for i := range buf {
		buf[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(buf[:])
}
