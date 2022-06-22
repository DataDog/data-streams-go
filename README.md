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

The instrumentation relies on creating checkpoints on a pathway and passing them via headers inside the Kafka application. The product measures latency between the created Checkpoints.

You can set a checkpoint and add the pathway into your Kafka message headers all in one go like this:
```
import (ddkafka "github.com/DataDog/data-streams-go/integrations/kafka")
...
ctx = ddkafka.TraceKafkaProduce(ctx, &kafkaMsg)
```

Note that the output `ctx` from `TraceKafkaProduce()` contains information about the updated pathway. If you are sending multiple Kafka messages in one go, do not reuse the output `ctx` across calls.
