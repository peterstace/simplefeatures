package main

import (
	"errors"
	"strconv"

	"github.com/peterstace/simplefeatures/geom"
)

func convertToGeometries(candidates []string) ([]geom.Geometry, error) {
	var geoms []geom.Geometry
	for _, c := range candidates {
		g, err := geom.UnmarshalWKT(c, geom.NoValidate{})
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if len(geoms) == 0 {
		return nil, errors.New("could not extract any WKT geoms")
	}

	oldCount := len(geoms)
	for _, c := range candidates {
		buf, err := hexStringToBytes(c)
		if err != nil {
			continue
		}
		g, err := geom.UnmarshalWKB(buf, geom.NoValidate{})
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if oldCount == len(geoms) {
		return nil, errors.New("could not extract any WKB geoms")
	}

	oldCount = len(geoms)
	for _, c := range candidates {
		g, err := geom.UnmarshalGeoJSON([]byte(c), geom.NoValidate{})
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if oldCount == len(geoms) {
		return nil, errors.New("could not extract any geojson geoms")
	}

	return geoms, nil
}

func hexStringToBytes(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errors.New("hex string must have even length")
	}
	var buf []byte
	for i := 0; i < len(s); i += 2 {
		x, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return nil, err
		}
		buf = append(buf, byte(x))
	}
	return buf, nil
}
