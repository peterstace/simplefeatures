package jts

var _ Noding_Noder = (*Noding_SegmentExtractingNoder)(nil)

// Noding_SegmentExtractingNoder is a noder which extracts all line segments as
// SegmentStrings. This enables fast overlay of geometries which are known to
// be already fully noded. In particular, it provides fast union of polygonal
// and linear coverages. Unioning a noded set of lines is an effective way to
// perform line merging and line dissolving.
//
// No precision reduction is carried out. If that is required, another noder
// must be used (such as a snap-rounding noder), or the input must be
// precision-reduced beforehand.
type Noding_SegmentExtractingNoder struct {
	segList []Noding_SegmentString
}

// IsNoding_Noder is a marker method for interface identification.
func (sen *Noding_SegmentExtractingNoder) IsNoding_Noder() {}

// Noding_NewSegmentExtractingNoder creates a new segment-extracting noder.
func Noding_NewSegmentExtractingNoder() *Noding_SegmentExtractingNoder {
	return &Noding_SegmentExtractingNoder{}
}

// ComputeNodes extracts segments from the input segment strings.
func (sen *Noding_SegmentExtractingNoder) ComputeNodes(segStrings []Noding_SegmentString) {
	sen.segList = sen.extractSegments(segStrings)
}

func (sen *Noding_SegmentExtractingNoder) extractSegments(segStrings []Noding_SegmentString) []Noding_SegmentString {
	segList := make([]Noding_SegmentString, 0)
	for _, ss := range segStrings {
		sen.extractSegmentsFrom(ss, &segList)
	}
	return segList
}

func (sen *Noding_SegmentExtractingNoder) extractSegmentsFrom(ss Noding_SegmentString, segList *[]Noding_SegmentString) {
	for i := 0; i < ss.Size()-1; i++ {
		p0 := ss.GetCoordinate(i)
		p1 := ss.GetCoordinate(i + 1)
		seg := Noding_NewBasicSegmentString([]*Geom_Coordinate{p0, p1}, ss.GetData())
		*segList = append(*segList, seg)
	}
}

// GetNodedSubstrings returns the extracted segment strings.
func (sen *Noding_SegmentExtractingNoder) GetNodedSubstrings() []Noding_SegmentString {
	return sen.segList
}
