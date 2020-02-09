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

	log.Println("checking AsText forward")
	if err := checkAsTextForward(h, g, log); err != nil {
		return err
	}
	log.Println("checking AsText reverse")
	if err := checkAsTextReverse(h, g, log); err != nil {
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
	return nil

	//AsBinary   []byte
	//AsGeoJSON  sql.NullString
	//IsEmpty    bool
	//Dimension  int
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

func checkAsTextForward(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	wkt := g.AsText()
	gWKT, err := h.FromText(wkt)
	if err != nil {
		return err
	}
	log.Printf("libgeos FromText: %v", gWKT.AsText())
	if !gWKT.EqualsExact(g) {
		return mismatchErr
	}
	return nil
}

func checkAsTextReverse(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	wkt, err := h.AsText(g)
	if err != nil {
		return err
	}
	log.Printf("libgeos AsText: %v", wkt)
	gWKT, err := geom.UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		return err
	}
	log.Printf("unmarshalled geom via simplefeatures: %v", gWKT.AsText())
	if !gWKT.EqualsExact(g) {
		return mismatchErr
	}
	return nil
}

func checkAsBinary(h *libgeos.Handle, g geom.Geometry, log *log.Logger) error {
	want, err := h.AsBinary(g)
	if err != nil {
		return err
	}
	var got bytes.Buffer
	if err := g.AsBinary(&got); err != nil {
		return err
	}
	if bytes.Compare(want, got.Bytes()) != 0 {
		return errors.New("mismatch")
	}
	return nil
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
