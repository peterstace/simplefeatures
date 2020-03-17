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
	if err := detectCoordinatesLengths(rootObj, hasLength); err != nil {
		return Geometry{}, err
	}

	// We want to parse the geojson as a 3D geometry in the case where there is
	// at least 1 non-empty geometry, and there are no 2D coordinates (since
	// otherwise we would not be able to provide the height for the 2D
	// coordinates). In all other cases, we can only sensibly interpret the
	// geojson as being 2D.
	//
	// | hasEmpty | has2D | has3D | result |
	// | ---      | ---   | ---   | ---    |
	// | false    | false | false | DimXY  |
	// | false    | false | true  | XYZ    |
	// | false    | true  | false | DimXY  |
	// | false    | true  | true  | DimXY  |
	// | true     | false | false | DimXY  |
	// | true     | false | true  | XYZ    |
	// | true     | true  | false | DimXY  |
	// | true     | true  | true  | DimXY  |
	var has2D, has3D bool
	for length := range hasLength {
		if length == 2 {
			has2D = true
		}

		// The GeoJSON spec allows parsers to ignore any "extra" coordinate
		// values in addition to the normal 3 coordinate values used to specify
		// a 3D position. So having a length strictly greater than 3 is not an
		// error.
		if length >= 3 {
			has3D = true
		}
	}
	ctype := DimXY
	if !has2D && has3D {
		ctype = DimXYZ
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

func detectCoordinatesLengths(node interface{}, hasLength map[int]bool) error {
	switch node := node.(type) {
	case geojsonPoint:
		n := len(node.coords)
		hasLength[n] = true
		if n == 1 {
			return geojsonInvalidCoordinatesLengthError{n}
		}
		return nil
	case geojsonLineString:
		for _, c := range node.coords {
			n := len(c)
			hasLength[n] = true
			if n < 2 {
				return geojsonInvalidCoordinatesLengthError{n}
			}
		}
		return nil
	case geojsonPolygon:
		for _, outer := range node.coords {
			for _, inner := range outer {
				n := len(inner)
				hasLength[n] = true
				if n < 2 {
					return geojsonInvalidCoordinatesLengthError{n}
				}
			}
		}
		return nil
	case geojsonMultiPoint:
		for _, c := range node.coords {
			// GeoJSON MultiPoints do not allow empty Points inside them.
			n := len(c)
			hasLength[n] = true
			if n < 2 {
				return geojsonInvalidCoordinatesLengthError{n}
			}
		}
		return nil
	case geojsonMultiLineString:
		for _, outer := range node.coords {
			for _, inner := range outer {
				n := len(inner)
				hasLength[n] = true
				if n < 2 {
					return geojsonInvalidCoordinatesLengthError{n}
				}
			}
		}
		return nil
	case geojsonMultiPolygon:
		for _, outer := range node.coords {
			for _, middle := range outer {
				for _, inner := range middle {
					n := len(inner)
					hasLength[n] = true
					if n < 2 {
						return geojsonInvalidCoordinatesLengthError{n}
					}
				}
			}
		}
		return nil
	case geojsonGeometryCollection:
		for _, child := range node.geoms {
			if err := detectCoordinatesLengths(child, hasLength); err != nil {
				return err
			}
		}
		return nil
	default:
		panic(fmt.Sprintf("unexpected node: %#v", node))
	}
}

func geojsonNodeToGeometry(node interface{}, ctype CoordinatesType) (Geometry, error) {
	switch node := node.(type) {
	case geojsonPoint:
		coords, ok := oneDimFloat64sToCoordinates(node.coords, ctype)
		if ok {
			return NewPoint(coords).AsGeometry(), nil
		} else {
			return NewEmptyPoint(ctype).AsGeometry(), nil
		}
	case geojsonLineString:
		seq := twoDimFloat64sToSequence(node.coords, ctype)
		ls, err := NewLineString(seq)
		return ls.AsGeometry(), err
	case geojsonPolygon:
		if len(node.coords) == 0 {
			return Polygon{}.ForceCoordinatesType(ctype).AsGeometry(), nil
		}
		rings := make([]LineString, len(node.coords))
		for i, coords := range node.coords {
			seq := twoDimFloat64sToSequence(coords, ctype)
			var err error
			rings[i], err = NewLineString(seq)
			if err != nil {
				return Geometry{}, err
			}
		}
		poly, err := NewPolygonFromRings(rings)
		return poly.AsGeometry(), err
	case geojsonMultiPoint:
		// GeoJSON MultiPoints cannot contain empty Points.
		seq := twoDimFloat64sToSequence(node.coords, ctype)
		return NewMultiPoint(seq, BitSet{}).AsGeometry(), nil
	case geojsonMultiLineString:
		if len(node.coords) == 0 {
			return MultiLineString{}.ForceCoordinatesType(ctype).AsGeometry(), nil
		}
		lss := make([]LineString, len(node.coords))
		for i, coords := range node.coords {
			seq := twoDimFloat64sToSequence(coords, ctype)
			var err error
			lss[i], err = NewLineString(seq)
			if err != nil {
				return Geometry{}, err
			}
		}
		return NewMultiLineStringFromLineStrings(lss).AsGeometry(), nil
	case geojsonMultiPolygon:
		if len(node.coords) == 0 {
			return MultiPolygon{}.ForceCoordinatesType(ctype).AsGeometry(), nil
		}
		polys := make([]Polygon, len(node.coords))
		for i, coords := range node.coords {
			rings := make([]LineString, len(coords))
			for j, coords := range coords {
				seq := twoDimFloat64sToSequence(coords, ctype)
				var err error
				rings[j], err = NewLineString(seq)
				if err != nil {
					return Geometry{}, err
				}
			}
			var err error
			polys[i], err = NewPolygonFromRings(rings)
			if err != nil {
				return Geometry{}, err
			}
			polys[i] = polys[i].ForceCoordinatesType(ctype)
		}
		mp, err := NewMultiPolygonFromPolygons(polys)
		return mp.AsGeometry(), err
	case geojsonGeometryCollection:
		if len(node.geoms) == 0 {
			return GeometryCollection{}.ForceCoordinatesType(ctype).AsGeometry(), nil
		}
		children := make([]Geometry, len(node.geoms))
		for i, child := range node.geoms {
			var err error
			children[i], err = geojsonNodeToGeometry(child, ctype)
			if err != nil {
				return Geometry{}, err
			}
		}
		return NewGeometryCollection(children).AsGeometry(), nil
	default:
		panic(fmt.Sprintf("unexpected node: %#v", node))
	}
}

func oneDimFloat64sToCoordinates(fs []float64, ctype CoordinatesType) (Coordinates, bool) {
	if len(fs) == 0 {
		return Coordinates{}, false
	}
	coords := Coordinates{
		XY:   XY{fs[0], fs[1]},
		Type: ctype,
	}
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
	return NewSequence(floats, ctype)
}
