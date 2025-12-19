package jts

// OperationOverlayng_EdgeMerger performs merging on the noded edges of the
// input geometries. Merging takes place on edges which are coincident (i.e.
// have the same coordinate list, modulo direction). The following situations
// can occur:
//   - Coincident edges from different input geometries have their labels combined
//   - Coincident edges from the same area geometry indicate a topology collapse.
//     In this case the topology locations are "summed" to provide a final
//     assignment of side location
//   - Coincident edges from the same linear geometry can simply be merged using
//     the same ON location
//
// The merging attempts to preserve the direction of linear edges if possible
// (which is the case if there is no other coincident edge, or if all coincident
// edges have the same direction). This ensures that the overlay output line
// direction will be as consistent as possible with input lines.
//
// The merger also preserves the order of the edges in the input. This means
// that for polygon-line overlay the result lines will be in the same order as
// in the input (possibly with multiple result lines for a single input line).
type OperationOverlayng_EdgeMerger struct{}

// OperationOverlayng_EdgeMerger_Merge merges edges with the same coordinates.
func OperationOverlayng_EdgeMerger_Merge(edges []*OperationOverlayng_Edge) []*OperationOverlayng_Edge {
	// use a list to collect the final edges, to preserve order
	mergedEdges := make([]*OperationOverlayng_Edge, 0)
	edgeMap := make(map[operationOverlayng_EdgeMerger_edgeKeyStruct]*OperationOverlayng_Edge)

	for _, edge := range edges {
		edgeKey := OperationOverlayng_EdgeKey_Create(edge)
		keyStruct := operationOverlayng_EdgeMerger_toKeyStruct(edgeKey)
		baseEdge, exists := edgeMap[keyStruct]
		if !exists {
			// this is the first (and maybe only) edge for this line
			edgeMap[keyStruct] = edge
			mergedEdges = append(mergedEdges, edge)
		} else {
			// found an existing edge
			// Assert: edges are identical (up to direction)
			// this is a fast (but incomplete) sanity check
			Util_Assert_IsTrueWithMessage(baseEdge.Size() == edge.Size(),
				"Merge of edges of different sizes - probable noding error.")

			baseEdge.Merge(edge)
		}
	}
	return mergedEdges
}

// operationOverlayng_EdgeMerger_edgeKeyStruct is a struct used as a map key
// for edge merging.
type operationOverlayng_EdgeMerger_edgeKeyStruct struct {
	p0x, p0y, p1x, p1y float64
}

func operationOverlayng_EdgeMerger_toKeyStruct(ek *OperationOverlayng_EdgeKey) operationOverlayng_EdgeMerger_edgeKeyStruct {
	return operationOverlayng_EdgeMerger_edgeKeyStruct{
		p0x: ek.p0x,
		p0y: ek.p0y,
		p1x: ek.p1x,
		p1y: ek.p1y,
	}
}
