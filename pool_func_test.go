package turbopool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPoolWithFunc(t *testing.T) {
	pool, _ := NewPoolWithFuncDefaultHandler(5, WithExpiryDuration(10*time.Second))
	defer pool.Release()
	var wg sync.WaitGroup
	wg.Add(20)
	for j := 0; j < 20; j++ {
		i := j
		err := pool.Submit(func() {
			fmt.Println(i)
			time.Sleep(10 * time.Millisecond)
			wg.Done()
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	wg.Wait()
	fmt.Println("done")
}

func TestPoolWithFunc_2(t *testing.T) {
	pool, _ := NewPoolWithFuncDefaultWorkers(5, func(f func()) {
		f()
	}, WithExpiryDuration(10*time.Second))
	defer pool.Release()
	var wg sync.WaitGroup
	wg.Add(20)
	for j := 0; j < 20; j++ {
		i := j
		err := pool.Submit(func() {
			fmt.Println(i)
			time.Sleep(10 * time.Millisecond)
			wg.Done()
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	wg.Wait()
	fmt.Println("done")
}
