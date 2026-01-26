package jts

// OperationOverlayng_FastOverlayFilter provides fast filtering of overlay
// operations that can be determined by simple envelope checks.
type OperationOverlayng_FastOverlayFilter struct {
	targetGeom        *Geom_Geometry
	isTargetRectangle bool
}

// OperationOverlayng_NewFastOverlayFilter creates a new FastOverlayFilter.
func OperationOverlayng_NewFastOverlayFilter(geom *Geom_Geometry) *OperationOverlayng_FastOverlayFilter {
	return &OperationOverlayng_FastOverlayFilter{
		targetGeom:        geom,
		isTargetRectangle: geom.IsRectangle(),
	}
}

// Overlay computes the overlay operation on the input geometries, if it can be
// determined that the result is either empty or equal to one of the input
// values. Otherwise nil is returned, indicating that a full overlay operation
// must be performed.
func (fof *OperationOverlayng_FastOverlayFilter) Overlay(geom *Geom_Geometry, overlayOpCode int) *Geom_Geometry {
	// For now only INTERSECTION is handled.
	if overlayOpCode != OperationOverlayng_OverlayNG_INTERSECTION {
		return nil
	}
	return fof.intersection(geom)
}

func (fof *OperationOverlayng_FastOverlayFilter) intersection(geom *Geom_Geometry) *Geom_Geometry {
	// Handle rectangle case.
	resultForRect := fof.intersectionRectangle(geom)
	if resultForRect != nil {
		return resultForRect
	}

	// Handle general case.
	if !fof.isEnvelopeIntersects(fof.targetGeom, geom) {
		return fof.createEmpty(geom)
	}

	return nil
}

func (fof *OperationOverlayng_FastOverlayFilter) createEmpty(geom *Geom_Geometry) *Geom_Geometry {
	// Empty result has dimension of non-rectangle input.
	return OperationOverlayng_OverlayUtil_CreateEmptyResult(geom.GetDimension(), geom.GetFactory())
}

func (fof *OperationOverlayng_FastOverlayFilter) intersectionRectangle(geom *Geom_Geometry) *Geom_Geometry {
	if !fof.isTargetRectangle {
		return nil
	}

	if fof.isEnvelopeCovers(fof.targetGeom, geom) {
		return geom.Copy()
	}
	if !fof.isEnvelopeIntersects(fof.targetGeom, geom) {
		return fof.createEmpty(geom)
	}
	return nil
}

func (fof *OperationOverlayng_FastOverlayFilter) isEnvelopeIntersects(a, b *Geom_Geometry) bool {
	return a.GetEnvelopeInternal().IntersectsEnvelope(b.GetEnvelopeInternal())
}

func (fof *OperationOverlayng_FastOverlayFilter) isEnvelopeCovers(a, b *Geom_Geometry) bool {
	return a.GetEnvelopeInternal().CoversEnvelope(b.GetEnvelopeInternal())
}
