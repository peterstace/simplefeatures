package jts

// NodingSnapround_GeometryNoder nodes the linework in a list of Geometries
// using Snap-Rounding to a given PrecisionModel.
//
// Input coordinates do not need to be rounded to the precision model. All
// output coordinates are rounded to the precision model.
//
// This class does not dissolve the output linework, so there may be duplicate
// linestrings in the output. Subsequent processing (e.g. polygonization) may
// require the linework to be unique. Using UnaryUnion is one way to do this
// (although this is an inefficient approach).
type NodingSnapround_GeometryNoder struct {
	geomFact          *Geom_GeometryFactory
	pm                *Geom_PrecisionModel
	isValidityChecked bool
}

// NodingSnapround_NewGeometryNoder creates a new noder which snap-rounds to a
// grid specified by the given PrecisionModel.
func NodingSnapround_NewGeometryNoder(pm *Geom_PrecisionModel) *NodingSnapround_GeometryNoder {
	return &NodingSnapround_GeometryNoder{
		pm: pm,
	}
}

// SetValidate sets whether noding validity is checked after noding is
// performed.
func (gn *NodingSnapround_GeometryNoder) SetValidate(isValidityChecked bool) {
	gn.isValidityChecked = isValidityChecked
}

// Node nodes the linework of a set of Geometries using SnapRounding.
func (gn *NodingSnapround_GeometryNoder) Node(geoms []*Geom_Geometry) []*Geom_LineString {
	if len(geoms) == 0 {
		return nil
	}

	// Get geometry factory.
	gn.geomFact = geoms[0].GetFactory()

	segStrings := gn.toSegmentStrings(gn.extractLines(geoms))
	sr := NodingSnapround_NewSnapRoundingNoder(gn.pm)
	sr.ComputeNodes(segStrings)
	nodedLines := sr.GetNodedSubstrings()

	if gn.isValidityChecked {
		nv := Noding_NewNodingValidator(nodedLines)
		nv.CheckValid()
	}

	return gn.toLineStrings(nodedLines)
}

func (gn *NodingSnapround_GeometryNoder) toLineStrings(segStrings []Noding_SegmentString) []*Geom_LineString {
	lines := make([]*Geom_LineString, 0)
	for _, ss := range segStrings {
		// Skip collapsed lines.
		if ss.Size() < 2 {
			continue
		}
		lines = append(lines, gn.geomFact.CreateLineStringFromCoordinates(ss.GetCoordinates()))
	}
	return lines
}

func (gn *NodingSnapround_GeometryNoder) extractLines(geoms []*Geom_Geometry) []*Geom_LineString {
	var lines []*Geom_LineString
	for _, geom := range geoms {
		lines = GeomUtil_LinearComponentExtracter_GetLinesFromGeometryToSlice(geom, lines)
	}
	return lines
}

func (gn *NodingSnapround_GeometryNoder) toSegmentStrings(lines []*Geom_LineString) []Noding_SegmentString {
	segStrings := make([]Noding_SegmentString, len(lines))
	for i, line := range lines {
		nss := Noding_NewNodedSegmentString(line.GetCoordinates(), nil)
		segStrings[i] = nss
	}
	return segStrings
}
