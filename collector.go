package aura

import "time"

// Collector is the interface implemented by anything that can be used by
// Reporter to report metrics. A Collector has to be registered for collection.
type Collector interface {
	Interval() time.Duration
	Describe(chan<- *Desc)
	Collect(ch chan<- Metric)
}
