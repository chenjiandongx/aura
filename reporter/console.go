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

func (r *StreamReporter) Report(ch chan aura.Metric) {
	for i := 0; i < r.MaxConcurrency; i++ {
		go func() {
			ms := make([]aura.Metric, 0)
			for {
				select {
				case metric := <-ch:
					if len(ms) >= r.Batch {
						if err := r.report(ms); err != nil {

						}
						ms = make([]aura.Metric, 0)
					}
					ms = append(ms, metric)
				case <-r.Ticker:
					if err := r.report(ms); err != nil {

					}
					ms = make([]aura.Metric, 0)
				}
			}
		}()
	}
}

func (r *StreamReporter) Convert(met aura.Metric) interface{} {
	return met
}

func (r *StreamReporter) report(mets []aura.Metric) error {
	for _, met := range mets {
		if _, err := fmt.Fprintf(r.Writer, "%+v\n", r.Convert(met)); err != nil {
			return err
		}
	}
	return nil
}
