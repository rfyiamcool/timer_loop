package timerLoop

import (
	"sync"
	"time"

	"container/heap"
	"runtime/debug"
)

const (
	defaultMinDelay = 1 * time.Millisecond
)

// null logger
type loggerType func(tmpl string, s ...interface{})

func SetLogger(logger loggerType) {
	defaultLogger = logger
}

var (
	defaultLogger = func(tmpl string, s ...interface{}) {}
)

// time heap
type TimerHeapHandler struct {
	nextAddSeq        uint
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

func (h *TimerHeapHandler) AddFuncWithID(delay time.Duration, taskID string, callback CallbackFunc) *TimerEntry {
	return h.addCallback(
		delay,
		callback,
		taskID,
		false,
	)
}

func (h *TimerHeapHandler) AddCronFuncWithID(delay time.Duration, taskID string, callback CallbackFunc) *TimerEntry {
	return h.addCallback(
		delay,
		callback,
		taskID,
		true,
	)
}

func (h *TimerHeapHandler) addCallback(delay time.Duration, callback CallbackFunc, taskID string, repeat bool) *TimerEntry {
	h.locker.Lock()
	defer h.locker.Unlock()

	if delay < defaultMinDelay {
		delay = defaultMinDelay
	}

	if taskID == "" {
		taskID = makeTaskId()
	}

	t := &TimerEntry{
		runTime:  time.Now().Add(delay),
		taskID:   taskID,
		interval: delay,
		callback: callback,
		repeat:   repeat,
		active:   true,
	}
	// set seq
	t.addSeq = h.nextAddSeq
	h.nextAddSeq += 1

	// push heap
	heap.Push(h, t)

	// link registy
	h.TaskTimerEntryMap[taskID] = t

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
		if t.taskID != "" {
			delete(h.TaskTimerEntryMap, t.taskID)
		}

		callback := t.callback
		if callback == nil {
			continue
		}

		if !t.repeat {
			t.callback = nil
		}
		h.locker.Unlock()

		h.runCallback(callback) // don't use goroutine

		h.locker.Lock()
		if t.repeat {
			t.runTime = t.runTime.Add(t.interval)
			if !t.runTime.After(now) {
				t.runTime = now.Add(t.interval)
			}
			// t.addSeq = nextAddSeq
			// nextAddSeq += 1
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
			defaultLogger("Callback %v paniced: %v\n", callback, err)
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
