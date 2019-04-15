package main

import (
	"fmt"
	"sync/atomic"
	"time"

	timerLoop "github.com/rfyiamcool/timer_loop"
)

var (
	timerCtl       = timerLoop.New()
	c        int64 = 0
	num            = 5000
)

func init() {
	timerCtl.StartTimerLoop(time.Millisecond)
}

func main() {
	var (
		start    = time.Now()
		interval = 1000 * time.Millisecond
	)

	go func() {
		for i := 0; i < num; i++ {
			timerCtl.AddFuncWithID(interval, fmt.Sprintf("%d", i), func() {
				atomic.AddInt64(&c, 1)
			})
		}
	}()

	for {
		fmt.Println(timerCtl.GetLength())
		if c == int64(num) {
			break
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Println("time cost: ", time.Now().Sub(start))
}
