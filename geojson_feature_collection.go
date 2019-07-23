package simplefeatures

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GeoJSONFeature represents a Geometry with associated free-form properties.
type GeoJSONFeature struct {
	// Geometry is the Geometry that is associated with the Feature. It must be
	// populated.
	Geometry Geometry

	// ID is an identifier to refer to the feature. If an identifier isn't
	// applicable, ID can be left as nil. If it is set, then it's value should
	// marshal into a JSON string or number (this is not enforced).
	ID interface{}

	// Properties are free-form properties that are associated with the
	// feature. It can either be left as nil, or populated with any value that
	// can be marshalled to JSON.
	Properties map[string]interface{}
}

func (f *GeoJSONFeature) UnmarshalJSON(p []byte) error {
	var topLevel struct {
		Type       string                 `json:"type"`
		Geometry   AnyGeometry            `json:"geometry"`
		ID         interface{}            `json:"id,omitempty"`
		Properties map[string]interface{} `json:"properties,omitempty"`
	}
	if err := json.Unmarshal(p, &topLevel); err != nil {
		return err
	}

	// TODO: check for type not set correctly

	f.ID = topLevel.ID
	f.Properties = topLevel.Properties
	f.Geometry = topLevel.Geometry.Geom // TODO: check for case where geometry is nil

	return nil
}

// GeoJSONFeatureCollection is a collection of GeoJSONFeatures.
type GeoJSONFeatureCollection []GeoJSONFeature

func (c *GeoJSONFeatureCollection) UnmarshalJSON(p []byte) error {
	var topLevel struct {
		Type     string           `json:"type"`
		Features []GeoJSONFeature `json:"features"`
	}
	if err := json.Unmarshal(p, &topLevel); err != nil {
		return err
	}
	if topLevel.Type == "" {
		return errors.New("type field missing or empty")
	}
	if topLevel.Type != "FeatureCollection" {
		return fmt.Errorf("type field isn't set to FeatureCollection: '%s'", topLevel.Type)
	}
	(*c) = make([]GeoJSONFeature, len(topLevel.Features))
	copy(*c, topLevel.Features)
	return nil
}
