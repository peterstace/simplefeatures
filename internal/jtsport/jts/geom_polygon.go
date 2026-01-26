package jts

import (
	"sort"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geom_Polygon represents a polygon with linear edges, which may include holes. The
// outer boundary (shell) and inner boundaries (holes) of the polygon are
// represented by LinearRings. The boundary rings of the polygon may have any
// orientation. Polygons are closed, simple geometries by definition.
//
// The polygon model conforms to the assertions specified in the OpenGIS Simple
// Features Specification for SQL.
//
// A Polygon is topologically valid if and only if:
//   - the coordinates which define it are valid coordinates
//   - the linear rings for the shell and holes are valid (i.e. are closed and
//     do not self-intersect)
//   - holes touch the shell or another hole at at most one point (which implies
//     that the rings of the shell and holes must not cross)
//   - the interior of the polygon is connected, or equivalently no sequence of
//     touching holes makes the interior of the polygon disconnected (i.e.
//     effectively split the polygon into two pieces).
type Geom_Polygon struct {
	*Geom_Geometry
	shell *Geom_LinearRing
	holes []*Geom_LinearRing
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (p *Geom_Polygon) GetChild() java.Polymorphic {
	return p.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (p *Geom_Polygon) GetParent() java.Polymorphic {
	return p.Geom_Geometry
}

// Geom_NewPolygonWithPrecisionModelAndSRID constructs a Polygon with the given
// exterior boundary.
//
// Deprecated: Use GeometryFactory instead.
func Geom_NewPolygonWithPrecisionModelAndSRID(shell *Geom_LinearRing, precisionModel *Geom_PrecisionModel, SRID int) *Geom_Polygon {
	return Geom_NewPolygon(shell, []*Geom_LinearRing{}, Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, SRID))
}

// Geom_NewPolygonWithPrecisionModelSRIDAndHoles constructs a Polygon with the given
// exterior boundary and interior boundaries.
//
// Deprecated: Use GeometryFactory instead.
func Geom_NewPolygonWithPrecisionModelSRIDAndHoles(shell *Geom_LinearRing, holes []*Geom_LinearRing, precisionModel *Geom_PrecisionModel, SRID int) *Geom_Polygon {
	return Geom_NewPolygon(shell, holes, Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, SRID))
}

// Geom_NewPolygon constructs a Polygon with the given exterior boundary and interior
// boundaries.
//
// Parameters:
//   - shell: the outer boundary of the new Polygon, or nil or an empty
//     LinearRing if the empty geometry is to be created.
//   - holes: the inner boundaries of the new Polygon, or nil or empty
//     LinearRings if the empty geometry is to be created.
func Geom_NewPolygon(shell *Geom_LinearRing, holes []*Geom_LinearRing, factory *Geom_GeometryFactory) *Geom_Polygon {
	geom := &Geom_Geometry{factory: factory}
	if shell == nil {
		shell = factory.CreateLinearRing()
	}
	if holes == nil {
		holes = []*Geom_LinearRing{}
	}
	if geom_Polygon_hasNullElements(holes) {
		panic("holes must not contain nil elements")
	}
	if shell.IsEmpty() && geom_Polygon_hasNonEmptyElements(holes) {
		panic("shell is empty but holes are not")
	}
	poly := &Geom_Polygon{
		Geom_Geometry: geom,
		shell:        shell,
		holes:        holes,
	}
	geom.child = poly
	return poly
}

func geom_Polygon_hasNullElements(holes []*Geom_LinearRing) bool {
	for _, hole := range holes {
		if hole == nil {
			return true
		}
	}
	return false
}

func geom_Polygon_hasNonEmptyElements(holes []*Geom_LinearRing) bool {
	for _, hole := range holes {
		if !hole.IsEmpty() {
			return true
		}
	}
	return false
}

func (p *Geom_Polygon) GetCoordinate_BODY() *Geom_Coordinate {
	return p.shell.GetCoordinate()
}

func (p *Geom_Polygon) GetCoordinates_BODY() []*Geom_Coordinate {
	if p.IsEmpty() {
		return []*Geom_Coordinate{}
	}
	coordinates := make([]*Geom_Coordinate, p.GetNumPoints())
	k := -1
	shellCoordinates := p.shell.GetCoordinates()
	for x := 0; x < len(shellCoordinates); x++ {
		k++
		coordinates[k] = shellCoordinates[x]
	}
	for i := 0; i < len(p.holes); i++ {
		childCoordinates := p.holes[i].GetCoordinates()
		for j := 0; j < len(childCoordinates); j++ {
			k++
			coordinates[k] = childCoordinates[j]
		}
	}
	return coordinates
}

func (p *Geom_Polygon) GetNumPoints_BODY() int {
	numPoints := p.shell.GetNumPoints()
	for i := 0; i < len(p.holes); i++ {
		numPoints += p.holes[i].GetNumPoints()
	}
	return numPoints
}

func (p *Geom_Polygon) GetDimension_BODY() int {
	return 2
}

func (p *Geom_Polygon) GetBoundaryDimension_BODY() int {
	return 1
}

func (p *Geom_Polygon) IsEmpty_BODY() bool {
	return p.shell.IsEmpty()
}

// IsRectangle tests whether this Polygon is a rectangle.
func (p *Geom_Polygon) IsRectangle() bool {
	if p.GetNumInteriorRing() != 0 {
		return false
	}
	if p.shell == nil {
		return false
	}
	if p.shell.GetNumPoints() != 5 {
		return false
	}

	seq := p.shell.GetCoordinateSequence()

	env := p.GetEnvelopeInternal()
	for i := 0; i < 5; i++ {
		x := seq.GetX(i)
		if !(x == env.GetMinX() || x == env.GetMaxX()) {
			return false
		}
		y := seq.GetY(i)
		if !(y == env.GetMinY() || y == env.GetMaxY()) {
			return false
		}
	}

	prevX := seq.GetX(0)
	prevY := seq.GetY(0)
	for i := 1; i <= 4; i++ {
		x := seq.GetX(i)
		y := seq.GetY(i)
		xChanged := x != prevX
		yChanged := y != prevY
		if xChanged == yChanged {
			return false
		}
		prevX = x
		prevY = y
	}
	return true
}

// GetExteriorRing returns the exterior ring of this Polygon.
func (p *Geom_Polygon) GetExteriorRing() *Geom_LinearRing {
	return p.shell
}

// GetNumInteriorRing returns the number of interior rings.
func (p *Geom_Polygon) GetNumInteriorRing() int {
	return len(p.holes)
}

// GetInteriorRingN returns the N'th interior ring.
func (p *Geom_Polygon) GetInteriorRingN(n int) *Geom_LinearRing {
	return p.holes[n]
}

func (p *Geom_Polygon) GetGeometryType_BODY() string {
	return Geom_Geometry_TypeNamePolygon
}

func (p *Geom_Polygon) GetArea_BODY() float64 {
	area := 0.0
	area += Algorithm_Area_OfRingSeq(p.shell.GetCoordinateSequence())
	for i := 0; i < len(p.holes); i++ {
		area -= Algorithm_Area_OfRingSeq(p.holes[i].GetCoordinateSequence())
	}
	return area
}

func (p *Geom_Polygon) GetLength_BODY() float64 {
	length := 0.0
	length += p.shell.GetLength()
	for i := 0; i < len(p.holes); i++ {
		length += p.holes[i].GetLength()
	}
	return length
}

func (p *Geom_Polygon) GetBoundary_BODY() *Geom_Geometry {
	if p.IsEmpty() {
		return p.GetFactory().CreateMultiLineString().Geom_Geometry
	}
	rings := make([]*Geom_LinearRing, len(p.holes)+1)
	rings[0] = p.shell
	for i := 0; i < len(p.holes); i++ {
		rings[i+1] = p.holes[i]
	}
	if len(rings) <= 1 {
		return p.GetFactory().CreateLinearRingFromCoordinateSequence(rings[0].GetCoordinateSequence()).Geom_Geometry
	}
	lineStrings := make([]*Geom_LineString, len(rings))
	for i, ring := range rings {
		lineStrings[i] = ring.Geom_LineString
	}
	return p.GetFactory().CreateMultiLineStringFromLineStrings(lineStrings).Geom_Geometry
}

func (p *Geom_Polygon) ComputeEnvelopeInternal_BODY() *Geom_Envelope {
	return p.shell.GetEnvelopeInternal()
}

func (p *Geom_Polygon) EqualsExactWithTolerance_BODY(other *Geom_Geometry, tolerance float64) bool {
	if !p.IsEquivalentClass(other) {
		return false
	}
	otherPolygon := java.Cast[*Geom_Polygon](other)
	thisShell := p.shell
	otherPolygonShell := otherPolygon.shell
	if !thisShell.Geom_Geometry.EqualsExactWithTolerance(otherPolygonShell.Geom_Geometry, tolerance) {
		return false
	}
	if len(p.holes) != len(otherPolygon.holes) {
		return false
	}
	for i := 0; i < len(p.holes); i++ {
		if !p.holes[i].Geom_Geometry.EqualsExactWithTolerance(otherPolygon.holes[i].Geom_Geometry, tolerance) {
			return false
		}
	}
	return true
}

func (p *Geom_Polygon) ApplyCoordinateFilter_BODY(filter Geom_CoordinateFilter) {
	p.shell.ApplyCoordinateFilter(filter)
	for i := 0; i < len(p.holes); i++ {
		p.holes[i].ApplyCoordinateFilter(filter)
	}
}

func (p *Geom_Polygon) ApplyCoordinateSequenceFilter_BODY(filter Geom_CoordinateSequenceFilter) {
	p.shell.ApplyCoordinateSequenceFilter(filter)
	if !filter.IsDone() {
		for i := 0; i < len(p.holes); i++ {
			p.holes[i].ApplyCoordinateSequenceFilter(filter)
			if filter.IsDone() {
				break
			}
		}
	}
	if filter.IsGeometryChanged() {
		p.GeometryChanged()
	}
}

func (p *Geom_Polygon) ApplyGeometryFilter_BODY(filter Geom_GeometryFilter) {
	filter.Filter(p.Geom_Geometry)
}

func (p *Geom_Polygon) Apply_BODY(filter Geom_GeometryComponentFilter) {
	filter.Filter(p.Geom_Geometry)
	p.shell.Apply(filter)
	for i := range p.holes {
		p.holes[i].Apply(filter)
	}
}

func (p *Geom_Polygon) CopyInternal_BODY() *Geom_Geometry {
	shellCopy := java.Cast[*Geom_LinearRing](p.shell.Copy())
	holeCopies := make([]*Geom_LinearRing, len(p.holes))
	for i := 0; i < len(p.holes); i++ {
		holeCopies[i] = java.Cast[*Geom_LinearRing](p.holes[i].Copy())
	}
	return Geom_NewPolygon(shellCopy, holeCopies, p.factory).Geom_Geometry
}

func (p *Geom_Polygon) ConvexHull_BODY() *Geom_Geometry {
	return p.GetExteriorRing().ConvexHull()
}

func (p *Geom_Polygon) Normalize_BODY() {
	p.shell = p.normalized(p.shell, true)
	for i := 0; i < len(p.holes); i++ {
		p.holes[i] = p.normalized(p.holes[i], false)
	}
	sort.Slice(p.holes, func(i, j int) bool {
		return p.holes[i].CompareTo(p.holes[j].Geom_Geometry) < 0
	})
}

func (p *Geom_Polygon) normalized(ring *Geom_LinearRing, clockwise bool) *Geom_LinearRing {
	res := java.Cast[*Geom_LinearRing](ring.Copy())
	p.normalizeRing(res, clockwise)
	return res
}

func (p *Geom_Polygon) normalizeRing(ring *Geom_LinearRing, clockwise bool) {
	if ring.IsEmpty() {
		return
	}

	seq := ring.GetCoordinateSequence()
	minCoordinateIndex := Geom_CoordinateSequences_MinCoordinateIndexInRange(seq, 0, seq.Size()-2)
	Geom_CoordinateSequences_ScrollToIndexWithRing(seq, minCoordinateIndex, true)
	if Algorithm_Orientation_IsCCWSeq(seq) == clockwise {
		Geom_CoordinateSequences_Reverse(seq)
	}
}

func (p *Geom_Polygon) CompareToSameClass_BODY(o any) int {
	poly := java.Cast[*Geom_Polygon](o.(*Geom_Geometry))

	thisShell := p.shell
	otherShell := poly.shell
	shellComp := thisShell.CompareTo(otherShell.Geom_Geometry)
	if shellComp != 0 {
		return shellComp
	}

	nHole1 := p.GetNumInteriorRing()
	nHole2 := poly.GetNumInteriorRing()
	i := 0
	for i < nHole1 && i < nHole2 {
		thisHole := p.GetInteriorRingN(i)
		otherHole := poly.GetInteriorRingN(i)
		holeComp := thisHole.CompareTo(otherHole.Geom_Geometry)
		if holeComp != 0 {
			return holeComp
		}
		i++
	}
	if i < nHole1 {
		return 1
	}
	if i < nHole2 {
		return -1
	}
	return 0
}

func (p *Geom_Polygon) CompareToSameClassWithComparator_BODY(o any, comp *Geom_CoordinateSequenceComparator) int {
	poly := java.Cast[*Geom_Polygon](o.(*Geom_Geometry))

	thisShell := p.shell
	otherShell := poly.shell
	shellComp := thisShell.CompareToWithComparator(otherShell.Geom_Geometry, comp)
	if shellComp != 0 {
		return shellComp
	}

	nHole1 := p.GetNumInteriorRing()
	nHole2 := poly.GetNumInteriorRing()
	i := 0
	for i < nHole1 && i < nHole2 {
		thisHole := p.GetInteriorRingN(i)
		otherHole := poly.GetInteriorRingN(i)
		holeComp := thisHole.CompareToWithComparator(otherHole.Geom_Geometry, comp)
		if holeComp != 0 {
			return holeComp
		}
		i++
	}
	if i < nHole1 {
		return 1
	}
	if i < nHole2 {
		return -1
	}
	return 0
}

func (p *Geom_Polygon) GetTypeCode_BODY() int {
	return Geom_Geometry_TypeCodePolygon
}

func (p *Geom_Polygon) Reverse_BODY() *Geom_Geometry {
	return p.ReverseInternal().Geom_Geometry
}

func (p *Geom_Polygon) ReverseInternal() *Geom_Polygon {
	shell := java.Cast[*Geom_LinearRing](p.GetExteriorRing().Reverse())
	holes := make([]*Geom_LinearRing, p.GetNumInteriorRing())
	for i := 0; i < len(holes); i++ {
		holes[i] = java.Cast[*Geom_LinearRing](p.GetInteriorRingN(i).Reverse())
	}

	return p.GetFactory().CreatePolygonWithLinearRingAndHoles(shell, holes)
}

// Marker interface implementation.
func (p *Geom_Polygon) IsPolygonal() {}
