package aura

import (
	"fmt"
	"time"

	"github.com/rcrowley/go-metrics"
)

// Counter is just a gauge for an AtomicLong instance. You can increment or decrement its value.
type Counter interface {
	Collector

	Clear()
	Count() int64
	Rate() float64
	Dec(int64)
	Inc(int64)
}

type counter struct {
	*Desc

	prev     int64
	self     metrics.Counter
	labels   map[string]string
	interval time.Duration
}

type CounterVec struct {
	*Desc

	counters map[string]*counter
	interval time.Duration
}

func (c *counter) popMetric(desc *Desc) Metric {
	cnt := c.self.Count()

	m := Metric{
		Endpoint:  c.labels["endpoint"],
		Metric:    desc.fqName,
		Step:      desc.step,
		Value:     c.Rate(),
		Type:      GaugeValue,
		Labels:    c.labels,
		Timestamp: time.Now().Unix(),
	}

	c.prev = cnt
	return m
}

// Inc increases the counter.
func (c *counter) Inc(i int64) {
	c.self.Inc(i)
}

// Dec decreases the counter.
func (c *counter) Dec(i int64) {
	c.self.Dec(i)
}

// Clear resets the counter to zero.
func (c *counter) Clear() {
	c.self.Clear()
}

// Count returns the number of the counter.
func (c *counter) Count() int64 {
	return c.self.Count()
}

// Rate returns the increasing rate of the counter
func (c *counter) Rate() float64 {
	return float64(c.self.Count()-c.prev) / float64(c.Desc.step)
}

// Interval implements aura.Collector.
func (c *counter) Interval() time.Duration {
	return c.interval
}

// Describe implements aura.Collector.
func (c *counter) Describe(ch chan<- *Desc) {
	ch <- c.Desc
}

// Collect implements aura.Collector.
func (c *counter) Collect(ch chan<- Metric) {
	ch <- c.popMetric(c.Desc)
}

func (cv *CounterVec) WithLabelValues(lvs ...string) Counter {
	if len(cv.Desc.labelKeys) != len(lvs) {
		panic(fmt.Sprintf("counter(%s): expected %d label values but go %d",
			cv.Desc.fqName, len(cv.Desc.labelKeys), len(lvs)),
		)
	}

	return cv.searchCounter(lvs...)
}

func (cv *CounterVec) With(labels map[string]string) Counter {
	for k := range labels {
		if !cv.Desc.IsKeyIn(k) {
			panic(fmt.Sprintf("counter(%s): expected label key: %s, but it dosen't exists", cv.Desc.fqName, k))
		}
	}

	lvs := make([]string, 0)
	for _, key := range cv.Desc.labelKeys {
		lvs = append(lvs, labels[key])
	}

	return cv.searchCounter(lvs...)
}

func (cv *CounterVec) searchCounter(lvs ...string) Counter {
	lbp := makeLabelPairs(cv.Desc.fqName, cv.Desc.labelKeys, lvs)
	lbm := makeLabelMap(cv.Desc.labelKeys, lvs)
	_, ok := cv.counters[lbp]
	if !ok {
		cv.counters[lbp] = &counter{self: metrics.NewCounter(), labels: lbm, Desc: &Desc{step: cv.step}}
	}

	return cv.counters[lbp]
}

// Interval implements aura.Collector.
func (cv *CounterVec) Interval() time.Duration {
	return cv.interval
}

// Describe implements aura.Collector.
func (cv *CounterVec) Describe(ch chan<- *Desc) {
	ch <- cv.Desc
}

// Collect implements aura.Collector.
func (cv *CounterVec) Collect(ch chan<- Metric) {
	for _, c := range cv.counters {
		ch <- c.popMetric(cv.Desc)
	}
}

func NewCounter(fqName, help string, step uint32, interval time.Duration) Counter {
	return &counter{
		Desc:     NewDesc(fqName, help, step, nil),
		self:     metrics.NewCounter(),
		labels:   map[string]string{},
		interval: interval,
	}
}

func NewCounterVec(fqName, help string, step uint32, interval time.Duration, labelKeys []string) *CounterVec {
	return &CounterVec{
		Desc:     NewDesc(fqName, help, step, labelKeys),
		counters: map[string]*counter{},
		interval: interval,
	}
}
