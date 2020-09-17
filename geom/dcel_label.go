package geom

import "fmt"

const (
	// TODO: Better names?
	// Perhaps "in set" to indicate whether the part of the DCEL is in one of
	// the geometries, and "populated" to indicate whether or not the
	// aforementioned bit has been set yet.

	inputAValue   uint8 = 0b0001
	inputAPresent uint8 = 0b0010
	inputBValue   uint8 = 0b0100
	inputBPresent uint8 = 0b1000

	presenceMask uint8 = 0b1010
	valueMask    uint8 = 0b0101

	inputAMask uint8 = 0b0011
	inputBMask uint8 = 0b1100

	extracted uint8 = 0b010000
)

func assertPresenceBits(label uint8) {
	if presenceMask&label != presenceMask {
		panic(fmt.Sprintf("all presence bits in label not set: %v", label))
	}
}

func selectUnion(label uint8) bool {
	assertPresenceBits(label)
	return label&valueMask != 0
}

func selectIntersection(label uint8) bool {
	assertPresenceBits(label)
	return label&valueMask == valueMask
}

func selectDifference(label uint8) bool {
	assertPresenceBits(label)
	return label&valueMask == inputAValue
}

func selectSymmetricDifference(label uint8) bool {
	assertPresenceBits(label)
	vals := label & valueMask
	return vals == inputAValue || vals == inputBValue
}
