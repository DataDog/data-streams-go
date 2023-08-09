package kafka

import (
	"context"

	"github.com/DataDog/data-streams-go/integrations/messaging"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type kafkaWrapper struct {
	*kafka.Message
}

func (m *kafkaWrapper) GetSize() int64 {
	return int64(len(m.Key) + len(m.Value))
}

var (
	_ messaging.ProducerMessage = (*kafkaWrapper)(nil)
	_ messaging.ConsumerMessage = (*kafkaWrapper)(nil)
)

func (m *kafkaWrapper) GetTopic() *string { return m.Message.TopicPartition.Topic }

func (m *kafkaWrapper) GetPartition() int32 {
	p := m.Message.TopicPartition.Partition
	if p == kafka.PartitionAny {
		return messaging.PartitionAny
	}
	return p
}

func (m *kafkaWrapper) AppendToHeaders(hs ...messaging.Header) {
	kafkaHeaders := make([]kafka.Header, len(hs))
	for i, h := range hs {
		kafkaHeaders[i] = kafka.Header{Key: h.Key, Value: h.Value}
	}
	m.Headers = append(m.Headers, kafkaHeaders...)
}

func (m *kafkaWrapper) GetHeaders() []messaging.Header {
	hs := make([]messaging.Header, len(m.Headers))
	for i, h := range m.Headers {
		hs[i] = messaging.Header{Key: h.Key, Value: h.Value}
	}
	return hs
}

func TraceKafkaProduce(ctx context.Context, msg *kafka.Message) context.Context {
	return messaging.TraceKafkaProduce(ctx, &kafkaWrapper{msg})
}

func TraceKafkaConsume(ctx context.Context, msg *kafka.Message, group string) context.Context {
	return messaging.TraceKafkaConsume(ctx, &kafkaWrapper{msg}, group)
}
