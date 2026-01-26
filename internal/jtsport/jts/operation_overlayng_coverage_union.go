package jts

// OperationOverlayng_CoverageUnion unions a valid coverage of polygons or lines
// in an efficient way.
//
// A polygonal coverage is a collection of Polygons which satisfy the following
// conditions:
//  1. Vector-clean - Line segments within the collection must either be
//     identical or intersect only at endpoints.
//  2. Non-overlapping - No two polygons may overlap. Equivalently, polygons
//     must be interior-disjoint.
//
// A linear coverage is a collection of LineStrings which satisfies the
// Vector-clean condition. Note that this does not require the LineStrings to be
// fully noded - i.e. they may contain coincident linework. Coincident line
// segments are dissolved by the union. Currently linear output is not merged
// (this may be added in a future release.)
//
// No checking is done to determine whether the input is a valid coverage. This
// is because coverage validation involves segment intersection detection, which
// is much more expensive than the union phase. If the input is not a valid
// coverage then in some cases this will be detected during processing and a
// TopologyException is thrown. Otherwise, the computation will produce output,
// but it will be invalid.
//
// Unioning a valid coverage implies that no new vertices are created. This
// means that a precision model does not need to be specified. The precision of
// the vertices in the output geometry is not changed.

// OperationOverlayng_CoverageUnion_Union unions a valid polygonal coverage or
// linear network.
func OperationOverlayng_CoverageUnion_Union(coverage *Geom_Geometry) *Geom_Geometry {
	var noder Noding_Noder = Noding_NewBoundaryChainNoder()

	// Linear networks require a segment-extracting noder.
	if coverage.GetDimension() < 2 {
		noder = Noding_NewSegmentExtractingNoder()
	}

	// A precision model is not needed since no noding is done.
	return OperationOverlayng_OverlayNG_UnionGeomWithNoder(coverage, nil, noder)
}
