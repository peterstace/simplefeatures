package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_LinearComponentExtracter extracts all the 1-dimensional (LineString)
// components from a Geometry. For polygonal geometries, this will extract all
// the component LinearRings. If desired, LinearRings can be forced to be
// returned as LineStrings.
type GeomUtil_LinearComponentExtracter struct {
	lines                []*Geom_LineString
	isForcedToLineString bool
}

var _ Geom_GeometryComponentFilter = (*GeomUtil_LinearComponentExtracter)(nil)

func (lce *GeomUtil_LinearComponentExtracter) IsGeom_GeometryComponentFilter() {}

// GeomUtil_NewLinearComponentExtracter constructs a LinearComponentExtracter
// with a list in which to store LineStrings found.
func GeomUtil_NewLinearComponentExtracter(lines []*Geom_LineString) *GeomUtil_LinearComponentExtracter {
	return &GeomUtil_LinearComponentExtracter{
		lines:                lines,
		isForcedToLineString: false,
	}
}

// GeomUtil_NewLinearComponentExtracterWithForce constructs a
// LinearComponentExtracter with a list in which to store LineStrings found and
// a flag to force LinearRings to LineStrings.
func GeomUtil_NewLinearComponentExtracterWithForce(lines []*Geom_LineString, isForcedToLineString bool) *GeomUtil_LinearComponentExtracter {
	return &GeomUtil_LinearComponentExtracter{
		lines:                lines,
		isForcedToLineString: isForcedToLineString,
	}
}

// GeomUtil_LinearComponentExtracter_GetLinesFromCollection extracts the linear
// components from a collection of Geometries and adds them to the provided
// slice.
func GeomUtil_LinearComponentExtracter_GetLinesFromCollection(geoms []*Geom_Geometry, lines []*Geom_LineString) []*Geom_LineString {
	for _, g := range geoms {
		lines = GeomUtil_LinearComponentExtracter_GetLinesFromGeometryToSlice(g, lines)
	}
	return lines
}

// GeomUtil_LinearComponentExtracter_GetLinesFromCollectionWithForce extracts
// the linear components from a collection of Geometries and adds them to the
// provided slice, with optional forcing of LinearRings to LineStrings.
func GeomUtil_LinearComponentExtracter_GetLinesFromCollectionWithForce(geoms []*Geom_Geometry, lines []*Geom_LineString, forceToLineString bool) []*Geom_LineString {
	for _, g := range geoms {
		lines = GeomUtil_LinearComponentExtracter_GetLinesFromGeometryToSliceWithForce(g, lines, forceToLineString)
	}
	return lines
}

// GeomUtil_LinearComponentExtracter_GetLinesFromGeometryToSlice extracts the
// linear components from a single Geometry and adds them to the provided slice.
func GeomUtil_LinearComponentExtracter_GetLinesFromGeometryToSlice(geom *Geom_Geometry, lines []*Geom_LineString) []*Geom_LineString {
	if java.InstanceOf[*Geom_LineString](geom) {
		return append(lines, java.Cast[*Geom_LineString](geom))
	}
	extracter := GeomUtil_NewLinearComponentExtracter(lines)
	geom.Apply(extracter)
	return extracter.lines
}

// GeomUtil_LinearComponentExtracter_GetLinesFromGeometryToSliceWithForce
// extracts the linear components from a single Geometry and adds them to the
// provided slice, with optional forcing of LinearRings to LineStrings.
func GeomUtil_LinearComponentExtracter_GetLinesFromGeometryToSliceWithForce(geom *Geom_Geometry, lines []*Geom_LineString, forceToLineString bool) []*Geom_LineString {
	extracter := GeomUtil_NewLinearComponentExtracterWithForce(lines, forceToLineString)
	geom.Apply(extracter)
	return extracter.lines
}

// GeomUtil_LinearComponentExtracter_GetLines extracts the linear components
// from a single geometry. If more than one geometry is to be processed, it is
// more efficient to create a single LinearComponentExtracter instance and pass
// it to multiple geometries.
func GeomUtil_LinearComponentExtracter_GetLines(geom *Geom_Geometry) []*Geom_LineString {
	return GeomUtil_LinearComponentExtracter_GetLinesWithForce(geom, false)
}

// GeomUtil_LinearComponentExtracter_GetLinesWithForce extracts the linear
// components from a single geometry with optional forcing of LinearRings to
// LineStrings.
func GeomUtil_LinearComponentExtracter_GetLinesWithForce(geom *Geom_Geometry, forceToLineString bool) []*Geom_LineString {
	var lines []*Geom_LineString
	extracter := GeomUtil_NewLinearComponentExtracterWithForce(lines, forceToLineString)
	geom.Apply(extracter)
	return extracter.lines
}

// GeomUtil_LinearComponentExtracter_GetGeometry extracts the linear components
// from a single Geometry and returns them as either a LineString or
// MultiLineString.
func GeomUtil_LinearComponentExtracter_GetGeometry(geom *Geom_Geometry) *Geom_Geometry {
	lines := GeomUtil_LinearComponentExtracter_GetLines(geom)
	geoms := make([]*Geom_Geometry, len(lines))
	for i, line := range lines {
		geoms[i] = line.Geom_Geometry
	}
	return geom.GetFactory().BuildGeometry(geoms)
}

// GeomUtil_LinearComponentExtracter_GetGeometryWithForce extracts the linear
// components from a single Geometry and returns them as either a LineString or
// MultiLineString, with optional forcing of LinearRings to LineStrings.
func GeomUtil_LinearComponentExtracter_GetGeometryWithForce(geom *Geom_Geometry, forceToLineString bool) *Geom_Geometry {
	lines := GeomUtil_LinearComponentExtracter_GetLinesWithForce(geom, forceToLineString)
	geoms := make([]*Geom_Geometry, len(lines))
	for i, line := range lines {
		geoms[i] = line.Geom_Geometry
	}
	return geom.GetFactory().BuildGeometry(geoms)
}

// SetForceToLineString indicates that LinearRing components should be converted
// to pure LineStrings.
func (lce *GeomUtil_LinearComponentExtracter) SetForceToLineString(isForcedToLineString bool) {
	lce.isForcedToLineString = isForcedToLineString
}

// Filter implements the GeometryComponentFilter interface.
func (lce *GeomUtil_LinearComponentExtracter) Filter(geom *Geom_Geometry) {
	if lce.isForcedToLineString && java.InstanceOf[*Geom_LinearRing](geom) {
		ring := java.Cast[*Geom_LinearRing](geom)
		line := geom.GetFactory().CreateLineStringFromCoordinateSequence(ring.GetCoordinateSequence())
		lce.lines = append(lce.lines, line)
		return
	}
	// Check if this is a LineString (or subtype like LinearRing).
	// The InstanceOf check traverses the child chain and will match both
	// LineString and LinearRing since LinearRing embeds LineString.
	if java.InstanceOf[*Geom_LineString](geom) {
		// Walk the chain to find the LineString level.
		self := java.GetLeaf(geom)
		if ls, ok := self.(*Geom_LineString); ok {
			lce.lines = append(lce.lines, ls)
		} else if ring, ok := self.(*Geom_LinearRing); ok {
			lce.lines = append(lce.lines, ring.Geom_LineString)
		}
	}
}
