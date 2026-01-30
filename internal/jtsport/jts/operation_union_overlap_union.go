package jts

// OperationUnion_OverlapUnion unions MultiPolygons efficiently by using full
// topological union only for polygons which may overlap, and combining with the
// remaining polygons. Polygons which may overlap are those which intersect the
// common extent of the inputs. Polygons wholly outside this extent must be
// disjoint to the computed union. They can thus be simply combined with the
// union result, which is much more performant.
//
// This situation is likely to occur during cascaded polygon union, since the
// partitioning of polygons is done heuristically and thus may group disjoint
// polygons which can lie far apart. It may also occur in real world data which
// contains many disjoint polygons (e.g. polygons representing parcels on
// different street blocks).
//
// Deprecated: due to impairing performance.
type OperationUnion_OverlapUnion struct {
	geomFactory *Geom_GeometryFactory
	g0          *Geom_Geometry
	g1          *Geom_Geometry
	isUnionSafe bool
	unionFun    OperationUnion_UnionStrategy
}

// OperationUnion_OverlapUnion_Union unions a pair of geometries, using the more
// performant overlap union algorithm if possible.
func OperationUnion_OverlapUnion_Union(g0, g1 *Geom_Geometry, unionFun OperationUnion_UnionStrategy) *Geom_Geometry {
	union := OperationUnion_NewOverlapUnionWithStrategy(g0, g1, unionFun)
	return union.Union()
}

// OperationUnion_NewOverlapUnion creates a new instance for unioning the given
// geometries.
func OperationUnion_NewOverlapUnion(g0, g1 *Geom_Geometry) *OperationUnion_OverlapUnion {
	return OperationUnion_NewOverlapUnionWithStrategy(g0, g1, OperationUnion_CascadedPolygonUnion_ClassicUnion)
}

// OperationUnion_NewOverlapUnionWithStrategy creates a new instance for
// unioning the given geometries with a custom union strategy.
func OperationUnion_NewOverlapUnionWithStrategy(g0, g1 *Geom_Geometry, unionFun OperationUnion_UnionStrategy) *OperationUnion_OverlapUnion {
	return &OperationUnion_OverlapUnion{
		g0:          g0,
		g1:          g1,
		geomFactory: g0.GetFactory(),
		unionFun:    unionFun,
	}
}

// Union unions the input geometries, using the more performant overlap union
// algorithm if possible.
func (ou *OperationUnion_OverlapUnion) Union() *Geom_Geometry {
	overlapEnv := operationUnion_OverlapUnion_overlapEnvelope(ou.g0, ou.g1)

	// If no overlap, can just combine the geometries.
	if overlapEnv.IsNull() {
		g0Copy := ou.g0.Copy()
		g1Copy := ou.g1.Copy()
		return GeomUtil_GeometryCombiner_Combine2(g0Copy, g1Copy)
	}

	disjointPolys := make([]*Geom_Geometry, 0)

	g0Overlap := ou.extractByEnvelope(overlapEnv, ou.g0, &disjointPolys)
	g1Overlap := ou.extractByEnvelope(overlapEnv, ou.g1, &disjointPolys)

	unionGeom := ou.unionFull(g0Overlap, g1Overlap)

	var result *Geom_Geometry
	ou.isUnionSafe = ou.isBorderSegmentsSame(unionGeom, overlapEnv)
	if !ou.isUnionSafe {
		// Overlap union changed border segments... need to do full union.
		result = ou.unionFull(ou.g0, ou.g1)
	} else {
		result = ou.combine(unionGeom, disjointPolys)
	}
	return result
}

// IsUnionOptimized allows checking whether the optimized or full union was
// performed. Used for unit testing.
func (ou *OperationUnion_OverlapUnion) IsUnionOptimized() bool {
	return ou.isUnionSafe
}

func operationUnion_OverlapUnion_overlapEnvelope(g0, g1 *Geom_Geometry) *Geom_Envelope {
	g0Env := g0.GetEnvelopeInternal()
	g1Env := g1.GetEnvelopeInternal()
	overlapEnv := g0Env.Intersection(g1Env)
	return overlapEnv
}

func (ou *OperationUnion_OverlapUnion) combine(unionGeom *Geom_Geometry, disjointPolys []*Geom_Geometry) *Geom_Geometry {
	if len(disjointPolys) <= 0 {
		return unionGeom
	}

	disjointPolys = append(disjointPolys, unionGeom)
	result := GeomUtil_GeometryCombiner_CombineSlice(disjointPolys)
	return result
}

func (ou *OperationUnion_OverlapUnion) extractByEnvelope(env *Geom_Envelope, geom *Geom_Geometry, disjointGeoms *[]*Geom_Geometry) *Geom_Geometry {
	intersectingGeoms := make([]*Geom_Geometry, 0)
	for i := 0; i < geom.GetNumGeometries(); i++ {
		elem := geom.GetGeometryN(i)
		if elem.GetEnvelopeInternal().IntersectsEnvelope(env) {
			intersectingGeoms = append(intersectingGeoms, elem)
		} else {
			copy := elem.Copy()
			*disjointGeoms = append(*disjointGeoms, copy)
		}
	}
	return ou.geomFactory.BuildGeometry(intersectingGeoms)
}

func (ou *OperationUnion_OverlapUnion) unionFull(geom0, geom1 *Geom_Geometry) *Geom_Geometry {
	// If both are empty collections, just return a copy of one of them.
	if geom0.GetNumGeometries() == 0 && geom1.GetNumGeometries() == 0 {
		return geom0.Copy()
	}
	union := ou.unionFun.Union(geom0, geom1)
	return union
}

func (ou *OperationUnion_OverlapUnion) isBorderSegmentsSame(result *Geom_Geometry, env *Geom_Envelope) bool {
	segsBefore := ou.extractBorderSegments2(ou.g0, ou.g1, env)

	segsAfter := make([]*Geom_LineSegment, 0)
	operationUnion_OverlapUnion_extractBorderSegments(result, env, &segsAfter)

	return operationUnion_OverlapUnion_isEqual(segsBefore, segsAfter)
}

func operationUnion_OverlapUnion_isEqual(segs0, segs1 []*Geom_LineSegment) bool {
	if len(segs0) != len(segs1) {
		return false
	}

	// Build a map indexed by hash code for efficient lookup.
	segIndex := make(map[int][]*Geom_LineSegment)
	for _, seg := range segs0 {
		hash := seg.HashCode()
		segIndex[hash] = append(segIndex[hash], seg)
	}

	for _, seg := range segs1 {
		hash := seg.HashCode()
		bucket := segIndex[hash]
		found := false
		for _, s := range bucket {
			if s.Equals(seg) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (ou *OperationUnion_OverlapUnion) extractBorderSegments2(geom0, geom1 *Geom_Geometry, env *Geom_Envelope) []*Geom_LineSegment {
	segs := make([]*Geom_LineSegment, 0)
	operationUnion_OverlapUnion_extractBorderSegments(geom0, env, &segs)
	if geom1 != nil {
		operationUnion_OverlapUnion_extractBorderSegments(geom1, env, &segs)
	}
	return segs
}

func operationUnion_OverlapUnion_intersects(env *Geom_Envelope, p0, p1 *Geom_Coordinate) bool {
	return env.IntersectsCoordinate(p0) || env.IntersectsCoordinate(p1)
}

func operationUnion_OverlapUnion_containsProperlyBoth(env *Geom_Envelope, p0, p1 *Geom_Coordinate) bool {
	return operationUnion_OverlapUnion_containsProperly(env, p0) && operationUnion_OverlapUnion_containsProperly(env, p1)
}

func operationUnion_OverlapUnion_containsProperly(env *Geom_Envelope, p *Geom_Coordinate) bool {
	if env.IsNull() {
		return false
	}
	return p.GetX() > env.GetMinX() &&
		p.GetX() < env.GetMaxX() &&
		p.GetY() > env.GetMinY() &&
		p.GetY() < env.GetMaxY()
}

func operationUnion_OverlapUnion_extractBorderSegments(geom *Geom_Geometry, env *Geom_Envelope, segs *[]*Geom_LineSegment) {
	filter := operationUnion_newBorderSegmentFilter(env, segs)
	geom.ApplyCoordinateSequenceFilter(filter)
}

// operationUnion_borderSegmentFilter is a filter that extracts border segments
// from a geometry.
type operationUnion_borderSegmentFilter struct {
	env  *Geom_Envelope
	segs *[]*Geom_LineSegment
}

var _ Geom_CoordinateSequenceFilter = (*operationUnion_borderSegmentFilter)(nil)

func (f *operationUnion_borderSegmentFilter) IsGeom_CoordinateSequenceFilter() {}

func operationUnion_newBorderSegmentFilter(env *Geom_Envelope, segs *[]*Geom_LineSegment) *operationUnion_borderSegmentFilter {
	return &operationUnion_borderSegmentFilter{
		env:  env,
		segs: segs,
	}
}

func (f *operationUnion_borderSegmentFilter) Filter(seq Geom_CoordinateSequence, i int) {
	if i <= 0 {
		return
	}

	// Extract LineSegment.
	p0 := seq.GetCoordinate(i - 1)
	p1 := seq.GetCoordinate(i)
	isBorder := operationUnion_OverlapUnion_intersects(f.env, p0, p1) && !operationUnion_OverlapUnion_containsProperlyBoth(f.env, p0, p1)
	if isBorder {
		seg := Geom_NewLineSegmentFromCoordinates(p0, p1)
		*f.segs = append(*f.segs, seg)
	}
}

func (f *operationUnion_borderSegmentFilter) IsDone() bool {
	return false
}

func (f *operationUnion_borderSegmentFilter) IsGeometryChanged() bool {
	return false
}
