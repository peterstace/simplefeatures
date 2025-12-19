package jts

// IndexStrtree_BoundablePairDistanceComparator implements a comparator that is
// used to sort the BoundablePair list.
type IndexStrtree_BoundablePairDistanceComparator struct {
	// normalOrder when true puts the lowest record at the head of the queue.
	// This is the natural order. Priority queue peek will get the least element.
	normalOrder bool
}

// IndexStrtree_NewBoundablePairDistanceComparator creates a new comparator.
// When normalOrder is true, the lowest distance is at the head. When false,
// the highest distance is at the head.
func IndexStrtree_NewBoundablePairDistanceComparator(normalOrder bool) *IndexStrtree_BoundablePairDistanceComparator {
	return &IndexStrtree_BoundablePairDistanceComparator{
		normalOrder: normalOrder,
	}
}

// Compare compares two BoundablePairs by their distances.
func (c *IndexStrtree_BoundablePairDistanceComparator) Compare(p1, p2 *IndexStrtree_BoundablePair) int {
	distance1 := p1.GetDistance()
	distance2 := p2.GetDistance()
	if c.normalOrder {
		if distance1 > distance2 {
			return 1
		} else if distance1 == distance2 {
			return 0
		}
		return -1
	}
	if distance1 > distance2 {
		return -1
	} else if distance1 == distance2 {
		return 0
	}
	return 1
}
