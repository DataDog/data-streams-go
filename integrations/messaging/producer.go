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

// TraceKafkaProduce appends the pathway in the context to the kafka message header. It returns the
// newly updated context which records the updated pathway. Do not pass the resulting context from
// this function to another call of TraceKafkaProduce, as it will modify the pathway incorrectly.
func TraceKafkaProduce(ctx context.Context, msg ProducerMessage) context.Context {
	edges := []string{"type:kafka", "direction:out"}
	if msg.GetTopic() != nil {
		edges = append(edges, "topic:"+*msg.GetTopic())
	}
	if msg.GetPartition() != PartitionAny {
		edges = append(edges, "partition:"+strconv.Itoa(int(msg.GetPartition())))
	}
	p, ctx := datastreams.SetCheckpointWithParams(ctx, datastreams.NewCheckpointParams().WithPayloadSize(msg.GetSize()), edges...)
	msg.AppendToHeaders(Header{Key: datastreams.PropagationKey, Value: p.Encode()})
	return ctx
}
