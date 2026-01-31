package main

import (
	"math"
	"strings"
	"text/scanner"

	"github.com/peterstace/simplefeatures/geom"
)

func containsNonEmptyPointInMultiPoint(g geom.Geometry) bool {
	switch {
	case g.IsGeometryCollection():
		gc := g.MustAsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if containsNonEmptyPointInMultiPoint(gc.GeometryN(i)) {
				return true
			}
		}
	case g.IsMultiPoint():
		mp := g.MustAsMultiPoint()
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
		gc := g.MustAsGeometryCollection()
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
		mp := g.MustAsMultiPoint()
		return mp.IsEmpty() && mp.NumPoints() > 0
	case g.IsMultiLineString():
		mls := g.MustAsMultiLineString()
		return mls.IsEmpty() && mls.NumLineStrings() > 0
	case g.IsMultiPolygon():
		mp := g.MustAsMultiPolygon()
		return mp.IsEmpty() && mp.NumPolygons() > 0
	default:
		return false
	}
}

func containsOnlyGeometryCollections(g geom.Geometry) bool {
	if !g.IsGeometryCollection() {
		return false
	}
	gc := g.MustAsGeometryCollection()
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
		mp := g.MustAsMultiPolygon()
		for i := 0; i < mp.NumPolygons(); i++ {
			if mp.PolygonN(i).IsEmpty() {
				return true
			}
		}
		return false
	case g.IsGeometryCollection():
		gc := g.MustAsGeometryCollection()
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

func containsMultiPointWithEmptyPoint(g geom.Geometry) bool {
	switch {
	case g.IsMultiPoint():
		mp := g.MustAsMultiPoint()
		for i := 0; i < mp.NumPoints(); i++ {
			if mp.PointN(i).IsEmpty() {
				return true
			}
		}
	case g.IsGeometryCollection():
		gc := g.MustAsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if containsMultiPointWithEmptyPoint(gc.GeometryN(i)) {
				return true
			}
		}
	}
	return false
}

func containsMultiLineStringWithEmptyLineString(g geom.Geometry) bool {
	switch {
	case g.IsMultiLineString():
		mls := g.MustAsMultiLineString()
		for i := 0; i < mls.NumLineStrings(); i++ {
			if mls.LineStringN(i).IsEmpty() {
				return true
			}
		}
	case g.IsGeometryCollection():
		gc := g.MustAsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if containsMultiLineStringWithEmptyLineString(gc.GeometryN(i)) {
				return true
			}
		}
	}
	return false
}

func hasEmptyRing(g geom.Geometry) bool {
	// NOTE: Valid geometries _don't_ have empty rings. This function gets
	// called with invalid geometries.
	switch {
	case g.IsPolygon():
		p := g.MustAsPolygon()
		if p.ExteriorRing().IsEmpty() {
			return true
		}
		for i := 0; i < p.NumInteriorRings(); i++ {
			if p.InteriorRingN(i).IsEmpty() {
				return true
			}
		}
	case g.IsMultiPolygon():
		mp := g.MustAsMultiPolygon()
		for i := 0; i < mp.NumPolygons(); i++ {
			if hasEmptyRing(mp.PolygonN(i).AsGeometry()) {
				return true
			}
		}
	case g.IsGeometryCollection():
		gc := g.MustAsGeometryCollection()
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
		mp := g.MustAsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			if mp.PointN(i).IsEmpty() {
				return true
			}
		}
	case g.IsGeometryCollection():
		gc := g.MustAsGeometryCollection()
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

// hasLargeCoordinates returns true if the geometry has any coordinates with
// magnitude large enough to cause floating point precision issues when
// comparing the results of operations performed on this geometry. The
// operations themselves work fine, but comparing the results fails because
// geom.ExactEquals only supports absolute tolerance. A relative tolerance
// option for ExactEquals would allow these comparisons to succeed.
func hasLargeCoordinates(g geom.Geometry) bool {
	env := g.Envelope()
	lo, hi, ok := env.MinMaxXYs()
	if !ok {
		return false
	}
	const threshold = 1e6
	return math.Abs(lo.X) > threshold ||
		math.Abs(lo.Y) > threshold ||
		math.Abs(hi.X) > threshold ||
		math.Abs(hi.Y) > threshold
}
