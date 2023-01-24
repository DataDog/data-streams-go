package datastreams

import "sync/atomic"

func TrackKafkaCommitOffset(group string, topic string, partition int32, offset int64) {
	if aggregator := getGlobalAggregator(); aggregator != nil {
		select {
		case aggregator.inKafka <- kafkaOffset{
			offset:     offset,
			group:      group,
			topic:      topic,
			partition:  partition,
			offsetType: commitOffset,
		}:
		default:
			atomic.AddInt64(&aggregator.stats.dropped, 1)
		}
	}
}

func TrackKafkaProduce(topic string, partition int32, offset int64) {
	if aggregator := getGlobalAggregator(); aggregator != nil {
		select {
		case aggregator.inKafka <- kafkaOffset{
			offset:     offset,
			topic:      topic,
			partition:  partition,
			offsetType: produceOffset,
		}:
		default:
			atomic.AddInt64(&aggregator.stats.dropped, 1)
		}
	}
}
