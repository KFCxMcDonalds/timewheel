package timewheel

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/KFCxMcDonalds/timewheel/delayqueue"
	"github.com/panjf2000/ants/v2"

)

var POOL_SIZE = 1000

type TimeWheel struct {
	currentTime atomic.Int64 // ms
	tick     int64 // ms, tick time span
	wheelSize int64 // number of slots
	interval int64 // ms, circle time, interval = tick * wheelSize 

	buckets []*Bucket             // bucket list, each bucket contains a bi-directional linked list of tasks
	queue *delayqueue.DelayQueue // priority queue for slots

	overflowTW atomic.Pointer[TimeWheel] // contain tasks that exceed current time wheel's range

	wg sync.WaitGroup
	pool *ants.Pool
	exitCh chan struct{}
}

func New(tick time.Duration, wheelSize int64) *TimeWheel {
	tickMS := int64(tick / time.Millisecond)
	if tickMS <= 0 || wheelSize <= 0 {
		panic("tick span must be >= 1ms and wheelSize must be > 0")
	}
	startMS := time2MS(time.Now())
	pool , _:= ants.NewPool(POOL_SIZE, ants.WithNonblocking(true)) // if goroutine pool is full, degraded to goroutine
	return new(startMS, tickMS, wheelSize, delayqueue.New(int(wheelSize)), pool)
}

func new(startMS, tickMS, wheelSize int64, queue *delayqueue.DelayQueue, pool *ants.Pool) *TimeWheel {
	// initialize slots
	buckets := make([]*Bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	tw :=  &TimeWheel {
		currentTime: atomic.Int64{},
		tick: tickMS,
		wheelSize: wheelSize,
		interval: tickMS * wheelSize,
		buckets: buckets,
		queue: queue,
		pool: pool,
		exitCh: make(chan struct{}),
	}
	tw.currentTime.Store(truncate(startMS, tickMS))
	return tw
}

func (tw *TimeWheel) Start() {
	// start two goroutine, one for polling delay queue, another for running tasks
	// goroutine 1: poll delay queue
	tw.wg.Add(1)
	go func() {
		defer tw.wg.Done()
		tw.queue.Poll(tw.exitCh, func()int64{
			return time2MS(time.Now())
		})
	}()

	// goroutine 2: run tasks
	tw.wg.Add(1)
	go func() {
		defer tw.wg.Done()
		for {
			select {
			case item := <- tw.queue.C:
			    b := item.(*Bucket)
			    // advance time wheel clock
				tw.advanceClock(b.expiration.Load())
				// flush bucket's tasks
				b.Flush(tw.addOrRun)
			case <- tw.exitCh:
				return
			}
		}
	}()
}

func (tw *TimeWheel) Stop() {
	close(tw.exitCh)
	tw.wg.Wait()
	tw.pool.Release()
}

func (tw *TimeWheel) PlaceTimer(after time.Duration, run func()) {
	timer := &Timer{
		expiration: time2MS(time.Now().Add(after)),
		run:      run,
	}
	tw.addOrRun(timer)
}

func (tw *TimeWheel) addOrRun(timer *Timer) {
	// if expireAt < now, run immediately
	// INFO: maybe need rethink
	if timer.expiration < tw.currentTime.Load()+tw.tick {
		if tw.pool.Submit(timer.run) != nil {
			go timer.run()
		}
		return
	}
	// else add task into TW
	tw.add(timer)
}

func (tw *TimeWheel) add(timer *Timer) {
	// if expireAt is within current time wheel range
	if timer.expiration < tw.currentTime.Load()+tw.interval {
		// calculate bucket index
		virtualID := timer.expiration / tw.tick
		bucketInd := virtualID % tw.wheelSize    // abusolute bucket index in current time wheel
		// if index exists, add task into bucket (reuse bucket)
		b := tw.buckets[bucketInd]
		b.AddTimer(timer)
		
		// bucket expiration time is aligned with interval
		if b.SetExpiration(virtualID * tw.tick){
			// bucket which has been reused into delay queue
			expiration := b.expiration.Load()
			tw.queue.Enqueue(b, expiration)
		}
	// else recursively add into overflow time wheel
	} else {
		overflowTW := tw.getOrCreateOverflowTW()
		// recursively add timer into overflowTW
		overflowTW.add(timer)
	}
}

func (tw *TimeWheel) getOrCreateOverflowTW() *TimeWheel {
	if wheel := tw.overflowTW.Load(); wheel != nil {
		return wheel
	}
	// create new overflow time wheel
	newTW := new(tw.currentTime.Load(), tw.interval, tw.wheelSize, tw.queue, tw.pool)
	if tw.overflowTW.CompareAndSwap(nil, newTW) {
		// create successfully, return new wheel
		return newTW
	}
	// false create, return existing wheel
	return tw.overflowTW.Load()
}

func (tw *TimeWheel) advanceClock(expiration int64) {
	currentTime := tw.currentTime.Load()
	// move tick to next bucket
	if expiration >= currentTime+tw.tick {
		currentTime = truncate(expiration, tw.tick)
		tw.currentTime.Store(currentTime)

		// recursively advance overflow time wheel clock
		overflow := tw.overflowTW.Load()
		if overflow != nil {
			overflow.advanceClock(expiration)
		}
	}
}
