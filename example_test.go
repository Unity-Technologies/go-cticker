package cticker_test

import (
	"fmt"
	"time"

	"github.com/multiplay/go-cticker"
)

func ExampleTicker() {
	// Create a ticker that ticks on the minute with 1 second accuracy.
	t := cticker.New(time.Minute, time.Second)
	defer t.Stop()

	<-t.C

	// Process tick
	fmt.Println("tick")
	// Output: tick
}
