package delayqueue

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

type DelayQueue struct {
	pq priorityQueue

	C        chan any
	wakeupCh chan struct{}

	mu    sync.Mutex
	sleep atomic.Int32
}

func New(size int) *DelayQueue {
	return &DelayQueue{
		pq:       make(priorityQueue, 0, size),
		C:        make(chan any, size), // buffered channel to avoid blocking
		wakeupCh: make(chan struct{}),
		mu:       sync.Mutex{},
		sleep:    atomic.Int32{},
	}
}

// use this function to peek the first item in pq, you have to hold the lock before calling it
func (dq *DelayQueue) peek(now int64) (item *Item, after int64) {
	// if no item in pq, return
	if dq.pq.Len() == 0 {
		return nil, 0
	}
	// else at least one item in pq
	// if first item's expiration time > now, return duration time poller should wait
	first := dq.pq[0]
	if first.priority > now {
		return nil, first.priority - now
	}
	heap.Pop(&dq.pq)
	// else it needs to be execute now
	return first, 0
}

func (dq *DelayQueue) Enqueue(b any, priority int64) {
	item := &Item{
		value:    b,
		priority: priority,
	}
	dq.mu.Lock()
	heap.Push(&dq.pq, item)
	index := item.index
	dq.mu.Unlock()
	if index == 0 {
		// new the earlist block enqueue, wake up poller
		if dq.sleep.CompareAndSwap(1, 0) {
			dq.wakeupCh <- struct{}{}
		}
	}
}

func (dq *DelayQueue) Poll(exitCh chan struct{}, nowF func() int64) {
	for {
		now := nowF()

		dq.mu.Lock()
		item, after := dq.peek(now)
		if item == nil {
			dq.sleep.Store(1) // set sleep status
		}
		dq.mu.Unlock()

		if item == nil {
			if after == 0 {
				// no item to process now, wait for wakeupCh
				select {
				case <-exitCh:
					goto exit
				case <-dq.wakeupCh:
					continue
				}
			} else if after > 0 {
				// items in pq, but no item expired, wait for time.After or wakeupCh
				select {
				case <-dq.wakeupCh:
					continue
				case <-time.After(time.Duration(after) * time.Millisecond):
					if dq.sleep.Swap(0) == 0 {
						// A caller of Enqueue() is being blocked on sending to wakeupC,
						// drain wakeupCh to unblock the caller.
						<-dq.wakeupCh
					}
					continue
				case <-exitCh:
					goto exit
				}
			}
		}
		// process item
		select {
		case dq.C <- item.value:
			// bucket sent successfully
		case <-exitCh:
			goto exit
		}
	}
exit:
	// reset sleep status
	dq.sleep.Store(0)
}
