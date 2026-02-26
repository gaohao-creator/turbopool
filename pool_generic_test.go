package turbopool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPoolWithGeneric(t *testing.T) {
	pool, _ := NewPoolDefaultHandler(5, WithExpiryDuration(10*time.Second))
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
}

func TestPoolDefaultWorkers_int(t *testing.T) {
	var process func(int)
	process = func(i int) {
		fmt.Println(i)
	}
	var wg sync.WaitGroup
	pool, _ := NewPoolDefaultWorkers(5, func(task int) {
		defer wg.Done()
		process(task)
	}, WithExpiryDuration(10*time.Second))
	defer pool.Release()

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
