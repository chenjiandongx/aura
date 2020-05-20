package aura

import "time"

type Collector interface {
	Interval() time.Duration
	Describe(chan<- *Desc)
	Collect(ch chan<- Metric)
}
