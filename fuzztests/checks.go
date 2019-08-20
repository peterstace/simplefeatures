package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
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

func CheckWKT(t *testing.T, pg PostGIS, g geom.Geometry) {
	t.Run("CheckWKT", func(t *testing.T) {
		got := g.AsText()
		if strings.Contains(got, "MULTIPOINT") {
			// Skip Multipoints. This is because Postgis doesn't follow the SFA
			// spec by not including parenthesis around each individual point.
			// The simplefeatures library follows the spec correctly.
			return
		}
		want := pg.AsText(t, g)
		if want != got {
			t.Logf("got:  %v", got)
			t.Logf("want: %v", want)
			t.Error("mismatch")
		}
	})
}

func CheckWKB(t *testing.T, pg PostGIS, g geom.Geometry) {
	t.Run("CheckWKB", func(t *testing.T) {
		if _, ok := g.(geom.EmptySet); ok && g.AsText() == "POINT EMPTY" {
			// Empty point WKB use NaN as part of their representation.
			// Go's math.NaN() and Postgis use slightly different (but
			// compatible) representations of NaN.
			return
		}
		if _, ok := g.(geom.GeometryCollection); ok && g.IsEmpty() {
			// The behaviour for GeometryCollections in Postgis is to just
			// give 'GEOMETRYCOLLECTION EMPTY' whenever the contents of a
			// geometry collection are all empty geometries. This doesn't
			// seem like correct behaviour, so these cases are skipped.
			return
		}
		var got bytes.Buffer
		if err := g.AsBinary(&got); err != nil {
			t.Fatalf("writing wkb: %v", err)
		}
		want := pg.AsBinary(t, g)
		if !bytes.Equal(got.Bytes(), want) {
			t.Logf("got:  %v", got.Bytes())
			t.Logf("want: %v", want)
			t.Error("mismatch")
		}
	})
}

func CheckGeoJSON(t *testing.T, pg PostGIS, g geom.Geometry) {
	t.Run("CheckGeoJSON", func(t *testing.T) {

		// PostGIS cannot convert to geojson in the case where it has
		// nested geometry collections. That seems to be based on the
		// following section of https://tools.ietf.org/html/rfc7946:
		//
		// To maximize interoperability, implementations SHOULD avoid
		// nested GeometryCollections.  Furthermore, GeometryCollections
		// composed of a single part or a number of parts of a single type
		// SHOULD be avoided when that single part or a single object of
		// multipart type (MultiPoint, MultiLineString, or MultiPolygon)
		// could be used instead.
		if gc, ok := g.(geom.GeometryCollection); ok {
			for i := 0; i < gc.NumGeometries(); i++ {
				if _, ok := gc.GeometryN(i).(geom.GeometryCollection); ok {
					return
				}
			}
		}

		got, err := g.MarshalJSON()
		if err != nil {
			t.Fatalf("could not convert to geojson: %v", err)
		}
		want := pg.AsGeoJSON(t, g)
		if !bytes.Equal(got, want) {
			t.Logf("got:  %v", string(got))
			t.Logf("want: %v", string(want))
			t.Error("mismatch")
		}
	})
}

func CheckIsEmpty(t *testing.T, pg PostGIS, g geom.Geometry) {
	t.Run("CheckIsEmpty", func(t *testing.T) {
		got := g.IsEmpty()
		want := pg.IsEmpty(t, g)
		if got != want {
			t.Logf("got:  %v", got)
			t.Logf("want: %v", want)
			t.Error("mismatch")
		}
	})
}

func CheckDimension(t *testing.T, pg PostGIS, g geom.Geometry) {
	t.Run("CheckDimension", func(t *testing.T) {
		got := g.Dimension()
		want := pg.Dimension(t, g)
		if got != want {
			t.Logf("got:  %v", got)
			t.Logf("want: %v", want)
			t.Error("mismatch")
		}
	})
}

func CheckEnvelope(t *testing.T, pg PostGIS, g geom.Geometry) {
	t.Run("CheckEnvelope", func(t *testing.T) {
		if g.IsEmpty() {
			// PostGIS allows envelopes on empty geometries, but they are empty
			// envelopes. In simplefeatures, an envelope is never empty, so we
			// skip testing that case.
			return
		}
		env, ok := g.Envelope()
		if !ok {
			// We just checked IsEmpty, so this should never happen.
			panic("could not get envelope")
		}
		got := env.AsGeometry()
		want := pg.Envelope(t, g)

		if !reflect.DeepEqual(got, want) {
			t.Logf("got:  %v", got.AsText())
			t.Logf("want: %v", want.AsText())
			t.Error("mismatch")
		}
	})
}
