package aura

import (
	"fmt"
	"strings"
)

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

type Desc struct {
	// fqName has been built from Namespace, Subsystem, and Name.
	fqName string
	// help provides some helpful information about this metric.
	help string

	labelKeys []string

	pairs []string

	tags map[string]string

	step uint32

	err error
}

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
