package geom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

// TODO: Shouldn't need to use a "first pass" -> "second pass" approach. Can
// just use json.RawMessage and then decode the inner part.

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
			return NewEmptyPoint(XYOnly).AsGeometry(), nil
		}
		coords, ctype, ok, err := oneDimFloat64sToCoordinates(secondPass.Coords)
		if err != nil {
			return Geometry{}, err
		}
		if !ok {
			return NewEmptyPoint(ctype).AsGeometry(), nil
		}
		return NewPointC(coords, ctype, opts...).AsGeometry(), nil
	case "LineString", "MultiPoint":
		var secondPass struct {
			Coords [][]float64 `json:"coordinates"`
		}
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&secondPass); err != nil {
			return Geometry{}, err
		}
		switch firstPass.Type {
		case "LineString":
			seq, err := twoDimFloat64sToSequence(secondPass.Coords)
			if err != nil {
				return Geometry{}, err
			}
			if seq.Length() == 2 {
				ln, err := NewLineC(
					seq.Get(0), seq.Get(1),
					seq.CoordinatesType(), opts...,
				)
				return ln.AsGeometry(), err
			} else {
				ls, err := NewLineStringFromSequence(seq, opts...)
				return ls.AsGeometry(), err
			}
		case "MultiPoint":
			seq, empty, err := twoDimFloat64sToOptionalSequence(secondPass.Coords)
			if err != nil {
				return Geometry{}, err
			}
			return NewMultiPointFromSequence(seq, empty, opts...).AsGeometry(), nil
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
		lss := make([]LineString, len(secondPass.Coords))
		for i, coords := range secondPass.Coords {
			seq, err := twoDimFloat64sToSequence(coords)
			if err != nil {
				return Geometry{}, err
			}
			lss[i], err = NewLineStringFromSequence(seq, opts...)
			if err != nil {
				return Geometry{}, err
			}
		}
		switch firstPass.Type {
		case "Polygon":
			poly, err := NewPolygon(lss, opts...)
			return poly.AsGeometry(), err
		case "MultiLineString":
			mls, err := NewMultiLineString(lss, opts...)
			return mls.AsGeometry(), err
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
		polys := make([]Polygon, len(secondPass.Coords))
		for i, polyCoords := range secondPass.Coords {
			rings := make([]LineString, len(polyCoords))
			for j, ringCoords := range polyCoords {
				seq, err := twoDimFloat64sToSequence(ringCoords)
				if err != nil {
					return Geometry{}, err
				}
				rings[j], err = NewLineStringFromSequence(seq, opts...)
				if err != nil {
					return Geometry{}, err
				}
			}
			var err error
			polys[i], err = NewPolygon(rings, opts...)
			if err != nil {
				return Geometry{}, err
			}
		}
		mp, err := NewMultiPolygon(polys, opts...)
		return mp.AsGeometry(), err
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
		gc, err := NewGeometryCollection(geoms, opts...)
		return gc.AsGeometry(), err
	case "":
		return Geometry{}, errors.New("type field missing or empty")
	default:
		return Geometry{}, fmt.Errorf("unknown geojson type: %s", firstPass.Type)
	}
}

func oneDimFloat64sToCoordinates(fs []float64) (Coordinates, CoordinatesType, bool, error) {
	for _, f := range fs {
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return Coordinates{}, 0, false, GeoJSONNaNOrInfErr
		}
	}

	switch len(fs) {
	case 0:
		return Coordinates{}, XYOnly, false, nil
	case 2:
		return Coordinates{XY: XY{fs[0], fs[1]}}, XYOnly, true, nil
	case 3:
		return Coordinates{XY: XY{fs[0], fs[1]}, Z: fs[2]}, XYZ, true, nil
	default:
		return Coordinates{}, 0, false, GeoJSONInvalidCoordinatesLengthError{len(fs)}
	}
}

func twoDimFloat64sToSequence(outer [][]float64) (Sequence, error) {
	var count2D, count3D int
	for _, c := range outer {
		switch len(c) {
		case 2:
			count2D++
		case 3:
			count3D++
		default:
			return Sequence{}, GeoJSONInvalidCoordinatesLengthError{len(c)}
		}
	}
	ctype := XYOnly
	if count2D == 0 && count3D > 0 {
		ctype = XYZ
	}
	stride := ctype.Dimension()
	floats := make([]float64, stride*len(outer))
	for i, c := range outer {
		for _, f := range c {
			if math.IsNaN(f) || math.IsInf(f, 0) {
				return Sequence{}, GeoJSONNaNOrInfErr
			}
		}
		for j := 0; j < stride; j++ {
			floats[stride*i+j] = c[j]
		}
	}
	return NewSequenceNoCopy(floats, ctype), nil
}

func twoDimFloat64sToOptionalSequence(outer [][]float64) (Sequence, BitSet, error) {
	var count2D, count3D int
	for _, c := range outer {
		switch len(c) {
		case 0:
		case 2:
			count2D++
		case 3:
			count3D++
		default:
			return Sequence{}, BitSet{}, GeoJSONInvalidCoordinatesLengthError{len(c)}
		}
	}
	ctype := XYOnly
	if count2D == 0 && count3D > 0 {
		ctype = XYZ
	}
	var empty BitSet
	stride := ctype.Dimension()
	floats := make([]float64, stride*len(outer))
	for i, c := range outer {
		if len(c) == 0 {
			empty.Set(i)
			continue
		}
		for j := 0; j < stride; j++ {
			floats[stride*i+j] = c[j]
		}
	}
	return NewSequenceNoCopy(floats, ctype), empty, nil
}

func appendGeoJSONCoordinate(
	dst []byte,
	ctype CoordinatesType,
	coords Coordinates,
	empty bool,
) []byte {
	if empty {
		return append(dst, "[]"...)
	}
	dst = append(dst, '[')
	dst = appendFloat(dst, coords.X)
	dst = append(dst, ',')
	dst = appendFloat(dst, coords.Y)
	if (ctype & XYZ) != 0 {
		dst = append(dst, ',')
		dst = appendFloat(dst, coords.Z)
	}
	// GeoJSON explicitly prohibits including M values.
	return append(dst, ']')
}

func appendGeoJSONSequence(dst []byte, seq Sequence, empty BitSet) []byte {
	dst = append(dst, '[')
	n := seq.Length()
	for i := 0; i < n; i++ {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendGeoJSONCoordinate(
			dst, seq.CoordinatesType(), seq.Get(i), empty.Get(i),
		)
	}
	dst = append(dst, ']')
	return dst
}

func appendGeoJSONSequences(dst []byte, seqs []Sequence) []byte {
	dst = append(dst, '[')
	for i, seq := range seqs {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendGeoJSONSequence(dst, seq, BitSet{})
	}
	dst = append(dst, ']')
	return dst
}

func appendGeoJSONSequenceMatrix(dst []byte, matrix [][]Sequence) []byte {
	dst = append(dst, '[')
	for i, seqs := range matrix {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendGeoJSONSequences(dst, seqs)
	}
	dst = append(dst, ']')
	return dst
}
