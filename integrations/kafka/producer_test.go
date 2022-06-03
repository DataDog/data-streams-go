// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package kafka

import (
	"context"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
)

func testTraceKafkaProduce(t *testing.T) {
	t.Run("Context and kafka headers are set appropriately", function(t *testing.T) {
		ctx := context.Context()
		msg := kafka.Headers()
		ok = ddKafka.traceKafkaProduce(ctx, msg)
		assert.Equal(ok, true)
		assert.Equal(ctx, context.Context())
		assert.Equal(msg, kafka.Headers())
	})
}
