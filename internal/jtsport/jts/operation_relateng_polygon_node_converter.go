package jts

import "sort"

// OperationRelateng_PolygonNodeConverter converts the node sections at a
// polygon node where a shell and one or more holes touch, or two or more holes
// touch. This converts the node topological structure from the OGC
// "touching-rings" (AKA "minimal-ring") model to the equivalent "self-touch"
// (AKA "inverted/exverted ring" or "maximal ring") model. In the "self-touch"
// model the converted NodeSection corners enclose areas which all lies inside
// the polygon (i.e. they does not enclose hole edges). This allows RelateNode
// to use simple area-additive semantics for adding edges and propagating edge
// locations.
//
// The input node sections are assumed to have canonical orientation (CW shells
// and CCW holes). The arrangement of shells and holes must be topologically
// valid. Specifically, the node sections must not cross or be collinear.
//
// This supports multiple shell-shell touches (including ones containing holes),
// and hole-hole touches, This generalizes the relate algorithm to support both
// the OGC model and the self-touch model.

// OperationRelateng_PolygonNodeConverter_Convert converts a list of sections of
// valid polygon rings to have "self-touching" structure. There are the same
// number of output sections as input ones.
func OperationRelateng_PolygonNodeConverter_Convert(polySections []*OperationRelateng_NodeSection) []*OperationRelateng_NodeSection {
	// Sort by edge angle.
	comparator := &OperationRelateng_NodeSection_EdgeAngleComparator{}
	sort.Slice(polySections, func(i, j int) bool {
		return comparator.Compare(polySections[i], polySections[j]) < 0
	})

	// TODO: move uniquing up to caller.
	sections := operationRelateng_PolygonNodeConverter_extractUnique(polySections)
	if len(sections) == 1 {
		return sections
	}

	// Find shell section index.
	shellIndex := operationRelateng_PolygonNodeConverter_findShell(sections)
	if shellIndex < 0 {
		return operationRelateng_PolygonNodeConverter_convertHoles(sections)
	}
	// At least one shell is present. Handle multiple ones if present.
	var convertedSections []*OperationRelateng_NodeSection
	nextShellIndex := shellIndex
	for {
		nextShellIndex = operationRelateng_PolygonNodeConverter_convertShellAndHoles(sections, nextShellIndex, &convertedSections)
		if nextShellIndex == shellIndex {
			break
		}
	}

	return convertedSections
}

func operationRelateng_PolygonNodeConverter_convertShellAndHoles(sections []*OperationRelateng_NodeSection, shellIndex int, convertedSections *[]*OperationRelateng_NodeSection) int {
	shellSection := sections[shellIndex]
	inVertex := shellSection.GetVertex(0)
	i := operationRelateng_PolygonNodeConverter_next(sections, shellIndex)
	var holeSection *OperationRelateng_NodeSection
	for !sections[i].IsShell() {
		holeSection = sections[i]
		// Assert: holeSection.isShell() = false.
		outVertex := holeSection.GetVertex(1)
		ns := operationRelateng_PolygonNodeConverter_createSection(shellSection, inVertex, outVertex)
		*convertedSections = append(*convertedSections, ns)

		inVertex = holeSection.GetVertex(0)
		i = operationRelateng_PolygonNodeConverter_next(sections, i)
	}
	// Create final section for corner from last hole to shell.
	outVertex := shellSection.GetVertex(1)
	ns := operationRelateng_PolygonNodeConverter_createSection(shellSection, inVertex, outVertex)
	*convertedSections = append(*convertedSections, ns)
	return i
}

func operationRelateng_PolygonNodeConverter_convertHoles(sections []*OperationRelateng_NodeSection) []*OperationRelateng_NodeSection {
	var convertedSections []*OperationRelateng_NodeSection
	copySection := sections[0]
	for i := range sections {
		inext := operationRelateng_PolygonNodeConverter_next(sections, i)
		inVertex := sections[i].GetVertex(0)
		outVertex := sections[inext].GetVertex(1)
		ns := operationRelateng_PolygonNodeConverter_createSection(copySection, inVertex, outVertex)
		convertedSections = append(convertedSections, ns)
	}
	return convertedSections
}

func operationRelateng_PolygonNodeConverter_createSection(ns *OperationRelateng_NodeSection, v0, v1 *Geom_Coordinate) *OperationRelateng_NodeSection {
	return OperationRelateng_NewNodeSection(ns.IsA(),
		Geom_Dimension_A, ns.Id(), 0, ns.GetPolygonal(),
		ns.IsNodeAtVertex(),
		v0, ns.NodePt(), v1)
}

func operationRelateng_PolygonNodeConverter_extractUnique(sections []*OperationRelateng_NodeSection) []*OperationRelateng_NodeSection {
	var uniqueSections []*OperationRelateng_NodeSection
	lastUnique := sections[0]
	uniqueSections = append(uniqueSections, lastUnique)
	for _, ns := range sections {
		if lastUnique.CompareTo(ns) != 0 {
			uniqueSections = append(uniqueSections, ns)
			lastUnique = ns
		}
	}
	return uniqueSections
}

func operationRelateng_PolygonNodeConverter_next(ns []*OperationRelateng_NodeSection, i int) int {
	next := i + 1
	if next >= len(ns) {
		next = 0
	}
	return next
}

func operationRelateng_PolygonNodeConverter_findShell(polySections []*OperationRelateng_NodeSection) int {
	for i := range polySections {
		if polySections[i].IsShell() {
			return i
		}
	}
	return -1
}
