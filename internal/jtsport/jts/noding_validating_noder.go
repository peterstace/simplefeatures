package jts

var _ Noding_Noder = (*Noding_ValidatingNoder)(nil)

// Noding_ValidatingNoder is a wrapper for Noders which validates the output
// arrangement is correctly noded. An arrangement of line segments is fully
// noded if there is no line segment which has another segment intersecting its
// interior. If the noding is not correct, a TopologyException is thrown with
// details of the first invalid location found.
type Noding_ValidatingNoder struct {
	noder   Noding_Noder
	nodedSS []Noding_SegmentString
}

// IsNoding_Noder is a marker method for interface identification.
func (vn *Noding_ValidatingNoder) IsNoding_Noder() {}

// Noding_NewValidatingNoder creates a noding validator wrapping the given Noder.
func Noding_NewValidatingNoder(noder Noding_Noder) *Noding_ValidatingNoder {
	return &Noding_ValidatingNoder{
		noder: noder,
	}
}

// ComputeNodes checks whether the output of the wrapped noder is fully noded.
// Throws an exception if it is not.
func (vn *Noding_ValidatingNoder) ComputeNodes(segStrings []Noding_SegmentString) {
	vn.noder.ComputeNodes(segStrings)
	vn.nodedSS = vn.noder.GetNodedSubstrings()
	vn.validate()
}

func (vn *Noding_ValidatingNoder) validate() {
	// Convert to SegmentString slice for FastNodingValidator.
	segStrings := make([]Noding_SegmentString, len(vn.nodedSS))
	for i, ss := range vn.nodedSS {
		segStrings[i] = Noding_NewBasicSegmentString(ss.GetCoordinates(), ss.GetData())
	}
	nv := Noding_NewFastNodingValidator(segStrings)
	nv.CheckValid()
}

// GetNodedSubstrings returns the noded substrings.
func (vn *Noding_ValidatingNoder) GetNodedSubstrings() []Noding_SegmentString {
	return vn.nodedSS
}
