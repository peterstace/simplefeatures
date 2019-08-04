package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func CheckWKTParse(t *testing.T, pg PostGIS, candidates []string) {
	var any bool
	for i, wkt := range candidates {
		any = true
		t.Run(fmt.Sprintf("CheckWKTParse_%d", i), func(t *testing.T) {

			// The simple feature library accepts LINEARRING WKTs. However,
			// postgis doesn't accept them. A workaround for this is to just
			// substitute LINEARRING for LINESTRING. However, this will give a
			// false negative if the corpus contains a LINEARRING WKT that
			// isn't closed (and thus won't be accepted by simple features).
			wkt := strings.ReplaceAll(wkt, "LINEARRING", "LINESTRING")

			_, sfErr := geom.UnmarshalWKT(strings.NewReader(wkt))
			isValid, reason := pg.WKTIsValidWithReason(t, wkt)
			if (sfErr == nil) != isValid {
				t.Logf("WKT: %v", wkt)
				t.Logf("SimpleFeatures err: %v", sfErr)
				t.Logf("PostGIS IsValid: %v", isValid)
				t.Logf("PostGIS Reason: %v", reason)
				t.Errorf("mismatch")
			}
		})
	}
	if !any {
		// We know there are some some valid WKT strings, so if this happens
		// then something is wrong with the extraction or conversion logic.
		t.Errorf("could not extract any WKTs")
	}
}

func CheckWKBParse(t *testing.T, pg PostGIS, candidates []string) {
	var any bool
	for i, wkb := range candidates {
		buf, err := hexStringToBytes(wkb)
		if err != nil {
			continue
		}
		any = true
		t.Run(fmt.Sprintf("CheckWKBParse_%d", i), func(t *testing.T) {
			_, sfErr := geom.UnmarshalWKB(bytes.NewReader(buf))
			isValid, reason := pg.WKBIsValidWithReason(t, wkb)
			if (sfErr == nil) != isValid {
				t.Logf("WKB: %v", wkb)
				t.Logf("SimpleFeatures err: %v", sfErr)
				t.Logf("PostGIS IsValid: %v", isValid)
				t.Logf("PostGIS Reason: %v", reason)
				t.Errorf("mismatch")
			}
		})
	}
	if !any {
		// We know there are some some valid hex strings, so if this happens
		// then something is wrong with the extraction or conversion logic.
		t.Errorf("could not extract any WKBs")
	}
}

func hexStringToBytes(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errors.New("hex string must have even length")
	}
	var buf []byte
	for i := 0; i < len(s); i += 2 {
		x, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return nil, err
		}
		buf = append(buf, byte(x))
	}
	return buf, nil
}

func CheckGeoJSONParse(t *testing.T, pg PostGIS, candidates []string) {
	var any bool
	for i, geojson := range candidates {
		if geojson == `{"type":"Point","coordinates":[]}` {
			// From https://tools.ietf.org/html/rfc7946#section-3.1:
			//
			// > GeoJSON processors MAY interpret Geometry objects with
			// > empty "coordinates" arrays as null objects.
			//
			// Simplefeatures chooses to accept this as an empty point, but
			// Postgres rejects it.
			continue
		}
		any = true
		t.Run(fmt.Sprintf("CheckGeoJSONParse_%d", i), func(t *testing.T) {
			_, sfErr := geom.UnmarshalGeoJSON([]byte(geojson))
			isValid, reason := pg.GeoJSONIsValidWithReason(t, geojson)
			if (sfErr == nil) != isValid {
				t.Logf("GeoJSON: %v", geojson)
				t.Logf("SimpleFeatures err: %v", sfErr)
				t.Logf("PostGIS IsValid: %v", isValid)
				t.Logf("PostGIS Reason: %v", reason)
				t.Errorf("mismatch")
			}
		})
	}
	if !any {
		// We know there are some some valid geojson strings, so if this happens
		// then something is wrong with the extraction or conversion logic.
		t.Errorf("could not extract any geojsons")
	}
}
