package geom

import (
	"fmt"

	"github.com/peterstace/simplefeatures/de9im"
)

func Relate(a, b Geometry) (de9im.Matrix, error) {
	if a.IsEmpty() || b.IsEmpty() {
		return de9im.Matrix(0), nil
	}
	overlay, err := createOverlay(a, b)
	if err != nil {
		return 0, fmt.Errorf("internal error creating overlay: %v", err)
	}
	return overlay.extractIntersectionMatrix(), nil
}
