package turbopool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	RunTimes           = 1e6
	PoolCap            = 5e4
	BenchParam         = 10
	DefaultExpiredTime = 10 * time.Second
)

var benchSink uint64

func BenchmarkDirectGoroutine_FixedTasks(b *testing.B) {
	b.ReportAllocs()
	var counter uint64
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				atomic.AddUint64(&counter, 1)
				wg.Done()
			}()
		}
		wg.Wait()
	}
	benchSink = atomic.LoadUint64(&counter)
}

func BenchmarkGoroutine(b *testing.B) {
	b.ReportAllocs()
	sleepDuration := time.Duration(BenchParam) * time.Millisecond

	var counter uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				time.Sleep(sleepDuration)
				atomic.AddUint64(&counter, 1)
				wg.Done()
			}()
		}
		wg.Wait()
	}
	benchSink = atomic.LoadUint64(&counter)
}

func BenchmarkChannel(b *testing.B) {
	b.ReportAllocs()
	sleepDuration := time.Duration(BenchParam) * time.Millisecond

	var wg sync.WaitGroup
	sema := make(chan struct{}, PoolCap)
	var counter uint64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			sema <- struct{}{}
			go func() {
				time.Sleep(sleepDuration)
				<-sema
				atomic.AddUint64(&counter, 1)
				wg.Done()
			}()
		}
		wg.Wait()
	}
	benchSink = atomic.LoadUint64(&counter)
}

func BenchmarkErrGroup(b *testing.B) {
	b.ReportAllocs()
	sleepDuration := time.Duration(BenchParam) * time.Millisecond

	var counter uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var pool errgroup.Group
		pool.SetLimit(PoolCap)

		var wg sync.WaitGroup
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			pool.Go(func() error {
				time.Sleep(sleepDuration)
				atomic.AddUint64(&counter, 1)
				wg.Done()
				return nil
			})
		}
		wg.Wait()
		if err := pool.Wait(); err != nil {
			b.Fatalf("errgroup 执行失败: %v", err)
		}
	}
	benchSink = atomic.LoadUint64(&counter)
}

func BenchmarkTurboPool_FixedTasks(b *testing.B) {
	b.ReportAllocs()
	pool, _ := NewPoolWithFuncDefaultHandler(
		PoolCap,
		WithExpiryDuration(DefaultExpiredTime),
	)
	defer pool.Release()

	var counter uint64
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			if err := pool.Submit(func() {
				atomic.AddUint64(&counter, 1)
				wg.Done()
			}); err != nil {
				b.Fatalf("提交任务失败: %v", err)
			}
		}
		wg.Wait()
	}
	benchSink = atomic.LoadUint64(&counter)
}

func BenchmarkTurboPool_Sleep(b *testing.B) {
	b.ReportAllocs()
	sleepDuration := time.Duration(BenchParam) * time.Millisecond

	pool, _ := NewPoolWithFuncDefaultHandler(
		PoolCap,
		WithExpiryDuration(DefaultExpiredTime),
	)
	defer pool.Release()

	var counter uint64
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			if err := pool.Submit(func() {
				time.Sleep(sleepDuration)
				atomic.AddUint64(&counter, 1)
				wg.Done()
			}); err != nil {
				b.Fatalf("提交任务失败: %v", err)
			}
		}
		wg.Wait()
	}
	benchSink = atomic.LoadUint64(&counter)
}

func BenchmarkDirectGoroutine_RunParallel(b *testing.B) {
	b.ReportAllocs()
	var counter uint64
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		for pb.Next() {
			wg.Add(1)
			go func() {
				atomic.AddUint64(&counter, 1)
				wg.Done()
			}()
			wg.Wait()
		}
	})
	benchSink = atomic.LoadUint64(&counter)
}

func BenchmarkTurboPool_RunParallel(b *testing.B) {
	b.ReportAllocs()
	pool, _ := NewPoolWithFuncDefaultHandler(
		PoolCap,
		WithExpiryDuration(DefaultExpiredTime),
	)
	defer pool.Release()

	var counter uint64
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		for pb.Next() {
			wg.Add(1)
			if err := pool.Submit(func() {
				atomic.AddUint64(&counter, 1)
				wg.Done()
			}); err != nil {
				panic(err)
			}
			wg.Wait()
		}
	})
	benchSink = atomic.LoadUint64(&counter)
}
