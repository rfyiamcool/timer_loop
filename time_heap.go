package main

import (
	"sync"
	"time"

	"container/heap"
	"runtime/debug"

	"github.com/Sirupsen/logrus"
)

const (
	MIN_TIMER = 1 * time.Millisecond // 最少的时间戳, 小于 1ms 没意义, cpu hz.
)

var (
	nextAddSeq uint = 1

	logger *logrus.Logger
)

// time heap
type TimerHeapHandler struct {
	timers            []*TimerEntry
	locker            *sync.Mutex
	TaskTimerEntryMap map[string]*TimerEntry
}

func New() *TimerHeapHandler {
	var ttimers TimerHeapHandler
	heap.Init(&ttimers) // heap interface method

	return &TimerHeapHandler{
		locker:            new(sync.Mutex),
		TaskTimerEntryMap: make(map[string]*TimerEntry),
	}
}

func (h *TimerHeapHandler) Len() int {
	return len(h.timers)
}

func (h *TimerHeapHandler) Less(i, j int) bool {
	t1, t2 := h.timers[i].runTime, h.timers[j].runTime
	if t1.Before(t2) {
		return true
	}

	if t1.After(t2) {
		return false
	}

	return h.timers[i].addSeq < h.timers[j].addSeq
}

func (h *TimerHeapHandler) Swap(i, j int) {
	var tmp *TimerEntry
	tmp = h.timers[i]
	h.timers[i] = h.timers[j]
	h.timers[j] = tmp
}

func (h *TimerHeapHandler) Push(x interface{}) {
	h.timers = append(h.timers, x.(*TimerEntry))
}

func (h *TimerHeapHandler) Pop() (ret interface{}) {
	l := len(h.timers)
	h.timers, ret = h.timers[:l-1], h.timers[l-1]
	return
}

func (h *TimerHeapHandler) GetLength() map[string]int {
	return map[string]int{
		"heap_len": h.Len(),
		"id_map":   len(h.TaskTimerEntryMap),
	}
}

func (h *TimerHeapHandler) CancelById(id string) {
	h.locker.Lock()
	defer h.locker.Unlock()

	if entry, ok := h.TaskTimerEntryMap[id]; ok {
		entry.Cancel()
	}
}

func (h *TimerHeapHandler) AddFuncWithId(d time.Duration, taskId string, callback CallbackFunc) *TimerEntry {
	return h.addCallback(d,
		callback,
		map[string]interface{}{
			"TaskId": taskId,
		})
}

func (h *TimerHeapHandler) AddCronFuncWithId(d time.Duration, taskId string, callback CallbackFunc) *TimerEntry {
	return h.addCallback(d,
		callback,
		map[string]interface{}{
			"TaskId": taskId,
			"Repeat": true,
		})
}

func (h *TimerHeapHandler) addCallback(d time.Duration, callback CallbackFunc, argument map[string]interface{}) *TimerEntry {
	h.locker.Lock()
	defer h.locker.Unlock()

	if d < MIN_TIMER {
		d = MIN_TIMER
	}
	var taskId string
	var repeat bool
	if v, ok := argument["TaskId"]; ok {
		taskId = v.(string)
	}

	if v, ok := argument["Repeat"]; ok {
		repeat = v.(bool)
	}

	t := &TimerEntry{
		runTime:  time.Now().Add(d),
		taskId:   taskId,
		interval: d,
		callback: callback,
		repeat:   repeat,
		active:   true,
	}
	t.addSeq = nextAddSeq // set addseq when locked
	nextAddSeq += 1

	heap.Push(h, t)

	h.TaskTimerEntryMap[taskId] = t
	return t
}

func (h *TimerHeapHandler) EventLoop() {
	now := time.Now()
	h.locker.Lock()

	for {
		if h.Len() <= 0 {
			break
		}

		nextRunTime := h.timers[0].runTime
		if nextRunTime.After(now) {
			// not due time
			break
		}

		t := heap.Pop(h).(*TimerEntry)
		if t.taskId != "" {
			delete(h.TaskTimerEntryMap, t.taskId)
		}

		callback := t.callback
		if callback == nil {
			continue
		}

		if !t.repeat {
			t.callback = nil
		}
		// 减少在callback逻辑出阻塞.
		h.locker.Unlock()

		h.runCallback(callback) // don't use goroutine

		h.locker.Lock()
		if t.repeat {
			t.runTime = t.runTime.Add(t.interval)
			if !t.runTime.After(now) {
				t.runTime = now.Add(t.interval)
			}
			t.addSeq = nextAddSeq
			nextAddSeq += 1
			heap.Push(h, t)
		}
	}
	h.locker.Unlock()
}

func (h *TimerHeapHandler) StartTimerLoop(tickInterval time.Duration) {
	go func() {
		for {
			time.Sleep(tickInterval)
			h.EventLoop()
		}
	}()
}

func (h *TimerHeapHandler) runCallback(callback CallbackFunc) {
	defer func() {
		err := recover()
		if err != nil {
			if logger != nil {
				logger.Warnf("Callback %v paniced: %v\n", callback, err)
			}
			debug.PrintStack()
		}
	}()
	if callback == nil {
		// is not func
		return
	}
	callback()
}

type CallbackFunc func()

func SetLogger(cLogger *logrus.Logger) {
	logger = cLogger
}
