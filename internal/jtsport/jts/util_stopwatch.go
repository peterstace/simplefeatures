package jts

import (
	"fmt"
	"time"
)

// Util_Stopwatch implements a timer function which can compute
// elapsed time as well as split times.
type Util_Stopwatch struct {
	startTimestamp time.Time
	totalTime      int64
	isRunning      bool
}

func Util_NewStopwatch() *Util_Stopwatch {
	sw := &Util_Stopwatch{
		totalTime: 0,
		isRunning: false,
	}
	sw.Start()
	return sw
}

func (sw *Util_Stopwatch) Start() {
	if sw.isRunning {
		return
	}
	sw.startTimestamp = time.Now()
	sw.isRunning = true
}

func (sw *Util_Stopwatch) Stop() int64 {
	if sw.isRunning {
		sw.updateTotalTime()
		sw.isRunning = false
	}
	return sw.totalTime
}

func (sw *Util_Stopwatch) Reset() {
	sw.totalTime = 0
	sw.startTimestamp = time.Now()
}

func (sw *Util_Stopwatch) Split() int64 {
	if sw.isRunning {
		sw.updateTotalTime()
	}
	return sw.totalTime
}

func (sw *Util_Stopwatch) updateTotalTime() {
	endTimestamp := time.Now()
	elapsedTime := endTimestamp.Sub(sw.startTimestamp).Milliseconds()
	sw.startTimestamp = endTimestamp
	sw.totalTime += elapsedTime
}

func (sw *Util_Stopwatch) GetTime() int64 {
	sw.updateTotalTime()
	return sw.totalTime
}

func (sw *Util_Stopwatch) GetTimeString() string {
	totalTime := sw.GetTime()
	return Util_Stopwatch_GetTimeStringFromMillis(totalTime)
}

func Util_Stopwatch_GetTimeStringFromMillis(timeMillis int64) string {
	var totalTimeStr string
	if timeMillis < 10000 {
		totalTimeStr = fmt.Sprintf("%d ms", timeMillis)
	} else {
		totalTimeStr = fmt.Sprintf("%v s", float64(timeMillis)/1000.0)
	}
	return totalTimeStr
}
