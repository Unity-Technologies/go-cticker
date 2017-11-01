# go-cticker [![Go Report Card](https://goreportcard.com/badge/github.com/multiplay/go-cticker)](https://goreportcard.com/report/github.com/multiplay/go-cticker) [![License](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/multiplay/go-cticker/blob/master/LICENSE) [![GoDoc](https://godoc.org/github.com/multiplay/go-cticker?status.svg)](https://godoc.org/github.com/multiplay/go-cticker) [![Build Status](https://travis-ci.org/multiplay/go-cticker.svg?branch=master)](https://travis-ci.org/multiplay/go-cticker)

go-cticker is a [Go](http://golang.org/) library that provides a ticker which ticks according to wall clock and is reliable under clock drift and clock adjustments; that is if you ask it to tick on the minute it will ensure that it does so even if the underlying clock is inaccurate or gets adjusted.

Features
--------
* Reliable under clock drift.
* Reliable under clock adjustments.

Installation
------------
```sh
go get -u github.com/multiplay/go-cticker
```

Examples
--------

Using go-cticker is very much like time.NewTicker with the addition of an accuracy and start time.

The following creates a ticker which ticks on the minute according the hosts wall clock with an accuracy of plus or minus one second.
```go
package main

import (
	"fmt"
	"time"

	"github.com/multiplay/go-cticker"
)

func main() {
	t := cticker.New(time.Minute, time.Second)
	for tick := range t.C {
		// Process tick
		fmt.Println("tick:", tick)
	}
}
```

Documentation
-------------
- [GoDoc API Reference](http://godoc.org/github.com/multiplay/go-cticker).

License
-------
go-cticker is available under the [BSD 2-Clause License](https://opensource.org/licenses/BSD-2-Clause).
```
