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

// TraceKafkaProduce appends the pathway in the context to the kafka message header. It returns the
// newly updated context which records the updated pathway. Do not pass the resulting context from
// this function to another call of TraceKafkaProduce, as it will modify the pathway incorrectly.
func TraceKafkaProduce(ctx context.Context, msg *kafka.Message) context.Context {
	// tags need to be sorted, to ensure hash consistency across implementations
	edges := []string{"direction:out"}
	if msg.TopicPartition.Partition != kafka.PartitionAny {
		edges = append(edges, "partition:"+strconv.Itoa(int(msg.TopicPartition.Partition)))
	}
	if msg.TopicPartition.Topic != nil {
		edges = append(edges, "topic:"+*msg.TopicPartition.Topic)
	}
	edges = append(edges, "type:kafka")

	p, ctx := datastreams.SetCheckpoint(ctx, edges...)
	msg.Headers = append(msg.Headers, kafka.Header{Key: datastreams.PropagationKey, Value: p.Encode()})
	return ctx
}
