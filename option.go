package timewheel

import "github.com/sirupsen/logrus"

type Option func(*TimeWheel)

func DefaultOptions() []Option {
	return []Option{
		WithPoolSize(1000),
		WithLogger(logrus.New()),
		WithPanicHandler(func(p any) {
			logrus.Errorf("time wheel panic: %v", p)
		}),
	}
}

func WithLogger(logger Logger) Option {
	return func(tw *TimeWheel) {
		tw.logger = logger
	}
}

func WithPanicHandler(handler func(any)) Option {
	return func(tw *TimeWheel) {
		tw.panicHanlder = handler
	}
}

func WithPoolSize(size int) Option {
	return func(tw *TimeWheel) {
		tw.poolSize = size
	}
}

func inheritOptions(tw *TimeWheel, parent *TimeWheel) {
	tw.logger = parent.logger
	tw.panicHanlder = parent.panicHanlder
	tw.pool = parent.pool
	tw.exitCh = parent.exitCh
	tw.queue = parent.queue
}

