package utils

import "time"

type Benchmark int64

func StartBenchmark() Benchmark {
	return Benchmark(NanoTime())
}

func (b *Benchmark) Reset() {
	*b = Benchmark(NanoTime())
}

func (b Benchmark) ElapsedSeconds() float64 {
	return float64((NanoTime() - int64(b)) / int64(time.Millisecond)) / 1000
}

func (b Benchmark) ElapsedMillis() float64 {
	return float64((NanoTime() - int64(b)) / int64(time.Microsecond)) / 1000
}

func (b Benchmark) ElapsedMicros() float64 {
	return float64(NanoTime() - int64(b)) / 1000
}

func (b Benchmark) ElapsedNanos() int64 {
	return NanoTime() - int64(b)
}

func (b Benchmark) ElapsedDuration() time.Duration {
	return time.Duration(NanoTime() - int64(b))
}
