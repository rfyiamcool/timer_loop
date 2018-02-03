package main

import (
	"time"
)

type TimerEntry struct {
	addSeq   uint
	runTime  time.Time     // 到期时间
	interval time.Duration // repeat的间隔
	callback CallbackFunc  // 回调方法
	taskId   string        // 业务ID
	repeat   bool          // 是否重复执行
	active   bool          // 任务状态
}

func (t *TimerEntry) Cancel() {
	t.active = false
	t.callback = nil
	t.repeat = false
}

func (t *TimerEntry) IsActive() bool {
	return t.callback != nil
}
