// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package datastreams

import (
	"hash/fnv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPathway(t *testing.T) {
	t.Run("test SetCheckpoint", func(t *testing.T) {
		aggregator := aggregator{
			stopped:    1,
			in:         newFastQueue(),
			service:    "service-1",
			env:        "env",
			primaryTag: "d:1",
		}
		setGlobalAggregator(&aggregator)
		defer setGlobalAggregator(nil)
		start := time.Now()
		middle := start.Add(time.Hour)
		end := middle.Add(time.Hour)
		p := newPathway(start, NewCheckpointParams())
		p = p.setCheckpoint(middle, NewCheckpointParams(), []string{"edge-1"})
		p = p.setCheckpoint(end, NewCheckpointParams(), []string{"edge-2"})
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
		}, aggregator.in.pop())
		assert.Equal(t, statsPoint{
			edgeTags:       []string{"edge-1"},
			hash:           hash2,
			parentHash:     hash1,
			timestamp:      middle.UnixNano(),
			pathwayLatency: middle.Sub(start).Nanoseconds(),
			edgeLatency:    middle.Sub(start).Nanoseconds(),
		}, aggregator.in.pop())
		assert.Equal(t, statsPoint{
			edgeTags:       []string{"edge-2"},
			hash:           hash3,
			parentHash:     hash2,
			timestamp:      end.UnixNano(),
			pathwayLatency: end.Sub(start).Nanoseconds(),
			edgeLatency:    end.Sub(middle).Nanoseconds(),
		}, aggregator.in.pop())
	})

	t.Run("test NewPathway", func(t *testing.T) {
		aggregator := aggregator{
			stopped:    1,
			in:         newFastQueue(),
			service:    "service-1",
			env:        "env",
			primaryTag: "d:1",
		}
		setGlobalAggregator(&aggregator)
		defer setGlobalAggregator(nil)

		pathwayWithNoEdgeTags := NewPathway()
		pathwayWith1EdgeTag := NewPathway("type:internal")
		pathwayWith2EdgeTags := NewPathway("type:internal", "some_other_key:some_other_val")

		hash1 := pathwayHash(nodeHash("service-1", "env", "d:1", nil), 0)
		hash2 := pathwayHash(nodeHash("service-1", "env", "d:1", []string{"type:internal"}), 0)
		hash3 := pathwayHash(nodeHash("service-1", "env", "d:1", []string{"type:internal", "some_other_key:some_other_val"}), 0)
		assert.Equal(t, hash1, pathwayWithNoEdgeTags.hash)
		assert.Equal(t, hash2, pathwayWith1EdgeTag.hash)
		assert.Equal(t, hash3, pathwayWith2EdgeTags.hash)

		var statsPointWithNoEdgeTags statsPoint = *aggregator.in.pop()
		var statsPointWith1EdgeTag statsPoint = *aggregator.in.pop()
		var statsPointWith2EdgeTags statsPoint = *aggregator.in.pop()
		assert.Equal(t, hash1, statsPointWithNoEdgeTags.hash)
		assert.Equal(t, []string(nil), statsPointWithNoEdgeTags.edgeTags)
		assert.Equal(t, hash2, statsPointWith1EdgeTag.hash)
		assert.Equal(t, []string{"type:internal"}, statsPointWith1EdgeTag.edgeTags)
		assert.Equal(t, hash3, statsPointWith2EdgeTags.hash)
		assert.Equal(t, []string{"some_other_key:some_other_val", "type:internal"}, statsPointWith2EdgeTags.edgeTags)
	})

	t.Run("test nodeHash", func(t *testing.T) {
		assert.NotEqual(t,
			nodeHash("service-1", "env", "d:1", []string{"type:internal"}),
			nodeHash("service-1", "env", "d:1", []string{"type:kafka"}),
		)
		assert.NotEqual(t,
			nodeHash("service-1", "env", "d:1", []string{"exchange:1"}),
			nodeHash("service-1", "env", "d:1", []string{"exchange:2"}),
		)
		assert.NotEqual(t,
			nodeHash("service-1", "env", "d:1", []string{"topic:1"}),
			nodeHash("service-1", "env", "d:1", []string{"topic:2"}),
		)
		assert.NotEqual(t,
			nodeHash("service-1", "env", "d:1", []string{"group:1"}),
			nodeHash("service-1", "env", "d:1", []string{"group:2"}),
		)
		assert.NotEqual(t,
			nodeHash("service-1", "env", "d:1", []string{"event_type:1"}),
			nodeHash("service-1", "env", "d:1", []string{"event_type:2"}),
		)
		assert.Equal(t,
			nodeHash("service-1", "env", "d:1", []string{"partition:0"}),
			nodeHash("service-1", "env", "d:1", []string{"partition:1"}),
		)
	})

	t.Run("test isWellFormedEdgeTag", func(t *testing.T) {
		for _, tc := range []struct {
			s string
			b bool
		}{
			{"", false},
			{"dog", false},
			{"dog:", false},
			{"dog:bark", false},
			{"type:", true},
			{"type:dog", true},
			{"type::dog", false},
			{"type:d:o:g", false},
			{"type::", false},
			{":", false},
		} {
			assert.Equal(t, isWellFormedEdgeTag(tc.s), tc.b)
		}
	})

	// nodeHash assumes that the go Hash interface produces the same result
	// for a given series of Write calls as for a single Write of the same
	// byte sequence. This unit test asserts that assumption.
	t.Run("test hashWriterIsomorphism", func(t *testing.T) {
		h := fnv.New64()
		var b []byte
		b = append(b, "dog"...)
		b = append(b, "cat"...)
		b = append(b, "pig"...)
		h.Write(b)
		s1 := h.Sum64()
		h.Reset()
		h.Write([]byte("dog"))
		h.Write([]byte("cat"))
		h.Write([]byte("pig"))
		assert.Equal(t, s1, h.Sum64())
	})

	t.Run("test GetHash", func(t *testing.T) {
		pathway := NewPathway("type:kafka", "topic:my-topic", "direction:in")
		assert.Equal(t, pathway.hash, pathway.GetHash())
	})
}

// Sample results at time of writing this benchmark:
// goos: darwin
// goarch: amd64
// pkg: github.com/DataDog/data-streams-go/datastreams
// cpu: Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz
// BenchmarkNodeHash-8   	 5167707	       232.5 ns/op	      24 B/op	       1 allocs/op
func BenchmarkNodeHash(b *testing.B) {
	service := "benchmark-runner"
	env := "test"
	primaryTag := "foo:bar"
	edgeTags := []string{"event_type:dog", "exchange:local", "group:all", "topic:off", "type:writer"}
	for i := 0; i < b.N; i++ {
		nodeHash(service, env, primaryTag, edgeTags)
	}
}
