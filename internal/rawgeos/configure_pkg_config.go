//go:build !sfnopkgconfig

package rawgeos

/*
#cgo pkg-config: geos
#cgo CFLAGS: -Wall
*/
import "C"
