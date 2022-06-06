// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package kafka

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/DataDog/data-streams-go/datastreams"
)

// TraceKafkaProduce appends the pathway in the context to the kafka message header, and returns true if
// it is successful.
func TraceKafkaProduce(ctx context.Context, msg kafka.Message, edgeTags ...string) (bool, context.Context, kafka.Message) {
	_, ctx = datastreams.SetCheckpoint(ctx, edgeTags...)
	p, ok := datastreams.PathwayFromContext(ctx)
	if ok {
		msg.Headers = append(msg.Headers, kafka.Header{Key: datastreams.PropagationKey, Value: p.Encode()})
	}
	return ok, ctx, msg
}
