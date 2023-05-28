// Package cticker provides a ticker which ticks according to wall clock
// instead of monotonic clock. It is reliable under both clock drift and
// clock adjustments, that is if you ask it to tick on the minute it will
// ensure that it does so even if the underlying clock is inaccurate or
// gets adjusted.
package cticker

import (
	"fmt"
	"sync"
	"time"
)

// ticker is an interface that allows us to mock the runtime ticker for
// testing.
type ticker interface {
	Ch() <-chan time.Time
	Stop()
}

// runtimeTicker wraps time.Ticker implementing ticker.
type runtimeTicker struct {
	*time.Ticker
}

func newRuntimeTicker(d time.Duration) ticker {
	return &runtimeTicker{Ticker: time.NewTicker(d)}
}

// Ch implements ticker.
func (t *runtimeTicker) Ch() <-chan time.Time {
	return t.C
}

// newTicker is the method used to create a new ticker so we can mock it
// for testing.
var newTicker = newRuntimeTicker

// Ticker holds a channel that delivers `ticks` on wall clock boundaries.
// Unlike time.Ticker it will fire early if the clock is adjusted to an
// earlier time, therefore it can be used for events which need to trigger
// on wall clock boundaries e.g. every minute on the minute.
type Ticker struct {
	mtx    sync.Mutex
	C      <-chan time.Time // The channel on which ticks are delivered.
	d      time.Duration
	a      time.Duration
	next   time.Time
	done   chan struct{}
	ticker ticker
}

// New returns a new Ticker containing a channel that will send the
// time at d wall clock boundaries plus accuracy.
//
// It will drop ticks to make up for slow receivers.
// The duration d must be greater than zero; if not, NewTicker will panic.
// The accuracy must be less than d; it not, NewTicker will panic.
//
// If the accuracy is too small, the ticker may not be able to tick at
// the requested time. Instead it will tick at the next available time
// after the target time, which should be within accuracy.
//
// Stop the ticker to release its associated resources.
func New(d, accuracy time.Duration) *Ticker {
	if d <= accuracy {
		panic(fmt.Errorf("accuracy %v is not less than duration %v", accuracy, d))
	}
	// We hold one element time buffer, if the consumer falls behind
	// while reading the values, we drop the ticks until it catches
	// back up.
	c := make(chan time.Time, 1)
	now := time.Now()
	t := &Ticker{
		C:    c,
		d:    d,
		a:    accuracy,
		next: now.Truncate(d).Add(d),
		done: make(chan struct{}),
	}

	go func() {
		// Synchronise to the accuracy.
		time.Sleep(now.Truncate(accuracy).Add(accuracy).Sub(now))

		t.mtx.Lock()
		defer t.mtx.Unlock()
		select {
		case <-t.done:
			// Already stopped.
		default:
			t.ticker = newTicker(accuracy)
			go t.tick(c)
		}

	}()

	return t
}

// Stop turns off the ticker. After Stop, no more ticks will be sent.
// Stop does not close the channel to prevent a read from the channel
// succeeding incorrectly.
func (t *Ticker) Stop() {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	close(t.done)
	if t.ticker != nil {
		t.ticker.Stop()
	}
}

// tick sends ticks to t.C with the tickers defined accuracy.
func (t *Ticker) tick(c chan time.Time) {
	ticks := t.ticker.Ch()
	for {
		select {
		case now := <-ticks:
			// Remove monotonic clock as we want wall clock comparisons.
			now = now.Truncate(t.a)

			// Tick if are within accuracy of the target time or it has past.
			// This ensures that we tick even if the requested accuracy is
			// not achievable, for example time.NanoSecond.
			if now.Compare(t.next) >= 0 {
				t.next = t.next.Add(t.d)
				select {
				case c <- now:
				default:
					// Consumer is running slowly
				}
			}
		case <-t.done:
			return
		}
	}
}
