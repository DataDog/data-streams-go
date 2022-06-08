// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package datastreams

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	t.Run("SetCheckpoint", func(t *testing.T) {
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
		hash1 := pathwayHash(nodeHash("service-1", "env", "d:1", nil), 0)
		hash2 := pathwayHash(nodeHash("service-1", "env", "d:1", []string{"type:internal"}), hash1)

		ctx := context.Background()
		pathway := newPathway(start)

		ctx = ContextWithPathway(ctx, pathway)
		updatedPathway, _ := SetCheckpoint(ctx, "type:internal")

		statsPt1 := <-aggregator.in
		statsPt2 := <-aggregator.in

		assert.Equal(t, []string(nil), statsPt1.edgeTags)
		assert.Equal(t, hash1, statsPt1.hash)
		assert.Equal(t, uint64(0), statsPt1.parentHash)
		assert.Equal(t, start.UnixNano(), statsPt1.timestamp)

		assert.Equal(t, []string{"type:internal"}, statsPt2.edgeTags)
		assert.Equal(t, hash2, statsPt2.hash)
		assert.Equal(t, hash1, statsPt2.parentHash)

		assert.Equal(t, statsPt2.hash, updatedPathway.hash)
	})
}
