package turbopool

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gaohao-creator/turbopool/options"
)

func TestPoolWithGeneric(t *testing.T) {
	pool, _ := NewPoolDefaultHandler(5, options.WithExpiryDuration(10*time.Second))
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

func TestPoolDefaultWorkers_func(t *testing.T) {
	pool, _ := NewPoolDefaultWorkers(5, func(task func()) {
		task()
	}, options.WithExpiryDuration(10*time.Second))
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
}

func TestPoolDefaultWorkers_int(t *testing.T) {
	pool, _ := NewPoolDefaultWorkers(5, func(task int) {
		task++
	}, options.WithExpiryDuration(10*time.Second))
	defer pool.Release()
	var wg sync.WaitGroup
	wg.Add(20)
	for j := 0; j < 20; j++ {
		i := j
		err := pool.Submit(i)
		if err != nil {
			fmt.Println(err)
		}
	}
	wg.Wait()
}
