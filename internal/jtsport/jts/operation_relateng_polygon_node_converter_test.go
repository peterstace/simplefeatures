package jts_test

import (
	"sort"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestPolygonNodeConverterShells(t *testing.T) {
	checkPolygonNodeConversion(t,
		collectSections(
			sectionShell(1, 1, 5, 5, 9, 9),
			sectionShell(8, 9, 5, 5, 6, 9),
			sectionShell(4, 9, 5, 5, 2, 9)),
		collectSections(
			sectionShell(1, 1, 5, 5, 9, 9),
			sectionShell(8, 9, 5, 5, 6, 9),
			sectionShell(4, 9, 5, 5, 2, 9)))
}

func TestPolygonNodeConverterShellAndHole(t *testing.T) {
	checkPolygonNodeConversion(t,
		collectSections(
			sectionShell(1, 1, 5, 5, 9, 9),
			sectionHole(6, 0, 5, 5, 4, 0)),
		collectSections(
			sectionShell(1, 1, 5, 5, 4, 0),
			sectionShell(6, 0, 5, 5, 9, 9)))
}

func TestPolygonNodeConverterShellsAndHoles(t *testing.T) {
	checkPolygonNodeConversion(t,
		collectSections(
			sectionShell(1, 1, 5, 5, 9, 9),
			sectionHole(6, 0, 5, 5, 4, 0),
			sectionShell(8, 8, 5, 5, 1, 8),
			sectionHole(4, 8, 5, 5, 6, 8)),
		collectSections(
			sectionShell(1, 1, 5, 5, 4, 0),
			sectionShell(6, 0, 5, 5, 9, 9),
			sectionShell(4, 8, 5, 5, 1, 8),
			sectionShell(8, 8, 5, 5, 6, 8)))
}

func TestPolygonNodeConverterShellAnd2Holes(t *testing.T) {
	checkPolygonNodeConversion(t,
		collectSections(
			sectionShell(1, 1, 5, 5, 9, 9),
			sectionHole(7, 0, 5, 5, 6, 0),
			sectionHole(4, 0, 5, 5, 3, 0)),
		collectSections(
			sectionShell(1, 1, 5, 5, 3, 0),
			sectionShell(4, 0, 5, 5, 6, 0),
			sectionShell(7, 0, 5, 5, 9, 9)))
}

func TestPolygonNodeConverterHoles(t *testing.T) {
	checkPolygonNodeConversion(t,
		collectSections(
			sectionHole(7, 0, 5, 5, 6, 0),
			sectionHole(4, 0, 5, 5, 3, 0)),
		collectSections(
			sectionShell(4, 0, 5, 5, 6, 0),
			sectionShell(7, 0, 5, 5, 3, 0)))
}

func checkPolygonNodeConversion(t *testing.T, input, expected []*jts.OperationRelateng_NodeSection) {
	t.Helper()
	actual := jts.OperationRelateng_PolygonNodeConverter_Convert(input)
	isEqual := checkSectionsEqual(actual, expected)
	if !isEqual {
		t.Errorf("Sections not equal.\nExpected: %v\nActual: %v", formatSections(expected), formatSections(actual))
	}
}

func formatSections(sections []*jts.OperationRelateng_NodeSection) string {
	result := ""
	for _, ns := range sections {
		result += ns.String() + "\n"
	}
	return result
}

func checkSectionsEqual(ns1, ns2 []*jts.OperationRelateng_NodeSection) bool {
	if len(ns1) != len(ns2) {
		return false
	}
	sortSections(ns1)
	sortSections(ns2)
	for i := range ns1 {
		comp := ns1[i].CompareTo(ns2[i])
		if comp != 0 {
			return false
		}
	}
	return true
}

func sortSections(ns []*jts.OperationRelateng_NodeSection) {
	sort.Slice(ns, func(i, j int) bool {
		return jts.OperationRelateng_NodeSection_EdgeAngleComparator_Compare(ns[i], ns[j]) < 0
	})
}

func collectSections(sections ...*jts.OperationRelateng_NodeSection) []*jts.OperationRelateng_NodeSection {
	return sections
}

func sectionHole(v0x, v0y, nx, ny, v1x, v1y float64) *jts.OperationRelateng_NodeSection {
	return section(1, v0x, v0y, nx, ny, v1x, v1y)
}

func section(ringId int, v0x, v0y, nx, ny, v1x, v1y float64) *jts.OperationRelateng_NodeSection {
	return jts.OperationRelateng_NewNodeSection(true, jts.Geom_Dimension_A, 1, ringId, nil, false,
		&jts.Geom_Coordinate{X: v0x, Y: v0y},
		&jts.Geom_Coordinate{X: nx, Y: ny},
		&jts.Geom_Coordinate{X: v1x, Y: v1y})
}

func sectionShell(v0x, v0y, nx, ny, v1x, v1y float64) *jts.OperationRelateng_NodeSection {
	return section(0, v0x, v0y, nx, ny, v1x, v1y)
}
