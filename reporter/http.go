package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/go-resty/resty/v2"
)

type MetricReported struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Step      uint32      `json:"step"`
	Value     interface{} `json:"value"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}

var defaultHTTPClient *resty.Client

func init() {
	defaultHTTPClient = resty.New()
	defaultHTTPClient.SetTimeout(5 * time.Second)
	defaultHTTPClient.SetRetryCount(3)
}

var DefaultHTTPReporter = &HTTPReporter{
	client:         defaultHTTPClient,
	Urls:           []string{},
	Batch:          200,
	Ticker:         time.Tick(5 * time.Second),
	Timeout:        5 * time.Second,
	RetryCount:     3,
	MaxConcurrency: 3,
}

type HTTPReporter struct {
	client         *resty.Client
	Urls           []string
	Batch          int
	Ticker         <-chan time.Time
	Timeout        time.Duration
	RetryCount     int
	MaxConcurrency int
}

func (r *HTTPReporter) Convert(m aura.Metric) interface{} {
	keys := make([]string, 0)
	for k := range m.Tags {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	buf := &bytes.Buffer{}
	for idx, k := range keys {
		if idx == len(keys)-1 {
			buf.WriteString(fmt.Sprintf("%s=%s", k, m.Tags[k]))
			continue
		}
		buf.WriteString(fmt.Sprintf("%s=%s,", k, m.Tags[k]))
	}
	return MetricReported{
		Endpoint:  m.Endpoint,
		Metric:    m.Metric,
		Step:      m.Step,
		Value:     m.Value,
		Type:      string(m.Type),
		Tags:      buf.String(),
		Timestamp: m.Timestamp,
	}
}

func (r *HTTPReporter) Report(ch chan aura.Metric) {
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

func (r *HTTPReporter) report(mets []aura.Metric) error {
	items := make([]interface{}, 0)
	for _, met := range mets {
		items = append(items, r.Convert(met))
	}

	bs, err := json.Marshal(items)
	if err != nil {
		return err
	}

	ok := false

	indexes := rand.Perm(len(r.Urls))
	for _, i := range indexes {
		_, err := r.client.R().SetBody(bs).Post(r.Urls[i])
		if err != nil {
			continue
		}
		ok = true
	}

	if !ok {
		// todo
		return fmt.Errorf("")
	}

	return nil
}
