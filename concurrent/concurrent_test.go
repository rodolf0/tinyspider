package concurrent

import (
	"fmt"
	"runtime"
	"testing"
)

// TODO: improve test: race conditions, num-goroutines assumptions
func TestWorkerPool(t *testing.T) {
	var poolsize int = 25
	var p = MakeWorkerPool(uint(poolsize))
	var concurrence int

	var startg = runtime.NumGoroutine()

	for i := 0; i < 10000; i++ {
		p.Schedule(func(...interface{}) {
			var n = runtime.NumGoroutine()
			if n > concurrence {
				concurrence = n
			}
		}, nil)
	}
	if concurrence-poolsize != startg {
		t.Fail()
	}
	fmt.Printf("Max concurrence %v\n", concurrence-startg)
}
