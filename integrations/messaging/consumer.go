// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package messaging

import (
	"context"
	"strconv"

	"github.com/DataDog/data-streams-go/datastreams"
)

// TraceKafkaConsume extracts the pathway from to the kafka message header to the context.
// It returns the newly updated context which records the extracted pathway. Do not pass the resulting context from
// this function to another call of TraceKafkaConsume, as it will modify the pathway incorrectly.
func TraceKafkaConsume(ctx context.Context, msg ConsumerMessage, group string) context.Context {
	ctx = extractPipelineToContext(ctx, msg)
	edges := []string{"type:kafka", "direction:in", "group:" + group}
	if msg.GetTopic() != nil {
		edges = append(edges, "topic:"+*msg.GetTopic())
	}
	edges = append(edges, "partition:"+strconv.Itoa(int(msg.GetPartition())))
	_, ctx = datastreams.SetCheckpointWithPayloadSize(ctx, msg.GetSize(), edges...)
	return ctx
}

func extractPipelineToContext(ctx context.Context, m ConsumerMessage) context.Context {
	for _, header := range m.GetHeaders() {
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
