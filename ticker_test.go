package cticker

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockTicker struct {
	c      chan time.Time
	done   chan struct{}
	period time.Duration
	times  []time.Time
}

func (t *mockTicker) run() {
	tick := time.NewTicker(t.period)
	defer tick.Stop()

	for _, tm := range t.times {
		select {
		case <-tick.C:
			t.c <- tm
		case <-t.done:
			return
		}
	}
}

func (t *mockTicker) Ch() <-chan time.Time {
	return t.c
}

func (t *mockTicker) Stop() {
	close(t.done)
}

func testSetup() (time.Duration, time.Duration, time.Time) {
	d := time.Minute
	a := time.Second
	now := time.Now()

	// Ensure we won't trigger the missed tick on the first value.
	if now.Truncate(a).Add(-a).Equal(now.Truncate(d)) {
		now = now.Add(a * 2)
	}

	return d, a, now
}

// TestTickerDrift checks that Ticker ticks based on the 'values' returned
// by the underlying ticker, confirming it will adjust to clock drifts.
func TestTickerDrift(t *testing.T) {
	d, a, now := testSetup()

	times := make([]time.Time, 0, 4*d/time.Second)
	times = addTimes(times, now, now.Add(time.Minute*4), a)

	ticks := make([]time.Time, 0, 4)
	ticks = addTicks(ticks, now, now.Add(time.Minute*3), d, a)

	last := testTicker(t, now, d, a, times, ticks)
	now = time.Now()
	assert.True(t, now.Before(last), fmt.Sprintln(now, "is not before", last))
}

// TestTickerAdjustment checks that Ticker deals with clock adjustments.
func TestTickerAdjustment(t *testing.T) {
	d, a, now := testSetup()
	summer := now.Add(-time.Hour)

	times := make([]time.Time, 0, 4*d/a)
	times = addTimes(times, now, now.Add(time.Minute), a)
	times = addTimes(times, summer.Add(d*2), summer.Add(d*5), a)

	ticks := make([]time.Time, 0, 4)
	ticks = append(ticks, now.Truncate(d).Add(d))
	ticks = addTicks(ticks, summer.Add(d*2), summer.Add(d*4), d, a)

	last := testTicker(t, now, d, a, times, ticks)
	now = time.Now()
	assert.True(t, now.After(last), fmt.Sprintln(now, "is not after", last))
}

// TestTickerLostTick checks that Ticker deals with clock adjustments
// which result in the underlying tick that we would tick on being missed.
func TestTickerLostTick(t *testing.T) {
	d, a, now := testSetup()

	times := make([]time.Time, 0, d/a)
	times = addTimes(times, now, now.Add(time.Minute*4), a)

	ticks := make([]time.Time, 0, 4)
	ticks = addTicks(ticks, now, now.Add(time.Minute*3), d, a)

	// Remove the time which would match the first tick
	for i, t := range times {
		if t.Truncate(d) == ticks[0] {
			times = append(times[:i], times[i+1:]...)
			ticks[0] = ticks[0].Add(a)
			break
		}
	}

	last := testTicker(t, now, d, a, times, ticks)
	now = time.Now()
	assert.True(t, now.Before(last), fmt.Sprintln(now, "is not before", last))
}

// addTimes adds underlying times that we will replay to times and returns the resultant slice.
func addTimes(times []time.Time, start, end time.Time, a time.Duration) []time.Time {
	for tick := start; tick.Before(end) || tick.Equal(end); tick = tick.Add(a) {
		times = append(times, tick)
	}

	return times
}

// addTicks adds the expected ticks to ticks and returns the resultant slice.
func addTicks(ticks []time.Time, start, end time.Time, d, a time.Duration) []time.Time {
	for tick := start; tick.Before(end) || tick.Equal(end); tick = tick.Add(d) {
		t := tick.Truncate(d)
		if t.Equal(tick.Truncate(a)) {
			// On tick
			ticks = append(ticks, t)
		} else {
			// After the tick
			ticks = append(ticks, t.Add(d))
		}
	}
	return ticks
}

func testTicker(t *testing.T, now time.Time, d, a time.Duration, times, ticks []time.Time) time.Time {
	newTicker = func(d time.Duration) ticker {
		t := &mockTicker{
			c:      make(chan time.Time, 4),
			done:   make(chan struct{}),
			period: time.Millisecond,
			times:  times,
		}
		go t.run()
		return t
	}

	tk := New(d, a)
	defer tk.Stop()

	var i int
	var tick time.Time
	for {
		tick = <-tk.C
		assert.Equal(t, ticks[i], tick)
		i++
		if i == len(ticks) {
			break
		}
	}

	return tick
}

func Test_issue4(t *testing.T) {
	duration := time.Second
	ticker := New(duration, time.Millisecond)
	timeout := time.After(time.Second * 2)
	select {
	case <-ticker.C:
	case <-timeout:
		t.Fatal("timeout")
	}
}

func TestTicker_StopClose(t *testing.T) {
	ticker := New(time.Millisecond, time.Nanosecond)
	go func() {
		time.Sleep(time.Millisecond * 10)
		ticker.StopClose()
	}()

	ticks := 0
	for range ticker.C {
		ticks++
	}
	assert.GreaterOrEqual(t, 9, ticks)
}
