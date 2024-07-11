# Protocol Buffers




<a name="kubefoxprotov1component"></a>

### Component



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  |  |
| app | [string](#string) |  |  |
| name | [string](#string) |  |  |
| hash | [string](#string) |  |  |
| id | [string](#string) |  |  |
| broker_id | [string](#string) |  |  |






<a name="kubefoxprotov1event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| parent_id | [string](#string) |  |  |
| parent_span | [SpanContext](#kubefoxprotov1spancontext) |  |  |
| type | [string](#string) |  |  |
| category | [Category](#kubefoxprotov1category) |  |  |
| create_time | [int64](#int64) |  | Unix time in nanosecond |
| ttl | [int64](#int64) |  | TTL in nanosecond |
| context | [EventContext](#kubefoxprotov1eventcontext) |  |  |
| source | [Component](#kubefoxprotov1component) |  |  |
| target | [Component](#kubefoxprotov1component) |  |  |
| params | [Event.ParamsEntry](#kubefoxprotov1eventparamsentry) | repeated |  |
| values | [Event.ValuesEntry](#kubefoxprotov1eventvaluesentry) | repeated |  |
| content_type | [string](#string) |  |  |
| content | [bytes](#bytes) |  |  |






<a name="kubefoxprotov1eventparamsentry"></a>

### Event.ParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Value](#googleprotobufvalue) |  |  |






<a name="kubefoxprotov1eventvaluesentry"></a>

### Event.ValuesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Value](#googleprotobufvalue) |  |  |






<a name="kubefoxprotov1eventcontext"></a>

### EventContext



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| platform | [string](#string) |  |  |
| virtual_environment | [string](#string) |  |  |
| app_deployment | [string](#string) |  |  |
| release_manifest | [string](#string) |  |  |






<a name="kubefoxprotov1matchedevent"></a>

### MatchedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#kubefoxprotov1event) |  |  |
| route_id | [int64](#int64) |  |  |
| env | [MatchedEvent.EnvEntry](#kubefoxprotov1matchedeventenventry) | repeated |  |






<a name="kubefoxprotov1matchedeventenventry"></a>

### MatchedEvent.EnvEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Value](#googleprotobufvalue) |  |  |






<a name="kubefoxprotov1spancontext"></a>

### SpanContext



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trace_id | [bytes](#bytes) |  |  |
| span_id | [bytes](#bytes) |  |  |
| trace_state | [string](#string) |  |  |
| flags | [fixed32](#fixed32) |  |  |






<a name="kubefoxprotov1telemetry"></a>

### Telemetry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trace_id | [bytes](#bytes) |  |  |
| log_records | [opentelemetry.proto.logs.v1.LogRecord](#opentelemetryprotologsv1logrecord) | repeated |  |
| metrics | [opentelemetry.proto.metrics.v1.Metric](#opentelemetryprotometricsv1metric) | repeated |  |
| spans | [opentelemetry.proto.trace.v1.Span](#opentelemetryprototracev1span) | repeated |  |





 <!-- end messages -->


<a name="kubefoxprotov1category"></a>

### Category


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| MESSAGE | 1 |  |
| REQUEST | 2 |  |
| RESPONSE | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kubefoxprotov1broker"></a>

### Broker


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Subscribe | [Event](#kubefoxprotov1event) stream | [MatchedEvent](#kubefoxprotov1matchedevent) stream |  |

 <!-- end services -->





<a name="googleprotobuflistvalue"></a>

### ListValue
`ListValue` is a wrapper around a repeated field of values.

The JSON representation for `ListValue` is JSON array.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [Value](#googleprotobufvalue) | repeated | Repeated field of dynamically typed values. |






<a name="googleprotobufstruct"></a>

### Struct
`Struct` represents a structured data value, consisting of fields
which map to dynamically typed values. In some languages, `Struct`
might be supported by a native representation. For example, in
scripting languages like JS a struct is represented as an
object. The details of that representation are described together
with the proto support for the language.

The JSON representation for `Struct` is JSON object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [Struct.FieldsEntry](#googleprotobufstructfieldsentry) | repeated | Unordered map of dynamically typed values. |






<a name="googleprotobufstructfieldsentry"></a>

### Struct.FieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#googleprotobufvalue) |  |  |






<a name="googleprotobufvalue"></a>

### Value
`Value` represents a dynamically typed value which can be either
null, a number, a string, a boolean, a recursive struct value, or a
list of values. A producer of value is expected to set one of these
variants. Absence of any variant indicates an error.

The JSON representation for `Value` is JSON value.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| null_value | [NullValue](#googleprotobufnullvalue) |  | Represents a null value. |
| number_value | [double](#double) |  | Represents a double value. |
| string_value | [string](#string) |  | Represents a string value. |
| bool_value | [bool](#bool) |  | Represents a boolean value. |
| struct_value | [Struct](#googleprotobufstruct) |  | Represents a structured value. |
| list_value | [ListValue](#googleprotobuflistvalue) |  | Represents a repeated `Value`. |





 <!-- end messages -->


<a name="googleprotobufnullvalue"></a>

### NullValue
`NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

 The JSON representation for `NullValue` is JSON `null`.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NULL_VALUE | 0 | Null value. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->





<a name="opentelemetryprotocommonv1anyvalue"></a>

### AnyValue
AnyValue is used to represent any type of attribute value. AnyValue may contain a
primitive value such as a string or integer or it may contain an arbitrary nested
object containing arrays, key-value lists and primitives.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| string_value | [string](#string) |  |  |
| bool_value | [bool](#bool) |  |  |
| int_value | [int64](#int64) |  |  |
| double_value | [double](#double) |  |  |
| array_value | [ArrayValue](#opentelemetryprotocommonv1arrayvalue) |  |  |
| kvlist_value | [KeyValueList](#opentelemetryprotocommonv1keyvaluelist) |  |  |
| bytes_value | [bytes](#bytes) |  |  |






<a name="opentelemetryprotocommonv1arrayvalue"></a>

### ArrayValue
ArrayValue is a list of AnyValue messages. We need ArrayValue as a message
since oneof in AnyValue does not allow repeated fields.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [AnyValue](#opentelemetryprotocommonv1anyvalue) | repeated | Array of values. The array may be empty (contain 0 elements). |






<a name="opentelemetryprotocommonv1instrumentationscope"></a>

### InstrumentationScope
InstrumentationScope is a message representing the instrumentation scope information
such as the fully qualified name and version.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | An empty instrumentation scope name means the name is unknown. |
| version | [string](#string) |  |  |
| attributes | [KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | Additional attributes that describe the scope. [Optional]. Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| dropped_attributes_count | [uint32](#uint32) |  |  |






<a name="opentelemetryprotocommonv1keyvalue"></a>

### KeyValue
KeyValue is a key-value pair that is used to store Span attributes, Link
attributes, etc.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [AnyValue](#opentelemetryprotocommonv1anyvalue) |  |  |






<a name="opentelemetryprotocommonv1keyvaluelist"></a>

### KeyValueList
KeyValueList is a list of KeyValue messages. We need KeyValueList as a message
since `oneof` in AnyValue does not allow repeated fields. Everywhere else where we need
a list of KeyValue messages (e.g. in Span) we use `repeated KeyValue` directly to
avoid unnecessary extra wrapping (which slows down the protocol). The 2 approaches
are semantically equivalent.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | A collection of key/value pairs of key-value pairs. The list may be empty (may contain 0 elements). The keys MUST be unique (it is not allowed to have more than one value with the same key). |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->





<a name="opentelemetryprotologsv1logrecord"></a>

### LogRecord
A log record according to OpenTelemetry Log Data Model:
https://github.com/open-telemetry/oteps/blob/main/text/logs/0097-log-data-model.md


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| time_unix_nano | [fixed64](#fixed64) |  | time_unix_nano is the time when the event occurred. Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. Value of 0 indicates unknown or missing timestamp. |
| observed_time_unix_nano | [fixed64](#fixed64) |  | Time when the event was observed by the collection system. For events that originate in OpenTelemetry (e.g. using OpenTelemetry Logging SDK) this timestamp is typically set at the generation time and is equal to Timestamp. For events originating externally and collected by OpenTelemetry (e.g. using Collector) this is the time when OpenTelemetry's code observed the event measured by the clock of the OpenTelemetry code. This field MUST be set once the event is observed by OpenTelemetry.

For converting OpenTelemetry log data to formats that support only one timestamp or when receiving OpenTelemetry log data by recipients that support only one timestamp internally the following logic is recommended: - Use time_unix_nano if it is present, otherwise use observed_time_unix_nano.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. Value of 0 indicates unknown or missing timestamp. |
| severity_number | [SeverityNumber](#opentelemetryprotologsv1severitynumber) |  | Numerical value of the severity, normalized to values described in Log Data Model. [Optional]. |
| severity_text | [string](#string) |  | The severity text (also known as log level). The original string representation as it is known at the source. [Optional]. |
| body | [opentelemetry.proto.common.v1.AnyValue](#opentelemetryprotocommonv1anyvalue) |  | A value containing the body of the log record. Can be for example a human-readable string message (including multi-line) describing the event in a free form or it can be a structured data composed of arrays and maps of other values. [Optional]. |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | Additional attributes that describe the specific event occurrence. [Optional]. Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| dropped_attributes_count | [uint32](#uint32) |  |  |
| flags | [fixed32](#fixed32) |  | Flags, a bit field. 8 least significant bits are the trace flags as defined in W3C Trace Context specification. 24 most significant bits are reserved and must be set to 0. Readers must not assume that 24 most significant bits will be zero and must correctly mask the bits when reading 8-bit trace flag (use flags & LOG_RECORD_FLAGS_TRACE_FLAGS_MASK). [Optional]. |
| trace_id | [bytes](#bytes) |  | A unique identifier for a trace. All logs from the same trace share the same `trace_id`. The ID is a 16-byte array. An ID with all zeroes OR of length other than 16 bytes is considered invalid (empty string in OTLP/JSON is zero-length and thus is also invalid).

This field is optional.

The receivers SHOULD assume that the log record is not associated with a trace if any of the following is true: - the field is not present, - the field contains an invalid value. |
| span_id | [bytes](#bytes) |  | A unique identifier for a span within a trace, assigned when the span is created. The ID is an 8-byte array. An ID with all zeroes OR of length other than 8 bytes is considered invalid (empty string in OTLP/JSON is zero-length and thus is also invalid).

This field is optional. If the sender specifies a valid span_id then it SHOULD also specify a valid trace_id.

The receivers SHOULD assume that the log record is not associated with a span if any of the following is true: - the field is not present, - the field contains an invalid value. |






<a name="opentelemetryprotologsv1logsdata"></a>

### LogsData
LogsData represents the logs data that can be stored in a persistent storage,
OR can be embedded by other protocols that transfer OTLP logs data but do not
implement the OTLP protocol.

The main difference between this message and collector protocol is that
in this message there will not be any "control" or "metadata" specific to
OTLP protocol.

When new fields are added into this message, the OTLP request MUST be updated
as well.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_logs | [ResourceLogs](#opentelemetryprotologsv1resourcelogs) | repeated | An array of ResourceLogs. For data coming from a single resource this array will typically contain one element. Intermediary nodes that receive data from multiple origins typically batch the data before forwarding further and in that case this array will contain multiple elements. |






<a name="opentelemetryprotologsv1resourcelogs"></a>

### ResourceLogs
A collection of ScopeLogs from a Resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [opentelemetry.proto.resource.v1.Resource](#opentelemetryprotoresourcev1resource) |  | The resource for the logs in this message. If this field is not set then resource info is unknown. |
| scope_logs | [ScopeLogs](#opentelemetryprotologsv1scopelogs) | repeated | A list of ScopeLogs that originate from a resource. |
| schema_url | [string](#string) |  | The Schema URL, if known. This is the identifier of the Schema that the resource data is recorded in. To learn more about Schema URL see https://opentelemetry.io/docs/specs/otel/schemas/#schema-url This schema_url applies to the data in the "resource" field. It does not apply to the data in the "scope_logs" field which have their own schema_url field. |






<a name="opentelemetryprotologsv1scopelogs"></a>

### ScopeLogs
A collection of Logs produced by a Scope.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| scope | [opentelemetry.proto.common.v1.InstrumentationScope](#opentelemetryprotocommonv1instrumentationscope) |  | The instrumentation scope information for the logs in this message. Semantically when InstrumentationScope isn't set, it is equivalent with an empty instrumentation scope name (unknown). |
| log_records | [LogRecord](#opentelemetryprotologsv1logrecord) | repeated | A list of log records. |
| schema_url | [string](#string) |  | The Schema URL, if known. This is the identifier of the Schema that the log data is recorded in. To learn more about Schema URL see https://opentelemetry.io/docs/specs/otel/schemas/#schema-url This schema_url applies to all logs in the "logs" field. |





 <!-- end messages -->


<a name="opentelemetryprotologsv1logrecordflags"></a>

### LogRecordFlags
LogRecordFlags represents constants used to interpret the
LogRecord.flags field, which is protobuf 'fixed32' type and is to
be used as bit-fields. Each non-zero value defined in this enum is
a bit-mask.  To extract the bit-field, for example, use an
expression like:

  (logRecord.flags & LOG_RECORD_FLAGS_TRACE_FLAGS_MASK)

| Name | Number | Description |
| ---- | ------ | ----------- |
| LOG_RECORD_FLAGS_DO_NOT_USE | 0 | The zero value for the enum. Should not be used for comparisons. Instead use bitwise "and" with the appropriate mask as shown above. |
| LOG_RECORD_FLAGS_TRACE_FLAGS_MASK | 255 | Bits 0-7 are used for trace flags. |



<a name="opentelemetryprotologsv1severitynumber"></a>

### SeverityNumber
Possible values for LogRecord.SeverityNumber.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SEVERITY_NUMBER_UNSPECIFIED | 0 | UNSPECIFIED is the default SeverityNumber, it MUST NOT be used. |
| SEVERITY_NUMBER_TRACE | 1 |  |
| SEVERITY_NUMBER_TRACE2 | 2 |  |
| SEVERITY_NUMBER_TRACE3 | 3 |  |
| SEVERITY_NUMBER_TRACE4 | 4 |  |
| SEVERITY_NUMBER_DEBUG | 5 |  |
| SEVERITY_NUMBER_DEBUG2 | 6 |  |
| SEVERITY_NUMBER_DEBUG3 | 7 |  |
| SEVERITY_NUMBER_DEBUG4 | 8 |  |
| SEVERITY_NUMBER_INFO | 9 |  |
| SEVERITY_NUMBER_INFO2 | 10 |  |
| SEVERITY_NUMBER_INFO3 | 11 |  |
| SEVERITY_NUMBER_INFO4 | 12 |  |
| SEVERITY_NUMBER_WARN | 13 |  |
| SEVERITY_NUMBER_WARN2 | 14 |  |
| SEVERITY_NUMBER_WARN3 | 15 |  |
| SEVERITY_NUMBER_WARN4 | 16 |  |
| SEVERITY_NUMBER_ERROR | 17 |  |
| SEVERITY_NUMBER_ERROR2 | 18 |  |
| SEVERITY_NUMBER_ERROR3 | 19 |  |
| SEVERITY_NUMBER_ERROR4 | 20 |  |
| SEVERITY_NUMBER_FATAL | 21 |  |
| SEVERITY_NUMBER_FATAL2 | 22 |  |
| SEVERITY_NUMBER_FATAL3 | 23 |  |
| SEVERITY_NUMBER_FATAL4 | 24 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->





<a name="opentelemetryprotometricsv1exemplar"></a>

### Exemplar
A representation of an exemplar, which is a sample input measurement.
Exemplars also hold information about the environment when the measurement
was recorded, for example the span and trace ID of the active span when the
exemplar was recorded.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filtered_attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | The set of key/value pairs that were filtered out by the aggregator, but recorded alongside the original measurement. Only key/value pairs that were filtered out by the aggregator should be included |
| time_unix_nano | [fixed64](#fixed64) |  | time_unix_nano is the exact time when this exemplar was recorded

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| as_double | [double](#double) |  |  |
| as_int | [sfixed64](#sfixed64) |  |  |
| span_id | [bytes](#bytes) |  | (Optional) Span ID of the exemplar trace. span_id may be missing if the measurement is not recorded inside a trace or if the trace is not sampled. |
| trace_id | [bytes](#bytes) |  | (Optional) Trace ID of the exemplar trace. trace_id may be missing if the measurement is not recorded inside a trace or if the trace is not sampled. |






<a name="opentelemetryprotometricsv1exponentialhistogram"></a>

### ExponentialHistogram
ExponentialHistogram represents the type of a metric that is calculated by aggregating
as a ExponentialHistogram of all reported double measurements over a time interval.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data_points | [ExponentialHistogramDataPoint](#opentelemetryprotometricsv1exponentialhistogramdatapoint) | repeated |  |
| aggregation_temporality | [AggregationTemporality](#opentelemetryprotometricsv1aggregationtemporality) |  | aggregation_temporality describes if the aggregator reports delta changes since last report time, or cumulative changes since a fixed start time. |






<a name="opentelemetryprotometricsv1exponentialhistogramdatapoint"></a>

### ExponentialHistogramDataPoint
ExponentialHistogramDataPoint is a single data point in a timeseries that describes the
time-varying values of a ExponentialHistogram of double values. A ExponentialHistogram contains
summary statistics for a population of values, it may optionally contain the
distribution of those values across a set of buckets.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | The set of key/value pairs that uniquely identify the timeseries from where this point belongs. The list may be empty (may contain 0 elements). Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| start_time_unix_nano | [fixed64](#fixed64) |  | StartTimeUnixNano is optional but strongly encouraged, see the the detailed comments above Metric.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| time_unix_nano | [fixed64](#fixed64) |  | TimeUnixNano is required, see the detailed comments above Metric.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| count | [fixed64](#fixed64) |  | count is the number of values in the population. Must be non-negative. This value must be equal to the sum of the "bucket_counts" values in the positive and negative Buckets plus the "zero_count" field. |
| sum | [double](#double) | optional | sum of the values in the population. If count is zero then this field must be zero.

Note: Sum should only be filled out when measuring non-negative discrete events, and is assumed to be monotonic over the values of these events. Negative events *can* be recorded, but sum should not be filled out when doing so. This is specifically to enforce compatibility w/ OpenMetrics, see: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#histogram |
| scale | [sint32](#sint32) |  | scale describes the resolution of the histogram. Boundaries are located at powers of the base, where:

 base = (2^(2^-scale))

The histogram bucket identified by `index`, a signed integer, contains values that are greater than (base^index) and less than or equal to (base^(index+1)).

The positive and negative ranges of the histogram are expressed separately. Negative values are mapped by their absolute value into the negative range using the same scale as the positive range.

scale is not restricted by the protocol, as the permissible values depend on the range of the data. |
| zero_count | [fixed64](#fixed64) |  | zero_count is the count of values that are either exactly zero or within the region considered zero by the instrumentation at the tolerated degree of precision. This bucket stores values that cannot be expressed using the standard exponential formula as well as values that have been rounded to zero.

Implementations MAY consider the zero bucket to have probability mass equal to (zero_count / count). |
| positive | [ExponentialHistogramDataPoint.Buckets](#opentelemetryprotometricsv1exponentialhistogramdatapointbuckets) |  | positive carries the positive range of exponential bucket counts. |
| negative | [ExponentialHistogramDataPoint.Buckets](#opentelemetryprotometricsv1exponentialhistogramdatapointbuckets) |  | negative carries the negative range of exponential bucket counts. |
| flags | [uint32](#uint32) |  | Flags that apply to this specific data point. See DataPointFlags for the available flags and their meaning. |
| exemplars | [Exemplar](#opentelemetryprotometricsv1exemplar) | repeated | (Optional) List of exemplars collected from measurements that were used to form the data point |
| min | [double](#double) | optional | min is the minimum value over (start_time, end_time]. |
| max | [double](#double) | optional | max is the maximum value over (start_time, end_time]. |
| zero_threshold | [double](#double) |  | ZeroThreshold may be optionally set to convey the width of the zero region. Where the zero region is defined as the closed interval [-ZeroThreshold, ZeroThreshold]. When ZeroThreshold is 0, zero count bucket stores values that cannot be expressed using the standard exponential formula as well as values that have been rounded to zero. |






<a name="opentelemetryprotometricsv1exponentialhistogramdatapointbuckets"></a>

### ExponentialHistogramDataPoint.Buckets
Buckets are a set of bucket counts, encoded in a contiguous array
of counts.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| offset | [sint32](#sint32) |  | Offset is the bucket index of the first entry in the bucket_counts array.

Note: This uses a varint encoding as a simple form of compression. |
| bucket_counts | [uint64](#uint64) | repeated | bucket_counts is an array of count values, where bucket_counts[i] carries the count of the bucket at index (offset+i). bucket_counts[i] is the count of values greater than base^(offset+i) and less than or equal to base^(offset+i+1).

Note: By contrast, the explicit HistogramDataPoint uses fixed64. This field is expected to have many buckets, especially zeros, so uint64 has been selected to ensure varint encoding. |






<a name="opentelemetryprotometricsv1gauge"></a>

### Gauge
Gauge represents the type of a scalar metric that always exports the
"current value" for every data point. It should be used for an "unknown"
aggregation.

A Gauge does not support different aggregation temporalities. Given the
aggregation is unknown, points cannot be combined using the same
aggregation, regardless of aggregation temporalities. Therefore,
AggregationTemporality is not included. Consequently, this also means
"StartTimeUnixNano" is ignored for all data points.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data_points | [NumberDataPoint](#opentelemetryprotometricsv1numberdatapoint) | repeated |  |






<a name="opentelemetryprotometricsv1histogram"></a>

### Histogram
Histogram represents the type of a metric that is calculated by aggregating
as a Histogram of all reported measurements over a time interval.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data_points | [HistogramDataPoint](#opentelemetryprotometricsv1histogramdatapoint) | repeated |  |
| aggregation_temporality | [AggregationTemporality](#opentelemetryprotometricsv1aggregationtemporality) |  | aggregation_temporality describes if the aggregator reports delta changes since last report time, or cumulative changes since a fixed start time. |






<a name="opentelemetryprotometricsv1histogramdatapoint"></a>

### HistogramDataPoint
HistogramDataPoint is a single data point in a timeseries that describes the
time-varying values of a Histogram. A Histogram contains summary statistics
for a population of values, it may optionally contain the distribution of
those values across a set of buckets.

If the histogram contains the distribution of values, then both
"explicit_bounds" and "bucket counts" fields must be defined.
If the histogram does not contain the distribution of values, then both
"explicit_bounds" and "bucket_counts" must be omitted and only "count" and
"sum" are known.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | The set of key/value pairs that uniquely identify the timeseries from where this point belongs. The list may be empty (may contain 0 elements). Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| start_time_unix_nano | [fixed64](#fixed64) |  | StartTimeUnixNano is optional but strongly encouraged, see the the detailed comments above Metric.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| time_unix_nano | [fixed64](#fixed64) |  | TimeUnixNano is required, see the detailed comments above Metric.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| count | [fixed64](#fixed64) |  | count is the number of values in the population. Must be non-negative. This value must be equal to the sum of the "count" fields in buckets if a histogram is provided. |
| sum | [double](#double) | optional | sum of the values in the population. If count is zero then this field must be zero.

Note: Sum should only be filled out when measuring non-negative discrete events, and is assumed to be monotonic over the values of these events. Negative events *can* be recorded, but sum should not be filled out when doing so. This is specifically to enforce compatibility w/ OpenMetrics, see: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#histogram |
| bucket_counts | [fixed64](#fixed64) | repeated | bucket_counts is an optional field contains the count values of histogram for each bucket.

The sum of the bucket_counts must equal the value in the count field.

The number of elements in bucket_counts array must be by one greater than the number of elements in explicit_bounds array. |
| explicit_bounds | [double](#double) | repeated | explicit_bounds specifies buckets with explicitly defined bounds for values.

The boundaries for bucket at index i are:

(-infinity, explicit_bounds[i]] for i == 0 (explicit_bounds[i-1], explicit_bounds[i]] for 0 < i < size(explicit_bounds) (explicit_bounds[i-1], +infinity) for i == size(explicit_bounds)

The values in the explicit_bounds array must be strictly increasing.

Histogram buckets are inclusive of their upper boundary, except the last bucket where the boundary is at infinity. This format is intentionally compatible with the OpenMetrics histogram definition. |
| exemplars | [Exemplar](#opentelemetryprotometricsv1exemplar) | repeated | (Optional) List of exemplars collected from measurements that were used to form the data point |
| flags | [uint32](#uint32) |  | Flags that apply to this specific data point. See DataPointFlags for the available flags and their meaning. |
| min | [double](#double) | optional | min is the minimum value over (start_time, end_time]. |
| max | [double](#double) | optional | max is the maximum value over (start_time, end_time]. |






<a name="opentelemetryprotometricsv1metric"></a>

### Metric
Defines a Metric which has one or more timeseries.  The following is a
brief summary of the Metric data model.  For more details, see:

  https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/data-model.md


The data model and relation between entities is shown in the
diagram below. Here, "DataPoint" is the term used to refer to any
one of the specific data point value types, and "points" is the term used
to refer to any one of the lists of points contained in the Metric.

- Metric is composed of a metadata and data.
- Metadata part contains a name, description, unit.
- Data is one of the possible types (Sum, Gauge, Histogram, Summary).
- DataPoint contains timestamps, attributes, and one of the possible value type
  fields.

    Metric
 +------------+
 |name        |
 |description |
 |unit        |     +------------------------------------+
 |data        |---> |Gauge, Sum, Histogram, Summary, ... |
 +------------+     +------------------------------------+

   Data [One of Gauge, Sum, Histogram, Summary, ...]
 +-----------+
 |...        |  // Metadata about the Data.
 |points     |--+
 +-----------+  |
                |      +---------------------------+
                |      |DataPoint 1                |
                v      |+------+------+   +------+ |
             +-----+   ||label |label |...|label | |
             |  1  |-->||value1|value2|...|valueN| |
             +-----+   |+------+------+   +------+ |
             |  .  |   |+-----+                    |
             |  .  |   ||value|                    |
             |  .  |   |+-----+                    |
             |  .  |   +---------------------------+
             |  .  |                   .
             |  .  |                   .
             |  .  |                   .
             |  .  |   +---------------------------+
             |  .  |   |DataPoint M                |
             +-----+   |+------+------+   +------+ |
             |  M  |-->||label |label |...|label | |
             +-----+   ||value1|value2|...|valueN| |
                       |+------+------+   +------+ |
                       |+-----+                    |
                       ||value|                    |
                       |+-----+                    |
                       +---------------------------+

Each distinct type of DataPoint represents the output of a specific
aggregation function, the result of applying the DataPoint's
associated function of to one or more measurements.

All DataPoint types have three common fields:
- Attributes includes key-value pairs associated with the data point
- TimeUnixNano is required, set to the end time of the aggregation
- StartTimeUnixNano is optional, but strongly encouraged for DataPoints
  having an AggregationTemporality field, as discussed below.

Both TimeUnixNano and StartTimeUnixNano values are expressed as
UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970.

# TimeUnixNano

This field is required, having consistent interpretation across
DataPoint types.  TimeUnixNano is the moment corresponding to when
the data point's aggregate value was captured.

Data points with the 0 value for TimeUnixNano SHOULD be rejected
by consumers.

# StartTimeUnixNano

StartTimeUnixNano in general allows detecting when a sequence of
observations is unbroken.  This field indicates to consumers the
start time for points with cumulative and delta
AggregationTemporality, and it should be included whenever possible
to support correct rate calculation.  Although it may be omitted
when the start time is truly unknown, setting StartTimeUnixNano is
strongly encouraged.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name of the metric. |
| description | [string](#string) |  | description of the metric, which can be used in documentation. |
| unit | [string](#string) |  | unit in which the metric value is reported. Follows the format described by http://unitsofmeasure.org/ucum.html. |
| gauge | [Gauge](#opentelemetryprotometricsv1gauge) |  |  |
| sum | [Sum](#opentelemetryprotometricsv1sum) |  |  |
| histogram | [Histogram](#opentelemetryprotometricsv1histogram) |  |  |
| exponential_histogram | [ExponentialHistogram](#opentelemetryprotometricsv1exponentialhistogram) |  |  |
| summary | [Summary](#opentelemetryprotometricsv1summary) |  |  |
| metadata | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | Additional metadata attributes that describe the metric. [Optional]. Attributes are non-identifying. Consumers SHOULD NOT need to be aware of these attributes. These attributes MAY be used to encode information allowing for lossless roundtrip translation to / from another data model. Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |






<a name="opentelemetryprotometricsv1metricsdata"></a>

### MetricsData
MetricsData represents the metrics data that can be stored in a persistent
storage, OR can be embedded by other protocols that transfer OTLP metrics
data but do not implement the OTLP protocol.

The main difference between this message and collector protocol is that
in this message there will not be any "control" or "metadata" specific to
OTLP protocol.

When new fields are added into this message, the OTLP request MUST be updated
as well.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_metrics | [ResourceMetrics](#opentelemetryprotometricsv1resourcemetrics) | repeated | An array of ResourceMetrics. For data coming from a single resource this array will typically contain one element. Intermediary nodes that receive data from multiple origins typically batch the data before forwarding further and in that case this array will contain multiple elements. |






<a name="opentelemetryprotometricsv1numberdatapoint"></a>

### NumberDataPoint
NumberDataPoint is a single data point in a timeseries that describes the
time-varying scalar value of a metric.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | The set of key/value pairs that uniquely identify the timeseries from where this point belongs. The list may be empty (may contain 0 elements). Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| start_time_unix_nano | [fixed64](#fixed64) |  | StartTimeUnixNano is optional but strongly encouraged, see the the detailed comments above Metric.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| time_unix_nano | [fixed64](#fixed64) |  | TimeUnixNano is required, see the detailed comments above Metric.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| as_double | [double](#double) |  |  |
| as_int | [sfixed64](#sfixed64) |  |  |
| exemplars | [Exemplar](#opentelemetryprotometricsv1exemplar) | repeated | (Optional) List of exemplars collected from measurements that were used to form the data point |
| flags | [uint32](#uint32) |  | Flags that apply to this specific data point. See DataPointFlags for the available flags and their meaning. |






<a name="opentelemetryprotometricsv1resourcemetrics"></a>

### ResourceMetrics
A collection of ScopeMetrics from a Resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [opentelemetry.proto.resource.v1.Resource](#opentelemetryprotoresourcev1resource) |  | The resource for the metrics in this message. If this field is not set then no resource info is known. |
| scope_metrics | [ScopeMetrics](#opentelemetryprotometricsv1scopemetrics) | repeated | A list of metrics that originate from a resource. |
| schema_url | [string](#string) |  | The Schema URL, if known. This is the identifier of the Schema that the resource data is recorded in. To learn more about Schema URL see https://opentelemetry.io/docs/specs/otel/schemas/#schema-url This schema_url applies to the data in the "resource" field. It does not apply to the data in the "scope_metrics" field which have their own schema_url field. |






<a name="opentelemetryprotometricsv1scopemetrics"></a>

### ScopeMetrics
A collection of Metrics produced by an Scope.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| scope | [opentelemetry.proto.common.v1.InstrumentationScope](#opentelemetryprotocommonv1instrumentationscope) |  | The instrumentation scope information for the metrics in this message. Semantically when InstrumentationScope isn't set, it is equivalent with an empty instrumentation scope name (unknown). |
| metrics | [Metric](#opentelemetryprotometricsv1metric) | repeated | A list of metrics that originate from an instrumentation library. |
| schema_url | [string](#string) |  | The Schema URL, if known. This is the identifier of the Schema that the metric data is recorded in. To learn more about Schema URL see https://opentelemetry.io/docs/specs/otel/schemas/#schema-url This schema_url applies to all metrics in the "metrics" field. |






<a name="opentelemetryprotometricsv1sum"></a>

### Sum
Sum represents the type of a scalar metric that is calculated as a sum of all
reported measurements over a time interval.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data_points | [NumberDataPoint](#opentelemetryprotometricsv1numberdatapoint) | repeated |  |
| aggregation_temporality | [AggregationTemporality](#opentelemetryprotometricsv1aggregationtemporality) |  | aggregation_temporality describes if the aggregator reports delta changes since last report time, or cumulative changes since a fixed start time. |
| is_monotonic | [bool](#bool) |  | If "true" means that the sum is monotonic. |






<a name="opentelemetryprotometricsv1summary"></a>

### Summary
Summary metric data are used to convey quantile summaries,
a Prometheus (see: https://prometheus.io/docs/concepts/metric_types/#summary)
and OpenMetrics (see: https://github.com/OpenObservability/OpenMetrics/blob/4dbf6075567ab43296eed941037c12951faafb92/protos/prometheus.proto#L45)
data type. These data points cannot always be merged in a meaningful way.
While they can be useful in some applications, histogram data points are
recommended for new applications.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data_points | [SummaryDataPoint](#opentelemetryprotometricsv1summarydatapoint) | repeated |  |






<a name="opentelemetryprotometricsv1summarydatapoint"></a>

### SummaryDataPoint
SummaryDataPoint is a single data point in a timeseries that describes the
time-varying values of a Summary metric.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | The set of key/value pairs that uniquely identify the timeseries from where this point belongs. The list may be empty (may contain 0 elements). Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| start_time_unix_nano | [fixed64](#fixed64) |  | StartTimeUnixNano is optional but strongly encouraged, see the the detailed comments above Metric.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| time_unix_nano | [fixed64](#fixed64) |  | TimeUnixNano is required, see the detailed comments above Metric.

Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970. |
| count | [fixed64](#fixed64) |  | count is the number of values in the population. Must be non-negative. |
| sum | [double](#double) |  | sum of the values in the population. If count is zero then this field must be zero.

Note: Sum should only be filled out when measuring non-negative discrete events, and is assumed to be monotonic over the values of these events. Negative events *can* be recorded, but sum should not be filled out when doing so. This is specifically to enforce compatibility w/ OpenMetrics, see: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#summary |
| quantile_values | [SummaryDataPoint.ValueAtQuantile](#opentelemetryprotometricsv1summarydatapointvalueatquantile) | repeated | (Optional) list of values at different quantiles of the distribution calculated from the current snapshot. The quantiles must be strictly increasing. |
| flags | [uint32](#uint32) |  | Flags that apply to this specific data point. See DataPointFlags for the available flags and their meaning. |






<a name="opentelemetryprotometricsv1summarydatapointvalueatquantile"></a>

### SummaryDataPoint.ValueAtQuantile
Represents the value at a given quantile of a distribution.

To record Min and Max values following conventions are used:
- The 1.0 quantile is equivalent to the maximum value observed.
- The 0.0 quantile is equivalent to the minimum value observed.

See the following issue for more context:
https://github.com/open-telemetry/opentelemetry-proto/issues/125


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| quantile | [double](#double) |  | The quantile of a distribution. Must be in the interval [0.0, 1.0]. |
| value | [double](#double) |  | The value at the given quantile of a distribution.

Quantile values must NOT be negative. |





 <!-- end messages -->


<a name="opentelemetryprotometricsv1aggregationtemporality"></a>

### AggregationTemporality
AggregationTemporality defines how a metric aggregator reports aggregated
values. It describes how those values relate to the time interval over
which they are aggregated.

| Name | Number | Description |
| ---- | ------ | ----------- |
| AGGREGATION_TEMPORALITY_UNSPECIFIED | 0 | UNSPECIFIED is the default AggregationTemporality, it MUST not be used. |
| AGGREGATION_TEMPORALITY_DELTA | 1 | DELTA is an AggregationTemporality for a metric aggregator which reports changes since last report time. Successive metrics contain aggregation of values from continuous and non-overlapping intervals.

The values for a DELTA metric are based only on the time interval associated with one measurement cycle. There is no dependency on previous measurements like is the case for CUMULATIVE metrics.

For example, consider a system measuring the number of requests that it receives and reports the sum of these requests every second as a DELTA metric:

 1. The system starts receiving at time=t_0. 2. A request is received, the system measures 1 request. 3. A request is received, the system measures 1 request. 4. A request is received, the system measures 1 request. 5. The 1 second collection cycle ends. A metric is exported for the number of requests received over the interval of time t_0 to t_0+1 with a value of 3. 6. A request is received, the system measures 1 request. 7. A request is received, the system measures 1 request. 8. The 1 second collection cycle ends. A metric is exported for the number of requests received over the interval of time t_0+1 to t_0+2 with a value of 2. |
| AGGREGATION_TEMPORALITY_CUMULATIVE | 2 | CUMULATIVE is an AggregationTemporality for a metric aggregator which reports changes since a fixed start time. This means that current values of a CUMULATIVE metric depend on all previous measurements since the start time. Because of this, the sender is required to retain this state in some form. If this state is lost or invalidated, the CUMULATIVE metric values MUST be reset and a new fixed start time following the last reported measurement time sent MUST be used.

For example, consider a system measuring the number of requests that it receives and reports the sum of these requests every second as a CUMULATIVE metric:

 1. The system starts receiving at time=t_0. 2. A request is received, the system measures 1 request. 3. A request is received, the system measures 1 request. 4. A request is received, the system measures 1 request. 5. The 1 second collection cycle ends. A metric is exported for the number of requests received over the interval of time t_0 to t_0+1 with a value of 3. 6. A request is received, the system measures 1 request. 7. A request is received, the system measures 1 request. 8. The 1 second collection cycle ends. A metric is exported for the number of requests received over the interval of time t_0 to t_0+2 with a value of 5. 9. The system experiences a fault and loses state. 10. The system recovers and resumes receiving at time=t_1. 11. A request is received, the system measures 1 request. 12. The 1 second collection cycle ends. A metric is exported for the number of requests received over the interval of time t_1 to t_0+1 with a value of 1.

Note: Even though, when reporting changes since last report time, using CUMULATIVE is valid, it is not recommended. This may cause problems for systems that do not use start_time to determine when the aggregation value was reset (e.g. Prometheus). |



<a name="opentelemetryprotometricsv1datapointflags"></a>

### DataPointFlags
DataPointFlags is defined as a protobuf 'uint32' type and is to be used as a
bit-field representing 32 distinct boolean flags.  Each flag defined in this
enum is a bit-mask.  To test the presence of a single flag in the flags of
a data point, for example, use an expression like:

  (point.flags & DATA_POINT_FLAGS_NO_RECORDED_VALUE_MASK) == DATA_POINT_FLAGS_NO_RECORDED_VALUE_MASK

| Name | Number | Description |
| ---- | ------ | ----------- |
| DATA_POINT_FLAGS_DO_NOT_USE | 0 | The zero value for the enum. Should not be used for comparisons. Instead use bitwise "and" with the appropriate mask as shown above. |
| DATA_POINT_FLAGS_NO_RECORDED_VALUE_MASK | 1 | This DataPoint is valid but has no recorded value. This value SHOULD be used to reflect explicitly missing data in a series, as for an equivalent to the Prometheus "staleness marker". |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->





<a name="opentelemetryprotoresourcev1resource"></a>

### Resource
Resource information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | Set of attributes that describe the resource. Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| dropped_attributes_count | [uint32](#uint32) |  | dropped_attributes_count is the number of dropped attributes. If the value is 0, then no attributes were dropped. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->





<a name="opentelemetryprototracev1resourcespans"></a>

### ResourceSpans
A collection of ScopeSpans from a Resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [opentelemetry.proto.resource.v1.Resource](#opentelemetryprotoresourcev1resource) |  | The resource for the spans in this message. If this field is not set then no resource info is known. |
| scope_spans | [ScopeSpans](#opentelemetryprototracev1scopespans) | repeated | A list of ScopeSpans that originate from a resource. |
| schema_url | [string](#string) |  | The Schema URL, if known. This is the identifier of the Schema that the resource data is recorded in. To learn more about Schema URL see https://opentelemetry.io/docs/specs/otel/schemas/#schema-url This schema_url applies to the data in the "resource" field. It does not apply to the data in the "scope_spans" field which have their own schema_url field. |






<a name="opentelemetryprototracev1scopespans"></a>

### ScopeSpans
A collection of Spans produced by an InstrumentationScope.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| scope | [opentelemetry.proto.common.v1.InstrumentationScope](#opentelemetryprotocommonv1instrumentationscope) |  | The instrumentation scope information for the spans in this message. Semantically when InstrumentationScope isn't set, it is equivalent with an empty instrumentation scope name (unknown). |
| spans | [Span](#opentelemetryprototracev1span) | repeated | A list of Spans that originate from an instrumentation scope. |
| schema_url | [string](#string) |  | The Schema URL, if known. This is the identifier of the Schema that the span data is recorded in. To learn more about Schema URL see https://opentelemetry.io/docs/specs/otel/schemas/#schema-url This schema_url applies to all spans and span events in the "spans" field. |






<a name="opentelemetryprototracev1span"></a>

### Span
A Span represents a single operation performed by a single component of the system.

The next available field id is 17.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trace_id | [bytes](#bytes) |  | A unique identifier for a trace. All spans from the same trace share the same `trace_id`. The ID is a 16-byte array. An ID with all zeroes OR of length other than 16 bytes is considered invalid (empty string in OTLP/JSON is zero-length and thus is also invalid).

This field is required. |
| span_id | [bytes](#bytes) |  | A unique identifier for a span within a trace, assigned when the span is created. The ID is an 8-byte array. An ID with all zeroes OR of length other than 8 bytes is considered invalid (empty string in OTLP/JSON is zero-length and thus is also invalid).

This field is required. |
| trace_state | [string](#string) |  | trace_state conveys information about request position in multiple distributed tracing graphs. It is a trace_state in w3c-trace-context format: https://www.w3.org/TR/trace-context/#tracestate-header See also https://github.com/w3c/distributed-tracing for more details about this field. |
| parent_span_id | [bytes](#bytes) |  | The `span_id` of this span's parent span. If this is a root span, then this field must be empty. The ID is an 8-byte array. |
| flags | [fixed32](#fixed32) |  | Flags, a bit field.

Bits 0-7 (8 least significant bits) are the trace flags as defined in W3C Trace Context specification. To read the 8-bit W3C trace flag, use `flags & SPAN_FLAGS_TRACE_FLAGS_MASK`.

See https://www.w3.org/TR/trace-context-2/#trace-flags for the flag definitions.

Bits 8 and 9 represent the 3 states of whether a span's parent is remote. The states are (unknown, is not remote, is remote). To read whether the value is known, use `(flags & SPAN_FLAGS_CONTEXT_HAS_IS_REMOTE_MASK) != 0`. To read whether the span is remote, use `(flags & SPAN_FLAGS_CONTEXT_IS_REMOTE_MASK) != 0`.

When creating span messages, if the message is logically forwarded from another source with an equivalent flags fields (i.e., usually another OTLP span message), the field SHOULD be copied as-is. If creating from a source that does not have an equivalent flags field (such as a runtime representation of an OpenTelemetry span), the high 22 bits MUST be set to zero. Readers MUST NOT assume that bits 10-31 (22 most significant bits) will be zero.

[Optional]. |
| name | [string](#string) |  | A description of the span's operation.

For example, the name can be a qualified method name or a file name and a line number where the operation is called. A best practice is to use the same display name at the same call point in an application. This makes it easier to correlate spans in different traces.

This field is semantically required to be set to non-empty string. Empty value is equivalent to an unknown span name.

This field is required. |
| kind | [Span.SpanKind](#opentelemetryprototracev1spanspankind) |  | Distinguishes between spans generated in a particular context. For example, two spans with the same name may be distinguished using `CLIENT` (caller) and `SERVER` (callee) to identify queueing latency associated with the span. |
| start_time_unix_nano | [fixed64](#fixed64) |  | start_time_unix_nano is the start time of the span. On the client side, this is the time kept by the local machine where the span execution starts. On the server side, this is the time when the server's application handler starts running. Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970.

This field is semantically required and it is expected that end_time >= start_time. |
| end_time_unix_nano | [fixed64](#fixed64) |  | end_time_unix_nano is the end time of the span. On the client side, this is the time kept by the local machine where the span execution ends. On the server side, this is the time when the server application handler stops running. Value is UNIX Epoch time in nanoseconds since 00:00:00 UTC on 1 January 1970.

This field is semantically required and it is expected that end_time >= start_time. |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | attributes is a collection of key/value pairs. Note, global attributes like server name can be set using the resource API. Examples of attributes:

 "/http/user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36" "/http/server_latency": 300 "example.com/myattribute": true "example.com/score": 10.239

The OpenTelemetry API specification further restricts the allowed value types: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/common/README.md#attribute Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| dropped_attributes_count | [uint32](#uint32) |  | dropped_attributes_count is the number of attributes that were discarded. Attributes can be discarded because their keys are too long or because there are too many attributes. If this value is 0, then no attributes were dropped. |
| events | [Span.Event](#opentelemetryprototracev1spanevent) | repeated | events is a collection of Event items. |
| dropped_events_count | [uint32](#uint32) |  | dropped_events_count is the number of dropped events. If the value is 0, then no events were dropped. |
| links | [Span.Link](#opentelemetryprototracev1spanlink) | repeated | links is a collection of Links, which are references from this span to a span in the same or different trace. |
| dropped_links_count | [uint32](#uint32) |  | dropped_links_count is the number of dropped links after the maximum size was enforced. If this value is 0, then no links were dropped. |
| status | [Status](#opentelemetryprototracev1status) |  | An optional final status for this span. Semantically when Status isn't set, it means span's status code is unset, i.e. assume STATUS_CODE_UNSET (code = 0). |






<a name="opentelemetryprototracev1spanevent"></a>

### Span.Event
Event is a time-stamped annotation of the span, consisting of user-supplied
text description and key-value pairs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| time_unix_nano | [fixed64](#fixed64) |  | time_unix_nano is the time the event occurred. |
| name | [string](#string) |  | name of the event. This field is semantically required to be set to non-empty string. |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | attributes is a collection of attribute key/value pairs on the event. Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| dropped_attributes_count | [uint32](#uint32) |  | dropped_attributes_count is the number of dropped attributes. If the value is 0, then no attributes were dropped. |






<a name="opentelemetryprototracev1spanlink"></a>

### Span.Link
A pointer from the current span to another span in the same trace or in a
different trace. For example, this can be used in batching operations,
where a single batch handler processes multiple requests from different
traces or when the handler receives a request from a different project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trace_id | [bytes](#bytes) |  | A unique identifier of a trace that this linked span is part of. The ID is a 16-byte array. |
| span_id | [bytes](#bytes) |  | A unique identifier for the linked span. The ID is an 8-byte array. |
| trace_state | [string](#string) |  | The trace_state associated with the link. |
| attributes | [opentelemetry.proto.common.v1.KeyValue](#opentelemetryprotocommonv1keyvalue) | repeated | attributes is a collection of attribute key/value pairs on the link. Attribute keys MUST be unique (it is not allowed to have more than one attribute with the same key). |
| dropped_attributes_count | [uint32](#uint32) |  | dropped_attributes_count is the number of dropped attributes. If the value is 0, then no attributes were dropped. |
| flags | [fixed32](#fixed32) |  | Flags, a bit field.

Bits 0-7 (8 least significant bits) are the trace flags as defined in W3C Trace Context specification. To read the 8-bit W3C trace flag, use `flags & SPAN_FLAGS_TRACE_FLAGS_MASK`.

See https://www.w3.org/TR/trace-context-2/#trace-flags for the flag definitions.

Bits 8 and 9 represent the 3 states of whether the link is remote. The states are (unknown, is not remote, is remote). To read whether the value is known, use `(flags & SPAN_FLAGS_CONTEXT_HAS_IS_REMOTE_MASK) != 0`. To read whether the link is remote, use `(flags & SPAN_FLAGS_CONTEXT_IS_REMOTE_MASK) != 0`.

Readers MUST NOT assume that bits 10-31 (22 most significant bits) will be zero. When creating new spans, bits 10-31 (most-significant 22-bits) MUST be zero.

[Optional]. |






<a name="opentelemetryprototracev1status"></a>

### Status
The Status type defines a logical error model that is suitable for different
programming environments, including REST APIs and RPC APIs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  | A developer-facing human readable error message. |
| code | [Status.StatusCode](#opentelemetryprototracev1statusstatuscode) |  | The status code. |






<a name="opentelemetryprototracev1tracesdata"></a>

### TracesData
TracesData represents the traces data that can be stored in a persistent storage,
OR can be embedded by other protocols that transfer OTLP traces data but do
not implement the OTLP protocol.

The main difference between this message and collector protocol is that
in this message there will not be any "control" or "metadata" specific to
OTLP protocol.

When new fields are added into this message, the OTLP request MUST be updated
as well.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_spans | [ResourceSpans](#opentelemetryprototracev1resourcespans) | repeated | An array of ResourceSpans. For data coming from a single resource this array will typically contain one element. Intermediary nodes that receive data from multiple origins typically batch the data before forwarding further and in that case this array will contain multiple elements. |





 <!-- end messages -->


<a name="opentelemetryprototracev1spanspankind"></a>

### Span.SpanKind
SpanKind is the type of span. Can be used to specify additional relationships between spans
in addition to a parent/child relationship.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SPAN_KIND_UNSPECIFIED | 0 | Unspecified. Do NOT use as default. Implementations MAY assume SpanKind to be INTERNAL when receiving UNSPECIFIED. |
| SPAN_KIND_INTERNAL | 1 | Indicates that the span represents an internal operation within an application, as opposed to an operation happening at the boundaries. Default value. |
| SPAN_KIND_SERVER | 2 | Indicates that the span covers server-side handling of an RPC or other remote network request. |
| SPAN_KIND_CLIENT | 3 | Indicates that the span describes a request to some remote service. |
| SPAN_KIND_PRODUCER | 4 | Indicates that the span describes a producer sending a message to a broker. Unlike CLIENT and SERVER, there is often no direct critical path latency relationship between producer and consumer spans. A PRODUCER span ends when the message was accepted by the broker while the logical processing of the message might span a much longer time. |
| SPAN_KIND_CONSUMER | 5 | Indicates that the span describes consumer receiving a message from a broker. Like the PRODUCER kind, there is often no direct critical path latency relationship between producer and consumer spans. |



<a name="opentelemetryprototracev1spanflags"></a>

### SpanFlags
SpanFlags represents constants used to interpret the
Span.flags field, which is protobuf 'fixed32' type and is to
be used as bit-fields. Each non-zero value defined in this enum is
a bit-mask.  To extract the bit-field, for example, use an
expression like:

  (span.flags & SPAN_FLAGS_TRACE_FLAGS_MASK)

See https://www.w3.org/TR/trace-context-2/#trace-flags for the flag definitions.

Note that Span flags were introduced in version 1.1 of the
OpenTelemetry protocol.  Older Span producers do not set this
field, consequently consumers should not rely on the absence of a
particular flag bit to indicate the presence of a particular feature.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SPAN_FLAGS_DO_NOT_USE | 0 | The zero value for the enum. Should not be used for comparisons. Instead use bitwise "and" with the appropriate mask as shown above. |
| SPAN_FLAGS_TRACE_FLAGS_MASK | 255 | Bits 0-7 are used for trace flags. |
| SPAN_FLAGS_CONTEXT_HAS_IS_REMOTE_MASK | 256 | Bits 8 and 9 are used to indicate that the parent span or link span is remote. Bit 8 (`HAS_IS_REMOTE`) indicates whether the value is known. Bit 9 (`IS_REMOTE`) indicates whether the span or link is remote. |
| SPAN_FLAGS_CONTEXT_IS_REMOTE_MASK | 512 |  |



<a name="opentelemetryprototracev1statusstatuscode"></a>

### Status.StatusCode
For the semantics of status codes see
https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#set-status

| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_CODE_UNSET | 0 | The default status. |
| STATUS_CODE_OK | 1 | The Span has been validated by an Application developer or Operator to have completed successfully. |
| STATUS_CODE_ERROR | 2 | The Span contains an error. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

