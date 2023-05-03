# Events

KubeFox is built around event sourcing. All requests and responses, even
synchronous calls, are modeled as events and are exchanged as messages via
brokers.

[CloudEvents](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md)
are used to wrap all events passed thru the message broker. This includes things
like requests/responses, audit records, etc.

To fulfil the requirement that the `source` attribute of the `CloudEvent` be a
URI a KubeFox scheme is described below:

```plain
kubefox://[componentId.componentHash.]componentName
```

Where `componentId.componentHash` is optional. If provided `componentId` must be
length 5 and `componentHash` must be length 7. Both must only contain lowercase
alpha-numeric characters.

As an example:

```plain
kubefox://6cqbp.ab883d2.joker
```

This is the URI to the component named `joker` with the id `6cqbp`.

An complete example of a `CloudEvent` request coming from the `api-gateway`
component with a response coming from the `joker` component is provided below.

Request:

```yaml
specversion: 1.0
type: io.kubefox.http_request
id: 69c6aac6-f839-4c98-a380-3c3a79bb2788
time: 2022-04-12T23:20:50.52Z
source: kubefox://vwqbx.223aa15.api-gateway/
dataschema: kubefox.proto.v1.KubeFoxData
datacontenttype: application/protobuf
```

Response:

```yaml
specversion: 1.0
type: io.kubefox.http_response
id: abb8ab78-c78e-4ab2-acf0-5fae242a5b6e
requestid: 69c6aac6-f839-4c98-a380-3c3a79bb2788
time: 2022-04-12T23:20:51.10Z
source: kubefox://6cqbp.ab883d2.joker/
dataschema: kubefox.proto.v1.KubeFoxData
datacontenttype: application/protobuf
```

## Common Attributes

These are common attributes that are included in events, logs, traces, and audit
records to allow correlation.

| Name           | Type            | Description | Cloud Event Key               | Log Key         | Trace Key                |
| -------------- | --------------- | ----------- | ----------------------------- | --------------- | ------------------------ |
| SysRef         | git hash, len 7 |             |                               | `systemTag`     | `kubefox.system-tag`     |
| Environment    | string          |             |                               | `environment`   | `kubefox.environment`    |
| Component Id   | string, len 5   |             |                               | `componentId`   | `kubefox.component-id`   |
| Component Name | string          |             |                               | `componentName` | `kubefox.component-name` |
| Component Hash | git hash, len 7 |             |                               | `componentHash` | `kubefox.component-hash` |
| Event Id       | UUID            |             | `id`                          | `eventId`       | `kubefox.event-id`       |
| Event Type     | string          |             | `type`                        | `eventType`     | `kubefox.event-type`     |
| Event Source   | URI             |             | `source`                      | `eventSource`   | `kubefox.event-source`   |
| Event Target   | URI             |             | `data.metadata.target`        |                 |                          |
| Trace Id       | 64-bit uint     |             | `data.metadata.span.trace_id` | `traceId`       | `trace`                  |
| Span Id        | 64-bit unit     |             | `data.metadata.span.span_id`  | `spanId`        | `span`                   |

## Event Types

### Request/Response

- `io.kubefox.component_request`
- `io.kubefox.component_response`
- `io.kubefox.cron_request`
- `io.kubefox.cron_response`
- `io.kubefox.dapr_request`
- `io.kubefox.dapr_response`
- `io.kubefox.http_request`
- `io.kubefox.http_response`
- `io.kubefox.kubernetes_request`
- `io.kubefox.kubernetes_response`
- `io.kubefox.metrics_request`
- `io.kubefox.metrics_response`
- `io.kubefox.telemetry_request`
- `io.kubefox.telemetry_response`

### Audit

- `io.kubefox.audit_record`

### Errors

- `io.kubefox.error`
- `io.kubefox.rejected`

## Message Subjects

| Event Type | Subject                                    |
| ---------- | ------------------------------------------ |
| Request    | `{componentName}.{componentHash}.req`      |
| Response   | `{componentName}.{componentHash}.{id}.res` |
