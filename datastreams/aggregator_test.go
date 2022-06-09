// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package datastreams

import (
	"sort"
	"testing"
	"time"

	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/DataDog/sketches-go/ddsketch/store"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func buildSketch(values ...float64) []byte {
	sketch := ddsketch.NewDDSketch(sketchMapping, store.DenseStoreConstructor(), store.DenseStoreConstructor())
	for _, v := range values {
		sketch.Add(v)
	}
	bytes, _ := proto.Marshal(sketch.ToProto())
	return bytes
}

func TestAggregator(t *testing.T) {
	p := newAggregator(nil, "env", "datacenter:us1.prod.dog", "service", "agent-addr", nil, "datadoghq.com", "key", true)
	tp1 := time.Now()
	tp2 := tp1.Add(time.Second * 40)
	p.add(statsPoint{
		edgeTags:       []string{"edge-1"},
		hash:           2,
		parentHash:     1,
		timestamp:      tp2.UnixNano(),
		pathwayLatency: time.Second.Nanoseconds(),
		edgeLatency:    time.Second.Nanoseconds(),
	})
	p.add(statsPoint{
		edgeTags:       []string{"edge-1"},
		hash:           2,
		parentHash:     1,
		timestamp:      tp2.UnixNano(),
		pathwayLatency: (5 * time.Second).Nanoseconds(),
		edgeLatency:    (2 * time.Second).Nanoseconds(),
	})
	p.add(statsPoint{
		edgeTags:       []string{"edge-1"},
		hash:           3,
		parentHash:     1,
		timestamp:      tp2.UnixNano(),
		pathwayLatency: (5 * time.Second).Nanoseconds(),
		edgeLatency:    (2 * time.Second).Nanoseconds(),
	})
	p.add(statsPoint{
		edgeTags:       []string{"edge-1"},
		hash:           2,
		parentHash:     1,
		timestamp:      tp1.UnixNano(),
		pathwayLatency: (5 * time.Second).Nanoseconds(),
		edgeLatency:    (2 * time.Second).Nanoseconds(),
	})
	// flush at tp2 doesn't flush tp2 (current bucket)
	assert.Equal(t, StatsPayload{
		Env:        "env",
		Service:    "service",
		PrimaryTag: "datacenter:us1.prod.dog",
		Stats: []StatsBucket{{
			Start:    uint64(alignTs(tp1.UnixNano(), bucketDuration.Nanoseconds())),
			Duration: uint64(bucketDuration.Nanoseconds()),
			Stats: []StatsPoint{{
				EdgeTags:       []string{"edge-1"},
				Hash:           2,
				ParentHash:     1,
				PathwayLatency: buildSketch(5),
				EdgeLatency:    buildSketch(2),
			}},
		}},
		TracerVersion: "v0.2",
		Lang:          "go",
	}, p.flush(tp2))
	sp := p.flush(tp2.Add(bucketDuration).Add(time.Second))
	sort.Slice(sp.Stats[0].Stats, func(i, j int) bool {
		return sp.Stats[0].Stats[i].Hash < sp.Stats[0].Stats[j].Hash
	})
	assert.Equal(t, StatsPayload{
		Env:        "env",
		Service:    "service",
		PrimaryTag: "datacenter:us1.prod.dog",
		Stats: []StatsBucket{{
			Start:    uint64(alignTs(tp2.UnixNano(), bucketDuration.Nanoseconds())),
			Duration: uint64(bucketDuration.Nanoseconds()),
			Stats: []StatsPoint{
				{
					EdgeTags:       []string{"edge-1"},
					Hash:           2,
					ParentHash:     1,
					PathwayLatency: buildSketch(1, 5),
					EdgeLatency:    buildSketch(1, 2),
				},
				{
					EdgeTags:       []string{"edge-1"},
					Hash:           3,
					ParentHash:     1,
					PathwayLatency: buildSketch(5),
					EdgeLatency:    buildSketch(2),
				},
			},
		}},
		TracerVersion: "v0.2",
		Lang:          "go",
	}, sp)
}
