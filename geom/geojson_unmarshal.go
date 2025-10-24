package geom

import (
	"encoding/json"
	"fmt"
)

// UnmarshalGeoJSON unmarshals a geometry that is encoded as a GeoJSON Geometry Object.
//
// NoValidate{} can be passed in to disable geometry constraint validation.
func UnmarshalGeoJSON(input []byte, nv ...NoValidate) (Geometry, error) {
	var root geojsonNode
	if err := json.Unmarshal(input, &root); err != nil {
		return Geometry{}, wrapWithGeoJSONSyntaxError(err)
	}

	rootObj, err := decodeGeoJSON(root)
	if err != nil {
		return Geometry{}, err
	}

	hasLength := make(map[int]struct{})
	if err := detectCoordinatesLengths(rootObj, hasLength); err != nil {
		return Geometry{}, err
	}
	ctype := chooseGeoJSONCoordinatesType(hasLength)

	g := geojsonNodeToGeometry(rootObj, ctype)
	if len(nv) == 0 {
		if err := g.Validate(); err != nil {
			return Geometry{}, err
		}
	}
	return g, nil
}

func chooseGeoJSONCoordinatesType(hasLength map[int]struct{}) CoordinatesType {
	// The GeoJSON spec allows (but doesn't require) parsers to ignore any
	// "extra" coordinate values in addition to the normal 3 coordinate values
	// used to specify a 3D position. We choose to interpret a 4th dimension as
	// M values, and ignore any further (5th and above) dimensions without
	// error.
	//
	// We choose the highest ctype such that we can unambiguously parse any of
	// the points in the geometry (i.e. we don't have to make up a Z or M
	// value).
	//
	// | has2  | has3  | has4OrMore | ctype   |
	// | ---   | ---   | ---        | ---     |
	// | false | false | false      | DimXY   |
	// | false | false | true       | DimXYZM |
	// | false | true  | false      | DimXYZ  |
	// | false | true  | true       | DimXYZ  |
	// | true  | false | false      | DimXY   |
	// | true  | false | true       | DimXY   |
	// | true  | true  | false      | DimXY   |
	// | true  | true  | true       | DimXY   |
	var has2, has3, has4OrMore bool
	for length := range hasLength {
		has2 = has2 || length == 2
		has3 = has3 || length == 3
		has4OrMore = has4OrMore || length >= 4
	}
	if !has2 && has3 {
		return DimXYZ
	}
	if !has2 && !has3 && has4OrMore {
		return DimXYZM
	}
	return DimXY
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
	geoms []any
}

func decodeGeoJSON(node geojsonNode) (any, error) {
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
			geoms: make([]any, len(node.Geoms)),
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
		return nil, geojsonSyntaxError{fmt.Sprintf("unknown geometry type: '%s'", node.Type)}
	}
}

func extract1DimFloat64s(coords json.RawMessage) ([]float64, error) {
	var result []float64
	err := json.Unmarshal(coords, &result)
	return result, wrapWithGeoJSONSyntaxError(err)
}

func extract2DimFloat64s(coords json.RawMessage) ([][]float64, error) {
	var result [][]float64
	err := json.Unmarshal(coords, &result)
	return result, wrapWithGeoJSONSyntaxError(err)
}

func extract3DimFloat64s(coords json.RawMessage) ([][][]float64, error) {
	var result [][][]float64
	err := json.Unmarshal(coords, &result)
	return result, wrapWithGeoJSONSyntaxError(err)
}

func extract4DimFloat64s(coords json.RawMessage) ([][][][]float64, error) {
	var result [][][][]float64
	err := json.Unmarshal(coords, &result)
	return result, wrapWithGeoJSONSyntaxError(err)
}

func geojsonInvalidCoordinatesLengthError(n int) error {
	return geojsonSyntaxError{fmt.Sprintf("invalid geojson coordinate length: %d", n)}
}

func detectCoordinatesLengths(node any, hasLength map[int]struct{}) error {
	switch node := node.(type) {
	case geojsonPoint:
		n := len(node.coords)
		hasLength[n] = struct{}{}
		if n == 1 {
			return geojsonInvalidCoordinatesLengthError(n)
		}
		return nil
	case geojsonLineString:
		for _, c := range node.coords {
			n := len(c)
			hasLength[n] = struct{}{}
			if n < 2 {
				return geojsonInvalidCoordinatesLengthError(n)
			}
		}
		return nil
	case geojsonPolygon:
		for _, outer := range node.coords {
			for _, inner := range outer {
				n := len(inner)
				hasLength[n] = struct{}{}
				if n < 2 {
					return geojsonInvalidCoordinatesLengthError(n)
				}
			}
		}
		return nil
	case geojsonMultiPoint:
		for _, c := range node.coords {
			// GeoJSON MultiPoints do not allow empty Points inside them.
			n := len(c)
			hasLength[n] = struct{}{}
			if n < 2 {
				return geojsonInvalidCoordinatesLengthError(n)
			}
		}
		return nil
	case geojsonMultiLineString:
		for _, outer := range node.coords {
			for _, inner := range outer {
				n := len(inner)
				hasLength[n] = struct{}{}
				if n < 2 {
					return geojsonInvalidCoordinatesLengthError(n)
				}
			}
		}
		return nil
	case geojsonMultiPolygon:
		for _, outer := range node.coords {
			for _, middle := range outer {
				for _, inner := range middle {
					n := len(inner)
					hasLength[n] = struct{}{}
					if n < 2 {
						return geojsonInvalidCoordinatesLengthError(n)
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

func geojsonNodeToGeometry(node any, ctype CoordinatesType) Geometry {
	switch node := node.(type) {
	case geojsonPoint:
		coords, ok := oneDimFloat64sToCoordinates(node.coords, ctype)
		if ok {
			return NewPoint(coords).AsGeometry()
		}
		return NewEmptyPoint(ctype).AsGeometry()
	case geojsonLineString:
		seq := twoDimFloat64sToSequence(node.coords, ctype)
		return NewLineString(seq).AsGeometry()
	case geojsonPolygon:
		if len(node.coords) == 0 {
			return Polygon{}.ForceCoordinatesType(ctype).AsGeometry()
		}
		rings := make([]LineString, len(node.coords))
		for i, coords := range node.coords {
			seq := twoDimFloat64sToSequence(coords, ctype)
			rings[i] = NewLineString(seq)
		}
		return NewPolygon(rings).AsGeometry()
	case geojsonMultiPoint:
		// GeoJSON MultiPoints cannot contain empty Points.
		if len(node.coords) == 0 {
			return MultiPoint{}.ForceCoordinatesType(ctype).AsGeometry()
		}
		points := make([]Point, len(node.coords))
		for i, coords := range node.coords {
			coords, ok := oneDimFloat64sToCoordinates(coords, ctype)
			if ok {
				points[i] = NewPoint(coords)
			} else {
				points[i] = NewEmptyPoint(ctype)
			}
		}
		return NewMultiPoint(points).AsGeometry()
	case geojsonMultiLineString:
		if len(node.coords) == 0 {
			return MultiLineString{}.ForceCoordinatesType(ctype).AsGeometry()
		}
		lss := make([]LineString, len(node.coords))
		for i, coords := range node.coords {
			seq := twoDimFloat64sToSequence(coords, ctype)
			lss[i] = NewLineString(seq)
		}
		return NewMultiLineString(lss).AsGeometry()
	case geojsonMultiPolygon:
		if len(node.coords) == 0 {
			return MultiPolygon{}.ForceCoordinatesType(ctype).AsGeometry()
		}
		polys := make([]Polygon, len(node.coords))
		for i, coords := range node.coords {
			rings := make([]LineString, len(coords))
			for j, coords := range coords {
				seq := twoDimFloat64sToSequence(coords, ctype)
				rings[j] = NewLineString(seq)
			}
			polys[i] = NewPolygon(rings).ForceCoordinatesType(ctype)
		}
		return NewMultiPolygon(polys).AsGeometry()
	case geojsonGeometryCollection:
		if len(node.geoms) == 0 {
			return GeometryCollection{}.ForceCoordinatesType(ctype).AsGeometry()
		}
		children := make([]Geometry, len(node.geoms))
		for i, child := range node.geoms {
			children[i] = geojsonNodeToGeometry(child, ctype)
		}
		return NewGeometryCollection(children).AsGeometry()
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

	if len(fs) < ctype.Dimension() {
		// This should not happen because we used the lengths to choose the ctype.
		panic(fmt.Sprintf("unexpected length %d for ctype %s", len(fs), ctype))
	}

	switch ctype {
	case DimXY:
		// Do nothing, already set XY.
	case DimXYZ:
		coords.Z = fs[2]
	case DimXYZM:
		coords.Z = fs[2]
		coords.M = fs[3]
	case DimXYM:
		fallthrough // Cannot happen because DimXYM is never chosen as the ctype.
	default:
		panic(fmt.Sprintf("unexpected ctype %v in oneDimFloat64sToCoordinates", ctype))
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
