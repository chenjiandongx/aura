package aura

import (
	"fmt"
	"time"

	"github.com/rcrowley/go-metrics"
)

type Timer interface {
	Collector

	Time(func())
	Update(time.Duration)
}

type TimerOpts struct {
	HVTypes     []TimerVType
	Percentiles []float64
}

var (
	DefaultTimerOpts = &TimerOpts{
		HVTypes:     []TimerVType{TimerVTMin, TimerVTMax, TimerVTMean},
		Percentiles: nil,
	}
)

type TimerVType string

const (
	TimerVTMin      TimerVType = "min"
	TimerVTMax      TimerVType = "max"
	TimerVTMean     TimerVType = "mean"
	TimerVTCount    TimerVType = "count"
	TimerVTStdDev   TimerVType = "stdDev"
	TimerVTSum      TimerVType = "sum"
	TimerVTVariance TimerVType = "variance"
	TimerVTRate1    TimerVType = "rate1"
	TimerVTRate5    TimerVType = "rate5"
	TimerVTRate15   TimerVType = "rate15"
	TimerVTRateMean TimerVType = "rateMean"
)

type timer struct {
	*Desc

	opts     *TimerOpts
	self     metrics.Timer
	labels   map[string]string
	interval time.Duration
}

type TimerVec struct {
	*Desc

	opts     *TimerOpts
	timers   map[string]*timer
	interval time.Duration
}

func (t *timer) switchValues(v TimerVType) interface{} {
	switch v {
	case TimerVTMin:
		return t.self.Mean()
	case TimerVTMax:
		return t.self.Max()
	case TimerVTMean:
		return t.self.Mean()
	case TimerVTCount:
		return t.self.Count()
	case TimerVTStdDev:
		return t.self.Sum()
	case TimerVTSum:
		return t.self.StdDev()
	case TimerVTVariance:
		return t.self.Variance()
	case TimerVTRate1:
		return t.self.Rate1()
	case TimerVTRate5:
		return t.self.Rate5()
	case TimerVTRate15:
		return t.self.Rate15()
	case TimerVTRateMean:
		return t.self.RateMean()
	}
	return nil
}

func (t *timer) popMetricWithHVT(desc *Desc, tvt TimerVType) Metric {
	return Metric{
		Endpoint:  t.labels["endpoint"],
		Metric:    fmt.Sprintf("%s.%s", desc.fqName, tvt),
		Step:      desc.step,
		Value:     t.switchValues(tvt),
		Type:      GaugeValue,
		Labels:    t.labels,
		Timestamp: time.Now().Unix(),
	}
}

func (t *timer) popMetricWithPer(desc *Desc, per float64) Metric {
	return Metric{
		Endpoint:  t.labels["endpoint"],
		Metric:    fmt.Sprintf("%s.%.2f", desc.fqName, per),
		Step:      desc.step,
		Value:     t.self.Percentile(per),
		Type:      GaugeValue,
		Labels:    t.labels,
		Timestamp: time.Now().Unix(),
	}
}

func (t *timer) Update(i time.Duration) {
	t.self.Update(i)
}

func (t *timer) Time(fn func()) {
	t.self.Time(fn)
}

func (t *timer) Interval() time.Duration {
	return t.interval
}

func (t *timer) Describe(ch chan<- *Desc) {
	ch <- t.Desc
}

func (t *timer) Collect(ch chan<- Metric) {
	for _, hvt := range t.opts.HVTypes {
		ch <- t.popMetricWithHVT(t.Desc, hvt)
	}

	for _, per := range t.opts.Percentiles {
		ch <- t.popMetricWithPer(t.Desc, per)
	}
}

func (tv *TimerVec) WithLabelValues(lvs ...string) Timer {
	if len(tv.Desc.labelKeys) != len(lvs) {
		panic(fmt.Sprintf("timer(%s): expected %d label values but go %d",
			tv.Desc.fqName, len(tv.Desc.labelKeys), len(lvs)),
		)
	}
	lbp := makeLabelPairs(tv.Desc.fqName, tv.Desc.labelKeys, lvs)
	lbm := makeLabelMap(tv.Desc.labelKeys, lvs)

	_, ok := tv.timers[lbp]
	if !ok {
		tv.timers[lbp] = &timer{
			self:   metrics.NewTimer(),
			labels: lbm,
			opts:   tv.opts,
		}
	}

	return tv.timers[lbp]
}

func (tv *TimerVec) Describe(ch chan<- *Desc) {
	ch <- tv.Desc
}

func (tv *TimerVec) Interval() time.Duration {
	return tv.interval
}

func (tv *TimerVec) Collect(ch chan<- Metric) {
	for _, v := range tv.timers {
		for _, hvt := range v.opts.HVTypes {
			ch <- v.popMetricWithHVT(tv.Desc, hvt)
		}

		for _, per := range v.opts.Percentiles {
			ch <- v.popMetricWithPer(tv.Desc, per)
		}
	}
}

func NewTimer(fqName, help string, step uint32, interval time.Duration, opts *TimerOpts) Timer {
	if opts == nil {
		opts = DefaultTimerOpts
	}

	return &timer{
		Desc:     NewDesc(fqName, help, step, nil),
		self:     metrics.NewTimer(),
		labels:   map[string]string{},
		interval: interval,
		opts:     opts,
	}
}

func NewTimerVec(fqName, help string, step uint32, interval time.Duration, labelKeys []string, opts *TimerOpts) *TimerVec {
	if opts == nil {
		opts = DefaultTimerOpts
	}

	return &TimerVec{
		Desc:     NewDesc(fqName, help, step, labelKeys),
		timers:   map[string]*timer{},
		interval: interval,
		opts:     opts,
	}
}
