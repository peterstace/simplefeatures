package main

import (
	"bytes"
	"fmt"
	"strings"

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

	// TODO: Check is valid before doing anything at all.

	if valid, err := checkIsValid(h, g); err != nil {
		return err
	} else if !valid {
		return nil
	}

	if err := checkAsTextForward(h, g); err != nil {
		return err
	}
	if err := checkAsTextReverse(h, g); err != nil {
		return err
	}
	return nil
}

type mismatchError struct {
	operation string
	operands  []interface{}
	want      interface{}
	got       interface{}
}

func (e mismatchError) Error() string {
	var buf strings.Builder
	buf.WriteByte('\n')
	fmt.Fprintf(&buf, "\toperation: %v\n", e.operation)
	fmt.Fprintf(&buf, "\toperands: %d\n", len(e.operands))
	for i, o := range e.operands {
		fmt.Fprintf(&buf, "\t\t[%d]: %v\n", i, o)
	}
	fmt.Fprintf(&buf, "\twant: %s\n", e.want)
	fmt.Fprintf(&buf, "\tgot:  %s\n", e.got)
	return buf.String()
}

func checkIsValid(h *libgeos.Handle, g geom.Geometry) (bool, error) {
	var wkb bytes.Buffer
	if err := g.AsBinary(&wkb); err != nil {
		return false, err
	}
	var validAsPerSimpleFeatures bool
	if _, err := geom.UnmarshalWKB(&wkb); err == nil {
		validAsPerSimpleFeatures = true
	}

	validAsPerLibgeos, err := h.IsValid(g)
	if err != nil {
		// The geometry is _so_ invalid that libgeos can't even tell if it's
		// invalid or not.
		validAsPerLibgeos = false
	}

	if validAsPerLibgeos != validAsPerSimpleFeatures {
		return false, fmt.Errorf("validAsPerLibgeos:%v "+
			"validAsPerSimpleFeatures:%v",
			validAsPerLibgeos,
			validAsPerSimpleFeatures,
		)
	}
	return validAsPerSimpleFeatures, nil
}

func checkAsTextForward(h *libgeos.Handle, g geom.Geometry) error {
	wkt := g.AsText()
	gWKT, err := h.FromText(wkt)
	if err != nil {
		return err
	}
	if !gWKT.EqualsExact(g) {
		return mismatchError{
			operation: "AsText_Forward",
			operands:  []interface{}{g},
			want:      wkt,
			got:       gWKT.AsText(),
		}
	}
	return nil
}

func checkAsTextReverse(h *libgeos.Handle, g geom.Geometry) error {
	wkt, err := h.AsText(g)
	if err != nil {
		return err
	}
	gWKT, err := geom.UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		return err
	}
	if !gWKT.EqualsExact(g) {
		return mismatchError{
			operation: "AsText_Reverse",
			operands:  []interface{}{g},
			want:      g.AsText(),
			got:       gWKT.AsText(),
		}
	}
	return nil
}
