// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package kafka

import (
	"context"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/DataDog/data-streams-go/datastreams"
)

// TraceKafkaProduce appends the pathway in the context to the kafka message header. It returns the
// newly updated context which records the updated pathway. Do not pass the resulting context from
// this function to another call of TraceKafkaProduce, as it will modify the pathway incorrectly.
func TraceKafkaProduce(ctx context.Context, msg *kafka.Message) context.Context {
	span, ctx := tracer.StartSpanFromContext(ctx, "dsm.trace_kafka_produce")
	edges := []string{"type:kafka", "direction:out"}
	if msg.TopicPartition.Topic != nil {
		edges = append(edges, "topic:"+*msg.TopicPartition.Topic)
	}
	if msg.TopicPartition.Partition != kafka.PartitionAny {
		edges = append(edges, "partition:"+strconv.Itoa(int(msg.TopicPartition.Partition)))
	}
	p, ctx := datastreams.SetCheckpoint(ctx, edges...)
	msg.Headers = append(msg.Headers, kafka.Header{Key: datastreams.PropagationKey, Value: p.Encode()})
	span.SetTag("pathway.hash", strconv.FormatUint(p.GetHash(), 10))
	span.Finish()
	return ctx
}
