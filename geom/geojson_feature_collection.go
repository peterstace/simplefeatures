package geom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// GeoJSONFeature represents a Geometry with associated free-form properties.
// GeoJSONFeature values have a one to one correspondence with GeoJSON Features.
type GeoJSONFeature struct {
	// Geometry is the geometry that is associated with the Feature.
	Geometry Geometry

	// ID is an identifier to refer to the feature. If an identifier isn't
	// applicable, ID can be left as nil. If it's set, then its value should
	// marshal into a JSON string or number (this is not enforced).
	ID interface{}

	// Properties are free-form properties that are associated with the
	// feature. If there are no properties associated with the feature, then it
	// can either be set to an empty map or left as nil.
	Properties map[string]interface{}

	// ForeignMembers are additional fields that are not explicitly described
	// in the GeoJSON specification, but are allowed (as per the specification)
	// to be present at the top level of GeoJSON features nonetheless.
	ForeignMembers map[string]interface{}
}

// UnmarshalJSON implements the encoding/json Unmarshaler interface by
// unmarshalling a GeoJSON Feature Collection object.
func (f *GeoJSONFeature) UnmarshalJSON(p []byte) error {
	var topLevel map[string]json.RawMessage
	if err := json.Unmarshal(p, &topLevel); err != nil {
		return err
	}

	typeJSON, ok := topLevel["type"]
	if !ok {
		return errors.New("feature type field missing")
	}
	var typeStr string
	if err := json.Unmarshal(typeJSON, &typeStr); err != nil {
		return err
	}
	if typeStr != "Feature" {
		return fmt.Errorf("type field not set to Feature: '%s'", typeStr)
	}

	gJSON, ok := topLevel["geometry"]
	if !ok {
		return errors.New("geometry field missing")
	}
	var g Geometry
	if err := json.Unmarshal(gJSON, &g); err != nil {
		return err
	}

	idJSON, ok := topLevel["id"]
	var id interface{}
	if ok {
		if err := json.Unmarshal(idJSON, &id); err != nil {
			return err
		}
	}

	propsJSON, ok := topLevel["properties"]
	var props map[string]interface{}
	if ok {
		if err := json.Unmarshal(propsJSON, &props); err != nil {
			return err
		}
	}

	foreignMembers := make(map[string]interface{})
	for k, vJSON := range topLevel {
		switch k {
		case "type", "geometry", "id", "properties":
			continue
		default:
			var v interface{}
			if err := json.Unmarshal(vJSON, &v); err != nil {
				return err
			}
			foreignMembers[k] = v
		}
	}

	*f = GeoJSONFeature{
		Geometry:       g,
		ID:             id,
		Properties:     props,
		ForeignMembers: foreignMembers,
	}
	return nil
}

// MarshalJSON implements the encoding/json Marshaler interface by marshalling
// into a GeoJSON FeatureCollection object.
func (f GeoJSONFeature) MarshalJSON() ([]byte, error) {
	props := f.Properties
	if props == nil {
		// As per the GeoJSON spec, the properties field must be an object (not null).
		props = map[string]interface{}{}
	}

	buf, err := json.Marshal(struct {
		Type       string                 `json:"type"`
		Geometry   Geometry               `json:"geometry"`
		ID         interface{}            `json:"id,omitempty"`
		Properties map[string]interface{} `json:"properties"`
	}{
		"Feature",
		f.Geometry,
		f.ID,
		props,
	})
	if err != nil {
		return nil, err
	}

	if len(f.ForeignMembers) == 0 {
		return buf, nil
	}
	fms, err := json.Marshal(f.ForeignMembers)
	if err != nil {
		return nil, err
	}
	if len(fms) == 0 || fms[0] != '{' {
		return nil, errors.New("ForeignMembers must marshal to a JSON object")
	}
	if bytes.Equal(fms, []byte("{}")) {
		// {} is a special case due to the ',' that would be added below.
		return buf, nil
	}
	buf = buf[:len(buf)-1] // remove trailing '}'
	buf = append(buf, ',')
	buf = append(buf, fms[1:]...) // skip leading '{'
	return buf, nil
}

// GeoJSONFeatureCollection is a collection of GeoJSONFeatures.
// GeoJSONFeatureCollection values have a one to one correspondence with
// GeoJSON FeatureCollections.
type GeoJSONFeatureCollection []GeoJSONFeature

// UnmarshalJSON implements the encoding/json Unmarshaler interface by
// unmarshalling a GeoJSON FeatureCollection object.
func (c *GeoJSONFeatureCollection) UnmarshalJSON(p []byte) error {
	var topLevel struct {
		Type     string           `json:"type"`
		Features []GeoJSONFeature `json:"features"`
	}
	if err := json.Unmarshal(p, &topLevel); err != nil {
		return err
	}
	if topLevel.Type == "" {
		return errors.New("feature collection type field missing or empty")
	}
	if topLevel.Type != "FeatureCollection" {
		return fmt.Errorf("type field not set to FeatureCollection: '%s'", topLevel.Type)
	}
	*c = topLevel.Features
	return nil
}

// MarshalJSON implements the encoding/json Marshaler interface by marshalling
// into a GeoJSON FeatureCollection object.
func (c GeoJSONFeatureCollection) MarshalJSON() ([]byte, error) {
	var col []GeoJSONFeature = c
	if col == nil {
		col = []GeoJSONFeature{}
	}
	return json.Marshal(struct {
		Type     string           `json:"type"`
		Features []GeoJSONFeature `json:"features"`
	}{"FeatureCollection", col})
}
