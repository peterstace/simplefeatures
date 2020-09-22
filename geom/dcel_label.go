package geom

import "fmt"

const (
	inputAInSet     uint8 = 0b0001
	inputAPopulated uint8 = 0b0010
	inputBInSet     uint8 = 0b0100
	inputBPopulated uint8 = 0b1000

	populatedMask uint8 = 0b1010
	inSetMask     uint8 = 0b0101

	inputAMask uint8 = 0b0011
	inputBMask uint8 = 0b1100

	extracted uint8 = 0b010000
)

func assertPresenceBits(label uint8) {
	if populatedMask&label != populatedMask {
		panic(fmt.Sprintf("all presence bits in label not set: %v", label))
	}
}

func selectUnion(label uint8) bool {
	assertPresenceBits(label)
	return label&inSetMask != 0
}

func selectIntersection(label uint8) bool {
	assertPresenceBits(label)
	return label&inSetMask == inSetMask
}

func selectDifference(label uint8) bool {
	assertPresenceBits(label)
	return label&inSetMask == inputAInSet
}

func selectSymmetricDifference(label uint8) bool {
	assertPresenceBits(label)
	vals := label & inSetMask
	return vals == inputAInSet || vals == inputBInSet
}
