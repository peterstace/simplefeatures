package jts

import "math"

// OperationOverlaySnap_LineStringSnapper snaps the vertices and segments of a
// LineString to a set of target snap vertices. A snap distance tolerance is
// used to control where snapping is performed.
//
// The implementation handles empty geometry and empty snap vertex sets.
type OperationOverlaySnap_LineStringSnapper struct {
	snapTolerance                 float64
	srcPts                        []*Geom_Coordinate
	seg                           *Geom_LineSegment
	allowSnappingToSourceVertices bool
	isClosed                      bool
}

// OperationOverlaySnap_NewLineStringSnapperFromLineString creates a new snapper
// using the points in the given LineString as source snap points.
func OperationOverlaySnap_NewLineStringSnapperFromLineString(srcLine *Geom_LineString, snapTolerance float64) *OperationOverlaySnap_LineStringSnapper {
	return OperationOverlaySnap_NewLineStringSnapperFromCoordinates(srcLine.GetCoordinates(), snapTolerance)
}

// OperationOverlaySnap_NewLineStringSnapperFromCoordinates creates a new snapper
// using the given points as source points to be snapped.
func OperationOverlaySnap_NewLineStringSnapperFromCoordinates(srcPts []*Geom_Coordinate, snapTolerance float64) *OperationOverlaySnap_LineStringSnapper {
	return &OperationOverlaySnap_LineStringSnapper{
		srcPts:        srcPts,
		isClosed:      operationOverlaySnap_lineStringSnapper_isClosed(srcPts),
		snapTolerance: snapTolerance,
		seg:           Geom_NewLineSegment(),
	}
}

// SetAllowSnappingToSourceVertices sets whether snapping to source vertices is allowed.
func (lss *OperationOverlaySnap_LineStringSnapper) SetAllowSnappingToSourceVertices(allowSnappingToSourceVertices bool) {
	lss.allowSnappingToSourceVertices = allowSnappingToSourceVertices
}

func operationOverlaySnap_lineStringSnapper_isClosed(pts []*Geom_Coordinate) bool {
	if len(pts) <= 1 {
		return false
	}
	return pts[0].Equals2D(pts[len(pts)-1])
}

// SnapTo snaps the vertices and segments of the source LineString to the given
// set of snap vertices.
func (lss *OperationOverlaySnap_LineStringSnapper) SnapTo(snapPts []*Geom_Coordinate) []*Geom_Coordinate {
	coordList := Geom_NewCoordinateListFromCoordinates(lss.srcPts)

	lss.snapVertices(coordList, snapPts)
	lss.snapSegments(coordList, snapPts)

	return coordList.ToCoordinateArray()
}

// snapVertices snaps source vertices to vertices in the target.
func (lss *OperationOverlaySnap_LineStringSnapper) snapVertices(srcCoords *Geom_CoordinateList, snapPts []*Geom_Coordinate) {
	// Try snapping vertices.
	// If src is a ring then don't snap final vertex.
	end := srcCoords.Size()
	if lss.isClosed {
		end = srcCoords.Size() - 1
	}
	for i := 0; i < end; i++ {
		srcPt := srcCoords.Get(i)
		snapVert := lss.findSnapForVertex(srcPt, snapPts)
		if snapVert != nil {
			// Update src with snap pt.
			srcCoords.Set(i, Geom_NewCoordinateFromCoordinate(snapVert))
			// Keep final closing point in synch (rings only).
			if i == 0 && lss.isClosed {
				srcCoords.Set(srcCoords.Size()-1, Geom_NewCoordinateFromCoordinate(snapVert))
			}
		}
	}
}

func (lss *OperationOverlaySnap_LineStringSnapper) findSnapForVertex(pt *Geom_Coordinate, snapPts []*Geom_Coordinate) *Geom_Coordinate {
	for i := 0; i < len(snapPts); i++ {
		// If point is already equal to a src pt, don't snap.
		if pt.Equals2D(snapPts[i]) {
			return nil
		}
		if pt.Distance(snapPts[i]) < lss.snapTolerance {
			return snapPts[i]
		}
	}
	return nil
}

// snapSegments snaps segments of the source to nearby snap vertices. Source
// segments are "cracked" at a snap vertex. A single input segment may be
// snapped several times to different snap vertices.
//
// For each distinct snap vertex, at most one source segment is snapped to.
// This prevents "cracking" multiple segments at the same point, which would
// likely cause topology collapse when being used on polygonal linework.
func (lss *OperationOverlaySnap_LineStringSnapper) snapSegments(srcCoords *Geom_CoordinateList, snapPts []*Geom_Coordinate) {
	// Guard against empty input.
	if len(snapPts) == 0 {
		return
	}

	distinctPtCount := len(snapPts)

	// Check for duplicate snap pts when they are sourced from a linear ring.
	// TODO: Need to do this better - need to check *all* snap points for dups (using a Set?).
	if snapPts[0].Equals2D(snapPts[len(snapPts)-1]) {
		distinctPtCount = len(snapPts) - 1
	}

	for i := 0; i < distinctPtCount; i++ {
		snapPt := snapPts[i]
		index := lss.findSegmentIndexToSnap(snapPt, srcCoords)
		// If a segment to snap to was found, "crack" it at the snap pt.
		// The new pt is inserted immediately into the src segment list,
		// so that subsequent snapping will take place on the modified segments.
		// Duplicate points are not added.
		if index >= 0 {
			srcCoords.AddCoordinateAtIndex(index+1, Geom_NewCoordinateFromCoordinate(snapPt), false)
		}
	}
}

// findSegmentIndexToSnap finds a src segment which snaps to (is close to) the
// given snap point.
//
// Only a single segment is selected for snapping. This prevents multiple
// segments snapping to the same snap vertex, which would almost certainly cause
// invalid geometry to be created. (The heuristic approach to snapping used here
// is really only appropriate when snap pts snap to a unique spot on the src
// geometry.)
//
// Also, if the snap vertex occurs as a vertex in the src coordinate list, no
// snapping is performed.
//
// Returns the index of the snapped segment or -1 if no segment snaps to the
// snap point.
func (lss *OperationOverlaySnap_LineStringSnapper) findSegmentIndexToSnap(snapPt *Geom_Coordinate, srcCoords *Geom_CoordinateList) int {
	minDist := math.MaxFloat64
	snapIndex := -1
	for i := 0; i < srcCoords.Size()-1; i++ {
		lss.seg.P0 = srcCoords.Get(i)
		lss.seg.P1 = srcCoords.Get(i + 1)

		// Check if the snap pt is equal to one of the segment endpoints.
		//
		// If the snap pt is already in the src list, don't snap at all.
		if lss.seg.P0.Equals2D(snapPt) || lss.seg.P1.Equals2D(snapPt) {
			if lss.allowSnappingToSourceVertices {
				continue
			}
			return -1
		}

		dist := lss.seg.DistanceToPoint(snapPt)
		if dist < lss.snapTolerance && dist < minDist {
			minDist = dist
			snapIndex = i
		}
	}
	return snapIndex
}
