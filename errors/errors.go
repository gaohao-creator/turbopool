package errors

import "errors"

var (

	// stack Errors
	ErrorsWorkerStackFull = errors.New("workerstack full")

	// Scheduler Errors
	ErrorSchedulerClosed   = errors.New("scheduler_func is closed")
	ErrorSchedulerOpened   = errors.New("scheduler_func is opened")
	ErrorSchedulerIsFull   = errors.New("scheduler_func is full")
	ErrorWorkersIsEmpty    = errors.New("workers is empty")
	ErrorFetchChanIsClosed = errors.New("fetch channel is closed")

	// Pool Errors
	ErrorPoolClosed         = errors.New("pool is closed")
	ErrorPoolOpened         = errors.New("pool is opened")
	ErrorPoolReleaseTimeout = errors.New("release pool timeout")
	ErrorSubmitTaskFail     = errors.New("submit task fail")
	ErrorSubmitTaskTimeout  = errors.New("submit task timeout")
)
