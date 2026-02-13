package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationBuffer_BufferCurveSetBuilder creates all the raw offset curves for a buffer of a Geometry.
// Raw curves need to be noded together and polygonized to form the final buffer area.
type OperationBuffer_BufferCurveSetBuilder struct {
	inputGeom    *Geom_Geometry
	distance     float64
	curveBuilder *OperationBuffer_OffsetCurveBuilder

	curveList []Noding_SegmentString

	isInvertOrientation bool
}

// OperationBuffer_NewBufferCurveSetBuilder creates a new BufferCurveSetBuilder.
func OperationBuffer_NewBufferCurveSetBuilder(inputGeom *Geom_Geometry, distance float64, precisionModel *Geom_PrecisionModel, bufParams *OperationBuffer_BufferParameters) *OperationBuffer_BufferCurveSetBuilder {
	return &OperationBuffer_BufferCurveSetBuilder{
		inputGeom:    inputGeom,
		distance:     distance,
		curveBuilder: OperationBuffer_NewOffsetCurveBuilder(precisionModel, bufParams),
		curveList:    make([]Noding_SegmentString, 0),
	}
}

// SetInvertOrientation sets whether the offset curve is generated
// using the inverted orientation of input rings.
// This allows generating a buffer(0) polygon from the smaller lobes
// of self-crossing rings.
func (bcsb *OperationBuffer_BufferCurveSetBuilder) SetInvertOrientation(isInvertOrientation bool) {
	bcsb.isInvertOrientation = isInvertOrientation
}

// isRingCCW computes orientation of a ring using a signed-area orientation test.
// For invalid (self-crossing) rings this ensures the largest enclosed area
// is taken to be the interior of the ring.
// This produces a more sensible result when
// used for repairing polygonal geometry via buffer-by-zero.
// For buffer use the lower robustness of orientation-by-area
// doesn't matter, since narrow or flat rings
// produce an acceptable offset curve for either orientation.
func (bcsb *OperationBuffer_BufferCurveSetBuilder) isRingCCW(coord []*Geom_Coordinate) bool {
	isCCW := Algorithm_Orientation_IsCCWArea(coord)
	// invert orientation if required
	if bcsb.isInvertOrientation {
		return !isCCW
	}
	return isCCW
}

// GetCurves computes the set of raw offset curves for the buffer.
// Each offset curve has an attached Label indicating
// its left and right location.
//
// Returns a Collection of SegmentStrings representing the raw buffer curves.
func (bcsb *OperationBuffer_BufferCurveSetBuilder) GetCurves() []Noding_SegmentString {
	bcsb.add(bcsb.inputGeom)
	return bcsb.curveList
}

// addCurve creates a SegmentString for a coordinate list which is a raw offset curve,
// and adds it to the list of buffer curves.
// The SegmentString is tagged with a Label giving the topology of the curve.
// The curve may be oriented in either direction.
// If the curve is oriented CW, the locations will be:
//
//	Left: Location.EXTERIOR
//	Right: Location.INTERIOR
func (bcsb *OperationBuffer_BufferCurveSetBuilder) addCurve(coord []*Geom_Coordinate, leftLoc int, rightLoc int) {
	// don't add null or trivial curves
	if coord == nil || len(coord) < 2 {
		return
	}
	// add the edge for a coordinate list which is a raw offset curve
	e := Noding_NewNodedSegmentString(coord, Geomgraph_NewLabelGeomOnLeftRight(0, Geom_Location_Boundary, leftLoc, rightLoc))
	bcsb.curveList = append(bcsb.curveList, e)
}

func (bcsb *OperationBuffer_BufferCurveSetBuilder) add(g *Geom_Geometry) {
	if g.IsEmpty() {
		return
	}

	if java.InstanceOf[*Geom_Polygon](g) {
		bcsb.addPolygon(java.Cast[*Geom_Polygon](g))
	} else if java.InstanceOf[*Geom_LineString](g) {
		// LineString also handles LinearRings
		bcsb.addLineString(java.Cast[*Geom_LineString](g))
	} else if java.InstanceOf[*Geom_Point](g) {
		bcsb.addPoint(java.Cast[*Geom_Point](g))
	} else if java.InstanceOf[*Geom_MultiPoint](g) {
		bcsb.addCollection(java.Cast[*Geom_MultiPoint](g).Geom_GeometryCollection)
	} else if java.InstanceOf[*Geom_MultiLineString](g) {
		bcsb.addCollection(java.Cast[*Geom_MultiLineString](g).Geom_GeometryCollection)
	} else if java.InstanceOf[*Geom_MultiPolygon](g) {
		bcsb.addCollection(java.Cast[*Geom_MultiPolygon](g).Geom_GeometryCollection)
	} else if java.InstanceOf[*Geom_GeometryCollection](g) {
		bcsb.addCollection(java.Cast[*Geom_GeometryCollection](g))
	} else {
		panic("unsupported geometry type")
	}
}

func (bcsb *OperationBuffer_BufferCurveSetBuilder) addCollection(gc *Geom_GeometryCollection) {
	for i := 0; i < gc.GetNumGeometries(); i++ {
		g := gc.GetGeometryN(i)
		bcsb.add(g)
	}
}

// addPoint adds a Point to the graph.
func (bcsb *OperationBuffer_BufferCurveSetBuilder) addPoint(p *Geom_Point) {
	// a zero or negative width buffer of a point is empty
	if bcsb.distance <= 0.0 {
		return
	}
	coord := p.GetCoordinates()
	// skip if coordinate is invalid
	if len(coord) >= 1 && !coord[0].IsValid() {
		return
	}
	curve := bcsb.curveBuilder.GetLineCurve(coord, bcsb.distance)
	bcsb.addCurve(curve, Geom_Location_Exterior, Geom_Location_Interior)
}

func (bcsb *OperationBuffer_BufferCurveSetBuilder) addLineString(line *Geom_LineString) {
	if bcsb.curveBuilder.IsLineOffsetEmpty(bcsb.distance) {
		return
	}

	coord := operationBuffer_bufferCurveSetBuilder_clean(line.GetCoordinates())

	// Rings (closed lines) are generated with a continuous curve,
	// with no end arcs. This produces better quality linework,
	// and avoids noding issues with arcs around almost-parallel end segments.
	// See JTS #523 and #518.
	//
	// Singled-sided buffers currently treat rings as if they are lines.
	if Geom_CoordinateArrays_IsRing(coord) && !bcsb.curveBuilder.GetBufferParameters().IsSingleSided() {
		bcsb.addRingBothSides(coord, bcsb.distance)
	} else {
		curve := bcsb.curveBuilder.GetLineCurve(coord, bcsb.distance)
		bcsb.addCurve(curve, Geom_Location_Exterior, Geom_Location_Interior)
	}
}

// operationBuffer_bufferCurveSetBuilder_clean keeps only valid coordinates, and removes repeated points.
func operationBuffer_bufferCurveSetBuilder_clean(coords []*Geom_Coordinate) []*Geom_Coordinate {
	return Geom_CoordinateArrays_RemoveRepeatedOrInvalidPoints(coords)
}

func (bcsb *OperationBuffer_BufferCurveSetBuilder) addPolygon(p *Geom_Polygon) {
	offsetDistance := bcsb.distance
	offsetSide := Geom_Position_Left
	if bcsb.distance < 0.0 {
		offsetDistance = -bcsb.distance
		offsetSide = Geom_Position_Right
	}

	shell := p.GetExteriorRing()
	shellCoord := operationBuffer_bufferCurveSetBuilder_clean(shell.GetCoordinates())
	// optimization - don't bother computing buffer
	// if the polygon would be completely eroded
	if bcsb.distance < 0.0 && operationBuffer_bufferCurveSetBuilder_isErodedCompletely(shell, bcsb.distance) {
		return
	}
	// don't attempt to buffer a polygon with too few distinct vertices
	if bcsb.distance <= 0.0 && len(shellCoord) < 3 {
		return
	}

	bcsb.addRingSide(
		shellCoord,
		offsetDistance,
		offsetSide,
		Geom_Location_Exterior,
		Geom_Location_Interior)

	for i := 0; i < p.GetNumInteriorRing(); i++ {
		hole := p.GetInteriorRingN(i)
		holeCoord := operationBuffer_bufferCurveSetBuilder_clean(hole.GetCoordinates())

		// optimization - don't bother computing buffer for this hole
		// if the hole would be completely covered
		if bcsb.distance > 0.0 && operationBuffer_bufferCurveSetBuilder_isErodedCompletely(hole, -bcsb.distance) {
			continue
		}

		// Holes are topologically labelled opposite to the shell, since
		// the interior of the polygon lies on their opposite side
		// (on the left, if the hole is oriented CCW)
		bcsb.addRingSide(
			holeCoord,
			offsetDistance,
			Geom_Position_Opposite(offsetSide),
			Geom_Location_Interior,
			Geom_Location_Exterior)
	}
}

func (bcsb *OperationBuffer_BufferCurveSetBuilder) addRingBothSides(coord []*Geom_Coordinate, distance float64) {
	bcsb.addRingSide(coord, distance,
		Geom_Position_Left,
		Geom_Location_Exterior, Geom_Location_Interior)
	// Add the opposite side of the ring
	bcsb.addRingSide(coord, distance,
		Geom_Position_Right,
		Geom_Location_Interior, Geom_Location_Exterior)
}

// addRingSide adds an offset curve for one side of a ring.
// The side and left and right topological location arguments
// are provided as if the ring is oriented CW.
// (If the ring is in the opposite orientation,
// this is detected and the left and right locations are interchanged and the side is flipped.)
func (bcsb *OperationBuffer_BufferCurveSetBuilder) addRingSide(coord []*Geom_Coordinate, offsetDistance float64, side int, cwLeftLoc int, cwRightLoc int) {
	// don't bother adding ring if it is "flat" and will disappear in the output
	if offsetDistance == 0.0 && len(coord) < Geom_LinearRing_MinimumValidSize {
		return
	}

	leftLoc := cwLeftLoc
	rightLoc := cwRightLoc
	isCCW := bcsb.isRingCCW(coord)
	if len(coord) >= Geom_LinearRing_MinimumValidSize && isCCW {
		leftLoc = cwRightLoc
		rightLoc = cwLeftLoc
		side = Geom_Position_Opposite(side)
	}
	curve := bcsb.curveBuilder.GetRingCurve(coord, side, offsetDistance)

	// If the offset curve has inverted completely it will produce
	// an unwanted artifact in the result, so skip it.
	if operationBuffer_bufferCurveSetBuilder_isRingCurveInverted(coord, offsetDistance, curve) {
		return
	}

	bcsb.addCurve(curve, leftLoc, rightLoc)
}

const operationBuffer_bufferCurveSetBuilder_MAX_INVERTED_RING_SIZE = 9
const operationBuffer_bufferCurveSetBuilder_INVERTED_CURVE_VERTEX_FACTOR = 4
const operationBuffer_bufferCurveSetBuilder_NEARNESS_FACTOR = 0.99

// operationBuffer_bufferCurveSetBuilder_isRingCurveInverted tests whether the offset curve for a ring is fully inverted.
// An inverted ("inside-out") curve occurs in some specific situations
// involving a buffer distance which should result in a fully-eroded (empty) buffer.
// It can happen that the sides of a small, convex polygon
// produce offset segments which all cross one another to form
// a curve with inverted orientation.
// This happens at buffer distances slightly greater than the distance at
// which the buffer should disappear.
// The inverted curve will produce an incorrect non-empty buffer (for a shell)
// or an incorrect hole (for a hole).
// It must be discarded from the set of offset curves used in the buffer.
// Heuristics are used to reduce the number of cases which area checked,
// for efficiency and correctness.
func operationBuffer_bufferCurveSetBuilder_isRingCurveInverted(inputRing []*Geom_Coordinate, distance float64, curveRing []*Geom_Coordinate) bool {
	if distance == 0.0 {
		return false
	}
	// Only proper rings can invert.
	if len(inputRing) <= 3 {
		return false
	}
	// Heuristic based on low chance that a ring with many vertices will invert.
	// This low limit ensures this test is fairly efficient.
	if len(inputRing) >= operationBuffer_bufferCurveSetBuilder_MAX_INVERTED_RING_SIZE {
		return false
	}

	// Don't check curves which are much larger than the input.
	// This improves performance by avoiding checking some concave inputs
	// (which can produce fillet arcs with many more vertices)
	if len(curveRing) > operationBuffer_bufferCurveSetBuilder_INVERTED_CURVE_VERTEX_FACTOR*len(inputRing) {
		return false
	}

	// If curve contains points which are on the buffer,
	// it is not inverted and can be included in the raw curves.
	if operationBuffer_bufferCurveSetBuilder_hasPointOnBuffer(inputRing, distance, curveRing) {
		return false
	}

	// curve is inverted, so discard it
	return true
}

// operationBuffer_bufferCurveSetBuilder_hasPointOnBuffer tests if there are points on the raw offset curve which may
// lie on the final buffer curve
// (i.e. they are (approximately) at the buffer distance from the input ring).
// For efficiency this only tests a limited set of points on the curve.
func operationBuffer_bufferCurveSetBuilder_hasPointOnBuffer(inputRing []*Geom_Coordinate, distance float64, curveRing []*Geom_Coordinate) bool {
	distTol := operationBuffer_bufferCurveSetBuilder_NEARNESS_FACTOR * math.Abs(distance)

	for i := 0; i < len(curveRing)-1; i++ {
		v := curveRing[i]

		// check curve vertices
		dist := Algorithm_Distance_PointToSegmentString(v, inputRing)
		if dist > distTol {
			return true
		}

		// check curve segment midpoints
		iNext := i + 1
		if i >= len(curveRing)-1 {
			iNext = 0
		}
		vnext := curveRing[iNext]
		midPt := Geom_LineSegment_MidPoint(v, vnext)

		distMid := Algorithm_Distance_PointToSegmentString(midPt, inputRing)
		if distMid > distTol {
			return true
		}
	}
	return false
}

// operationBuffer_bufferCurveSetBuilder_isErodedCompletely tests whether a ring buffer is eroded completely (is empty)
// based on simple heuristics.
//
// The ringCoord is assumed to contain no repeated points.
// It may be degenerate (i.e. contain only 1, 2, or 3 points).
// In this case it has no area, and hence has a minimum diameter of 0.
func operationBuffer_bufferCurveSetBuilder_isErodedCompletely(ring *Geom_LinearRing, bufferDistance float64) bool {
	ringCoord := ring.GetCoordinates()
	// degenerate ring has no area
	if len(ringCoord) < 4 {
		return bufferDistance < 0
	}

	// important test to eliminate inverted triangle bug
	// also optimizes erosion test for triangles
	if len(ringCoord) == 4 {
		return operationBuffer_bufferCurveSetBuilder_isTriangleErodedCompletely(ringCoord, bufferDistance)
	}

	// if envelope is narrower than twice the buffer distance, ring is eroded
	env := ring.GetEnvelopeInternal()
	envMinDimension := math.Min(env.GetHeight(), env.GetWidth())
	if bufferDistance < 0.0 && 2*math.Abs(bufferDistance) > envMinDimension {
		return true
	}

	return false
}

// operationBuffer_bufferCurveSetBuilder_isTriangleErodedCompletely tests whether a triangular ring would be eroded completely by the given
// buffer distance.
// This is a precise test. It uses the fact that the inner buffer of a
// triangle converges on the inCentre of the triangle (the point
// equidistant from all sides). If the buffer distance is greater than the
// distance of the inCentre from a side, the triangle will be eroded completely.
//
// This test is important, since it removes a problematic case where
// the buffer distance is slightly larger than the inCentre distance.
// In this case the triangle buffer curve "inverts" with incorrect topology,
// producing an incorrect hole in the buffer.
func operationBuffer_bufferCurveSetBuilder_isTriangleErodedCompletely(triangleCoord []*Geom_Coordinate, bufferDistance float64) bool {
	tri := Geom_NewTriangle(triangleCoord[0], triangleCoord[1], triangleCoord[2])
	inCentre := tri.InCentre()
	distToCentre := Algorithm_Distance_PointToSegment(inCentre, tri.P0, tri.P1)
	return distToCentre < math.Abs(bufferDistance)
}
