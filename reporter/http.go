package reporter

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/chenjiandongx/aura"
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

func ConvertMetric(m aura.Metric) MetricReported {
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
