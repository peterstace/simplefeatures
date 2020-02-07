package main

import (
	"fmt"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/libgeos"
)

func unaryChecks(h *libgeos.Handle, g geom.Geometry) error {
	//AsText     string
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

	return checkAsText(h, g)
}

type mismatchError struct {
	want interface{}
	got  interface{}
}

func (e mismatchError) Error() string {
	return fmt.Sprintf("\nwant: %q\ngot:  %q", e.want, e.got)
}

func checkAsText(h *libgeos.Handle, g geom.Geometry) error {
	want, err := h.AsText(g)
	if err != nil {
		return err
	}
	got := g.AsText()
	if want != got {
		return mismatchError{want: want, got: got}
	}
	return nil
}
