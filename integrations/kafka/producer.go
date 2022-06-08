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

// TraceKafkaProduce appends the pathway in the context to the kafka message header. It returns the
// newly updated context which records the updated pathway. Do not pass the resulting context from
// this function to another call of TraceKafkaProduce, as it will modify the pathway incorrectly.
func TraceKafkaProduce(ctx context.Context, msg *kafka.Message) context.Context {
	p, ctx := datastreams.SetCheckpoint(ctx, "type:internal")
	msg.Headers = append(msg.Headers, kafka.Header{Key: datastreams.PropagationKey, Value: p.Encode()})
	return ctx
}
