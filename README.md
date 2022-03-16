# data-streams-go

## Introduction

This product is meant to measure end to end latency in async pipelines.
It's in an Alpha phase.

## Glossary

- Data stream: A set of services connected together via *queues*
- Pathway: A single branch of connected services
- Queue: A connection between two services
- Edge latency: Latency of a queue between two services
- Latency from origin: Latency from the first tracked service, down to the current service
- Checkpoint: records at what time a specific operation on a payload occurred (eg: The payload was sent to Kafka). The product can then measure latency between checkpoints.

The product can measure edge latency, and latency from origin, for a set of checkpoints connected together via queues.
To do so, we propagate timestamps, and a hash of the path that messages took with the payload.

## Go instrumentation
**Prerequisites**
- Datadog Agent 7.34.0+
- latest version of [the data streams library](https://github.com/DataDog/data-streams-go)

You need to start the pipeline with `datastreams.Start()` at the start of your application.
Default trace agent URL is `localhost:8126`. If it's different for you, use the option:
```
datastreams.Start(datastreams.WithAgentAddr("notlocalhost:8126"))
```

The instrumentation relies on passing headers inside the Kafka application.
Right now, in Go, it's done manually.

Then, the product measures latency between the created Checkpoints.
We recommend putting the pipeline inside the context. So you will need:

- `p, ok := datapipeline.PathwayFromContext(ctx)` to get the pipeline from the context (in order to propagate it in headers)
- `_, ctx = datapipeline.SetCheckpoint(ctx, "type:kafka")` to set a checkpoint (if no pipeline exists in the context, it will create a new one).
- `datapipeline.ContextWithPathway(ctx, p)` to put a datapipeline inside the context (after extracting it from the headers).

Then, to put the data pipeline in headers, you will need:
```
p, err := datapipeline.Decode(bytes)
and
bytes := p.Encode()
``` 

This interface is going to evolve to be easier to use.
