package timewheel_test

import (
	"testing"
	"time"

	"github.com/KFCxMcDonalds/timewheel"
)

func TestTimeWheel_Start(t *testing.T) {
	tw := timewheel.New(1*time.Millisecond,60, 1000)
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
		t.Run("", func(t *testing.T){
			start := time.Now()
			timeC := make(chan time.Time)

			tw.PlaceTimer(d, func(){
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
