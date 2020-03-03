package geom

import (
	"encoding/json"
	"fmt"
)

func UnmarshalGeoJSON(input []byte, opts ...ConstructorOption) (Geometry, error) {
	var root geojsonNode
	if err := json.Unmarshal(input, &root); err != nil {
		return Geometry{}, err
	}

	rootObj, err := decodeGeoJSON(root)
	if err != nil {
		return Geometry{}, err
	}

	hasLength := make(map[int]bool)
	detectCoordinatesLengths(rootObj, hasLength)

	has2D := hasLength[2]
	has3D := hasLength[3]
	delete(hasLength, 0)
	delete(hasLength, 2)
	delete(hasLength, 3)

	// If there are any lengths other than 0, 2, or 3, then an error is given.
	for length := range hasLength {
		return Geometry{}, GeoJSONInvalidCoordinatesLengthError{length}
	}

	// We want to parse the geojson as a 3D geometry in the case where there is
	// at least 1 non-empty geometry, and there are no 2D coordinates (since
	// otherwise we would not be able to provide the height for the 2D
	// coordinates). In all other cases, we can only sensibly interpret the
	// geojson as being 2D.
	//
	// | hasEmpty | has2D | has3D | result |
	// | ---      | ---   | ---   | ---    |
	// | false    | false | false | XYOnly |
	// | false    | false | true  | XYZ    |
	// | false    | true  | false | XYOnly |
	// | false    | true  | true  | XYOnly |
	// | true     | false | false | XYOnly |
	// | true     | false | true  | XYZ    |
	// | true     | true  | false | XYOnly |
	// | true     | true  | true  | XYOnly |
	ctype := XYOnly
	if !has2D && has3D {
		ctype = XYZ
	}

	return geojsonNodeToGeometry(rootObj, ctype)
}

type geojsonNode struct {
	Type   string          `json:"type"`
	Coords json.RawMessage `json:"coordinates"`
	Geoms  []geojsonNode   `json:"geometries"`
}

type geojsonPoint struct {
	coords []float64
}

type geojsonLineString struct {
	coords [][]float64
}

type geojsonPolygon struct {
	coords [][][]float64
}

type geojsonMultiPoint struct {
	coords [][]float64
}

type geojsonMultiLineString struct {
	coords [][][]float64
}

type geojsonMultiPolygon struct {
	coords [][][][]float64
}

type geojsonGeometryCollection struct {
	geoms []interface{}
}

func decodeGeoJSON(node geojsonNode) (interface{}, error) {
	switch node.Type {
	case "Point":
		c, err := extract1DimFloat64s(node.Coords)
		return geojsonPoint{c}, err
	case "LineString":
		c, err := extract2DimFloat64s(node.Coords)
		return geojsonLineString{c}, err
	case "Polygon":
		c, err := extract3DimFloat64s(node.Coords)
		return geojsonPolygon{c}, err
	case "MultiPoint":
		c, err := extract2DimFloat64s(node.Coords)
		return geojsonMultiPoint{c}, err
	case "MultiLineString":
		c, err := extract3DimFloat64s(node.Coords)
		return geojsonMultiLineString{c}, err
	case "MultiPolygon":
		c, err := extract4DimFloat64s(node.Coords)
		return geojsonMultiPolygon{c}, err
	case "GeometryCollection":
		parent := geojsonGeometryCollection{
			geoms: make([]interface{}, len(node.Geoms)),
		}
		for i, g := range node.Geoms {
			child, err := decodeGeoJSON(g)
			if err != nil {
				return nil, err
			}
			parent.geoms[i] = child
		}
		return parent, nil
	default:
		return nil, fmt.Errorf("unknown geojson type: %s", node.Type)
	}
}

func extract1DimFloat64s(coords json.RawMessage) ([]float64, error) {
	var result []float64
	err := json.Unmarshal(coords, &result)
	return result, err

}

func extract2DimFloat64s(coords json.RawMessage) ([][]float64, error) {
	var result [][]float64
	err := json.Unmarshal(coords, &result)
	return result, err
}

func extract3DimFloat64s(coords json.RawMessage) ([][][]float64, error) {
	var result [][][]float64
	err := json.Unmarshal(coords, &result)
	return result, err
}

func extract4DimFloat64s(coords json.RawMessage) ([][][][]float64, error) {
	var result [][][][]float64
	err := json.Unmarshal(coords, &result)
	return result, err
}

func detectCoordinatesLengths(node interface{}, hasLength map[int]bool) {
	switch node := node.(type) {
	case geojsonPoint:
		hasLength[len(node.coords)] = true
	case geojsonLineString:
		for _, c := range node.coords {
			hasLength[len(c)] = true
		}
	case geojsonPolygon:
		for _, outer := range node.coords {
			for _, inner := range outer {
				hasLength[len(inner)] = true
			}
		}
	case geojsonMultiPoint:
		for _, c := range node.coords {
			hasLength[len(c)] = true
		}
	case geojsonMultiLineString:
		for _, outer := range node.coords {
			for _, inner := range outer {
				hasLength[len(inner)] = true
			}
		}
	case geojsonMultiPolygon:
		for _, outer := range node.coords {
			for _, middle := range outer {
				for _, inner := range middle {
					hasLength[len(inner)] = true
				}
			}
		}
	case geojsonGeometryCollection:
		for _, child := range node.geoms {
			detectCoordinatesLengths(child, hasLength)
		}
	default:
		panic(fmt.Sprintf("unexpected node: %#v", node))
	}
}

func geojsonNodeToGeometry(node interface{}, ctype CoordinatesType) (Geometry, error) {
	switch node := node.(type) {
	case geojsonPoint:
		coords, ok := oneDimFloat64sToCoordinates(node.coords, ctype)
		if ok {
			return NewPointC(coords, ctype).AsGeometry(), nil
		} else {
			return NewEmptyPoint(ctype).AsGeometry(), nil
		}
	case geojsonLineString:
		seq := twoDimFloat64sToSequence(node.coords, ctype)
		ls, err := NewLineStringFromSequence(seq)
		return ls.AsGeometry(), err
	case geojsonPolygon:
		rings := make([]LineString, len(node.coords))
		for i, coords := range node.coords {
			seq := twoDimFloat64sToSequence(coords, ctype)
			var err error
			rings[i], err = NewLineStringFromSequence(seq)
			if err != nil {
				return Geometry{}, err
			}
		}
		if len(rings) == 0 {
			return NewEmptyPolygon(ctype).AsGeometry(), nil
		}
		poly, err := NewPolygon(rings)
		return poly.AsGeometry(), err
	case geojsonMultiPoint:
		seq, empty := twoDimFloat64sToOptionalSequence(node.coords, ctype)
		return NewMultiPointFromSequence(seq, empty).AsGeometry(), nil
	case geojsonMultiLineString:
		lss := make([]LineString, len(node.coords))
		for i, coords := range node.coords {
			seq := twoDimFloat64sToSequence(coords, ctype)
			var err error
			lss[i], err = NewLineStringFromSequence(seq)
			if err != nil {
				return Geometry{}, err
			}
		}
		if len(lss) == 0 {
			return NewEmptyMultiLineString(ctype).AsGeometry(), nil
		}
		poly, err := NewMultiLineString(lss)
		return poly.AsGeometry(), err
	case geojsonMultiPolygon:
		polys := make([]Polygon, len(node.coords))
		for i, coords := range node.coords {
			rings := make([]LineString, len(coords))
			for j, coords := range coords {
				seq := twoDimFloat64sToSequence(coords, ctype)
				var err error
				rings[j], err = NewLineStringFromSequence(seq)
				if err != nil {
					return Geometry{}, err
				}
			}
			if len(rings) == 0 {
				polys[i] = NewEmptyPolygon(ctype)
			} else {
				var err error
				polys[i], err = NewPolygon(rings)
				if err != nil {
					return Geometry{}, err
				}
			}
		}
		if len(polys) == 0 {
			return NewEmptyMultiPolygon(ctype).AsGeometry(), nil
		}
		mp, err := NewMultiPolygon(polys)
		return mp.AsGeometry(), err
	case geojsonGeometryCollection:
		children := make([]Geometry, len(node.geoms))
		for i, child := range node.geoms {
			var err error
			children[i], err = geojsonNodeToGeometry(child, ctype)
			if err != nil {
				return Geometry{}, err
			}
		}
		if len(children) == 0 {
			return NewEmptyGeometryCollection(ctype).AsGeometry(), nil
		}
		gc, err := NewGeometryCollection(children)
		return gc.AsGeometry(), err
	default:
		panic(fmt.Sprintf("unexpected node: %#v", node))
	}
}

func oneDimFloat64sToCoordinates(fs []float64, ctype CoordinatesType) (Coordinates, bool) {
	if len(fs) == 0 {
		return Coordinates{}, false
	}
	var coords Coordinates
	coords.X = fs[0]
	coords.Y = fs[1]
	if ctype.Is3D() {
		coords.Z = fs[2]
	}
	return coords, true
}

func twoDimFloat64sToSequence(input [][]float64, ctype CoordinatesType) Sequence {
	stride := ctype.Dimension()
	floats := make([]float64, stride*len(input))
	for i, c := range input {
		for j := 0; j < stride; j++ {
			floats[stride*i+j] = c[j]
		}
	}
	return NewSequenceNoCopy(floats, ctype)
}

func twoDimFloat64sToOptionalSequence(outer [][]float64, ctype CoordinatesType) (Sequence, BitSet) {
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
	return NewSequenceNoCopy(floats, ctype), empty
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
	if ctype.Is3D() {
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
