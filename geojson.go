package simplefeatures

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
)

func UnmarshalGeoJSON(r io.Reader) (Geometry, error) {
	var firstPass struct {
		Type string `json:"type"`
	}
	var rCopy bytes.Buffer
	if err := json.NewDecoder(io.TeeReader(r, &rCopy)).Decode(&firstPass); err != nil {
		return nil, err
	}

	switch firstPass.Type {
	case "Point":
		var secondPass struct {
			Coords []float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(&rCopy).Decode(&secondPass); err != nil {
			return nil, err
		}
		if len(secondPass.Coords) == 0 {
			return NewEmptyPoint(), nil
		}
		coords, err := oneDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return nil, err
		}
		return NewPointFromCoords(coords), nil
	case "LineString", "MultiPoint":
		var secondPass struct {
			Coords [][]float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(&rCopy).Decode(&secondPass); err != nil {
			return nil, err
		}
		coords, err := twoDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return nil, err
		}
		switch firstPass.Type {
		case "LineString":
			switch len(coords) {
			case 0:
				return NewEmptyLineString(), nil
			case 2:
				return NewLine(coords[0], coords[1])
			default:
				return NewLineString(coords)
			}
		case "MultiPoint":
			return NewMultiPointFromCoords(coords), nil
		default:
			panic("switch case bug")
		}
	case "Polygon", "MultiLineString":
		var secondPass struct {
			Coords [][][]float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(&rCopy).Decode(&secondPass); err != nil {
			return nil, err
		}
		coords, err := threeDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return nil, err
		}
		switch firstPass.Type {
		case "Polygon":
			switch len(coords) {
			case 0:
				return NewEmptyPolygon(), nil
			default:
				return NewPolygonFromCoords(coords)
			}
		case "MultiLineString":
			return NewMultiLineStringFromCoords(coords)
		default:
			panic("switch case bug")
		}
	case "MultiPolygon":
		var secondPass struct {
			Coords [][][][]float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(&rCopy).Decode(&secondPass); err != nil {
			return nil, err
		}
		coords, err := fourDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return nil, err
		}
		return NewMultiPolygonFromCoords(coords)
	case "GeometryCollection":
		var secondPass struct {
			Geometries []AnyGeometry `json:"geometries"`
		}
		if err := json.NewDecoder(&rCopy).Decode(&secondPass); err != nil {
			return nil, err
		}
		geoms := make([]Geometry, len(secondPass.Geometries))
		for i := range geoms {
			geoms[i] = secondPass.Geometries[i].Geom
		}
		return NewGeometryCollection(geoms), nil
	case "":
		return nil, errors.New("type field missing or empty")
	default:
		return nil, fmt.Errorf("unknown geojson type: %s", firstPass.Type)
	}
}

func oneDimFloat64sToCoordinates(fs []float64) (Coordinates, error) {
	if len(fs) < 2 || len(fs) > 4 {
		return Coordinates{}, fmt.Errorf("coordinates have incorrect dimension: %d", len(fs))
	}
	for _, f := range fs {
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return Coordinates{}, errors.New("coordinate is NaN or inf")
		}
	}
	xstr := strconv.FormatFloat(fs[0], 'f', -1, 64)
	ystr := strconv.FormatFloat(fs[1], 'f', -1, 64)
	x, err := NewScalar(xstr)
	if err != nil {
		return Coordinates{}, err
	}
	y, err := NewScalar(ystr)
	if err != nil {
		return Coordinates{}, err
	}
	return Coordinates{XY{x, y}}, nil
}

func twoDimFloat64sToCoordinates(outer [][]float64) ([]Coordinates, error) {
	var coords []Coordinates
	for _, inner := range outer {
		cs, err := oneDimFloat64sToCoordinates(inner)
		if err != nil {
			return nil, err
		}
		coords = append(coords, cs)
	}
	return coords, nil
}

func threeDimFloat64sToCoordinates(outer [][][]float64) ([][]Coordinates, error) {
	var coords [][]Coordinates
	for _, inner := range outer {
		cs, err := twoDimFloat64sToCoordinates(inner)
		if err != nil {
			return nil, err
		}
		coords = append(coords, cs)
	}
	return coords, nil
}

func fourDimFloat64sToCoordinates(outer [][][][]float64) ([][][]Coordinates, error) {
	var coords [][][]Coordinates
	for _, inner := range outer {
		cs, err := threeDimFloat64sToCoordinates(inner)
		if err != nil {
			return nil, err
		}
		coords = append(coords, cs)
	}
	return coords, nil
}
