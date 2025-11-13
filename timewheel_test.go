package timewheel_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/KFCxMcDonalds/timewheel"
)

func TestTimeWheel_PlaceTimerAfter(t *testing.T) {
	tw := timewheel.New(1*time.Millisecond,60)
	tw.Start()
	defer tw.Stop()

	durations := []time.Duration{
		1 *time.Millisecond,
		10 *time.Millisecond,
		100 *time.Millisecond,
		500 *time.Millisecond,
		1 *time.Second,
		2 *time.Second,
	}

	for _, d := range durations {
		t.Run("Timewheel_PlaceTimerAfter", func(t *testing.T){
			start := time.Now()
			timeC := make(chan time.Time)

			tw.PlaceTimerAfter(d, func(){
				timeC <- time.Now()
			})
			got := (<-timeC).Truncate(time.Millisecond)

			min := start.Add(d).Truncate(time.Millisecond)
			errT := 3 * time.Millisecond
			max := min.Add(errT).Truncate(time.Millisecond)
			if got.Before(min) || got.After(max) {
				t.Errorf("Timer executed at %v, expected between %v and %v", got, min, max)
			}
		})
	}
}

func TestTimeWheel_PlaceTimerAt(t *testing.T) {
	tw := timewheel.New(1*time.Millisecond, 60)
	tw.Start()
	defer tw.Stop()

	durations := []time.Duration{
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
	}

	for _, d := range durations {
		t.Run("Timewheel_PlaceTimerAt", func(t *testing.T) {
			start := time.Now()
			timeC := make(chan time.Time)
			targetTime := start.Add(d)

			tw.PlaceTimerAt(targetTime, func() {
				timeC <- time.Now()
			})
			got := (<-timeC).Truncate(time.Millisecond)

			min := targetTime.Truncate(time.Millisecond)
			errT := 3 * time.Millisecond
			max := min.Add(errT).Truncate(time.Millisecond)
			if got.Before(min) || got.After(max) {
				t.Errorf("Timer executed at %v, expected between %v and %v", got, min, max)
			}
		})
	}
}

func TestTimeWheel_PlaceTimerAt_PastTime(t *testing.T) {
	tw := timewheel.New(1*time.Millisecond, 60)
	tw.Start()
	defer tw.Stop()

	t.Run("Timewheel_PlaceTimerAt_PastTime", func(t *testing.T) {
		executed := make(chan bool, 1)
		pastTime := time.Now().Add(-100 * time.Millisecond)

		tw.PlaceTimerAt(pastTime, func() {
			executed <- true
		})

		select {
		case <-executed:
			// Task should execute immediately
		case <-time.After(50 * time.Millisecond):
			t.Error("Task with past time should execute immediately")
		}
	})
}

func TestTimeWheel_Panic(t *testing.T) {
	panicMsg := ""
	handler := func(p any) {
		panicMsg = fmt.Sprintf("%v", p)
	}
	tw := timewheel.New(1*time.Millisecond,60,
		timewheel.WithPanicHandler(handler),
	)
	tw.Start()
	defer tw.Stop()

	t.Run("Timewheel_PanicRecovery", func(t *testing.T){
		tw.PlaceTimerAfter(10*time.Millisecond, func(){
			panic("test panic")
		})

		time.Sleep(50 * time.Millisecond) // wait for the timer to execute

		if !strings.Contains(panicMsg, "test panic") {
			t.Errorf("Panic handler was not called correctly, got: %s", panicMsg)
		}

	})

}
