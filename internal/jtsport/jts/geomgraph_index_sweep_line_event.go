package jts

import (
	"reflect"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const (
	geomgraphIndex_SweepLineEvent_INSERT = 1
	geomgraphIndex_SweepLineEvent_DELETE = 2
)

// GeomgraphIndex_SweepLineEvent represents an event in a sweep line algorithm.
type GeomgraphIndex_SweepLineEvent struct {
	child java.Polymorphic

	label            any // used for red-blue intersection detection
	xValue           float64
	eventType        int
	insertEvent      *GeomgraphIndex_SweepLineEvent // nil if this is an INSERT event
	deleteEventIndex int
	obj              any
}

// GetChild returns the immediate child in the type hierarchy chain.
func (e *GeomgraphIndex_SweepLineEvent) GetChild() java.Polymorphic {
	return e.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (e *GeomgraphIndex_SweepLineEvent) GetParent() java.Polymorphic {
	return nil
}

// GeomgraphIndex_NewSweepLineEventInsert creates an INSERT event.
// label is the edge set label for this object.
// x is the event location.
// obj is the object being inserted.
func GeomgraphIndex_NewSweepLineEventInsert(label any, x float64, obj any) *GeomgraphIndex_SweepLineEvent {
	return &GeomgraphIndex_SweepLineEvent{
		eventType: geomgraphIndex_SweepLineEvent_INSERT,
		label:     label,
		xValue:    x,
		obj:       obj,
	}
}

// GeomgraphIndex_NewSweepLineEventDelete creates a DELETE event.
// x is the event location.
// insertEvent is the corresponding INSERT event.
func GeomgraphIndex_NewSweepLineEventDelete(x float64, insertEvent *GeomgraphIndex_SweepLineEvent) *GeomgraphIndex_SweepLineEvent {
	return &GeomgraphIndex_SweepLineEvent{
		eventType:   geomgraphIndex_SweepLineEvent_DELETE,
		xValue:      x,
		insertEvent: insertEvent,
	}
}

// IsInsert returns true if this is an INSERT event.
func (e *GeomgraphIndex_SweepLineEvent) IsInsert() bool {
	return e.eventType == geomgraphIndex_SweepLineEvent_INSERT
}

// IsDelete returns true if this is a DELETE event.
func (e *GeomgraphIndex_SweepLineEvent) IsDelete() bool {
	return e.eventType == geomgraphIndex_SweepLineEvent_DELETE
}

// GetInsertEvent returns the corresponding INSERT event for a DELETE event.
func (e *GeomgraphIndex_SweepLineEvent) GetInsertEvent() *GeomgraphIndex_SweepLineEvent {
	return e.insertEvent
}

// GetDeleteEventIndex returns the index of the corresponding DELETE event.
func (e *GeomgraphIndex_SweepLineEvent) GetDeleteEventIndex() int {
	return e.deleteEventIndex
}

// SetDeleteEventIndex sets the index of the corresponding DELETE event.
func (e *GeomgraphIndex_SweepLineEvent) SetDeleteEventIndex(deleteEventIndex int) {
	e.deleteEventIndex = deleteEventIndex
}

// GetObject returns the object associated with this event.
func (e *GeomgraphIndex_SweepLineEvent) GetObject() any {
	return e.obj
}

// IsSameLabel returns true if this event has the same label as the given event.
// No label set indicates single group.
func (e *GeomgraphIndex_SweepLineEvent) IsSameLabel(ev *GeomgraphIndex_SweepLineEvent) bool {
	if e.label == nil {
		return false
	}
	// In Java, == compares object identity. For Go, we need to handle both
	// pointer types (which work with ==) and slice types (which don't).
	// Use reflect to safely compare identity for slices.
	v1 := reflect.ValueOf(e.label)
	v2 := reflect.ValueOf(ev.label)
	if v1.Kind() == reflect.Slice && v2.Kind() == reflect.Slice {
		// Compare slice header pointers for identity.
		return v1.UnsafePointer() == v2.UnsafePointer()
	}
	return e.label == ev.label
}

// CompareTo compares two events.
// Events are ordered first by their x-value, and then by their eventType.
// Insert events are sorted before Delete events, so that items whose Insert
// and Delete events occur at the same x-value will be correctly handled.
func (e *GeomgraphIndex_SweepLineEvent) CompareTo(o *GeomgraphIndex_SweepLineEvent) int {
	if e.xValue < o.xValue {
		return -1
	}
	if e.xValue > o.xValue {
		return 1
	}
	if e.eventType < o.eventType {
		return -1
	}
	if e.eventType > o.eventType {
		return 1
	}
	return 0
}
