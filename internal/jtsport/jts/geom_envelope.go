package jts

import (
	"fmt"
	"math"
)

// HashCode computes a hash code for this envelope.
func (e *Geom_Envelope) HashCode() int {
	result := 17
	result = 37*result + Geom_Coordinate_HashCodeFloat64(e.minx)
	result = 37*result + Geom_Coordinate_HashCodeFloat64(e.maxx)
	result = 37*result + Geom_Coordinate_HashCodeFloat64(e.miny)
	result = 37*result + Geom_Coordinate_HashCodeFloat64(e.maxy)
	return result
}

// Geom_Envelope_IntersectsPointEnvelope tests if the point q intersects the Geom_Envelope defined
// by p1-p2.
func Geom_Envelope_IntersectsPointEnvelope(p1, p2, q *Geom_Coordinate) bool {
	x1, x2 := p1.X, p2.X
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	y1, y2 := p1.Y, p2.Y
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if q.X >= x1 && q.X <= x2 && q.Y >= y1 && q.Y <= y2 {
		return true
	}
	return false
}

// Geom_Envelope_IntersectsEnvelopeEnvelope tests whether the envelope defined by p1-p2 and
// the envelope defined by q1-q2 intersect.
func Geom_Envelope_IntersectsEnvelopeEnvelope(p1, p2, q1, q2 *Geom_Coordinate) bool {
	minq := math.Min(q1.X, q2.X)
	maxq := math.Max(q1.X, q2.X)
	minp := math.Min(p1.X, p2.X)
	maxp := math.Max(p1.X, p2.X)

	if minp > maxq {
		return false
	}
	if maxp < minq {
		return false
	}

	minq = math.Min(q1.Y, q2.Y)
	maxq = math.Max(q1.Y, q2.Y)
	minp = math.Min(p1.Y, p2.Y)
	maxp = math.Max(p1.Y, p2.Y)

	if minp > maxq {
		return false
	}
	if maxp < minq {
		return false
	}
	return true
}

// Geom_Envelope defines a rectangular region of the 2D coordinate plane. It is often
// used to represent the bounding box of a Geometry, e.g. the minimum and
// maximum x and y values of the Coordinates.
//
// Envelopes support infinite or half-infinite regions, by using the values of
// math.Inf(1) and math.Inf(-1). Geom_Envelope objects may have a null value.
//
// When Geom_Envelope objects are created or initialized, the supplied extent values
// are automatically sorted into the correct order.
type Geom_Envelope struct {
	minx float64
	maxx float64
	miny float64
	maxy float64
}

// Geom_NewEnvelope creates a null Geom_Envelope.
func Geom_NewEnvelope() *Geom_Envelope {
	e := &Geom_Envelope{}
	e.Init()
	return e
}

// Geom_NewEnvelopeFromXY creates a Geom_Envelope for a region defined by maximum and
// minimum values.
func Geom_NewEnvelopeFromXY(x1, x2, y1, y2 float64) *Geom_Envelope {
	e := &Geom_Envelope{}
	e.InitFromXY(x1, x2, y1, y2)
	return e
}

// Geom_NewEnvelopeFromCoordinates creates a Geom_Envelope for a region defined by two
// Geom_Coordinates.
func Geom_NewEnvelopeFromCoordinates(p1, p2 *Geom_Coordinate) *Geom_Envelope {
	e := &Geom_Envelope{}
	e.InitFromXY(p1.X, p2.X, p1.Y, p2.Y)
	return e
}

// Geom_NewEnvelopeFromCoordinate creates a Geom_Envelope for a region defined by a
// single Geom_Coordinate.
func Geom_NewEnvelopeFromCoordinate(p *Geom_Coordinate) *Geom_Envelope {
	e := &Geom_Envelope{}
	e.InitFromXY(p.X, p.X, p.Y, p.Y)
	return e
}

// Geom_NewEnvelopeFromEnvelope creates a Geom_Envelope from an existing Geom_Envelope.
func Geom_NewEnvelopeFromEnvelope(env *Geom_Envelope) *Geom_Envelope {
	e := &Geom_Envelope{}
	e.InitFromEnvelope(env)
	return e
}

// Init initializes to a null Geom_Envelope.
func (e *Geom_Envelope) Init() {
	e.SetToNull()
}

// InitFromXY initializes a Geom_Envelope for a region defined by maximum and
// minimum values.
func (e *Geom_Envelope) InitFromXY(x1, x2, y1, y2 float64) {
	if x1 < x2 {
		e.minx = x1
		e.maxx = x2
	} else {
		e.minx = x2
		e.maxx = x1
	}
	if y1 < y2 {
		e.miny = y1
		e.maxy = y2
	} else {
		e.miny = y2
		e.maxy = y1
	}
}

// Copy creates a copy of this envelope object.
func (e *Geom_Envelope) Copy() *Geom_Envelope {
	return Geom_NewEnvelopeFromEnvelope(e)
}

// InitFromCoordinates initializes a Geom_Envelope to a region defined by two
// Geom_Coordinates.
func (e *Geom_Envelope) InitFromCoordinates(p1, p2 *Geom_Coordinate) {
	e.InitFromXY(p1.X, p2.X, p1.Y, p2.Y)
}

// InitFromCoordinate initializes a Geom_Envelope to a region defined by a single
// Geom_Coordinate.
func (e *Geom_Envelope) InitFromCoordinate(p *Geom_Coordinate) {
	e.InitFromXY(p.X, p.X, p.Y, p.Y)
}

// InitFromEnvelope initializes a Geom_Envelope from an existing Geom_Envelope.
func (e *Geom_Envelope) InitFromEnvelope(env *Geom_Envelope) {
	e.minx = env.minx
	e.maxx = env.maxx
	e.miny = env.miny
	e.maxy = env.maxy
}

// SetToNull makes this Geom_Envelope a "null" envelope, that is, the envelope of
// the empty geometry.
func (e *Geom_Envelope) SetToNull() {
	e.minx = 0
	e.maxx = -1
	e.miny = 0
	e.maxy = -1
}

// IsNull returns true if this Geom_Envelope is a "null" envelope.
func (e *Geom_Envelope) IsNull() bool {
	return e.maxx < e.minx
}

// GetWidth returns the difference between the maximum and minimum x values.
func (e *Geom_Envelope) GetWidth() float64 {
	if e.IsNull() {
		return 0
	}
	return e.maxx - e.minx
}

// GetHeight returns the difference between the maximum and minimum y values.
func (e *Geom_Envelope) GetHeight() float64 {
	if e.IsNull() {
		return 0
	}
	return e.maxy - e.miny
}

// GetDiameter gets the length of the diameter (diagonal) of the envelope.
func (e *Geom_Envelope) GetDiameter() float64 {
	if e.IsNull() {
		return 0
	}
	w := e.GetWidth()
	h := e.GetHeight()
	return math.Hypot(w, h)
}

// GetMinX returns the Geom_Envelope's minimum x-value. min x > max x indicates that
// this is a null Geom_Envelope.
func (e *Geom_Envelope) GetMinX() float64 {
	return e.minx
}

// GetMaxX returns the Geom_Envelope's maximum x-value. min x > max x indicates that
// this is a null Geom_Envelope.
func (e *Geom_Envelope) GetMaxX() float64 {
	return e.maxx
}

// GetMinY returns the Geom_Envelope's minimum y-value. min y > max y indicates that
// this is a null Geom_Envelope.
func (e *Geom_Envelope) GetMinY() float64 {
	return e.miny
}

// GetMaxY returns the Geom_Envelope's maximum y-value. min y > max y indicates that
// this is a null Geom_Envelope.
func (e *Geom_Envelope) GetMaxY() float64 {
	return e.maxy
}

// GetArea gets the area of this envelope.
func (e *Geom_Envelope) GetArea() float64 {
	return e.GetWidth() * e.GetHeight()
}

// MinExtent gets the minimum extent of this envelope across both dimensions.
func (e *Geom_Envelope) MinExtent() float64 {
	if e.IsNull() {
		return 0.0
	}
	w := e.GetWidth()
	h := e.GetHeight()
	if w < h {
		return w
	}
	return h
}

// MaxExtent gets the maximum extent of this envelope across both dimensions.
func (e *Geom_Envelope) MaxExtent() float64 {
	if e.IsNull() {
		return 0.0
	}
	w := e.GetWidth()
	h := e.GetHeight()
	if w > h {
		return w
	}
	return h
}

// ExpandToIncludeCoordinate enlarges this Geom_Envelope so that it contains the
// given Geom_Coordinate. Has no effect if the point is already on or within the
// envelope.
func (e *Geom_Envelope) ExpandToIncludeCoordinate(p *Geom_Coordinate) {
	e.ExpandToIncludeXY(p.X, p.Y)
}

// ExpandBy expands this envelope by a given distance in all directions. Both
// positive and negative distances are supported.
func (e *Geom_Envelope) ExpandBy(distance float64) {
	e.ExpandByXY(distance, distance)
}

// ExpandByXY expands this envelope by given distances in the X and Y
// directions. Both positive and negative distances are supported.
func (e *Geom_Envelope) ExpandByXY(deltaX, deltaY float64) {
	if e.IsNull() {
		return
	}

	e.minx -= deltaX
	e.maxx += deltaX
	e.miny -= deltaY
	e.maxy += deltaY

	if e.minx > e.maxx || e.miny > e.maxy {
		e.SetToNull()
	}
}

// ExpandToIncludeXY enlarges this Geom_Envelope so that it contains the given point.
// Has no effect if the point is already on or within the envelope.
func (e *Geom_Envelope) ExpandToIncludeXY(x, y float64) {
	if e.IsNull() {
		e.minx = x
		e.maxx = x
		e.miny = y
		e.maxy = y
	} else {
		if x < e.minx {
			e.minx = x
		}
		if x > e.maxx {
			e.maxx = x
		}
		if y < e.miny {
			e.miny = y
		}
		if y > e.maxy {
			e.maxy = y
		}
	}
}

// ExpandToIncludeEnvelope enlarges this Geom_Envelope so that it contains the other
// Geom_Envelope. Has no effect if other is wholly on or within the envelope.
func (e *Geom_Envelope) ExpandToIncludeEnvelope(other *Geom_Envelope) {
	if other.IsNull() {
		return
	}
	if e.IsNull() {
		e.minx = other.GetMinX()
		e.maxx = other.GetMaxX()
		e.miny = other.GetMinY()
		e.maxy = other.GetMaxY()
	} else {
		if other.minx < e.minx {
			e.minx = other.minx
		}
		if other.maxx > e.maxx {
			e.maxx = other.maxx
		}
		if other.miny < e.miny {
			e.miny = other.miny
		}
		if other.maxy > e.maxy {
			e.maxy = other.maxy
		}
	}
}

// Translate translates this envelope by given amounts in the X and Y direction.
func (e *Geom_Envelope) Translate(transX, transY float64) {
	if e.IsNull() {
		return
	}
	e.InitFromXY(e.GetMinX()+transX, e.GetMaxX()+transX,
		e.GetMinY()+transY, e.GetMaxY()+transY)
}

// Centre computes the coordinate of the centre of this envelope (as long as it
// is non-null). Returns nil if the envelope is null.
func (e *Geom_Envelope) Centre() *Geom_Coordinate {
	if e.IsNull() {
		return nil
	}
	return Geom_NewCoordinateWithXY(
		(e.GetMinX()+e.GetMaxX())/2.0,
		(e.GetMinY()+e.GetMaxY())/2.0)
}

// Intersection computes the intersection of two Geom_Envelopes. Returns a new
// Geom_Envelope representing the intersection of the envelopes (this will be the
// null envelope if either argument is null, or they do not intersect).
func (e *Geom_Envelope) Intersection(env *Geom_Envelope) *Geom_Envelope {
	if e.IsNull() || env.IsNull() || !e.IntersectsEnvelope(env) {
		return Geom_NewEnvelope()
	}
	intMinX := e.minx
	if env.minx > intMinX {
		intMinX = env.minx
	}
	intMinY := e.miny
	if env.miny > intMinY {
		intMinY = env.miny
	}
	intMaxX := e.maxx
	if env.maxx < intMaxX {
		intMaxX = env.maxx
	}
	intMaxY := e.maxy
	if env.maxy < intMaxY {
		intMaxY = env.maxy
	}
	return Geom_NewEnvelopeFromXY(intMinX, intMaxX, intMinY, intMaxY)
}

// IntersectsEnvelope tests if the region defined by other intersects the region
// of this Geom_Envelope. A null envelope never intersects.
func (e *Geom_Envelope) IntersectsEnvelope(other *Geom_Envelope) bool {
	if e.IsNull() || other.IsNull() {
		return false
	}
	return !(other.minx > e.maxx ||
		other.maxx < e.minx ||
		other.miny > e.maxy ||
		other.maxy < e.miny)
}

// IntersectsCoordinates tests if the extent defined by two extremal points
// intersects the extent of this Geom_Envelope.
func (e *Geom_Envelope) IntersectsCoordinates(a, b *Geom_Coordinate) bool {
	if e.IsNull() {
		return false
	}

	envminx := a.X
	if b.X < envminx {
		envminx = b.X
	}
	if envminx > e.maxx {
		return false
	}

	envmaxx := a.X
	if b.X > envmaxx {
		envmaxx = b.X
	}
	if envmaxx < e.minx {
		return false
	}

	envminy := a.Y
	if b.Y < envminy {
		envminy = b.Y
	}
	if envminy > e.maxy {
		return false
	}

	envmaxy := a.Y
	if b.Y > envmaxy {
		envmaxy = b.Y
	}
	if envmaxy < e.miny {
		return false
	}

	return true
}

// Disjoint tests if the region defined by other is disjoint from the region of
// this Geom_Envelope. A null envelope is always disjoint.
func (e *Geom_Envelope) Disjoint(other *Geom_Envelope) bool {
	return !e.IntersectsEnvelope(other)
}

// IntersectsCoordinate tests if the point p intersects (lies inside) the region
// of this Geom_Envelope.
func (e *Geom_Envelope) IntersectsCoordinate(p *Geom_Coordinate) bool {
	return e.IntersectsXY(p.X, p.Y)
}

// IntersectsXY checks if the point (x, y) intersects (lies inside) the region
// of this Geom_Envelope.
func (e *Geom_Envelope) IntersectsXY(x, y float64) bool {
	if e.IsNull() {
		return false
	}
	return !(x > e.maxx ||
		x < e.minx ||
		y > e.maxy ||
		y < e.miny)
}

// ContainsEnvelope tests if the Geom_Envelope other lies wholly inside this Geom_Envelope
// (inclusive of the boundary).
//
// Note that this is not the same definition as the SFS contains, which would
// exclude the envelope boundary.
func (e *Geom_Envelope) ContainsEnvelope(other *Geom_Envelope) bool {
	return e.CoversEnvelope(other)
}

// ContainsCoordinate tests if the given point lies in or on the envelope.
//
// Note that this is not the same definition as the SFS contains, which would
// exclude the envelope boundary.
func (e *Geom_Envelope) ContainsCoordinate(p *Geom_Coordinate) bool {
	return e.CoversCoordinate(p)
}

// ContainsXY tests if the given point lies in or on the envelope.
//
// Note that this is not the same definition as the SFS contains, which would
// exclude the envelope boundary.
func (e *Geom_Envelope) ContainsXY(x, y float64) bool {
	return e.CoversXY(x, y)
}

// ContainsProperly tests if an envelope is properly contained in this one. The
// envelope is properly contained if it is contained by this one but not equal
// to it.
func (e *Geom_Envelope) ContainsProperly(other *Geom_Envelope) bool {
	if e.Equals(other) {
		return false
	}
	return e.CoversEnvelope(other)
}

// CoversXY tests if the given point lies in or on the envelope.
func (e *Geom_Envelope) CoversXY(x, y float64) bool {
	if e.IsNull() {
		return false
	}
	return x >= e.minx &&
		x <= e.maxx &&
		y >= e.miny &&
		y <= e.maxy
}

// CoversCoordinate tests if the given point lies in or on the envelope.
func (e *Geom_Envelope) CoversCoordinate(p *Geom_Coordinate) bool {
	return e.CoversXY(p.X, p.Y)
}

// CoversEnvelope tests if the Geom_Envelope other lies wholly inside this Geom_Envelope
// (inclusive of the boundary).
func (e *Geom_Envelope) CoversEnvelope(other *Geom_Envelope) bool {
	if e.IsNull() || other.IsNull() {
		return false
	}
	return other.GetMinX() >= e.minx &&
		other.GetMaxX() <= e.maxx &&
		other.GetMinY() >= e.miny &&
		other.GetMaxY() <= e.maxy
}

// Distance computes the distance between this and another Geom_Envelope. The
// distance between overlapping Geom_Envelopes is 0. Otherwise, the distance is the
// Euclidean distance between the closest points.
func (e *Geom_Envelope) Distance(env *Geom_Envelope) float64 {
	if e.IntersectsEnvelope(env) {
		return 0
	}

	dx := 0.0
	if e.maxx < env.minx {
		dx = env.minx - e.maxx
	} else if e.minx > env.maxx {
		dx = e.minx - env.maxx
	}

	dy := 0.0
	if e.maxy < env.miny {
		dy = env.miny - e.maxy
	} else if e.miny > env.maxy {
		dy = e.miny - env.maxy
	}

	if dx == 0.0 {
		return dy
	}
	if dy == 0.0 {
		return dx
	}
	return math.Hypot(dx, dy)
}

// Equals tests if this envelope equals another envelope.
func (e *Geom_Envelope) Equals(other *Geom_Envelope) bool {
	if e.IsNull() {
		return other.IsNull()
	}
	return e.maxx == other.GetMaxX() &&
		e.maxy == other.GetMaxY() &&
		e.minx == other.GetMinX() &&
		e.miny == other.GetMinY()
}

// String returns a string representation of this envelope.
func (e *Geom_Envelope) String() string {
	return "Env[" + geom_floatToString(e.minx) + " : " + geom_floatToString(e.maxx) +
		", " + geom_floatToString(e.miny) + " : " + geom_floatToString(e.maxy) + "]"
}

// geom_floatToString converts a float64 to a string representation.
func geom_floatToString(f float64) string {
	return fmt.Sprintf("%v", f)
}

// CompareTo compares two envelopes using lexicographic ordering. The ordering
// comparison is based on the usual numerical comparison between the sequence of
// ordinates. Null envelopes are less than all non-null envelopes.
func (e *Geom_Envelope) CompareTo(other *Geom_Envelope) int {
	if e.IsNull() {
		if other.IsNull() {
			return 0
		}
		return -1
	}
	if other.IsNull() {
		return 1
	}
	if e.minx < other.minx {
		return -1
	}
	if e.minx > other.minx {
		return 1
	}
	if e.miny < other.miny {
		return -1
	}
	if e.miny > other.miny {
		return 1
	}
	if e.maxx < other.maxx {
		return -1
	}
	if e.maxx > other.maxx {
		return 1
	}
	if e.maxy < other.maxy {
		return -1
	}
	if e.maxy > other.maxy {
		return 1
	}
	return 0
}
