package jts

// IndexChain_MonotoneChainBuilder_GetChains computes a list of the MonotoneChains
// for a list of coordinates.
func IndexChain_MonotoneChainBuilder_GetChains(pts []*Geom_Coordinate) []*IndexChain_MonotoneChain {
	return IndexChain_MonotoneChainBuilder_GetChainsWithContext(pts, nil)
}

// IndexChain_MonotoneChainBuilder_GetChainsWithContext computes a list of the
// MonotoneChains for a list of coordinates, attaching a context data object to
// each.
func IndexChain_MonotoneChainBuilder_GetChainsWithContext(pts []*Geom_Coordinate, context any) []*IndexChain_MonotoneChain {
	var mcList []*IndexChain_MonotoneChain
	if len(pts) == 0 {
		return mcList
	}
	chainStart := 0
	for {
		chainEnd := indexChain_MonotoneChainBuilder_findChainEnd(pts, chainStart)
		mc := IndexChain_NewMonotoneChain(pts, chainStart, chainEnd, context)
		mcList = append(mcList, mc)
		chainStart = chainEnd
		if chainStart >= len(pts)-1 {
			break
		}
	}
	return mcList
}

// indexChain_MonotoneChainBuilder_findChainEnd finds the index of the last
// point in a monotone chain starting at a given point. Repeated points
// (0-length segments) are included in the monotone chain returned.
func indexChain_MonotoneChainBuilder_findChainEnd(pts []*Geom_Coordinate, start int) int {
	safeStart := start
	// Skip any zero-length segments at the start of the sequence (since they
	// cannot be used to establish a quadrant).
	for safeStart < len(pts)-1 && pts[safeStart].Equals2D(pts[safeStart+1]) {
		safeStart++
	}
	// Check if there are NO non-zero-length segments.
	if safeStart >= len(pts)-1 {
		return len(pts) - 1
	}
	// Determine overall quadrant for chain (which is the starting quadrant).
	chainQuad := Geom_Quadrant_QuadrantFromCoords(pts[safeStart], pts[safeStart+1])
	last := start + 1
	for last < len(pts) {
		// Skip zero-length segments, but include them in the chain.
		if !pts[last-1].Equals2D(pts[last]) {
			// Compute quadrant for next possible segment in chain.
			quad := Geom_Quadrant_QuadrantFromCoords(pts[last-1], pts[last])
			if quad != chainQuad {
				break
			}
		}
		last++
	}
	return last - 1
}
