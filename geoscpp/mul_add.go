package geoscpp

/*
#cgo CFLAGS: -Wall
#include "bridge_mul_add.h"
*/
import "C"

// MulAdd computes a*b+c.
func MulAdd(a, b, c float64) float64 {
	mul := C.LIB_NewMul(C.double(a))
	defer C.LIB_DestroyMul(mul)
	aSumB := C.LIB_multiply(mul, C.double(b))

	add := C.LIB_NewAdd(aSumB)
	defer C.LIB_DestroyAdd(add)
	result := C.LIB_sum(add, C.double(c))
	return float64(result)
}

func MulAddSimple(a, b, c float64) float64 {
	return float64(C.LIB_MulAdd(
		C.double(a),
		C.double(b),
		C.double(c),
	))
}
