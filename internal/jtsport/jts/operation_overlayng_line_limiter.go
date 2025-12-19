package jts

// OperationOverlayng_LineLimiter limits the segments in a list of segments to
// those which intersect an envelope. This creates zero or more sections of the
// input segment sequences, containing only line segments which intersect the
// limit envelope. Segments are not clipped, since that can move line segments
// enough to alter topology, and it happens in the overlay in any case. This can
// substantially reduce the number of vertices which need to be processed during
// overlay.
//
// This optimization is only applicable to Line geometries, since it does not
// maintain the closed topology of rings. Polygonal geometries are optimized
// using the RingClipper.
type OperationOverlayng_LineLimiter struct {
	limitEnv    *Geom_Envelope
	ptList      *Geom_CoordinateList
	lastOutside *Geom_Coordinate
	sections    [][]*Geom_Coordinate
}

// OperationOverlayng_NewLineLimiter creates a new limiter for a given envelope.
func OperationOverlayng_NewLineLimiter(env *Geom_Envelope) *OperationOverlayng_LineLimiter {
	return &OperationOverlayng_LineLimiter{
		limitEnv: env,
	}
}

// Limit limits a list of segments.
func (ll *OperationOverlayng_LineLimiter) Limit(pts []*Geom_Coordinate) [][]*Geom_Coordinate {
	ll.lastOutside = nil
	ll.ptList = nil
	ll.sections = make([][]*Geom_Coordinate, 0)

	for i := 0; i < len(pts); i++ {
		p := pts[i]
		if ll.limitEnv.IntersectsCoordinate(p) {
			ll.addPoint(p)
		} else {
			ll.addOutside(p)
		}
	}
	// Finish last section, if any.
	ll.finishSection()
	return ll.sections
}

func (ll *OperationOverlayng_LineLimiter) addPoint(p *Geom_Coordinate) {
	if p == nil {
		return
	}
	ll.startSection()
	ll.ptList.AddCoordinate(p, false)
}

func (ll *OperationOverlayng_LineLimiter) addOutside(p *Geom_Coordinate) {
	segIntersects := ll.isLastSegmentIntersecting(p)
	if !segIntersects {
		ll.finishSection()
	} else {
		ll.addPoint(ll.lastOutside)
		ll.addPoint(p)
	}
	ll.lastOutside = p
}

func (ll *OperationOverlayng_LineLimiter) isLastSegmentIntersecting(p *Geom_Coordinate) bool {
	if ll.lastOutside == nil {
		// Last point must have been inside.
		if ll.isSectionOpen() {
			return true
		}
		return false
	}
	return ll.limitEnv.IntersectsCoordinates(ll.lastOutside, p)
}

func (ll *OperationOverlayng_LineLimiter) isSectionOpen() bool {
	return ll.ptList != nil
}

func (ll *OperationOverlayng_LineLimiter) startSection() {
	if ll.ptList == nil {
		ll.ptList = Geom_NewCoordinateList()
	}
	if ll.lastOutside != nil {
		ll.ptList.AddCoordinate(ll.lastOutside, false)
	}
	ll.lastOutside = nil
}

func (ll *OperationOverlayng_LineLimiter) finishSection() {
	if ll.ptList == nil {
		return
	}
	// Finish off this section.
	if ll.lastOutside != nil {
		ll.ptList.AddCoordinate(ll.lastOutside, false)
		ll.lastOutside = nil
	}

	section := ll.ptList.ToCoordinateArray()
	ll.sections = append(ll.sections, section)
	ll.ptList = nil
}
