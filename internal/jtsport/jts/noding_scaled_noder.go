package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Noding_ScaledNoder wraps a Noder and transforms its input into the integer
// domain. This is intended for use with Snap-Rounding noders, which typically
// are only intended to work in the integer domain. Offsets can be provided to
// increase the number of digits of available precision.
//
// Clients should be aware that rescaling can involve loss of precision, which
// can cause zero-length line segments to be created. These in turn can cause
// problems when used to build a planar graph. This situation should be checked
// for and collapsed segments removed if necessary.
type Noding_ScaledNoder struct {
	noder       Noding_Noder
	scaleFactor float64
	offsetX     float64
	offsetY     float64
	isScaled    bool
}

var _ Noding_Noder = (*Noding_ScaledNoder)(nil)

func (sn *Noding_ScaledNoder) IsNoding_Noder() {}

// Noding_NewScaledNoder creates a new ScaledNoder with the given noder and
// scale factor.
func Noding_NewScaledNoder(noder Noding_Noder, scaleFactor float64) *Noding_ScaledNoder {
	return Noding_NewScaledNoderWithOffsets(noder, scaleFactor, 0, 0)
}

// Noding_NewScaledNoderWithOffsets creates a new ScaledNoder with the given
// noder, scale factor, and offsets.
func Noding_NewScaledNoderWithOffsets(noder Noding_Noder, scaleFactor, offsetX, offsetY float64) *Noding_ScaledNoder {
	sn := &Noding_ScaledNoder{
		noder:       noder,
		scaleFactor: scaleFactor,
		offsetX:     offsetX,
		offsetY:     offsetY,
		// no need to scale if input precision is already integral
	}
	sn.isScaled = !sn.IsIntegerPrecision()
	return sn
}

// IsIntegerPrecision returns true if the scale factor is 1.0.
func (sn *Noding_ScaledNoder) IsIntegerPrecision() bool {
	return sn.scaleFactor == 1.0
}

// GetNodedSubstrings returns a collection of fully noded SegmentStrings.
func (sn *Noding_ScaledNoder) GetNodedSubstrings() []Noding_SegmentString {
	splitSS := sn.noder.GetNodedSubstrings()
	if sn.isScaled {
		sn.rescaleSegmentStrings(splitSS)
	}
	return splitSS
}

// ComputeNodes computes the noding for a collection of SegmentStrings.
func (sn *Noding_ScaledNoder) ComputeNodes(inputSegStrings []Noding_SegmentString) {
	intSegStrings := inputSegStrings
	if sn.isScaled {
		intSegStrings = sn.scale(inputSegStrings)
	}
	sn.noder.ComputeNodes(intSegStrings)
}

func (sn *Noding_ScaledNoder) scale(segStrings []Noding_SegmentString) []Noding_SegmentString {
	nodedSegmentStrings := make([]Noding_SegmentString, 0, len(segStrings))
	for _, ss := range segStrings {
		scaledCoords := sn.scaleCoords(ss.GetCoordinates())
		nodedSegmentStrings = append(nodedSegmentStrings, Noding_NewNodedSegmentString(scaledCoords, ss.GetData()))
	}
	return nodedSegmentStrings
}

func (sn *Noding_ScaledNoder) scaleCoords(pts []*Geom_Coordinate) []*Geom_Coordinate {
	roundPts := make([]*Geom_Coordinate, len(pts))
	for i := 0; i < len(pts); i++ {
		roundPts[i] = Geom_NewCoordinateWithXYZ(
			java.Round(float64(float64(pts[i].GetX()-sn.offsetX)*sn.scaleFactor)),
			java.Round(float64(float64(pts[i].GetY()-sn.offsetY)*sn.scaleFactor)),
			pts[i].GetZ(),
		)
	}
	roundPtsNoDup := Geom_CoordinateArrays_RemoveRepeatedPoints(roundPts)
	return roundPtsNoDup
}

// private double scale(double val) { return (double) Math.round(val * scaleFactor); }

func (sn *Noding_ScaledNoder) rescaleSegmentStrings(segStrings []Noding_SegmentString) {
	for _, ss := range segStrings {
		sn.rescale(ss.GetCoordinates())
	}
}

func (sn *Noding_ScaledNoder) rescale(pts []*Geom_Coordinate) {
	for i := 0; i < len(pts); i++ {
		pts[i].SetX(pts[i].GetX()/sn.scaleFactor + sn.offsetX)
		pts[i].SetY(pts[i].GetY()/sn.scaleFactor + sn.offsetY)
	}
	// if (pts.length == 2 && pts[0].equals2D(pts[1])) {
	//     System.out.println(pts);
	// }
}

// private double rescale(double val) { return val / scaleFactor; }
