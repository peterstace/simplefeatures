package jts

import "fmt"

// OperationRelateng_NodeSection represents a computed node along with the
// incident edges on either side of it (if they exist). This captures the
// information about a node in a geometry component required to determine the
// component's contribution to the node topology. A node in an area geometry
// always has edges on both sides of the node. A node in a linear geometry may
// have one or other incident edge missing, if the node occurs at an endpoint of
// the line. The edges of an area node are assumed to be provided with CW-shell
// orientation (as per JTS norm). This must be enforced by the caller.
type OperationRelateng_NodeSection struct {
	isA            bool
	dim            int
	id             int
	ringId         int
	isNodeAtVertex bool
	nodePt         *Geom_Coordinate
	v0             *Geom_Coordinate
	v1             *Geom_Coordinate
	poly           *Geom_Geometry
}

// OperationRelateng_NewNodeSection creates a new NodeSection.
func OperationRelateng_NewNodeSection(isA bool, dimension, id, ringId int, poly *Geom_Geometry, isNodeAtVertex bool, v0, nodePt, v1 *Geom_Coordinate) *OperationRelateng_NodeSection {
	return &OperationRelateng_NodeSection{
		isA:            isA,
		dim:            dimension,
		id:             id,
		ringId:         ringId,
		poly:           poly,
		isNodeAtVertex: isNodeAtVertex,
		nodePt:         nodePt,
		v0:             v0,
		v1:             v1,
	}
}

// OperationRelateng_NodeSection_IsAreaArea tests if both sections are from area
// geometries.
func OperationRelateng_NodeSection_IsAreaArea(a, b *OperationRelateng_NodeSection) bool {
	return a.Dimension() == Geom_Dimension_A && b.Dimension() == Geom_Dimension_A
}

// OperationRelateng_NodeSection_IsProperSections tests if both sections are
// proper intersections (not at a vertex).
func OperationRelateng_NodeSection_IsProperSections(a, b *OperationRelateng_NodeSection) bool {
	return a.IsProper() && b.IsProper()
}

// GetVertex returns the vertex at the given index (0 for v0, 1 for v1).
func (ns *OperationRelateng_NodeSection) GetVertex(i int) *Geom_Coordinate {
	if i == 0 {
		return ns.v0
	}
	return ns.v1
}

// NodePt returns the node point.
func (ns *OperationRelateng_NodeSection) NodePt() *Geom_Coordinate {
	return ns.nodePt
}

// Dimension returns the dimension of the geometry.
func (ns *OperationRelateng_NodeSection) Dimension() int {
	return ns.dim
}

// Id returns the element id.
func (ns *OperationRelateng_NodeSection) Id() int {
	return ns.id
}

// RingId returns the ring id.
func (ns *OperationRelateng_NodeSection) RingId() int {
	return ns.ringId
}

// GetPolygonal gets the polygon this section is part of. Will be nil if section
// is not on a polygon boundary.
func (ns *OperationRelateng_NodeSection) GetPolygonal() *Geom_Geometry {
	return ns.poly
}

// IsShell tests if this is a shell ring (ring id 0).
func (ns *OperationRelateng_NodeSection) IsShell() bool {
	return ns.ringId == 0
}

// IsArea tests if this section is from an area geometry.
func (ns *OperationRelateng_NodeSection) IsArea() bool {
	return ns.dim == Geom_Dimension_A
}

// IsA tests if this section is from geometry A.
func (ns *OperationRelateng_NodeSection) IsA() bool {
	return ns.isA
}

// IsSameGeometry tests if this section is from the same geometry as another.
func (ns *OperationRelateng_NodeSection) IsSameGeometry(other *OperationRelateng_NodeSection) bool {
	return ns.IsA() == other.IsA()
}

// IsSamePolygon tests if this section is from the same polygon as another.
func (ns *OperationRelateng_NodeSection) IsSamePolygon(other *OperationRelateng_NodeSection) bool {
	return ns.IsA() == other.IsA() && ns.Id() == other.Id()
}

// IsNodeAtVertex tests if the node is at a vertex of the geometry.
func (ns *OperationRelateng_NodeSection) IsNodeAtVertex() bool {
	return ns.isNodeAtVertex
}

// IsProper tests if this is a proper intersection (not at a vertex).
func (ns *OperationRelateng_NodeSection) IsProper() bool {
	return !ns.isNodeAtVertex
}

// String returns a string representation of this NodeSection.
func (ns *OperationRelateng_NodeSection) String() string {
	geomName := OperationRelateng_RelateGeometry_Name(ns.isA)
	atVertexInd := "---"
	if ns.isNodeAtVertex {
		atVertexInd = "-V-"
	}
	polyId := ""
	if ns.id >= 0 {
		polyId = fmt.Sprintf("[%d:%d]", ns.id, ns.ringId)
	}
	return fmt.Sprintf("%s%d%s: %s %s %s",
		geomName, ns.dim, polyId, ns.edgeRep(ns.v0, ns.nodePt), atVertexInd, ns.edgeRep(ns.nodePt, ns.v1))
}

func (ns *OperationRelateng_NodeSection) edgeRep(p0, p1 *Geom_Coordinate) string {
	if p0 == nil || p1 == nil {
		return "null"
	}
	return Io_WKTWriter_ToLineStringFromTwoCoords(p0, p1)
}

// CompareTo compares node sections by parent geometry, dimension, element id
// and ring id, and edge vertices. Sections are assumed to be at the same node
// point.
func (ns *OperationRelateng_NodeSection) CompareTo(o *OperationRelateng_NodeSection) int {
	// Assert: nodePt.equals2D(o.nodePt())

	// Sort A before B.
	if ns.isA != o.isA {
		if ns.isA {
			return -1
		}
		return 1
	}
	// Sort on dimensions.
	if ns.dim < o.dim {
		return -1
	}
	if ns.dim > o.dim {
		return 1
	}

	// Sort on id and ring id.
	if ns.id < o.id {
		return -1
	}
	if ns.id > o.id {
		return 1
	}

	if ns.ringId < o.ringId {
		return -1
	}
	if ns.ringId > o.ringId {
		return 1
	}

	// Sort on edge coordinates.
	compV0 := operationRelateng_NodeSection_compareWithNull(ns.v0, o.v0)
	if compV0 != 0 {
		return compV0
	}

	return operationRelateng_NodeSection_compareWithNull(ns.v1, o.v1)
}

func operationRelateng_NodeSection_compareWithNull(v0, v1 *Geom_Coordinate) int {
	if v0 == nil {
		if v1 == nil {
			return 0
		}
		// Nil is lower than non-nil.
		return -1
	}
	// v0 is non-nil.
	if v1 == nil {
		return 1
	}
	return v0.CompareTo(v1)
}

// OperationRelateng_NodeSection_EdgeAngleComparator compares sections by the
// angle the entering edge makes with the positive X axis.
type OperationRelateng_NodeSection_EdgeAngleComparator struct{}

// Compare compares two NodeSections by edge angle.
func (c *OperationRelateng_NodeSection_EdgeAngleComparator) Compare(ns1, ns2 *OperationRelateng_NodeSection) int {
	return OperationRelateng_NodeSection_EdgeAngleComparator_Compare(ns1, ns2)
}

// OperationRelateng_NodeSection_EdgeAngleComparator_Compare compares two
// NodeSections by the angle the entering edge makes with the positive X axis.
func OperationRelateng_NodeSection_EdgeAngleComparator_Compare(ns1, ns2 *OperationRelateng_NodeSection) int {
	return Algorithm_PolygonNodeTopology_CompareAngle(ns1.nodePt, ns1.GetVertex(0), ns2.GetVertex(0))
}
