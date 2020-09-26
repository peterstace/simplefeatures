package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
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

func containsMultiPointWithEmptyPoint(g geom.Geometry) bool {
	switch {
	case g.IsMultiPoint():
		mp := g.AsMultiPoint()
		for i := 0; i < mp.NumPoints(); i++ {
			if mp.PointN(i).IsEmpty() {
				return true
			}
		}
	case g.IsGeometryCollection():
		gc := g.AsGeometryCollection()
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
		mls := g.AsMultiLineString()
		for i := 0; i < mls.NumLineStrings(); i++ {
			if mls.LineStringN(i).IsEmpty() {
				return true
			}
		}
	case g.IsGeometryCollection():
		gc := g.AsGeometryCollection()
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

func isNonEmptyGeometryCollection(g geom.Geometry) bool {
	return g.IsGeometryCollection() && !g.IsEmpty()
}

// checkEqualIgnoringLinerStructure compares two geometries, taking into
// consideration differences between GEOS and simplefeatures with regards to
// linear element structure.
func checkEqualIgnoringLinerStructure(g1, g2 geom.Geometry) error {
	g1Pts, g1Lines, g1Polys := separateParts(g1)
	g2Pts, g2Lines, g2Polys := separateParts(g2)

	if !g1Pts.EqualsExact(g2Pts.AsGeometry()) {
		return errors.New("points not equal")
	}
	if !g1Lines.EqualsExact(g2Lines.AsGeometry()) {
		return errors.New("lines not equal")
	}
	if !g1Polys.EqualsExact(g2Polys.AsGeometry()) {
		return errors.New("polys not equal")
	}
	return nil
}

func separateParts(g geom.Geometry) (geom.MultiPoint, geom.MultiLineString, geom.MultiPolygon) {
	var (
		pts   []geom.Point
		lines []geom.LineString
		polys []geom.Polygon
	)

	switch g.Type() {
	case geom.TypePoint:
		pts = append(pts, g.AsPoint())
	case geom.TypeMultiPoint:
		mp := g.AsMultiPoint()
		for i := 0; i < mp.NumPoints(); i++ {
			pts = append(pts, mp.PointN(i))
		}
	case geom.TypeLineString:
		ls := g.AsLineString()
		seq := ls.Coordinates()
		n := seq.Length()
		for i := 0; i+1 < n; i++ {
			ptA := seq.GetXY(i)
			ptB := seq.GetXY(i + 1)
			if ptA == ptB {
				continue
			}
			line, err := geom.NewLineString(geom.NewSequence([]float64{ptA.X, ptA.Y, ptB.X, ptB.Y}, geom.DimXY))
			if err != nil {
				panic("invalid 2 point line: " + err.Error())
			}
			lines = append(lines, line)
		}
	case geom.TypeMultiLineString:
		mls := g.AsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			_, child, _ := separateParts(mls.LineStringN(i).AsGeometry())
			for j := 0; j < child.NumLineStrings(); j++ {
				lines = append(lines, child.LineStringN(j))
			}
		}
	case geom.TypePolygon:
		polys = append(polys, g.AsPolygon())
	case geom.TypeMultiPolygon:
		mp := g.AsMultiPolygon()
		for i := 0; i < mp.NumPolygons(); i++ {
			polys = append(polys, mp.PolygonN(i))
		}
	case geom.TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			newPts, newLines, newPolys := separateParts(gc.GeometryN(i))
			for j := 0; j < newPts.NumPoints(); j++ {
				pts = append(pts, newPts.PointN(j))
			}
			for j := 0; j < newLines.NumLineStrings(); j++ {
				lines = append(lines, newLines.LineStringN(j))
			}
			for j := 0; j < newPolys.NumPolygons(); j++ {
				polys = append(polys, newPolys.PolygonN(j))
			}
		}
	default:
		panic("unknown geometry type")
	}

	mp, err := geom.NewMultiPolygonFromPolygons(polys)
	if err != nil {
		panic(fmt.Sprintf("invalid multipolygon: %v", err))
	}
	return geom.NewMultiPointFromPoints(pts), geom.NewMultiLineStringFromLineStrings(lines), mp
}

// usesNonSimpleFloats return true iff each control point in the geometry uses
// at least one non-simple floats. A non-simple float is defined as a float
// that uses more than 4 decimal places past the decimal separator in its
// decimal expansion.
func usesNonSimpleFloats(g geom.Geometry) bool {
	nonSimpleF := func(f float64) bool {
		decExpansion := strconv.FormatFloat(f, 'f', -1, 64)
		sepIdx := strings.Index(decExpansion, ".")
		if sepIdx == -1 {
			return false // it's an integer
		}
		decPlaces := len(decExpansion) - sepIdx - 1
		return decPlaces > 4
	}
	nonSimpleXY := func(c geom.XY) bool {
		return nonSimpleF(c.X) || nonSimpleF(c.Y)
	}
	switch g.Type() {
	case geom.TypePoint:
		xy, ok := g.AsPoint().XY()
		return ok && nonSimpleXY(xy)
	case geom.TypeMultiPoint:
		mp := g.AsMultiPoint()
		for i := 0; i < mp.NumPoints(); i++ {
			if usesNonSimpleFloats(mp.PointN(i).AsGeometry()) {
				return true
			}
		}
		return false
	case geom.TypeLineString:
		seq := g.AsLineString().Coordinates()
		for i := 0; i < seq.Length(); i++ {
			xy := seq.GetXY(i)
			if nonSimpleXY(xy) {
				return true
			}
		}
		return false
	case geom.TypeMultiLineString:
		mls := g.AsMultiLineString()
		for i := 0; i < mls.NumLineStrings(); i++ {
			ls := mls.LineStringN(i).AsGeometry()
			if usesNonSimpleFloats(ls) {
				return true
			}
		}
		return false
	case geom.TypePolygon:
		return usesNonSimpleFloats(g.AsPolygon().Boundary().AsGeometry())
	case geom.TypeMultiPolygon:
		return usesNonSimpleFloats(g.AsMultiPolygon().Boundary().AsGeometry())
	case geom.TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			if usesNonSimpleFloats(gc.GeometryN(i)) {
				return true
			}
		}
		return false
	default:
		panic("unknown geometry type")
	}
}

func mantissaTerminatesQuickly(g geom.Geometry) bool {
	termF := func(f float64) bool {
		const (
			mantissaMask        = ^uint64(0) >> 12
			allowedMantissaMask = (mantissaMask >> 28) << 28
		)
		mant := math.Float64bits(f) & mantissaMask
		return mant & ^allowedMantissaMask == 0
	}
	termXY := func(xy geom.XY) bool {
		return termF(xy.X) && termF(xy.Y)
	}

	switch g.Type() {
	case geom.TypePoint:
		xy, ok := g.AsPoint().XY()
		return !ok || termXY(xy)
	case geom.TypeLineString:
		seq := g.AsLineString().Coordinates()
		for i := 0; i < seq.Length(); i++ {
			if !termXY(seq.GetXY(i)) {
				return false
			}
		}
		return true
	case geom.TypePolygon:
		return g.IsEmpty() || mantissaTerminatesQuickly(g.Boundary())
	case geom.TypeMultiPoint:
		mp := g.AsMultiPoint()
		for i := 0; i < mp.NumPoints(); i++ {
			pt := mp.PointN(i)
			if !mantissaTerminatesQuickly(pt.AsGeometry()) {
				return false
			}
		}
		return true
	case geom.TypeMultiLineString:
		mls := g.AsMultiLineString()
		for i := 0; i < mls.NumLineStrings(); i++ {
			ls := mls.LineStringN(i)
			if !mantissaTerminatesQuickly(ls.AsGeometry()) {
				return false
			}
		}
		return true
	case geom.TypeMultiPolygon:
		return g.IsEmpty() || mantissaTerminatesQuickly(g.Boundary())
	case geom.TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		for i := 0; i < gc.NumGeometries(); i++ {
			g := gc.GeometryN(i)
			if !mantissaTerminatesQuickly(g) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("unknown type: %v", g.Type()))
	}
}
