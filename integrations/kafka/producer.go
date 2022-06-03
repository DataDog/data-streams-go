package kafka

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/DataDog/data-streams-go/datastreams"
)

// traceKafkaProduce appends the pathway in the context to the kafka message headers, and returns true if
// it is successful.
func traceKafkaProduce(ctx context.Context, msg kafka.Headers) bool {
	p, ok := datastreams.PathwayFromContext(ctx)
	if ok {
		msg.Headers = append(msg.Headers, kafka.Header{Key: datastreams.PropagationKey, Value: p.Encode()})
	}
	return ok
}
