# timer_loop

easy timer loop controller.

`to do list`

* add goroutine pool

* add more add func

### deps

```
go get github.com/timer_loop
go get github.com/Sirupsen/logrus
```

### example

```
package main

import (
	"fmt"
	"sync/atomic"
	"time"

	timerLoop "github.com/timer_loop"
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
	start := time.Now()
	interval := 1000 * time.Millisecond
	for i := 0; i < num; i++ {
		timerCtl.AddFuncWithId(interval, string(i), func() {
			atomic.AddInt64(&c, 1)
		})
		fmt.Println(timerCtl.GetLength())
		// fmt.Println(timerCtl.TaskTimerEntryMap)
		time.Sleep(1 * time.Second)
	}

	for {
		fmt.Println(timerCtl.GetLength())
		if c == int64(num) {
			fmt.Println(c)
			break
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
	fmt.Println(timerCtl.GetLength())
	fmt.Println("time cost: ", time.Now().Sub(start))
}

```


