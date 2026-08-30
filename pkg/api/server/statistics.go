package server

import (
	"encoding/json"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/shtypes"
)

type (
	// StatisticParameter is one way a metric can be broken down —
	// which partition, or which interface.
	//
	// Exactly one field is populated per element, depending on the
	// metric. A metric that takes no parameters reports an empty
	// element, which arrives as "[]" rather than "{}" because PHP
	// serialises an empty map as a list.
	StatisticParameter struct {
		// Partition is a disk label, e.g. "<server>-disk" or
		// "<server>-swap".
		Partition string `json:"partition,omitempty"`

		// Iface is a network interface name, e.g. "e0".
		Iface string `json:"iface,omitempty"`
	}

	// StatisticTypes maps a metric name to the parameter sets it can
	// be requested with.
	//
	// It is a map on servers that expose metrics, and an empty list on
	// servers that do not — the same value expressed two ways, which
	// is why it needs its own decoder.
	StatisticTypes map[string][]StatisticParameter
)

// UnmarshalJSON tolerates the empty-list form.
//
// Declaring this field as a plain map fails on a server with no
// metrics ("cannot unmarshal array into Go value of type map"), and
// declaring it as a list fails on a server that has some. Before this
// it was []string, so [Client.ListStatisticTypes] worked only on
// servers that had nothing to report — which is to say it had never
// returned a metric name.
func (s *StatisticTypes) UnmarshalJSON(b []byte) error {
	if shtypes.IsEmptyMapShape(b) {
		*s = StatisticTypes{}
		return nil
	}
	var raw map[string][]StatisticParameter
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("server: decoding statistic types: %w", err)
	}
	*s = raw
	return nil
}

// UnmarshalJSON tolerates the empty-list form for a single parameter.
func (p *StatisticParameter) UnmarshalJSON(b []byte) error {
	if shtypes.IsEmptyMapShape(b) {
		*p = StatisticParameter{}
		return nil
	}
	type plain StatisticParameter
	var known plain
	if err := json.Unmarshal(b, &known); err != nil {
		return fmt.Errorf("server: decoding statistic parameter: %w", err)
	}
	*p = StatisticParameter(known)
	return nil
}

// Names lists the metric names, so a caller does not have to iterate
// the map to find out what is available.
func (s StatisticTypes) Names() []string {
	out := make([]string, 0, len(s))
	for name := range s {
		out = append(out, name)
	}
	return out
}
