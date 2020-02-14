package main

import (
	"bytes"
	"errors"
	"log"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/libgeos"
)

func unaryChecks(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	if valid, err := checkIsValid(h, g, log); err != nil {
		return err
	} else if !valid {
		return nil
	}

	log.Println("checking AsText")
	if err := checkAsText(h, g, log); err != nil {
		return err
	}
	log.Println("checking FromText")
	if err := checkFromText(h, g, log); err != nil {
		return err
	}
	log.Println("checking AsBinary")
	if err := checkAsBinary(h, g, log); err != nil {
		return err
	}
	log.Println("checking FromBinary")
	if err := checkFromBinary(h, g, log); err != nil {
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
	return nil

	//Area       float64
	//Cetroid    geom.Geometry
	//Reverse    geom.Geometry
}

var mismatchErr = errors.New("mismatch")

func checkIsValid(h *libgeos.Handle, g geom.Geometry, log *log.Logger) (bool, error) {
	var wkb bytes.Buffer
	if err := g.AsBinary(&wkb); err != nil {
		return false, err
	}
	var validAsPerSimpleFeatures bool
	if _, err := geom.UnmarshalWKB(&wkb); err == nil {
		validAsPerSimpleFeatures = true
	}
	log.Printf("Valid as per simplefeatures: %v", validAsPerSimpleFeatures)

	validAsPerLibgeos, err := h.IsValid(g)
	if err != nil {
		// The geometry is _so_ invalid that libgeos can't even tell if it's
		// invalid or not.
		validAsPerLibgeos = false
	}
	log.Printf("Valid as per libgeos: %v", validAsPerLibgeos)

	if validAsPerLibgeos != validAsPerSimpleFeatures {
		return false, mismatchErr
	}
	return validAsPerSimpleFeatures, nil
}

func checkAsText(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.AsText(g)
	if err != nil {
		return err
	}

	// Account for acceptable spacing differeneces between libgeos and simplefeatures.
	want = strings.ReplaceAll(want, " (", "(")
	want = strings.ReplaceAll(want, ", ", ",")

	got := g.AsText()
	if got != want {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return mismatchErr
	}
	return nil
}

func checkFromText(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	wkt := g.AsText()

	want, err := h.FromText(wkt)
	if err != nil {
		return err
	}

	got, err := geom.UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		return err
	}

	if !got.EqualsExact(want) {
		log.Printf("want: %v", want.AsText())
		log.Printf("got:  %v", got.AsText())
		return mismatchErr
	}
	return nil
}

func checkAsBinary(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	var wantDefined bool
	want, err := h.AsBinary(g)
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

	var got bytes.Buffer
	if err := g.AsBinary(&got); err != nil {
		return err
	}
	if bytes.Compare(want, got.Bytes()) != 0 {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errors.New("mismatch")
	}
	return nil
}

func hasEmptyPoint(g geom.Geometry) bool {
	if g.AsText() == "POINT EMPTY" {
		return true
	}
	// TODO: Should also support MultiPoints here. However, simplefeatures
	// doesn't support empty points in multipoint collections. We'll need to
	// update this when we add support for that.
	if !g.IsGeometryCollection() {
		return false
	}
	gc := g.AsGeometryCollection()
	for i := 0; i < gc.NumGeometries(); i++ {
		if hasEmptyPoint(gc.GeometryN(i)) {
			return true
		}
	}
	return false
}

func checkFromBinary(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	var wkb bytes.Buffer
	if err := g.AsBinary(&wkb); err != nil {
		return err
	}

	want, err := h.FromBinary(wkb.Bytes())
	if err != nil {
		return err
	}

	got, err := geom.UnmarshalWKB(bytes.NewReader(wkb.Bytes()))
	if err != nil {
		return err
	}

	if !want.EqualsExact(got) {
		return errors.New("mismatch")
	}
	return nil
}

func checkIsEmpty(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.IsEmpty(g)
	if err != nil {
		return err
	}
	got := g.IsEmpty()

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got: %v", got)
		return errors.New("mismatch")
	}
	return nil
}

func checkDimension(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	var want int
	if !(g.IsGeometryCollection() &&
		g.AsGeometryCollection().NumGeometries() == 0) {
		// Libgeos gives -1 dimension for GeometryCollections with zero
		// elements. This is very weird behaviour, and the dimension should
		// actually be zero. So we don't get 'want' from libgeos in that case.
		var err error
		want, err = h.Dimension(g)
		if err != nil {
			return err
		}
	}
	got := g.Dimension()

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got: %v", got)
		return errors.New("mismatch")
	}
	return nil
}

func checkEnvelope(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, wantDefined, err := h.Envelope(g)
	if err != nil {
		return err
	}
	got, gotDefined := g.Envelope()

	if wantDefined != gotDefined {
		log.Println("disagreement about envelope being defined")
		log.Printf("simplefeatures: %v", gotDefined)
		log.Printf("libgeos: %v", wantDefined)
		return errors.New("mismatch")
	}

	if !wantDefined {
		return nil
	}
	if want.Min() != got.Min() || want.Max() != got.Max() {
		log.Printf("want: %v", want.AsGeometry().AsText())
		log.Printf("got:  %v", got.AsGeometry().AsText())
		return errors.New("mismatch")
	}
	return nil
}

func checkIsSimple(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, wantDefined, err := h.IsSimple(g)
	if err != nil {
		return err
	}
	got, gotDefined := g.IsSimple()

	if wantDefined != gotDefined {
		log.Printf("want defined: %v", wantDefined)
		log.Printf("got defined: %v", gotDefined)
		return errors.New("mismatch")
	}
	if !gotDefined {
		return nil
	}
	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errors.New("mismatch")
	}
	return nil
}

func checkBoundary(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, wantDefined, err := h.Boundary(g)
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

	// PostGIS and libgeos have different behaviour for Boundary.
	// Simplefeatures currently uses the PostGIS behaviour (the difference in
	// behaviour has to do with the geometry type of empty geometries).
	if got.IsEmpty() && want.IsEmpty() {
		return nil
	}

	if !want.EqualsExact(got, geom.IgnoreOrder) {
		log.Printf("want: %v", want.AsText())
		log.Printf("got:  %v", got.AsText())
		return errors.New("mismatch")
	}
	return nil
}

func checkConvexHull(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.ConvexHull(g)
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

	if !want.EqualsExact(got, geom.IgnoreOrder) {
		log.Printf("want: %v", want.AsText())
		log.Printf("got:  %v", got.AsText())
		return errors.New("mismatch")
	}
	return nil
}

func checkIsRing(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.IsRing(g)
	if err != nil {
		return err
	}
	got := g.IsLineString() && g.AsLineString().IsRing()

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errors.New("mismatch")
	}
	return nil
}

func checkLength(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.Length(g)
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

	if want != got {
		log.Printf("want: %v", want)
		log.Printf("got:  %v", got)
		return errors.New("mismatch")
	}
	return nil
}

func isArealGeometry(g geom.Geometry) bool {
	switch {
	case g.IsPolygon() || g.IsMultiPolygon():
		return true
	case g.IsGeometryCollection():
		gc := g.AsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if isArealGeometry(gc.GeometryN(i)) {
				return true
			}
		}
	}
	return false
}
