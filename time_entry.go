package timerLoop

import (
	"time"
)

type TimerEntry struct {
	addSeq   uint
	runTime  time.Time
	interval time.Duration
	callback CallbackFunc
	taskID   string
	repeat   bool
	active   bool
}

func (t *TimerEntry) Cancel() {
	t.active = false
	t.callback = nil
	t.repeat = false
}

func (t *TimerEntry) IsActive() bool {
	return t.callback != nil
}
