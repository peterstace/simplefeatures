package simplefeatures_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
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
	expectDeepEqual(t, f0.Geometry, geomFromWKT(t, "LINESTRING(102 0,103 1,104 0,105 1)"))

	expectDeepEqual(t, f1.ID, "id1")
	expectDeepEqual(t, f1.Properties, map[string]interface{}{"prop0": "value2", "prop1": "value3"})
	expectDeepEqual(t, f1.Geometry, geomFromWKT(t, "POLYGON((100 0,101 0,101 1,100 1,100 0))"))
}

func TestGeoJSONFeatureCollectionInvalidUnmarshal(t *testing.T) {
	for i, tt := range []struct {
		input       string
		errFragment string
	}{
		{
			input:       `{"type":"foo","features":[{"type":"Point","coordinates":[1,2]}]}`,
			errFragment: "type field isn't set to FeatureCollection",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var fc GeoJSONFeatureCollection
			err := json.NewDecoder(strings.NewReader(tt.input)).Decode(&fc)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.errFragment) {
				t.Errorf("expected to contain '%s' but got '%s'", tt.errFragment, err.Error())
			}
		})
	}
}
