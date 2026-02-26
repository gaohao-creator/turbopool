package turbopool

import (
	"context"

	"sync/atomic"
	"time"

	ctx "github.com/gaohao-creator/turbopool/context"

	"github.com/gaohao-creator/turbopool/errors"
	"github.com/gaohao-creator/turbopool/scheduler_func"
)

type PoolWithFunc struct {
	// 池子配置选项
	options *Options
	// 任务调度器，负责管理worker和任务分发
	scheduler scheduler_func.Scheduler
	// 池子状态，关闭后无法提交任务
	state atomic.Int32
	// 时钟上下文，池子关闭时取消
	clockCtxCancel *ctx.CtxCancel
	// 清理上下文，池子关闭时取消
	clearCtxCancel *ctx.CtxCancel
}

// 提交任务到worker，worker从调度器获取
func (p *PoolWithFunc) Submit(task func()) error {
	if p.Closed() {
		return errors.ErrorPoolClosed
	}
	if w, err := p.scheduler.Get(); err == nil {
		w.Put(task)
		return nil
	}
	return errors.ErrorSubmitTaskFail
}

// 释放调度器资源
func (p *PoolWithFunc) Release() {
	p.Close()
	p.scheduler.Release()
	// 停止时钟和清理goroutine
	p.clearCtxCancel.Cancel()
	p.clockCtxCancel.Cancel()
}

// 等待调度器所有任务完成
func (p *PoolWithFunc) Wait() {
	p.scheduler.Wait()
}

// 释放调度器并等待所有任务完成
func (p *PoolWithFunc) ReleaseWithWait() {
	p.Close()
	p.scheduler.Release()
	p.scheduler.Wait() // 会坚持等待任务执行完成
	// 停止时钟和清理goroutine
	p.clearCtxCancel.Cancel()
	p.clockCtxCancel.Cancel()
}

// 带超时的释放调度器
func (p *PoolWithFunc) ReleaseWithTimeout(t time.Duration) error {
	p.Close()
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	p.scheduler.Release()
	select {
	case <-ctx.Done():
		return errors.ErrorPoolReleaseTimeout
	case <-p.scheduler.Done():
	}
	// 停止时钟和清理goroutine
	p.clearCtxCancel.Cancel()
	p.clockCtxCancel.Cancel()
	return nil
}

/* ------------------------------------------------- */
/* 监控需求 */
/* ------------------------------------------------- */

// 获取调度器的容量（最大worker数量）
func (p *PoolWithFunc) Cap() int32 {
	return p.scheduler.Cap()
}

// 获取调度器的空闲worker数量
func (p *PoolWithFunc) Free() int32 {
	return p.scheduler.Free()
}

// 获取调度器中正在运行的worker数量
func (p *PoolWithFunc) Running() int32 {
	return p.scheduler.Running()
}

// // 获取调度器中正在工作的worker数量
// func (p *PoolWithFunc) Working() int32 {
// 	return p.scheduler.Working()
// }

// 获取调度器中等待执行的任务数量
func (p *PoolWithFunc) Waiting() int32 {
	return p.scheduler.Waiting()
}

// 关闭池子
func (p *PoolWithFunc) Close() {
	p.state.Store(STATE_CLOSED)
}

// 打开池子
func (p *PoolWithFunc) Open() {
	p.state.Store(STATE_OPENED)
}

// 获取调度器是否已关闭
func (p *PoolWithFunc) Closed() bool {
	return p.scheduler.Closed()
}

// 获取调度器是否已打开
func (p *PoolWithFunc) Opened() bool {
	return p.scheduler.Opened()
}

/* ------------------------------------------------- */
/* 对外开放 */
/* ------------------------------------------------- */

// 池子的时钟功能，定期更新调度器的时间戳
//func (p *PoolWithFunc) clock(d time.Duration) {
//	p.clockCtxCancel = ctx.NewContextWithCancel(context.Background())
//	go func() {
//		ticker := time.NewTicker(d)
//		defer ticker.Stop()
//
//		// 先设置初始时间
//		p.scheduler.SetNowTime(time.Now())
//
//		for {
//			select {
//			case <-p.clockCtxCancel.ctx.Done():
//				return
//			case <-ticker.C:
//				if p.Closed() {
//					return
//				}
//				// 使用真实时间，或者如果想避免系统调用，至少保证初始值正确
//				p.scheduler.SetNowTime(time.Now())
//			}
//		}
//	}()
//}

// 清理过期的worker
func (p *PoolWithFunc) clear(d time.Duration) {
	if d == 0 {
		return
	}
	p.clearCtxCancel = ctx.NewContextWithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(d)
		defer func() {
			ticker.Stop()
		}()
		context := p.clearCtxCancel
		for {
			select {
			case <-context.Ctx.Done():
				return
			case <-ticker.C:
			}
			if p.Closed() {
				break
			}
			// 清理过期worker
			p.scheduler.ClearExpired(p.options.ExpiryDuration)
		}
	}()
}

type WorkersWithFuncCreator func(int) (scheduler_func.WorkersWithFunc, error)

func WorkerWithFuncCreator(s scheduler_func.Scheduler) scheduler_func.WorkerWithFunc {
	return scheduler_func.NewWorkerWithFunc(s)
}

func NewPoolWithFunc(
	cap int,
	workersCreator WorkersWithFuncCreator,
	fn func(func()),
	opt ...Option,
) (*PoolWithFunc, error) {
	workers, _ := workersCreator(cap)
	opts := NewOptions(opt...)
	scheduler := NewScheduler(int32(cap), workers, WorkerWithFuncCreator, fn, opts)

	// New pool
	p := &PoolWithFunc{
		options:        opts,
		scheduler:      scheduler,
		clockCtxCancel: ctx.NewContextWithCancel(context.Background()),
		clearCtxCancel: ctx.NewContextWithCancel(context.Background()),
	}
	p.Open()
	//p.clock(500 * time.Millisecond)
	p.clear(p.options.ExpiryDuration)
	return p, nil
}

// 使用默认的worker栈创建池子。
// cap是调度器容量，fn是任务处理函数，options是调度器配置选项。
func NewPoolWithFuncDefaultWorkers(
	cap int,
	fn func(func()),
	options ...Option,
) (*PoolWithFunc, error) {
	return NewPoolWithFunc(cap, scheduler_func.NewWorkersStackWithFunc, fn, options...)
}

// 使用默认的worker栈和默认的任务处理函数创建池子。
// cap是调度器容量，options是调度器配置选项。
func NewPoolWithFuncDefaultHandler(
	cap int,
	options ...Option,
) (*PoolWithFunc, error) {
	return NewPoolWithFunc(cap, scheduler_func.NewWorkersStackWithFunc, func(task func()) {
		task()
	}, options...)
}
