package libgeos

import (
	"sync"

	"github.com/peterstace/simplefeatures/geom"
)

var (
	globalMutex  sync.Mutex
	globalHandle *Handle
)

func getGlobalHandle() (*Handle, error) {
	if globalHandle != nil {
		return globalHandle, nil
	}

	var err error
	globalHandle, err = NewHandle()
	if err != nil {
		return nil, err
	}
	return globalHandle, nil
}

// Equals returns true if and only if the input geometries are spatially equal
// in the XY plane. It does not support GeometryCollections.
func Equals(g1, g2 geom.Geometry) (bool, error) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	h, err := getGlobalHandle()
	if err != nil {
		return false, err
	}
	return h.Equals(g1, g2)
}
