package messaging

// This file abstracts away the interface between dd-streams
// and the underlying messaging library which it is used to monitor.
// It defines interface for producer & consumer messages which can then
// be chained together to provide stream analyses.

const (
	// PartitionAny is used to represent that partition is not
	// specified at the time of message production.
	PartitionAny int32 = -1
)

// Headers are transparent blobs of data that are stored along with
// each message. They serve as the vehicles of delivery to persist
// runtime observability metadata on disk.
type Header struct {
	Key   string
	Value []byte
}

// ProducerMessage interface abstracts away the various touch-points
// between dd-streams and messages produced to the underlying storage.
type ProducerMessage interface {
	GetTopic() *string
	GetPartition() int32
	AppendToHeaders(...Header)
	GetSize() int64
}

// ConsumerMessage inteface serves as the point of entry where messages
// read from underlying storage engine can be inspected for metadata
// and their corresponding metadata be injected back into `context.Context`.
type ConsumerMessage interface {
	GetTopic() *string
	GetPartition() int32
	GetHeaders() []Header
	GetSize() int64
}
