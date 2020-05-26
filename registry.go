package aura

import (
	"fmt"
	"sync"
	"time"
)

const (
	// capacity for the channel to collect metrics and descriptors.
	defaultCapMetricChan = 2500
	defaultCapDescChan   = 20
)

type Reporter interface {
	Convert(Metric) interface{}
	Report(ch chan Metric)
}

type MetaData struct {
	Metric string `json:"metric"`
	Help   string `json:"help"`
	Step   uint32 `json:"step"`
}

type Registry struct {
	opts       *RegistryOpts
	reporter   Reporter
	mtx        sync.RWMutex
	collectors map[string]Collector
	metricChs  chan Metric
	metadata   map[string]*MetaData
	stop       chan struct{}
	exit       chan struct{}
}

type RegistryOpts struct {
	CapMetricChan int
	CapDescChan   int
}

var DefaultRegistryOpts = &RegistryOpts{
	CapMetricChan: defaultCapMetricChan,
	CapDescChan:   defaultCapDescChan,
}

func NewRegistry(opts *RegistryOpts) *Registry {
	if opts == nil {
		opts = DefaultRegistryOpts
	}

	if opts.CapDescChan < 1 {
		opts.CapDescChan = defaultCapDescChan
	}
	if opts.CapMetricChan < 1 {
		opts.CapMetricChan = defaultCapMetricChan
	}

	return &Registry{
		opts:       opts,
		reporter:   nil,
		mtx:        sync.RWMutex{},
		collectors: map[string]Collector{},
		metricChs:  make(chan Metric, opts.CapMetricChan),
		metadata:   map[string]*MetaData{},
		stop:       make(chan struct{}),
		exit:       make(chan struct{}),
	}
}

func (r *Registry) AddReporter(reporter Reporter) {
	r.reporter = reporter
}

func (r *Registry) Register(c Collector) error {
	descChan := make(chan *Desc, r.opts.CapDescChan)

	go func() {
		c.Describe(descChan)
		close(descChan)
	}()

	r.mtx.Lock()
	defer func() {
		// Drain channel in case of premature return to not leak a goroutine.
		for range descChan {
		}
		r.mtx.Unlock()
	}()

	for desc := range descChan {
		if desc.err != nil {
			return desc.err
		}

		if _, ok := r.metadata[desc.fqName]; ok {
			return fmt.Errorf("duplicated mertric FQName:(%s)", desc.fqName)
		}

		r.metadata[desc.fqName] = &MetaData{
			Metric: desc.fqName,
			Help:   desc.help,
			Step:   desc.step,
		}
		r.collectors[desc.fqName] = c
	}

	return nil
}

func (r *Registry) MustRegister(cs ...Collector) {
	for _, c := range cs {
		if err := r.Register(c); err != nil {
			panic(err)
		}
	}
}

func (r *Registry) gather() {
	for _, collector := range r.collectors {
		go func(c Collector) {
			ticker := time.Tick(c.Interval())

			time.Sleep(5 * time.Second)
			c.Collect(r.metricChs)

			for {
				select {
				case <-ticker:
					c.Collect(r.metricChs)
				case <-r.stop:
					return
				}
			}
		}(collector)
	}
}

func (r *Registry) Run() {
	if r.reporter == nil {
		panic("reporter cannot be nil")
	}
	r.gather()
	r.reporter.Report(r.metricChs)
	<-r.exit
}

func (r *Registry) Stop() {
	for i := 0; i < len(r.collectors); i++ {
		r.stop <- struct{}{}
	}
	r.exit <- struct{}{}
}
