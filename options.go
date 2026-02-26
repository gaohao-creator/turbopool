package turbopool

import "time"

const (
	STATE_OPENED = int32(iota)
	STATE_CLOSED
)

// Custom Logger interface
type Logger interface {
	Printf(format string, args ...any)
}

type Options struct {
	// Blocking option, blocking submit task when no free worker.
	Nonblocking bool
	// blocking submit task max value.
	MaxBlockingTasks int
	// Worker expiry duration
	ExpiryDuration time.Duration
	// Recover panic handler.
	PanicHandler func(any)
	// Custom Logger
	Logger Logger
}

type Option func(opts *Options)

func WithNonblocking(nonblocking bool) Option {
	return func(opts *Options) {
		opts.Nonblocking = nonblocking
	}
}

func WithMaxBlockingTasks(maxBlockingTasks int) Option {
	return func(opts *Options) {
		opts.MaxBlockingTasks = maxBlockingTasks
	}
}

func WithExpiryDuration(expiryDuration time.Duration) Option {
	return func(opts *Options) {
		opts.ExpiryDuration = expiryDuration
	}
}

func WithPanicHandler(panicHandler func(any)) Option {
	return func(opts *Options) {
		opts.PanicHandler = panicHandler
	}
}

func WithLogger(logger Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

func NewOptions(options ...Option) *Options {
	opts := &Options{
		Nonblocking:      false,
		MaxBlockingTasks: 0,
		ExpiryDuration:   1000 * time.Millisecond,
	}
	for _, option := range options {
		option(opts)
	}
	return opts
}

func SetOptions(opts *Options, options ...Option) *Options {
	for _, option := range options {
		option(opts)
	}
	return opts
}
