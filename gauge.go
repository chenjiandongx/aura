package aura

import (
	"fmt"
	"time"

	"github.com/rcrowley/go-metrics"
)

type Gauge interface {
	Collector

	Update(float64)
	Value() float64
}

type gauge struct {
	*Desc

	self     metrics.GaugeFloat64
	labels   map[string]string
	interval time.Duration
}

type GaugeVec struct {
	*Desc

	gauges   map[string]*gauge
	interval time.Duration
}

func (g *gauge) popMetric(desc *Desc) Metric {
	return Metric{
		Endpoint:  g.labels["endpoint"],
		Metric:    desc.fqName,
		Step:      desc.step,
		Value:     g.self.Value(),
		Type:      CounterValue,
		Labels:    g.labels,
		Timestamp: time.Now().Unix(),
	}
}

func (g *gauge) Update(i float64) {
	g.self.Update(i)
}

func (g *gauge) Value() float64 {
	return g.self.Value()
}

func (g *gauge) Interval() time.Duration {
	return g.interval
}

func (g *gauge) Describe(ch chan<- *Desc) {
	ch <- g.Desc
}

func (g *gauge) Collect(ch chan<- Metric) {
	ch <- g.popMetric(g.Desc)
}

func (gv *GaugeVec) WithLabelValues(lvs ...string) Gauge {
	if len(gv.Desc.labelKeys) != len(lvs) {
		panic(fmt.Sprintf("gauge(%s): expected %d label values but go %d",
			gv.Desc.fqName, len(gv.Desc.labelKeys), len(lvs)),
		)
	}
	lbp := makeLabelPairs(gv.Desc.fqName, gv.Desc.labelKeys, lvs)
	lbm := makeLabelMap(gv.Desc.labelKeys, lvs)

	_, ok := gv.gauges[lbp]
	if !ok {
		gv.gauges[lbp] = &gauge{self: metrics.NewGaugeFloat64(), labels: lbm}
	}

	return gv.gauges[lbp]
}

func (gv *GaugeVec) Describe(ch chan<- *Desc) {
	ch <- gv.Desc
}

func (gv *GaugeVec) Interval() time.Duration {
	return gv.interval
}

func (gv *GaugeVec) Collect(ch chan<- Metric) {
	for _, v := range gv.gauges {
		ch <- v.popMetric(gv.Desc)
	}
}

func NewGauge(fqName, help string, step uint32, interval time.Duration) Gauge {
	return &gauge{
		Desc:     NewDesc(fqName, help, step, nil),
		self:     metrics.NewGaugeFloat64(),
		labels:   map[string]string{},
		interval: interval,
	}
}

func NewGaugeVec(fqName, help string, step uint32, interval time.Duration, labelKeys []string) *GaugeVec {
	return &GaugeVec{
		Desc:     NewDesc(fqName, help, step, labelKeys),
		gauges:   map[string]*gauge{},
		interval: interval,
	}
}
