package geom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

func UnmarshalGeoJSON(input []byte, opts ...ConstructorOption) (Geometry, error) {
	var firstPass struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(bytes.NewReader(input)).Decode(&firstPass); err != nil {
		return Geometry{}, err
	}

	switch firstPass.Type {
	case "Point":
		var secondPass struct {
			Coords []float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&secondPass); err != nil {
			return Geometry{}, err
		}
		if len(secondPass.Coords) == 0 {
			return NewEmptyPoint(opts...).AsGeometry(), nil
		}
		coords, err := oneDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return Geometry{}, err
		}
		return NewPointC(coords, opts...).AsGeometry(), nil
	case "LineString", "MultiPoint":
		var secondPass struct {
			Coords [][]float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&secondPass); err != nil {
			return Geometry{}, err
		}
		coords, err := twoDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return Geometry{}, err
		}
		switch firstPass.Type {
		case "LineString":
			switch len(coords) {
			case 0:
				return NewEmptyLineString(opts...).AsGeometry(), nil
			case 2:
				ln, err := NewLineC(coords[0], coords[1], opts...)
				if err != nil {
					return Geometry{}, err
				}
				return ln.AsGeometry(), nil
			default:
				ls, err := NewLineStringC(coords, opts...)
				if err != nil {
					return Geometry{}, err
				}
				return ls.AsGeometry(), nil
			}
		case "MultiPoint":
			return NewMultiPointC(coords, opts...).AsGeometry(), nil
		default:
			panic("switch case bug")
		}
	case "Polygon", "MultiLineString":
		var secondPass struct {
			Coords [][][]float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&secondPass); err != nil {
			return Geometry{}, err
		}
		coords, err := threeDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return Geometry{}, err
		}
		switch firstPass.Type {
		case "Polygon":
			switch len(coords) {
			case 0:
				return NewEmptyPolygon(opts...).AsGeometry(), nil
			default:
				poly, err := NewPolygonC(coords, opts...)
				if err != nil {
					return Geometry{}, err
				}
				return poly.AsGeometry(), nil
			}
		case "MultiLineString":
			mls, err := NewMultiLineStringC(coords, opts...)
			if err != nil {
				return Geometry{}, err
			}
			return mls.AsGeometry(), nil
		default:
			panic("switch case bug")
		}
	case "MultiPolygon":
		var secondPass struct {
			Coords [][][][]float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&secondPass); err != nil {
			return Geometry{}, err
		}
		coords, err := fourDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return Geometry{}, err
		}
		mp, err := NewMultiPolygonC(coords, opts...)
		if err != nil {
			return Geometry{}, err
		}
		return mp.AsGeometry(), nil
	case "GeometryCollection":
		var secondPass struct {
			Geometries []Geometry `json:"geometries"`
		}
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&secondPass); err != nil {
			return Geometry{}, err
		}
		geoms := make([]Geometry, len(secondPass.Geometries))
		for i := range geoms {
			geoms[i] = secondPass.Geometries[i]
		}
		return NewGeometryCollection(geoms, opts...).AsGeometry(), nil
	case "":
		return Geometry{}, errors.New("type field missing or empty")
	default:
		return Geometry{}, fmt.Errorf("unknown geojson type: %s", firstPass.Type)
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
	return Coordinates{XY{fs[0], fs[1]}}, nil
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

func marshalGeoJSON(geomType string, coordinates interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`{"type":"`)
	buf.WriteString(geomType)
	buf.WriteString(`","coordinates":`)
	coordJSON, err := json.Marshal(coordinates)
	if err != nil {
		return nil, err
	}
	buf.Write(coordJSON)
	buf.WriteRune('}')
	return buf.Bytes(), nil
}
