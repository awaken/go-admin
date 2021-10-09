package utils

import _ "unsafe"

// NanoTime returns the current time in nanoseconds from a monotonic clock.
//go:linkname NanoTime runtime.nanotime
func NanoTime() int64

