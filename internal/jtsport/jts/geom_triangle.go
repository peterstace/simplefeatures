package jts

import "math"

// Geom_Triangle represents a planar triangle, and provides methods for
// calculating various properties of triangles.
type Geom_Triangle struct {
	// P0 is a vertex of the triangle.
	P0 *Geom_Coordinate
	// P1 is a vertex of the triangle.
	P1 *Geom_Coordinate
	// P2 is a vertex of the triangle.
	P2 *Geom_Coordinate
}

// Geom_NewTriangle creates a new triangle with the given vertices.
func Geom_NewTriangle(p0, p1, p2 *Geom_Coordinate) *Geom_Triangle {
	return &Geom_Triangle{
		P0: p0,
		P1: p1,
		P2: p2,
	}
}

// Geom_Triangle_IsAcute tests whether a triangle is acute. A triangle is acute
// if all interior angles are acute. This is a strict test - right triangles
// will return false. A triangle which is not acute is either right or obtuse.
//
// Note: this implementation is not robust for angles very close to 90 degrees.
func Geom_Triangle_IsAcute(a, b, c *Geom_Coordinate) bool {
	if !Algorithm_Angle_IsAcute(a, b, c) {
		return false
	}
	if !Algorithm_Angle_IsAcute(b, c, a) {
		return false
	}
	if !Algorithm_Angle_IsAcute(c, a, b) {
		return false
	}
	return true
}

// Geom_Triangle_IsCCW tests whether a triangle is oriented counter-clockwise.
func Geom_Triangle_IsCCW(a, b, c *Geom_Coordinate) bool {
	return Algorithm_Orientation_Counterclockwise == Algorithm_Orientation_Index(a, b, c)
}

// Geom_Triangle_Intersects tests whether a triangle intersects a point.
func Geom_Triangle_Intersects(a, b, c, p *Geom_Coordinate) bool {
	exteriorIndex := Algorithm_Orientation_Counterclockwise
	if Geom_Triangle_IsCCW(a, b, c) {
		exteriorIndex = Algorithm_Orientation_Clockwise
	}
	if exteriorIndex == Algorithm_Orientation_Index(a, b, p) {
		return false
	}
	if exteriorIndex == Algorithm_Orientation_Index(b, c, p) {
		return false
	}
	if exteriorIndex == Algorithm_Orientation_Index(c, a, p) {
		return false
	}
	return true
}

// Geom_Triangle_PerpendicularBisector computes the line which is the
// perpendicular bisector of the line segment a-b.
func Geom_Triangle_PerpendicularBisector(a, b *Geom_Coordinate) *Algorithm_HCoordinate {
	// Returns the perpendicular bisector of the line segment ab.
	dx := b.X - a.X
	dy := b.Y - a.Y
	l1 := Algorithm_NewHCoordinateWithXYW(a.X+dx/2.0, a.Y+dy/2.0, 1.0)
	l2 := Algorithm_NewHCoordinateWithXYW(a.X-dy+dx/2.0, a.Y+dx+dy/2.0, 1.0)
	return Algorithm_NewHCoordinateFromHCoordinates(l1, l2)
}

// Geom_Triangle_Circumradius computes the radius of the circumcircle of a
// triangle.
//
// Formula is as per https://math.stackexchange.com/a/3610959
func Geom_Triangle_Circumradius(a, b, c *Geom_Coordinate) float64 {
	A := a.Distance(b)
	B := b.Distance(c)
	C := c.Distance(a)
	area := Geom_Triangle_Area(a, b, c)
	if area == 0.0 {
		return math.Inf(1)
	}
	return (A * B * C) / (4 * area)
}

// Geom_Triangle_Circumcentre computes the circumcentre of a triangle. The
// circumcentre is the centre of the circumcircle, the smallest circle which
// encloses the triangle. It is also the common intersection point of the
// perpendicular bisectors of the sides of the triangle, and is the only point
// which has equal distance to all three vertices of the triangle.
//
// The circumcentre does not necessarily lie within the triangle. For example,
// the circumcentre of an obtuse isosceles triangle lies outside the triangle.
//
// This method uses an algorithm due to J.R.Shewchuk which uses normalization to
// the origin to improve the accuracy of computation. (See "Lecture Notes on
// Geometric Robustness", Jonathan Richard Shewchuk, 1999).
func Geom_Triangle_Circumcentre(a, b, c *Geom_Coordinate) *Geom_Coordinate {
	cx := c.X
	cy := c.Y
	ax := a.X - cx
	ay := a.Y - cy
	bx := b.X - cx
	by := b.Y - cy

	denom := 2 * geom_triangle_det(ax, ay, bx, by)
	numx := geom_triangle_det(ay, ax*ax+ay*ay, by, bx*bx+by*by)
	numy := geom_triangle_det(ax, ax*ax+ay*ay, bx, bx*bx+by*by)

	ccx := cx - numx/denom
	ccy := cy + numy/denom

	return Geom_NewCoordinateWithXY(ccx, ccy)
}

// Geom_Triangle_CircumcentreDD computes the circumcentre of a triangle using
// DD extended-precision arithmetic to provide more accurate results than
// Geom_Triangle_Circumcentre.
//
// The circumcentre is the centre of the circumcircle, the smallest circle which
// encloses the triangle. It is also the common intersection point of the
// perpendicular bisectors of the sides of the triangle, and is the only point
// which has equal distance to all three vertices of the triangle.
//
// The circumcentre does not necessarily lie within the triangle. For example,
// the circumcentre of an obtuse isosceles triangle lies outside the triangle.
func Geom_Triangle_CircumcentreDD(a, b, c *Geom_Coordinate) *Geom_Coordinate {
	ax := Math_DD_ValueOfFloat64(a.X).Subtract(Math_DD_ValueOfFloat64(c.X))
	ay := Math_DD_ValueOfFloat64(a.Y).Subtract(Math_DD_ValueOfFloat64(c.Y))
	bx := Math_DD_ValueOfFloat64(b.X).Subtract(Math_DD_ValueOfFloat64(c.X))
	by := Math_DD_ValueOfFloat64(b.Y).Subtract(Math_DD_ValueOfFloat64(c.Y))

	denom := Math_DD_DeterminantDD(ax, ay, bx, by).MultiplyFloat64(2)
	asqr := ax.Sqr().Add(ay.Sqr())
	bsqr := bx.Sqr().Add(by.Sqr())
	numx := Math_DD_DeterminantDD(ay, asqr, by, bsqr)
	numy := Math_DD_DeterminantDD(ax, asqr, bx, bsqr)

	ccx := Math_DD_ValueOfFloat64(c.X).Subtract(numx.Divide(denom)).DoubleValue()
	ccy := Math_DD_ValueOfFloat64(c.Y).Add(numy.Divide(denom)).DoubleValue()

	return Geom_NewCoordinateWithXY(ccx, ccy)
}

// geom_triangle_det computes the determinant of a 2x2 matrix. Uses standard
// double-precision arithmetic, so is susceptible to round-off error.
func geom_triangle_det(m00, m01, m10, m11 float64) float64 {
	return m00*m11 - m01*m10
}

// Geom_Triangle_InCentre computes the incentre of a triangle. The incentre of a
// triangle is the point which is equidistant from the sides of the triangle. It
// is also the point at which the bisectors of the triangle's angles meet. It is
// the centre of the triangle's incircle, which is the unique circle that is
// tangent to each of the triangle's three sides.
//
// The incentre always lies within the triangle.
func Geom_Triangle_InCentre(a, b, c *Geom_Coordinate) *Geom_Coordinate {
	// The lengths of the sides, labelled by their opposite vertex.
	len0 := b.Distance(c)
	len1 := a.Distance(c)
	len2 := a.Distance(b)
	circum := len0 + len1 + len2

	inCentreX := (len0*a.X + len1*b.X + len2*c.X) / circum
	inCentreY := (len0*a.Y + len1*b.Y + len2*c.Y) / circum
	return Geom_NewCoordinateWithXY(inCentreX, inCentreY)
}

// Geom_Triangle_Centroid computes the centroid (centre of mass) of a triangle.
// This is also the point at which the triangle's three medians intersect (a
// triangle median is the segment from a vertex of the triangle to the midpoint
// of the opposite side). The centroid divides each median in a ratio of 2:1.
//
// The centroid always lies within the triangle.
func Geom_Triangle_Centroid(a, b, c *Geom_Coordinate) *Geom_Coordinate {
	x := (a.X + b.X + c.X) / 3
	y := (a.Y + b.Y + c.Y) / 3
	return Geom_NewCoordinateWithXY(x, y)
}

// Geom_Triangle_Length computes the length of the perimeter of a triangle.
func Geom_Triangle_Length(a, b, c *Geom_Coordinate) float64 {
	return a.Distance(b) + b.Distance(c) + c.Distance(a)
}

// Geom_Triangle_LongestSideLength computes the length of the longest side of a
// triangle.
func Geom_Triangle_LongestSideLength(a, b, c *Geom_Coordinate) float64 {
	lenAB := a.Distance(b)
	lenBC := b.Distance(c)
	lenCA := c.Distance(a)
	maxLen := lenAB
	if lenBC > maxLen {
		maxLen = lenBC
	}
	if lenCA > maxLen {
		maxLen = lenCA
	}
	return maxLen
}

// Geom_Triangle_AngleBisector computes the point at which the bisector of the
// angle ABC cuts the segment AC.
func Geom_Triangle_AngleBisector(a, b, c *Geom_Coordinate) *Geom_Coordinate {
	// Uses the fact that the lengths of the parts of the split segment are
	// proportional to the lengths of the adjacent triangle sides.
	len0 := b.Distance(a)
	len2 := b.Distance(c)
	frac := len0 / (len0 + len2)
	dx := c.X - a.X
	dy := c.Y - a.Y

	splitPt := Geom_NewCoordinateWithXY(a.X+frac*dx, a.Y+frac*dy)
	return splitPt
}

// Geom_Triangle_Area computes the 2D area of a triangle. The area value is
// always non-negative.
func Geom_Triangle_Area(a, b, c *Geom_Coordinate) float64 {
	return math.Abs(((c.X-a.X)*(b.Y-a.Y) - (b.X-a.X)*(c.Y-a.Y)) / 2)
}

// Geom_Triangle_SignedArea computes the signed 2D area of a triangle. The area
// value is positive if the triangle is oriented CW, and negative if it is
// oriented CCW.
//
// The signed area value can be used to determine point orientation, but the
// implementation in this method is susceptible to round-off errors. Use
// Algorithm_Orientation_Index for robust orientation calculation.
func Geom_Triangle_SignedArea(a, b, c *Geom_Coordinate) float64 {
	// Uses the formula 1/2 * | u x v | where u,v are the side vectors of the
	// triangle x is the vector cross-product. For 2D vectors, this formula
	// simplifies to the expression below.
	return ((c.X-a.X)*(b.Y-a.Y) - (b.X-a.X)*(c.Y-a.Y)) / 2
}

// Geom_Triangle_Area3D computes the 3D area of a triangle. The value computed
// is always non-negative.
func Geom_Triangle_Area3D(a, b, c *Geom_Coordinate) float64 {
	// Uses the formula 1/2 * | u x v | where u,v are the side vectors of the
	// triangle x is the vector cross-product.
	// Side vectors u and v.
	ux := b.X - a.X
	uy := b.Y - a.Y
	uz := b.GetZ() - a.GetZ()

	vx := c.X - a.X
	vy := c.Y - a.Y
	vz := c.GetZ() - a.GetZ()

	// Cross-product = u x v.
	crossx := uy*vz - uz*vy
	crossy := uz*vx - ux*vz
	crossz := ux*vy - uy*vx

	// Tri area = 1/2 * | u x v |.
	absSq := crossx*crossx + crossy*crossy + crossz*crossz
	area3D := math.Sqrt(absSq) / 2

	return area3D
}

// Geom_Triangle_InterpolateZ computes the Z-value (elevation) of an XY point on
// a three-dimensional plane defined by a triangle whose vertices have Z-values.
// The defining triangle must not be degenerate (in other words, the triangle
// must enclose a non-zero area), and must not be parallel to the Z-axis.
//
// This method can be used to interpolate the Z-value of a point inside a
// triangle (for example, of a TIN facet with elevations on the vertices).
func Geom_Triangle_InterpolateZ(p, v0, v1, v2 *Geom_Coordinate) float64 {
	x0 := v0.X
	y0 := v0.Y
	a := v1.X - x0
	b := v2.X - x0
	c := v1.Y - y0
	d := v2.Y - y0
	det := a*d - b*c
	dx := p.X - x0
	dy := p.Y - y0
	t := (d*dx - b*dy) / det
	u := (-c*dx + a*dy) / det
	z := v0.GetZ() + t*(v1.GetZ()-v0.GetZ()) + u*(v2.GetZ()-v0.GetZ())
	return z
}

// InCentre computes the incentre of this triangle. The incentre of a triangle
// is the point which is equidistant from the sides of the triangle. It is also
// the point at which the bisectors of the triangle's angles meet. It is the
// centre of the triangle's incircle, which is the unique circle that is tangent
// to each of the triangle's three sides.
func (t *Geom_Triangle) InCentre() *Geom_Coordinate {
	return Geom_Triangle_InCentre(t.P0, t.P1, t.P2)
}

// IsAcute tests whether this triangle is acute. A triangle is acute if all
// interior angles are acute. This is a strict test - right triangles will
// return false. A triangle which is not acute is either right or obtuse.
//
// Note: this implementation is not robust for angles very close to 90 degrees.
func (t *Geom_Triangle) IsAcute() bool {
	return Geom_Triangle_IsAcute(t.P0, t.P1, t.P2)
}

// IsCCW tests whether this triangle is oriented counter-clockwise.
func (t *Geom_Triangle) IsCCW() bool {
	return Geom_Triangle_IsCCW(t.P0, t.P1, t.P2)
}

// Circumcentre computes the circumcentre of this triangle. The circumcentre is
// the centre of the circumcircle, the smallest circle which passes through all
// the triangle vertices. It is also the common intersection point of the
// perpendicular bisectors of the sides of the triangle, and is the only point
// which has equal distance to all three vertices of the triangle.
//
// The circumcentre does not necessarily lie within the triangle.
//
// This method uses an algorithm due to J.R.Shewchuk which uses normalization to
// the origin to improve the accuracy of computation. (See "Lecture Notes on
// Geometric Robustness", Jonathan Richard Shewchuk, 1999).
func (t *Geom_Triangle) Circumcentre() *Geom_Coordinate {
	return Geom_Triangle_Circumcentre(t.P0, t.P1, t.P2)
}

// Circumradius computes the radius of the circumcircle of this triangle.
func (t *Geom_Triangle) Circumradius() float64 {
	return Geom_Triangle_Circumradius(t.P0, t.P1, t.P2)
}

// Centroid computes the centroid (centre of mass) of this triangle. This is
// also the point at which the triangle's three medians intersect (a triangle
// median is the segment from a vertex of the triangle to the midpoint of the
// opposite side). The centroid divides each median in a ratio of 2:1.
//
// The centroid always lies within the triangle.
func (t *Geom_Triangle) Centroid() *Geom_Coordinate {
	return Geom_Triangle_Centroid(t.P0, t.P1, t.P2)
}

// Length computes the length of the perimeter of this triangle.
func (t *Geom_Triangle) Length() float64 {
	return Geom_Triangle_Length(t.P0, t.P1, t.P2)
}

// LongestSideLength computes the length of the longest side of this triangle.
func (t *Geom_Triangle) LongestSideLength() float64 {
	return Geom_Triangle_LongestSideLength(t.P0, t.P1, t.P2)
}

// Area computes the 2D area of this triangle. The area value is always
// non-negative.
func (t *Geom_Triangle) Area() float64 {
	return Geom_Triangle_Area(t.P0, t.P1, t.P2)
}

// SignedArea computes the signed 2D area of this triangle. The area value is
// positive if the triangle is oriented CW, and negative if it is oriented CCW.
//
// The signed area value can be used to determine point orientation, but the
// implementation in this method is susceptible to round-off errors. Use
// Algorithm_Orientation_Index for robust orientation calculation.
func (t *Geom_Triangle) SignedArea() float64 {
	return Geom_Triangle_SignedArea(t.P0, t.P1, t.P2)
}

// Area3D computes the 3D area of this triangle. The value computed is always
// non-negative.
func (t *Geom_Triangle) Area3D() float64 {
	return Geom_Triangle_Area3D(t.P0, t.P1, t.P2)
}

// InterpolateZ computes the Z-value (elevation) of an XY point on a
// three-dimensional plane defined by this triangle (whose vertices must have
// Z-values). This triangle must not be degenerate (in other words, the triangle
// must enclose a non-zero area), and must not be parallel to the Z-axis.
//
// This method can be used to interpolate the Z-value of a point inside this
// triangle (for example, of a TIN facet with elevations on the vertices).
//
// Panics if p is nil.
func (t *Geom_Triangle) InterpolateZ(p *Geom_Coordinate) float64 {
	if p == nil {
		panic("Supplied point is null.")
	}
	return Geom_Triangle_InterpolateZ(p, t.P0, t.P1, t.P2)
}
