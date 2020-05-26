package aura

import (
	"fmt"
	"strings"
)

// BuildFQName joins the given three name components by ".". Empty name components are ignored.
func BuildFQName(namespace, subsystem, name string) string {
	if name == "" {
		return ""
	}

	switch {
	case namespace != "" && subsystem != "":
		return strings.Join([]string{namespace, subsystem, name}, ".")
	case namespace != "":
		return strings.Join([]string{namespace, name}, ".")
	case subsystem != "":
		return strings.Join([]string{subsystem, name}, ".")
	}
	return name
}

// Desc is the descriptor used by every Metric.
type Desc struct {
	// fqName has been built from Namespace, Subsystem, and Name.
	fqName string
	// help provides some helpful information about this metric.
	help string
	// labelKeys contains the keys of label
	labelKeys []string
	// step is the reporting interval of a metric
	step uint32
	// err is an error that occurred during construction.
	err error
}

// IsKeyIn returns true if the key is in the labels.
func (d *Desc) IsKeyIn(k string) bool {
	for _, key := range d.labelKeys {
		if key == k {
			return true
		}
	}
	return false
}

// NewDesc allocates and initializes a new Desc. Errors are recorded in the Desc
// and will be reported on registration time.
func NewDesc(fqName, help string, step uint32, labelKeys []string) *Desc {
	d := &Desc{help: help, labelKeys: labelKeys}
	if fqName == "" {
		d.err = fmt.Errorf("fqname should not be empty")
		return d
	}
	d.fqName = fqName

	if step == 0 {
		d.err = fmt.Errorf("step should greater than 0")
		return d
	}

	d.step = step
	return d
}
