package cticker_test

import (
	"fmt"
	"time"

	"github.com/multiplay/go-cticker"
)

// TODO(steve): remove this nolint when go tool vet is fixed.
// nolint: vet
func ExampleTicker() {
	t := cticker.New(time.Minute, time.Second)
	for tick := range t.C {
		// Process tick
		fmt.Println("tick:", tick)
	}
}
