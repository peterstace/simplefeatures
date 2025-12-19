package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

const nodingSnapround_HotPixel_TOLERANCE = 0.5

// NodingSnapround_HotPixel implements a "hot pixel" as used in the Snap
// Rounding algorithm. A hot pixel is a square region centred on the rounded
// value of the coordinate given, and of width equal to the size of the scale
// factor. It is a partially open region, which contains the interior of the
// tolerance square and the boundary minus the top and right segments. This
// ensures that every point of the space lies in a unique hot pixel. It also
// matches the rounding semantics for numbers.
//
// The hot pixel operations are all computed in the integer domain to avoid
// rounding problems.
//
// Hot Pixels support being marked as nodes. This is used to prevent
// introducing nodes at line vertices which do not have other lines snapped to
// them.
type NodingSnapround_HotPixel struct {
	originalPt  *Geom_Coordinate
	scaleFactor float64
	// The scaled ordinates of the hot pixel point.
	hpx float64
	hpy float64
	// Indicates if this hot pixel must be a node in the output.
	isNode bool
}

// NodingSnapround_NewHotPixel creates a new hot pixel centered on a rounded
// point, using a given scale factor. The scale factor must be strictly
// positive (non-zero).
func NodingSnapround_NewHotPixel(pt *Geom_Coordinate, scaleFactor float64) *NodingSnapround_HotPixel {
	if scaleFactor <= 0 {
		panic("Scale factor must be non-zero")
	}
	hp := &NodingSnapround_HotPixel{
		originalPt:  pt,
		scaleFactor: scaleFactor,
	}
	if scaleFactor != 1.0 {
		hp.hpx = hp.scaleRound(pt.GetX())
		hp.hpy = hp.scaleRound(pt.GetY())
	} else {
		hp.hpx = pt.GetX()
		hp.hpy = pt.GetY()
	}
	return hp
}

// GetCoordinate gets the coordinate this hot pixel is based at.
func (hp *NodingSnapround_HotPixel) GetCoordinate() *Geom_Coordinate {
	return hp.originalPt
}

// GetScaleFactor gets the scale factor for the precision grid for this pixel.
func (hp *NodingSnapround_HotPixel) GetScaleFactor() float64 {
	return hp.scaleFactor
}

// GetWidth gets the width of the hot pixel in the original coordinate system.
func (hp *NodingSnapround_HotPixel) GetWidth() float64 {
	return 1.0 / hp.scaleFactor
}

// IsNode tests whether this pixel has been marked as a node.
func (hp *NodingSnapround_HotPixel) IsNode() bool {
	return hp.isNode
}

// SetToNode sets this pixel to be a node.
func (hp *NodingSnapround_HotPixel) SetToNode() {
	hp.isNode = true
}

func (hp *NodingSnapround_HotPixel) scaleRound(val float64) float64 {
	return float64(java.Round(val * hp.scaleFactor))
}

// scale scales without rounding. This ensures intersections are checked
// against original linework. This is required to ensure that intersections are
// not missed because the segment is moved by snapping.
func (hp *NodingSnapround_HotPixel) scale(val float64) float64 {
	return val * hp.scaleFactor
}

// IntersectsPoint tests whether a coordinate lies in (intersects) this hot pixel.
func (hp *NodingSnapround_HotPixel) IntersectsPoint(p *Geom_Coordinate) bool {
	x := hp.scale(p.GetX())
	y := hp.scale(p.GetY())
	if x >= hp.hpx+nodingSnapround_HotPixel_TOLERANCE {
		return false
	}
	// Check Left side.
	if x < hp.hpx-nodingSnapround_HotPixel_TOLERANCE {
		return false
	}
	// Check Top side.
	if y >= hp.hpy+nodingSnapround_HotPixel_TOLERANCE {
		return false
	}
	// Check Bottom side.
	if y < hp.hpy-nodingSnapround_HotPixel_TOLERANCE {
		return false
	}
	return true
}

// IntersectsSegment tests whether the line segment (p0-p1) intersects this hot
// pixel.
func (hp *NodingSnapround_HotPixel) IntersectsSegment(p0, p1 *Geom_Coordinate) bool {
	if hp.scaleFactor == 1.0 {
		return hp.intersectsScaled(p0.GetX(), p0.GetY(), p1.GetX(), p1.GetY())
	}

	sp0x := hp.scale(p0.GetX())
	sp0y := hp.scale(p0.GetY())
	sp1x := hp.scale(p1.GetX())
	sp1y := hp.scale(p1.GetY())
	return hp.intersectsScaled(sp0x, sp0y, sp1x, sp1y)
}

func (hp *NodingSnapround_HotPixel) intersectsScaled(p0x, p0y, p1x, p1y float64) bool {
	// Determine oriented segment pointing in positive X direction.
	px := p0x
	py := p0y
	qx := p1x
	qy := p1y
	if px > qx {
		px = p1x
		py = p1y
		qx = p0x
		qy = p0y
	}

	// Report false if segment env does not intersect pixel env. This check
	// reflects the fact that the pixel Top and Right sides are open (not part
	// of the pixel).

	// Check Right side.
	maxx := hp.hpx + nodingSnapround_HotPixel_TOLERANCE
	segMinx := px
	if qx < px {
		segMinx = qx
	}
	if segMinx >= maxx {
		return false
	}
	// Check Left side.
	minx := hp.hpx - nodingSnapround_HotPixel_TOLERANCE
	segMaxx := px
	if qx > px {
		segMaxx = qx
	}
	if segMaxx < minx {
		return false
	}
	// Check Top side.
	maxy := hp.hpy + nodingSnapround_HotPixel_TOLERANCE
	segMiny := py
	if qy < py {
		segMiny = qy
	}
	if segMiny >= maxy {
		return false
	}
	// Check Bottom side.
	miny := hp.hpy - nodingSnapround_HotPixel_TOLERANCE
	segMaxy := py
	if qy > py {
		segMaxy = qy
	}
	if segMaxy < miny {
		return false
	}

	// Vertical or horizontal segments must now intersect the segment interior
	// or Left or Bottom sides.

	// Check vertical segment.
	if px == qx {
		return true
	}
	// Check horizontal segment.
	if py == qy {
		return true
	}

	// Now know segment is not horizontal or vertical.
	//
	// Compute orientation WRT each pixel corner. If corner orientation == 0,
	// segment intersects the corner. From the corner and whether segment is
	// heading up or down, can determine intersection or not.
	//
	// Otherwise, check whether segment crosses interior of pixel side. This is
	// the case if the orientations for each corner of the side are different.

	orientUL := Algorithm_CGAlgorithmsDD_OrientationIndexFloat64(px, py, qx, qy, minx, maxy)
	if orientUL == 0 {
		// Upward segment does not intersect pixel interior.
		if py < qy {
			return false
		}
		// Downward segment must intersect pixel interior.
		return true
	}

	orientUR := Algorithm_CGAlgorithmsDD_OrientationIndexFloat64(px, py, qx, qy, maxx, maxy)
	if orientUR == 0 {
		// Downward segment does not intersect pixel interior.
		if py > qy {
			return false
		}
		// Upward segment must intersect pixel interior.
		return true
	}
	// Check crossing Top side.
	if orientUL != orientUR {
		return true
	}

	orientLL := Algorithm_CGAlgorithmsDD_OrientationIndexFloat64(px, py, qx, qy, minx, miny)
	if orientLL == 0 {
		// Segment crossed LL corner, which is the only one in pixel interior.
		return true
	}
	// Check crossing Left side.
	if orientLL != orientUL {
		return true
	}

	orientLR := Algorithm_CGAlgorithmsDD_OrientationIndexFloat64(px, py, qx, qy, maxx, miny)
	if orientLR == 0 {
		// Upward segment does not intersect pixel interior.
		if py < qy {
			return false
		}
		// Downward segment must intersect pixel interior.
		return true
	}

	// Check crossing Bottom side.
	if orientLL != orientLR {
		return true
	}
	// Check crossing Right side.
	if orientLR != orientUR {
		return true
	}

	// Segment does not intersect pixel.
	return false
}

const (
	nodingSnapround_HotPixel_UPPER_RIGHT = 0
	nodingSnapround_HotPixel_UPPER_LEFT  = 1
	nodingSnapround_HotPixel_LOWER_LEFT  = 2
	nodingSnapround_HotPixel_LOWER_RIGHT = 3
)

// intersectsPixelClosure tests whether a segment intersects the closure of
// this hot pixel. This is NOT the test used in the standard snap-rounding
// algorithm, which uses the partially-open tolerance square instead. This
// method is provided for testing purposes only.
func (hp *NodingSnapround_HotPixel) intersectsPixelClosure(p0, p1 *Geom_Coordinate) bool {
	minx := hp.hpx - nodingSnapround_HotPixel_TOLERANCE
	maxx := hp.hpx + nodingSnapround_HotPixel_TOLERANCE
	miny := hp.hpy - nodingSnapround_HotPixel_TOLERANCE
	maxy := hp.hpy + nodingSnapround_HotPixel_TOLERANCE

	corner := make([]*Geom_Coordinate, 4)
	corner[nodingSnapround_HotPixel_UPPER_RIGHT] = Geom_NewCoordinateWithXY(maxx, maxy)
	corner[nodingSnapround_HotPixel_UPPER_LEFT] = Geom_NewCoordinateWithXY(minx, maxy)
	corner[nodingSnapround_HotPixel_LOWER_LEFT] = Geom_NewCoordinateWithXY(minx, miny)
	corner[nodingSnapround_HotPixel_LOWER_RIGHT] = Geom_NewCoordinateWithXY(maxx, miny)

	li := Algorithm_NewRobustLineIntersector()
	li.ComputeIntersection(p0, p1, corner[0], corner[1])
	if li.HasIntersection() {
		return true
	}
	li.ComputeIntersection(p0, p1, corner[1], corner[2])
	if li.HasIntersection() {
		return true
	}
	li.ComputeIntersection(p0, p1, corner[2], corner[3])
	if li.HasIntersection() {
		return true
	}
	li.ComputeIntersection(p0, p1, corner[3], corner[0])
	if li.HasIntersection() {
		return true
	}

	return false
}

// String returns a string representation of this HotPixel.
func (hp *NodingSnapround_HotPixel) String() string {
	return "HP(" + io_WKTWriter_FormatCoord(hp.originalPt) + ")"
}
