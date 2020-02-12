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
	return nil

	//AsGeoJSON  sql.NullString
	//Envelope   geom.Geometry
	//IsSimple   sql.NullBool
	//Boundary   geom.NullGeometry
	//ConvexHull geom.Geometry
	//IsValid    bool
	//IsRing     sql.NullBool
	//Length     float64
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
