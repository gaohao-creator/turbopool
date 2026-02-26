package turbopool

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gaohao-creator/turbopool/errors"
	"github.com/gaohao-creator/turbopool/scheduler_func"
)

type SchedulerWithFunc struct {
	// 整体状态
	state    atomic.Int32  // 状态（开、关）
	lock     *sync.Mutex   // 互斥锁
	cond     *sync.Cond    // 条件锁
	done     chan struct{} // 完成信号
	doneOnce *sync.Once    // 仅关闭一次

	// worker容器
	capacity     atomic.Int32                   // 允许同时存在的最多worker数量
	readyWorkers scheduler_func.WorkersWithFunc // 就绪的worker队列（正在Run的）
	cacheWorkers *sync.Pool                     // 对象池 （没在Run的）
	running      atomic.Int32                   // 正在运行的worker数量
	waiting      atomic.Int32                   // 等待的任务数

	// 任务运行层次控制
	preHook  func()       // 前置钩子
	postHook func()       // 后置钩子
	handler  func(func()) // 任务处理函数,可设置一些前置钩子和后置钩子（pre-hook \ post-hook）

	// option
	options *Options // 配置选项
	//Nonblocking      bool             // 是否开始阻塞功能
	//MaxBlockingTasks int32            // 阻塞模式下最大阻塞数量
	//ExpiryDuration   time.Duration    // worker 空闲过期时间
	//PanicHandler     func(any)        // 自定义 panic 处理器
	//Logger           logger.LoggerV1  // 自定义日志处理器
}

// 获取worker
func (s *SchedulerWithFunc) Get() (scheduler_func.WorkerWithFunc, error) {
	// 1) 先尝试从 ready 队列获取
	if w, err := s.readyWorkers.Pop(); err == nil {
		return w, nil
	}

	// 2) ready 为空时再判断是否需要阻塞
	if s.state.Load() == STATE_OPENED && s.Free() <= 0 {
		if err := s.blocking(); err != nil {
			return nil, err
		}
		// 阻塞结束后再尝试 Pop 一次
		if w, err := s.readyWorkers.Pop(); err == nil {
			return w, nil
		}
	}

	// 3) 需要新建 worker
	w := s.cacheWorkers.Get().(scheduler_func.WorkerWithFunc)
	w.Run()
	s.addRunning(1)
	return w, nil
}

// 将worker放入就绪队列
func (s *SchedulerWithFunc) PutReady(w scheduler_func.WorkerWithFunc) error {
	// 调度器已关闭或无空闲容量时不归还，返回 false 使 worker 结束并走 Cache
	if s.state.Load() == STATE_CLOSED {
		return errors.ErrorSchedulerClosed
	}
	if err := s.readyWorkers.Push(w); err != nil {
		return err
	}
	w.Refresh()
	s.cond.Signal()
	return nil
}

// 将worker放入sync.Pool
func (s *SchedulerWithFunc) PutCache(w scheduler_func.WorkerWithFunc) error {
	s.addRunning(-1)
	s.cacheWorkers.Put(w)
	s.cond.Signal()
	return nil
}

// 统一处理任务 panic，优先使用自定义处理器或日志
func (s *SchedulerWithFunc) Recover() {
	p := recover()
	if p == nil {
		return
	}
	if ph := s.options.PanicHandler; ph != nil {
		ph(p)
		return
	}
	if logger := s.options.Logger; logger != nil {
		logger.Printf("worker exits from panic: %v\n%s\n", p, debug.Stack())
		return
	}
}

func (s *SchedulerWithFunc) Handler() func(func()) {
	return s.handler
}

func (s *SchedulerWithFunc) ClearExpired(duration time.Duration) {
	if s.readyWorkers.IsEmpty() {
		return
	}
	if duration == 0 && s.options.ExpiryDuration != 0 {
		duration = s.options.ExpiryDuration
	}
	t := time.Now().Add(-duration)
	clearCount, _ := s.readyWorkers.ClearExpired(t)
	// 清理后如有等待任务则唤醒
	if clearCount > 0 && s.Waiting() > 0 {
		s.cond.Broadcast() // 唤醒阻塞队列，因为free的空间增大了
	}
}

// Release 关闭调度器并清空就绪的worker，同时避免阻塞的goroutine泄露
func (s *SchedulerWithFunc) Release() {
	s.Close()

	// 清空就绪 worker 释放内存
	s.readyWorkers.Clear()

	// 唤醒所有等待方,避免goroutine泄露
	s.cond.Broadcast()

	// 最终检查并关闭 done
	if s.Closed() && s.Running() == 0 {
		s.doneOnce.Do(func() {
			close(s.done) // 通知调度器已完成
		})
	}
}

func (s *SchedulerWithFunc) Wait() {
	if s.Running() > 0 {
		<-s.Done()
	}
}

// ReleaseWithWait 释放调度器并等待全部结束。
func (s *SchedulerWithFunc) ReleaseWithWait() {
	s.Release()
	s.Wait()
}

/* ------------------------------------------------- */
/* 监控需求 */
/* ------------------------------------------------- */

func (s *SchedulerWithFunc) Cap() int32 {
	return s.capacity.Load()
}

func (s *SchedulerWithFunc) Free() int32 {
	return s.capacity.Load() - s.running.Load()
}

func (s *SchedulerWithFunc) Running() int32 {
	return s.running.Load()
}

func (s *SchedulerWithFunc) Waiting() int32 {
	return s.waiting.Load()
}

func (s *SchedulerWithFunc) Open() {
	s.state.Store(STATE_OPENED)
}

func (s *SchedulerWithFunc) Close() {
	s.state.Store(STATE_CLOSED)
}

// Opened 判断调度器是否已开启。
func (s *SchedulerWithFunc) Opened() bool {
	return s.state.Load() == STATE_OPENED
}

// Closed 判断调度器是否已关闭。
func (s *SchedulerWithFunc) Closed() bool {
	return s.state.Load() == STATE_CLOSED
}

func (s *SchedulerWithFunc) Done() chan struct{} {
	return s.done
}

func (s *SchedulerWithFunc) Scale(cap int32) {
	s.capacity.Store(cap)
}

func (s *SchedulerWithFunc) addRunning(delta int32) int32 {
	return s.running.Add(delta)
}

// blocking 阻塞获取worker
func (s *SchedulerWithFunc) blocking() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	// 检查调度器是否开启
	if s.state.Load() == STATE_CLOSED {
		return errors.ErrorSchedulerClosed
	}
	s.waiting.Add(1)
	opened := s.Opened() //检查调度器是否处于开启状态
	free := s.Free()     // 获取空闲worker数量（容量 - 运行数）
	for opened && free <= 0 && s.readyWorkers.IsEmpty() {
		if s.options.Nonblocking ||
			(s.options.MaxBlockingTasks != 0 && s.Waiting() >= int32(s.options.MaxBlockingTasks)) {
			s.waiting.Add(-1)
			return errors.ErrorSchedulerIsFull // 返回"调度器已满"错误
		}
		s.cond.Wait()       // 开始阻塞等待
		opened = s.Opened() // 等待任务数 -1（获取到worker了，准备退出）
		free = s.Free()
	}
	s.waiting.Add(-1)
	return nil
}

func NewScheduler(
	cap int32,
	workers scheduler_func.WorkersWithFunc,
	workerFunc func(scheduler_func.Scheduler) scheduler_func.WorkerWithFunc, // 工厂函数
	handler func(func()),
	opts *Options) *SchedulerWithFunc {
	s := &SchedulerWithFunc{
		state:        atomic.Int32{},
		lock:         &sync.Mutex{},
		done:         make(chan struct{}),
		doneOnce:     &sync.Once{},
		readyWorkers: workers,
		cacheWorkers: &sync.Pool{},
		running:      atomic.Int32{},
		waiting:      atomic.Int32{},
		handler:      handler,
		options:      opts,
	}
	s.cond = sync.NewCond(s.lock)
	s.capacity.Store(cap)
	s.cacheWorkers.New = func() any {
		return workerFunc(s) // 如果不存在缓存，就调用工厂函数来生产一个新的对象
	}
	return s
}
