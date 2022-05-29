package geom

import "fmt"

// operand represents either the first (A) or second (B) geometry in a binary
// operation (such as Union or Covers).
type operand int

const (
	operandA operand = 0
	operandB operand = 1
)

type label struct {
	// Set to true once inSet has a valid value.
	populated bool

	// Indicates whether the thing being labelled is in a set or not (context
	// specific to how the label is used).
	inSet bool
}

func newUnpopulatedLabels() [2]label {
	return [2]label{}
}

func newHalfPopulatedLabels(operand operand, inSet bool) [2]label {
	var labels [2]label
	labels[operand].populated = true
	labels[operand].inSet = inSet
	return labels
}

func newPopulatedLabels(inSet bool) [2]label {
	var labels [2]label
	labels[0].populated = true
	labels[1].populated = true
	labels[0].inSet = inSet
	labels[1].inSet = inSet
	return labels
}

func mergeLabels(dst *[2]label, src [2]label) {
	dst[0].populated = dst[0].populated || src[0].populated
	dst[1].populated = dst[1].populated || src[1].populated
	dst[0].inSet = dst[0].inSet || src[0].inSet
	dst[1].inSet = dst[1].inSet || src[1].inSet
}

type location struct {
	interior bool
	boundary bool
}

func mergeLocations(dst *[2]location, src [2]location) {
	dst[0].interior = dst[0].interior || src[0].interior
	dst[1].interior = dst[1].interior || src[1].interior
	dst[0].boundary = dst[0].boundary || src[0].boundary
	dst[1].boundary = dst[1].boundary || src[1].boundary
}

func newLocationsOnBoundary(operand operand) [2]location {
	var locs [2]location
	locs[operand].boundary = true
	return locs
}

func assertPresence(labels [2]label) {
	if !labels[0].populated || !labels[1].populated {
		panic(fmt.Sprintf("all presence flags in labels not set: %v", labels))
	}
}

func selectUnion(labels [2]label) bool {
	assertPresence(labels)
	return labels[0].inSet || labels[1].inSet
}

func selectIntersection(labels [2]label) bool {
	assertPresence(labels)
	return labels[0].inSet && labels[1].inSet
}

func selectDifference(labels [2]label) bool {
	assertPresence(labels)
	return labels[0].inSet && !labels[1].inSet
}

func selectSymmetricDifference(labels [2]label) bool {
	assertPresence(labels)
	return labels[0].inSet != labels[1].inSet
}
