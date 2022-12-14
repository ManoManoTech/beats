// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Code generated by x-pack/dev-tools/cmd/buildspec/buildspec.go - DO NOT EDIT.

package program

import (
	"strings"

	"github.com/elastic/beats/v7/x-pack/elastic-agent/pkg/packer"
)

var Supported []Spec
var SupportedMap map[string]Spec

func init() {
	// Packed Files
	// spec/apm-server.yml
	// spec/endpoint.yml
	// spec/filebeat.yml
	// spec/fleet-server.yml
	// spec/heartbeat.yml
	// spec/metricbeat.yml
	// spec/packetbeat.yml
	unpacked := packer.MustUnpack("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")
	SupportedMap = make(map[string]Spec)

	for f, v := range unpacked {
		s, err := NewSpecFromBytes(v)
		if err != nil {
			panic("Cannot read spec from " + f)
		}
		Supported = append(Supported, s)
		SupportedMap[strings.ToLower(s.Cmd)] = s
	}
}
