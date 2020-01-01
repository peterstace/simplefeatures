package geom_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestGeoJSONFeatureCollectionValidUnmarshal(t *testing.T) {
	const input = `
	{
	   "type": "FeatureCollection",
	   "features": [
		   {
			   "type": "Feature",
			   "id": "id0",
			   "geometry": {
				   "type": "LineString",
				   "coordinates": [
					   [102.0, 0.0],
					   [103.0, 1.0],
					   [104.0, 0.0],
					   [105.0, 1.0]
				   ]
			   },
			   "properties": {
				   "prop0": "value0",
				   "prop1": "value1"
			   }
		   },
		   {
			   "type": "Feature",
			   "id": "id1",
			   "geometry": {
				   "type": "Polygon",
				   "coordinates": [
					   [
						   [100.0, 0.0],
						   [101.0, 0.0],
						   [101.0, 1.0],
						   [100.0, 1.0],
						   [100.0, 0.0]
					   ]
				   ]
			   },
			   "properties": {
				   "prop0": "value2",
				   "prop1": "value3"
			   }
		   }
	   ]
	}`

	var fc GeoJSONFeatureCollection
	err := json.NewDecoder(strings.NewReader(input)).Decode(&fc)
	expectNoErr(t, err)

	expectDeepEqual(t, len(fc), 2)
	f0 := fc[0]
	f1 := fc[1]

	expectDeepEqual(t, f0.ID, "id0")
	expectDeepEqual(t, f0.Properties, map[string]interface{}{"prop0": "value0", "prop1": "value1"})
	expectDeepEqual(t, f0.GeometryX, geomFromWKT(t, "LINESTRING(102 0,103 1,104 0,105 1)"))

	expectDeepEqual(t, f1.ID, "id1")
	expectDeepEqual(t, f1.Properties, map[string]interface{}{"prop0": "value2", "prop1": "value3"})
	expectDeepEqual(t, f1.GeometryX, geomFromWKT(t, "POLYGON((100 0,101 0,101 1,100 1,100 0))"))
}

func TestGeoJSONFeatureCollectionInvalidUnmarshal(t *testing.T) {
	for i, tt := range []struct {
		input       string
		errFragment string
	}{
		{
			// Valid case.
			input: `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]}}]}`,
		},
		{
			input:       `{"type":"Foo","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]}}]}`,
			errFragment: "type field not set to FeatureCollection",
		},
		{
			input:       `{"type":"FeatureCollection","features":[{"type":"Foo","geometry":{"type":"Point","coordinates":[1,2]}}]}`,
			errFragment: "type field not set to Feature",
		},
		{
			input:       `{"type":"FeatureCollection","features":[{"geometry":{"type":"Point","coordinates":[1,2]}}]}`,
			errFragment: "feature type field missing or empty",
		},
		{
			input:       `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"ZORT","coordinates":[1,2]}}]}`,
			errFragment: "unknown geojson type: ZORT",
		},
		{
			input:       `{"type":"FeatureCollection","features":[{"type":"Feature"}]}`,
			errFragment: "geometry field missing or empty",
		},
		{
			input: `{"type":"FeatureCollection","features":[{"type":"Feature","properties":"zoortle","geometry":{"type":"Point","coordinates":[1,2]}}]}`,
			// This error message is from the Go standard lib, so don't want to
			// string match the error too closely. Since it contains
			// 'features', we know it's about the features field.
			errFragment: "features",
		},
	} {
		if i == 0 {
			// Ensure that the first feature collection is valid, since that's
			// what the other test cases are based on.
			var fc GeoJSONFeatureCollection
			r := strings.NewReader(tt.input)
			expectNoErr(t, json.NewDecoder(r).Decode(&fc))
			continue
		}
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var fc GeoJSONFeatureCollection
			r := strings.NewReader(tt.input)
			err := json.NewDecoder(r).Decode(&fc)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.errFragment) {
				t.Errorf("expected to contain '%s' but got '%s'", tt.errFragment, err.Error())
			}
		})
	}
}

func TestGeoJSONFeatureCollectionEmpty(t *testing.T) {
	out, err := json.Marshal(GeoJSONFeatureCollection{})
	expectNoErr(t, err)
	expectDeepEqual(t, string(out), `{"type":"FeatureCollection","features":[]}`)
}

func TestGeoJSONFeatureCollectionNil(t *testing.T) {
	out, err := json.Marshal(GeoJSONFeatureCollection(nil))
	expectNoErr(t, err)
	expectDeepEqual(t, string(out), `{"type":"FeatureCollection","features":[]}`)
}

func TestGeoJSONFeatureCollectionNilGeometryX(t *testing.T) {
	if _, err := json.Marshal(GeoJSONFeatureCollection{{}}); err == nil {
		t.Error("expected error but got nil")
	}
}

func TestGeoJSONFeatureCollectionAndPropertiesNil(t *testing.T) {
	out, err := json.Marshal(GeoJSONFeatureCollection{{GeometryX: geomFromWKT(t, "POINT(1 2)")}})
	expectNoErr(t, err)
	expectDeepEqual(t, string(out), `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"properties":{}}]}`)
}

func TestGeoJSONFeatureCollectionAndPropertiesSet(t *testing.T) {
	out, err := json.Marshal(GeoJSONFeatureCollection{{
		GeometryX: geomFromWKT(t, "POINT(1 2)"),
		ID:       "myid",
		Properties: map[string]interface{}{
			"foo": "bar",
		},
	}})
	expectNoErr(t, err)
	expectDeepEqual(t, string(out), `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"id":"myid","properties":{"foo":"bar"}}]}`)
}
