package aura

import (
	"fmt"
	"time"

	"github.com/rcrowley/go-metrics"
)

type Histogram interface {
	Collector

	Observe(int64)
}

type HistogramOpts struct {
	HVTypes     []HistogramVType
	Percentiles []float64
}

var (
	defaultSample        = metrics.NewExpDecaySample(1028, 0.015)
	DefaultHistogramOpts = &HistogramOpts{
		HVTypes:     []HistogramVType{HVTMin, HVTMax, HVTMean},
		Percentiles: nil,
	}
)

type HistogramVType string

const (
	HVTMin      HistogramVType = "min"
	HVTMax      HistogramVType = "max"
	HVTMean     HistogramVType = "mean"
	HVTCount    HistogramVType = "count"
	HVTStdDev   HistogramVType = "stddev"
	HVTSum      HistogramVType = "sum"
	HVTVariance HistogramVType = "variance"
)

type histogram struct {
	*Desc

	opts     *HistogramOpts
	self     metrics.Histogram
	labels   map[string]string
	interval time.Duration
}

type HistogramVec struct {
	*Desc

	opts       *HistogramOpts
	histograms map[string]*histogram
	interval   time.Duration
}

func (h *histogram) switchValues(v HistogramVType) interface{} {
	switch v {
	case HVTMin:
		return h.self.Mean()
	case HVTMax:
		return h.self.Max()
	case HVTMean:
		return h.self.Mean()
	case HVTCount:
		return h.self.Count()
	case HVTSum:
		return h.self.Sum()
	case HVTStdDev:
		return h.self.StdDev()
	case HVTVariance:
		return h.self.Variance()
	}
	return nil
}

func (h *histogram) Observe(i int64) {
	h.self.Update(i)
}

func (h *histogram) Interval() time.Duration {
	return h.interval
}

func (h *histogram) Describe(ch chan<- *Desc) {
	ch <- h.Desc
}

func (h *histogram) popMetricWithHVT(hvt HistogramVType) Metric {
	return Metric{
		Endpoint:  h.labels["endpoint"],
		Metric:    fmt.Sprintf("%s.%s", h.Desc.fqName, hvt),
		Step:      h.Desc.step,
		Value:     h.switchValues(hvt),
		Type:      GaugeValue,
		Tags:      h.labels,
		Timestamp: time.Now().Unix(),
	}
}

func (h *histogram) popMetricWithPer(per float64) Metric {
	return Metric{
		Endpoint:  h.labels["endpoint"],
		Metric:    fmt.Sprintf("%s.%.2f", h.Desc.fqName, per),
		Step:      h.Desc.step,
		Value:     h.self.Percentile(per),
		Type:      GaugeValue,
		Tags:      h.labels,
		Timestamp: time.Now().Unix(),
	}
}

func (h *histogram) Collect(ch chan<- Metric) {
	for _, hvt := range h.opts.HVTypes {
		ch <- h.popMetricWithHVT(hvt)
	}

	for _, per := range h.opts.Percentiles {
		ch <- h.popMetricWithPer(per)
	}
}

func (hv *HistogramVec) WithLabelValues(lvs ...string) Histogram {
	if len(hv.Desc.labelKeys) != len(lvs) {
		// todo: panic message
		panic("")
	}
	lbp := makeLabelPairs(hv.Desc.fqName, hv.Desc.labelKeys, lvs)
	lbm := makeLabelMap(hv.Desc.labelKeys, lvs)

	_, ok := hv.histograms[lbp]
	if !ok {
		hv.histograms[lbp] = &histogram{
			self:   metrics.NewHistogram(defaultSample),
			labels: lbm,
		}
	}

	return hv.histograms[lbp]
}

func (hv *HistogramVec) Describe(ch chan<- *Desc) {
	ch <- hv.Desc
}

func (hv *HistogramVec) Interval() time.Duration {
	return hv.interval
}

func (hv *HistogramVec) Collect(ch chan<- Metric) {
	for _, v := range hv.histograms {
		for _, hvt := range v.opts.HVTypes {
			ch <- v.popMetricWithHVT(hvt)
		}

		for _, per := range v.opts.Percentiles {
			ch <- v.popMetricWithPer(per)
		}
	}
}

func NewHistogram(fqName, help string, step uint32, interval time.Duration, opts *HistogramOpts) Histogram {
	if opts == nil {
		opts = DefaultHistogramOpts
	}

	return &histogram{
		Desc:     NewDesc(fqName, help, step, nil),
		self:     metrics.NewHistogram(defaultSample),
		labels:   map[string]string{},
		interval: interval,
		opts:     opts,
	}
}

func NewHistogramVec(fqName, help string, step uint32, interval time.Duration, labelKeys []string, opts *HistogramOpts) *HistogramVec {
	if opts == nil {
		opts = DefaultHistogramOpts
	}

	return &HistogramVec{
		Desc:       NewDesc(fqName, help, step, labelKeys),
		histograms: map[string]*histogram{},
		interval:   interval,
		opts:       opts,
	}
}
