package scheduler_func

import (
	"time"
)

type workerWithFunc struct {
	task      chan func() // 需要执行的task
	scheduler Scheduler   // 这个worker受哪个scheduler控制
	usedTime  time.Time   // 上次运行的时间
}

func (w *workerWithFunc) Put(task func()) {
	w.task <- task
}

func (w *workerWithFunc) Run() {
	go func() {
		defer func() {
			_ = w.scheduler.PutCache(w) // 将对象放在缓冲池中
			w.scheduler.Recover()       // 有报错就处理报错
		}()
		for task := range w.task {
			if task == nil {
				return
			}
			handler := w.scheduler.Handler() // 获取该scheduler的handler处理函数
			handler(task)                    // 执行task
			if err := w.scheduler.PutReady(w); err != nil {
				return
			}
		}
	}()
}

func (w *workerWithFunc) Finish() {
	w.task <- nil
}

func (w *workerWithFunc) Refresh() {
	w.usedTime = time.Now()
}

func (w *workerWithFunc) GetUsedTime() time.Time {
	return w.usedTime
}

func NewWorkerWithFunc(s Scheduler) WorkerWithFunc {
	return &workerWithFunc{
		task:      make(chan func(), 1),
		scheduler: s,
		usedTime:  time.Now(),
	}
}
