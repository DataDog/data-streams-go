package sarama

import (
	"context"

	"github.com/DataDog/data-streams-go/integrations/messaging"
	"github.com/Shopify/sarama"
)

// Internal wrapper struct used for defining ProducerMessage interface
// on top of sarama's message format.
type saramaProducerWrapper struct {
	*sarama.ProducerMessage
}

// Internal wrapper struct used for defining ConsumerMessage interface
// on top of sarama's message format.
type saramaConsumerWrapper struct {
	*sarama.ConsumerMessage
}

var (
	_ messaging.ProducerMessage = (*saramaProducerWrapper)(nil)
	_ messaging.ConsumerMessage = (*saramaConsumerWrapper)(nil)
)

func (m *saramaProducerWrapper) GetTopic() *string { return &m.ProducerMessage.Topic }

// Returns the partition to which this message is produced to.
// Guaranteed to be defined only if message is delivered successfully.
// sarama doesn't provide a way to know if this is present or not since
// zero value might represent both "not known" or partition zero.
func (m *saramaProducerWrapper) GetPartition() int32 { return m.ProducerMessage.Partition }

func (m *saramaProducerWrapper) AppendToHeaders(hs ...messaging.Header) {
	kafkaHeaders := make([]sarama.RecordHeader, len(hs))
	for i, h := range hs {
		kafkaHeaders[i] = sarama.RecordHeader{Key: []byte(h.Key), Value: h.Value}
	}
	m.Headers = append(m.Headers, kafkaHeaders...)
}

func (m *saramaConsumerWrapper) GetTopic() *string { return &m.ConsumerMessage.Topic }

func (m *saramaConsumerWrapper) GetPartition() int32 { return m.ConsumerMessage.Partition }

func (m *saramaConsumerWrapper) GetHeaders() []messaging.Header {
	hs := make([]messaging.Header, len(m.Headers))
	for i, h := range m.Headers {
		hs[i] = messaging.Header{Key: string(h.Key), Value: h.Value}
	}
	return hs
}

func TraceKafkaProduce(ctx context.Context, msg *sarama.ProducerMessage) context.Context {
	return messaging.TraceKafkaProduce(ctx, &saramaProducerWrapper{msg})
}

func TraceKafkaConsume(ctx context.Context, msg *sarama.ConsumerMessage, group string) context.Context {
	return messaging.TraceKafkaConsume(ctx, &saramaConsumerWrapper{msg}, group)
}
