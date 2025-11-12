package timewheel_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/KFCxMcDonalds/timewheel"
)

func TestTimeWheel_Start(t *testing.T) {
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
		t.Run("Timewheel_PlaceTimer", func(t *testing.T){
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
		tw.PlaceTimer(10*time.Millisecond, func(){
			panic("test panic")
		})

		time.Sleep(50 * time.Millisecond) // wait for the timer to execute

		if !strings.Contains(panicMsg, "test panic") {
			t.Errorf("Panic handler was not called correctly, got: %s", panicMsg)
		}

	})

}
