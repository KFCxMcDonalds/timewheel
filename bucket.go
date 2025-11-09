package timewheel

import (
	"container/list"
	"sync"
	"sync/atomic"
)

type Timer struct {
	expiration int64  // ms, timer expire time
	run        func() // timer task to run
	element    *list.Element
}

type Bucket struct {
	expiration atomic.Int64 // ms, bucket expire time in unix ms
	timers     *list.List   // list of timers in this bucket

	mu sync.Mutex
}

func newBucket() *Bucket {
	b := &Bucket{
		expiration: atomic.Int64{},
		timers:     list.New(),
	}
	b.expiration.Store(-1)
	return b
}

func (b *Bucket) AddTimer(timer *Timer) {
	b.mu.Lock()
	// prev and next pointers are wrapped in list.Element
	e := b.timers.PushBack(timer)
	timer.element = e
	b.mu.Unlock()
}

func (b *Bucket) SetExpiration(newExpiration int64) bool {
	// set bucket expiration time
	// if old != new, means old bucket has been updated and reused, which need to be added into delay queue
	// else do nothing (bucket already in delay queue)
	return b.expiration.Swap(newExpiration) != newExpiration
}

func (b *Bucket) Flush(reinsertOrRun func(timer *Timer)) {
	ts := []*Timer{}

	b.mu.Lock()
	for e := b.timers.Front(); e != nil; {
		next := e.Next()

		timer := e.Value.(*Timer)
		b.timers.Remove(e)
		ts = append(ts, timer)

		e = next
	}
	b.mu.Unlock()

	// reset expiration
	b.expiration.Store(-1)

	// run all timers
	go func() {
		for _, t := range ts {
			// if current bucket is from overflowTW, timers will be reinsert into timewheel
			reinsertOrRun(t)
		}
	}()
}
