package reporter

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chenjiandongx/aura"
)

var DefaultStreamReporter = &StreamReporter{
	Writer:         os.Stdout,
	Batch:          200,
	Ticker:         time.Tick(5 * time.Second),
	MaxConcurrency: 3,
}

type StreamReporter struct {
	Writer         io.Writer
	Batch          int
	Ticker         <-chan time.Time
	MaxConcurrency int
}

func (s *StreamReporter) Report(ch chan aura.Metric) {
	for i := 0; i < s.MaxConcurrency; i++ {
		go func() {
			ms := make([]aura.Metric, 0)
			for {
				select {
				case metric := <-ch:
					if len(ms) >= s.Batch {
						s.report(ms)
						ms = make([]aura.Metric, 0)
					}
					ms = append(ms, metric)
				case <-s.Ticker:
					s.report(ms)
					ms = make([]aura.Metric, 0)
				}
			}
		}()
	}
}

func (s *StreamReporter) Convert(met aura.Metric) interface{} {
	return met
}

func (s *StreamReporter) report(mets []aura.Metric) error {
	for _, met := range mets {
		fmt.Fprintf(s.Writer, "%+v\n", s.Convert(met))
	}
	return nil
}
