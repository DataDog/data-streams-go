// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:generate msgp -unexported -marshal=false -o=payload_msgp.go -tests=false

package datastreams

// StatsPayload stores client computed stats.
type StatsPayload struct {
	// Env specifies the env. of the application, as defined by the user.
	Env string
	// Service is the service of the application
	Service string
	// PrimaryTag is the primary tag of the application.
	PrimaryTag string
	// Stats holds all stats buckets computed within this payload.
	Stats []StatsBucket
}

// StatsBucket specifies a set of stats computed over a duration.
type StatsBucket struct {
	// Start specifies the beginning of this bucket.
	Start uint64
	// Duration specifies the duration of this bucket.
	Duration uint64
	// Stats contains a set of statistics computed for the duration of this bucket.
	Stats []StatsPoint
}

// StatsPoint contains a set of statistics grouped under various aggregation keys.
type StatsPoint struct {
	// These fields indicate the properties under which the stats were aggregated.
	Service    string // deprecated
	EdgeTags   []string
	Hash       uint64
	ParentHash uint64
	// These fields specify the stats for the above aggregation.
	PathwayLatency []byte
	EdgeLatency    []byte
}
