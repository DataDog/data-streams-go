# data-streams-go

## Introduction

This product is meant to measure end to end latency in async pipelines.
It's in an Alpha phase. We currently support instrumentations of async pipelines using Kafka. Integrations with other systems will follow soon.

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

The instrumentation relies on creating checkpoints at various points in your data stream services, recording the pathway that messages take along the way. These pathways are stored within a Go Context, and are passed around via message headers.

### M1 support of the Kafka instrumentation

This library relies on the [Confluent Kafka Go Library](https://github.com/confluentinc/confluent-kafka-go). On M1 machines, the following parameters are needed when starting the instrumented service:
```
-tags dynamic
```

### Kafka

To instrument your data stream services that use Kafka queues, you can use the library provided under `github.com/DataDog/data-streams-go/integrations/kafka`.

On the producer side, before sending out a Kafka message you can call `TraceKafkaProduce()`, which sets a new checkpoint onto any existing pathway in the provided Go Context (or creates a new pathway if none are found). It then adds the pathway into your Kafka message headers.
```
import (ddkafka "github.com/DataDog/data-streams-go/integrations/kafka")
...
ctx = ddkafka.TraceKafkaProduce(ctx, &kafkaMsg)
```

Similarly, on the consumer side, you can call `TraceKafkaConsume()`, which extracts the pathway that a particular Kafka message has gone through so far. It also sets a new checkpoint on the pathway to record the successful consumption of a message, and then finally it stores the pathway into the provided Go Context. 
```
import (ddkafka "github.com/DataDog/data-streams-go/integrations/kafka")
...
ctx = ddkafka.TraceKafkaConsume(ctx, &kafkaMsg, consumer_group)
```

Please note that the output `ctx` from `TraceKafkaProduce()` and `TraceKafkaConsume()` both contains information about the updated pathway. For `TraceKafkaProduce()`, if you are sending multiple Kafka messages in one go (i.e. fan-out situations), do not reuse the output `ctx` across calls. And for `TraceKafkaConsume()`, if you are aggregating multiple messages to create a smaller number of payloads (i.e. fan-in situations), merge the resulting Contexts from `TraceKafkaConsume()` using `MergeContexts()` from `github.com/DataDog/data-streams-go`. This resulting Context can then be passed into the next `TraceKafkaProduce()` call.
```
import (
    datastreams "github.com/DataDog/data-streams-go"
    ddkafka "github.com/DataDog/data-streams-go/integrations/kafka"
)

...

contexts := []Context{}
for (...) {
    contexts = append(contexts, ddkafka.TraceKafkaConsume(ctx, &consumedMsg, consumer_group))
}
mergedContext = datastreams.MergeContexts(contexts...)

...

ddkafka.TraceKafkaProduce(mergedContext, &producedMsg)
```

### Manual instrumentation

the example below is for HTTP, but you can instrument any technology you want with these manual instrumentation.
In HTTP, we propagate the pathway through HTTP headers.

In the http client, to inject the pathway, use:
```
req, err := http.NewRequest(...)
...
p, ok := datastreams.PathwayFromContext(ctx)
if ok {
   req.Headers.Set(datastreams.PropagationKeyBase64, p.EncodeStr())
}
```

And to extract the pathway from HTTP headers (inside the HTTP server), use:
```
func extractPathwayToContext(req *http.Request) context.Context {
	ctx := req.Context()
	p, err := datastreams.DecodeStr(req.Header.Get(datastreams.PropagationKeyBase64))
	if err != nil {
		return ctx
	}
	ctx = datastreams.ContextWithPathway(ctx, p)
	_, ctx = datastreams.SetCheckpoint(ctx, string[]{"type:http", "direction:in"})
}

```
