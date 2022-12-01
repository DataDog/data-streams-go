// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strconv"

	"github.com/DataDog/data-streams-go/datastreams"
)

// TraceKafkaConsume extracts the pathway from to the kafka message header to the context.
// It returns the newly updated context which records the extracted pathway. Do not pass the resulting context from
// this function to another call of TraceKafkaConsume, as it will modify the pathway incorrectly.
func TraceKafkaConsume(ctx context.Context, msg *kafka.Message, group string) context.Context {
	// tags need to be sorted, to ensure hash consistency across implementations
	ctx = extractPipelineToContext(ctx, msg)
	edges := []string{
		"direction:in",
		"group:" + group,
		"partition:" + strconv.Itoa(int(msg.TopicPartition.Partition)),
	}

	if msg.TopicPartition.Topic != nil {
		edges = append(edges, "topic:"+*msg.TopicPartition.Topic)
	}
	edges = append(edges, "type:kafka")

	_, ctx = datastreams.SetCheckpoint(ctx, edges...)
	return ctx
}

func extractPipelineToContext(ctx context.Context, m *kafka.Message) context.Context {
	for _, header := range m.Headers {
		if header.Key == datastreams.PropagationKey {
			p, err := datastreams.Decode(header.Value)
			if err != nil {
				return ctx
			}
			return datastreams.ContextWithPathway(ctx, p)
		}
	}
	return ctx
}
