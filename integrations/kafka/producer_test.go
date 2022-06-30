// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package kafka

import (
	"context"
	"testing"

	"github.com/DataDog/data-streams-go/datastreams"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func TestTraceKafkaProduce(t *testing.T) {
	t.Run("Checkpoint should be created and pathway should be propagated to kafka headers", func(t *testing.T) {
		ctx := context.Background()
		initialPathway := datastreams.NewPathway()
		ctx = datastreams.ContextWithPathway(ctx, initialPathway)

		msg := kafka.Message{
			TopicPartition: kafka.TopicPartition{},
			Value:          []byte{},
		}

		ctx = TraceKafkaProduce(ctx, &msg)

		// The old pathway shouldn't be equal to the new pathway found in the ctx because we created a new checkpoint.
		ctxPathway, _ := datastreams.PathwayFromContext(ctx)
		assertPathwayNotEqual(t, initialPathway, ctxPathway)

		// The decoded pathway found in the kafka headers should be the same as the pathway found in the ctx.
		var encodedPathway []byte
		for _, header := range msg.Headers {
			if header.Key == datastreams.PropagationKey {
				encodedPathway = header.Value
			}
		}
		headersPathway, _ := datastreams.Decode(encodedPathway)
		assertPathwayEqual(t, ctxPathway, headersPathway)
	})
}
