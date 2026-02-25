package scheduler_generic

import (
	"time"
)

type worker[T any] struct {
	task      chan T        // 需要执行的task
	exit      chan struct{} // 退出信号通知
	scheduler Scheduler[T]  // 这个worker受哪个scheduler控制
	usedTime  time.Time     // 上次运行的时间
}

func (w *worker[T]) Put(task T) {
	w.task <- task
}

func (w *worker[T]) Run() {
	go func() {
		defer func() {
			_ = w.scheduler.PutCache(w) // 将对象放在缓冲池中
			w.scheduler.Recover()       // 有报错就处理报错
		}()

		for {
			select {
			case <-w.exit:
				return
			case task := <-w.task:
				handler := w.scheduler.Handler() // 获取该scheduler的handler处理函数
				handler(task)                    // 执行task
				if err := w.scheduler.PutReady(w); err != nil {
					return
				}
			}
		}
	}()
}

func (w *worker[T]) Finish() {
	w.exit <- struct{}{}
}

func (w *worker[T]) Refresh() {
	w.usedTime = time.Now()
}

func (w *worker[T]) GetUsedTime() time.Time {
	return w.usedTime
}

func NewWorker[T any](s Scheduler[T]) Worker[T] {
	return &worker[T]{
		task:      make(chan T, 1),
		exit:      make(chan struct{}, 1),
		scheduler: s,
		usedTime:  time.Now(),
	}
}
