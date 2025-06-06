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
	"github.com/peterstace/simplefeatures/internal/rawgeos"
)

func unaryChecks(g geom.Geometry, lg *log.Logger) error {
	if valid, err := checkIsValid(g, lg); err != nil {
		return err
	} else if !valid {
		return nil
	}

	for _, check := range []struct {
		name string
		fn   func(geom.Geometry, *log.Logger) error
	}{
		{"AsBinary", checkAsBinary},
		{"FromBinary", checkFromBinary},
		{"AsText", checkAsText},
		{"FromText", checkFromText},
		{"IsEmpty", checkIsEmpty},
		{"Dimension", checkDimension},
		{"Envelope", checkEnvelope},
		{"IsSimple", checkIsSimple},
		{"Boundary", checkBoundary},
		{"ConvexHull", checkConvexHull},
		{"IsRing", checkIsRing},
		{"Length", checkLength},
		{"Area", checkArea},
		{"Centroid", checkCentroid},
		{"PointOnSurface", checkPointOnSurface},
		{"Simplify", checkSimplify},
		{"RotatedMinimumAreaBoundingRectangle", checkRotatedMinimumAreaBoundingRectangle},
	} {
		lg.Printf("checking %s", check.name)
		if err := check.fn(g, lg); err != nil {
			return err
		}
	}

	return nil

	// TODO: Reverse isn't checked yet. There is some significant behaviour
	// differences between libgeos and PostGIS.
}

var errMismatch = errors.New("mismatch")

func checkIsValid(g geom.Geometry, log *log.Logger) (bool, error) {
	wkb := g.AsBinary()
	var validAsPerSimpleFeatures bool
	if _, err := geom.UnmarshalWKB(wkb); err == nil {
		validAsPerSimpleFeatures = true
	}
	log.Printf("Valid as per simplefeatures: %v", validAsPerSimpleFeatures)

	validAsPerLibgeos, err := rawgeos.IsValid(g)
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

func checkAsText(g geom.Geometry, log *log.Logger) error {
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

	want, err := rawgeos.AsText(g)
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

func checkFromText(g geom.Geometry, log *log.Logger) error {
	// libgeos is unable to parse MultiPoints if the *first* Point is empty. It
	// gives the following error: ParseException: Unexpected token: WORD EMPTY.
	// Skip the check in that case.
	if g.IsMultiPoint() &&
		g.MustAsMultiPoint().NumPoints() > 0 &&
		g.MustAsMultiPoint().PointN(0).IsEmpty() {
		return nil
	}

	wkt := g.AsText()
	want, err := rawgeos.FromText(wkt)
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

func checkAsBinary(g geom.Geometry, log *log.Logger) error {
	var wantDefined bool
	want, err := rawgeos.AsBinary(g)
	if err == nil {
		wantDefined = true
	}
	hAsPointEmpty := hasEmptyPoint(g)
	if !wantDefined && !hAsPointEmpty {
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

func checkFromBinary(g geom.Geometry, log *log.Logger) error {
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

	want, err := rawgeos.FromBinary(wkb)
	if err != nil {
		return err
	}

	got, err := geom.UnmarshalWKB(wkb)
	if err != nil {
		return err
	}

	if !geom.ExactEquals(want, got) {
		log.Printf("wkb:\n%s", hex.Dump(wkb))
		log.Printf("want:\n%s", want.AsText())
		log.Printf("got:\n%s", got.AsText())
		return errMismatch
	}
	return nil
}

func checkIsEmpty(g geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.IsEmpty(g)
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

func checkDimension(g geom.Geometry, log *log.Logger) error {
	var want int
	if !containsOnlyGeometryCollections(g) {
		// Libgeos gives -1 dimension for GeometryCollection trees that only
		// contain other GeometryCollections (all the way to the leaf nodes).
		// This is weird behaviour, and the dimension should actually be zero.
		// So we don't get 'want' from libgeos in that case (and allow want to
		// default to 0).
		var err error
		want, err = rawgeos.Dimension(g)
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

func checkEnvelope(g geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.Envelope(g)
	if err != nil {
		return err
	}
	got := g.Envelope()
	gotMin, gotMax, gotNonEmpty := got.MinMaxXYs()

	wantCoords := want.DumpCoordinates()
	if wantCoords.Length() == 0 {
		if gotNonEmpty {
			log.Printf("want: %v", want.AsText())
			log.Printf("got: %v", got.AsGeometry().AsText())
			return errMismatch
		}
		return nil
	}

	minx := wantCoords.GetXY(0).X
	miny := wantCoords.GetXY(0).Y
	maxx := minx
	maxy := miny
	for i := 1; i < wantCoords.Length(); i++ {
		xy := wantCoords.GetXY(i)
		minx = math.Min(minx, xy.X)
		miny = math.Min(miny, xy.Y)
		maxx = math.Max(maxx, xy.X)
		maxy = math.Max(maxy, xy.Y)
	}

	if gotMin != (geom.XY{X: minx, Y: miny}) || gotMax != (geom.XY{X: maxx, Y: maxy}) {
		log.Printf("want: %v", want.AsText())
		log.Printf("got: %v", got.AsGeometry().AsText())
		return errMismatch
	}
	return nil
}

func checkIsSimple(g geom.Geometry, log *log.Logger) error {
	if containsMultiPointWithEmptyPoint(g) {
		// libgeos crashes when GEOSisSimple_r is called with MultiPoints
		// containing empty Points.
		return nil
	}

	var wantDefined, wantSimple bool
	if !g.IsGeometryCollection() {
		wantDefined = true
		var err error
		wantSimple, err = rawgeos.IsSimple(g)
		if err != nil {
			return err
		}
	}

	gotSimple, gotDefined := g.IsSimple()

	if wantDefined != gotDefined {
		log.Printf("want defined: %v", wantDefined)
		log.Printf("got defined: %v", gotDefined)
		return errMismatch
	}
	if !gotDefined {
		return nil
	}

	if wantSimple != gotSimple {
		log.Printf("want: %v", wantSimple)
		log.Printf("got:  %v", gotSimple)
		return errMismatch
	}
	return nil
}

func checkBoundary(g geom.Geometry, log *log.Logger) error {
	if g.Type() == geom.TypeGeometryCollection {
		// libgeos doesn't define the boundary of GeometryCollections, but
		// simplefeatures does. So we skip the test in this case.
		return nil
	}

	want, err := rawgeos.Boundary(g)
	if err != nil {
		return err
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

func checkConvexHull(g geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.ConvexHull(g)
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

func checkIsRing(g geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.IsRing(g)
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

func checkLength(g geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.Length(g)
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

func checkArea(g geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.Area(g)
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

func checkCentroid(g geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.Centroid(g)
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

func checkPointOnSurface(g geom.Geometry, log *log.Logger) error {
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
		intersects, err := rawgeos.Intersects(pt, g)
		if err != nil {
			return err
		}
		if !intersects {
			log.Printf("the pt doesn't intersect with the input")
			return errMismatch
		}
	}

	if g.Dimension() == 2 && !g.IsEmpty() && !g.IsGeometryCollection() {
		contains, err := rawgeos.Contains(g, pt)
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

func checkSimplify(g geom.Geometry, log *log.Logger) error {
	for _, threshold := range []float64{0.125, 0.25, 0.5, 1, 2, 4, 8, 16} {
		// If we get an error from GEOS, then we may or may not get an error from
		// simplefeatures.
		want, err := rawgeos.Simplify(g, threshold)
		wantIsValid := err == nil

		// Even if GEOS couldn't simplify, we still want to attempt to simplify
		// with simplefeatures to ensure it doesn't crash (even if it may give an
		// error).
		got, err := rawgeos.Simplify(g, threshold)
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

func checkRotatedMinimumAreaBoundingRectangle(g geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.MinimumRotatedRectangle(g)
	if err != nil {
		return err
	}
	wantArea := want.Area()

	got := geom.RotatedMinimumAreaBoundingRectangle(g)
	gotArea := got.Area()

	// The rotated bounding rectangle with minimum area is not always unique
	// (multiple could exist with the same minimum area). Simplefeatures and
	// GEOS (both correctly) choose different ones in some cases. To account
	// for this, the comparison between the GEOS result and the simplefeatures
	// result is broken into two parts...

	// ...First, the areas are compared.
	const areaDiffThreshold = 1e-10
	if math.Abs(wantArea-gotArea) > areaDiffThreshold {
		log.Printf("areas differ by more than %v", areaDiffThreshold)
		log.Printf("want: (area %v) %v", wantArea, want.AsText())
		log.Printf("got:  (area %v) %v", gotArea, got.AsText())
		return errMismatch
	}

	// ...Second, we check if any of the input geometry is outside of the
	// minimum bounding rectangle. Since GEOS cannot compute the difference of
	// two geometries if one of them is a GeometryCollection, we break into
	// parts and check the difference of each part individually.
	var parts []geom.Geometry
	if gc, ok := g.AsGeometryCollection(); ok {
		parts = gc.Dump()
	} else {
		parts = []geom.Geometry{g}
	}
	for i, part := range parts {
		overhang, err := rawgeos.Difference(part, got)
		if err != nil {
			return err
		}
		overhangArea := overhang.Area()
		const overhangAreaThreshold = 1e-14
		if overhangArea > overhangAreaThreshold {
			log.Printf("part WKT (%d of %d): %v", i+1, len(parts), part.AsText())
			log.Printf("part area overhangs MBR by %v (threshold %v)", overhangArea, overhangAreaThreshold)
			log.Printf("overhang: (area %v) %v", overhangArea, overhang.AsText())
			log.Printf("want: (area %v) %v", wantArea, want.AsText())
			log.Printf("got:  (area %v) %v", gotArea, got.AsText())
			return errMismatch
		}
	}
	return nil
}

func binaryChecks(g1, g2 geom.Geometry, lg *log.Logger) error {
	for _, g := range []geom.Geometry{g1, g2} {
		if valid, err := checkIsValid(g, lg); err != nil {
			return err
		} else if !valid {
			return nil
		}
	}

	for _, check := range []struct {
		name string
		fn   func(geom.Geometry, geom.Geometry, *log.Logger) error
	}{
		{"Intersects", checkIntersects},
		{"ExactEquals", checkExactEquals},
		{"Distance", checkDistance},
		{"DCELOperations", checkDCELOperations},
	} {
		lg.Printf("checking %s", check.name)
		if err := check.fn(g1, g2, lg); err != nil {
			return err
		}
	}
	return nil
}

func checkIntersects(g1, g2 geom.Geometry, log *log.Logger) error {
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

		// GEOS gives the wrong result for the intersection of these two inputs:
		"POLYGON((4.4 8.2,2.8 7.4,5.4 2.2,7 3,4.4 8.2))": true,
		"POLYGON((1 4,3 4,3 7,1 7,1 4))":                 true,
		"POLYGON((1.5827586206896551 -0.49310344827586206,7.575862068965518 6.589655172413793,5.424137931034483 8.410344827586208,-0.5689655172413792 1.3275862068965518,1.5827586206896551 -0.49310344827586206))": true,
		"POLYGON((-0.057692307692307696 -0.038461538461538464,3 1.9999999999999998,2.230769230769231 3.1538461538461537,-0.826923076923077 1.1153846153846154,-0.057692307692307696 -0.038461538461538464))":        true,
	}
	if skipList[g1.AsText()] || skipList[g2.AsText()] {
		// Skipping test because GEOS gives the incorrect result for *some*
		// intersection operations involving this input.
		return nil
	}

	want, err := rawgeos.Intersects(g1, g2)
	if err != nil {
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

func checkExactEquals(g1, g2 geom.Geometry, log *log.Logger) error {
	want, err := rawgeos.EqualsExact(g1, g2)
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

func checkDistance(g1, g2 geom.Geometry, log *log.Logger) error {
	willCauseGEOSCrash := false ||
		containsMultiLineStringWithEmptyLineString(g1) ||
		containsMultiLineStringWithEmptyLineString(g2) ||
		containsMultiPointWithEmptyPoint(g1) ||
		containsMultiPointWithEmptyPoint(g2) ||
		containsMultiPolygonWithEmptyPolygon(g1) ||
		containsMultiPolygonWithEmptyPolygon(g2)
	if willCauseGEOSCrash {
		// Skip test since attempting to calculate distance will cause a GEOS crash.
		return nil
	}

	want, err := rawgeos.Distance(g1, g2)
	if err != nil {
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
	"POINT(333673.327 6252387.751)": true,
	"POINT(456567.479 3973182.99)":  true,

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
	"POINT(456567.479 3973182.99)":  true,
	"POINT(333673.327 6252387.751)": true,

	// Causes simplefeatures "multipolygon has nested child polygons" error.
	"POLYGON((1 0,0.9807852804032304 -0.19509032201612825,0.9238795325112867 -0.3826834323650898,0.8314696123025452 -0.5555702330196022,0.7071067811865476 -0.7071067811865475,0.5555702330196023 -0.8314696123025452,0.38268343236508984 -0.9238795325112867,0.19509032201612833 -0.9807852804032304,0.00000000000000006123233995736766 -1,-0.1950903220161282 -0.9807852804032304,-0.3826834323650897 -0.9238795325112867,-0.555570233019602 -0.8314696123025455,-0.7071067811865475 -0.7071067811865476,-0.8314696123025453 -0.5555702330196022,-0.9238795325112867 -0.3826834323650899,-0.9807852804032304 -0.1950903220161286,-1 -0.00000000000000012246467991473532,-0.9807852804032304 0.19509032201612836,-0.9238795325112868 0.38268343236508967,-0.8314696123025455 0.555570233019602,-0.7071067811865477 0.7071067811865475,-0.5555702330196022 0.8314696123025452,-0.38268343236509034 0.9238795325112865,-0.19509032201612866 0.9807852804032303,-0.00000000000000018369701987210297 1,0.1950903220161283 0.9807852804032304,0.38268343236509 0.9238795325112866,0.5555702330196018 0.8314696123025455,0.7071067811865474 0.7071067811865477,0.8314696123025452 0.5555702330196022,0.9238795325112865 0.3826834323650904,0.9807852804032303 0.19509032201612872,1 0))": true,

	// Causes simplefeatures DCEL operations to fail with "polygon ring not simple" error.
	"POLYGON((-1 0,-0.9 0.2,-0.80952380952381 0.19047619047619,0 1,0 0.105263157894737,1 0,-0.9 -0.2,-1 0))":            true,
	"POLYGON((0 -0.1414213562373095,1.1414213562373097 1,0.9292893218813453 1.0707106781186548,0 -0.1414213562373095))": true,
	"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.0717462120245884 0.06467514421272098,0.9353248557872769 -0.07174621202458649,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))": true,
	"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.1 0,1 -0.1,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))":                                                                   true,
	"POLYGON((0.9292893218813453 1.0707106781186548,1.0707106781186548 0.9292893218813453,0.07071067811865475 -0.07071067811865475,-0.07071067811865475 0.07071067811865475,0.9292893218813453 1.0707106781186548))": true,
	"POLYGON((0.9292893218813453 1.0707106781186548,0.9444429766980399 1.0831469612302547,0.9617316567634911 1.0923879532511287,0.9804909677983873 1.0980785280403231,1 1.1,1.019509032201613 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302544,1.0707106781186548 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.038268343236509,1.0980785280403231 1.019509032201613,1.1 1,1.0980785280403231 0.9804909677983873,1.0923879532511287 0.9617316567634911,1.0831469612302547 0.9444429766980399,1.0707106781186548 0.9292893218813453,0.07071067811865475 -0.07071067811865475,0.05555702330196012 -0.0831469612302546,0.03826834323650888 -0.09238795325112872,0.019509032201612746 -0.09807852804032308,-0.00000000000000006049014748177263 -0.1,-0.019509032201612864 -0.09807852804032305,-0.03826834323650899 -0.09238795325112868,-0.0555570233019602 -0.08314696123025456,-0.07071067811865475 -0.07071067811865477,-0.08314696123025449 -0.05555702330196029,-0.09238795325112865 -0.03826834323650903,-0.09807852804032302 -0.019509032201612948,-0.1 -0.00000000000000010106430996148606,-0.09807852804032308 0.019509032201612663,-0.09238795325112874 0.038268343236508844,-0.08314696123025465 0.055557023301960044,-0.07071067811865475 0.07071067811865475,0.9292893218813453 1.0707106781186548))": true,
	"POLYGON((-83.5825305152402 32.7316823944815,-83.58376293006216 32.73315376178507,-83.58504085655653 32.734597137036324,-83.58636334101533 32.73601156235818,-83.58772946946186 32.73739605462324,-83.58913829287843 32.738749652823024,-83.59058883640313 32.740071417020225,-83.59208009385833 32.74136043125909,-83.59361103333862 32.74261579844315,-83.59518059080796 32.743836647669376,-83.59678768034422 32.745022130852156,-83.59843118482598 32.746171426099316,-83.60010996734121 32.74728373328835,-83.60182286094272 32.74835828151699,-83.60356867749573 32.74939432363171,-83.6053462070958 32.75039113983657,-83.60715421364506 32.75134803897381,-83.60899144521312 32.75226435508957,-83.61085662379259 32.75313945319649,-83.61274845635847 32.75397272333648,-83.61466562799973 32.7547635884388,-83.61660680960262 32.75551149954699,-83.61857065109865 32.756215934654755,-83.62055578961366 32.75687640673859,-83.62256084632457 32.75749245578336,-83.62458442890414 32.75806365483586,-83.62662513245226 32.75858960587211,-83.62868153809899 32.759069943900975,-83.63075221766115 32.759504334926895,-83.6328357331767 32.759892478947584,-83.6349306369047 32.76023410294001,-83.63703547248946 32.76052897293848,-83.63914877915153 32.76077688192745,-83.64126908772954 32.760977657903275,-83.64339492533685 32.76113116287015,-83.64552481489585 32.76123728881648,-83.64765727560365 32.761295961756694,-83.64979082712301 32.761307141684846,-83.65192398562425 32.76127082060284,-83.65405527030447 32.761187023508135,-83.65618319896379 32.76105580840831,-83.65830629359328 32.76087726729329,-83.66042307921083 32.7606515241745,-83.66253208572374 32.760378736038,-83.66463184629896 32.76005909090909,-83.666720902951 32.75969281274174,-83.66879780351513 32.75928015456466,-83.67086110700251 32.758821403399814,-83.67290937847787 32.75831687924383,-83.67494119511315 32.75776693105163,-83.67695514665314 32.75717194084009,-83.67894983308719 32.756532323629834,-83.68092387070277 32.75584852243937,-83.68287588719608 32.755121013232745,-83.68480452667785 32.75435029997188,-83.6867084511868 32.75353691874934,-83.68858633871054 32.752681435518205,-83.69043688423329 32.75178444323992,-83.69225880474181 32.75084656496929,-83.69405083421947 32.74986845274852,-83.69581173063119 32.74885078644311,-83.69754027103403 32.74779427211197,-83.69923525541884 32.746699642822385,-83.70089550880579 32.74556766051264,-83.70251987926525 32.74439911017125,-83.70410724071183 32.74319480286376,-83.70565649104168 32.74195557654766,-83.70716655737118 32.740682290251605,-83.70863639277735 32.73937582791714,-83.71006497711272 32.738037097583764,-83.71145132142914 32.73666702917708,-83.71279446471812 32.735266572878935,-83.71409347705406 32.73383670052444,-83.71534745831372 32.73237840618403,-83.71655554062085 32.73089270080836,-83.71771688788061 32.729380616419256,-83.7188306961288 32.72784320203518,-83.71989619434675 32.7262815256503,-83.72091264550903 32.72469667027635,-83.72187934768948 32.72308973592171,-83.72279563289698 32.721461838543654,-83.72366086608594 32.71981410620665,-83.72447445225765 32.718147683855385,-83.72523582737989 32.71646372644627,-83.72594446851502 32.71476340406971,-83.72659988462323 32.713047893684596,-83.7272016266906 32.71131838726742,-83.72774927777567 32.70957608482731,-83.72824246383601 32.70782219347475,-83.72868084389114 32.70605793103041,-83.72906411790967 32.70428451962234,-83.7293920229094 32.7025031912739,-83.72966433589676 32.700715179871075,-83.7298808698651 32.69892172546988,-83.7300414788296 32.69712407208444,-83.73014605577549 32.69532346570795,-83.73019452970033 32.693521154312656,-83.73018687161033 32.691718388897606,-83.73012309050767 32.689916417551466,-83.7300032323846 32.68811649010913,-83.7298273852511 32.68631985477568,-83.72959567309675 32.684527756380156,-83.72930826092573 32.68274143695762,-83.72896535074773 32.680962133537285,-83.72856718354245 32.6791910811693,-83.72811403832833 32.67742950582342,-83.72760623312291 32.67567862846354,-83.72704412086836 32.67393966108957,-83.72642809662005 32.672213810695574,-83.72575858936847 32.670502271353506,-83.72503606605545 32.66880622898625,-83.72426102976982 32.66712685869006,-83.7234340194563 32.66546532333759,-83.72255561311698 32.66382277497487,-83.7216264197786 32.66220034969901,-83.72064708740885 32.66059917243128,-83.7196182950583 32.659020351096295,-83.71854075868117 32.65746497883392,-83.71741522729366 32.655934134581365,-83.71624248192614 32.65442887527333,-83.71502333853363 32.65295024597017,-83.71375864112719 32.65149926868198,-83.71244926771118 32.650076949469984,-83.7110961292355 32.64868427122874,-83.70970016179575 32.64732219892485,-83.70826233431666 32.64599167470751,-83.70678364482677 32.6446936215174,-83.70526511731526 32.64342893528683,-83.70370780580646 32.642198493088834,-83.70211278830624 32.64100314591942,-83.70048117076016 32.639843721724354,-83.5825305152402 32.7316823944815))": true,
}

var skipUnion = map[string]bool{
	// Cause simplefeatures DCEL operations to fail with "no rings" error. See
	// https://github.com/peterstace/simplefeatures/pull/497 for details.
	"POLYGON((-83.58253051 32.73168239,-83.59843118 32.74617142,-83.70048117 32.63984372,-83.58253051 32.73168239))": true,
	"POLYGON((-83.70047745 32.63984661,-83.68891846 32.5989632,-83.58253417 32.73167955,-83.70047745 32.63984661))":  true,
	"POINT(456567.479 3973182.99)":  true,
	"POINT(333673.327 6252387.751)": true,
}

func checkDCELOperations(g1, g2 geom.Geometry, log *log.Logger) error {
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
			geom.Union,
			rawgeos.Union,
			skipUnion,
		},
		{
			"Intersection",
			geom.Intersection,
			rawgeos.Intersection,
			skipIntersection,
		},
		{
			"Difference",
			geom.Difference,
			rawgeos.Difference,
			skipDifference,
		},
		{
			"SymmetricDifference",
			geom.SymmetricDifference,
			rawgeos.SymmetricDifference,
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
	return checkRelate(g1, g2, log)
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
		return err
	}
	return checkEqualityHeuristic(want, got, log)
}

// checkEqualityHeuristic checks some necessary but not sufficient properties
// of two geometries if they are to be equal.
func checkEqualityHeuristic(want, got geom.Geometry, log *log.Logger) error {
	symDiff, err := rawgeos.SymmetricDifference(want, got)
	if err != nil {
		return err
	}
	symDiffArea := symDiff.Area()

	floatEq := float64EqualityChecker{
		absoluteThreshold: 1e-3,
		relativeThreshold: 1e-3,
	}.eq

	wantArea := want.Area()
	gotArea := got.Area()

	if !floatEq(symDiffArea, 0) {
		log.Printf("wantWKT: %v\n", want.AsText())
		log.Printf("gotWKT:  %v\n", got.AsText())
		log.Printf("wantArea: %v\n", wantArea)
		log.Printf("gotArea:  %v\n", gotArea)
		log.Printf("wantSymDiffArea: %v\n", 0)
		log.Printf("gotSymDiffArea:  %v\n", symDiffArea)
		return errMismatch
	}
	return nil
}

type float64EqualityChecker struct {
	absoluteThreshold float64
	relativeThreshold float64
}

func (c float64EqualityChecker) eq(a, b float64) bool {
	absDiff := math.Abs(a - b)
	magnitude := math.Max(math.Abs(a), math.Abs(b))
	return absDiff < c.absoluteThreshold || absDiff < magnitude*c.relativeThreshold
}

func checkRelate(g1, g2 geom.Geometry, log *log.Logger) error {
	got, err := geom.Relate(g1, g2)
	if err != nil {
		return err
	}
	want, err := rawgeos.Relate(g1, g2)
	if err != nil {
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

func checkRelateMatch(log *log.Logger) error {
	for i := 0; i < 1_000_000; i++ {
		mat := rand9("F012")
		pat := rand9("F012T*")
		want, err := rawgeos.RelatePatternMatch(mat, pat)
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
