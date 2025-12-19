package jts

// Noding_TestUtil provides test utilities for noding tests.

// Noding_TestUtil_ToLines converts a collection of NodedSegmentStrings to a
// Geometry.
func Noding_TestUtil_ToLines(nodedList []*Noding_NodedSegmentString, geomFact *Geom_GeometryFactory) *Geom_Geometry {
	lines := make([]*Geom_LineString, len(nodedList))
	for i, nss := range nodedList {
		pts := nss.GetCoordinates()
		lines[i] = geomFact.CreateLineStringFromCoordinates(pts)
	}
	if len(lines) == 1 {
		return lines[0].Geom_Geometry
	}
	return geomFact.CreateMultiLineStringFromLineStrings(lines).Geom_Geometry
}

// Noding_TestUtil_ToSegmentStrings converts a list of LineStrings to
// NodedSegmentStrings.
func Noding_TestUtil_ToSegmentStrings(lines []*Geom_LineString) []*Noding_NodedSegmentString {
	nssList := make([]*Noding_NodedSegmentString, len(lines))
	for i, line := range lines {
		nssList[i] = Noding_NewNodedSegmentString(line.GetCoordinates(), line)
	}
	return nssList
}

// Noding_TestUtil_NodeValidated runs a noder on one or two sets of input
// geometries and validates that the result is fully noded.
func Noding_TestUtil_NodeValidated(geom1, geom2 *Geom_Geometry, noder Noding_Noder) *Geom_Geometry {
	lines := GeomUtil_LinearComponentExtracter_GetLines(geom1)
	if geom2 != nil {
		lines2 := GeomUtil_LinearComponentExtracter_GetLines(geom2)
		lines = append(lines, lines2...)
	}
	ssList := Noding_TestUtil_ToSegmentStrings(lines)

	// Convert to SegmentString slice.
	ssSlice := make([]Noding_SegmentString, len(ssList))
	for i, nss := range ssList {
		ssSlice[i] = nss
	}

	noderValid := Noding_NewValidatingNoder(noder)
	noderValid.ComputeNodes(ssSlice)
	nodedList := noderValid.GetNodedSubstrings()

	// Convert back to NodedSegmentString.
	nssResult := make([]*Noding_NodedSegmentString, len(nodedList))
	for i, ss := range nodedList {
		nssResult[i] = ss.(*Noding_NodedSegmentString)
	}

	return Noding_TestUtil_ToLines(nssResult, geom1.GetFactory())
}
