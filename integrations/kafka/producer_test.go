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
	"github.com/stretchr/testify/assert"
)

func getHeaderValue(headers []kafka.Header, key string) []byte {
	for _, header := range headers {
		if header.Key == key {
			return header.Value
		}
	}
	return nil
}

func assertPathwayNotEqual(t *testing.T, p1 datastreams.Pathway, p2 datastreams.Pathway) {
	decodedP1, _ := datastreams.Decode(p1.Encode())
	decodedP2, _ := datastreams.Decode(p2.Encode())

	assert.NotEqual(t, decodedP1, decodedP2)
}

func assertPathwayEqual(t *testing.T, p1 datastreams.Pathway, p2 datastreams.Pathway) {
	decodedP1, _ := datastreams.Decode(p1.Encode())
	decodedP2, _ := datastreams.Decode(p2.Encode())

	assert.Equal(t, decodedP1, decodedP2)
}

func TestTraceKafkaProduce(t *testing.T) {
	t.Run("Checkpoint should be created and pathway should be propagated to kafka headers", func(t *testing.T) {
		ctx := context.Background()
		initialPathway := datastreams.NewPathway()
		ctx = datastreams.ContextWithPathway(ctx, initialPathway)

		msg := kafka.Message{
			TopicPartition: kafka.TopicPartition{},
			Value:          []byte{},
		}

		ok := TraceKafkaProduce(&ctx, &msg)
		// Operation should be successful.
		assert.Equal(t, ok, true)

		// The old pathway shouldn't be equal to the new pathway found in the ctx because we created a new checkpoint.
		ctxPathway, _ := datastreams.PathwayFromContext(ctx)
		assertPathwayNotEqual(t, initialPathway, ctxPathway)

		// The decoded pathway found in the kafka headers should be the same as the pathway found in the ctx.
		headersPathway, _ := datastreams.Decode(getHeaderValue(msg.Headers, datastreams.PropagationKey))
		assertPathwayEqual(t, ctxPathway, headersPathway)
	})
}