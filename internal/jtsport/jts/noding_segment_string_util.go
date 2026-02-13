package jts

import (
	"fmt"
	"strings"
)

// Noding_SegmentStringUtil provides utility methods for processing SegmentStrings.
type Noding_SegmentStringUtil struct{}

// Noding_SegmentStringUtil_ExtractSegmentStrings extracts all linear components
// from a given Geometry to SegmentStrings. The SegmentString data item is set
// to be the source Geometry.
func Noding_SegmentStringUtil_ExtractSegmentStrings(geom *Geom_Geometry) []Noding_SegmentString {
	return Noding_SegmentStringUtil_ExtractNodedSegmentStrings(geom)
}

// Noding_SegmentStringUtil_ExtractNodedSegmentStrings extracts all linear
// components from a given Geometry to NodedSegmentStrings. The SegmentString
// data item is set to be the source Geometry.
func Noding_SegmentStringUtil_ExtractNodedSegmentStrings(geom *Geom_Geometry) []Noding_SegmentString {
	var segStr []Noding_SegmentString
	lines := GeomUtil_LinearComponentExtracter_GetLines(geom)
	for _, line := range lines {
		pts := line.GetCoordinates()
		segStr = append(segStr, Noding_NewNodedSegmentString(pts, geom))
	}
	return segStr
}

// Noding_SegmentStringUtil_ExtractBasicSegmentStrings extracts all linear
// components from a given Geometry to BasicSegmentStrings. The SegmentString
// data item is set to be the source Geometry.
func Noding_SegmentStringUtil_ExtractBasicSegmentStrings(geom *Geom_Geometry) []Noding_SegmentString {
	var segStr []Noding_SegmentString
	lines := GeomUtil_LinearComponentExtracter_GetLines(geom)
	for _, line := range lines {
		pts := line.GetCoordinates()
		segStr = append(segStr, Noding_NewBasicSegmentString(pts, geom))
	}
	return segStr
}

// Noding_SegmentStringUtil_ToGeometry converts a collection of SegmentStrings
// into a Geometry. The geometry will be either a LineString or a
// MultiLineString (possibly empty).
func Noding_SegmentStringUtil_ToGeometry(segStrings []Noding_SegmentString, geomFact *Geom_GeometryFactory) *Geom_Geometry {
	lines := make([]*Geom_LineString, len(segStrings))
	index := 0
	for _, ss := range segStrings {
		line := geomFact.CreateLineStringFromCoordinates(ss.GetCoordinates())
		lines[index] = line
		index++
	}
	if len(lines) == 1 {
		return lines[0].Geom_Geometry
	}
	return geomFact.CreateMultiLineStringFromLineStrings(lines).Geom_Geometry
}

// Noding_SegmentStringUtil_ToString returns a string representation of a list
// of SegmentStrings.
func Noding_SegmentStringUtil_ToString(segStrings []Noding_SegmentString) string {
	var buf strings.Builder
	for _, segStr := range segStrings {
		buf.WriteString(fmt.Sprintf("%v", segStr))
		buf.WriteString("\n")
	}
	return buf.String()
}
