package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const operationBuffer_variableBuffer_MIN_CAP_SEG_LEN_FACTOR = 4

// OperationBuffer_VariableBuffer_Buffer creates a buffer polygon along a line with the buffer distance interpolated
// between a start distance and an end distance.
func OperationBuffer_VariableBuffer_Buffer(line *Geom_Geometry, startDistance, endDistance float64) *Geom_Geometry {
	distance := operationBuffer_variableBuffer_interpolate(java.Cast[*Geom_LineString](line), startDistance, endDistance)
	vb := operationBuffer_newVariableBuffer(line, distance)
	return vb.getResult()
}

// OperationBuffer_VariableBuffer_BufferWithMidDistance creates a buffer polygon along a line with the buffer distance interpolated
// between a start distance, a middle distance and an end distance.
// The middle distance is attained at
// the vertex at or just past the half-length of the line.
// For smooth buffering of a LinearRing (or the rings of a Polygon)
// the start distance and end distance should be equal.
func OperationBuffer_VariableBuffer_BufferWithMidDistance(line *Geom_Geometry, startDistance, midDistance, endDistance float64) *Geom_Geometry {
	distance := operationBuffer_variableBuffer_interpolateMid(java.Cast[*Geom_LineString](line), startDistance, midDistance, endDistance)
	vb := operationBuffer_newVariableBuffer(line, distance)
	return vb.getResult()
}

// OperationBuffer_VariableBuffer_BufferWithDistances creates a buffer polygon along a line with the distance specified
// at each vertex.
func OperationBuffer_VariableBuffer_BufferWithDistances(line *Geom_Geometry, distance []float64) *Geom_Geometry {
	vb := operationBuffer_newVariableBuffer(line, distance)
	return vb.getResult()
}

// operationBuffer_variableBuffer_interpolate computes a list of values for the points along a line by
// interpolating between values for the start and end point.
// The interpolation is based on the distance of each point along the line
// relative to the total line length.
func operationBuffer_variableBuffer_interpolate(line *Geom_LineString, startValue, endValue float64) []float64 {
	startValue = math.Abs(startValue)
	endValue = math.Abs(endValue)
	values := make([]float64, line.GetNumPoints())
	values[0] = startValue
	values[len(values)-1] = endValue

	totalLen := line.GetLength()
	pts := line.GetCoordinates()
	currLen := 0.0
	for i := 1; i < len(values)-1; i++ {
		segLen := pts[i].Distance(pts[i-1])
		currLen += segLen
		lenFrac := currLen / totalLen
		delta := lenFrac * (endValue - startValue)
		values[i] = startValue + delta
	}
	return values
}

// operationBuffer_variableBuffer_interpolateMid computes a list of values for the points along a line by
// interpolating between values for the start, middle and end points.
// The interpolation is based on the distance of each point along the line
// relative to the total line length.
// The middle distance is attained at the vertex at or just past the half-length of the line.
func operationBuffer_variableBuffer_interpolateMid(line *Geom_LineString, startValue, midValue, endValue float64) []float64 {
	startValue = math.Abs(startValue)
	midValue = math.Abs(midValue)
	endValue = math.Abs(endValue)

	values := make([]float64, line.GetNumPoints())
	values[0] = startValue
	values[len(values)-1] = endValue

	pts := line.GetCoordinates()
	lineLen := line.GetLength()
	midIndex := operationBuffer_variableBuffer_indexAtLength(pts, lineLen/2)

	delMidStart := midValue - startValue
	delEndMid := endValue - midValue

	lenSM := operationBuffer_variableBuffer_length(pts, 0, midIndex)
	currLen := 0.0
	for i := 1; i <= midIndex; i++ {
		segLen := pts[i].Distance(pts[i-1])
		currLen += segLen
		lenFrac := currLen / lenSM
		val := startValue + lenFrac*delMidStart
		values[i] = val
	}

	lenME := operationBuffer_variableBuffer_length(pts, midIndex, len(pts)-1)
	currLen = 0
	for i := midIndex + 1; i < len(values)-1; i++ {
		segLen := pts[i].Distance(pts[i-1])
		currLen += segLen
		lenFrac := currLen / lenME
		val := midValue + lenFrac*delEndMid
		values[i] = val
	}
	return values
}

func operationBuffer_variableBuffer_indexAtLength(pts []*Geom_Coordinate, targetLen float64) int {
	length := 0.0
	for i := 1; i < len(pts); i++ {
		length += pts[i].Distance(pts[i-1])
		if length > targetLen {
			return i
		}
	}
	return len(pts) - 1
}

func operationBuffer_variableBuffer_length(pts []*Geom_Coordinate, i1, i2 int) float64 {
	length := 0.0
	for i := i1 + 1; i <= i2; i++ {
		length += pts[i].Distance(pts[i-1])
	}
	return length
}

// operationBuffer_VariableBuffer creates a buffer polygon with a varying buffer distance
// at each vertex along a line.
type operationBuffer_VariableBuffer struct {
	line         *Geom_LineString
	distance     []float64
	geomFactory  *Geom_GeometryFactory
	quadrantSegs int
}

// operationBuffer_newVariableBuffer creates a generator for a variable-distance line buffer.
func operationBuffer_newVariableBuffer(line *Geom_Geometry, distance []float64) *operationBuffer_VariableBuffer {
	vb := &operationBuffer_VariableBuffer{
		line:         java.Cast[*Geom_LineString](line),
		distance:     distance,
		geomFactory:  line.GetFactory(),
		quadrantSegs: OperationBuffer_BufferParameters_DEFAULT_QUADRANT_SEGMENTS,
	}

	if len(distance) != vb.line.GetNumPoints() {
		panic("Number of distances is not equal to number of vertices")
	}

	return vb
}

// getResult computes the variable buffer polygon.
func (vb *operationBuffer_VariableBuffer) getResult() *Geom_Geometry {
	var parts []*Geom_Geometry

	pts := vb.line.GetCoordinates()
	// construct segment buffers
	for i := 1; i < len(pts); i++ {
		dist0 := vb.distance[i-1]
		dist1 := vb.distance[i]
		if dist0 > 0 || dist1 > 0 {
			poly := vb.segmentBuffer(pts[i-1], pts[i], dist0, dist1)
			if poly != nil {
				parts = append(parts, poly.Geom_Geometry)
			}
		}
	}

	partsGeom := vb.geomFactory.CreateGeometryCollectionFromGeometries(Geom_GeometryFactory_ToGeometryArray(parts))
	buffer := partsGeom.Geom_Geometry.UnionSelf()

	//-- ensure an empty polygon is returned if needed
	if buffer.IsEmpty() {
		return vb.geomFactory.CreatePolygon().Geom_Geometry
	}
	return buffer
}

// segmentBuffer computes a variable buffer polygon for a single segment,
// with the given endpoints and buffer distances.
// The individual segment buffers are unioned to form the final buffer.
// If one distance is zero, the end cap at that segment end is the endpoint of the segment.
// If both distances are zero, no polygon is returned.
func (vb *operationBuffer_VariableBuffer) segmentBuffer(p0, p1 *Geom_Coordinate, dist0, dist1 float64) *Geom_Polygon {
	// Skip buffer polygon if both distances are zero
	if dist0 <= 0 && dist1 <= 0 {
		return nil
	}

	// Generation algorithm requires increasing distance, so flip if needed
	if dist0 > dist1 {
		return vb.segmentBufferOriented(p1, p0, dist1, dist0)
	}
	return vb.segmentBufferOriented(p0, p1, dist0, dist1)
}

func (vb *operationBuffer_VariableBuffer) segmentBufferOriented(p0, p1 *Geom_Coordinate, dist0, dist1 float64) *Geom_Polygon {
	//-- Assert: dist0 <= dist1

	//-- forward tangent line
	tangent := operationBuffer_variableBuffer_outerTangent(p0, dist0, p1, dist1)

	//-- if tangent is null then compute a buffer for largest circle
	if tangent == nil {
		center := p0
		dist := dist0
		if dist1 > dist0 {
			center = p1
			dist = dist1
		}
		return vb.circle(center, dist)
	}

	//-- reverse tangent line on other side of segment
	tangentReflect := vb.reflect(tangent, p0, p1, dist0)

	coords := Geom_NewCoordinateList()
	//-- end cap
	vb.addCap(p1, dist1, tangent.P1, tangentReflect.P1, coords)
	//-- start cap
	vb.addCap(p0, dist0, tangentReflect.P0, tangent.P0, coords)

	coords.CloseRing()

	pts := coords.ToCoordinateArray()
	polygon := vb.geomFactory.CreatePolygonFromCoordinates(pts)
	return polygon
}

func (vb *operationBuffer_VariableBuffer) reflect(seg *Geom_LineSegment, p0, p1 *Geom_Coordinate, dist0 float64) *Geom_LineSegment {
	line := Geom_NewLineSegmentFromCoordinates(p0, p1)
	r0 := line.Reflect(seg.P0)
	r1 := line.Reflect(seg.P1)
	//-- avoid numeric jitter if first distance is zero (second dist must be > 0)
	if dist0 == 0 {
		r0 = p0.Copy()
	}
	return Geom_NewLineSegmentFromCoordinates(r0, r1)
}

// circle returns a circular polygon.
func (vb *operationBuffer_VariableBuffer) circle(center *Geom_Coordinate, radius float64) *Geom_Polygon {
	if radius <= 0 {
		return nil
	}
	nPts := 4 * vb.quadrantSegs
	pts := make([]*Geom_Coordinate, nPts+1)
	angInc := math.Pi / 2 / float64(vb.quadrantSegs)
	for i := 0; i < nPts; i++ {
		pts[i] = operationBuffer_variableBuffer_projectPolar(center, radius, float64(i)*angInc)
	}
	pts[len(pts)-1] = pts[0].Copy()
	return vb.geomFactory.CreatePolygonFromCoordinates(pts)
}

// addCap adds a semi-circular cap CCW around the point p.
// The vertices in caps are generated at fixed angles around a point.
// This allows caps at the same point to share vertices,
// which reduces artifacts when the segment buffers are merged.
func (vb *operationBuffer_VariableBuffer) addCap(p *Geom_Coordinate, r float64, t1, t2 *Geom_Coordinate, coords *Geom_CoordinateList) {
	//-- if radius is zero just copy the vertex
	if r == 0 {
		coords.AddCoordinate(p.Copy(), false)
		return
	}

	coords.AddCoordinate(t1, false)

	angStart := Algorithm_Angle_AngleBetweenPoints(p, t1)
	angEnd := Algorithm_Angle_AngleBetweenPoints(p, t2)
	if angStart < angEnd {
		angStart += 2 * math.Pi
	}

	indexStart := vb.capAngleIndex(angStart)
	indexEnd := vb.capAngleIndex(angEnd)

	capSegLen := r * 2 * math.Sin(math.Pi/4/float64(vb.quadrantSegs))
	minSegLen := capSegLen / operationBuffer_variableBuffer_MIN_CAP_SEG_LEN_FACTOR

	for i := indexStart; i >= indexEnd; i-- {
		//-- use negative increment to create points CW
		ang := vb.capAngle(i)
		capPt := operationBuffer_variableBuffer_projectPolar(p, r, ang)

		isCapPointHighQuality := true
		// Due to the fixed locations of the cap points,
		// a start or end cap point might create
		// a "reversed" segment to the next tangent point.
		// This causes an unwanted narrow spike in the buffer curve,
		// which can cause holes in the final buffer polygon.
		// These checks remove these points.
		if i == indexStart && Algorithm_Orientation_Clockwise != Algorithm_Orientation_Index(p, t1, capPt) {
			isCapPointHighQuality = false
		} else if i == indexEnd && Algorithm_Orientation_Counterclockwise != Algorithm_Orientation_Index(p, t2, capPt) {
			isCapPointHighQuality = false
		}

		// Remove short segments between the cap and the tangent segments.
		if capPt.Distance(t1) < minSegLen {
			isCapPointHighQuality = false
		} else if capPt.Distance(t2) < minSegLen {
			isCapPointHighQuality = false
		}

		if isCapPointHighQuality {
			coords.AddCoordinate(capPt, false)
		}
	}

	coords.AddCoordinate(t2, false)
}

// capAngle computes the actual angle for a cap angle index.
func (vb *operationBuffer_VariableBuffer) capAngle(index int) float64 {
	capSegAng := math.Pi / 2 / float64(vb.quadrantSegs)
	return float64(index) * capSegAng
}

// capAngleIndex computes the canonical cap point index for a given angle.
// The angle is rounded down to the next lower index.
// In order to reduce the number of points created by overlapping end caps,
// cap points are generated at the same locations around a circle.
// The index is the index of the points around the circle,
// with 0 being the point at (1,0).
// The total number of points around the circle is 4 * quadrantSegs.
func (vb *operationBuffer_VariableBuffer) capAngleIndex(ang float64) int {
	capSegAng := math.Pi / 2 / float64(vb.quadrantSegs)
	index := int(ang / capSegAng)
	return index
}

// operationBuffer_variableBuffer_outerTangent computes the two circumference points defining the outer tangent line
// between two circles.
// The tangent line may be null if one circle mostly overlaps the other.
// For the algorithm see https://en.wikipedia.org/wiki/Tangent_lines_to_circles#Outer_tangent.
func operationBuffer_variableBuffer_outerTangent(c1 *Geom_Coordinate, r1 float64, c2 *Geom_Coordinate, r2 float64) *Geom_LineSegment {
	// If distances are inverted then flip to compute and flip result back.
	if r1 > r2 {
		seg := operationBuffer_variableBuffer_outerTangent(c2, r2, c1, r1)
		return Geom_NewLineSegmentFromCoordinates(seg.P1, seg.P0)
	}
	x1 := c1.GetX()
	y1 := c1.GetY()
	x2 := c2.GetX()
	y2 := c2.GetY()
	// TODO: handle r1 == r2?
	a3 := -math.Atan2(y2-y1, x2-x1)

	dr := r2 - r1
	d := math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))

	a2 := math.Asin(dr / d)
	// check if no tangent exists
	if math.IsNaN(a2) {
		return nil
	}

	a1 := a3 - a2

	aa := math.Pi/2 - a1
	x3 := x1 + r1*math.Cos(aa)
	y3 := y1 + r1*math.Sin(aa)
	x4 := x2 + r2*math.Cos(aa)
	y4 := y2 + r2*math.Sin(aa)

	return Geom_NewLineSegmentFromXY(x3, y3, x4, y4)
}

func operationBuffer_variableBuffer_projectPolar(p *Geom_Coordinate, r, ang float64) *Geom_Coordinate {
	x := p.GetX() + r*operationBuffer_variableBuffer_snapTrig(math.Cos(ang))
	y := p.GetY() + r*operationBuffer_variableBuffer_snapTrig(math.Sin(ang))
	return Geom_NewCoordinateWithXY(x, y)
}

const operationBuffer_variableBuffer_SNAP_TRIG_TOL = 1e-6

// operationBuffer_variableBuffer_snapTrig snaps trig values to integer values for better consistency.
func operationBuffer_variableBuffer_snapTrig(x float64) float64 {
	if x > (1 - operationBuffer_variableBuffer_SNAP_TRIG_TOL) {
		return 1
	}
	if x < (-1 + operationBuffer_variableBuffer_SNAP_TRIG_TOL) {
		return -1
	}
	if math.Abs(x) < operationBuffer_variableBuffer_SNAP_TRIG_TOL {
		return 0
	}
	return x
}
