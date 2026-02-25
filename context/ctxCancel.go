package context

import (
	"context"
	"time"
)

type CtxCancel struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

func NewContextWithCancel(parent context.Context) *CtxCancel {
	ctx, cancel := context.WithCancel(parent)
	return &CtxCancel{
		Ctx:    ctx,
		Cancel: cancel,
	}
}

func NewContextWithTimeout(parent context.Context, timeout time.Duration) *CtxCancel {
	ctx, cancel := context.WithTimeout(parent, timeout)
	return &CtxCancel{
		Ctx:    ctx,
		Cancel: cancel,
	}
}
