package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
)

func loadStringsFromFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	strSet := map[string]struct{}{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		strSet[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var strs []string
	for s := range strSet {
		strs = append(strs, s)
	}
	sort.Strings(strs)
	return strs, nil
}

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
		buf, err := hex.DecodeString(c)
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
