package aura

import (
	"bytes"
	"fmt"
	"sort"
	"time"
)

type ValueType string

const (
	CounterValue ValueType = "Counter"
	GaugeValue   ValueType = "Gauge"
)

// todo: rethink tags
// tags: {status: 200, url: "/index"}
// tags: {status: 200, url: "/api"}
// tags: {status: 400, url: "/token"}
type Metric struct {
	Endpoint  string
	Metric    string
	Step      uint32
	Value     interface{}
	Type      ValueType
	Tags      map[string]string
	Timestamp int64
}

func makeLabelPairs(fqname string, keys []string, values []string) string {
	buf := &bytes.Buffer{}
	buf.WriteString(fqname)
	for idx, k := range keys {
		if idx == len(keys)-1 {
			buf.WriteString(fmt.Sprintf("%s=%s", k, values[idx]))
			continue
		}
		buf.WriteString(fmt.Sprintf("%s=%s,", k, values[idx]))
	}

	return buf.String()
}

func makeLabelMap(keys []string, values []string) map[string]string {
	m := make(map[string]string)
	for i, k := range keys {
		m[k] = values[i]
	}

	return m
}

type MetricReported struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Step      uint32      `json:"step"`
	Value     interface{} `json:"value"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}

func ConvertMetric(m Metric) MetricReported {
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

func NewConstMetric(desc *Desc, valueType ValueType, value interface{}, lvs ...string) (Metric, error) {
	if len(lvs) != len(desc.labelKeys) {
		return Metric{}, fmt.Errorf("%s: expected %d label values but got %d in %#v",
			desc.fqName, len(desc.labelKeys), len(lvs), lvs,
		)
	}

	tags := make(map[string]string)
	for i, lb := range desc.labelKeys {
		tags[lb] = lvs[i]
	}

	return Metric{
		Endpoint:  tags["endpoint"],
		Metric:    desc.fqName,
		Value:     value,
		Step:      desc.step,
		Type:      valueType,
		Tags:      tags,
		Timestamp: time.Now().Unix(),
	}, nil
}

func MustNewConstMetric(desc *Desc, valueType ValueType, value interface{}, lvs ...string) Metric {
	metric, err := NewConstMetric(desc, valueType, value, lvs...)
	if err != nil {
		panic(err)
	}
	return metric
}
