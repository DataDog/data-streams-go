// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package datastreams

import (
	"log"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/DataDog/sketches-go/ddsketch/mapping"
	"github.com/DataDog/sketches-go/ddsketch/store"
	"github.com/golang/protobuf/proto"

	"github.com/DataDog/data-streams-go/datastreams/version"
)

const (
	bucketDuration     = time.Second * 10
	defaultServiceName = "unnamed-go-service"
)

var sketchMapping, _ = mapping.NewLogarithmicMapping(0.01)

type statsPoint struct {
	edgeTags       []string
	hash           uint64
	parentHash     uint64
	timestamp      int64
	pathwayLatency int64
	edgeLatency    int64
}

type statsGroup struct {
	service        string
	edgeTags       []string
	hash           uint64
	parentHash     uint64
	pathwayLatency *ddsketch.DDSketch
	edgeLatency    *ddsketch.DDSketch
}

type bucket map[uint64]statsGroup

func (b bucket) export(timestampType string) []StatsPoint {
	stats := make([]StatsPoint, 0, len(b))
	for _, s := range b {
		pathwayLatency, err := proto.Marshal(s.pathwayLatency.ToProto())
		if err != nil {
			log.Printf("ERROR: can't serialize pathway latency. Ignoring: %v", err)
			continue
		}
		edgeLatency, err := proto.Marshal(s.edgeLatency.ToProto())
		if err != nil {
			log.Printf("ERROR: can't serialize edge latency. Ignoring: %v", err)
			continue
		}
		stats = append(stats, StatsPoint{
			PathwayLatency: pathwayLatency,
			EdgeLatency:    edgeLatency,
			Service:        s.service,
			EdgeTags:       s.edgeTags,
			Hash:           s.hash,
			ParentHash:     s.parentHash,
			TimestampType:  timestampType,
		})
	}
	return stats
}

type aggregatorStats struct {
	payloadsIn      int64
	flushedPayloads int64
	flushedBuckets  int64
	flushErrors     int64
	dropped         int64
}

type aggregator struct {
	in                   chan statsPoint
	tsTypeCurrentBuckets map[int64]bucket
	tsTypeOriginBuckets  map[int64]bucket
	wg                   sync.WaitGroup
	stopped              uint64
	stop                 chan struct{} // closing this channel triggers shutdown
	stats                aggregatorStats
	transport            *httpTransport
	statsd               statsd.ClientInterface
	env                  string
	primaryTag           string
	service              string
}

func newAggregator(statsd statsd.ClientInterface, env, primaryTag, service, agentAddr string, httpClient *http.Client, site, apiKey string, agentLess bool) *aggregator {
	return &aggregator{
		tsTypeCurrentBuckets: make(map[int64]bucket),
		tsTypeOriginBuckets:  make(map[int64]bucket),
		in:                   make(chan statsPoint, 10000),
		stopped:              1,
		statsd:               statsd,
		env:                  env,
		primaryTag:           primaryTag,
		service:              service,
		transport:            newHTTPTransport(agentAddr, site, apiKey, httpClient, agentLess),
	}
}

// alignTs returns the provided timestamp truncated to the bucket size.
// It gives us the start time of the time bucket in which such timestamp falls.
func alignTs(ts, bucketSize int64) int64 { return ts - ts%bucketSize }

func (a *aggregator) addToBuckets(point statsPoint, btime int64, buckets map[int64]bucket) {
	b, ok := buckets[btime]
	if !ok {
		b = make(bucket)
		buckets[btime] = b
	}
	group, ok := b[point.hash]
	if !ok {
		group = statsGroup{
			edgeTags:       point.edgeTags,
			parentHash:     point.parentHash,
			hash:           point.hash,
			pathwayLatency: ddsketch.NewDDSketch(sketchMapping, store.DenseStoreConstructor(), store.DenseStoreConstructor()),
			edgeLatency:    ddsketch.NewDDSketch(sketchMapping, store.DenseStoreConstructor(), store.DenseStoreConstructor()),
		}
		b[point.hash] = group
	}
	if err := group.pathwayLatency.Add(math.Max(float64(point.pathwayLatency)/float64(time.Second), 0)); err != nil {
		log.Printf("ERROR: failed to add pathway latency. Ignoring %v.", err)
	}
	if err := group.edgeLatency.Add(math.Max(float64(point.edgeLatency)/float64(time.Second), 0)); err != nil {
		log.Printf("ERROR: failed to add edge latency. Ignoring %v.", err)
	}
}

func (a *aggregator) add(point statsPoint) {
	currentBucketTime := alignTs(point.timestamp, bucketDuration.Nanoseconds())
	a.addToBuckets(point, currentBucketTime, a.tsTypeCurrentBuckets)
	originTimestamp := point.timestamp - point.pathwayLatency
	originBucketTime := alignTs(originTimestamp, bucketDuration.Nanoseconds())
	a.addToBuckets(point, originBucketTime, a.tsTypeOriginBuckets)
}

func (a *aggregator) run(tick <-chan time.Time) {
	for {
		select {
		case s := <-a.in:
			atomic.AddInt64(&a.stats.payloadsIn, 1)
			a.add(s)
		case now := <-tick:
			a.sendToAgent(a.flush(now))
		case <-a.stop:
			// drop in flight payloads on the input channel
			a.sendToAgent(a.flush(time.Now().Add(bucketDuration * 10)))
			return
		}
	}
}

func (a *aggregator) Start() {
	if atomic.SwapUint64(&a.stopped, 0) == 0 {
		// already running
		log.Print("WARN: (*aggregator).Start called more than once. This is likely a programming error.")
		return
	}
	a.stop = make(chan struct{})
	a.wg.Add(1)
	go a.reportStats()
	go func() {
		defer a.wg.Done()
		tick := time.NewTicker(bucketDuration)
		defer tick.Stop()
		a.run(tick.C)
	}()
}

func (a *aggregator) Stop() {
	if atomic.SwapUint64(&a.stopped, 1) > 0 {
		return
	}
	close(a.stop)
	a.wg.Wait()
}

func (a *aggregator) reportStats() {
	for range time.NewTicker(time.Second * 10).C {
		a.statsd.Count("datadog.datastreams.aggregator.payloads_in", atomic.SwapInt64(&a.stats.payloadsIn, 0), nil, 1)
		a.statsd.Count("datadog.datastreams.aggregator.flushed_payloads", atomic.SwapInt64(&a.stats.flushedPayloads, 0), nil, 1)
		a.statsd.Count("datadog.datastreams.aggregator.flushed_buckets", atomic.SwapInt64(&a.stats.flushedBuckets, 0), nil, 1)
		a.statsd.Count("datadog.datastreams.aggregator.flush_errors", atomic.SwapInt64(&a.stats.flushErrors, 0), nil, 1)
		a.statsd.Count("datadog.datastreams.dropped_payloads", atomic.SwapInt64(&a.stats.dropped, 0), nil, 1)
	}
}

func (a *aggregator) runFlusher() {
	for {
		select {
		case <-a.stop:
			// flush everything, so add a few bucketDurations to the current time in order to get a good margin.
			return
		}
	}
}

func (a *aggregator) flushBucket(buckets map[int64]bucket, bucketStart int64, timestampType string) StatsBucket {
	bucket := buckets[bucketStart]
	delete(buckets, bucketStart)
	return StatsBucket{
		Start:    uint64(bucketStart),
		Duration: uint64(bucketDuration.Nanoseconds()),
		Stats:    bucket.export(timestampType),
	}
}

func (a *aggregator) flush(now time.Time) StatsPayload {
	nowNano := now.UnixNano()
	sp := StatsPayload{
		Service:       a.service,
		Env:           a.env,
		PrimaryTag:    a.primaryTag,
		Lang:          "go",
		TracerVersion: version.Tag,
		Stats:         make([]StatsBucket, 0, len(a.tsTypeCurrentBuckets)+len(a.tsTypeOriginBuckets)),
	}
	for ts := range a.tsTypeCurrentBuckets {
		if ts > nowNano-bucketDuration.Nanoseconds() {
			// do not flush the bucket at the current time
			continue
		}
		sp.Stats = append(sp.Stats, a.flushBucket(a.tsTypeCurrentBuckets, ts, "current"))
	}
	for ts := range a.tsTypeOriginBuckets {
		if ts > nowNano-bucketDuration.Nanoseconds() {
			// do not flush the bucket at the current time
			continue
		}
		sp.Stats = append(sp.Stats, a.flushBucket(a.tsTypeOriginBuckets, ts, "origin"))
	}
	return sp
}

func (a *aggregator) sendToAgent(payload StatsPayload) {
	atomic.AddInt64(&a.stats.flushedPayloads, 1)
	atomic.AddInt64(&a.stats.flushedBuckets, int64(len(payload.Stats)))
	if err := a.transport.sendPipelineStats(&payload); err != nil {
		atomic.AddInt64(&a.stats.flushErrors, 1)
	}
}
