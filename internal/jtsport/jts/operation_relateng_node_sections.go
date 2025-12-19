package jts

import "sort"

// OperationRelateng_NodeSections manages a collection of NodeSections at a
// single node point.
type OperationRelateng_NodeSections struct {
	nodePt   *Geom_Coordinate
	sections []*OperationRelateng_NodeSection
}

// OperationRelateng_NewNodeSections creates a new NodeSections for the given
// point.
func OperationRelateng_NewNodeSections(pt *Geom_Coordinate) *OperationRelateng_NodeSections {
	return &OperationRelateng_NodeSections{
		nodePt:   pt,
		sections: make([]*OperationRelateng_NodeSection, 0),
	}
}

// GetCoordinate returns the coordinate of this node.
func (ns *OperationRelateng_NodeSections) GetCoordinate() *Geom_Coordinate {
	return ns.nodePt
}

// AddNodeSection adds a NodeSection to this collection.
func (ns *OperationRelateng_NodeSections) AddNodeSection(e *OperationRelateng_NodeSection) {
	ns.sections = append(ns.sections, e)
}

// HasInteractionAB tests if there are sections from both geometries A and B.
func (ns *OperationRelateng_NodeSections) HasInteractionAB() bool {
	isA := false
	isB := false
	for _, section := range ns.sections {
		if section.IsA() {
			isA = true
		} else {
			isB = true
		}
		if isA && isB {
			return true
		}
	}
	return false
}

// GetPolygonal returns the polygonal geometry for the given input (A or B).
func (ns *OperationRelateng_NodeSections) GetPolygonal(isA bool) *Geom_Geometry {
	for _, section := range ns.sections {
		if section.IsA() == isA {
			poly := section.GetPolygonal()
			if poly != nil {
				return poly
			}
		}
	}
	return nil
}

// CreateNode creates a RelateNode from the sections at this point.
func (ns *OperationRelateng_NodeSections) CreateNode() *OperationRelateng_RelateNode {
	ns.prepareSections()

	node := OperationRelateng_NewRelateNode(ns.nodePt)
	i := 0
	for i < len(ns.sections) {
		section := ns.sections[i]
		// If there multiple polygon sections incident at node convert them to
		// maximal-ring structure.
		if section.IsArea() && operationRelateng_NodeSections_hasMultiplePolygonSections(ns.sections, i) {
			polySections := operationRelateng_NodeSections_collectPolygonSections(ns.sections, i)
			nsConvert := OperationRelateng_PolygonNodeConverter_Convert(polySections)
			node.AddEdges(nsConvert)
			i += len(polySections)
		} else {
			// The most common case is a line or a single polygon ring section.
			node.AddEdgesFromSection(section)
			i++
		}
	}
	return node
}

// prepareSections sorts the sections so that:
//   - lines are before areas
//   - edges from the same polygon are contiguous
func (ns *OperationRelateng_NodeSections) prepareSections() {
	sort.Slice(ns.sections, func(i, j int) bool {
		return ns.sections[i].CompareTo(ns.sections[j]) < 0
	})
	// TODO: remove duplicate sections.
}

func operationRelateng_NodeSections_hasMultiplePolygonSections(sections []*OperationRelateng_NodeSection, i int) bool {
	// If last section can only be one.
	if i >= len(sections)-1 {
		return false
	}
	// Check if there are at least two sections for same polygon.
	section := sections[i]
	sectionNext := sections[i+1]
	return section.IsSamePolygon(sectionNext)
}

func operationRelateng_NodeSections_collectPolygonSections(sections []*OperationRelateng_NodeSection, i int) []*OperationRelateng_NodeSection {
	var polySections []*OperationRelateng_NodeSection
	// Note ids are only unique to a geometry.
	polySection := sections[i]
	for i < len(sections) && polySection.IsSamePolygon(sections[i]) {
		polySections = append(polySections, sections[i])
		i++
	}
	return polySections
}
