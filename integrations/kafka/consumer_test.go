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

func TestTraceKafkaConsume(t *testing.T) {
	t.Run("Checkpoint should be created and pathway should be extracted from kafka headers into context", func(t *testing.T) {
		// First, set up pathway and context as it would have been from the producer view.
		producerCtx := context.Background()
		initialPathway := datastreams.NewPathway()
		producerCtx = datastreams.ContextWithPathway(producerCtx, initialPathway)

		topic := "my-topic"
		msg := kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic: &topic,
			},
			Value: []byte{},
		}
		producerCtx = TraceKafkaProduce(producerCtx, &msg)

		// Calls TraceKafkaConsume
		group := "my-consumer-group"
		consumerCtx := context.Background()
		consumerCtx = TraceKafkaConsume(consumerCtx, &msg, group)

		// Check that the resulting consumerCtx contains an expected pathway.
		consumerCtxPathway, _ := datastreams.PathwayFromContext(consumerCtx)
		_, expectedCtx := datastreams.SetCheckpoint(producerCtx, "type:kafka", "group:my-consumer-group", "topic:my-topic")
		expectedCtxPathway, _ := datastreams.PathwayFromContext(expectedCtx)
		assertPathwayEqual(t, expectedCtxPathway, consumerCtxPathway)
	})
}
