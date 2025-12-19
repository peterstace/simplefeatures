package jts

// String constants for DE-9IM matrix patterns for topological relationships.
// These can be used with RelateNG.Evaluate and RelateNG.Relate.
//
// DE-9IM Pattern Matching:
// Matrix patterns are specified as a 9-character string containing the pattern
// symbols for the DE-9IM 3x3 matrix entries, listed row-wise.
// The pattern symbols are:
//   - '0' - topological interaction has dimension 0
//   - '1' - topological interaction has dimension 1
//   - '2' - topological interaction has dimension 2
//   - 'F' - no topological interaction
//   - 'T' - topological interaction of any dimension
//   - '*' - any topological interaction is allowed, including none

// OperationRelateng_IntersectionMatrixPattern_ADJACENT is a DE-9IM pattern to
// detect whether two polygonal geometries are adjacent along an edge, but do
// not overlap.
const OperationRelateng_IntersectionMatrixPattern_ADJACENT = "F***1****"

// OperationRelateng_IntersectionMatrixPattern_CONTAINS_PROPERLY is a DE-9IM
// pattern to detect a geometry which properly contains another geometry
// (i.e. which lies entirely in the interior of the first geometry).
const OperationRelateng_IntersectionMatrixPattern_CONTAINS_PROPERLY = "T**FF*FF*"

// OperationRelateng_IntersectionMatrixPattern_INTERIOR_INTERSECTS is a DE-9IM
// pattern to detect if two geometries intersect in their interiors. This can
// be used to determine if a polygonal coverage contains any overlaps (although
// not whether they are correctly noded).
const OperationRelateng_IntersectionMatrixPattern_INTERIOR_INTERSECTS = "T********"
