package jts

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var Util_Debug_DEBUG_PROPERTY_NAME = "jts.debug"
var Util_Debug_DEBUG_PROPERTY_VALUE_ON = "on"
var Util_Debug_DEBUG_PROPERTY_VALUE_TRUE = "true"

var util_debug_debugOn = false

// TRANSLITERATION NOTE: In Java, the static initializer reads the system property
// "jts.debug" at class load time. In Go, we use an IIFE in a package-level var
// to read the environment variable at package initialization.
var _ = func() struct{} {
	debugValue := os.Getenv(Util_Debug_DEBUG_PROPERTY_NAME)
	if debugValue != "" {
		if strings.EqualFold(debugValue, Util_Debug_DEBUG_PROPERTY_VALUE_ON) ||
			strings.EqualFold(debugValue, Util_Debug_DEBUG_PROPERTY_VALUE_TRUE) {
			util_debug_debugOn = true
		}
	}
	return struct{}{}
}()

var util_debug_stopwatch = Util_NewStopwatch()
var util_debug_lastTimePrinted int64

// Util_Debug_Main prints the status of debugging to stdout.
func Util_Debug_Main(args []string) {
	status := "OFF"
	if util_debug_debugOn {
		status = "ON"
	}
	fmt.Println("JTS Debugging is " + status)
}

var util_debug_debug = util_newDebug()
var util_debug_fact = Geom_NewGeometryFactoryDefault()

const util_debug_DEBUG_LINE_TAG = "D! "

type util_Debug struct {
	out      io.Writer
	watchObj any
}

func Util_Debug_IsDebugging() bool {
	return util_debug_debugOn
}

func Util_Debug_ToLine(p0, p1 *Geom_Coordinate) *Geom_LineString {
	return util_debug_fact.CreateLineStringFromCoordinates([]*Geom_Coordinate{p0, p1})
}

func Util_Debug_ToLine3(p0, p1, p2 *Geom_Coordinate) *Geom_LineString {
	return util_debug_fact.CreateLineStringFromCoordinates([]*Geom_Coordinate{p0, p1, p2})
}

func Util_Debug_ToLine4(p0, p1, p2, p3 *Geom_Coordinate) *Geom_LineString {
	return util_debug_fact.CreateLineStringFromCoordinates([]*Geom_Coordinate{p0, p1, p2, p3})
}

func Util_Debug_PrintString(str string) {
	if !util_debug_debugOn {
		return
	}
	util_debug_debug.instancePrintString(str)
}

func Util_Debug_Print(obj any) {
	if !util_debug_debugOn {
		return
	}
	util_debug_debug.instancePrint(obj)
}

func Util_Debug_PrintIf(isTrue bool, obj any) {
	if !util_debug_debugOn {
		return
	}
	if !isTrue {
		return
	}
	util_debug_debug.instancePrint(obj)
}

func Util_Debug_Println(obj any) {
	if !util_debug_debugOn {
		return
	}
	util_debug_debug.instancePrint(obj)
	util_debug_debug.println()
}

func Util_Debug_ResetTime() {
	util_debug_stopwatch.Reset()
	util_debug_lastTimePrinted = util_debug_stopwatch.GetTime()
}

func Util_Debug_PrintTime(tag string) {
	if !util_debug_debugOn {
		return
	}
	time := util_debug_stopwatch.GetTime()
	elapsedTime := time - util_debug_lastTimePrinted
	util_debug_debug.instancePrint(
		util_debug_formatField(Util_Stopwatch_GetTimeStringFromMillis(time), 10) +
			" (" + util_debug_formatField(Util_Stopwatch_GetTimeStringFromMillis(elapsedTime), 10) + " ) " +
			tag)
	util_debug_debug.println()
	util_debug_lastTimePrinted = time
}

func util_debug_formatField(s string, fieldLen int) string {
	nPad := fieldLen - len(s)
	if nPad <= 0 {
		return s
	}
	padStr := util_debug_spaces(nPad) + s
	return padStr[len(padStr)-fieldLen:]
}

func util_debug_spaces(n int) string {
	ch := make([]byte, n)
	for i := 0; i < n; i++ {
		ch[i] = ' '
	}
	return string(ch)
}

func Util_Debug_Equals(c1, c2 *Geom_Coordinate, tolerance float64) bool {
	return c1.Distance(c2) <= tolerance
}

// Util_Debug_AddWatch adds an object to be watched.
// A watched object can be printed out at any time.
// Currently only supports one watched object at a time.
func Util_Debug_AddWatch(obj any) {
	util_debug_debug.instanceAddWatch(obj)
}

func Util_Debug_PrintWatch() {
	util_debug_debug.instancePrintWatch()
}

func Util_Debug_PrintIfWatch(obj any) {
	util_debug_debug.instancePrintIfWatch(obj)
}

func Util_Debug_BreakIf(cond bool) {
	if cond {
		util_debug_doBreak()
	}
}

func Util_Debug_BreakIfEqual(o1, o2 any) {
	if o1 == o2 {
		util_debug_doBreak()
	}
}

func Util_Debug_BreakIfEqualCoords(p0, p1 *Geom_Coordinate, tolerance float64) {
	if p0.Distance(p1) <= tolerance {
		util_debug_doBreak()
	}
}

func util_debug_doBreak() {
	// Put breakpoint on following statement to break here
	return
}

func Util_Debug_HasSegment(geom Geom_Geometry, p0, p1 *Geom_Coordinate) bool {
	filter := util_debug_newSegmentFindingFilter(p0, p1)
	geom.ApplyCoordinateSequenceFilter(filter)
	return filter.hasSegment()
}

type util_debug_SegmentFindingFilter struct {
	p0, p1        *Geom_Coordinate
	hasSegmentVal bool
}

func util_debug_newSegmentFindingFilter(p0, p1 *Geom_Coordinate) *util_debug_SegmentFindingFilter {
	return &util_debug_SegmentFindingFilter{
		p0:            p0,
		p1:            p1,
		hasSegmentVal: false,
	}
}

func (f *util_debug_SegmentFindingFilter) hasSegment() bool {
	return f.hasSegmentVal
}

func (f *util_debug_SegmentFindingFilter) Filter(seq Geom_CoordinateSequence, i int) {
	if i == 0 {
		return
	}
	f.hasSegmentVal = f.p0.Equals2D(seq.GetCoordinate(i-1)) &&
		f.p1.Equals2D(seq.GetCoordinate(i))
}

func (f *util_debug_SegmentFindingFilter) IsDone() bool {
	return f.hasSegmentVal
}

func (f *util_debug_SegmentFindingFilter) IsGeometryChanged() bool {
	return false
}

// Implement Geom_CoordinateSequenceFilter interface marker
func (f *util_debug_SegmentFindingFilter) IsGeom_CoordinateSequenceFilter() {}

func util_newDebug() *util_Debug {
	return &util_Debug{
		out: os.Stdout,
	}
}

func (d *util_Debug) instancePrintWatch() {
	if d.watchObj == nil {
		return
	}
	d.instancePrint(d.watchObj)
}

func (d *util_Debug) instancePrintIfWatch(obj any) {
	if obj != d.watchObj {
		return
	}
	if d.watchObj == nil {
		return
	}
	d.instancePrint(d.watchObj)
}

func (d *util_Debug) instancePrint(obj any) {
	// TRANSLITERATION NOTE: Java uses reflection to check if the object has a
	// print(PrintStream) method and calls it if present. Go doesn't have this
	// capability, so we use a type switch for known types and fall back to
	// fmt.Sprint for others.
	switch v := obj.(type) {
	case []any:
		for _, item := range v {
			d.instancePrintObject(item)
		}
	default:
		d.instancePrintObject(obj)
	}
}

func (d *util_Debug) instancePrintObject(obj any) {
	// TRANSLITERATION NOTE: In Java, this method uses reflection to find and
	// invoke a print(PrintStream) method on the object. Since Go doesn't have
	// this capability, we print using fmt.Sprint and rely on the object's
	// String() method if it implements fmt.Stringer.
	d.instancePrintString(fmt.Sprint(obj))
}

func (d *util_Debug) println() {
	fmt.Fprintln(d.out)
}

func (d *util_Debug) instanceAddWatch(obj any) {
	d.watchObj = obj
}

func (d *util_Debug) instancePrintString(str string) {
	fmt.Fprint(d.out, util_debug_DEBUG_LINE_TAG)
	fmt.Fprint(d.out, str)
}
