package datastreams

import (
	"time"
)

func TrackKafkaCommitOffset(group string, topic string, partition int32, offset int64) {
	if aggregator := getGlobalAggregator(); aggregator != nil {
		aggregator.in.push(&aggregatorInput{typ: pointTypeKafkaOffset, kafkaOffset: kafkaOffset{
			offset:     offset,
			group:      group,
			topic:      topic,
			partition:  partition,
			offsetType: commitOffset,
			timestamp:  time.Now().UnixNano(),
		}})
	}
}

func TrackKafkaProduce(topic string, partition int32, offset int64) {
	if aggregator := getGlobalAggregator(); aggregator != nil {
		aggregator.in.push(&aggregatorInput{typ: pointTypeKafkaOffset, kafkaOffset: kafkaOffset{
			offset:     offset,
			topic:      topic,
			partition:  partition,
			offsetType: produceOffset,
			timestamp:  time.Now().UnixNano(),
		}})
	}
}
