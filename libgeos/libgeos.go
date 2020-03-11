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
	globalMutex.Lock() // caller must unlock if this function is successful

	if globalHandle != nil {
		return globalHandle, nil
	}

	var err error
	globalHandle, err = NewHandle()
	if err != nil {
		globalMutex.Unlock()
		return nil, err
	}
	return globalHandle, nil
}

func Equals(g1, g2 geom.Geometry) (bool, error) {
	h, err := getGlobalHandle()
	if err != nil {
		return false, err
	}
	defer globalMutex.Unlock()
	return h.Equals(g1, g2)
}
