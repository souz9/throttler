package throttler

import (
	"sync/atomic"
	"time"
)

const minDelay = int64(time.Millisecond)

type Throttler struct {
	delay    int64 // time.Duration
	maxDelay int64 // time.Duration
	ticks    chan struct{}
}

// New creates a new instance of throttler.
// maxDelay controls what min time interval between consecutive
// operations is allowed at the top throttling level.
func New(maxDelay time.Duration) *Throttler {
	t := &Throttler{
		maxDelay: int64(maxDelay),
		ticks:    make(chan struct{}),
	}
	go t.ticker()
	return t
}

func (t *Throttler) ticker() {
	for {
		if delay := atomic.LoadInt64(&t.delay); delay > 0 {
			time.Sleep(time.Duration(delay))
		}
		t.ticks <- struct{}{}
	}
}

// Allow tries to run next operation.
// It returns true if the operation is actually allowed to run,
// otherwise the caller should wait for some time before running next operation.
func (t *Throttler) Allow() bool {
	if delay := atomic.LoadInt64(&t.delay); delay > 0 {
		select {
		case <-t.ticks:
		default:
			return false
		}
	}
	return true
}

// Up increases the throughput of the throttler.
func (t *Throttler) Up() {
	if delay := atomic.LoadInt64(&t.delay); delay > 0 {
		newDelay := delay / 2
		if newDelay < minDelay {
			newDelay = 0
		}
		atomic.CompareAndSwapInt64(&t.delay, delay, newDelay)
	}
}

// Down decreases the throughput of the throttler.
func (t *Throttler) Down() {
	if delay := atomic.LoadInt64(&t.delay); delay < t.maxDelay {
		newDelay := delay * 2
		if newDelay < minDelay {
			newDelay = minDelay
		}
		if newDelay > t.maxDelay {
			newDelay = t.maxDelay
		}
		atomic.CompareAndSwapInt64(&t.delay, delay, newDelay)
	}
}

// Adjust controls the throttling level depending on a result of some task execution.
// It throttles UP if the task is done successfully (done=true) or throttles
// DOWN otherwise (done=false).
func (t *Throttler) Adjust(done bool) {
	if done {
		t.Up()
	} else {
		t.Down()
	}
}
