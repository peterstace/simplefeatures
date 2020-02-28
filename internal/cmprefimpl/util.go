package main

import (
	"strings"
	"text/scanner"

	"github.com/peterstace/simplefeatures/geom"
)

func containsNonEmptyPointInMultiPoint(g geom.Geometry) bool {
	switch {
	case g.IsGeometryCollection():
		gc := g.AsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if containsNonEmptyPointInMultiPoint(gc.GeometryN(i)) {
				return true
			}
		}
	case g.IsMultiPoint():
		mp := g.AsMultiPoint()
		for i := 0; i < mp.NumPoints(); i++ {
			if !mp.PointN(i).IsEmpty() {
				return true
			}
		}
	}
	return false
}

func containsCollectionWithOnlyEmptyElements(g geom.Geometry) bool {
	switch {
	case g.IsGeometryCollection():
		gc := g.AsGeometryCollection()
		if gc.IsEmpty() && gc.NumGeometries() > 0 {
			return true
		}
		for i := 0; i < gc.NumGeometries(); i++ {
			if containsCollectionWithOnlyEmptyElements(gc.GeometryN(i)) {
				return true
			}
		}
		return false
	case g.IsMultiPoint():
		mp := g.AsMultiPoint()
		return mp.IsEmpty() && mp.NumPoints() > 0
	case g.IsMultiLineString():
		mls := g.AsMultiLineString()
		return mls.IsEmpty() && mls.NumLineStrings() > 0
	case g.IsMultiPolygon():
		mp := g.AsMultiPolygon()
		return mp.IsEmpty() && mp.NumPolygons() > 0
	default:
		return false
	}
}

func containsOnlyGeometryCollections(g geom.Geometry) bool {
	if !g.IsGeometryCollection() {
		return false
	}
	gc := g.AsGeometryCollection()
	for i := 0; i < gc.NumGeometries(); i++ {
		if !containsOnlyGeometryCollections(gc.GeometryN(i)) {
			return false
		}
	}
	return true
}

func containsMultiPolygonWithEmptyPolygon(g geom.Geometry) bool {
	switch {
	case g.IsMultiPolygon():
		mp := g.AsMultiPolygon()
		for i := 0; i < mp.NumPolygons(); i++ {
			if mp.PolygonN(i).IsEmpty() {
				return true
			}
		}
		return false
	case g.IsGeometryCollection():
		gc := g.AsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if containsMultiPolygonWithEmptyPolygon(gc.GeometryN(i)) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func hasEmptyRing(g geom.Geometry) bool {
	// NOTE: Valid geometries _don't_ have empty rings. This function gets
	// called with invalid geometries.
	switch {
	case g.IsPolygon():
		p := g.AsPolygon()
		if p.ExteriorRing().IsEmpty() {
			return true
		}
		for i := 0; i < p.NumInteriorRings(); i++ {
			if p.InteriorRingN(i).IsEmpty() {
				return true
			}
		}
	case g.IsMultiPolygon():
		mp := g.AsMultiPolygon()
		for i := 0; i < mp.NumPolygons(); i++ {
			if hasEmptyRing(mp.PolygonN(i).AsGeometry()) {
				return true
			}
		}
	case g.IsGeometryCollection():
		gc := g.AsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if hasEmptyRing(gc.GeometryN(i)) {
				return true
			}
		}
	}
	return false
}

func hasEmptyPoint(g geom.Geometry) bool {
	switch {
	case g.IsPoint():
		return g.IsEmpty()
	case g.IsMultiPoint():
		mp := g.AsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			if mp.PointN(i).IsEmpty() {
				return true
			}
		}
	case g.IsGeometryCollection():
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			if hasEmptyPoint(gc.GeometryN(i)) {
				return true
			}
		}
	}
	return false
}

func tokenizeWKT(wkt string) []string {
	var scn scanner.Scanner
	scn.Init(strings.NewReader(wkt))
	scn.Error = func(_ *scanner.Scanner, msg string) {
		panic(msg)
	}
	var tokens []string
	for tok := scn.Scan(); tok != scanner.EOF; tok = scn.Scan() {
		tokens = append(tokens, scn.TokenText())
	}
	return tokens
}
