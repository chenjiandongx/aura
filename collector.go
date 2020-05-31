package aura

import "time"

// Collector is the interface implemented by anything that can be used by
// Reporter to report metrics. A Collector has to be registered for collection.
type Collector interface {
	// Interval sets the interval of a metric collecting period.
	Interval() time.Duration

	// Describe sends the super-set of all possible descriptors of metrics
	// collected by this Collector to the provided channel.
	Describe(chan<- *Desc)

	// Collect do the collecting stuffs and the implementation sends each
	// collected metric via the provided channel.
	Collect(ch chan<- Metric)
}
