package geom_test

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
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

	var fc geom.GeoJSONFeatureCollection
	err := json.NewDecoder(strings.NewReader(input)).Decode(&fc)
	expectNoErr(t, err)

	expectIntEq(t, len(fc.Features), 2)
	f0 := fc.Features[0]
	f1 := fc.Features[1]

	expectStringEq(t, f0.ID.(string), "id0")
	expectBoolEq(t, reflect.DeepEqual(f0.Properties, map[string]interface{}{"prop0": "value0", "prop1": "value1"}), true)
	expectGeomEq(t, f0.Geometry, geomFromWKT(t, "LINESTRING(102 0,103 1,104 0,105 1)"))

	expectStringEq(t, f1.ID.(string), "id1")
	expectBoolEq(t, reflect.DeepEqual(f1.Properties, map[string]interface{}{"prop0": "value2", "prop1": "value3"}), true)
	expectGeomEq(t, f1.Geometry, geomFromWKT(t, "POLYGON((100 0,101 0,101 1,100 1,100 0))"))
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
			errFragment: "feature type field missing",
		},
		{
			input:       `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"ZORT","coordinates":[1,2]}}]}`,
			errFragment: "unknown geometry type: 'ZORT'",
		},
		{
			input:       `{"type":"FeatureCollection","features":[{"type":"Feature"}]}`,
			errFragment: "geometry field missing",
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
			var fc geom.GeoJSONFeatureCollection
			r := strings.NewReader(tt.input)
			expectNoErr(t, json.NewDecoder(r).Decode(&fc))
			continue
		}
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var fc geom.GeoJSONFeatureCollection
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
	out, err := json.Marshal(geom.GeoJSONFeatureCollection{})
	expectNoErr(t, err)
	expectStringEq(t, string(out), `{"type":"FeatureCollection","features":[]}`)
}

func TestGeoJSONFeatureCollectionNil(t *testing.T) {
	out, err := json.Marshal(geom.GeoJSONFeatureCollection{
		Features:       nil,
		ForeignMembers: nil,
	})
	expectNoErr(t, err)
	expectStringEq(t, string(out), `{"type":"FeatureCollection","features":[]}`)
}

func TestGeoJSONFeatureCollectionAndPropertiesNil(t *testing.T) {
	out, err := json.Marshal(geom.GeoJSONFeatureCollection{Features: []geom.GeoJSONFeature{{Geometry: geomFromWKT(t, "POINT(1 2)")}}})
	expectNoErr(t, err)
	expectStringEq(t, string(out), `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"properties":{}}]}`)
}

func TestGeoJSONFeatureCollectionAndPropertiesSet(t *testing.T) {
	out, err := json.Marshal(geom.GeoJSONFeatureCollection{
		Features: []geom.GeoJSONFeature{{
			Geometry: geomFromWKT(t, "POINT(1 2)"),
			ID:       "myid",
			Properties: map[string]interface{}{
				"foo": "bar",
			},
		}},
	})
	expectNoErr(t, err)
	expectStringEq(t, string(out), `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"id":"myid","properties":{"foo":"bar"}}]}`)
}

func TestGeoJSONFeatureForeignMembers(t *testing.T) {
	for _, tc := range []struct {
		name    string
		members map[string]interface{}
		json    string
	}{
		{
			name:    "nil foreign members",
			members: nil,
			json:    `{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{}}`,
		},
		{
			name:    "empty foreign members",
			members: map[string]interface{}{},
			json:    `{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{}}`,
		},
		{
			name:    "one foreign member",
			members: map[string]interface{}{"foo": "bar"},
			json:    `{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{},"foo":"bar"}`,
		},
		{
			name:    "two foreign members",
			members: map[string]interface{}{"foo": "bar", "baz": 42.0},
			json:    `{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{},"baz":42,"foo":"bar"}`,
		},
		{
			name:    "nested",
			members: map[string]interface{}{"metadata": map[string]interface{}{"foo": "bar", "baz": 42.0}},
			json:    `{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{},"metadata":{"baz":42,"foo":"bar"}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("marshal", func(t *testing.T) {
				feat := geom.GeoJSONFeature{
					Geometry:       geom.XY{}.AsPoint().AsGeometry(),
					ForeignMembers: tc.members,
				}
				got, err := json.Marshal(feat)
				expectNoErr(t, err)
				expectStringEq(t, string(got), tc.json)
			})
			t.Run("unmarshal", func(t *testing.T) {
				var feat geom.GeoJSONFeature
				err := json.Unmarshal([]byte(tc.json), &feat)
				expectNoErr(t, err)
				if len(feat.ForeignMembers) != 0 || len(tc.members) != 0 {
					expectDeepEq(t, feat.ForeignMembers, tc.members)
				}
			})
		})
	}
}

func TestGeoJSONFeatureForeignMembersForbidden(t *testing.T) {
	for _, tc := range []struct {
		name    string
		members map[string]interface{}
	}{
		{
			name:    "has type foreign member",
			members: map[string]interface{}{"type": "dummy"},
		},
		{
			name:    "has geometry foreign member",
			members: map[string]interface{}{"geometry": "dummy"},
		},
		{
			name:    "has properties foreign member",
			members: map[string]interface{}{"properties": "dummy"},
		},
		{
			name:    "has id foreign member",
			members: map[string]interface{}{"id": "dummy"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			feat := geom.GeoJSONFeature{
				ForeignMembers: tc.members,
			}
			_, err := json.Marshal(feat)
			test.ErrAs(t, err, &geom.DisallowedForeignMemberError{})
		})
	}
}

func TestGeoJSONFeatureCollectionForeignMembers(t *testing.T) {
	for _, tc := range []struct {
		name    string
		members map[string]interface{}
		json    string
	}{
		{
			name:    "nil foreign members",
			members: nil,
			json:    `{"type":"FeatureCollection","features":[]}`,
		},
		{
			name:    "empty foreign members",
			members: map[string]interface{}{},
			json:    `{"type":"FeatureCollection","features":[]}`,
		},
		{
			name:    "one foreign member",
			members: map[string]interface{}{"foo": "bar"},
			json:    `{"type":"FeatureCollection","features":[],"foo":"bar"}`,
		},
		{
			name:    "two foreign members",
			members: map[string]interface{}{"foo": "bar", "baz": 42.0},
			json:    `{"type":"FeatureCollection","features":[],"baz":42,"foo":"bar"}`,
		},
		{
			name:    "nested",
			members: map[string]interface{}{"metadata": map[string]interface{}{"foo": "bar", "baz": 42.0}},
			json:    `{"type":"FeatureCollection","features":[],"metadata":{"baz":42,"foo":"bar"}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("marshal", func(t *testing.T) {
				fc := geom.GeoJSONFeatureCollection{
					ForeignMembers: tc.members,
				}
				got, err := json.Marshal(fc)
				expectNoErr(t, err)
				expectStringEq(t, string(got), tc.json)
			})
			t.Run("unmarshal", func(t *testing.T) {
				var fc geom.GeoJSONFeatureCollection
				err := json.Unmarshal([]byte(tc.json), &fc)
				expectNoErr(t, err)
				if len(fc.ForeignMembers) != 0 || len(tc.members) != 0 {
					expectDeepEq(t, fc.ForeignMembers, tc.members)
				}
			})
		})
	}
}

func TestGeoJSONFeatureCollectionForeignMembersForbidden(t *testing.T) {
	for _, tc := range []struct {
		name    string
		members map[string]interface{}
	}{
		{
			name:    "has type foreign member",
			members: map[string]interface{}{"type": "dummy"},
		},
		{
			name:    "has features foreign member",
			members: map[string]interface{}{"features": "dummy"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fc := geom.GeoJSONFeatureCollection{
				ForeignMembers: tc.members,
			}
			_, err := json.Marshal(fc)
			test.ErrAs(t, err, &geom.DisallowedForeignMemberError{})
		})
	}
}
