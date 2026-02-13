package java

import (
	"fmt"
	"reflect"
)

type Polymorphic interface {
	GetChild() Polymorphic
	GetParent() Polymorphic
}

// GetLeaf walks the child chain to find the leaf (concrete) type. This is used
// by dispatchers to find the most-derived implementation of a method.
func GetLeaf(obj Polymorphic) Polymorphic {
	for {
		child := obj.GetChild()
		if child == nil {
			return obj
		}
		obj = child
	}
}

// InstanceOf checks if obj's type hierarchy includes T. This is equivalent to
// Java's instanceof operator for polymorphic types that use the child-chain
// dispatch pattern.
//
// The function checks:
// 1. If obj itself is of type T
// 2. If any parent type (via GetParent chain) is of type T
// 3. If any child type (via GetChild chain) is of type T
//
// This correctly handles inheritance hierarchies:
//
//	// Given a LinearRing (which extends LineString which extends Geometry)
//	InstanceOf[*Geom_Geometry](ring)    // true
//	InstanceOf[*Geom_LineString](ring)  // true
//	InstanceOf[*Geom_LinearRing](ring)  // true
//	InstanceOf[*Geom_Polygon](ring)     // false
//
// Returns false if obj is nil.
func InstanceOf[T any](obj Polymorphic) bool {
	if obj == nil {
		return false
	}
	// Check for nil pointer wrapped in interface.
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return false
	}

	// Check the object itself.
	if _, ok := obj.(T); ok {
		return true
	}

	// Traverse UP the parent chain via GetParent().
	for parent := obj.GetParent(); parent != nil; parent = parent.GetParent() {
		if _, ok := parent.(T); ok {
			return true
		}
	}

	// Traverse DOWN the child chain via GetChild().
	for child := obj.GetChild(); child != nil; child = child.GetChild() {
		if _, ok := child.(T); ok {
			return true
		}
	}

	return false
}

// Cast extracts type T from obj's type hierarchy, panicking if the cast fails.
// This is equivalent to Java's cast operator: (T) obj
//
// Panics with a descriptive message if obj cannot be cast to T (equivalent to
// Java's ClassCastException).
func Cast[T Polymorphic](obj Polymorphic) T {
	var zero T
	if obj == nil {
		panic(fmt.Sprintf("cannot cast nil to %T", zero))
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		panic(fmt.Sprintf("cannot cast nil to %T", zero))
	}
	if val, ok := obj.(T); ok {
		return val
	}
	for parent := obj.GetParent(); parent != nil; parent = parent.GetParent() {
		if val, ok := parent.(T); ok {
			return val
		}
	}
	for child := obj.GetChild(); child != nil; child = child.GetChild() {
		if val, ok := child.(T); ok {
			return val
		}
	}
	panic(fmt.Sprintf("cannot cast %T to %T", GetLeaf(obj), zero))
}
