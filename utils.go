package timerLoop

import (
	"fmt"
	"math/rand"
	"time"
)

func makeTaskId() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Int63())
}
