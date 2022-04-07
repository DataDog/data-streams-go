// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package datastreams

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPathway(t *testing.T) {
	aggregator := aggregator{
		stopped:    1,
		in:         make(chan statsPoint, 10),
		service:    "service-1",
		env:        "env",
		primaryTag: "d:1",
	}
	setGlobalAggregator(&aggregator)
	defer setGlobalAggregator(nil)
	start := time.Now()
	middle := start.Add(time.Hour)
	end := middle.Add(time.Hour)
	p := newPathway(start)
	p = p.setCheckpoint(middle, []string{"edge-1"})
	p = p.setCheckpoint(end, []string{"edge-2"})
	hash1 := pathwayHash(nodeHash("service-1", "env", "d:1", nil), 0)
	hash2 := pathwayHash(nodeHash("service-1", "env", "d:1", []string{"edge-1"}), hash1)
	hash3 := pathwayHash(nodeHash("service-1", "env", "d:1", []string{"edge-2"}), hash2)
	assert.Equal(t, Pathway{
		hash:         hash3,
		pathwayStart: start,
		edgeStart:    end,
	}, p)
	assert.Equal(t, statsPoint{
		edgeTags:       nil,
		hash:           hash1,
		parentHash:     0,
		timestamp:      start.UnixNano(),
		pathwayLatency: 0,
		edgeLatency:    0,
	}, <-aggregator.in)
	assert.Equal(t, statsPoint{
		edgeTags:       []string{"edge-1"},
		hash:           hash2,
		parentHash:     hash1,
		timestamp:      middle.UnixNano(),
		pathwayLatency: middle.Sub(start).Nanoseconds(),
		edgeLatency:    middle.Sub(start).Nanoseconds(),
	}, <-aggregator.in)
	assert.Equal(t, statsPoint{
		edgeTags:       []string{"edge-2"},
		hash:           hash3,
		parentHash:     hash2,
		timestamp:      end.UnixNano(),
		pathwayLatency: end.Sub(start).Nanoseconds(),
		edgeLatency:    end.Sub(middle).Nanoseconds(),
	}, <-aggregator.in)
}
