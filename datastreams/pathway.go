// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package datastreams

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

var hashableEdgeTags = map[string]struct{}{"event_type": {}, "exchange": {}, "group": {}, "topic": {}, "type": {}}

// Pathway is used to monitor how payloads are sent across different services.
// An example Pathway would be:
// service A -- edge 1 --> service B -- edge 2 --> service C
// So it's a branch of services (we also call them "nodes") connected via edges.
// As the payload is sent around, we save the start time (start of service A),
// and the start time of the previous service.
// This allows us to measure the latency of each edge, as well as the latency from origin of any service.
type Pathway struct {
	// hash is the hash of the current node, of the parent node, and of the edge that connects the parent node
	// to this node.
	hash uint64
	// pathwayStart is the start of the first node in the Pathway
	pathwayStart time.Time
	// edgeStart is the start of the previous node.
	edgeStart time.Time
	nChildren *int32
}

// Merge merges multiple pathways into one.
// The current implementation samples one resulting Pathway. A future implementation could be more clever
// and actually merge the Pathways.
func Merge(pathways []Pathway) Pathway {
	if len(pathways) == 0 {
		return Pathway{}
	}
	if len(pathways) == 1 {
		return pathways[0]
	}
	// Randomly select a pathway to propagate downstream.
	n := rand.Intn(len(pathways))
	now := time.Now()
	if aggregator := getGlobalAggregator(); aggregator != nil {
		for i, p := range pathways {
			if i != n {
				select {
				case aggregator.in <- statsPoint{
					// we hope that the point will already be in the store (I believe it should be since this pathway was
					// created before
					hash:      p.hash,
					timestamp: now.UnixNano(),
					// we pass this latency so we can compute origin timestamp
					pathwayLatency:  now.Sub(p.pathwayStart).Nanoseconds(),
					fanIn:           true,
					ignoreLatencies: true,
				}:
				default:
					atomic.AddInt64(&aggregator.stats.dropped, 1)
				}
			}
		}
	}
	return pathways[n]
}

func nodeHash(service, env, primaryTag string, edgeTags []string) uint64 {
	n := len(service) + len(env) + len(primaryTag)
	sort.Strings(edgeTags)
	for _, t := range edgeTags {
		n += len(t)
	}
	b := make([]byte, 0, n)
	b = append(b, service...)
	b = append(b, env...)
	b = append(b, primaryTag...)
	for _, t := range edgeTags {
		s := strings.Split(t, ":")
		if len(s) == 2 {
			if _, ok := hashableEdgeTags[s[0]]; !ok {
				continue
			}
			b = append(b, t...)
		}
	}
	h := fnv.New64()
	h.Write(b)
	return h.Sum64()
}

func pathwayHash(nodeHash, parentHash uint64) uint64 {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, nodeHash)
	binary.LittleEndian.PutUint64(b[8:], parentHash)
	h := fnv.New64()
	h.Write(b)
	return h.Sum64()
}

// NewPathway creates a new pathway.
func NewPathway(edgeTags ...string) Pathway {
	return newPathway(time.Now(), edgeTags...)
}

func newPathway(now time.Time, edgeTags ...string) Pathway {
	var nChildren int32
	p := Pathway{
		hash:         0,
		pathwayStart: now,
		edgeStart:    now,
		nChildren:    &nChildren,
	}
	return p.setCheckpoint(now, edgeTags)
}

// SetCheckpoint sets a checkpoint on a pathway.
func (p Pathway) SetCheckpoint(edgeTags ...string) Pathway {
	return p.setCheckpoint(time.Now(), edgeTags)
}

func (p Pathway) setCheckpoint(now time.Time, edgeTags []string) Pathway {
	fanOut := false
	if atomic.AddInt32(p.nChildren, 1) > 1 {
		fanOut = true
	}
	aggr := getGlobalAggregator()
	service := defaultServiceName
	primaryTag := ""
	env := ""
	if aggr != nil {
		service = aggr.service
		primaryTag = aggr.primaryTag
		env = aggr.env
	}
	var nChildren int32
	child := Pathway{
		hash:         pathwayHash(nodeHash(service, env, primaryTag, edgeTags), p.hash),
		pathwayStart: p.pathwayStart,
		edgeStart:    now,
		nChildren:    &nChildren,
	}
	if aggregator := getGlobalAggregator(); aggregator != nil {
		select {
		case aggregator.in <- statsPoint{
			edgeTags:       edgeTags,
			parentHash:     p.hash,
			hash:           child.hash,
			timestamp:      now.UnixNano(),
			pathwayLatency: now.Sub(p.pathwayStart).Nanoseconds(),
			edgeLatency:    now.Sub(p.edgeStart).Nanoseconds(),
			fanOut:         fanOut,
		}:
		default:
			atomic.AddInt64(&aggregator.stats.dropped, 1)
		}
	}
	return child
}
