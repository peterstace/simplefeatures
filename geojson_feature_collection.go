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
	// applicable, ID can be left as nil. If it's set, then its value should
	// marshal into a JSON string or number (this is not enforced).
	ID interface{}

	// Properties are free-form properties that are associated with the
	// feature. If there are no properties associated with the feature, then it
	// can either be left as nil.
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

	if topLevel.Type == "" {
		return errors.New("feature type field missing or empty")
	}
	if topLevel.Type != "Feature" {
		return fmt.Errorf("type field not set to Feature: '%s'", topLevel.Type)
	}

	f.ID = topLevel.ID
	f.Properties = topLevel.Properties

	if topLevel.Geometry.Geom == nil {
		return fmt.Errorf("geometry field missing or empty")
	}
	f.Geometry = topLevel.Geometry.Geom

	return nil
}

func (f GeoJSONFeature) MarshalJSON() ([]byte, error) {
	if f.Geometry == nil {
		return nil, errors.New("Geometry field not set")
	}
	props := f.Properties
	if props == nil {
		props = map[string]interface{}{}
	}
	return json.Marshal(struct {
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
		return errors.New("feature collection type field missing or empty")
	}
	if topLevel.Type != "FeatureCollection" {
		return fmt.Errorf("type field not set to FeatureCollection: '%s'", topLevel.Type)
	}
	(*c) = make([]GeoJSONFeature, len(topLevel.Features))
	copy(*c, topLevel.Features)
	return nil
}

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
