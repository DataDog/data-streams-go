// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package datastreams

import (
	"context"
)

type contextKey struct{}

type CheckpointParams struct {
	PayloadSize int64
}

func NewCheckpointParams() CheckpointParams {
	return CheckpointParams{}
}

func (p CheckpointParams) WithPayloadSize(size int64) CheckpointParams {
	p.PayloadSize = size
	return p
}

var activePathwayKey = contextKey{}

// ContextWithPathway returns a copy of the given context which includes the pathway p.
func ContextWithPathway(ctx context.Context, p Pathway) context.Context {
	return context.WithValue(ctx, activePathwayKey, p)
}

// PathwayFromContext returns the pathway contained in the given context, and whether a
// pathway is found in ctx.
func PathwayFromContext(ctx context.Context) (p Pathway, ok bool) {
	if ctx == nil {
		return Pathway{}, false
	}
	v := ctx.Value(activePathwayKey)
	if s, ok := v.(Pathway); ok {
		return s, true
	}
	return Pathway{}, false
}

// SetCheckpoint sets a checkpoint on the pathway found in ctx.
// If there is no pathway in ctx, a new Pathway is returned.
func SetCheckpoint(ctx context.Context, edgeTags ...string) (Pathway, context.Context) {
	return SetCheckpointWithParams(ctx, CheckpointParams{}, edgeTags...)
}

func SetCheckpointWithParams(ctx context.Context, params CheckpointParams, edgeTags ...string) (Pathway, context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	p, ok := PathwayFromContext(ctx)
	if ok {
		p = p.SetCheckpointWithParams(params, edgeTags...)
	} else {
		p = NewPathwayWithParams(params, edgeTags...)
	}
	ctx = ContextWithPathway(ctx, p)
	return p, ctx
}

// contained in all contexts.
// This function should be used in fan-in situations. The current implementation keeps only 1 Pathway.
// A future implementation could merge multiple Pathways together and put the resulting Pathway in the context.
func MergeContexts(ctxs ...context.Context) context.Context {
	if len(ctxs) == 0 {
		return context.Background()
	}
	pathways := make([]Pathway, 0, len(ctxs))
	for _, ctx := range ctxs {
		if p, ok := PathwayFromContext(ctx); ok {
			pathways = append(pathways, p)
		}
	}
	if len(pathways) == 0 {
		return ctxs[0]
	}
	return ContextWithPathway(ctxs[0], Merge(pathways))
}
