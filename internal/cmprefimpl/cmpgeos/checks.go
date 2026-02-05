package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
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
	// Skip any geometries that are collections or contain collections that
	// only contain empty geometries. Libgeos will render WKT for these
	// collections as being EMPTY, however this isn't correct behaviour.
	if containsCollectionWithOnlyEmptyElements(g) {
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
	want, err := rawgeos.AsBinary(g)
	if err != nil {
		return err
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
	want, err := rawgeos.Dimension(g)
	if err != nil {
		return err
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
	want, err := rawgeos.Boundary(g)
	if g.Type() == geom.TypeGeometryCollection {
		// libgeos doesn't define the boundary of GeometryCollections (it gives
		// an error), but simplefeatures does define a boundary. Explicitly
		// expect an error here, so that we can update these tests in case the
		// behaviour of libgeos changes.
		if err == nil {
			return errors.New("expected error for GeometryCollection boundary, got none")
		}
		return nil
	} else { //nolint:gocritic,revive // more readable with the current structure
		if err != nil {
			return err
		}
	}

	got := g.Boundary()

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
		// Skip the contains check for degenerate polygons with near-zero area,
		// since the Contains predicate is unreliable due to floating point
		// precision issues.
		if g.Area() < 1e-9 {
			return nil
		}
		contains, err := rawgeos.Contains(g, pt)
		if err != nil {
			return err
		}
		if !contains {
			log.Printf("the input doesn't contain the pt: %v", pt.AsText())
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

	// ...First, the areas are compared. Use both relative and absolute
	// tolerance since the areas can vary widely in magnitude.
	diff := math.Abs(wantArea - gotArea)
	maxArea := math.Max(math.Abs(wantArea), math.Abs(gotArea))
	const relTol = 1e-6
	const absTol = 1e-9
	if diff > relTol*maxArea && diff > absTol {
		log.Printf("areas differ beyond tolerance (rel=%v, abs=%v)", relTol, absTol)
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
	} {
		lg.Printf("checking %s", check.name)
		if err := check.fn(g1, g2, lg); err != nil {
			return err
		}
	}
	return nil
}

func checkIntersects(g1, g2 geom.Geometry, log *log.Logger) error {
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
	want, err := rawgeos.Distance(g1, g2)
	if err != nil {
		return err
	}
	got, ok := geom.Distance(g1, g2)
	if !ok {
		// GEOS gives 0 when distance is not defined.
		got = 0
	}

	if math.Abs(want-got) > 1e-9 {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errMismatch
	}
	return nil
}
