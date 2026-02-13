package jts

import "sort"

// OperationBuffer_OffsetCurveSection models a section of a raw offset curve,
// starting at a given location along the raw curve.
// The location is a decimal number, with the integer part
// containing the segment index and the fractional part
// giving the fractional distance along the segment.
// The location of the last section segment
// is also kept, to allow optimizing joining sections together.
type OperationBuffer_OffsetCurveSection struct {
	sectionPts []*Geom_Coordinate
	location   float64
	locLast    float64
}

// OperationBuffer_OffsetCurveSection_ToGeometry converts a list of offset curve sections to a geometry.
func OperationBuffer_OffsetCurveSection_ToGeometry(sections []*OperationBuffer_OffsetCurveSection, geomFactory *Geom_GeometryFactory) *Geom_Geometry {
	if len(sections) == 0 {
		return geomFactory.CreateLineString().Geom_Geometry
	}
	if len(sections) == 1 {
		return geomFactory.CreateLineStringFromCoordinates(sections[0].getCoordinates()).Geom_Geometry
	}

	//-- sort sections in order along the offset curve
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].CompareTo(sections[j]) < 0
	})
	lines := make([]*Geom_LineString, len(sections))

	for i := 0; i < len(sections); i++ {
		lines[i] = geomFactory.CreateLineStringFromCoordinates(sections[i].getCoordinates())
	}
	return geomFactory.CreateMultiLineStringFromLineStrings(lines).Geom_Geometry
}

// OperationBuffer_OffsetCurveSection_ToLine joins section coordinates into a LineString.
// Join vertices which lie in the same raw curve segment
// are removed, to simplify the result linework.
func OperationBuffer_OffsetCurveSection_ToLine(sections []*OperationBuffer_OffsetCurveSection, geomFactory *Geom_GeometryFactory) *Geom_Geometry {
	if len(sections) == 0 {
		return geomFactory.CreateLineString().Geom_Geometry
	}
	if len(sections) == 1 {
		return geomFactory.CreateLineStringFromCoordinates(sections[0].getCoordinates()).Geom_Geometry
	}

	//-- sort sections in order along the offset curve
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].CompareTo(sections[j]) < 0
	})
	pts := Geom_NewCoordinateList()

	removeStartPt := false
	for i := 0; i < len(sections); i++ {
		section := sections[i]

		removeEndPt := false
		if i < len(sections)-1 {
			nextStartLoc := sections[i+1].location
			removeEndPt = section.isEndInSameSegment(nextStartLoc)
		}
		sectionPts := section.getCoordinates()
		for j := 0; j < len(sectionPts); j++ {
			if (removeStartPt && j == 0) || (removeEndPt && j == len(sectionPts)-1) {
				continue
			}
			pts.AddCoordinate(sectionPts[j], false)
		}
		removeStartPt = removeEndPt
	}
	return geomFactory.CreateLineStringFromCoordinates(pts.ToCoordinateArray()).Geom_Geometry
}

// OperationBuffer_OffsetCurveSection_Create creates a new offset curve section.
func OperationBuffer_OffsetCurveSection_Create(srcPts []*Geom_Coordinate, start, end int, loc, locLast float64) *OperationBuffer_OffsetCurveSection {
	length := end - start + 1
	if end <= start {
		length = len(srcPts) - start + end
	}

	sectionPts := make([]*Geom_Coordinate, length)
	for i := 0; i < length; i++ {
		index := (start + i) % (len(srcPts) - 1)
		sectionPts[i] = srcPts[index].Copy()
	}
	return operationBuffer_newOffsetCurveSection(sectionPts, loc, locLast)
}

func operationBuffer_newOffsetCurveSection(pts []*Geom_Coordinate, loc, locLast float64) *OperationBuffer_OffsetCurveSection {
	return &OperationBuffer_OffsetCurveSection{
		sectionPts: pts,
		location:   loc,
		locLast:    locLast,
	}
}

func (ocs *OperationBuffer_OffsetCurveSection) getCoordinates() []*Geom_Coordinate {
	return ocs.sectionPts
}

func (ocs *OperationBuffer_OffsetCurveSection) isEndInSameSegment(nextLoc float64) bool {
	segIndex := int(ocs.locLast)
	nextIndex := int(nextLoc)
	return segIndex == nextIndex
}

// CompareTo orders sections by their location along the raw offset curve.
func (ocs *OperationBuffer_OffsetCurveSection) CompareTo(section *OperationBuffer_OffsetCurveSection) int {
	if ocs.location < section.location {
		return -1
	}
	if ocs.location > section.location {
		return 1
	}
	return 0
}
