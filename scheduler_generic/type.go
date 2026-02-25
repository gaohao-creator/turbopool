package scheduler_generic

import "time"

type Worker[T any] interface {
	Put(task T) // 添加任务
	Run()       // 开始运行
	Finish()    // 停止运行
	GetUsedTime() time.Time
	Refresh() // 更新运行时间
}

type Workers[T any] interface {
	Len() int
	IsEmpty() bool
	Push(e Worker[T]) error
	Pop() (Worker[T], error)
	Clear() error
	ClearExpired(t time.Time) (int, error)
	Scale(cap int32) error
}

type Scheduler[T any] interface {
	Get() (Worker[T], error)             // 获取worker
	Handler() func(T)                    // 任务处理逻辑
	PutReady(w Worker[T]) error          // 将worker放入就绪队列
	PutCache(w Worker[T]) error          // 将worker放入sync.Pool
	Recover()                            // 统一处理任务 panic，优先使用自定义处理器或日志
	ClearExpired(duration time.Duration) // 清理过期worker

	Cap() int32     // worker总容量
	Free() int32    // 当前还可容纳的worker数量
	Running() int32 // 当前正在运行的worker总数量
	Waiting() int32 // 阻塞模式下等待的任务数量
	Opened() bool
	Closed() bool

	Open()               // 开始调度
	Close()              // 结束调度
	Wait()               // 等待任务完成
	Release()            // 释放资源
	Done() chan struct{} // 调度器生命周期的监听

	Scale(cap int32)
}
