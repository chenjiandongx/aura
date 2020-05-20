package aura

import (
	"bytes"
	"fmt"
	"time"
)

type ValueType string

const (
	CounterValue ValueType = "Counter"
	GaugeValue   ValueType = "Gauge"
)

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
